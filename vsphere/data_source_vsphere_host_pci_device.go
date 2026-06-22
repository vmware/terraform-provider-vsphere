// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/hostsystem"
)

func dataSourceVSphereHostPciDevice() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostPciDeviceRead,

		Schema: map[string]*schema.Schema{
			"host_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Managed Object ID of the host system.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regular expression used to match the PCI device name.",
			},
			"class_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The hexadecimal value of the PCI device's class ID.",
			},
			"vendor_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The hexadecimal value of the PCI device's vendor ID.",
			},
			"pci_devices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of matching PCI Devices available on the host.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name ID of this PCI, composed of 'bus:slot.function'",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the PCI device.",
						},
						"bus": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The bus ID of the PCI device.",
						},
						"slot": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The slot ID of the PCI device.",
						},
						"function": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The function ID of the PCI device.",
						},
						"class_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hexadecimal value of the PCI device's class ID.",
						},
						"vendor_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hexadecimal value of the PCI device's vendor ID.",
						},
						"sub_vendor_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hexadecimal value of the PCI device's sub vendor ID.",
						},
						"vendor_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The vendor name of the PCI device.",
						},
						"device_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hexadecimal value of the PCI device's device ID.",
						},
						"sub_device_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hexadecimal value of the PCI device's sub device ID.",
						},
						"parent_bridge": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The parent bridge of the PCI device.",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereHostPciDeviceRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] DataHostPCIDev: Beginning PCI device lookup on %s", d.Get("host_id").(string))

	client := meta.(*Client).vimClient

	host, err := hostsystem.FromID(client, d.Get("host_id").(string))
	if err != nil {
		return err
	}

	hprops, err := hostsystem.Properties(host)
	if err != nil {
		return err
	}

	// Create unique ID based on the host_id
	idsum := sha256.New()
	if _, err := fmt.Fprintf(idsum, "%#v", d.Get("host_id").(string)); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%x", idsum.Sum(nil)))

	// Identify PCI devices matching name_regex (if any)
	devices, err := matchName(d, hprops.Hardware.PciDevice)
	if err != nil {
		return err
	}

	// Output slice
	pciDevices := make([]interface{}, 0, len(devices))

	log.Printf("[DEBUG] DataHostPCIDev: Looking for a device with matching class_id and vendor_id")

	// Loop through devices
	for _, device := range devices {
		// Match the class_id if it is set.
		if class, exists := d.GetOk("class_id"); exists {
			classInt, err := strconv.ParseInt(class.(string), 16, 16)
			if err != nil {
				return err
			}
			if device.ClassId != int16(classInt) {
				continue
			}
		}

		// Now match the vendor_id if it is set.
		if vendor, exists := d.GetOk("vendor_id"); exists {
			vendorInt, err := strconv.ParseInt(vendor.(string), 16, 16)
			if err != nil {
				return err
			}
			if device.VendorId != int16(vendorInt) {
				continue
			}
		}

		// Convertions
		classHex := strconv.FormatInt(int64(device.ClassId), 16)
		vendorHex := strconv.FormatInt(int64(device.VendorId), 16)
		subVendorHex := strconv.FormatInt(int64(device.SubVendorId), 16)
		deviceHex := strconv.FormatInt(int64(device.DeviceId), 16)
		subDeviceHex := strconv.FormatInt(int64(device.SubDeviceId), 16)
		busString := fmt.Sprintf("%v", device.Bus)
		slotString := fmt.Sprintf("%v", device.Slot)
		functionString := fmt.Sprintf("%v", device.Function)

		dev := map[string]interface{}{
			"id":            device.Id,
			"name":          device.DeviceName,
			"class_id":      classHex,
			"vendor_id":     vendorHex,
			"sub_vendor_id": subVendorHex,
			"device_id":     deviceHex,
			"sub_device_id": subDeviceHex,
			"bus":           busString,
			"slot":          slotString,
			"function":      functionString,
			"parent_bridge": device.ParentBridge,
			"vendor_name":   device.VendorName,
		}

		// Add PCI device to output slice
		pciDevices = append(pciDevices, dev)

		log.Printf("[DEBUG] DataHostPCIDev: Matching PCI device found: %s", device.DeviceName)
	}

	// Set the `pci_devices` output to all PCI devices
	if err := d.Set("pci_devices", pciDevices); err != nil {
		return err
	}

	return nil
}

func matchName(d *schema.ResourceData, devices []types.HostPciDevice) ([]types.HostPciDevice, error) {
	log.Printf("[DEBUG] DataHostPCIDev: Selecting devices which match name regex")
	var matches []types.HostPciDevice
	re, err := regexp.Compile(d.Get("name_regex").(string))
	if err != nil {
		return nil, err
	}
	for _, device := range devices {
		if re.Match([]byte(device.DeviceName)) {
			matches = append(matches, device)
		}
	}
	return matches, nil
}
