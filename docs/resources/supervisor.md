---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_supervisor"
sidebar_current: "docs-vsphere-resource-vsphere-supervisor"
description: |-
  Provides a VMware vSphere Supervisor resource..
---

# vsphere_supervisor

Provides a resource for configuring Workload Management.

This resource is deprecated and will be removed in a future major version of the provider.
Use `vsphere_supervisor_v2` instead.

## Example Usages

### Enable Workload Management on a Compute Cluster

```hcl
resource "vsphere_virtual_machine_class" "vm_class" {
  name   = "custom-class"
  cpus   = 4
  memory = 4096
}

resource "vsphere_supervisor" "supervisor" {
  cluster         = "<compute_cluster_id>"
  storage_policy  = "<storage_policy_name>"
  content_library = "<content_library_id>"
  main_dns        = "10.0.0.250"
  worker_dns      = "10.0.0.250"
  edge_cluster    = "<edge_cluster_id>"
  dvs_uuid        = "<distributed_switch_uuid>"
  sizing_hint     = "MEDIUM"

  management_network {
    network          = "<portgroup_id>"
    subnet_mask      = "255.255.255.0"
    starting_address = "10.0.0.150"
    gateway          = "10.0.0.250"
    address_count    = 5
  }

  ingress_cidr {
    address = "10.10.10.0"
    prefix  = 24
  }

  egress_cidr {
    address = "10.10.11.0"
    prefix  = 24
  }

  pod_cidr {
    address = "10.244.10.0"
    prefix  = 23
  }

  service_cidr {
    address = "10.10.12.0"
    prefix  = 24
  }

  search_domains = ["vsphere.local"]

  namespace {
    name              = "custom-namespace"
    content_libraries = []
    vm_classes        = ["${vsphere_virtual_machine_class.vm_class.id}"]
  }
}
```

## Argument Reference

* `cluster` - The identifier of the compute cluster.
* `storage_policy` - The name of the storage policy.
* `management_network` - The configuration for the management network which the control plane VMs will be connected to.
* * `network` - ID of the network. (e.g. a distributed port group).
* * `starting_address` - Starting address of the management network range.
* * `subnet_mask` - Subnet mask.
* * `gateway` - Gateway IP address.
* * `address_count` - Number of addresses to allocate. Starts from `starting_address`
* `content_library` - The identifier of the subscribed content library.
* `main_dns` - The list of addresses of the primary DNS servers.
* `worker_dns` - The list of addresses of the DNS servers to use for the worker nodes.
* `main_ntp` - The list of addresses of the primary NTP servers.
* `worker_ntp` - The list of addresses of the NTP servers to use for the worker nodes.
* `edge_cluster` - The identifier of the NSX Edge Cluster.
* `dvs_uuid` - The UUID of the distributed switch.
* `sizing_hint` - The size of the Kubernetes API server.
* `egress_cidr` - CIDR blocks from which NSX assigns IP addresses used for performing SNAT from container IPs to external IPs.
* `ingress_cidr` - CIDR blocks from which NSX assigns IP addresses for Kubernetes Ingresses and Kubernetes Services of type LoadBalancer.
* `pod_cidr` - CIDR blocks from which Kubernetes allocates pod IP addresses. Minimum subnet size is 23.
* `service_cidr` - CIDR block from which Kubernetes allocates service cluster IP addresses.
* `search_domains` - List of DNS search domains.
* `namespace` - The list of namespaces to create in the Supervisor cluster

### Namespace schema

* `name` - The name of the namespace
* `content_libraries` - The list of content libraries to associate with the namespace
* `vm_classes` - The list of virtual machine classes to add to the namespace
