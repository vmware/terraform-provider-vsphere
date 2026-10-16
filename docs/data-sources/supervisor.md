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

### Look up a Supervisor by name

```hcl
data "vsphere_supervisor" "example" {
  name = "example-supervisor"
}
```

### Look up a Supervisor by compute cluster

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
