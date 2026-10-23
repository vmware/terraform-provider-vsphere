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

  zone {
    name               = "zone-1"
    cpu_limit          = 4000
    cpu_reservation    = 1000
    memory_limit       = 8192
    memory_reservation = 2048

    vm_reservation {
      reserved_vm_class = "example-vm-class"
      count             = 2
    }
  }

  vpc_network {
    vpc = "/orgs/default/projects/example-project/vpcs/example-vpc"
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
* `storage_policy` - (Optional) A storage policy that will be available in the vSphere Namespace.
  * `policy` - (Required) The identifier of the storage policy.
  * `limit` - (Optional) The maximum storage, in mebibytes (MiB), that the vSphere Namespace can use for the storage policy. Unset or `0` means no limit.
* `access` - (Optional) A principal with access to the vSphere Namespace.
  * `subject_type` - (Required) The type of the subject. One of `USER` or `GROUP`.
  * `subject` - (Required) The name of the user or group, without the domain.
  * `domain` - (Required) The domain of the identity source that owns the subject.
  * `role` - (Required) The role of the subject in the vSphere Namespace. One of `OWNER`, `EDIT` or `VIEW`.
  * `identity_provider` - (Optional) The identifier of the identity provider. If unset, vCenter Single Sign-On is used.
* `zone` - (Optional) A vSphere Zone to associate with the vSphere Namespace. See [Zone](#zone).

  ~> **NOTE:** Requires vCenter 9.0.0 or later.

  * `name` - (Required) The identifier of the vSphere Zone.
  * `cpu_limit` - (Optional) The CPU limit, in megahertz (MHz), for the vSphere Namespace in the zone. Unset or `0` means no zone limit.
  * `cpu_reservation` - (Optional) The CPU reservation, in megahertz (MHz), for the vSphere Namespace in the zone. Unset or `0` means no reservation.
  * `memory_limit` - (Optional) The memory limit, in mebibytes (MiB), for the vSphere Namespace in the zone. Unset or `0` means no zone limit.
  * `memory_reservation` - (Optional) The memory reservation, in mebibytes (MiB), for the vSphere Namespace in the zone. Unset or `0` means no reservation.
  * `disallow_unreserved_direct_path_usage` - (Optional) Whether workloads that do not use a reserved VM class are disallowed from using a DirectPath device. Defaults to `false`.
  * `vm_reservation` - (Optional) The VM class instances reserved for the vSphere Namespace in the zone.
    * `reserved_vm_class` - (Required) The identifier of the VM class.
    * `count` - (Required) The number of VM class instances to reserve.
* `vpc_network` - (Optional) The NSX VPC network configuration. Requires vCenter 9.0.0 or later. If omitted, the Supervisor creates a VPC with its default settings for the vSphere Namespace, and its network configuration is exported. To customize the VPC, manage it with the [NSX Terraform provider](https://registry.terraform.io/providers/vmware/nsxt/latest/docs) and set its path in `vpc`.
  * `vpc` - (Required) The NSX path of an existing VPC, for example `/orgs/default/projects/<project>/vpcs/<vpc>`. Changing this forces a new resource to be created.
  * `shared_subnet` - (Optional) The NSX subnets created outside the Supervisor to associate with the vSphere Namespace. Requires vCenter 9.1.0 or later.
    * `path` - (Required) The NSX path of the subnet.
    * `pod_default` - (Optional) Whether the subnet is the default network for vSphere Pods. Requires vCenter 9.1.1 or later.
    * `vm_default` - (Optional) Whether the subnet is the default network for VMs. Requires vCenter 9.1.1 or later.

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

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

* `vpc_network.0.vpc` - The NSX path of the VPC, including one created by the Supervisor.

## Zone

* If `zone` is omitted, the vSphere Zones the namespace already has are kept and exported. To remove a zone, list the zones to keep.
* Zone removal is asynchronous. A zone that is marked for removal is not exported, and it is deleted once its workloads are drained.
* vCenter does not return the limits, reservations, `vm_reservation` or `disallow_unreserved_direct_path_usage` of a zone. The provider keeps the values from the configuration. Changes made outside of Terraform are not detected.
