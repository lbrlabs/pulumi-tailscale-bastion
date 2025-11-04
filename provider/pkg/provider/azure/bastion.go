package azure

import (
	"bytes"
	_ "embed" // embed needs to be a blank import
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"

	"github.com/pulumi/pulumi-azure/sdk/v6/go/azure/compute"
	"github.com/pulumi/pulumi-tailscale/sdk/go/tailscale"
	tls "github.com/pulumi/pulumi-tls/sdk/v5/go/tls"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var (
	//go:embed userdata.tmpl
	userData string
)

type PeerRelaySettings struct {
	Enable bool `pulumi:"enable"`
	Port   int  `pulumi:"port"`
}

// The set of arguments for creating a Bastion component resource.
type BastionArgs struct {
	ResourceGroupName  pulumi.StringInput      `pulumi:"resourceGroupName"`
	SubnetID           pulumi.StringInput      `pulumi:"subnetId"`
	Location           pulumi.StringInput      `pulumi:"location"`
	Routes             pulumi.StringArrayInput `pulumi:"routes"`
	InstanceSku        pulumi.StringInput      `pulumi:"instanceSku"`
	TailscaleTags      pulumi.StringArrayInput `pulumi:"tailscaleTags"`
	Hostname           pulumi.StringInput      `pulumi:"hostname"`
	HighAvailability   bool                    `pulumi:"highAvailability"`
	EnableSSH          bool                    `pulumi:"enableSSH"`
	EnableExitNode     bool                    `pulumi:"enableExitNode"`
	EnableAppConnector bool                    `pulumi:"enableAppConnector"`
	Public             bool                    `pulumi:"public"`
	PeerRelaySettings  *PeerRelaySettings      `pulumi:"peerRelaySettings"`
}

type UserDataArgs struct {
	AuthKey            string
	Routes             string
	TailscaleTags      string
	EnableSSH          bool
	EnableExitNode     bool
	EnableAppConnector bool
	Hostname           string
	PeerRelayEnable    bool
	PeerRelayPort      int
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

	var hostname pulumi.StringInput

	if args.Hostname == nil {
		hostname = pulumi.String(name)
	} else {
		hostname = args.Hostname
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

	// Default peer relay settings if not provided
	peerRelayEnable := false
	peerRelayPort := 12345
	if args.PeerRelaySettings != nil {
		if args.PeerRelaySettings.Enable && !args.Public {
			return nil, fmt.Errorf("peer relay can only be enabled when public=true")
		}
		peerRelayEnable = args.PeerRelaySettings.Enable
		peerRelayPort = args.PeerRelaySettings.Port

		// Note: When using peer relay in Azure, ensure your subnet's Network Security Group
		// allows inbound TCP traffic on the specified peer relay port (%d) from 0.0.0.0/0
		if peerRelayEnable {
			fmt.Printf("WARNING: Peer relay enabled on port %d. Ensure your subnet's Network Security Group allows inbound TCP traffic on port %d from 0.0.0.0/0\n", peerRelayPort, peerRelayPort)
		}
	}

	data := pulumi.All(tailnetKey.Key, args.Routes, args.TailscaleTags, args.EnableSSH, hostname, args.EnableExitNode, args.EnableAppConnector, pulumi.Bool(peerRelayEnable), pulumi.Int(peerRelayPort)).ApplyT(
		func(args []interface{}) (string, error) {

			tagCSV := strings.Join(args[2].([]string), ",")

			var routesCsv string

			if args[1] != nil {
				routes := args[1].([]string)
				routesCsv = strings.Join(routes, ",")
			} else {
				routesCsv = ""
			}

			d := UserDataArgs{
				AuthKey:            args[0].(string),
				Routes:             routesCsv,
				TailscaleTags:      tagCSV,
				EnableSSH:          args[3].(bool),
				Hostname:           args[4].(string),
				EnableExitNode:     args[5].(bool),
				EnableAppConnector: args[6].(bool),
				PeerRelayEnable:    args[7].(bool),
				PeerRelayPort:      args[8].(int),
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

	var publicIPConfig compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArray

	if args.Public {

		publicIPConfig = compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArray{
			&compute.LinuxVirtualMachineScaleSetNetworkInterfaceIpConfigurationPublicIpAddressArgs{
				Name: pulumi.String("public"),
			},
		}
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
						Name:              pulumi.String("internal"),
						Primary:           pulumi.Bool(true),
						SubnetId:          args.SubnetID,
						PublicIpAddresses: publicIPConfig,
					},
				},
			},
		},
	}, pulumi.Parent(component))
	if err != nil {
		return nil, fmt.Errorf("error creating scale set: %v", err)
	}

	component.PrivateKey = key.PrivateKeyOpenssh
	component.ScaleSetName = scaleset.Name

	if err := ctx.RegisterResourceOutputs(component, pulumi.Map{
		"scaleSetName": scaleset.Name,
		"privateKey":   key.PrivateKeyOpenssh,
	}); err != nil {
		return nil, err
	}

	return component, nil
}
