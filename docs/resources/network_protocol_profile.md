---
subcategory: "Networking"
page_title: "VMware vSphere: vsphere_network_protocol_profile"
sidebar_current: "docs-vsphere-resource-networking-network-protocol-profile"
description: |-
  Provides a VMware vSphere network protocol profile resource. This can be used to manage network protocol profiles (IP pools) in vSphere.
---

# vsphere_network_protocol_profile

The `vsphere_network_protocol_profile` resource can be used to
create and manage network protocol profiles, also known as IP pools. A
network protocol profile defines a range of IP addresses, along with DNS,
gateway, and proxy settings, that can be associated with one or more
networks (standard port groups, distributed port groups, or opaque
networks) within a datacenter. vCenter Server uses this configuration to
automatically assign network settings to virtual machines during
provisioning and customization.

~> **NOTE:** Network protocol profiles are unsupported on direct ESXi host
connections and require vCenter Server.

~> **NOTE:** vCenter Server allows a network to be associated with only one
network protocol profile at a time. If a network in `network_ids` is already
associated with a different network protocol profile, vCenter Server will
silently move it away from that profile rather than rejecting the request.
To prevent this, the provider validates that none of the configured
`network_ids` are already assigned to another network protocol profile, and
will raise an error (both at plan time and apply time) if a conflict is
found.

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_network" "network" {
  name          = "VM Network"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_network_protocol_profile" "profile" {
  name          = "example-profile"
  datacenter_id = data.vsphere_datacenter.datacenter.id
  network_ids   = [data.vsphere_network.network.id]

  dns_domain      = "example.com"
  dns_search_path = "example.com"

  ipv4 {
    subnet  = "10.10.10.0"
    netmask = "255.255.255.0"
    gateway = "10.10.10.1"
    range   = "10.10.10.100#100"

    dns_servers = ["10.10.10.10", "10.10.10.11"]
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the network protocol profile.
* `datacenter_id` - (Required) The [managed object ID][docs-about-morefs] of
  the datacenter this network protocol profile is associated with. Forces a
  new resource if changed.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

* `network_ids` - (Optional) The managed object IDs of the networks
  (standard port groups, distributed port groups, or opaque networks)
  associated with this network protocol profile.
* `dns_domain` - (Optional) The DNS domain to use for this network protocol
  profile, for example `example.com`.
* `dns_search_path` - (Optional) The DNS search path to use for this network
  protocol profile, for example `eng.example.com;example.com`.
* `host_prefix` - (Optional) The prefix to use when generating host names
  for this network protocol profile.
* `http_proxy` - (Optional) The HTTP proxy to use on this network, in the
  form of a host and port, for example `proxy.example.com:3128`.
* `ipv4` - (Optional) An IPv4 configuration block, documented below. At
  least one of `ipv4` or `ipv6` must be specified.
* `ipv6` - (Optional) An IPv6 configuration block, documented below. At
  least one of `ipv4` or `ipv6` must be specified.

### IP Configuration Block (`ipv4` and `ipv6`)

* `subnet` - (Required) The address of the subnet, for example `10.10.10.0`.
* `netmask` - (Required) The netmask of the subnet, for example
  `255.255.255.0`.
* `gateway` - (Optional) The gateway of the subnet.
* `range` - (Required) The range(s) of addresses available for allocation,
  specified as one or more comma-separated `<start-address>#<count>` pairs,
  for example `10.10.10.100#100`.
* `dns_servers` - (Optional) The DNS server addresses to use for this
  network protocol profile.
* `dhcp_available` - (Optional) Whether a DHCP server is available on this
  network.
* `enabled` - (Optional) Whether addresses can be allocated from this range.
  Default: `true`.

## Attribute Reference

In addition to all arguments above, the following attributes are exported
for both the `ipv4` and `ipv6` configuration blocks:

* `available_addresses` - The number of addresses available for allocation
  from this range.
* `allocated_addresses` - The number of addresses currently allocated from
  this range.

## Importing

An existing network protocol profile can be [imported][docs-import] into
this resource via a colon-separated combination of the datacenter's managed
object ID and either the network protocol profile's name or its numeric ID,
via the following command:

[docs-import]: https://developer.hashicorp.com/terraform/cli/import

```shell
terraform import vsphere_network_protocol_profile.profile datacenter-21:example-profile
```

or, using the numeric ID:

```shell
terraform import vsphere_network_protocol_profile.profile datacenter-21:5
```
