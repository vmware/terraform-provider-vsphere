---
subcategory: "Virtual Machine"
page_title: "VMware vSphere: vsphere_virtual_machine_snapshot"
sidebar_current: "docs-vsphere-resource-vm-virtual-machine-snapshot"
description: |-
  Provides a VMware vSphere virtual machine snapshot resource. This can be used to create and delete virtual machine snapshots.
---

# vsphere_virtual_machine_snapshot

The `vsphere_virtual_machine_snapshot` resource can be used to manage snapshots
for a virtual machine.

For more information on managing snapshots and how they work in VMware, see
[here][ext-vm-snapshot-management].

[ext-vm-snapshot-management]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-single-host-management-vmware-host-client-8-0/virtual-machine-management-with-the-vsphere-host-client-vSphereSingleHostManagementVMwareHostClient/managing-virtual-machines-in-the-vsphere-host-client-vSphereSingleHostManagementVMwareHostClient/using-snapshots-to-manage-virtual-machines-vSphereSingleHostManagementVMwareHostClient.html

~> **NOTE:** A snapshot in VMware differs from traditional disk snapshots, and
can contain the actual running state of the virtual machine, data for all disks
that have not been set to be independent from the snapshot (including ones that
have been attached via the [attach][docs-vsphere-virtual-machine-disk-attach]
parameter to the `vsphere_virtual_machine` `disk` block), and even the
configuration of the virtual machine at the time of the snapshot. Virtual
machine, disk activity, and configuration changes post-snapshot are not
included in the original state. Use this resource with care! Neither VMware nor
HashiCorp recommends retaining snapshots for a extended period of time and does
NOT recommend using them as as backup feature. For more information on the
limitation of virtual machine snapshots, see [here][ext-vm-snap-limitations].

[docs-vsphere-virtual-machine-disk-attach]: /docs/providers/vsphere/r/virtual_machine.html#attach
[ext-vm-snap-limitations]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-virtual-machine-administration-guide-8-0/managing-virtual-machinesvsphere-vm-admin/using-snapshots-to-manage-virtual-machinesvsphere-vm-admin/snapshot-file-names-and-descriptionvsphere-vm-admin.html

## Example Usage

```hcl
resource "vsphere_virtual_machine_snapshot" "demo1" {
  virtual_machine_uuid = "9aac5551-a351-4158-8c5c-15a71e8ec5c9"
  snapshot_name        = "Snapshot Name"
  description          = "This is Demo Snapshot"
  memory               = "true"
  quiesce              = "true"
  remove_children      = "false"
  consolidate          = "true"
}
```

## Argument Reference

The following arguments are supported:

~> **NOTE:** All attributes in the `vsphere_virtual_machine_snapshot` resource
are immutable and force a new resource if changed.

* `virtual_machine_uuid` - (Required) The virtual machine UUID.
* `datacenter_id` - (Optional) The [managed object reference
  ID][docs-about-morefs] of the datacenter the virtual machine is located in.
  Set this when more than one virtual machine may share the same BIOS UUID (for
  example a source VM and a replica in another datacenter).
* `snapshot_name` - (Required) The name of the snapshot.
* `description` - (Required) A description for the snapshot.
* `memory` - (Required) If set to `true`, a dump of the internal state of the
  virtual machine is included in the snapshot.
* `quiesce` - (Required) If set to `true`, and the virtual machine is powered
  on when the snapshot is taken, VMware Tools is used to quiesce the file
  system in the virtual machine.
* `remove_children` - (Optional) If set to `true`, the entire snapshot subtree
  is removed when this resource is destroyed.
* `consolidate` - (Optional) If set to `true`, the delta disks involved in this
  snapshot will be consolidated into the parent when this resource is
  destroyed.

## Attribute Reference

The only attribute this resource exports is the resource `id`, which is set to
the [managed object reference ID][docs-about-morefs] of the snapshot.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider
