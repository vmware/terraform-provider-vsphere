---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_vnic"
sidebar_current: "docs-vsphere-resource-vnic"
description: |-
  Provides a VMware vSphere vnic resource..
---

# vsphere_vnic

Provides a VMware vSphere vnic resource.

## Example Usages

**Create a vnic attached to a distributed virtual switch using the vMotion TCP/IP stack:**

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_distributed_virtual_switch" "vds" {
  name          = "vds-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
  host {
    host_system_id = data.vsphere_host.host.id
    devices        = ["vnic3"]
  }
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "pg-01"
  vlan_id                         = 1234
  distributed_virtual_switch_uuid = vsphere_distributed_virtual_switch.vds.id
}

resource "vsphere_vnic" "vnic" {
  host                    = data.vsphere_host.host.id
  distributed_switch_port = vsphere_distributed_virtual_switch.vds.id
  distributed_port_group  = vsphere_distributed_port_group.pg.id
  ipv4 {
    dhcp = true
  }
  netstack = "vmotion"
}
```

**Create a vnic attached to a portgroup using the default TCP/IP stack:**

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_host_virtual_switch" "hvs" {
  name             = "hvs-01"
  host_system_id   = data.vsphere_host.host.id
  network_adapters = ["vmnic3", "vmnic4"]
  active_nics      = ["vmnic3"]
  standby_nics     = ["vmnic4"]
}

resource "vsphere_host_port_group" "pg" {
  name                = "pg-01"
  virtual_switch_name = vsphere_host_virtual_switch.hvs.name
  host_system_id      = data.vsphere_host.host.id
}

resource "vsphere_vnic" "vnic" {
  host      = data.vsphere_host.host.id
  portgroup = vsphere_host_port_group.pg.name
  ipv4 {
    dhcp = true
  }
  services = ["vsan", "management"]
}
```

## Argument Reference

* `portgroup` - (Optional) Portgroup to attach the nic to. Do not set if you set distributed_switch_port.
* `distributed_switch_port` - (Optional) UUID of the vdswitch the nic will be attached to. Do not set if you set portgroup.
* `distributed_port_group` - (Optional) Key of the distributed portgroup the nic will connect to.
* `ipv4` - (Optional) IPv4 settings. Either this or `ipv6` needs to be set. See [IPv4 options](#ipv4-options) below.
* `ipv6` - (Optional) IPv6 settings. Either this or `ipv6` needs to be set. See [IPv6 options](#ipv6-options) below.
* `mac` - (Optional) MAC address of the interface.
* `mtu` - (Optional) MTU of the interface.
* `netstack` - (Optional) TCP/IP stack setting for this interface. Possible values are `defaultTcpipStack``, 'vmotion', 'vSphereProvisioning'. Changing this will force the creation of a new interface since it's not possible to change the stack once it gets created. (Default:`defaultTcpipStack`)
* `services` - (Optional) Enabled services setting for this interface. Currently support values are `vmotion`, `management`, and `vsan`.

### IPv4 Options

Configures the IPv4 settings of the network interface. Either DHCP or Static IP has to be set.

* `dhcp` - Use DHCP to configure the interface's IPv4 stack.
* `ip` - Address of the interface, if DHCP is not set.
* `netmask` - Netmask of the interface, if DHCP is not set.
* `gw` - IP address of the default gateway, if DHCP is not set.

### IPv6 Options

Configures the IPv6 settings of the network interface. Either DHCP or Autoconfig or Static IP has to be set.

* `dhcp` - Use DHCP to configure the interface's IPv6 stack.
* `autoconfig` - Use IPv6 Autoconfiguration (RFC2462).
* `addresses` -  List of IPv6 addresses
* `gw` - IP address of the default gateway, if DHCP or autoconfig is not set.

## Attribute Reference

* `id` - The ID of the vNic.

## Importing

An existing vNic can be [imported][docs-import] into this resource
via supplying the vNic's ID. An example is below:

[docs-import]: https://developer.hashicorp.com/terraform/cli/import

```shell
terraform import vsphere_vnic.vnic host-123_vmk2
```

The above would import the vnic `vmk2` from host with ID `host-123`.
