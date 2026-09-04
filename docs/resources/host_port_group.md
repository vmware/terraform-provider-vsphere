---
subcategory: "Networking"
page_title: "VMware vSphere: vsphere_host_port_group"
sidebar_current: "docs-vsphere-resource-networking-host-port-group"
description: |-
  Provides a vSphere port group resource to manage port groups on ESXi hosts.
---

# vsphere_host_port_group

The `vsphere_host_port_group` resource can be used to manage port groups on
ESXi hosts. These port groups are connected to standard switches, which
can be managed by the [`vsphere_host_virtual_switch`][host-virtual-switch]
resource.

For an overview on vSphere networking concepts, see [the product documentation][ref-vsphere-net-concepts].

[host-virtual-switch]: /docs/providers/vsphere/r/host_virtual_switch.html
[ref-vsphere-net-concepts]: hhttps://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-networking-8-0/introduction-to-vsphere-networking.html

## Example Usages

**Create a Virtual Switch and Bind a Port Group:**

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_host_virtual_switch" "host_virtual_switch" {
  name           = "switch-01"
  host_system_id = data.vsphere_host.host.id

  network_adapters = ["vmnic0", "vmnic1"]

  active_nics  = ["vmnic0"]
  standby_nics = ["vmnic1"]
}

resource "vsphere_host_port_group" "pg" {
  name                = "portgroup-01"
  host_system_id      = data.vsphere_host.host.id
  virtual_switch_name = vsphere_host_virtual_switch.host_virtual_switch.name
}
```

**Create a Port Group with a VLAN and ab Override:**

This example sets the trunk mode VLAN (`4095`, which passes through all tags)
and sets
[`allow_promiscuous`](/docs/providers/vsphere/r/host_virtual_switch.html#allow_promiscuous)
to ensure that all traffic is seen on the port. The setting overrides
the implicit default of `false` set on the standard switch.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_host" "host" {
  name          = "esxi-01.example.com"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_host_virtual_switch" "host_virtual_switch" {
  name           = "switch-01"
  host_system_id = data.vsphere_host.host.id

  network_adapters = ["vmnic0", "vmnic1"]

  active_nics  = ["vmnic0"]
  standby_nics = ["vmnic1"]
}

resource "vsphere_host_port_group" "pg" {
  name                = "portgroup-01"
  host_system_id      = data.vsphere_host.host.id
  virtual_switch_name = vsphere_host_virtual_switch.host_virtual_switch.name

  vlan_id = 4095

  allow_promiscuous = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the port group.  Forces a new resource if
  changed.
* `host_system_id` - (Required) The [managed object ID][docs-about-morefs] of
  the host to set the port group up on. Forces a new resource if changed.
* `virtual_switch_name` - (Required) The name of the virtual switch to bind
  this port group to. Forces a new resource if changed.
* `vlan_id` - (Optional) The VLAN ID/trunk mode for this port group.  An ID of
  `0` denotes no tagging, an ID of `1`-`4094` tags with the specific ID, and an
  ID of `4095` enables trunk mode, allowing the guest to manage its own
  tagging. Default: `0`.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

### Policy Options

In addition to the above options, you can configure any policy option that is
available under the [`vsphere_host_virtual_switch` policy options
section][host-vswitch-policy-options].  Any policy option that is not set is
**inherited** from the virtual switch, its options propagating to the port
group.

Refer to the link for a full list of options that can be set.

[host-vswitch-policy-options]: /docs/providers/vsphere/r/host_virtual_switch.html#policy-options

## Attribute Reference

The following attributes are exported:

* `id` - An ID for the port group that is _unique_ to Terraform.
  The convention is a prefix, the host system ID, and the port group name.
  For example,`tf-HostPortGroup:host-10:portgroup-01`. Tracking a port group
  on a standard switch, which can be created with or without a vCenter Server,
  is different than a dvPortGroup which is tracked as a managed object ID in
  vCenter Server versus a key on a host.
* `computed_policy` - A map with a full set of the [policy
  options][host-vswitch-policy-options] computed from defaults and overrides,
  explaining the effective policy for this port group.
* `key` - The key for this port group as returned from the vSphere API.
* `ports` - A list of ports that currently exist and are used on this port group.

## Importing

An existing host port group can be [imported][docs-import] into this resource
using the host port group's ID. An example is below:

[docs-import]: https://developer.hashicorp.com/terraform/cli/import

```shell
terraform import vsphere_host_port_group.management tf-HostPortGroup:host-123:management
```

The above would import the `management` host port group from host with ID `host-123`.
