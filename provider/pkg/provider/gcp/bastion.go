package gcp

import (
	"bytes"
	_ "embed" // embed needs to be a blank import
	"encoding/base64"
	"fmt"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	tls "github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"text/template"
)

var (
	//go:embed userdata.tmpl
	userData string
)

// The set of arguments for creating a Bastion component resource.
type BastionArgs struct {
	SubNetworkID pulumi.StringInput `pulumi:"subnetNetworkId"`
	Route        pulumi.StringInput `pulumi:"route"`
	MachineType  pulumi.StringInput `pulumi:"machineType"`
}

type UserDataArgs struct {
	AuthKey string
	Route   string
}

// The Bastion component resource.
type Bastion struct {
	pulumi.ResourceState

	AutoscalerName   pulumi.StringOutput `pulumi:"autoscalerName"`
	GroupManagerName pulumi.StringOutput `pulumi:"groupManagerName"`
	TargetPoolName   pulumi.StringOutput `pulumi:"targetPoolName"`
	PrivateKey       pulumi.StringOutput `pulumi:"privateKey"`
}

// NewBastion creates a new Bastion component resource.
func NewBastion(ctx *pulumi.Context,
	name string, args *BastionArgs, opts ...pulumi.ResourceOption) (*Bastion, error) {
	if args == nil {
		args = &BastionArgs{}
	}

	component := &Bastion{}

	err := ctx.RegisterComponentResource("tailscale-bastion:gcp:Bastion", name, component, opts...)
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

	image, err := compute.LookupImage(ctx, &compute.LookupImageArgs{
		Family:  pulumi.StringRef("ubuntu-2004-lts"),
		Project: pulumi.StringRef("ubuntu-os-cloud"),
	}, pulumi.Parent((component)))
	if err != nil {
		return nil, fmt.Errorf("error looking up image: %v", err)
	}

	data := pulumi.All(tailnetKey.Key, args.Route).ApplyT(
		func(args []interface{}) (string, error) {
			d := UserDataArgs{
				AuthKey: args[0].(string),
				Route:   args[1].(string),
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

	key, err := tls.NewPrivateKey(ctx, name, &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.Int(4096),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	var machineType pulumi.String
	if args.MachineType == nil {
		machineType = pulumi.String("f1-micro")
	} else {
		machineType = args.MachineType.(pulumi.String)
	}

	template, err := compute.NewInstanceTemplate(ctx, name, &compute.InstanceTemplateArgs{
		MachineType:  machineType,
		CanIpForward: pulumi.Bool(true),
		Disks: compute.InstanceTemplateDiskArray{
			&compute.InstanceTemplateDiskArgs{
				SourceImage: pulumi.String(image.SelfLink),
				AutoDelete:  pulumi.Bool(true),
				Boot:        pulumi.Bool(true),
			},
		},
		NetworkInterfaces: compute.InstanceTemplateNetworkInterfaceArray{
			&compute.InstanceTemplateNetworkInterfaceArgs{
				Subnetwork: args.SubNetworkID,
			},
		},
		Metadata: pulumi.Map{
			"user-data": data,
			"ssh-keys":  pulumi.Sprintf("bastion:%s", key.PublicKeyOpenssh),
		},
	}, pulumi.Parent(component))

	targetPool, err := compute.NewTargetPool((ctx), name, &compute.TargetPoolArgs{}, pulumi.Parent(component))

	groupManager, err := compute.NewInstanceGroupManager(ctx, name, &compute.InstanceGroupManagerArgs{
		Versions: compute.InstanceGroupManagerVersionArray{
			&compute.InstanceGroupManagerVersionArgs{
				InstanceTemplate: template.SelfLink,
			},
		},
		TargetPools: pulumi.StringArray{
			targetPool.SelfLink,
		},
		BaseInstanceName: pulumi.String(fmt.Sprintf("%s-%s", name, "bastion")),
	}, pulumi.Parent(template))

	autoscaler, err := compute.NewAutoscalar(ctx, name, &compute.AutoscalarArgs{
		Target: groupManager.ID(),
		AutoscalingPolicy: &compute.AutoscalarAutoscalingPolicyArgs{
			MaxReplicas: pulumi.Int(1),
			MinReplicas: pulumi.Int(1),
		},
	}, pulumi.Parent(groupManager))

	component.AutoscalerName = autoscaler.Name
	component.GroupManagerName = groupManager.Name
	component.TargetPoolName = targetPool.Name
	component.PrivateKey = key.PrivateKeyOpenssh

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"autoscalerName":   autoscaler.Name,
		"groupManagerName": groupManager.Name,
		"targetPoolName":   targetPool.Name,
		"privateKey":       key.PrivateKeyOpenssh,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
