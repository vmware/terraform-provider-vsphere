---
subcategory: "Virtual Machine"
page_title: "VMware vSphere: vsphere_virtual_machine"
sidebar_current: "docs-vsphere-data-source-virtual-machine"
description: |-
  Provides a VMware vSphere virtual machine data source.
  This can be used to return data from a virtual machine or template.
---

# vsphere_virtual_machine

The `vsphere_virtual_machine` data source can be used to find the UUID of an
existing virtual machine or template. The most common purpose is for finding the
UUID of a template to be used as the source for cloning to a new
[`vsphere_virtual_machine`][docs-virtual-machine-resource] resource. It also
reads the guest ID so that can be supplied as well.

[docs-virtual-machine-resource]: /docs/providers/vsphere/r/virtual_machine.html

## Example Usage

In the following example, a virtual machine template is returned by its unique
name within the `vsphere_datacenter`.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_virtual_machine" "template" {
  name          = "ubuntu-server-template"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}
```

In the following example, each virtual machine template is returned by its
unique full path within the `vsphere_datacenter`.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_virtual_machine" "production_template" {
  name          = "production/templates/ubuntu-server-template"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_virtual_machine" "development_template" {
  name          = "development/templates/ubuntu-server-template"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the virtual machine. This can be a name or the
  full path relative to the datacenter. This is required if a UUID lookup is not
  performed.
* `uuid` - (Optional) Specify this field for a UUID lookup, `name` and
  `datacenter_id` are not required if this is specified.
* `folder` - (Optional) The name of the virtual machine folder where the virtual
  machine is located. The `name` argument is limited to 80 characters. If the
  `name` argument includes the full path to the virtual machine and exceeds the
  80 characters limit, the `folder` folder argument can be used.
* `datacenter_id` - (Optional) The [managed object reference
  ID][docs-about-morefs] of the datacenter the virtual machine is located in.
  This can be omitted if the search path used in `name` is an absolute path. For
  default datacenters, use the `id` attribute from an empty `vsphere_datacenter`
  data source.
* `scsi_controller_scan_count` - (Optional) The number of SCSI controllers to
  scan for disk attributes and controller types on. Default: `1`.
* `nvme_controller_scan_count` - (Optional) The number of NVMe controllers to
  scan for disk attributes and controller types on. Default: `1`.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

~> **NOTE:** For best results, ensure that all the disks on any templates you
use with this data source reside on the primary controller, and leave this value
at the default. See the
[`vsphere_virtual_machine`][docs-virtual-machine-resource] resource
documentation for the significance of this setting, specifically the
[additional requirements and notes for cloning][docs-virtual-machine-resource-cloning]
section.

[docs-virtual-machine-resource-cloning]: /docs/providers/vsphere/r/virtual_machine.html#additional-requirements-and-notes-for-cloning

## Attribute Reference

The following attributes are exported:

* `id` - The UUID of the virtual machine or template.
* `guest_id` - The guest ID of the virtual machine or template.
* `alternate_guest_name` - The alternate guest name of the virtual machine when
  `guest_id` is a non-specific operating system, like `otherGuest` or
  `otherGuest64`.
* `annotation` - The user-provided description of this virtual machine.
* `memory` - The size of the virtual machine's memory, in MB.
* `num_cpus` - The total number of virtual processor cores assigned to this
  virtual machine.
* `num_cores_per_socket` - The number of cores per socket for this virtual
  machine.
* `num_cores_per_numa_node` - The number of cores per NUMA node for this virtual machine.
* `firmware` - The firmware interface that is used by this virtual machine. Can
  be either `bios` or `efi`.
* `hardware_version` - The hardware version number on this virtual machine.
* `scsi_type` - The common type of all SCSI controllers on this virtual machine.
  Will be one of `lsilogic` (LSI Logic Parallel), `lsilogic-sas` (LSI Logic
  SAS), `pvscsi` (VMware Paravirtual), `buslogic` (BusLogic), or `mixed` when
  there are multiple controller types. Only the first number of controllers
  defined by `scsi_controller_scan_count` are scanned.
* `scsi_bus_sharing` - Mode for sharing the SCSI bus. The modes are
  physicalSharing, virtualSharing, and noSharing. Only the first number of
  controllers defined by `scsi_controller_scan_count` are scanned.
* `disks` - Information about each of the disks on this virtual machine or
  template. These are sorted by bus and unit number so that they can be applied
  to a `vsphere_virtual_machine` resource in the order the resource expects
  while cloning. This is useful for discovering certain disk settings while
  performing a linked clone, as all settings that are output by this data source
  must be the same on the destination virtual machine as the source. Only the
  first number of controllers defined by `scsi_controller_scan_count` are
  scanned for disks. The sub-attributes are:
  * `label` -  The label for the disk.
  * `size` - The size of the disk, in GIB.
  * `eagerly_scrub` - Set to `true` if the disk has been eager zeroed.
  * `thin_provisioned` - Set to `true` if the disk has been thin provisioned.
  * `unit_number` - The disk number on the storage bus.
* `network_interface_types` - The network interface types for each network
  interface found on the virtual machine, in device bus order. Will be one of
  `e1000`, `e1000e`, `pcnet32`, `sriov`, `vmxnet2`, `vmxnet3vrdma`, or
  `vmxnet3`.
* `network_interfaces` - Information about each of the network interfaces on
  this virtual machine or template. These are sorted by device bus order so that
  they can be applied to a `vsphere_virtual_machine` resource in the order the
  resource expects while cloning. This is useful for discovering certain network
  interface settings while performing a linked clone, as all settings that are
  output by this data source must be the same on the destination virtual machine
  as the source. The sub-attributes are:
  * `adapter_type` - The network interface types for each network interface found
     on the virtual machine, in device bus order. Will be one of `e1000`,
     `e1000e`, `vmxnet3vrdma`, or `vmxnet3`.
  * `bandwidth_limit` - The upper bandwidth limit of this network interface,
    in Mbits/sec.
  * `bandwidth_reservation` - The bandwidth reservation of this network
     interface, in Mbits/sec.
  * `bandwidth_share_level` - The bandwidth share allocation level for this
     interface. Can be one of `low`, `normal`, `high`, or `custom`.
  * `bandwidth_share_count` - The share count for this network interface when the
     share level is custom.
  * `mac_address` - The MAC address of this network interface.
  * `network_id` - The managed object reference ID of the network this interface
     is connected to.
* `video_card` - Information about the virtual video card
  * `num_displays` - The number of supported displays.
  * `total_video_memory` - The total video memory buffer in megabytes.
  * `graphics_3d` - 3D graphics options.
    * `renderer` - The 3D renderer - software, hardware or automatic.
    * `memory` - The dedicated 3D graphics memory in megabytes.
* `firmware` - The firmware type for this virtual machine. Can be `bios` or
  `efi`.
* `default_ip_address` - Whenever possible, this is the first IPv4 address that
  is reachable through the default gateway configured on the machine, then the
  first reachable IPv6 address, and then the first general discovered address if
  neither exist. If VMware Tools is not running on the virtual machine, or if
  the VM is powered off, this value will be blank.
* `guest_ip_addresses` - A list of IP addresses as reported by VMware Tools.
* `instance_uuid` - The instance UUID of the virtual machine or template.
* `vtpm` - Indicates whether a virtual Trusted Platform Module (TPM) device is present on the virtual machine.

~> **NOTE:** Keep in mind when using the results of `scsi_type` and
`network_interface_types`, that the `vsphere_virtual_machine` resource only
supports a subset of the types returned from this data source. See the
[resource docs][docs-virtual-machine-resource] for more details.
