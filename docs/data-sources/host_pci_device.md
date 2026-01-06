---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_host_pci_device"
sidebar_current: "docs-vsphere-data-source-host_pci_device"
description: |-
  A data source that can be used to get information for PCI passthrough
  device(s) on an ESXi host. The returned attribute `pci_devices` will
  be a list of matching PCI Passthrough devices, based on the criteria:
    - name_regex
    - class_id
    - vendor_id

  **NOTE** - The matching criteria above are evaluated in that order.
---

# vsphere_host_pci_device

The `vsphere_host_pci_device` data source can be used to discover the device ID(s)
of a vSphere host's PCI device(s). This can then be used with
`vsphere_virtual_machine`'s `pci_device_id`.

## Example Usage with Vendor ID and Class ID

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_host_pci_device" "dev" {
  host_id   = data.vsphere_host.host.id
  class_id  = 123
  vendor_id = 456
}
```

## Example Usage with Name Regular Expression

 ```hcl
 data "vsphere_datacenter" "datacenter" {
   name = "dc-01"
 }

 data "vsphere_host" "host" {
   name          = "esxi-01.example.com"
   datacenter_id = data.vsphere_datacenter.datacenter.id
 }

 data "vsphere_host_pci_device" "dev" {
   host_id    = data.vsphere_host.host.id
   name_regex = "MMC"
 }
 ```

## Argument Reference

The following arguments are supported:

* `host_id` - (Required) The [managed object reference ID][docs-about-morefs] of
  a host.
* `name_regex` - (Optional) A regular expression that will be used to match the
  host PCI device name.
* `class_id` - (Optional) The hexadecimal PCI device class ID.
* `vendor_id` - (Optional) The hexadecimal PCI device vendor ID.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

~> **NOTE:** `name_regex`, `class_id`, and `vendor_id` can all be used together.
The above arguments are evaluated and filter PCI Device results in the above order.

## Attribute Reference

The following attributes are exported:

* `host_id` - The [managed objectID][docs-about-morefs] of the ESXi host.
* `id` - Unique (SHA256) id based on the host_id if the ESXi host.
* `name_regex` - (Optional) A regular expression that will be used to match the
  host vGPU profile name.
* `class_id` - (Optional) The hexadecimal PCI device class ID.
* `vendor_id` - (Optional) The hexadecimal PCI device vendor ID.
* `pci_devices` - The list of matching PCI Devices available on the host.
  * `id` - The name ID of this PCI, composed of "bus:slot.function"
  * `name` - The name of the PCI device.
  * `bus` - The bus ID of the PCI device.
  * `class_id` - The hexadecimal value of the PCI device's class ID.
  * `device_id` - The hexadecimal value of the PCI device's device ID.
  * `function` - The function ID of the PCI device.
  * `parent_bridge` - The parent bridge of the PCI device.
  * `slot` - The slot ID of the PCI device.
  * `sub_device_id` - The hexadecimal value of the PCI device's sub device ID.
  * `sub_vendor_id` - The hexadecimal value of the PCI device's sub vendor ID.
  * `vendor_id` - The hexadecimal value of the PCI device's vendor ID.
  * `vendor_name` - The vendor name of the PCI device.
