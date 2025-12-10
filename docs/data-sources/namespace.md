---
subcategory: "Workload Management"
page_title: "VMware vSphere: namespace"
sidebar_current: "docs-vsphere-data-source-vsphere-namespace"
description: |-
  Provides a VMware Namespace resource.
---

# vsphere_namespace

The vsphere-namespace data source can be used to read the properties of a vSphere Namespace.

## Example Usages

### Create a namespace

```hcl
data vsphere_namespace "namespace" {
  name = "example-namespace"
}
```

## Argument Reference
The following arguments are supported:

* `name` - The name of the vSphere namespace.

## Attribute Reference

The following attributes are exported:

* `supervisor` - The identifier of the vSphere Supervisor managing the namespace.
* `vm_service` - The configuration for VM Service in the vSphere Namespace.
* * `content_libraries` - The list of content libraries associated with the VM Service.
* * `vm_classes` -  The list of VM Classes associated with the VM Service.
* `storage_policies` - The list of storage policies that are available in the vSphere Namespace.
