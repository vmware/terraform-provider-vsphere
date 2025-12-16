---
subcategory: "Workload Management"
page_title: "VMware vSphere: namespace"
sidebar_current: "docs-vsphere-resource-vsphere-namespace"
description: |-
  Provides a vSphere Namespace resource.
---

# vsphere_namespace

Provides a resource for configuring vSphere Namespaces.

## Example Usages

### Create a namespace

```hcl
resource vsphere_namespace "example" {
  name       = "example-namespace"
  supervisor = "ff69d7fb-4ad4-44a2-8d91-8b3bede80eaa"

  vm_service {
    content_libraries = [
      "ca616f93-506d-4f00-98bf-7708d30f68bc"
    ]
    vm_classes = [
      "example-vm-class"
    ]
  }

  storage_policies = [
    "1bb87d1e-49a5-49f8-ad4a-7094be13d4a0"
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vSphere namespace.
* `supervisor` - (Required) The identifier of the vSphere Supervisor managing the namespace.
* `vm_service` - (Optional) The configuration for VM Service in the vSphere Namespace.
  * `content_libraries` - (Optional) The list of content libraries to associate with the VM Service.
  * `vm_classes` - (Optional) The list of VM Classes to associate with the VM Service.
* `storage_policies` - (Optional) The list of storage policies that will be available in the vSphere Namespace.
