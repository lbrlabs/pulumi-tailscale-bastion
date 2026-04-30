package aws

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ssm"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	tls "github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	//go:embed userdata.tmpl
	userData string
)

type Architecture string

const (
	ArchX86_64 Architecture = "x86_64"
	ArchArm64  Architecture = "arm64"
)

type PeerRelaySettings struct {
	Enable bool `pulumi:"enable"`
	Port   int  `pulumi:"port"`
}

type BastionArgs struct {
	VpcID              pulumi.StringInput      `pulumi:"vpcId"`
	SubnetIds          pulumi.StringArrayInput `pulumi:"subnetIds"`
	TailscaleTags      pulumi.StringArrayInput `pulumi:"tailscaleTags"`
	Routes             pulumi.StringArrayInput `pulumi:"routes"`
	Region             pulumi.StringInput      `pulumi:"region"`
	InstanceType       pulumi.StringInput      `pulumi:"instanceType"`
	Hostname           pulumi.StringInput      `pulumi:"hostname"`
	OauthClientSecret  pulumi.StringInput      `pulumi:"oauthClientSecret"`
	HighAvailability   bool                    `pulumi:"highAvailability"`
	EnableSSH          bool                    `pulumi:"enableSSH"`
	Public             bool                    `pulumi:"public"`
	EnableExitNode     bool                    `pulumi:"enableExitNode"`
	EnableAppConnector bool                    `pulumi:"enableAppConnector"`
	Architecture       pulumi.StringInput      `pulumi:"architecture"`
	PeerRelaySettings  *PeerRelaySettings      `pulumi:"peerRelaySettings"`
}

type UserDataArgs struct {
	ParameterName      string
	Routes             string
	Region             string
	Partition          string
	TailscaleTags      string
	EnableSSH          bool
	EnableExitNode     bool
	EnableAppConnector bool
	Hostname           string
	PeerRelayEnable    bool
	PeerRelayPort      int
}

type Bastion struct {
	pulumi.ResourceState

	AsgName    pulumi.StringOutput `pulumi:"asgName"`
	PrivateKey pulumi.StringOutput `pulumi:"privateKey"`
}

func NewBastion(ctx *pulumi.Context,
	name string, args *BastionArgs, opts ...pulumi.ResourceOption) (*Bastion, error) {

	if args == nil {
		args = &BastionArgs{}
	}

	component := &Bastion{}

	err := ctx.RegisterComponentResource("tailscale-bastion:aws:Bastion", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// Get partition information for the current AWS region
	partition := aws.GetPartitionOutput(ctx, aws.GetPartitionOutputArgs{}, pulumi.Parent(component))

	var hostname pulumi.StringInput
	if args.Hostname == nil {
		hostname = pulumi.String(name)
	} else {
		hostname = args.Hostname
	}

	var arch pulumi.StringInput
	if args.Architecture == nil {
		arch = pulumi.String(string(ArchX86_64))
	} else {
		arch = args.Architecture
	}

	var tailnetKeyToUseForAuth pulumi.StringInput

	if args.OauthClientSecret == nil {
		tailnetKey, err := tailscale.NewTailnetKey(ctx, name, &tailscale.TailnetKeyArgs{
			Ephemeral:     pulumi.Bool(true),
			Preauthorized: pulumi.Bool(true),
			Reusable:      pulumi.Bool(true),
			Tags:          args.TailscaleTags,
			Description:   pulumi.Sprintf("Auth key for %s", hostname),
		}, pulumi.Parent(component))
		if err != nil {
			return nil, fmt.Errorf("error creating tailnet key: %v", err)
		}
		tailnetKeyToUseForAuth = tailnetKey.Key
	} else {
		tailnetKeyToUseForAuth = pulumi.Sprintf("%s?ephemeral=true&preauthorized=true", args.OauthClientSecret)
	}

	tailnetKeySsmParameter, err := ssm.NewParameter(ctx, name, &ssm.ParameterArgs{
		Type:        ssm.ParameterTypeSecureString,
		Value:       tailnetKeyToUseForAuth,
		Description: pulumi.Sprintf("Tailscale auth key for %s", hostname),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating SSM parameter: %v", err)
	}

	role, err := iam.NewRole(ctx, name, &iam.RoleArgs{
		AssumeRolePolicy: partition.DnsSuffix().ApplyT(func(dnsSuffix string) (string, error) {
			assumeRolePolicy := &AssumeRolePolicy{
				Version: "2012-10-17",
				Statement: []Statement{
					{
						Effect: "Allow",
						Action: "sts:AssumeRole",
						Principal: Principal{
							Service: []string{
								fmt.Sprintf("ec2.%s", dnsSuffix),
								fmt.Sprintf("ssm.%s", dnsSuffix),
							},
						},
					},
				},
			}
			assumeRolePolicyJSON, err := json.Marshal(assumeRolePolicy)
			if err != nil {
				return "", err
			}
			return string(assumeRolePolicyJSON), nil
		}).(pulumi.StringOutput),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating IAM role: %v", err)
	}

	ssmParameterPolicyJSON := tailnetKeySsmParameter.Arn.ApplyT(func(arn string) (string, error) {
		policyJSON, err := json.Marshal(map[string]interface{}{
			"Version": "2012-10-17",
			"Statement": []interface{}{
				map[string]interface{}{
					"Action": []string{
						"ssm:GetParameters",
					},
					"Effect": "Allow",
					"Resource": []string{
						arn,
					},
				},
				map[string]interface{}{
					"Action": []string{
						"ssm:DescribeParameters",
					},
					"Effect":   "Allow",
					"Resource": "*",
				},
			},
		})
		if err != nil {
			return "", err
		}
		return string(policyJSON), nil
	})

	ssmParameterPolicy, err := iam.NewPolicy(ctx, name, &iam.PolicyArgs{
		Policy: ssmParameterPolicyJSON,
	}, pulumi.Parent(role))
	if err != nil {
		return nil, fmt.Errorf("error creating SSM parameter policy: %v", err)
	}

	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-ssm-parameter", name), &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: ssmParameterPolicy.Arn,
	}, pulumi.Parent(role))
	if err != nil {
		return nil, fmt.Errorf("error creating SSM parameter policy attachment: %v", err)
	}

	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-ssm-manager", name), &iam.RolePolicyAttachmentArgs{
		Role: role.Name,
		PolicyArn: partition.Partition().ApplyT(func(partitionName string) string {
			return fmt.Sprintf("arn:%s:iam::aws:policy/AmazonSSMManagedInstanceCore", partitionName)
		}).(pulumi.StringOutput),
	}, pulumi.Parent(role))
	if err != nil {
		return nil, fmt.Errorf("error creating SSM manager policy attachment: %v", err)
	}

	profile, err := iam.NewInstanceProfile(ctx, name, &iam.InstanceProfileArgs{
		Role: role.Name,
	}, pulumi.Parent(role))
	if err != nil {
		return nil, fmt.Errorf("error creating IAM instance profile: %v", err)
	}

	// Default peer relay settings if not provided
	peerRelayEnable := false
	peerRelayPort := 12345
	if args.PeerRelaySettings != nil {
		if args.PeerRelaySettings.Enable && !args.Public {
			return nil, fmt.Errorf("peer relay can only be enabled when public=true")
		}
		peerRelayEnable = args.PeerRelaySettings.Enable
		peerRelayPort = args.PeerRelaySettings.Port
	}

	var ingress ec2.SecurityGroupIngressArray
	if args.Public {
		ingress = ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("icmp"),
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("udp"),
				FromPort: pulumi.Int(41641),
				ToPort:   pulumi.Int(41641),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		}

		// Add peer relay port if enabled
		if peerRelayEnable {
			ingress = append(ingress, ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("udp"),
				FromPort: pulumi.Int(peerRelayPort),
				ToPort:   pulumi.Int(peerRelayPort),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			})
		}
	} else {
		ingress = ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("icmp"),
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		}
	}

	sg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId:   args.VpcID,
		Ingress: ingress,
		Egress: ec2.SecurityGroupEgressArray{
			ec2.SecurityGroupEgressArgs{
				Protocol: pulumi.String("-1"),
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating security group: %v", err)
	}

	amiParameter := pulumi.All(arch).ApplyT(func(v []interface{}) (string, error) {
		switch v[0].(string) {
		case string(ArchArm64):
			return "resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-arm64", nil
		case string(ArchX86_64):
			return "resolve:ssm:/aws/service/ami-amazon-linux-latest/al2023-ami-kernel-default-x86_64", nil
		default:
			return "", fmt.Errorf("unsupported architecture: %s", v[0].(string))
		}
	}).(pulumi.StringOutput)

	data := pulumi.All(tailnetKeySsmParameter.Name, args.Routes, args.Region, args.TailscaleTags, args.EnableSSH, hostname, args.EnableExitNode, args.EnableAppConnector, pulumi.Bool(peerRelayEnable), pulumi.Int(peerRelayPort), partition.DnsSuffix()).ApplyT(
		func(args []interface{}) (string, error) {
			tagCSV := strings.Join(args[3].([]string), ",")

			var routesCsv string
			if args[1] != nil {
				routes := args[1].([]string)
				routesCsv = strings.Join(routes, ",")
			} else {
				routesCsv = ""
			}

			d := UserDataArgs{
				ParameterName:      args[0].(string),
				Routes:             routesCsv,
				Region:             args[2].(string),
				Partition:          args[10].(string),
				TailscaleTags:      tagCSV,
				EnableSSH:          args[4].(bool),
				Hostname:           args[5].(string),
				EnableExitNode:     args[6].(bool),
				EnableAppConnector: args[7].(bool),
				PeerRelayEnable:    args[8].(bool),
				PeerRelayPort:      args[9].(int),
			}

			var userDataBytes bytes.Buffer
			userDataTemplate := template.New("userdata")
			userDataTemplate, err = userDataTemplate.Parse(userData)
			if err != nil {
				return "", err
			}
			err = userDataTemplate.Execute(&userDataBytes, d)
			if err != nil {
				return "", err
			}
			return base64.StdEncoding.EncodeToString(userDataBytes.Bytes()), nil
		},
	).(pulumi.StringOutput)

	var instanceType pulumi.StringInput
	if args.InstanceType == nil {
		instanceType = pulumi.All(arch).ApplyT(func(v []interface{}) string {
			if v[0].(string) == string(ArchArm64) {
				return "t4g.micro"
			}
			return "t3.micro"
		}).(pulumi.StringOutput)
	} else {
		instanceType = args.InstanceType
	}

	key, err := tls.NewPrivateKey(ctx, name, &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.Int(4096),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	ec2Key, err := ec2.NewKeyPair(ctx, name, &ec2.KeyPairArgs{
		PublicKey: key.PublicKeyOpenssh,
	}, pulumi.Parent(component))

	launchTemplate, err := ec2.NewLaunchTemplate(ctx, name, &ec2.LaunchTemplateArgs{
		Description:          pulumi.String(fmt.Sprintf("Updated by Pulumi at %s", time.Now().UTC().Format(time.RFC3339))),
		InstanceType:         instanceType,
		ImageId:              amiParameter,
		UpdateDefaultVersion: pulumi.Bool(true),
		KeyName:              ec2Key.KeyName,
		IamInstanceProfile: ec2.LaunchTemplateIamInstanceProfileArgs{
			Arn: profile.Arn,
		},
		UserData: data,
		VpcSecurityGroupIds: pulumi.StringArray{
			sg.ID(),
		},
		TagSpecifications: ec2.LaunchTemplateTagSpecificationArray{
			ec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String("instance"),
				Tags: pulumi.StringMap{
					"Name": pulumi.String(fmt.Sprintf("%s-tailscale-bastion", name)),
				},
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating launch template: %v", err)
	}

	var size int
	if args.HighAvailability {
		size = 2
	} else {
		size = 1
	}

	minHealthyPercentage := 0
	if args.HighAvailability {
		minHealthyPercentage = 50
	}

	instanceRefresh := autoscaling.GroupInstanceRefreshArgs{
		Strategy: pulumi.String("Rolling"),
		Preferences: autoscaling.GroupInstanceRefreshPreferencesArgs{
			MinHealthyPercentage: pulumi.Int(minHealthyPercentage),
		},
	}

	asg, err := autoscaling.NewGroup(ctx, name, &autoscaling.GroupArgs{
		LaunchTemplate: autoscaling.GroupLaunchTemplateArgs{
			Id:      launchTemplate.ID(),
			Version: pulumi.Sprintf("%d", launchTemplate.LatestVersion),
		},
		MaxSize:                pulumi.Int(size),
		MinSize:                pulumi.Int(size),
		HealthCheckType:        pulumi.String("EC2"),
		HealthCheckGracePeriod: pulumi.Int(30),
		VpcZoneIdentifiers:     args.SubnetIds,
		InstanceRefresh:        instanceRefresh,
		Tags: autoscaling.GroupTagArray{
			autoscaling.GroupTagArgs{
				Key:               pulumi.String("Name"),
				Value:             pulumi.String(fmt.Sprintf("%s-tailscale-bastion", name)),
				PropagateAtLaunch: pulumi.Bool(true),
			},
		},
	}, pulumi.Parent(launchTemplate))
	if err != nil {
		return nil, fmt.Errorf("error creating asg: %v", err)
	}

	component.AsgName = asg.Name
	component.PrivateKey = key.PrivateKeyOpenssh

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"asgName":    asg.Name,
		"privateKey": key.PrivateKeyOpenssh,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
