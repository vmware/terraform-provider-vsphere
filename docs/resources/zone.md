---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_zone"
sidebar_current: "docs-vsphere-resource-zone"
description: |-
  Provides a vSphere Zone resource.
---

# vsphere_zone

The `vsphere_zone` resource can be used to create a vSphere Zone and associate it with one or more
compute clusters.

**NOTE:** vSphere Zones are available in vSphere 8. This resource requires vSphere 8 and later.

## Example Usage

### Create a zone without associating it with any compute clusters 

```hcl
resource "vsphere_zone" "zone1" {
  name        = "zone-1"
  description = "a sample zone"
}
```

### Create a zone and associate it with a compute cluster

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_zone" "zone1" {
  name        = "zone-1"
  description = "a sample zone"
  cluster_ids = [data.vsphere_compute_cluster.cluster.id]
}
```

## Argument Reference

The following arguments are supported:

* `name` - The display name of the vSphere Zone. Matches the identifier of the resource.
* `description` - (Optional) The plain text description of the vSphere Zone.
* `cluster_ids` - (Optional) The identifiers of the compute clusters (e.g. `domain-c123`) to associate with this vSphere Zone.

## Attribute Reference

The following attributes are exported:

* `id` - The identifier of the vSphere Zone. Matches the name of the Zone.
