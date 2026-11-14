package aws

import (
	"encoding/base64"
	"strings"
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type bastionMocks struct {
	sync.Mutex
	resources []pulumi.MockResourceArgs
}

func (m *bastionMocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return resource.NewPropertyMapFromMap(map[string]interface{}{
		"partition": "aws", "dnsSuffix": "amazonaws.com",
	}), nil
}

func (m *bastionMocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	m.Lock()
	m.resources = append(m.resources, args)
	m.Unlock()
	outputs := args.Inputs.Copy()
	outputs["name"] = resource.NewStringProperty(args.Name)
	outputs["arn"] = resource.NewStringProperty("arn:aws:iam::123456789012:role/" + args.Name)
	outputs["key"] = resource.NewStringProperty("test-key")
	outputs["privateKeyOpenssh"] = resource.NewStringProperty("private")
	outputs["publicKeyOpenssh"] = resource.NewStringProperty("public")
	outputs["rfc3339"] = resource.NewStringProperty("2026-01-01T00:00:00Z")
	outputs["latestVersion"] = resource.NewNumberProperty(1)
	return args.Name, outputs, nil
}

func TestBastionFederationResources(t *testing.T) {
	for _, mode := range []string{"federated", "oauth", "key"} {
		t.Run(mode, func(t *testing.T) {
			mocks := &bastionMocks{}
			err := pulumi.RunErr(func(ctx *pulumi.Context) error {
				args := &BastionArgs{
					VpcID: pulumi.String("vpc-123"), SubnetIds: pulumi.StringArray{pulumi.String("subnet-123")},
					Region: pulumi.String("us-west-2"), TailscaleTags: pulumi.StringArray{pulumi.String("tag:router")},
				}
				if mode == "federated" {
					args.WorkloadIdentity = &WorkloadIdentity{ClientID: pulumi.String("client"), Audience: pulumi.String("api.tailscale.com/client")}
					args.AcceptDNS = pulumi.Bool(false).ToBoolOutput()
					args.SNATSubnetRoutes = pulumi.Bool(false)
					args.RetryDelay = pulumi.Int(0).ToIntOutput()
					args.Track = pulumi.String("unstable").ToStringOutput()
				} else if mode == "oauth" {
					args.OauthClientSecret = pulumi.String("test-secret")
				}
				_, err := NewBastion(ctx, "router", args)
				return err
			}, pulumi.WithMocks("project", "stack", mocks))
			if err != nil {
				t.Fatal(err)
			}
			counts := map[string]int{}
			var policy string
			for _, r := range mocks.resources {
				counts[r.TypeToken]++
				if r.TypeToken == "aws:iam/rolePolicy:RolePolicy" {
					policy = r.Inputs["policy"].StringValue()
				}
				if r.TypeToken == "aws:ec2/launchTemplate:LaunchTemplate" {
					data, err := base64.StdEncoding.DecodeString(r.Inputs["userData"].StringValue())
					if err != nil {
						t.Fatal(err)
					}
					expected := []string{`TRACK="stable"`, `--accept-dns="true"`, `--snat-subnet-routes="true"`, "retry_delay=10"}
					if mode == "federated" {
						expected = []string{`TRACK="unstable"`, `--accept-dns="false"`, `--snat-subnet-routes="false"`, "retry_delay=0"}
					}
					for _, setting := range expected {
						if !strings.Contains(string(data), setting) {
							t.Errorf("launch template missing %s", setting)
						}
					}
				}
			}
			if mode == "federated" {
				if counts["aws:ssm/parameter:Parameter"] != 0 || counts["tailscale:index/tailnetKey:TailnetKey"] != 0 {
					t.Fatal("federation created stored credentials")
				}
				for _, expected := range []string{"sts:GetWebIdentityToken", "sts:IdentityTokenAudience", "api.tailscale.com/client", "sts:DurationSeconds", "300"} {
					if !strings.Contains(policy, expected) {
						t.Errorf("missing policy restriction %s", expected)
					}
				}
			} else {
				if counts["aws:ssm/parameter:Parameter"] != 1 {
					t.Fatal("legacy authentication must create SSM parameter")
				}
				if policy != "" {
					t.Fatal("legacy authentication should not grant token issuance")
				}
			}
		})
	}
}
