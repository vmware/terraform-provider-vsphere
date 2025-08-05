---
subcategory: "Virtual Machine"
page_title: "VMware vSphere: vsphere_ovf_vm_template"
sidebar_current: "docs-vsphere-data-source-ovf-vm-template"
description: |-
  A data source that can be used to extract the configuration of an OVF template.
---

# vsphere_ovf_vm_template

The `vsphere_ovf_vm_template` data source can be used to submit an OVF to
vSphere and extract its hardware settings in a form that can be then used as
inputs for a `vsphere_virtual_machine` resource.

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_datastore" "datastore" {
  name          = "datastore-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_resource_pool" "default" {
  name          = format("%s%s", data.vsphere_compute_cluster.cluster.name, "/Resources")
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_network" "network" {
  name          = "172.16.11.0"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

## Remote OVF/OVA Source
data "vsphere_ovf_vm_template" "ovfRemote" {
  name              = "ubuntu-server-cloud-image-01"
  disk_provisioning = "thin"
  resource_pool_id  = data.vsphere_resource_pool.default.id
  datastore_id      = data.vsphere_datastore.datastore.id
  host_system_id    = data.vsphere_host.host.id
  remote_ovf_url    = "https://cloud-images.ubuntu.com/releases/xx.xx/release/ubuntu-xx.xx-server-cloudimg-amd64.ova"
  ovf_network_map = {
    "VM Network" : data.vsphere_network.network.id
  }
}

## Local OVF/OVA Source
data "vsphere_ovf_vm_template" "ovfLocal" {
  name              = "ubuntu-server-cloud-image-02"
  disk_provisioning = "thin"
  resource_pool_id  = data.vsphere_resource_pool.default.id
  datastore_id      = data.vsphere_datastore.datastore.id
  host_system_id    = data.vsphere_host.host.id
  local_ovf_path    = "/Volume/Storage/OVA/ubuntu-xx-xx-server-cloudimg-amd64.ova"
  ovf_network_map = {
    "VM Network" : data.vsphere_network.network.id
  }
}

## Deployment of VM from Remote OVF
resource "vsphere_virtual_machine" "vmFromRemoteOvf" {
  name                 = "ubuntu-server-cloud-image-01"
  datacenter_id        = data.vsphere_datacenter.datacenter.id
  datastore_id         = data.vsphere_datastore.datastore.id
  host_system_id       = data.vsphere_host.host.id
  resource_pool_id     = data.vsphere_resource_pool.default.id
  num_cpus             = data.vsphere_ovf_vm_template.ovfRemote.num_cpus
  num_cores_per_socket = data.vsphere_ovf_vm_template.ovfRemote.num_cores_per_socket
  memory               = data.vsphere_ovf_vm_template.ovfRemote.memory
  guest_id             = data.vsphere_ovf_vm_template.ovfRemote.guest_id
  firmware             = data.vsphere_ovf_vm_template.ovfRemote.firmware
  scsi_type            = data.vsphere_ovf_vm_template.ovfRemote.scsi_type

  dynamic "network_interface" {
    for_each = data.vsphere_ovf_vm_template.ovfRemote.ovf_network_map
    content {
      network_id = network_interface.value
    }
  }

  wait_for_guest_net_timeout = 0
  wait_for_guest_ip_timeout  = 0

  ovf_deploy {
    allow_unverified_ssl_cert = false
    remote_ovf_url            = data.vsphere_ovf_vm_template.ovfRemote.remote_ovf_url
    disk_provisioning         = data.vsphere_ovf_vm_template.ovfRemote.disk_provisioning
    ovf_network_map           = data.vsphere_ovf_vm_template.ovfRemote.ovf_network_map
  }

  ovf_deploy {
    remote_ovf_url  = "https://cloud-images.ubuntu.com/focal/current/focal-server-cloudimg-amd64.ova"
    ovf_network_map = data.vsphere_ovf_vm_template.ovfRemote.ovf_network_map
  }

  cdrom {
    client_device = true
  }

  vapp {
    properties = {
      "hostname"    = var.remote_ovf_name
      "instance-id" = var.remote_ovf_uuid
      "public-keys" = var.remote_ovf_public_keys
      "password"    = var.remote_ovf_password
      "user-data"   = base64encode(var.remote_ovf_user_data)
    }
  }

  lifecycle {
    ignore_changes = [
      vapp[0].properties,
    ]
  }
}

## Deployment of VM from Local OVF
resource "vsphere_virtual_machine" "vmFromLocalOvf" {
  name                 = "ubuntu-server-cloud-image-02"
  datacenter_id        = data.vsphere_datacenter.datacenter.id
  datastore_id         = data.vsphere_datastore.datastore.id
  host_system_id       = data.vsphere_host.host.id
  resource_pool_id     = data.vsphere_resource_pool.default.id
  num_cpus             = data.vsphere_ovf_vm_template.ovfLocal.num_cpus
  num_cores_per_socket = data.vsphere_ovf_vm_template.ovfLocal.num_cores_per_socket
  memory               = data.vsphere_ovf_vm_template.ovfLocal.memory
  guest_id             = data.vsphere_ovf_vm_template.ovfLocal.guest_id
  firmware             = data.vsphere_ovf_vm_template.ovfLocal.firmware
  scsi_type            = data.vsphere_ovf_vm_template.ovfLocal.scsi_type

  dynamic "network_interface" {
    for_each = data.vsphere_ovf_vm_template.ovfLocal.ovf_network_map
    content {
      network_id = network_interface.value
    }
  }

  wait_for_guest_net_timeout = 0
  wait_for_guest_ip_timeout  = 0

  ovf_deploy {
    allow_unverified_ssl_cert = false
    local_ovf_path            = data.vsphere_ovf_vm_template.ovfLocal.local_ovf_path
    disk_provisioning         = data.vsphere_ovf_vm_template.ovfLocal.disk_provisioning
    ovf_network_map           = data.vsphere_ovf_vm_template.ovfLocal.ovf_network_map
  }

  cdrom {
    client_device = true
  }

  vapp {
    properties = {
      "hostname"    = var.local_ovf_name
      "instance-id" = var.local_ovf_uuid
      "public-keys" = var.local_ovf_public_keys
      "password"    = var.local_ovf_password
      "user-data"   = base64encode(var.local_ovf_user_data)
    }
  }

  lifecycle {
    ignore_changes = [
      vapp[0].properties,
    ]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - Name of the virtual machine to create.
* `resource_pool_id` - (Required) The ID of a resource pool in which to place
  the virtual machine.
* `host_system_id` - (Required) The ID of the ESXi host system to deploy the
  virtual machine.
* `datastore_id` - (Required) The ID of the virtual machine's datastore. The
  virtual machine configuration is placed here, along with any virtual disks
  that are created without datastores.
* `folder` - (Required) The name of the folder in which to place the virtual
  machine.
* `local_ovf_path` - (Optional) The absolute path to the OVF/OVA file on the
  local system. When deploying from an OVF, ensure all necessary files such as
  the `.vmdk` files are present in the same directory as the OVF.
* `remote_ovf_url` - (Optional) URL of the remote OVF/OVA file to be deployed.

~> **NOTE:** Either `local_ovf_path` or `remote_ovf_url` is required, both can
  not be empty.

* `ip_allocation_policy` - (Optional) The IP allocation policy.
* `ip_protocol` - (Optional) The IP protocol.
* `disk_provisioning` - (Optional) The disk provisioning type. If set, all the
  disks included in the OVF/OVA will have the same specified policy. Can be
  one of `thin`, `thick`, or `eagerZeroedThick`.
  * `thin`: Each disk is allocated and zeroed on demand as the space is used.
  * `thick`: Each disk is allocated at creation time and the space is zeroed
     on demand as the space is used.
  * `eagerZeroedThick`: Each disk is allocated and zeroed at creation time.
* `deployment_option` - (Optional) The key of the chosen deployment option. If
  empty, the default option is chosen.
* `ovf_network_map` - (Optional) The mapping of name of network identifiers
  from the OVF descriptor to network UUID in the environment.
* `allow_unverified_ssl_cert` - (Optional) Allow unverified SSL certificates
  when deploying OVF/OVA from a URL.
* `enable_hidden_properties` - (Optional) Allow properties with
  `ovf:userConfigurable=false` to be set.

## Attribute Reference

* `num_cpus` - The number of virtual CPUs to assign to the virtual machine.
* `num_cores_per_socket` - The number of cores per virtual CPU in the virtual
  machine.
* `cpu_hot_add_enabled` - Allow CPUs to be added to the virtual machine while
  powered on.
* `cpu_hot_remove_enabled` - Allow CPUs to be removed from the virtual machine
  while powered on.
* `nested_hv_enabled` - Enable nested hardware virtualization on the virtual
  machine, facilitating nested virtualization in the guest.
* `memory` - The size of the virtual machine memory, in MB.
* `memory_hot_add_enabled` - Allow memory to be added to the virtual machine
  while powered on.
* `swap_placement_policy` - The swap file placement policy for the virtual
  machine.
* `annotation` - A description of the virtual machine.
* `guest_id` - The ID for the guest operating system
* `alternate_guest_name` - An alternate guest operating system name.
* `firmware` - The firmware to use on the virtual machine.
