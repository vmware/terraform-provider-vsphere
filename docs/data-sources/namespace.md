---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_namespace"
sidebar_current: "docs-vsphere-data-source-vsphere-namespace"
description: |-
  Provides a data source to read the properties of a vSphere Namespace.
---

# vsphere_namespace

This data source can be used to read the properties of a vSphere Namespace. It requires vCenter 8.0.0.1 or later.

## Example Usages

### Read a Namespace

```hcl
data "vsphere_namespace" "example" {
  name = "example-namespace"
}
```

### Read the storage policies of a Namespace

```hcl
output "storage_policy_ids" {
  value = [for s in data.vsphere_namespace.example.storage_policy : s.policy]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vSphere Namespace.

## Attribute Reference

The following attributes are exported:

* `config_status` - The configuration status of the vSphere Namespace.
* `supervisor` - The identifier of the vSphere Supervisor managing the namespace.
* `cpu_usage` - The CPU usage of the vSphere Namespace in MHz.
* `memory_usage` - The memory usage of the vSphere Namespace in MiB.
* `storage_usage` - The storage usage of the vSphere Namespace in MiB.
* `vm_service` - The configuration for VM Service in the vSphere Namespace.
  * `content_libraries` - The list of content libraries associated with the VM Service.
  * `vm_classes` - The list of VM Classes associated with the VM Service.
* `storage_policy` - The storage policies that are available in the vSphere Namespace.
  * `policy` - The identifier of the storage policy.
  * `limit` - The maximum storage, in mebibytes (MiB), that the vSphere Namespace can use for the storage policy. `0` means no limit.

~> **NOTE:** In v2.18.0, `storage_policy` superseded the `storage_policies` attribute. To upgrade, replace references to `data.vsphere_namespace.example.storage_policies` with `[for s in data.vsphere_namespace.example.storage_policy : s.policy]`.
