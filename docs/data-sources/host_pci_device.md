---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_host_pci_device"
sidebar_current: "docs-vsphere-data-source-host_pci_device"
description: |-
  A data source that can be used to get information for a PCI passthrough
  device on an ESXi host.
---

# vsphere_host_pci_device

The `vsphere_host_pci_device` data source can be used to discover the device ID
of a vSphere host's PCI device. This can then be used with
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
* `vendor_id` - (Optional) The hexadecimal PCI device vendor ID.
* `class_id` - (Optional) The hexadecimal PCI device class ID

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

~> **NOTE:** `name_regex`, `vendor_id`, and `class_id` can all be used together.

## Attribute Reference

* `id` - The device ID of the PCI device.
* `name` - The name of the PCI device.
