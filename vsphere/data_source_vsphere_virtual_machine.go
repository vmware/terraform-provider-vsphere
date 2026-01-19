// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"log"
	"path"

	"context"
	"fmt"
	"log"
	"path"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/customattribute"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/virtualmachine"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/virtualdevice"
)

func dataSourceVSphereVirtualMachine() *schema.Resource {
	s := map[string]*schema.Schema{
		"datacenter_id": {
			Type:        schema.TypeString,
			Description: "The managed object ID of the datacenter the virtual machine is in. This is not required when using ESXi directly, or if there is only one datacenter in your infrastructure.",
			Optional:    true,
		},
		"folder": {
			Type:          schema.TypeString,
			Optional:      true,
			Description:   "The name of the folder the virtual machine is in. Allows distinguishing virtual machines with the same name in different folder paths",
			StateFunc:     folder.NormalizePath,
			ConflictsWith: []string{"uuid", "moid"},
		},
		"scsi_controller_scan_count": {
			Type:        schema.TypeInt,
			Description: "The number of SCSI controllers to scan for disk sizes and controller types on.",
			Optional:    true,
			Default:     1,
		},
		"sata_controller_scan_count": {
			Type:        schema.TypeInt,
			Description: "The number of SATA controllers to scan for disk sizes and controller types on.",
			Optional:    true,
			Default:     0,
		},
		"ide_controller_scan_count": {
			Type:        schema.TypeInt,
			Description: "The number of IDE controllers to scan for disk sizes and controller types on.",
			Optional:    true,
			Default:     2,
		},
		"nvme_controller_scan_count": {
			Type:        schema.TypeInt,
			Description: "The number of NVMe controllers to scan for disk sizes and controller types on.",
			Optional:    true,
			Default:     1,
		},
		"scsi_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The common SCSI bus type of all controllers on the virtual machine.",
		},
		"scsi_bus_sharing": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Mode for sharing the SCSI bus.",
		},
		"disks": {
			Type:        schema.TypeList,
			Description: "Select configuration attributes from the disks on this virtual machine, sorted by bus and unit number.",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"size": {
						Type:     schema.TypeInt,
						Computed: true,
					},
					"eagerly_scrub": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"thin_provisioned": {
						Type:     schema.TypeBool,
						Computed: true,
					},
					"label": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"unit_number": {
						Type:     schema.TypeInt,
						Computed: true,
					},
				},
			},
		},
		"network_interface_types": {
			Type:        schema.TypeList,
			Description: "The types of network interfaces found on the virtual machine, sorted by unit number.",
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"network_interfaces": {
			Type:        schema.TypeList,
			Description: "The types of network interfaces found on the virtual machine, sorted by unit number.",
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"adapter_type": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"physical_function": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "The ID of the Physical SR-IOV NIC to attach to, e.g. '0000:d8:00.0'",
					},
					"bandwidth_limit": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      -1,
						Description:  "The upper bandwidth limit of this network interface, in Mbits/sec.",
						ValidateFunc: validation.IntAtLeast(-1),
					},
					"bandwidth_reservation": {
						Type:         schema.TypeInt,
						Optional:     true,
						Default:      0,
						Description:  "The bandwidth reservation of this network interface, in Mbits/sec.",
						ValidateFunc: validation.IntAtLeast(0),
					},
					"bandwidth_share_level": {
						Type:         schema.TypeString,
						Optional:     true,
						Default:      string(types.SharesLevelNormal),
						Description:  "The bandwidth share allocation level for this interface. Can be one of low, normal, high, or custom.",
						ValidateFunc: validation.StringInSlice(sharesLevelAllowedValues, false),
					},
					"bandwidth_share_count": {
						Type:         schema.TypeInt,
						Optional:     true,
						Computed:     true,
						Description:  "The share count for this network interface when the share level is custom.",
						ValidateFunc: validation.IntAtLeast(0),
					},
					"mac_address": {
						Type:     schema.TypeString,
						Computed: true,
					},
					"network_id": {
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},
		},
		"video_card": {
			Type:     schema.TypeList,
			Computed: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"num_displays": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Number of supported displays",
					},
					"total_video_memory": {
						Type:        schema.TypeInt,
						Computed:    true,
						Description: "Video RAM size in megabytes",
					},
					"graphics_3d": {
						Type:     schema.TypeList,
						Computed: true,
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"renderer": {
									Type:        schema.TypeString,
									Computed:    true,
									Description: "Indicates how the virtual device renders 3D graphics",
								},
								"memory": {
									Type:        schema.TypeInt,
									Computed:    true,
									Description: "The amount of dedicated graphics memory in megabytes",
								},
							},
						},
					},
				},
			},
		},
		"default_ip_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The default IP address.",
		},
		"guest_ip_addresses": {
			Type:        schema.TypeList,
			Description: "The current list of IP addresses on this virtual machine.",
			Computed:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"instance_uuid": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Instance UUID of this virtual machine.",
		},
		"vtpm": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Indicates whether a virtual Trusted Platform Module (TPM) device is present on the virtual machine.",
		},
		"tags": {
			Type:     schema.TypeList,
			Computed: true,
			Description: "The tags applied to this virtual machine",
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"custom_attributes": {
			Type:     schema.TypeMap,
			Computed: true,
			Description: "The custom attributes applied to this virtual machine",
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}

	// Merge the VirtualMachineConfig structure so that we can include the number of
	// include the number of cpus, memory, firmware, disks, etc.
	structure.MergeSchema(s, schemaVirtualMachineConfigSpec())

	// make name/uuid/moid Optional/AtLeastOneOf
	s["name"].Required = false
	s["name"].Optional = true
	s["name"].AtLeastOneOf = []string{"name", "uuid", "moid"}

	s["uuid"].Required = false
	s["uuid"].Optional = true
	s["uuid"].AtLeastOneOf = []string{"name", "uuid", "moid"}

	s["moid"].Required = false
	s["moid"].Optional = true
	s["moid"].AtLeastOneOf = []string{"name", "uuid", "moid"}

	// Now that the schema has been composed and merged, we can attach our reader and
	// return the resource back to our host process.
	return &schema.Resource{
		Read:   dataSourceVSphereVirtualMachineRead,
		Schema: s,
	}
}

func dataSourceVSphereVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	providerClient := meta.(*Client)
	restClient := providerClient.restClient
	ctx := context.Background()
	uuid := d.Get("uuid").(string)
	moid := d.Get("moid").(string)
	name := d.Get("name").(string)
	folderName := d.Get("folder").(string)
	var vm *object.VirtualMachine
	var err error

	if uuid != "" {
		log.Printf("[DEBUG] Looking for VM or template by UUID %q", uuid)
		vm, err = virtualmachine.FromUUID(client, uuid)
	} else if moid != "" {
		log.Printf("[DEBUG] Looking for VM or template by MOID %q", moid)
		vm, err = virtualmachine.FromMOID(client, moid)
	} else {
		log.Printf("[DEBUG] Looking for VM or template by name/path %q", name)
		var dc *object.Datacenter
		if dcID, ok := d.GetOk("datacenter_id"); ok {
			dc, err = datacenterFromID(client, dcID.(string))
			if err != nil {
				return fmt.Errorf("cannot locate datacenter: %s", err)
			}
			log.Printf("[DEBUG] Datacenter for VM/template search: %s", dc.InventoryPath)
		}

		searchPath := name
		if len(folderName) > 0 {
			searchPath = path.Join(folderName, name)
		}
		vm, err = virtualmachine.FromPath(client, searchPath, dc)
	}

	if err != nil {
		return fmt.Errorf("error fetching virtual machine: %s", err)
	}

	// Set the managed object id.
	_ = d.Set("moid", vm.Reference().Value)

	props, err := virtualmachine.Properties(vm)
	if err != nil {
		return fmt.Errorf("error fetching virtual machine properties: %s", err)
	}

	if props.Config == nil {
		return fmt.Errorf("no configuration returned for virtual machine %q", vm.InventoryPath)
	}

	if props.Config.Uuid == "" {
		return fmt.Errorf("virtual machine %q does not have a UUID", vm.InventoryPath)
	}

	// Read general VM config info
	if err := flattenVirtualMachineConfigInfo(d, props.Config, client); err != nil {
		return fmt.Errorf("error reading virtual machine configuration: %s", err)
	}

	d.SetId(props.Config.Uuid)
	_ = d.Set("guest_id", props.Config.GuestId)
	_ = d.Set("alternate_guest_name", props.Config.AlternateGuestName)
	_ = d.Set("scsi_type", virtualdevice.ReadSCSIBusType(props.Config.Hardware.Device, d.Get("scsi_controller_scan_count").(int)))
	_ = d.Set("scsi_bus_sharing", virtualdevice.ReadSCSIBusSharing(props.Config.Hardware.Device, d.Get("scsi_controller_scan_count").(int)))
	_ = d.Set("firmware", props.Config.Firmware)
	_ = d.Set("instance_uuid", props.Config.InstanceUuid)
	_ = d.Set("evc_mode", props.Runtime.MinRequiredEVCModeKey)
	disks, err := virtualdevice.ReadDiskAttrsForDataSource(props.Config.Hardware.Device, d)
	if err != nil {
		return fmt.Errorf("error reading disk sizes: %s", err)
	}
	nics, err := virtualdevice.ReadNetworkInterfaceTypes(props.Config.Hardware.Device)
	if err != nil {
		return fmt.Errorf("error reading network interface types: %s", err)
	}
	networkInterfaces, err := virtualdevice.ReadNetworkInterfaces(props.Config.Hardware.Device)
	if err != nil {
		return fmt.Errorf("error reading network interfaces: %s", err)
	}
	videoCard, err := virtualdevice.ReadVideoCardForDataSource(props.Config.Hardware.Device)
	if err != nil {
		return fmt.Errorf("error reading video card: %s", err)
	}
	if err := d.Set("disks", disks); err != nil {
		return fmt.Errorf("error setting disk sizes: %s", err)
	}
	if err := d.Set("network_interface_types", nics); err != nil {
		return fmt.Errorf("error setting network interface types: %s", err)
	}
	if err := d.Set("network_interfaces", networkInterfaces); err != nil {
		return fmt.Errorf("error setting network interfaces: %s", err)
	}
	if err := d.Set("video_card", videoCard); err != nil {
		return fmt.Errorf("error setting video card: %s", err)
	}
	if props.Guest != nil {
		if err := buildAndSelectGuestIPs(d, *props.Guest); err != nil {
			return fmt.Errorf("error setting guest IP addresses: %s", err)
		}
	}

	var isVTPMPresent bool
	for _, dev := range props.Config.Hardware.Device {
		if _, ok := dev.(*types.VirtualTPM); ok {
			isVTPMPresent = true
			break
		}
	}
	_ = d.Set("vtpm", isVTPMPresent)

	var moVM mo.VirtualMachine
	pc := property.DefaultCollector(client.Client)

	if err := pc.RetrieveOne(
		ctx,
		vm.Reference(),
		[]string{"customValue"},
		&moVM,
	); err != nil {
		return err
	}

	customattribute.ReadFromResource(&moVM.ManagedEntity, d)

	tagManager := tags.NewManager(restClient)
	if err := readTagsForResource(tagManager, vm, d); err != nil {
		return fmt.Errorf("error reading tags for VM: %s", err)
	}

	log.Printf("[DEBUG] VM search for %q completed successfully (UUID %q)", name, props.Config.Uuid)
	return nil
}
