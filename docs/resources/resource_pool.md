---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_resource_pool"
sidebar_current: "docs-vsphere-resource-compute-resource-pool"
description: |-
  Provides a resource for VMware vSphere resource pools.
  This can be used to create and manage resource pools.
---

# vsphere_resource_pool

The `vsphere_resource_pool` resource can be used to create and manage
resource pools on DRS-enabled vSphere clusters or standalone ESXi hosts.

For more information on vSphere resource pools, please refer to the
[product documentation][ref-vsphere-resource_pools].

[ref-vsphere-resource_pools]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-resource-management-8-0/managing-resource-pools.html

## Example Usage

The following example sets up a resource pool in an existing compute cluster
with the default settings for CPU and memory reservations, shares, and limits.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "compute_cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "resource-pool-01"
  parent_resource_pool_id = data.vsphere_compute_cluster.compute_cluster.resource_pool_id
}
```

A virtual machine resource could be targeted to use the default resource pool
of the cluster using the following:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
  # ... other configuration ...
}
```

The following example sets up a parent resource pool in an existing compute cluster
with a child resource pool nested below. Each resource pool is configured with
the default settings for CPU and memory reservations, shares, and limits.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "compute_cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_resource_pool" "resource_pool_parent" {
  name                    = "parent"
  parent_resource_pool_id = data.vsphere_compute_cluster.compute_cluster.resource_pool_id
}

resource "vsphere_resource_pool" "resource_pool_child" {
  name                    = "child"
  parent_resource_pool_id = vsphere_resource_pool.resource_pool_parent.id
}
```

The following example set up a parent resource pool on a standalone ESXi host with the default
settings for CPU and memory reservations, shares, and limits.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_host_thumbprint" "thumbprint" {
  address  = "esxi-01.example.com"
  insecure = true
}

resource "vsphere_host" "esx-01" {
  hostname   = "esxi-01.example.com"
  username   = "root"
  password   = "password"
  license    = "00000-00000-00000-00000-00000"
  thumbprint = data.vsphere_host_thumbprint.thumbprint.id
  datacenter = data.vsphere_datacenter.datacenter.id
}
```

After the hosts are added to the datacenter, a resource pool can then be created on the host.

```hcl
data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_resource_pool" "resource_pool" {
  name                    = "site1-resource-pool"
  parent_resource_pool_id = data.vsphere_host.host.resource_pool_id
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the resource pool.
* `parent_resource_pool_id` - (Required) The [managed object ID][docs-about-morefs]
  of the parent resource pool. This can be the root resource pool for a cluster
  or standalone host, or a resource pool itself. When moving a resource pool
  from one parent resource pool to another, both must share a common root
  resource pool.
* `cpu_share_level` - (Optional) The CPU allocation level. The level is a
  simplified view of shares. Levels map to a pre-determined set of numeric
  values for shares. Can be one of `low`, `normal`, `high`, or `custom`. When
  `low`, `normal`, or `high` are specified values in `cpu_shares` will be
  ignored.  Default: `normal`
* `cpu_shares` - (Optional) The number of shares allocated for CPU. Used to
  determine resource allocation in case of resource contention. If this is set,
  `cpu_share_level` must be `custom`.
* `cpu_reservation` - (Optional) Amount of CPU (MHz) that is guaranteed
  available to the resource pool. Default: `0`
* `cpu_expandable` - (Optional) Determines if the reservation on a resource
  pool can grow beyond the specified value if the parent resource pool has
  unreserved resources. Default: `true`
* `cpu_limit` - (Optional) The CPU utilization of a resource pool will not
  exceed this limit, even if there are available resources. Set to `-1` for
  unlimited. Default: `-1`
* `memory_share_level` - (Optional) The memory allocation level. The level is a
  simplified view of shares. Levels map to a pre-determined set of numeric
  values for shares. Can be one of `low`, `normal`, `high`, or `custom`. When
  `low`, `normal`, or `high` are specified values in `memory_shares` will be
  ignored.  Default: `normal`
* `memory_shares` - (Optional) The number of shares allocated for memory. Used
  to determine resource allocation in case of resource contention. If this is
  set, `memory_share_level` must be `custom`.
* `memory_reservation` - (Optional) Amount of memory (MB) that is guaranteed
  available to the resource pool. Default: `0`
* `memory_expandable` - (Optional) Determines if the reservation on a resource
  pool can grow beyond the specified value if the parent resource pool has
  unreserved resources. Default: `true`
* `memory_limit` - (Optional) The memory utilization of a resource pool will not
  exceed this limit, even if there are available resources. Set to `-1` for
  unlimited. Default: `-1`
* `scale_descendants_shares` - (Optional) Determines if the shares of all
  descendants of the resource pool are scaled up or down when the shares
  of the resource pool are scaled up or down. Can be one of `disabled` or
  `scaleCpuAndMemoryShares`. Default: `disabled`.
* `tags` - (Optional) The IDs of any tags to attach to this resource. See
  [here][docs-applying-tags] for a reference on how to apply tags.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider
[docs-applying-tags]: /docs/providers/vsphere/r/tag.html#using-tags-in-a-supported-resource

## Attribute Reference

The only attribute this resource exports is the `id` of the resource, which is
the [managed object ID][docs-about-morefs] of the resource pool.

## Importing

An existing resource pool can be [imported][docs-import] into this resource via
the path to the resource pool, using the following command:

[docs-import]: https://developer.hashicorp.com/terraform/cli/import

```shell
terraform import vsphere_resource_pool.resource_pool /dc-01/host/cluster-01/Resources/resource-pool-01
```

The above would import the resource pool named `resource-pool-01` that is located
in the compute cluster `cluster-01` in the `dc-01` datacenter.

### Settings that Require vSphere 7.0 or higher

These settings require vSphere 7.0 or higher:

* [`scale_descendants_shares`](#scale_descendants_shares)
