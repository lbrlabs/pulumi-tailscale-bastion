package aws

import (
	"encoding/base64"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func TestCloudInitAuthentication(t *testing.T) {
	for _, federated := range []bool{false, true} {
		t.Run(map[bool]string{true: "federated", false: "ssm"}[federated], func(t *testing.T) {
			d := UserDataArgs{Hostname: "router", Region: "us-west-2", TailscaleTags: "tag:router", Routes: "10.0.0.0/8", EnableSSH: true, EnableExitNode: true, EnableAppConnector: true, PeerRelayEnable: true, PeerRelayPort: 12345}
			if federated {
				d.Options = map[string]interface{}{
					"TRACK": "unstable", "MAX_RETRIES": 3, "RETRY_DELAY": 0,
					"JSON": true, "LOGIN_SERVER": "https://control.example.com", "RESET": true,
					"TIMEOUT": "2m", "FORCE_REAUTH": true, "ACCEPT_DNS": false,
					"ACCEPT_ROUTES": true, "SHIELDS_UP": true, "SNAT_SUBNET_ROUTES": false,
					"NETFILTER_MODE": "nodivert", "STATEFUL_FILTERING": true,
					"EXIT_NODE": "100.64.0.1", "EXIT_NODE_ALLOW_LAN_ACCESS": true, "OPERATOR": "ec2-user",
				}
			}
			if federated {
				d.ClientID, d.Audience = "client123", "api.tailscale.com/client123"
			} else {
				d.ParameterName = "/bastion/key"
			}
			encoded, err := renderUserData(d)
			if err != nil {
				t.Fatal(err)
			}
			again, _ := renderUserData(d)
			if encoded != again {
				t.Fatal("user data must be deterministic")
			}
			data, err := base64.StdEncoding.DecodeString(encoded)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) > 16384 {
				t.Fatalf("user data exceeds EC2 limit: %d", len(data))
			}
			message, err := mail.ReadMessage(strings.NewReader(string(data)))
			if err != nil {
				t.Fatal(err)
			}
			_, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
			if err != nil {
				t.Fatal(err)
			}
			reader := multipart.NewReader(message.Body, params["boundary"])
			count := 0
			var setup string
			for {
				part, err := reader.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				script, _ := io.ReadAll(part)
				cmd := exec.Command("sh", "-n")
				cmd.Stdin = strings.NewReader(string(script))
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s: %s: %v", part.FileName(), out, err)
				}
				if strings.Contains(part.FileName(), "setup_tailscale") {
					setup = string(script)
				}
				if federated && strings.Contains(part.FileName(), "install_tailscale") {
					for _, expected := range []string{`TRACK="unstable"`, "max_retries=3", "retry_delay=0"} {
						if !strings.Contains(string(script), expected) {
							t.Errorf("missing install option %s", expected)
						}
					}
				}
				count++
			}
			if count != 5 {
				t.Fatalf("expected 5 cloud-init parts, got %d", count)
			}
			if !federated {
				if !strings.Contains(setup, "aws ssm get-parameter --name '/bastion/key'") || !strings.Contains(setup, "trap 'rm -f /run/tailscale-authkey' EXIT") {
					t.Fatal("missing SSM retrieval/cleanup")
				}
				return
			}
			if strings.Contains(setup, "aws ssm") {
				t.Fatal("federation must not read a key")
			}
			// Exercise the upstream shell command construction and retry behavior.
			dir := t.TempDir()
			log := filepath.Join(dir, "calls")
			for _, command := range []string{"sudo", "sleep", "tailscale"} {
				script := "#!/bin/sh\nexit 0\n"
				if command == "tailscale" {
					script = "#!/bin/sh\nprintf '%s\\n' \"$@\" >> \"$CALL_LOG\"\nif [ \"$1\" = set ] && [ \"${FAIL_SET:-0}\" = 1 ]; then exit 1; fi\nexit \"${FAIL:-0}\"\n"
				}
				if err := os.WriteFile(filepath.Join(dir, command), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			run := func(fail string) error {
				cmd := exec.Command("sh")
				cmd.Stdin = strings.NewReader(setup)
				cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "CALL_LOG="+log, "FAIL="+fail)
				return cmd.Run()
			}
			if err := run("0"); err != nil {
				t.Fatal(err)
			}
			calls, _ := os.ReadFile(log)
			for _, flag := range []string{"--client-id=client123?ephemeral=true&preauthorized=true", "--audience=api.tailscale.com/client123", "--advertise-routes=10.0.0.0/8", "--advertise-tags=tag:router", "--ssh=true", "--advertise-exit-node=true", "--advertise-connector=true", "--relay-server-port=12345"} {
				if !strings.Contains(string(calls), flag+"\n") {
					t.Errorf("missing argument %s", flag)
				}
			}
			for _, flag := range []string{"--json=true", "--login-server=https://control.example.com", "--reset=true", "--timeout=2m", "--force-reauth=true", "--accept-dns=false", "--accept-routes=true", "--shields-up=true", "--snat-subnet-routes=false", "--netfilter-mode=nodivert", "--stateful-filtering=true", "--exit-node=100.64.0.1", "--exit-node-allow-lan-access=true", "--operator=ec2-user"} {
				if !strings.Contains(string(calls), flag+"\n") {
					t.Errorf("missing argument %s", flag)
				}
			}
			if err := run("1"); err == nil {
				t.Fatal("exhausted authentication retries must fail")
			}
			calls, _ = os.ReadFile(log)
			if count := strings.Count(string(calls), "up\n"); count != 4 {
				t.Errorf("expected one successful call and three retry attempts, got %d", count)
			}
			cmd := exec.Command("sh")
			cmd.Stdin = strings.NewReader(setup)
			cmd.Env = append(os.Environ(), "PATH="+dir+":"+os.Getenv("PATH"), "CALL_LOG="+log, "FAIL_SET=1")
			if err := cmd.Run(); err == nil {
				t.Fatal("failed tailscale set must fail cloud-init")
			}
		})
	}
}

func TestCloudInitOptionValidation(t *testing.T) {
	for key, value := range map[string]interface{}{
		"TRACK": "invalid", "NETFILTER_MODE": "invalid", "MAX_RETRIES": 0,
		"RETRY_DELAY": -1, "TIMEOUT": "-1s", "LOGIN_SERVER": "$(id)",
		"EXIT_NODE": "`id`", "OPERATOR": "user\ncommand",
	} {
		t.Run(key, func(t *testing.T) {
			_, err := renderUserData(UserDataArgs{ParameterName: "/key", Options: map[string]interface{}{key: value}})
			if err == nil {
				t.Fatal("expected invalid option to be rejected")
			}
		})
	}
}

func TestCloudInitOutputInputs(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		options := cloudInitOptions(&BastionArgs{
			AcceptDNS:  pulumi.Bool(false).ToBoolOutput(),
			RetryDelay: pulumi.Int(0).ToIntOutput(),
			Track:      pulumi.String("unstable").ToStringOutput(),
		})
		options.ToMapOutput().ApplyT(func(values map[string]interface{}) error {
			if len(values) != 3 || values["ACCEPT_DNS"] != false || values["RETRY_DELAY"] != 0 || values["TRACK"] != "unstable" {
				t.Errorf("unexpected resolved options: %#v", values)
			}
			return nil
		})
		return nil
	}, pulumi.WithMocks("project", "stack", &bastionMocks{}))
	if err != nil {
		t.Fatal(err)
	}
}

func TestCloudInitRejectsInvalidIdentity(t *testing.T) {
	for _, d := range []UserDataArgs{{}, {ClientID: "client"}, {ClientID: "client", Audience: "$(id)"}, {ClientID: "client", Audience: "audience", Hostname: "`id`"}} {
		if _, err := renderUserData(d); err == nil {
			t.Fatal("expected invalid settings to be rejected")
		}
	}
}
