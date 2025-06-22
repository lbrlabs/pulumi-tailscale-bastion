package aws

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ssm"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	tls "github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strings"
	"text/template"
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
}

type UserDataArgs struct {
	ParameterName      string
	Routes             string
	Region             string
	TailscaleTags      string
	EnableSSH          bool
	EnableExitNode     bool
	EnableAppConnector bool
	Hostname           string
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

	assumeRolePolicy := &AssumeRolePolicy{
		Version: "2012-10-17",
		Statement: []Statement{
			{
				Effect: "Allow",
				Action: "sts:AssumeRole",
				Principal: Principal{
					Service: []string{"ec2.amazonaws.com", "ssm.amazonaws.com"},
				},
			},
		},
	}

	assumeRolePolicyJSON, err := json.Marshal(assumeRolePolicy)
	if err != nil {
		return nil, fmt.Errorf("error creating AssumeRolePolicy JSON: %v", err)
	}

	role, err := iam.NewRole(ctx, name, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(string(assumeRolePolicyJSON)),
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
		Role:      role.Name,
		PolicyArn: iam.ManagedPolicyAmazonSSMManagedInstanceCore,
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

	ami := ec2.LookupAmiOutput(ctx, ec2.LookupAmiOutputArgs{
		Owners: pulumi.StringArray{
			pulumi.String("amazon"),
		},
		Filters: ec2.GetAmiFilterArray{
			ec2.GetAmiFilterArgs{
				Name: pulumi.String("owner-alias"),
				Values: pulumi.StringArray{
					pulumi.String("amazon"),
				},
			},
			ec2.GetAmiFilterArgs{
				Name: pulumi.String("virtualization-type"),
				Values: pulumi.StringArray{
					pulumi.String("hvm"),
				},
			},
			ec2.GetAmiFilterArgs{
				Name: pulumi.String("name"),
				Values: pulumi.StringArray{
					pulumi.Sprintf("al2023-ami-*%s*", arch),
				},
			},
		},
		MostRecent: pulumi.BoolPtr(true),
	}, pulumi.Parent(component))

	data := pulumi.All(tailnetKeySsmParameter.Name, args.Routes, args.Region, args.TailscaleTags, args.EnableSSH, hostname, args.EnableExitNode, args.EnableAppConnector).ApplyT(
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
				TailscaleTags:      tagCSV,
				EnableSSH:          args[4].(bool),
				Hostname:           args[5].(string),
				EnableExitNode:     args[6].(bool),
				EnableAppConnector: args[7].(bool),
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

	launchConfiguration, err := ec2.NewLaunchConfiguration(ctx, name, &ec2.LaunchConfigurationArgs{
		InstanceType:             instanceType,
		AssociatePublicIpAddress: pulumi.Bool(args.Public),
		ImageId:                  ami.Id(),
		SecurityGroups: pulumi.StringArray{
			sg.ID(),
		},
		KeyName:            ec2Key.KeyName,
		IamInstanceProfile: profile.ID(),
		UserDataBase64:     data,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating launch configuration: %v", err)
	}

	var size int
	if args.HighAvailability {
		size = 2
	} else {
		size = 1
	}

	var instanceRefresh autoscaling.GroupInstanceRefreshPtrInput
	if args.HighAvailability {
		instanceRefresh = autoscaling.GroupInstanceRefreshArgs{
			Strategy: pulumi.String("Rolling"),
			Preferences: autoscaling.GroupInstanceRefreshPreferencesArgs{
				MinHealthyPercentage: pulumi.Int(50),
			},
		}
	} else {
		instanceRefresh = nil
	}

	asg, err := autoscaling.NewGroup(ctx, name, &autoscaling.GroupArgs{
		LaunchConfiguration:    launchConfiguration.ID(),
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
	}, pulumi.Parent(launchConfiguration))
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
