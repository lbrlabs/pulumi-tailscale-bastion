package aws

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

//go:embed cloudinit/*.sh cloudinit/*.tmpl
var cloudinitFiles embed.FS

var terraformVariable = regexp.MustCompile(`\$\{([A-Z_]+)\}`)

// Omitted inputs retain the renderer defaults; explicit false and zero values
// must survive resolution of Pulumi outputs.
func cloudInitOptions(args *BastionArgs) pulumi.Map {
	options := pulumi.Map{}
	for key, value := range map[string]pulumi.Input{
		"TRACK": args.Track, "MAX_RETRIES": args.MaxRetries, "RETRY_DELAY": args.RetryDelay,
		"JSON": args.JSON, "LOGIN_SERVER": args.LoginServer, "RESET": args.Reset,
		"TIMEOUT": args.Timeout, "FORCE_REAUTH": args.ForceReauth,
		"ACCEPT_DNS": args.AcceptDNS, "ACCEPT_ROUTES": args.AcceptRoutes,
		"SHIELDS_UP": args.ShieldsUp, "SNAT_SUBNET_ROUTES": args.SNATSubnetRoutes,
		"NETFILTER_MODE": args.NetfilterMode, "STATEFUL_FILTERING": args.StatefulFiltering,
		"EXIT_NODE": args.ExitNode, "EXIT_NODE_ALLOW_LAN_ACCESS": args.ExitNodeAllowLANAccess,
		"OPERATOR": args.Operator,
	} {
		if value != nil {
			options[key] = value
		}
	}
	return options
}

// renderUpstream renders the simple interpolation subset used by the pinned
// upstream templates. Fail on unknown variables so upstream updates are explicit.
func renderUpstream(name string, values map[string]string) (string, error) {
	content, err := cloudinitFiles.ReadFile("cloudinit/" + name)
	if err != nil {
		return "", err
	}
	var missing string
	result := terraformVariable.ReplaceAllStringFunc(string(content), func(match string) string {
		key := terraformVariable.FindStringSubmatch(match)[1]
		value, ok := values[key]
		if !ok {
			missing = key
		}
		return value
	})
	if missing != "" {
		return "", fmt.Errorf("unknown upstream template variable %s", missing)
	}
	return strings.ReplaceAll(result, "$${", "${"), nil
}

func renderUserData(d UserDataArgs) (string, error) {
	// Upstream builds an eval command. Prevent shell expansion at both its
	// interpolation and execution stages for user-controlled values.
	for _, value := range []string{d.ClientID, d.Audience, d.Hostname, d.Routes, d.TailscaleTags, d.ParameterName, d.Region} {
		if strings.ContainsAny(value, "\"'`$\\\r\n") {
			return "", fmt.Errorf("cloud-init settings contain unsupported shell characters")
		}
	}
	if d.ParameterName == "" && (strings.TrimSpace(d.ClientID) == "" || strings.TrimSpace(d.Audience) == "") {
		return "", fmt.Errorf("workloadIdentity requires non-empty clientId and audience")
	}
	values := map[string]string{
		"TRACK": "stable", "MAX_RETRIES": "10", "RETRY_DELAY": "10",
		"JSON": "false", "LOGIN_SERVER": "https://controlplane.tailscale.com",
		"RESET": "false", "TIMEOUT": "60s", "FORCE_REAUTH": "false",
		"AUTH_KEY": "", "ID_TOKEN": "", "CLIENT_ID": "", "CLIENT_SECRET": "", "AUDIENCE": d.Audience,
		"HOSTNAME": d.Hostname, "ADVERTISE_TAGS": d.TailscaleTags, "ADVERTISE_ROUTES": d.Routes,
		"ACCEPT_DNS": "true", "ACCEPT_ROUTES": "false",
		"ADVERTISE_EXIT_NODE": strconv.FormatBool(d.EnableExitNode),
		"ADVERTISE_CONNECTOR": strconv.FormatBool(d.EnableAppConnector),
		"SHIELDS_UP":          "false", "TAILSCALE_SSH": strconv.FormatBool(d.EnableSSH),
		"SNAT_SUBNET_ROUTES": "true", "NETFILTER_MODE": "on", "STATEFUL_FILTERING": "false",
		"EXIT_NODE": "", "EXIT_NODE_ALLOW_LAN_ACCESS": "false", "OPERATOR": "", "RELAY_SERVER_PORT": "",
	}
	for key, value := range d.Options {
		text := fmt.Sprint(value)
		if strings.ContainsAny(text, "\"'`$\\\r\n") {
			return "", fmt.Errorf("cloud-init setting %s contains unsupported shell characters", key)
		}
		values[key] = text
	}
	if values["TRACK"] != "stable" && values["TRACK"] != "unstable" {
		return "", fmt.Errorf("track must be stable or unstable")
	}
	if mode := values["NETFILTER_MODE"]; mode != "on" && mode != "off" && mode != "nodivert" {
		return "", fmt.Errorf("netfilterMode must be on, off, or nodivert")
	}
	for key, minimum := range map[string]int{"MAX_RETRIES": 1, "RETRY_DELAY": 0} {
		value, err := strconv.Atoi(values[key])
		if err != nil || value < minimum {
			return "", fmt.Errorf("%s must be an integer >= %d", key, minimum)
		}
	}
	if duration, err := time.ParseDuration(values["TIMEOUT"]); err != nil || duration < 0 {
		return "", fmt.Errorf("timeout must be a non-negative duration (for example 60s)")
	}
	if d.ClientID != "" {
		values["CLIENT_ID"] = d.ClientID + "?ephemeral=true&preauthorized=true"
	} else {
		values["AUTH_KEY"] = "file:/run/tailscale-authkey"
	}
	if d.PeerRelayEnable {
		values["RELAY_SERVER_PORT"] = strconv.Itoa(d.PeerRelayPort)
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	// Stable user data avoids launch template diffs on every update.
	if err := w.SetBoundary("tailscale-bastion-cloudinit"); err != nil {
		return "", err
	}
	writePart := func(name, script string) error {
		h := textproto.MIMEHeader{}
		h.Set("Content-Type", "text/x-shellscript; charset=us-ascii")
		h.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", name))
		part, err := w.CreatePart(h)
		if err != nil {
			return err
		}
		_, err = part.Write([]byte(script))
		return err
	}
	// Amazon Linux 2023 includes SSM and AWS CLI on both supported architectures.
	bootstrap := "#!/bin/sh\nset -eu\nsystemctl enable --now amazon-ssm-agent\ndnf install -y ec2-instance-connect ethtool\n"
	if err := writePart("00-aws.sh", bootstrap); err != nil {
		return "", err
	}
	for i, name := range []string{"ip_forwarding.sh", "udp_offloads.sh", "install_tailscale.sh.tmpl", "setup_tailscale.sh.tmpl"} {
		script, err := renderUpstream(name, values)
		if err != nil {
			return "", err
		}
		if name == "setup_tailscale.sh.tmpl" {
			prefix := "#!/bin/sh\nset -e\n"
			if d.ParameterName != "" {
				prefix += fmt.Sprintf("umask 077\naws ssm get-parameter --name '%s' --region '%s' --with-decryption --query Parameter.Value --output text > /run/tailscale-authkey\ntrap 'rm -f /run/tailscale-authkey' EXIT\n", d.ParameterName, d.Region)
			}
			// Keep errexit disabled for the upstream authentication retry loop,
			// then enable it before applying preferences so failures are reported.
			success := `echo "Tailscale up succeeded, now configuring settings..."`
			if strings.Count(script, success) != 1 {
				return "", fmt.Errorf("upstream setup script has unexpected success marker")
			}
			script = strings.Replace(script, success, "set -e\n    "+success, 1)
			// Make exhausted authentication retries fail cloud-init.
			script = prefix + "set +e\n" + script + "\nexit 1\n"
		}
		if err := writePart(fmt.Sprintf("%02d-%s", i+1, strings.TrimSuffix(name, ".tmpl")), script); err != nil {
			return "", err
		}
	}
	if err := w.Close(); err != nil {
		return "", err
	}
	data := "MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=" + w.Boundary() + "\r\n\r\n" + body.String()
	return base64.StdEncoding.EncodeToString([]byte(data)), nil
}
