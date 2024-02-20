package azure

import (
	"bytes"
	_ "embed" // embed needs to be a blank import
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"github.com/pulumi/pulumi-azure/sdk/v5/go/azure/compute"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	tls "github.com/pulumi/pulumi-tls/sdk/v4/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	//go:embed userdata.tmpl
	userData string
)

// The set of arguments for creating a Bastion component resource.
type BastionArgs struct {
	ResourceGroupName pulumi.StringInput      `pulumi:"resourceGroupName"`
	SubnetID          pulumi.StringInput      `pulumi:"subnetId"`
	Location          pulumi.StringInput      `pulumi:"location"`
	Route             pulumi.StringInput      `pulumi:"route"`
	InstanceSku       pulumi.StringInput      `pulumi:"instanceSku"`
	TailscaleTags     pulumi.StringArrayInput `pulumi:"tailscaleTags"`
	HighAvailability  bool                    `pulumi:"highAvailability"`
}

type UserDataArgs struct {
	AuthKey       string
	Route         string
	TailscaleTags []string
}

// Join the tags into a CSV
func (uda *UserDataArgs) JoinedTags() string {
	return strings.Join(uda.TailscaleTags, ",")
}

// The Bastion component resource.
type Bastion struct {
	pulumi.ResourceState

	ScaleSetName pulumi.StringOutput `pulumi:"scaleSetName"`
	PrivateKey   pulumi.StringOutput `pulumi:"privateKey"`
}

// NewBastion creates a new Bastion component resource.
func NewBastion(ctx *pulumi.Context,
	name string, args *BastionArgs, opts ...pulumi.ResourceOption) (*Bastion, error) {
	if args == nil {
		args = &BastionArgs{}
	}

	component := &Bastion{}

	err := ctx.RegisterComponentResource("tailscale-bastion:azure:Bastion", name, component, opts...)
	if err != nil {
		return nil, err
	}

	// create a tailnet key to auth devices
	tailnetKey, err := tailscale.NewTailnetKey(ctx, name, &tailscale.TailnetKeyArgs{
		Ephemeral:     pulumi.Bool(true),
		Preauthorized: pulumi.Bool(true),
		Reusable:      pulumi.Bool(true),
		Tags:          args.TailscaleTags,
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating tailnet key: %v", err)
	}

	data := pulumi.All(tailnetKey.Key, args.Route, args.TailscaleTags).ApplyT(
		func(args []interface{}) (string, error) {
			d := UserDataArgs{
				AuthKey:       args[0].(string),
				Route:         args[1].(string),
				TailscaleTags: args[2].([]string),
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

	var sku pulumi.String
	if args.InstanceSku == nil {
		sku = pulumi.String("Standard_B1s")
	} else {
		sku = args.InstanceSku.(pulumi.String)
	}

	key, err := tls.NewPrivateKey(ctx, name, &tls.PrivateKeyArgs{
		Algorithm: pulumi.String("RSA"),
		RsaBits:   pulumi.Int(4096),
	}, pulumi.Parent(component))
	if err != nil {
		return nil, err
	}

	var size int

	if args.HighAvailability {
		size = 2
	} else {
		size = 1
	}

	scaleset, err := compute.NewLinuxVirtualMachineScaleSet(ctx, name, &compute.LinuxVirtualMachineScaleSetArgs{
		ResourceGroupName: args.ResourceGroupName,
		Location:          args.Location,
		UpgradeMode:       pulumi.String("Manual"),
		Sku:               sku,
		Instances:         pulumi.Int(size),
		SourceImageReference: &compute.LinuxVirtualMachineScaleSetSourceImageReferenceArgs{
			Publisher: pulumi.String("Canonical"),
			Offer:     pulumi.String("0001-com-ubuntu-server-focal"),
			Sku:       pulumi.String("20_04-lts-gen2"),
			Version:   pulumi.String("latest"),
		},
		AdminUsername: pulumi.String(name),
		AdminSshKeys: compute.LinuxVirtualMachineScaleSetAdminSshKeyArray{
			&compute.LinuxVirtualMachineScaleSetAdminSshKeyArgs{
				PublicKey: key.PublicKeyOpenssh,
				Username:  pulumi.String(name),
			},
		},
		Identity: &compute.LinuxVirtualMachineScaleSetIdentityArgs{
			Type: pulumi.String("SystemAssigned"),
		},
		CustomData: data,
		OsDisk: &compute.LinuxVirtualMachineScaleSetOsDiskArgs{
			StorageAccountType: pulumi.String("Standard_LRS"),
			Caching:            pulumi.String("ReadWrite"),
		},
		NetworkInterfaces: compute.LinuxVirtualMachineScaleSetNetworkInterfaceArray{
			&compute.LinuxVirtualMachineScaleSetNetworkInterfaceArgs{
				Name:               pulumi.String("eth0"),
				Primary:            pulumi.Bool(true),
				EnableIpForwarding: pulumi.Bool(true),
				IpConfigurations: &compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationArray{
					&compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationArgs{
						Name:     pulumi.String("internal"),
						Primary:  pulumi.Bool(true),
						SubnetId: args.SubnetID,
					},
				},
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating scale set: %v", err)
	}

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"scaleSetName": scaleset.Name,
		"privateKey":   key.PrivateKeyOpenssh,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
