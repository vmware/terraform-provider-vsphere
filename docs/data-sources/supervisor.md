---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_supervisor"
sidebar_current: "docs-vsphere-data-source-vsphere-supervisor"
description: |-
  Provides a vSphere Supervisor data source.
---

# vsphere_supervisor

This data source can be used to read the properties of a vSphere Supervisor.

## Example Usages

### Look up a Supervisor by Name

```hcl
data "vsphere_supervisor" "example" {
  name = "example-supervisor"
}
```

### Look up a Supervisor by Compute Cluster

```hcl
data "vsphere_supervisor" "example" {
  cluster = data.vsphere_compute_cluster.cluster.id
}
```

### Look up a Supervisor by ID

```hcl
data "vsphere_supervisor" "example" {
  id = "435a6f10-ffd2-44f4-a2a8-65b4059b0b5f"
}
```

### Find the Supervisor network backed by a port group

```hcl
data "vsphere_network" "workload" {
  name          = "workload-pg"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_supervisor" "example" {
  name = "example-supervisor"
}

locals {
  workload_network = one([
    for network in data.vsphere_supervisor.example.workload_network :
    network.network if network.portgroup == data.vsphere_network.workload.id
  ])
}
```

The name can be used in the `vsphere_network` block of the [`vsphere_namespace`](/docs/providers/vsphere/r/namespace.html) resource.

## Argument Reference

Exactly one of the following arguments must be specified:

* `id` - The ID of the Supervisor.
* `name` - The name of the Supervisor. The lookup fails if more than one Supervisor has this name.
* `cluster` - The ID of a compute cluster managed by the Supervisor.

## Attribute Reference

* `id` - The ID of the Supervisor.
* `name` - The name of the Supervisor.
* `config_status` - The configuration status of the Supervisor.
* `kubernetes_status` - The Kubernetes status of the Supervisor.
* `cpu_usage` - The CPU usage of the Supervisor in megahertz.
* `memory_usage` - The memory usage of the Supervisor in mebibytes.
* `storage_usage` - The storage usage of the Supervisor in mebibytes.
* `zones` - The vSphere Zones of the Supervisor.
  * `zone` - The ID of the vSphere Zone.
  * `clusters` - The IDs of the compute clusters in the vSphere Zone.
* `workload_network` - The workload networks of the Supervisor.
  * `network` - The name of the Supervisor network. Use it in the `vsphere_network` block of the [`vsphere_namespace`](/docs/providers/vsphere/r/namespace.html) resource.
  * `network_provider` - The network provider of the Supervisor network, for example `VSPHERE_NETWORK`.
  * `portgroup` - The ID of the distributed port group backing the Supervisor network. Empty for NSX-backed networks.
  * `ip_assignment_mode` - The IP assignment mode of the Supervisor network. Empty for NSX-backed networks.
  * `gateway` - The gateway of the Supervisor network. Empty for NSX-backed networks.
  * `subnet_mask` - The subnet mask of the Supervisor network. Empty for NSX-backed networks.
  * `namespaces` - The vSphere Namespaces that use the Supervisor network.
