package aws

import (
	"bytes"
	_ "embed" // embed needs to be a blank import
	"encoding/base64"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/autoscaling"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ssm"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	//go:embed userdata.tmpl
	userData string
)

// The set of arguments for creating a Bastion component resource.
type BastionArgs struct {
	VpcID        pulumi.StringInput      `pulumi:"vpcId"`
	SubnetIds    pulumi.StringArrayInput `pulumi:"subnetIds"`
	Route        pulumi.StringInput      `pulumi:"route"`
	Region       pulumi.StringInput      `pulumi:"region"`
	InstanceType pulumi.StringInput      `pulumi:"instanceType"`
}

type UserDataArgs struct {
	ParameterName string
	Route         string
	Region        string
}

// The Bastion component resource.
type Bastion struct {
	pulumi.ResourceState

	AsgName pulumi.StringOutput `pulumi:"asgName"`
}

// NewBastion creates a new Bastion component resource.
func NewBastion(ctx *pulumi.Context,
	name string, args *BastionArgs, opts ...pulumi.ResourceOption) (*Bastion, error) {
	if args == nil {
		args = &BastionArgs{}
	}

	component := &Bastion{}

	err := ctx.RegisterComponentResource("aws-tailscale:index:Bastion", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// create a tailnet key to auth devices
	tailnetKey, err := tailscale.NewTailnetKey(ctx, name, &tailscale.TailnetKeyArgs{
		Ephemeral:     pulumi.Bool(true),
		Preauthorized: pulumi.Bool(true),
		Reusable:      pulumi.Bool(true),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating tailnet key: %v", err)
	}

	// store the key in an AWS SSM parameter
	tailnetKeySsmParameter, err := ssm.NewParameter(ctx, name, &ssm.ParameterArgs{
		Type:  ssm.ParameterTypeSecureString,
		Value: tailnetKey.Key,
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

	// Marshal the JSON into something the role can consume
	assumeRolePolicyJSON, err := json.Marshal(assumeRolePolicy)
	if err != nil {
		return nil, fmt.Errorf("error creating AssumeRolePolicy JSON: %v", err)
	}

	// create an IAM role the bastion uses
	// we give access to EC2 and SSM to read the parameter
	role, err := iam.NewRole(ctx, name, &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(string(assumeRolePolicyJSON)),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating IAM role: %v", err)
	}

	// FIXME: don't use the interface method here
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

	// allow access to the SSM parameter
	_, err = iam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-ssm-parameter", name), &iam.RolePolicyAttachmentArgs{
		Role:      role.Name,
		PolicyArn: ssmParameterPolicy.Arn,
	}, pulumi.Parent(role))
	if err != nil {
		return nil, fmt.Errorf("error creating SSM parameter policy attachment: %v", err)
	}

	// allow to EC2 instance to be managed by AWS SSM
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

	sg, err := ec2.NewSecurityGroup(ctx, name, &ec2.SecurityGroupArgs{
		VpcId: args.VpcID,
		Ingress: ec2.SecurityGroupIngressArray{
			ec2.SecurityGroupIngressArgs{
				Protocol: pulumi.String("icmp"),
				FromPort: pulumi.Int(0),
				ToPort:   pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{
					pulumi.String("0.0.0.0/0"),
				},
			},
		},
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
					pulumi.String("amzn2-ami-hvm*"),
				},
			},
		},
		MostRecent: pulumi.BoolPtr(true),
	}, pulumi.Parent(component))

	data := pulumi.All(tailnetKeySsmParameter.Name, args.Route, args.Region).ApplyT(
		func(args []interface{}) (string, error) {
			d := UserDataArgs{
				ParameterName: args[0].(string),
				Route:         args[1].(string),
				Region:        args[2].(string),
			}

			var userDataBytes bytes.Buffer

			userDataTemplate := template.New("userdata")
			userDataTemplate, err = userDataTemplate.Parse(userData)
			if err != nil {
				return "", err
			}
			err := userDataTemplate.Execute(&userDataBytes, d)
			if err != nil {
				return "", err
			}

			return base64.StdEncoding.EncodeToString(userDataBytes.Bytes()), nil

		},
	).(pulumi.StringOutput)

	var instanceType pulumi.String
	if args.InstanceType == nil {
		instanceType = pulumi.String("t3.micro")
	} else {
		instanceType = args.InstanceType.(pulumi.String)
	}

	launchConfiguration, err := ec2.NewLaunchConfiguration(ctx, name, &ec2.LaunchConfigurationArgs{
		InstanceType:             instanceType,
		AssociatePublicIpAddress: pulumi.Bool(false),
		ImageId:                  ami.Id(),
		SecurityGroups: pulumi.StringArray{
			sg.ID(),
		},
		IamInstanceProfile: profile.ID(),
		UserDataBase64:     data,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating launch configuration: %v", err)
	}

	asg, err := autoscaling.NewGroup(ctx, name, &autoscaling.GroupArgs{
		LaunchConfiguration:    launchConfiguration.ID(),
		MaxSize:                pulumi.Int(1),
		MinSize:                pulumi.Int(1),
		HealthCheckType:        pulumi.String("EC2"),
		HealthCheckGracePeriod: pulumi.Int(30),
		VpcZoneIdentifiers:     args.SubnetIds,
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

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"asgName": asg.Name,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
