---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_namespace"
sidebar_current: "docs-vsphere-resource-vsphere-namespace"
description: |-
  Provides a vSphere Namespace resource.
---

# vsphere_namespace

Provides a resource for configuring vSphere Namespaces.

## Example Usages

### Create a Namespace

```hcl
resource "vsphere_namespace" "example" {
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

  storage_policy {
    policy = "1bb87d1e-49a5-49f8-ad4a-7094be13d4a0"
    limit  = 102400
  }

  access {
    subject_type = "USER"
    subject      = "administrator"
    domain       = "vsphere.local"
    role         = "OWNER"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vSphere Namespace. Changing this forces a new resource to be created.
* `supervisor` - (Required) The identifier of the vSphere Supervisor managing the vSphere Namespace. Changing this forces a new resource to be created.
* `vm_service` - (Optional) The configuration for VM Service in the vSphere Namespace.
  * `content_libraries` - (Optional) The list of content libraries to associate with the VM Service.
  * `vm_classes` - (Optional) The list of VM Classes to associate with the VM Service.
* `storage_policy` - (Optional) The storage policies that will be available in the vSphere Namespace. Specify one block per storage policy.
  * `policy` - (Required) The identifier of the storage policy.
  * `limit` - (Optional) The maximum storage, in mebibytes (MiB), that the vSphere Namespace can use for the storage policy. Unset or `0` means no limit.
* `access` - (Optional) The users and groups with access to the vSphere Namespace. Specify one block per subject.
  * `subject_type` - (Required) The type of the subject. One of `USER` or `GROUP`.
  * `subject` - (Required) The name of the user or group, without the domain.
  * `domain` - (Required) The domain of the identity source that owns the subject.
  * `role` - (Required) The role of the subject in the vSphere Namespace. One of `OWNER`, `EDIT` or `VIEW`.
  * `identity_provider` - (Optional) The identifier of the identity provider. If unset, vCenter Single Sign-On is used.

~> **NOTE:** In v2.18.0, `storage_policy` superseded the `storage_policies` argument. To upgrade, replace

```hcl
storage_policies = ["policy-a", "policy-b"]
```

with one `storage_policy` block per storage policy:

```hcl
storage_policy {
  policy = "policy-a"
}

storage_policy {
  policy = "policy-b"
}
```
