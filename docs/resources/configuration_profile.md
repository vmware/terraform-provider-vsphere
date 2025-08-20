---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_configuration_profile"
sidebar_current: "docs-vsphere-resource-config-profile"
description: |-
  Provides a vSphere cluster configuration profile resource.
---

# vsphere_configuration_profile

The `vsphere_configuration_profile` resource can be used to configure profile-based host management on a vSphere compute cluster.
The source for the configuration can either be a ESXi host that is part of the compute cluster or a JSON file, but not both at the same time.

It is allowed to switch from one type of configuration source to the other at any time.

Deleting a `vsphere_configuration_profile` resource has no effect on the compute cluster. Once management via configuration
profiles is turned ot it is not possible to disable it.

~> **NOTE:** This resource requires a vCenter 8 or higher and will not work on
direct ESXi connections.

## Example Usage

### Creating a profile using an ESXi host as a reference

The following example sets up a configuration profile on a compute cluster using one of its hosts as a reference
and then propagates that configuration to two additional clusters.

Note that this example assumes that the hosts across all three clusters are compatible with the source configuration. 
This includes but is not limited to their ESXi versions and hardware capabilities.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "cluster1" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_compute_cluster" "cluster2" {
  name          = "cluster-02"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_compute_cluster" "cluster3" {
  name          = "cluster-03"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

# This host is assumed to be part of "cluster-01"
data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

# Configure a profile on "cluster-01" using one of its hosts as a reference
resource "vsphere_configuration_profile" "profile1" {
  cluster_id        = data.vsphere_compute_cluster.cluster1.id
  reference_host_id = data.vsphere_host.host.id
}

# Copy the configuration of "cluster-01" onto "cluster-02"
resource "vsphere_configuration_profile" "profile2" {
  cluster_id    = data.vsphere_compute_cluster.cluster2.id
  configuration = vsphere_configuration_profile.profile1.configuration
}

# Copy the configuration of "cluster-01" onto "cluster-03"
resource "vsphere_configuration_profile" "profile3" {
  cluster_id    = data.vsphere_compute_cluster.cluster3.id
  configuration = vsphere_configuration_profile.profile1.configuration
}
```

### Creating a profile using a configuration file

This example sets up a configuration profile on a cluster by reading a configuration from a JSON
file on the local filesystem. Reading files is natively supported by Terraform.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "cluster1" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_configuration_profile" "profile1" {
  cluster_id    = data.vsphere_compute_cluster.cluster1.id
  configuration = file("/path/to/cluster_config_1.json")
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) The identifier of the cluster.
* `reference_host_id` - (Optional) The identifier of the host to use as a configuration source.
The host needs to be a member of the cluster identified by `cluster_id`. This argument can only be specified if
`configuration` is not set.
* `configuration` - (Optional) The configuration JSON provided as a plain string. This argument can only be specified if `reference_host_id` is not set.

## Attribute Reference

The following attributes are exported:

* `id` - A custom identifier for the profile. The value for this attribute is constructed using the `cluster_id` in the following format - `configuration_profile_${cluster_id}`.
* `schema`- The JSON schema for the profile.
* `configuration` - The current configuration which is active on the cluster.
