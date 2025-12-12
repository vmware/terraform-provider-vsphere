---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_zone"
sidebar_current: "docs-vsphere-data-source-zone"
description: |-
  Provides a vSphere Zone data source.
---

# vsphere_zone

The `vsphere_zone` data source can be used to list the properties of a vSphere Zone including 
its associated compute clusters.

## Example Usage

```hcl
data "vsphere_zone" "zone1" {
  name = "zone-1"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vSphere Zone.

## Attribute Reference

The following attributes are exported:

* `id` - The identifier of the vSphere Zone. Matches the name of the Zone.
* `name` - The display name of the vSphere Zone. Matches the identifier of the resource.
* `description` - The plain text description of the vSphere Zone.
* `cluster_ids` - The identifiers of the compute clusters (e.g. `domain-c123`) associated with this vSphere Zone.
