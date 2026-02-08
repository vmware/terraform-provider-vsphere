---
subcategory: "Networking"
page_title: "VMware vSphere: vsphere_network"
sidebar_current: "docs-vsphere-data-source-network"
description: |-
  Provides a vSphere network data source.
  This can be used to get the general attributes of a vSphere network.
---

# vsphere_network

The `vsphere_network` data source can be used to discover the ID of a network in
vSphere. This can be any network that can be used as the backing for a network
interface for `vsphere_virtual_machine` or any other vSphere resource that
requires a network. This includes standard (host-based) port groups, distributed
port groups, or opaque networks such as those managed by NSX.

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_network" "network" {
  name          = "VM Network"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}
```

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_network" "my_port_group" {
  datacenter_id = data.vsphere_datacenter.datacenter.id
  name          = "VM Network"
  filter {
    network_type = "Network"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Optional) The name of the network. This can be a name or path. Exactly one of `name` or `vlan_id` must be set.
* `vlan_id` - (Optional) The VLAN ID of the network. Exactly one of `name` or `vlan_id` must be set. This is used to look up a Distributed Virtual Port Group by its VLAN ID.
* `datacenter_id` - (Optional) The
  [managed object reference ID][docs-about-morefs] of the datacenter the network
  is located in. This can be omitted if the search path used in `name` is an
  absolute path. For default datacenters, use the `id` attribute from an empty
  `vsphere_datacenter` data source.
* `distributed_virtual_switch_uuid` - (Optional) For distributed port group type
  network objects, the ID of the distributed virtual switch for which the port
  group belongs. It is useful to differentiate port groups with same name using
  the distributed virtual switch ID.
* `filter` - (Optional) Apply a filter for the discovered network.
  * `network_type`: This is required if you have multiple port groups with the same name. This will be one of `DistributedVirtualPortgroup` for distributed port groups, `Network` for standard (host-based) port groups, or `OpaqueNetwork` for networks managed externally, such as those managed by NSX.
* `retry_timeout` - (Optional) The timeout duration in seconds for the data source to retry read operations.
* `retry_interval` - (Optional) The interval in milliseconds to retry the read operation if `retry_timeout` is set. Default: 500.
* `vpc_project_id` - (Optional) Select a project scope for retrieval of VPC subnets.
* `vpc_id` - (Optional) Select a VPC scope for retrieval of VPC subnets.
  [docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

## Attribute Reference

The following attributes are exported:

* `id`: The [managed object ID][docs-about-morefs] of the network.
* `type`: The managed object type for the discovered network. This will be one
  of `DistributedVirtualPortgroup` for distributed port groups, `Network` for
  standard (host-based) port groups, or `OpaqueNetwork` for networks managed
  externally, such as those managed by NSX.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider
