---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_configuration_profile"
sidebar_current: "docs-vsphere-data-source-config-profile"
description: |-
  Provides a vSphere cluster configuration profile data source.
---

# vsphere_configuration_profile

The `vsphere_configuration_profile` data source can be used to export the configuration and schema
of a cluster that is already managed via configuration profiles.

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "compute_cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_configuration_profile" "profile" {
  cluster_id = data.vsphere_compute_cluster.compute_cluster.id
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The identifier of the compute cluster.

## Attribute Reference

The following attributes are exported:

* `id` - A custom identifier for the profile. The value for this attribute is constructed using the `cluster_id` in the following format - `configuration_profile_${cluster_id}`.
* `schema`- The JSON schema for the profile.
* `configuration` - The current configuration which is active on the cluster.
