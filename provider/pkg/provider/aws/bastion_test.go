package aws

import (
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
