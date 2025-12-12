---
subcategory: "Virtual Machine"
page_title: "VMware vSphere: vsphere_virtual_machine"
sidebar_current: "docs-vsphere-resource-vm-virtual-machine-resource"
description: |-
  Provides a resource for VMware vSphere virtual machines.
  This resource can be used to create, modify, and delete virtual machines.
---

# vsphere_virtual_machine

The `vsphere_virtual_machine` resource is used to manage the lifecycle of a virtual machine.

For details on working with virtual machines in VMware vSphere, please refer to the [product documentation][vmware-docs-vm-management].

[vmware-docs-vm-management]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-virtual-machine-administration-guide-8-0.html

## About Working with Virtual Machines in Terraform

A high degree of control and flexibility is available to a vSphere administrator to configure, deploy, and manage virtual machines. The Terraform provider enables you to manage the desired state of virtual machine resources.

This section provides information on configurations you should consider when setting up virtual machines, creating templates for cloning, and more.

### Disks

The `vsphere_virtual_machine` resource supports standard VMDK-backed virtual disks. It **does not** support raw device mappings (RDMs) to proxy use of raw physical storage device

Disks are managed by a label supplied to the [`label`](#label) attribute in a [`disk` block](#disk-options). This is separate from the automatic naming that vSphere assigns when a virtual machine is created. Control of the name for a virtual disk is not supported unless you are attaching an external disk with the [`attach`](#attach) attribute.

Virtual disks can be SCSI, SATA, NVMe or IDE. The storage controllers managed by the Terraform provider can vary, depending on the value supplied to [`scsi_controller_count`](#scsi_controller_count), [`sata_controller_count`](#sata_controller_count), [`nvme_controller_count`](#nvme_controller_count), or [`ide_controller_count`](#ide_controller_count). This also dictates the controllers that are checked when looking for disks during a cloning process. SCSI controllers are all configured with the controller type defined by the  [`scsi_type`](#scsi_type) setting. If you are cloning from a template, devices will be added or re-configured as necessary.

When cloning from a template, you must specify disks of either the same or greater size than the disks in the source template or the same size when cloning from a snapshot (also known as a linked clone).

See the section on [Creating a Virtual Machine from a Template](#creating-a-virtual-machine-from-a-template) for more information.

### Customization and Network Waiters

Terraform waits during various parts of a virtual machine deployment to ensure that the virtual machine is in an expected state before proceeding. These events occur when a virtual machine is created or updated, depending on the waiter.

The waiters include the following:

* **Customization Waiter**:

  This waiter watches events in vSphere to monitor when customization on a virtual machine completes during creation. Depending on your vSphere or virtual machine configuration, it may be necessary to change the timeout or turn off the waiter. This can be controlled by using the [`timeout`](#timeout-1) setting in the [customization settings](#virtual-machine-customizations) block.

* **Network Waiter**:

  This waiter waits for interfaces to appear on a virtual machine guest operating system and occurs close to the end of both virtual machine creation and update. This waiter ensures that the IP information gets reported to the guest operating system, mainly to facilitate the availability of a valid, reachable default IP address for any provisioners.

  The behavior of the waiter can be controlled with the [`wait_for_guest_net_timeout`](#wait_for_guest_net_timeout), [`wait_for_guest_net_routable`](#wait_for_guest_net_routable), [`wait_for_guest_ip_timeout`](#wait_for_guest_ip_timeout), and [`ignored_guest_ips`](#ignored_guest_ips) settings.

## Example Usage

### Creating a Virtual Machine

The following block contains the option necessary to create a virtual machine, with a single disk and network interface.

In this example, the resource makes use of the following data sources:

* [`vsphere_datacenter`][tf-vsphere-datacenter] to locate the datacenter,

* [`vsphere_datastore`][tf-vsphere-datastore] to locate the default datastore to place the virtual machine files,

* [`vsphere_compute-cluster`][tf-vsphere-compute-cluster] to locate a resource pool located in a cluster or standalone host, and

* [`vsphere_network`][tf-vsphere-network] to locate the network.

[tf-vsphere-datacenter]: /docs/providers/vsphere/d/datacenter.html
[tf-vsphere-datastore]: /docs/providers/vsphere/d/datastore.html
[tf-vsphere-compute-cluster]: /docs/providers/vsphere/d/compute-cluster.html
[tf-vsphere-network]: /docs/providers/vsphere/d/network.html

**Example**:

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

data "vsphere_network" "network" {
  name          = "VM Network"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_virtual_machine" "vm" {
  name             = "foo"
  resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
  datastore_id     = data.vsphere_datastore.datastore.id
  num_cpus         = 1
  memory           = 1024
  guest_id         = "otherLinux64Guest"
  network_interface {
    network_id = data.vsphere_network.network.id
  }
  disk {
    label = "Hard Disk 1"
    size  = 20
  }
}
```

### Cloning and Customization

Building on the above example, the below configuration creates a virtual machine by cloning it from a template, fetched using the [`vsphere_virtual_machine`][tf-vsphere-virtual-machine-ds] data source. This option allows you to locate the UUID of the template to clone, along with settings for network interface type, SCSI bus type, and disk attributes.

[tf-vsphere-virtual-machine-ds]: /docs/providers/vsphere/d/virtual_machine.html

~> **NOTE:** Cloning requires vCenter Server and is not supported on direct ESXi host connections.

**Examples**:

This example clones a Linux template and customizes with the provided settings:

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

data "vsphere_network" "network" {
  name          = "VM Network"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_virtual_machine" "template" {
  name          = "linux-template"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_virtual_machine" "vm" {
  name             = "foo"
  resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
  datastore_id     = data.vsphere_datastore.datastore.id
  num_cpus         = 1
  memory           = 1024
  guest_id         = data.vsphere_virtual_machine.template.guest_id
  scsi_type        = data.vsphere_virtual_machine.template.scsi_type
  network_interface {
    network_id   = data.vsphere_network.network.id
    adapter_type = data.vsphere_virtual_machine.template.network_interface_types[0]
  }
  disk {
    label            = "Hard Disk 1"
    size             = data.vsphere_virtual_machine.template.disks.0.size
    thin_provisioned = data.vsphere_virtual_machine.template.disks.0.thin_provisioned
  }
  clone {
    template_uuid = data.vsphere_virtual_machine.template.id
    customize {
      linux_options {
        host_name = "foo"
        domain    = "example.com"
      }
    }
  }
}
```

This example uses the same structure as the previous example, but customizes with an existing guest customization specification:

```hcl
# ... other configuration ...

data "vsphere_guest_os_customization" "linux" {
  name = "linux"
}

resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  template_uuid = data.vsphere_virtual_machine.template.id
  customization_spec {
    id = data.vsphere_guest_os_customization.linux.id
  }
  # ... other configuration ...
}
```

### Deploying Virtual Machines from OVF/OVA

Virtual machines can be deployed from OVF/OVA using either the local path and remote URL and the `ovf_deploy` property. When deploying from a local path, the path to the OVF/OVA must be provided. While deploying OVF, all other necessary files (_e.g._ `.vmdk`, `.mf`, etc) must be present in the same directory as the `.ovf` file.

~> **NOTE:** The vApp properties which are pre-defined in an OVF template can be overwritten. New vApp properties can not be created for an existing OVF template.

~> **NOTE:** An OVF/OVA deployment requires vCenter Server and is not supported on direct ESXi host connections.

The following example demonstrates a scenario deploying a simple OVF/OVA, using both the local path and remote URL options.

**Example**:

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

## Deployment of VM from Remote OVF
resource "vsphere_virtual_machine" "vmFromRemoteOvf" {
  name             = "remote-foo"
  datacenter_id    = data.vsphere_datacenter.datacenter.id
  datastore_id     = data.vsphere_datastore.datastore.i
  resource_pool_id = data.vsphere_resource_pool.default.id

  wait_for_guest_net_timeout = 0
  wait_for_guest_ip_timeout  = 0

  ovf_deploy {
    allow_unverified_ssl_cert = false
    remote_ovf_url            = "https://example.com/foo.ova"
    disk_provisioning         = "thin"
    ip_protocol               = "IPV4"
    ip_allocation_policy      = "STATIC_MANUAL"
    ovf_network_map = {
      "Network 1" = data.vsphere_network.network.id
      "Network 2" = data.vsphere_network.network.id
    }
  }
  vapp {
    properties = {
      "guestinfo.hostname"  = "remote-foo.example.com",
      "guestinfo.ipaddress" = "172.16.11.101",
      "guestinfo.netmask"   = "255.255.255.0",
      "guestinfo.gateway"   = "172.16.11.1",
      "guestinfo.dns"       = "172.16.11.4",
      "guestinfo.domain"    = "example.com",
      "guestinfo.ntp"       = "ntp.example.com",
      "guestinfo.password"  = "VMware1!",
      "guestinfo.ssh"       = "True"
    }
  }
}

## Deployment of VM from Local OVF
resource "vsphere_virtual_machine" "vmFromLocalOvf" {
  name             = "local-foo"
  datacenter_id    = data.vsphere_datacenter.datacenter.id
  datastore_id     = data.vsphere_datastore.datastore.id
  resource_pool_id = data.vsphere_resource_pool.default.id

  wait_for_guest_net_timeout = 0
  wait_for_guest_ip_timeout  = 0

  ovf_deploy {
    allow_unverified_ssl_cert = false
    local_ovf_path            = "/Volume/Storage/OVAs/foo.ova"
    disk_provisioning         = "thin"
    ip_protocol               = "IPV4"
    ip_allocation_policy      = "STATIC_MANUAL"
    ovf_network_map = {
      "Network 1" = data.vsphere_network.network.id
      "Network 2" = data.vsphere_network.network.id
    }
  }
  vapp {
    properties = {
      "guestinfo.hostname"  = "local-foo.example.com",
      "guestinfo.ipaddress" = "172.16.11.101",
      "guestinfo.netmask"   = "255.255.255.0",
      "guestinfo.gateway"   = "172.16.11.1",
      "guestinfo.dns"       = "172.16.11.4",
      "guestinfo.domain"    = "example.com",
      "guestinfo.ntp"       = "ntp.example.com",
      "guestinfo.password"  = "VMware1!",
      "guestinfo.ssh"       = "True"
    }
  }
}
```

In some scenarios, the Terraform provider may attempt to apply only the default settings. A virtual machine deployed directly from an OVF/OVA may not match the OVF specification. For example, if the `scsi_type` option is not included in a `vsphere_virtual_machine` resource, the provider will apply a default value of `pvscsi` and the virtual machine may not boot. In this scenario, use the `vsphere_ovf_vm_template` data source to parse the OVF properties and use the property value as parameters for the `vsphere_virtual_machine` resource.

The following example demonstrates a scenario deploying a nested ESXi host from an OVF/OVA, using the remote URL and local path options.

**Example**:

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
  name              = "ubuntu-xx.xx-server-cloudimg-amd64.ova"
  disk_provisioning = "thin"
  resource_pool_id  = data.vsphere_resource_pool.default.id
  datastore_id      = data.vsphere_datastore.datastore.id
  remote_ovf_url    = "https://cloud-images.ubuntu.com/releases/xx.xx/release/ubuntu-xx.xx-server-cloudimg-amd64.ova"
  ovf_network_map = {
    "VM Network" : data.vsphere_network.network.id
  }
}

## Local OVF/OVA Source
data "vsphere_ovf_vm_template" "ovfLocal" {
  name              = "ubuntu-xx.xx-server-cloudimg-amd64.ova"
  disk_provisioning = "thin"
  resource_pool_id  = data.vsphere_resource_pool.default.id
  datastore_id      = data.vsphere_datastore.datastore.id
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
    remote_ovf_url  = "https://cloud-images.ubuntu.com/releases/xx.xx/release/ubuntu-xx.xx-server-cloudimg-amd64.ova"
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

### Cloning from an OVF/OVA with vApp Properties

This alternate example illustrates how to clone a virtual machine from a template that originated from an OVF/OVA. This leverages the resource's [vApp properties](#using-vapp-properties-for-ovf-ova-configuration) capabilities to set appropriate keys that control various configuration settings on the virtual machine or virtual appliance. In this scenario, using `customize` is not recommended as the functionality tends to overlap.

**Example**:

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

data "vsphere_virtual_machine" "template_from_ovf" {
  name          = "ubuntu-server-template-from-ova"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_virtual_machine" "vm" {
  name             = "foo"
  resource_pool_id = data.vsphere_compute_cluster.cluster.resource_pool_id
  datastore_id     = data.vsphere_datastore.datastore.id
  num_cpus         = 2
  memory           = 1024
  guest_id         = data.vsphere_virtual_machine.template.guest_id
  scsi_type        = data.vsphere_virtual_machine.template.scsi_type
  network_interface {
    network_id   = data.vsphere_network.network.id
    adapter_type = data.vsphere_virtual_machine.template.network_interface_types[0]
  }
  disk {
    name             = "Hard Disk 1"
    size             = data.vsphere_virtual_machine.template_from_ovf.disks.0.size
    thin_provisioned = data.vsphere_virtual_machine.template_from_ovf.disks.0.thin_provisioned
  }
  clone {
    template_uuid = data.vsphere_virtual_machine.template_from_ovf.id
  }
  vapp {
    properties = {
      "guestinfo.hostname"  = "foo.example.com",
      "guestinfo.ipaddress" = "172.16.11.101",
      "guestinfo.netmask"   = "255.255.255.0",
      "guestinfo.gateway"   = "172.16.11.1",
      "guestinfo.dns"       = "172.16.11.4",
      "guestinfo.domain"    = "example.com",
      "guestinfo.ntp"       = "ntp.example.com",
      "guestinfo.password"  = "VMware1!",
      "guestinfo.ssh"       = "True"
    }
  }
}
```

### Using vSphere Storage DRS

The `vsphere_virtual_machine` resource also supports vSphere Storage DRS, allowing the assignment of virtual machines to datastore clusters. When assigned to a datastore cluster, changes to a virtual machine's underlying datastores are ignored unless disks drift outside of the datastore cluster. Note that the [`vsphere_datastore_cluster`][tf-vsphere-datastore-cluster-resource] resource also exists to allow for management of datastore clusters using the Terraform provider.

The following example demonstrates the use of the [`vsphere_datastore_cluster`] data source[tf-vsphere-datastore-cluster-data-source], and the [`datastore_cluster_id`](#datastore_cluster_id) configuration setting.

[tf-vsphere-datastore-cluster-resource]: /docs/providers/vsphere/r/datastore_cluster.html
[tf-vsphere-datastore-cluster-data-source]: /docs/providers/vsphere/d/datastore_cluster.html

~> **NOTE:** When managing datastore clusters, member datastores, and virtual machines within the same Terraform configuration, race conditions can apply. This is because datastore clusters must be created before datastores can be assigned to them, and the respective `vsphere_virtual_machine` resources will no longer have an implicit dependency on the specific datastore resources. Use [`depends_on`][tf-docs-depends-on] to create an explicit dependency on the datastores in the cluster, or manage datastore clusters and datastores in a separate configuration.

[tf-docs-depends-on]: /docs/configuration/resources.html#depends_on

**Example**:

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "datastore-cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_compute_cluster" "cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_network" "network" {
  name          = "VM Network"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_virtual_machine" "vm" {
  name                 = "foo"
  resource_pool_id     = data.vsphere_compute_cluster.cluster.resource_pool_id
  datastore_cluster_id = data.vsphere_datastore_cluster.datastore_cluster.id
  num_cpus             = 1
  memory               = 1024
  guest_id             = "otherLinux64Guest"
  network_interface {
    network_id = data.vsphere_network.network.id
  }
  disk {
    label = "Hard Disk 1"
    size  = 20
  }
}
```

## Argument Reference

The following arguments are supported:

### General Options

The following options are general virtual machine and provider workflow options:

* `alternate_guest_name` - (Optional) The guest name for the operating system when `guest_id` is `otherGuest` or `otherGuest64`.

* `annotation` - (Optional) A user-provided description of the virtual machine.

* `cdrom` - (Optional) A specification for a CD-ROM device on the virtual machine. See [CD-ROM options](#cd-rom-options) for more information.

* `clone` - (Optional) When specified, the virtual machine will be created as a clone of a specified template. Optional customization options can be submitted for the resource. See [creating a virtual machine from a template](#creating-a-virtual-machine-from-a-template) for more information.

* `extra_config_reboot_required` - (Optional) Allow the virtual machine to be rebooted when a change to `extra_config` occurs. Default: `true`.

* `custom_attributes` - (Optional) Map of custom attribute ids to attribute value strings to set for virtual machine. Please refer to the [`vsphere_custom_attributes`][docs-setting-custom-attributes] resource for more information on setting custom attributes.

[docs-setting-custom-attributes]: /docs/providers/vsphere/r/custom_attribute.html#using-custom-attributes-in-a-supported-resource

~> **NOTE:** Custom attributes requires vCenter Server and is not supported on direct ESXi host connections.

* `datastore_id` - (Optional) The [managed object reference ID][docs-about-morefs] of the datastore in which to place the virtual machine. The virtual machine configuration files is placed here, along with any virtual disks that are created where a datastore is not explicitly specified. See the section on [virtual machine migration](#virtual-machine-migration) for more information on modifying this value.

* `datastore_cluster_id` - (Optional) The [managed object reference ID][docs-about-morefs] of the datastore cluster in which to place the virtual machine. This setting applies to entire virtual machine and implies that you wish to use vSphere Storage DRS with the virtual machine. See the section on [virtual machine migration](#virtual-machine-migration) for more information on modifying this value.

~> **NOTE:** One of `datastore_id` or `datastore_cluster_id` must be specified.

~> **NOTE:** Use of `datastore_cluster_id` requires vSphere Storage DRS to be enabled on the specified datastore cluster.

~> **NOTE:** The `datastore_cluster_id` setting applies to the entire virtual machine resource. You cannot assign individual individual disks to datastore clusters. In addition, you cannot use the [`attach`](#attach) setting to attach external disks on virtual machines that are assigned to datastore clusters.

* `datacenter_id` - (Optional) The datacenter ID. Required only when deploying an OVF/OVA template.

* `disk` - (Required) A specification for a virtual disk device on the virtual machine. See [disk options](#disk-options) for more information.

* `extra_config` - (Optional) Extra configuration data for the virtual machine. Can be used to supply advanced parameters not normally in configuration, such as instance metadata and userdata.

~> **NOTE:** Do not use `extra_config` when working with a template imported from OVF/OVA as your settings may be ignored. Use the `vapp` block `properties` section as described in [Using vApp Properties for OVF/OVA Configuration](#using-vapp-properties-for-ovf-ova-configuration).

* `firmware` - (Optional) The firmware for the virtual machine. One of `bios` or `efi`.

* `folder` - (Optional) The path to the virtual machine folder in which to place the virtual machine, relative to the datacenter path (`/<datacenter-name>/vm`).  For example, `/dc-01/vm/foo`

* `guest_id` - (Optional) The guest ID for the operating system type. Default: `otherGuest64`.

~> **NOTE:** To get a list of supported guest operating system identifiers for your ESXi host, run the following PowerShell command using `VMware.PowerCLI`:

  ```powershell
  Connect-VIServer -Server "vcenter.example.com" -User "administrator@vsphere.local" -Password "password"
  $esxiHost = Get-VMHost -Name "esxi-01.example.com"
  $environmentBrowser = Get-View -Id $esxiHost.ExtensionData.Parent.ExtensionData.ConfigManager.EnvironmentBrowser
  $vmxVersion = ($environmentBrowser.QueryConfigOptionDescriptor() | Where-Object DefaultConfigOption).Key
  $osDescriptor = $environmentBrowser.QueryConfigOption($vmxVersion, $null).GuestOSDescriptor
  $osDescriptor | Select-Object Id, Fullname
  ```

* `hardware_version` - (Optional) The hardware version number. Allows versions within ranges: 4, 7-11, 13-15, 17-22. The hardware version cannot be downgraded. See virtual machine hardware [versions][virtual-machine-hardware-versions] and [compatibility][virtual-machine-hardware-compatibility] for more information on supported settings.

[virtual-machine-hardware-versions]: https://knowledge.broadcom.com/external/article?articleNumber=315655
[virtual-machine-hardware-compatibility]: https://knowledge.broadcom.com/external/article?articleNumber=312100

* `host_system_id` - (Optional) The [managed object reference ID][docs-about-morefs] of a host on which to place the virtual machine. See the section on [virtual machine migration](#virtual-machine-migration) for more information on modifying this value. When using a vSphere cluster, if a `host_system_id` is not supplied, vSphere will select a host in the cluster to place the virtual machine, according to any defaults or vSphere DRS placement policies.

* `name` - (Required) The name of the virtual machine.

* `network_interface` - (Required) A specification for a virtual NIC on the virtual machine. See [network interface options](#network-interface-options) for more information.

* `pci_device_id` - (Optional) List of host PCI device IDs in which to create PCI passthroughs.

~> **NOTE:** Cloning requires vCenter Server and is not supported on direct ESXi host connections.

* `ovf_deploy` - (Optional) When specified, the virtual machine will be deployed from the provided OVF/OVA template. See [creating a virtual machine from an OVF/OVA template](#creating-a-virtual-machine-from-an-ovf-ova-template) for more information.

* `replace_trigger` - (Optional) Triggers replacement of resource whenever it changes.

For example, `replace_trigger = sha256(format("%s-%s",data.template_file.cloud_init_metadata.rendered,data.template_file.cloud_init_userdata.rendered))` will fingerprint the changes in cloud-init metadata and userdata templates. This will enable a replacement of the resource whenever the dependant template renders a new configuration. (Forces a replacement.)

* `resource_pool_id` - (Required) The [managed object reference ID][docs-about-morefs] of the resource pool in which to place the virtual machine. See the [Virtual Machine Migration](#virtual-machine-migration) section for more information on modifying this value.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

~> **NOTE:** All clusters and standalone hosts have a default root resource pool. This resource argument does not directly accept the cluster or standalone host resource. For more information, see the section on [Specifying the Root Resource Pool][docs-resource-pool-cluster-default] in the `vsphere_resource_pool` data source documentation on using the root resource pool.

[docs-resource-pool-cluster-default]: /docs/data-sources/resource_pool#specifying-the-root-resource-pool-for-a-standalone-esxi-host

* `scsi_type` - (Optional) The SCSI controller type for the virtual machine. One of `lsilogic` (LSI Logic Parallel), `lsilogic-sas` (LSI Logic SAS) or `pvscsi` (VMware Paravirtual). Default: `pvscsi`.

* `scsi_bus_sharing` - (Optional) The type of SCSI bus sharing for the virtual machine SCSI controller. One of `physicalSharing`, `virtualSharing`, and `noSharing`. Default: `noSharing`.

* `storage_policy_id` - (Optional) The ID of the storage policy to assign to the home directory of a virtual machine.

* `tags` - (Optional) The IDs of any tags to attach to this resource. Please refer to the [`vsphere_tag`][docs-applying-tags] resource for more information on applying tags to virtual machine resources.

[docs-applying-tags]: /docs/providers/vsphere/r/tag.html#using-tags-in-a-supported-resource

~> **NOTE:** Tagging support is unsupported on direct ESXi host connections and requires vCenter Server instance.

* `vapp` - (Optional) Used for vApp configurations. The only sub-key available is `properties`, which is a key/value map of properties for virtual machines imported from and OVF/OVA. See [Using vApp Properties for OVF/OVA Configuration](#using-vapp-properties-for-ovf-ova-configuration) for more information.

### CPU and Memory Options

The following options control CPU and memory settings on a virtual machine:

* `cpu_hot_add_enabled` - (Optional) Allow CPUs to be added to the virtual machine while it is powered on.

* `cpu_hot_remove_enabled` - (Optional) Allow CPUs to be removed to the virtual machine while it is powered on.

* `memory` - (Optional) The memory size to assign to the virtual machine, in MB. Default: `1024` (1 GB).

* `memory_hot_add_enabled` - (Optional) Allow memory to be added to the virtual machine while it is powered on.

~> **NOTE:** CPU and memory hot add options are not available on all guest operating systems. Please refer to the [VMware Guest OS Compatibility Guide][vmware-docs-compat-guide] to which settings are allow for your guest operating system. In addition, at least one `terraform apply` must be run before you are able to use CPU and memory hot add.

[vmware-docs-compat-guide]: https://partnerweb.vmware.com/comp_guide2/pdf/VMware_GOS_Compatibility_Guide.pdf

~> **NOTE:** For Linux 64-bit guest operating systems with less than or equal to 3GB, the virtual machine must powered off to add memory beyond 3GB. Subsequent hot add of memory does not require the virtual machine to be powered-off to apply the plan. Please refer to [KB 2008405][kb-2008405].

[kb-2008405]: https://knowledge.broadcom.com/external/article?articleNumber=343190

* `num_cores_per_socket` - (Optional) The number of cores per socket in the virtual machine. The number of vCPUs on the virtual machine will be `num_cpus` divided by `num_cores_per_socket`. If specified, the value supplied to `num_cpus` must be evenly divisible by this value. If `0` is specified a value will be determined automatically at power on. Default: `1`.

* `num_cores_per_numa_node` - (Optional) The number of cores per NUMA node in the virtual machine. If `0` is specified a value will be determined automatically at power on.

* `num_cpus` - (Optional) The total number of virtual processor cores to assign to the virtual machine. Default: `1`.

### Boot Options

The following options control boot settings on a virtual machine:

* `boot_delay` - (Optional) The number of milliseconds to wait before starting the boot sequence. The default is no delay.

* `boot_retry_delay` - (Optional) The number of milliseconds to wait before retrying the boot sequence. This option is only valid if `boot_retry_enabled` is `true`. Default: `10000` (10 seconds).

* `boot_retry_enabled` - (Optional) If set to `true`, a virtual machine that fails to boot will try again after the delay defined in `boot_retry_delay`. Default: `false`.

* `efi_secure_boot_enabled` - (Optional) Use this option to enable EFI secure boot when the `firmware` type is set to is `efi`. Default: `false`.

### VMware Tools Options

The following options control VMware Tools settings on the virtual machine:

* `sync_time_with_host` - (Optional) Enable the guest operating system to synchronization its clock with the host when the virtual machine is powered on or resumed. Requires VMware Tools to be installed. Default: `true`.

~> **NOTE:** `sync_time_with_host` is only available on vSphere 7.0 Update 1 and later.

* `sync_time_with_host_periodically` - (Optional) Enable the guest operating system to periodically synchronize its clock with the host. Default: `false`.

~> **NOTE:** `sync_time_with_host_periodically` is only available on vSphere 7.0 Update 1 and later. On previous versions, setting `sync_time_with_host` is will enable periodic synchronization.

* `run_tools_scripts_after_power_on` - (Optional) Enable post-power-on scripts to run when VMware Tools is installed. Default: `true`.

* `run_tools_scripts_after_resume` - (Optional) Enable ost-resume scripts to run when VMware Tools is installed. Default: `true`.

* `run_tools_scripts_before_guest_reboot` - (Optional) Enable pre-reboot scripts to run when VMware Tools is installed. Default: `false`.

* `run_tools_scripts_before_guest_shutdown` - (Optional) Enable pre-shutdown scripts to run when VMware Tools is installed. Default: `true`.

* `run_tools_scripts_before_guest_standby` - (Optional) Enable pre-standby scripts to run when VMware Tools is installed. Default: `true`.

* `tools_upgrade_policy` - (Optional) Enable automatic upgrade of the VMware Tools version when the virtual machine is rebooted. If necessary, VMware Tools is upgraded to the latest version supported by the host on which the virtual machine is running. Requires VMware Tools to be installed. One of `manual` or `upgradeAtPowerCycle`. Default: `manual`.

### Resource Allocation Options

The following options control CPU and memory allocation on the virtual machine. Please note that the resource pool in which a virtual machine is placed may affect these options.

The options are:

* `cpu_limit` - (Optional) The maximum amount of CPU (in MHz) that the virtual machine can consume, regardless of available resources. The default is no limit.

* `cpu_reservation` - (Optional) The amount of CPU (in MHz) that the virtual machine is guaranteed. The default is no reservation.

* `cpu_share_level` - (Optional) The allocation level for the virtual machine CPU resources. One of `high`, `low`, `normal`, or `custom`. Default: `custom`.

* `cpu_share_count` - (Optional) The number of CPU shares allocated to the virtual machine when the `cpu_share_level` is `custom`.

* `memory_limit` - (Optional) The maximum amount of memory (in MB) that th virtual machine can consume, regardless of available resources. The default is no limit.

* `memory_reservation` - (Optional) The amount of memory (in MB) that the virtual machine is guaranteed. The default is no reservation.

* `memory_reservation_locked_to_max` - (Optional) If set true, memory resource reservation for this virtual machine will always be equal to the virtual machine's memory size.

* `memory_share_level` - (Optional) The allocation level for the virtual machine memory resources. One of `high`, `low`, `normal`, or `custom`. Default: `custom`.

* `memory_share_count` - (Optional) The number of memory shares allocated to the virtual machine when the `memory_share_level` is `custom`.

### Advanced Options

The following options control advanced operation of the virtual machine, or control various parts of Terraform workflow, and should not need to be modified during basic operation of the resource. Only change these options if they are explicitly required, or if you are having trouble with Terraform's default behavior.

The options are:

* `cpu_performance_counters_enabled` - (Optional) Enable CPU performance counters on the virtual machine. Default: `false`.

* `enable_disk_uuid` - (Optional) Expose the UUIDs of attached virtual disks to the virtual machine, allowing access to them in the guest. Default: `false`.

* `enable_logging` - (Optional) Enable logging of virtual machine events to a log file stored in the virtual machine directory. Default: `false`.

* `ept_rvi_mode` - (Optional) The EPT/RVI (hardware memory virtualization) setting for the virtual machine. One of `automatic`, `on`, or `off`. Default: `automatic`.

* `force_power_off` - (Optional) If a guest shutdown failed or times out while updating or destroying (see [`shutdown_wait_timeout`](#shutdown_wait_timeout)), force the power-off of the virtual machine. Default: `true`.

* `hv_mode` - (Optional) The hardware virtualization (non-nested) setting for the virtual machine. One of `hvAuto`, `hvOn`, or `hvOff`. Default: `hvAuto`.

* `ide_controller_count` - (Optional) The number of IDE controllers that the virtual machine. This directly affects the number of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers. Default: `2`.

* `ignored_guest_ips` - (Optional) List of IP addresses and CIDR networks to ignore while waiting for an available IP address using either of the waiters. Any IP addresses in this list will be ignored so that the waiter will continue to wait for a valid IP address. Default: `[]`.

* `latency_sensitivity` - (Optional) Controls the scheduling delay of the virtual machine. Use a higher sensitivity for applications that require lower latency, such as VOIP, media player applications, or applications that require frequent access to mouse or keyboard devices. One of `low`, `normal`, `medium`, or `high`.

~> **NOTE:** On higher sensitivities, you may need to adjust the [`memory_reservation`](#memory_reservation) to the full amount of memory provisioned for the virtual machine.

* `migrate_wait_timeout` - (Optional) The amount of time, in minutes, to wait for a virtual machine migration to complete before failing. Default: `10` minutes. See the section on [virtual machine migration](#virtual-machine-migration) for more information.

* `nested_hv_enabled` - (Optional) Enable nested hardware virtualization on the virtual machine, facilitating nested virtualization in the guest operating system. Default: `false`.

* `sata_controller_count` - (Optional) The number of SATA controllers that the virtual machine. This directly affects the number of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers. Default: `0`.

* `nvme_controller_count` - (Optional) The number of NVMe controllers that the virtual machine. This directly affects the number of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers. Default: `0`.

* `scsi_controller_count` - (Optional) The number of SCSI controllers on the virtual machine. This setting directly affects the number of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers. Default: `1`.

* `shutdown_wait_timeout` - (Optional) The amount of time, in minutes, to wait for a graceful guest shutdown when making necessary updates to the virtual machine. If `force_power_off` is set to `true`, the virtual machine will be forced to power-off after the timeout, otherwise an error is returned. Default: `3` minutes.

* `swap_placement_policy` - (Optional) The swap file placement policy for the virtual machine. One of `inherit`, `hostLocal`, or `vmDirectory`. Default: `inherit`.

* `vbs_enabled` - (Optional) Enable Virtualization Based Security. Requires `firmware` to be `efi`. In addition, `vvtd_enabled`, `nested_hv_enabled`, and `efi_secure_boot_enabled` must all have a value of `true`. Default: `false`.

* `vvtd_enabled` - (Optional) Enable Intel Virtualization Technology for Directed I/O for the virtual machine (_I/O MMU_ in the vSphere Client). Default: `false`.

* `wait_for_guest_ip_timeout` - (Optional) The amount of time, in minutes, to wait for an available guest IP address on the virtual machine. This should only be used if the version VMware Tools does not allow the [`wait_for_guest_net_timeout`](#wait_for_guest_net_timeout) waiter to be used. A value less than `1` disables the waiter. Default: `0`.

* `wait_for_guest_net_routable` - (Optional) Controls whether or not the guest network waiter waits for a routable address. When `false`, the waiter does not wait for a default gateway, nor are IP addresses checked against any discovered default gateways as part of its success criteria. This property is ignored if the [`wait_for_guest_ip_timeout`](#wait_for_guest_ip_timeout) waiter is used. Default: `true`.

* `wait_for_guest_net_timeout` - (Optional) The amount of time, in minutes, to wait for an available guest IP address on the virtual machine. Older versions of VMware Tools do not populate this property. In those cases, this waiter can be disabled and the [`wait_for_guest_ip_timeout`](#wait_for_guest_ip_timeout) waiter can be used instead. A value less than `1` disables the waiter. Default: `5` minutes.

### Disk Options

Virtual disks are managed by adding one or more instance of the `disk` block.

At a minimum, both the `label` and `size` attributes must be provided.  The `unit_number` is required for any disk other than the first, and there must be at least one resource with the implicit number of `0`.

The following example demonstrates an abridged multi-disk configuration:

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  disk {
    label = "Hard Disk 1"
    size  = "10"
  }
  disk {
    label       = "Hard Disk 2"
    size        = "100"
    unit_number = 1
  }
  # ... other configuration ...
}
```

The options are:

* `label` - (Required) A label for the virtual disk. Forces a new disk, if changed.

~> **NOTE:** It's recommended that you set the disk label to a format matching `diskN`, where `N` is the number of the disk, starting from disk number 0. This will ensure that your configuration is compatible when importing a virtual machine. See the section on [importing](#importing) for more information.

~> **NOTE:** Do not choose a label that starts with `orphaned_disk_` (_e.g._ `orphaned_disk_0`), as this prefix is reserved for disks that the provider does not recognize. Such as disks that are attached externally. The Terraform provider will issue
an error if you try to label a disk with this prefix.

* `size` - (Required) The size of the disk, in GB. Must be a whole number.

~> **NOTE:** Use the [`api_timeout`](/docs/providers/vsphere/index.html#argument-reference) option to increase the default timeout when creating virtual machines with large disk sizes and avoid "context deadline exceeded" errors.

* `unit_number` - (Optional) The disk number on the storage bus. The maximum value for this setting is the value of the controller count times the controller capacity (15 for SCSI, 30 for SATA, and 2 for IDE). Duplicate unit numbers are not allowed. Default `0`, for which one disk must be set to.

* `datastore_id` - (Optional) The [managed object reference ID][docs-about-morefs] for the datastore on which the virtual disk is placed. The default is to use the datastore of the virtual machine. See the section on [virtual machine migration](#virtual-machine-migration) for information on modifying this value.

~> **NOTE:** Datastores cannot be assigned to individual disks when [`datastore_cluster_id`](#datastore_cluster_id) is used.

* `attach` - (Optional) Attach an external disk instead of creating a new one. Implies and conflicts with `keep_on_remove`. If set, you cannot set `size`, `eagerly_scrub`, or `thin_provisioned`. Must set `path` if used.

~> **NOTE:** External disks cannot be attached when [`datastore_cluster_id`](#datastore_cluster_id) is used.

* `path` - (Optional) When using `attach`, this parameter controls the path of a virtual disk to attach externally. Otherwise, it is a computed attribute that contains the virtual disk filename.

* `keep_on_remove` - (Optional) Keep this disk when removing the device or destroying the virtual machine. Default: `false`.

* `disk_mode` - (Optional) The mode of this this virtual disk for purposes of writes and snapshots. One of `append`, `independent_nonpersistent`, `independent_persistent`, `nonpersistent`, `persistent`, or `undoable`. Default: `persistent`. For more information on these option, please refer to the [product documentation][vmware-docs-disk-mode].

[vmware-docs-disk-mode]: https://developer.broadcom.com/xapis/virtual-infrastructure-json-api/latest/data-structures/VirtualDiskMode/enum/

* `eagerly_scrub` - (Optional) If set to `true`, the disk space is zeroed out when the virtual machine is created. This will delay the creation of the virtual disk. Cannot be set to `true` when `thin_provisioned` is `true`.  See the section on [picking a disk type](#picking-a-disk-type) for more information.  Default: `false`.

* `thin_provisioned` - (Optional) If `true`, the disk is thin provisioned, with space for the file being allocated on an as-needed basis. Cannot be set to `true` when `eagerly_scrub` is `true`. See the section on [selecting a disk type](#selecting-a-disk-type) for more information. Default: `true`.

* `disk_sharing` - (Optional) The sharing mode of this virtual disk. One of `sharingMultiWriter` or `sharingNone`. Default: `sharingNone`.

* `write_through` - (Optional) If `true`, writes for this disk are sent directly to the filesystem immediately instead of being buffered. Default: `false`.

* `io_limit` - (Optional) The upper limit of IOPS that this disk can use. The default is no limit.

* `io_reservation` - (Optional) The I/O reservation (guarantee) for the virtual disk has, in IOPS.  The default is no reservation.

* `io_share_level` - (Optional) The share allocation level for the virtual disk. One of `low`, `normal`, `high`, or `custom`. Default: `normal`.

* `io_share_count` - (Optional) The share count for the virtual disk when the share level is `custom`.

* `storage_policy_id` - (Optional) The UUID of the storage policy to assign to the virtual disk.

* `controller_type` - (Optional) The type of storage controller to attach the  disk to. Can be `scsi`, `sata`, `nvme` or `ide`. You must have the appropriate number of controllers enabled for the selected type. Default `scsi`.

#### Computed Disk Attributes

* `uuid` - The UUID of the virtual disk VMDK file. This is used to track the virtual disk on the virtual machine.

#### Virtual Disk Provisioning Policies

The `eagerly_scrub` and `thin_provisioned` options control the virtual disk space allocation type. These appear in vSphere as a unified enumeration of options, the equivalents of which are explained below. The provider configuration defaults to thin provision.

The options are:

* **Thick Provision Lazy Zeroed**:

  Both `eagerly_scrub` and `thin_provisioned` should be set to `false`.

* **Thick Provision Eager Zeroed**:

  `eagerly_scrub` should be set to `true`, and `thin_provisioned` should be set to `false`.

* **Thin Provision**:

  `eagerly_scrub` should be set to `false`, and `thin_provisioned` should be set to `true`.

For more information on each provisioning policy, please refer to the [product documentation][docs-vmware-vm-disk-provisioning].

[docs-vmware-vm-disk-provisioning]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-single-host-management-vmware-host-client-8-0/virtual-machine-management-with-the-vsphere-host-client-vSphereSingleHostManagementVMwareHostClient/configuring-virtual-machines-in-the-vsphere-host-client-vSphereSingleHostManagementVMwareHostClient/virtual-disk-configuration-vSphereSingleHostManagementVMwareHostClient/about-virtual-disk-provisioning-policies-vSphereSingleHostManagementVMwareHostClient.html

~> **NOTE:** Not all provisioning policies are available for all datastore types. Setting and inappropriate provisioning policy for a datastore type may result in a successful initial apply as vSphere will silently correct the options; however, subsequent plans will fail with an appropriate error message until the settings are corrected.

~> **NOTE:** A disk type cannot be changed once set.

### Network Interface Options

Network interfaces are managed by adding one or more instance of the `network_interface` block.

Interfaces are assigned to devices in the order declared in the configuration and may have implications for different guest operating systems.

**Example**:

```hcl
data "vsphere_network" "routable" {
  name          = var.routable_network
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

data "vsphere_network" "non_routable" {
  name          = var.non_routable_network
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  network_interface {
    network_id = data.vsphere_network.routable.id
  }
  network_interface {
    network_id = data.vsphere_network.non_routable.id
  }
  # ... other configuration ...
}

clone {
  template_uuid = data.vsphere_virtual_machine.template.id
  customize {
    linux_options {
      host_name = "foo"
      domain    = "example.com"
    }
    network_interface {
      ipv4_address = "10.0.0.10"
      ipv4_netmask = 24
    }
    network_interface {
      ipv4_address = "172.16.0.10"
      ipv4_netmask = 24
    }
    ipv4_gateway = "10.0.0.1"
  }
}
```

In the example above, the first interface is assigned to the `routable` network (_e.g._, `10.0.0.10`) based on the order of sequential assignment. The second interface is assigned to the `non_routable` network (_e.g._, `172.16.0.10`).

On some Linux distributions, the first interface may be presented as `eth0` and the second may be presented as `eth1`.

The options are:

* `network_id` - (Required) The [managed object reference ID][docs-about-morefs] of the network on which to connect the virtual machine network interface.

* `adapter_type` - (Optional) The network interface type. One of `e1000`, `e1000e`, `sriov`, or `vmxnet3`. Default: `vmxnet3`.

* `use_static_mac` - (Optional) If true, the `mac_address` field is treated as a static MAC address and set accordingly. Setting this to `true` requires `mac_address` to be set. Default: `false`.

* `mac_address` - (Optional) The MAC address of the network interface. Can only be manually set if `use_static_mac` is `true`. Otherwise, the value is computed and presents the assigned MAC address for the interface.

* `bandwidth_limit` - (Optional) The upper bandwidth limit of the network interface, in Mbits/sec. The default is no limit. Ignored if `adapter_type` is set to `sriov`.

* `bandwidth_reservation` - (Optional) The bandwidth reservation of the network interface, in Mbits/sec. The default is no reservation.

* `bandwidth_share_level` - (Optional) The bandwidth share allocation level for the network interface. One of `low`, `normal`, `high`, or `custom`. Default: `normal`. Ignored if `adapter_type` is set to `sriov`.

* `bandwidth_share_count` - (Optional) The share count for the network interface when the share level is `custom`. Ignored if `adapter_type` is set to `sriov`.

* `ovf_mapping` - (Optional) Specifies which NIC in an OVF/OVA the `network_interface` should be associated. Only applies at creation when deploying from an OVF/OVA.

### Video Card Options

The virtual video card is managed by adding a `video_card` block.
A virtual machine always has exactly one virtual video card (this does not apply to passthrough devices). This device
is created automatically even if it is not included in the configuration.

This means that
* Creating a virtual machine with a `video_card` block in its configuration creates the video card device with the specified settings.
* Adding a `video_card` block to an existing machine modifies the existing device. 
* Removing the `video_card` block from a virtual machine does not delete the device.

Using the options under the `graphics_3d` section may result in a failure to power on the virtual machine.
This will happen if vCenter is unable to find a host which supports these options.

```hcl
resource "vsphere_virtual_machine" "vm" {
  name             = "example-vm"
  #...

  num_cpus = 2
  memory   = 2048
  #...

  video_card {
    num_displays = 2
    total_video_memory = 128
    graphics_3d {
      renderer = "hardware"
      memory = 512
    }
  }
}
```

The options are:

* `num_displays` - The number of supported displays.

* `total_video_memory` - The total video memory buffer in megabytes.

* `graphics_3d` - 3D graphics options.

  * `renderer` - The 3D renderer - software, hardware or automatic.
  
  * `memory` - The dedicated 3D graphics memory in megabytes.

#### Using SR-IOV Network Interfaces

In order to attach your virtual machine to an SR-IOV network interface,
there are a few requirements

* SR-IOV network interfaces must be declared after all non-SRIOV network interfaces.

* The target host must be known, if creating a VM from scratch, this means setting the `host_system_id` option.

* SR-IOV must be enabled on the relevant physical adapter on the host.

* The `memory_reservation` must be fully set (that is, equal to the `memory`) for the VM.

* The `network_interface` sub-resource takes a `physical_function` argument:
  * This **must** be set if your adapter type is `sriov`.
  * This **must not** be set if your adapter type is not `sriov`.
  * This can be found by navigating to the relevant host in the vSphere Client,
    going to the 'Configure' tab followed by 'Networking' then 'Physical adapters' and finding the
    relevant physical network adapter; one of the properties of the NIC is its PCI Location.
  * This is usually in the form of "0000:ab:cd.e"

* The `bandwidth_*` options on the network interface are not permitted.

* Adding, modifying, and deleting SR-IOV NICs is supported but requires a VM restart.

* Modifying the number of non-SR-IOV (_e.g._, VMXNET3) interfaces when there are SR-IOV interfaces existing is
  explicitly blocked (as the provider does not support modifying an interface at the same index from
  non-SR-IOV to SR-IOV or vice-versa). To work around this delete all SRIOV NICs for one terraform apply, and re-add
  them with any change to the number of non-SRIOV NICs on a second terraform apply.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  host_system_id     = data.vsphere_host.host.id
  memory             = var.memory
  memory_reservation = var.memory
  network_interface {
    network_id        = data.vsphere_network.network.id
    adapter_type      = "sriov"
    physical_function = "0000:3b:00.1"
  }
  # ... other network_interfaces...
}
```

### CD-ROM Options

A CD-ROM device is managed by adding an instance of the `cdrom` block.

Up to two virtual CD-ROM devices can be created and attached to the virtual machine. If adding multiple CD-ROM devices, add each device as a separate `cdrom` block. The resource supports attaching a CD-ROM from either a datastore ISO or using a remote client device.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  cdrom {
    datastore_id = data.vsphere_datastore.iso_datastore.id
    path         = "/Volume/Storage/ISO/foo.iso"
  }
  # ... other configuration ...
}
```

The options are:

* `client_device` - (Optional) Indicates whether the device should be backed by remote client device. Conflicts with `datastore_id` and `path`.

* `datastore_id` - (Optional) The datastore ID that on which the ISO is located. Required for using a datastore ISO. Conflicts with `client_device`.

* `path` - (Optional) The path to the ISO file. Required for using a datastore ISO. Conflicts with `client_device`.

~> **NOTE:** Either `client_device` (for a remote backed CD-ROM) or `datastore_id` and `path` (for a datastore ISO backed CD-ROM) are required to .

~> **NOTE:** Some CD-ROM drive types are not supported by this resource, such as pass-through devices. If these drives are present in a cloned template, or added outside of the provider, the desired state will be corrected to the defined device, or removed if no `cdrom` block is present.

### Virtual Device Computed Options

Virtual devices (`disk`, `network_interface`, and `cdrom`) all export the following attributes. These options help locate the device on subsequent application of the Terraform configuration.

The options are:

* `key` - The ID of the device within the virtual machine.

* `device_address` - An address internal to Terraform that helps locate the device when `key` is unavailable. This follows a convention of `CONTROLLER_TYPE:BUS_NUMBER:UNIT_NUMBER`. Example: `scsi:0:1` means device unit `1` on SCSI bus `0`.

## Creating a Virtual Machine from a Template

The `clone` block can be used to create a new virtual machine from an existing virtual machine or template. The resource supports both making a complete copy of a virtual machine, or cloning from a snapshot (also known as a linked clone).

See the section on [cloning and customization](#cloning-and-customization) for more information.

~> **NOTE:** Changing any option in `clone` after creation forces a new resource.

~> **NOTE:** Cloning requires vCenter Server and is not supported on direct ESXi host connections.

The options available in the `clone` block are:

* `template_uuid` - (Required) The UUID of the source virtual machine or template.

* `linked_clone` - (Optional) Clone the virtual machine from a snapshot or a template. Default: `false`.

* `timeout` - (Optional) The timeout, in minutes, to wait for the cloning process to complete. Default: 30 minutes.

* `customize` - (Optional) The customization spec for this clone. This allows the user to configure the virtual machine post-clone. For more details, see [virtual machine customizations](#virtual-machine-customizations).

### Virtual Machine Customizations

As part of the `clone` operation, a virtual machine can be [customized][vmware-docs-customize] to configure host, network, or licensing settings.

[vmware-docs-customize]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-virtual-machine-administration-guide-8-0/managing-virtual-machinesvsphere-vm-admin/customizing-guest-operating-systemsvsphere-vm-admin.html

To perform virtual machine customization as a part of the clone process,
specify the `customize` block with the respective customization options, nested within the `clone` block. Windows guests are customized using Sysprep, which will result in the machine SID being reset. Before using customization, check is that your source virtual machine meets the [requirements](https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-virtual-machine-administration-guide-8-0/managing-virtual-machinesvsphere-vm-admin/customizing-guest-operating-systemsvsphere-vm-admin.html) for guest OS customization on vSphere. See the section on [cloning and customization](#cloning-and-customization) for a usage synopsis.

The settings for `customize` are as follows:

#### Customization Timeout Settings

* `timeout` - (Optional) The time, in minutes, that the provider waits for customization to complete before failing. The default is `10` minutes. Setting the value to `0` or a negative value disables the waiter.

#### Network Interface Settings

These settings, which should be specified in nested `network_interface` blocks within [`customize`](#virtual-machine-customization) block, configure network interfaces on a per-interface basis and are matched up to [`network_interface`](#network-interface-options) devices in the order declared.

Static IP Address Example:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  network_interface {
    network_id = data.vsphere_network.public.id
  }
  network_interface {
    network_id = data.vsphere_network.private.id
  }
  clone {
    # ... other configuration ...
    customize {
      # ... other configuration ...
      network_interface {
        ipv4_address = "10.0.0.10"
        ipv4_netmask = 24
      }
      network_interface {
        ipv4_address = "172.16.0.10"
        ipv4_netmask = 24
      }
      ipv4_gateway = "10.0.0.1"
    }
  }
}
```

The first `network_interface` would be assigned to the `public` interface, and the second to the `private` interface.

To use DHCP, declare an empty `network_interface` block for each interface.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  network_interface {
    network_id = data.vsphere_network.public.id
  }
  network_interface {
    network_id = data.vsphere_network.private.id
  }
  clone {
    # ... other configuration ...
    customize {
      # ... other configuration ...
      network_interface {}
      network_interface {}
    }
  }
}
```

The options are:

* `dns_server_list` - (Optional) DNS servers for the network interface. Used by Windows guest operating systems, but ignored by Linux distribution guest operating systems. For Linux, please refer to the section on the [global DNS settings](#global-dns-settings).

* `dns_domain` - (Optional) DNS search domain for the network interface. Used by Windows guest operating systems, but ignored by Linux distribution guest operating systems. For Linux, please refer to the section on the [global DNS settings](#global-dns-settings).

* `ipv4_address` - (Optional) The IPv4 address assigned to the network adapter. If blank or not included, DHCP is used.

* `ipv4_netmask` The IPv4 subnet mask, in bits (_e.g._ `24` for 255.255.255.0).

* `ipv6_address` - (Optional) The IPv6 address assigned to the network adapter. If blank or not included, auto-configuration is used.

* `ipv6_netmask` - (Optional) The IPv6 subnet mask, in bits (_e.g._  `32`).

~> **NOTE:** The minimum setting for IPv4 in a customization specification is DHCP. If you are setting up an IPv6-exclusive network without DHCP, you may need to set [`wait_for_guest_net_timeout`](#wait_for_guest_net_timeout) to a high enough value to cover the DHCP timeout of your virtual machine, or disable by supplying a zero or negative value. Disabling `wait_for_guest_net_timeout` may result in IP addresses not being reported to any provisioners you may have configured on the resource.

#### Global Routing Settings

Virtual machine customization for the `vsphere_virtual_machine` resource does not take a per-interface gateway setting. Default routes are configured on a global basis. See the section on [network interface settings](#network-interface-settings) for more information.

The settings must match the IP address and netmask of at least one `network_interface` supplied to customization.

The options are:

* `ipv4_gateway` - (Optional) The IPv4 default gateway when using `network_interface` customization on the virtual machine.

* `ipv6_gateway` - (Optional) The IPv6 default gateway when using `network_interface` customization on the virtual machine.

#### Global DNS Settings

The following settings configure DNS globally, generally for Linux distribution guest operating systems. For Windows guest operating systems, this is performer per-interface. See the section on [network interface settings](#network-interface-settings) for more information.

* `dns_server_list` - The list of DNS servers to configure on the virtual machine.

* `dns_suffix_list` - A list of DNS search domains to add to the DNS configuration on the virtual machine.

#### Linux Customization Options

The settings in the `linux_options` block pertain to Linux distribution guest operating system customization. If you are customizing a Linux guest operating system, this section must be included.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  clone {
    # ... other configuration ...
    customize {
      # ... other configuration ...
      linux_options {
        host_name = "foo"
        domain    = "example.com"
      }
    }
  }
}
```

The options are:

* `host_name` - (Required) The host name for this machine. This, along with `domain`, make up the FQDN of the virtual machine.

* `domain` - (Required) The domain name for this machine. This, along with `host_name`, make up the FQDN of the virtual machine.

* `hw_clock_utc` - (Optional) Tells the operating system that the hardware clock is set to UTC. Default: `true`.

* `script_text` - (Optional) The customization script for the virtual machine that will be applied before and / or after guest customization. For more information on enabling and using a customization script, please refer to [VMware KB 74880][kb-74880]. The [Heredoc style][tf-heredoc-strings] of string literal is recommended.

[kb-74880]: https://knowledge.broadcom.com/external/article?articleNumber=313048
[tf-heredoc-strings]: https://www.terraform.io/language/expressions/strings#heredoc-strings

* `time_zone` - (Optional) Sets the time zone. For a list of possible combinations, please refer to [VMware KB 2145518][kb-2145518]. The default is UTC.

[kb-2145518]: https://knowledge.broadcom.com/external/article?articleNumber=320212

#### Windows Customization Options

The settings in the `windows_options` block pertain to Windows guest OS customization. If you are customizing a Windows operating system, this section must be included.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  clone {
    # ... other configuration ...
    customize {
      # ... other configuration ...
      windows_options {
        computer_name  = "foo"
        workgroup      = "BAR"
        admin_password = "VMware1!"
      }
    }
  }
}
```

The options are:

* `computer_name` - (Required) The computer name of the virtual machine.

* `admin_password` - (Optional) The administrator password for the virtual machine.

~> **NOTE:** `admin_password` is a sensitive field and will not be output on-screen, but is stored in state and sent to the virtual machine in plain text.

* `workgroup` - (Optional) The workgroup name for the virtual machine. One of this or `join_domain` must be included.

* `join_domain` - (Optional) The domain name in which to join  the virtual machine. One of this or `workgroup` must be included.

* `domain_ou` - (Optional) The `MachineObjectOU` which specifies the full LDAP path name of the OU to which the virtual machine belongs (_e.g._, OU=bar,OU=foo,DC=example,DC=com").

~> **NOTE:** `domain_ou` is only available on vSphere 8.0 Update 2 and later.

~> **NOTE:** `domain_ou` must **not** contain a spaces in the `MachineObjectOU` path (_e.g._, OU=foo bar,DC=example,DC=com").

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  clone {
    # ... other configuration ...
    customize {
      # ... other configuration ...
      windows_options {
        domain_ou = "OU=bar,OU=foo,DC=example,DC=com"
      }
    }
  }
}
```

* `domain_admin_user` - (Optional) The user account with administrative privileges to use to join the guest operating system to the domain. Required if setting `join_domain`.

* `domain_admin_password` - (Optional) The password user account with administrative privileges used to join the virtual machine to the domain. Required if setting `join_domain`.

~> **NOTE:** `domain_admin_password` is a sensitive field and will not be output on-screen, but is stored in state and sent to the virtual machine in plain text

* `full_name` - (Optional) The full name of the organization owner of the virtual machine. This populates the "user" field in the general Windows system information. Default: `Administrator`.

* `organization_name` - (Optional) The name of the organization for the virtual machine.  This option populates the "organization" field in the general Windows system information. Default: `Managed by Terraform`.

* `product_key` - (Optional) The product key for the virtual machine Windows guest operating system. The default is no key.

* `run_once_command_list` - (Optional) A list of commands to run at first user logon, after guest customization. Each run once command is limited by the API to 260 characters.

* `auto_logon` - (Optional) Specifies whether or not the virtual machine automatically logs on as Administrator. Default: `false`.

* `auto_logon_count` - (Optional) Specifies how many times the virtual machine should auto-logon the Administrator account when `auto_logon` is `true`. This option should be set accordingly to ensure that all of your commands that run in `run_once_command_list` can log in to run. Default: `1`.

* `time_zone` - (Optional) The time zone for the virtual machine. For a list of supported codes, please refer to the [MIcrosoft documentation][ms-docs-valid-sysprep-tzs]. The default is `85` (GMT/UTC).

[ms-docs-valid-sysprep-tzs]: https://msdn.microsoft.com/en-us/library/ms912391(v=winembedded.11).aspx

##### Using a Windows SysPrep Configuration

An alternative to the `windows_options` demonstrated above, you can provide a SysPrep `.XML` file using the `windows_sysprep_text` option. This option allows full control of the Windows guest operating system customization process.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  clone {
    # ... other configuration ...
    customize {
      # ... other configuration ...
      windows_sysprep_text = file("${path.module}/sysprep.xml")
    }
  }
}
```

~> **NOTE:** This option is mutually exclusive to `windows_options`. One must not be included if the other is specified.

### Creating a Virtual Machine from an OVF/OVA Template

The `ovf_deploy` block is used to create a new virtual machine from an OVF/OVA template from either local system or remote URL. While deploying, the virtual machines properties are taken from OVF properties and setting them in configuration file is not necessary.

See the [Deploying from OVF example](#deploying-vm-from-an-ovf-ova-template) for a usage synopsis.

~> **NOTE:** Changing options in the `ovf_deploy` block after creation forces a new resource.

The options available in the `ovf_deploy` block are:

* `allow_unverified_ssl_cert` - (Optional) Allow unverified SSL certificates while deploying OVF/OVA from a URL. Defaults `false`.

* `enable_hidden_properties` - (Optional) Allow properties with `ovf:userConfigurable=false` to be set. Defaults `false`.

* `local_ovf_path` - (Optional) The absolute path to the OVF/OVA file on the local system. When deploying from an OVF, ensure the necessary files, such as `.vmdk` and `.mf` files are also in the same directory as the `.ovf` file.

* `remote_ovf_url` - (Optional) URL to the OVF/OVA file.

~> **NOTE:** Either `local_ovf_path` or `remote_ovf_url` is required.

* `ip_allocation_policy` - (Optional) The IP allocation policy.

* `ip_protocol` - (Optional) The IP protocol.

* `disk_provisioning` - (Optional) The disk provisioning type. If set, all the disks included in the OVF/OVA will have the same specified policy. One of `thin`, `thick`, or `eagerZeroedThick`.
  * `thin`: Each disk is allocated and zeroed on demand as the space is used.
  * `thick`: Each disk is allocated at creation time and the space is zeroed on demand as the space is used.
  * `eagerZeroedThick`: Each disk is allocated and zeroed at creation time.

* `deployment_option` - (Optional) The key for the deployment option. If empty, the default option is selected.

* `ovf_network_map` - (Optional) The mapping of network identifiers from the OVF descriptor to a network UUID.

### Using vApp Properties for OVF/OVA Configuration

You can use the `properties` section of the `vapp` block to supply configuration parameters to a virtual machine cloned from a template that originated from an imported OVF/OVA file. Both GuestInfo and ISO transport methods are supported.

For templates that use ISO transport, a CD-ROM backed by a client device must be included.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  clone {
    template_uuid = data.vsphere_virtual_machine.template_from_ovf.id
  }
  cdrom {
    client_device = true
  }
  vapp {
    properties = {
      guestinfo.terraform.id = "foo"
    }
  }
}
```

See the section on [CD-ROM options](#cd-rom-options) for more information.

~> **NOTE:** The only supported usage path for vApp properties is for existing user-configurable keys. These generally come from an existing template created by importing an OVF or OVA file. You cannot set values for vApp properties on virtual machines created from scratch, virtual machines lacking a vApp configuration, or on property keys that do not exist.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  clone {
    template_uuid = data.vsphere_virtual_machine.template_from_ovf.id
  }
  vapp {
    properties = {
      guestinfo.terraform.id = "foo"
    }
  }
}
```

The vApp Properties for some OVF/OVA may require boolean values.

In Terraform a boolean is defined as `bool` with a value of either `true` or `false`.

**Example**: A boolean variable type for the Terraform provider configuration.

```hcl
variable "vsphere_insecure" {
  type        = bool
  description = "Allow insecure connections. Set to `true` for self-signed certificates."
  default     = false
}

provider "vsphere" {
  # ... other configuration ...
  allow_unverified_ssl = var.vsphere_insecure
}
```

However, for OVF properties, even though the type is boolean, the vApp Options in vSphere only accepts the values of `"True"` or `"False"`.

In these instances, it is recommended to define the variable as a string and pass the value in title case.

**Example**: A string variable for to pass to an OVF/OVA boolean OVF property.

```hcl
variable "ssh_enabled" {
  type        = string
  description = "Enable SSH on the virtual appliance. One of `True` or `False`."
  default     = "False"
}

resource "vsphere_virtual_machine" "vm" {
  # ... other configurations ...
  vapp {
    properties = {
      # ... other configurations ...
      "ssh_enabled" = var.ssh_enabled
    }
  }
}
```

### Additional Requirements for Cloning

When cloning from a template, there are additional requirements in both the resource configuration and source template:

* The virtual machine must not be powered on at the time of cloning.
* All disks on the virtual machine must be SCSI disks.
* You must specify at least the same number of `disk` devices as there are disks that exist in the template. These devices are ordered and lined up by the `unit_number` attribute. Additional disks can be added past this.
* The `size` of a virtual disk must be at least the same size as its counterpart disk in the source template.
* When using `linked_clone`, the `size`, `thin_provisioned`, and `eagerly_scrub` settings for each disk must be an exact match to the individual disk's counterpart in the source template.
* The storage controller count settings should be configured as necessary to cover all of the disks on the template. For best results, only configure this setting for the number of controllers you will need to cover your disk quantity and bandwidth needs, and configure your template accordingly. For most workloads, this setting should be kept at the default of `1` SCSI controller, and all disks in the template should reside on the single, primary controller.
* Some operating systems do not respond well to a change in disk controller type. Ensure that `scsi_type` is set to an exact match of the template's controller set. For maximum compatibility, make sure the SCSI controllers on the source template are all the same type.

You can use the [`vsphere_virtual_machine`][tf-vsphere-virtual-machine-ds] data source, which provides disk attributes, network interface types, SCSI bus types, and the guest ID of the source template, to return this information. See the section on [cloning and customization](#cloning-and-customization) for more information.

## Trusted Platform Module

When creating a virtual machine or cloning one from a template, you have the option to add a virtual Trusted Platform Module device. Refer to the requirements in the VMware vSphere [product documentation](https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-virtual-machine-administration-guide-8-0/configuring-virtual-machine-hardwarevsphere-vm-admin/securing-virtual-machines-with-virtual-trusted-platform-modulevsphere-vm-admin/vtpm-overviewvsphere-vm-admin.html).

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  vtpm {
    version = "2.0"
  }
  # ... other configuration ...
}
```

~> **NOTE:** Supported versions include 1.2 or 2.0.

## Virtual Machine Migration

The `vsphere_virtual_machine` resource supports live migration both on the host and storage level. You can migrate the virtual machine to another host, cluster, resource pool, or datastore. You can also migrate or pin a virtual disk to a specific datastore.

### Host, Cluster, and Resource Pool Migration

To migrate the virtual machine to another host or resource pool, change the `host_system_id` or `resource_pool_id` to the managed object IDs of the new host or resource pool. To change the virtual machine's cluster or standalone host, select a resource pool within the specific target.

The same rules apply for migration as they do for virtual machine creation - any host specified must contribute the resource pool supplied. When moving a virtual machine to a resource pool in another cluster (or standalone host), ensure that all hosts in the cluster (or the single standalone host) have access to the datastore on which the virtual machine is placed.

### Storage Migration

Storage migration can be done on two levels:

* Global datastore migration can be handled by changing the global `datastore_id` attribute. This triggers a storage migration for all disks that do not have an explicit `datastore_id` specified.
* When using Storage DRS through the `datastore_cluster_id` attribute, the entire virtual machine can be migrated from one datastore cluster to another by changing the value of this setting. In addition, when `datastore_cluster_id` is in use, any disks that drift to datastores outside of the datastore cluster via such actions as manual modification will be migrated back to the datastore cluster on the next apply.
* An individual `disk` device can be migrated by manually specifying the `datastore_id` in its configuration block. This also pins it to the specific datastore that is specified - if at a later time the virtual machine and any unpinned disks migrate to another host, the disk will stay on the specified datastore.

An example of datastore pinning is below. As long as the datastore in the `pinned_datastore` data source does not change, any change to the standard `vm_datastore` data source will not affect the data disk - the disk will stay where it is.

**Example**:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...
  datastore_id = data.vsphere_datastore.vm_datastore.id
  disk {
    label = "Hard Disk 1"
    size  = 10
  }
  disk {
    datastore_id = data.vsphere_datastore.pinned_datastore.id
    label        = "Hard Disk 2"
    size         = 100
    unit_number  = 1
  }
  # ... other configuration ...
}
```

#### Storage Migration Restrictions

You cannot migrate external disks added with the `attach` parameter. Typically, these disks are created and assigned to a datastore outside the scope of the `vsphere_virtual_machine` resource. For example, using the [`vsphere_virtual_disk`][tf-vsphere-virtual-disk] resource, management of the disks would render their configuration unstable.

[tf-vsphere-virtual-disk]: /docs/providers/vsphere/r/virtual_disk.html

## Virtual Machine Reboot

The virtual machine will be rebooted if any of the following parameters are changed:

* `alternate_guest_name`
* `cpu_hot_add_enabled`
* `cpu_hot_remove_enabled`
* `cpu_performance_counters_enabled`
* `disk.controller_type`
* `disk.unit_number`
* `disk.disk_mode`
* `disk.write_through`
* `disk.disk_sharing`
* `efi_secure_boot_enabled`
* `ept_rvi_mode`
* `enable_disk_uuid`
* `enable_logging`
* `extra_config`
* `firmware`
* `guest_id`
* `hardware_version`
* `hv_mode`
* `memory` -  When reducing the memory size, or when increasing the memory size and `memory_hot_add_enabled` is set to `false`
* `memory_hot_add_enabled`
* `nested_hv_enabled`
* `network_interface` - When deleting a network interface and VMware Tools is not running.
* `network_interface.adapter_type` - When VMware Tools is not running.
* `num_cores_per_socket`
* `pci_device_id`
* `run_tools_scripts_after_power_on`
* `run_tools_scripts_after_resume`
* `run_tools_scripts_before_guest_standby`
* `run_tools_scripts_before_guest_shutdown`
* `run_tools_scripts_before_guest_reboot`
* `swap_placement_policy`
* `tools_upgrade_policy`
* `vbs_enabled`
* `vvtd_enabled`
* `vtpm`

## Attribute Reference

The following attributes are exported on the base level of this resource:

* `id` - The UUID of the virtual machine.

* `reboot_required` - Value internal to Terraform used to determine if a configuration set change requires a reboot. This value is most useful during an update process and gets reset on refresh.

* `vmware_tools_status` - The state of  VMware Tools in the guest. This will determine the proper course of action for some device operations.

* `vmx_path` - The path of the virtual machine configuration file on the datastore in which the virtual machine is placed.

* `imported` - Indicates if the virtual machine resource has been imported, or if the state has been migrated from a previous version of the resource. It influences the behavior of the first post-import apply operation. See the section on [importing](#importing) below.

* `change_version` - A unique identifier for a given version of the last configuration was applied.

* `uuid` - The UUID of the virtual machine. Also exposed as the `id` of the resource.

* `default_ip_address` - The IP address selected by Terraform to be used with any provisioners configured on this resource. When possible, this is the first IPv4 address that is reachable through the default gateway configured on the machine, then the first reachable IPv6 address, and then the first general discovered address if neither exists. If VMware Tools is not running on the virtual machine, or if the virtual machine is powered off, this value will be blank.

* `guest_ip_addresses` - The current list of IP addresses on this machine, including the value of `default_ip_address`. If VMware Tools is not running on the virtual machine, or if the virtual machine is powered off, this list will be empty.

* `moid`: The [managed object reference ID][docs-about-morefs] of the created virtual machine.

* `vapp_transport` - Computed value which is only valid for cloned virtual machines. A list of vApp transport methods supported by the source virtual machine or template.

* `power_state` - A computed value for the current power state of the virtual machine. One of `on`, `off`, or `suspended`.

## Importing

An existing virtual machine can be [imported][docs-import] into the Terraform state by providing the full path to the virtual machine.

[docs-import]: /docs/import/index.html

**Examples**:

Import a virtual machine resource named `foo` located in the `dc-01` datacenter.

```shell
terraform import vsphere_virtual_machine.vm /dc-01/vm/foo
```

~> **NOTE:** The `vm` portion of the path is required by vSphere. If the virtual machine is located in a folder, the folder path needs to be included. This is because vSphere organizes virtual machines within a datacenter under the `vm` folder, and any additional folders created within the `vm` folder must be included in the path.

If the virtual machine `foo` is in a folder named `bar`, the import command would be:

```shell
terraform import vsphere_virtual_machine.vm /dc-01/vm/bar/foo
```

### Additional Importing Requirements

Many of the requirements for [cloning](#additional-requirements-and-notes-for-cloning) apply to importing. Although importing writes directly to the Terraform state, some rules can not be enforced during import time, so every effort should be made to ensure the correctness of the configuration before the import.

The following requirements apply to import:

* The disks must have a [`label`](#label) argument assigned in a convention matching `Hard Disk`, starting with disk number 0, based on each virtual disk order on the SCSI bus. As an example, a disk on SCSI controller `0` with a unit number of `0` would be labeled as `Hard Disk 0`, a disk on the same controller with a unit number of `1` would be `Hard Disk 1`, but the next disk, which is on SCSI controller `1` with a unit number of `0`, still becomes `Hard Disk 2`.

~> **NOTE:** Any custom `label` set at deployment of machine through Terraform, on import will not have the custom `label` and will default to `Hard Disk _x_`.

* Disks are always imported with [`keep_on_remove`](#keep_on_remove) enabled until the first `terraform apply` run which will remove the setting for known disks. This process safeguards against naming or accounting mistakes in the disk configuration.

* The storage controller count for the resource is set to the number of contiguous storage controllers found, starting with the controller at SCSI bus number `0`. If no storage controllers are discovered, the virtual machine is not eligible for import. For maximum compatibility, ensure that the virtual machine has the exact number of storage controllers needed and set the storage controller count accordingly.

After importing, you should run `terraform plan`. Unless you have changed anything else in the configuration that would cause other attributes to change. The only difference should be configuration-only changes, which are typically comprised of:

* The [`imported`](#imported) flag will transition from `true` to `false`.

* The [`keep_on_remove`](#keep_on_remove) of known disks will transition from `true` to `false`.

* Configuration supplied in the [`clone`](#clone) block, if present, will be persisted to state. This initial persistence operation does not perform any cloning or customization actions, nor does it force a new resource. After the first apply operation, further changes to `clone` will force the creation of a new resource.

~> **NOTE:** Do not make any configuration changes to `clone` after importing or upgrading from a legacy version of the provider before doing an initial `terraform apply` as these changes will not correctly force a new resource and your changes will have persisted to state, preventing further plans from correctly triggering a diff.

These changes only update Terraform state when applied. Hence, it is safe to run when the virtual machine is running. If more settings are modified, you may need to plan maintenance accordingly for any necessary virtual machine re-configurations.

## Migrating from a Previous Version of the Resource

~> **NOTE:** This section only applies this resource available in v0.4.2 or earlier of this provider.

The path for migrating to the current version of this resource is very similar to the [import](#importing) path; however, with the exception that the `terraform import` command does not need to be run. See that section for details on what is required before you run `terraform plan` on a provider resource that must be migrated.

A successful migration usually only results in a configuration-only diff - that is, Terraform reconciles the configuration settings that can not be set during the migration process with he Terraform state. In this event, no reconfiguration operations are sent to vSphere during the next `terraform apply`. For more information, see the [importing](#importing) section.
