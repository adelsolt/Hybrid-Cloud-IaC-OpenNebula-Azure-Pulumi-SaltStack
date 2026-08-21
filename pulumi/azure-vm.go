package main

import (
	"encoding/base64"
	"os"
	"strings"

	"github.com/pulumi/pulumi-azure-native-sdk/compute/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/network/v2"
	"github.com/pulumi/pulumi-azure-native-sdk/resources/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func buildAzure(ctx *pulumi.Context) error {
	cfg := config.New(ctx, "")
	sshKey := cfg.Require("sshPublicKey")
	masterPub := cfg.Require("masterAddrPublic")   
	adminIP := cfg.Require("adminIP")            

	// 1. Resource group: 
	rg, err := resources.NewResourceGroup(ctx, "edge-rg", nil)
	if err != nil { return err }

	// 2. Virtual network + subnet 
	vnet, err := network.NewVirtualNetwork(ctx, "edge-vnet", &network.VirtualNetworkArgs{
		ResourceGroupName: rg.Name,
		AddressSpace: &network.AddressSpaceArgs{
			AddressPrefixes: pulumi.StringArray{pulumi.String("10.20.0.0/16")},
		},
	})
	if err != nil { return err }

	subnet, err := network.NewSubnet(ctx, "edge-subnet", &network.SubnetArgs{
		ResourceGroupName:  rg.Name,
		VirtualNetworkName: vnet.Name,
		AddressPrefix:      pulumi.String("10.20.1.0/24"),
	})
	if err != nil { return err }

		// 3. NSG — Azure's firewall. SSH from you, HTTPS from anyone, WG UDP from OPNsense.
	nsg, err := network.NewNetworkSecurityGroup(ctx, "edge-nsg", &network.NetworkSecurityGroupArgs{
		ResourceGroupName: rg.Name,
		SecurityRules: network.SecurityRuleTypeArray{
			&network.SecurityRuleTypeArgs{
				Name: pulumi.String("ssh"), Priority: pulumi.Int(100),
				Direction: pulumi.String("Inbound"), Access: pulumi.String("Allow"),
				Protocol: pulumi.String("Tcp"), SourceAddressPrefix: pulumi.String(adminIP),
				SourcePortRange: pulumi.String("*"), DestinationAddressPrefix: pulumi.String("*"),
				DestinationPortRange: pulumi.String("22"),
			},
			&network.SecurityRuleTypeArgs{
				Name: pulumi.String("https"), Priority: pulumi.Int(110),
				Direction: pulumi.String("Inbound"), Access: pulumi.String("Allow"),
				Protocol: pulumi.String("Tcp"), SourceAddressPrefix: pulumi.String("*"),
				SourcePortRange: pulumi.String("*"), DestinationAddressPrefix: pulumi.String("*"),
				DestinationPortRange: pulumi.String("443"),
			},
			&network.SecurityRuleTypeArgs{
				Name: pulumi.String("wireguard"), Priority: pulumi.Int(120),
				Direction: pulumi.String("Inbound"), Access: pulumi.String("Allow"),
				Protocol: pulumi.String("Udp"), SourceAddressPrefix: pulumi.String(masterPub),
				SourcePortRange: pulumi.String("*"), DestinationAddressPrefix: pulumi.String("*"),
				DestinationPortRange: pulumi.String("51820"),
			},
		},
	})
	if err != nil { return err }

	// 4. Public IP the edge's internet address
	pip, err := network.NewPublicIPAddress(ctx, "edge-pip", &network.PublicIPAddressArgs{
    ResourceGroupName:        rg.Name,
    Location:                 rg.Location,
    PublicIPAllocationMethod: pulumi.String("Static"),
    Sku: &network.PublicIPAddressSkuArgs{
        Name: pulumi.String("Standard"),
    },
    })
	if err != nil { return err }

	// 5. NIC 
	nic, err := network.NewNetworkInterface(ctx, "edge-nic", &network.NetworkInterfaceArgs{
		ResourceGroupName: rg.Name,
		NetworkSecurityGroup: &network.NetworkSecurityGroupTypeArgs{Id: nsg.ID()},
		IpConfigurations: network.NetworkInterfaceIPConfigurationArray{
			&network.NetworkInterfaceIPConfigurationArgs{
				Name:                      pulumi.String("ipcfg"),
				Subnet:                    &network.SubnetTypeArgs{Id: subnet.ID()},
				PublicIPAddress:           &network.PublicIPAddressTypeArgs{Id: pip.ID()},
				PrivateIPAllocationMethod: pulumi.String("Dynamic"),
			},
		},
	})
	if err != nil { return err }

	// bootstrap same template as OpenNebula, edge role, PUBLIC master addr
	raw, err := os.ReadFile("bootstrap/minion.sh.tmpl")
	if err != nil { return err }
	script := strings.NewReplacer(
		"__MASTER_ADDR__", masterPub,
		"__MINION_ID__", "edge-az-01",
		"__ROLE__", "edge",
		"__CLOUD__", "azure",
	).Replace(string(raw))
	customData := base64.StdEncoding.EncodeToString([]byte(script))

	// 6.  Ubuntu, key, cloud-init runs the bootstrap
	_, err = compute.NewVirtualMachine(ctx, "edge-az-01", &compute.VirtualMachineArgs{
		ResourceGroupName: rg.Name,
        Location:          rg.Location,

		NetworkProfile: &compute.NetworkProfileArgs{
			NetworkInterfaces: compute.NetworkInterfaceReferenceArray{
				&compute.NetworkInterfaceReferenceArgs{
					Id:      nic.ID(),
					Primary: pulumi.Bool(true),
				},
			},
		},
		// --- Spot settings ---
		Priority:       pulumi.String("Spot"),
		EvictionPolicy: pulumi.String("Deallocate"),
		BillingProfile: &compute.BillingProfileArgs{
			MaxPrice: pulumi.Float64(-1),
		},
		HardwareProfile: &compute.HardwareProfileArgs{VmSize: pulumi.String("Standard_B2ts_v2")},
		OsProfile: &compute.OSProfileArgs{
			ComputerName:  pulumi.String("edge-az-01"),
			AdminUsername: pulumi.String("azureuser"),
			CustomData:    pulumi.String(customData),
			LinuxConfiguration: &compute.LinuxConfigurationArgs{
				DisablePasswordAuthentication: pulumi.Bool(true),
				Ssh: &compute.SshConfigurationArgs{
					PublicKeys: compute.SshPublicKeyTypeArray{
						&compute.SshPublicKeyTypeArgs{
							Path:    pulumi.String("/home/azureuser/.ssh/authorized_keys"),
							KeyData: pulumi.String(sshKey),
						},
					},
				},
			},
		},
		StorageProfile: &compute.StorageProfileArgs{
			ImageReference: &compute.ImageReferenceArgs{
				Publisher: pulumi.String("Canonical"),
				Offer:     pulumi.String("0001-com-ubuntu-server-jammy"),
				Sku:       pulumi.String("22_04-lts-gen2"),
				Version:   pulumi.String("latest"),
			},
			OsDisk: &compute.OSDiskArgs{
				CreateOption: pulumi.String("FromImage"),
				ManagedDisk:  &compute.ManagedDiskParametersArgs{StorageAccountType: pulumi.String("Standard_LRS")},
			},
		},
	})
	if err != nil { return err }

	ctx.Export("edge-public-ip", pip.IpAddress)
	return nil
}