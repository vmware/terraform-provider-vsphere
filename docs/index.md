---
page_title: "Provider: VMware vSphere"
sidebar_current: "docs-vsphere-index"
description: |-
  A Terraform provider to work with VMware vSphere, allowing management of virtual machines and other vSphere and vSAN resources.
---

<img src="https://raw.githubusercontent.com/vmware/terraform-provider-vsphere/main/docs/images/icon-color.svg" alt="VMware vSphere" width="150">

# Terraform Provider for VMware vSphere

This provider gives Terraform the ability to work with VMware vSphere. This
provider can be used to manage many aspects of a vSphere environment, including
virtual machines, standard and distributed switches, datastores, content
libraries, and more.

Use the navigation to read about the resources and data sources supported by
this provider.

This release is supported with:

- VMware vSphere 9.x
- VMware vSphere 8.x

Refer to the [Broadcom Product Lifecycle][product-lifecycle].

[product-lifecycle]: https://support.broadcom.com/group/ecx/productlifecycle

~> **NOTE:** This provider requires API write access and is therefore not
compatible with VMware vSphere Hypervisor version 8, the free entry-level
hypervisor.

## Example Usage

The following abridged example demonstrates basic usage of the provider to
provision a virtual machine using the
[`vsphere_virtual_machine`][tf-vsphere-virtual-machine-resource] resource.
The datacenter, datastore, resource pool, and network are discovered using the
[`vsphere_datacenter`][tf-vsphere-datacenter],
[`vsphere_datastore`][tf-vsphere-datastore],
[`vsphere_resource_pool`][tf-vsphere-resource-pool], and
[`vsphere_network`][tf-vsphere-network] data sources respectively.

[tf-vsphere-virtual-machine-resource]: /docs/providers/vsphere/r/virtual_machine.html
[tf-vsphere-datacenter]: /docs/providers/vsphere/d/datacenter.html
[tf-vsphere-datastore]: /docs/providers/vsphere/d/datastore.html
[tf-vsphere-resource-pool]: /docs/providers/vsphere/d/resource_pool.html
[tf-vsphere-network]: /docs/providers/vsphere/d/network.html

```hcl
provider "vsphere" {
  user                 = var.vsphere_user
  password             = var.vsphere_password
  vsphere_server       = var.vsphere_server
  allow_unverified_ssl = true
  api_timeout          = 10
}

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
    label = "disk0"
    size  = 20
  }
}
```

Refer to the provider documentation for information on all of the resources
and data sources supported by this provider. Each includes a detailed
description of the purpose and how to use it.

## Argument Reference

The following arguments are used to configure the provider:

* `user` - (Required) This is the username for vSphere API operations. Can also
  be specified with the `VSPHERE_USER` environment variable.
* `password` - (Required) This is the password for vSphere API operations. Can
  also be specified with the `VSPHERE_PASSWORD` environment variable.
* `vsphere_server` - (Required) This is the vCenter Server FQDN or IP Address
  for vSphere API operations. Can also be specified with the `VSPHERE_SERVER`
  environment variable.
* `allow_unverified_ssl` - (Optional) Boolean that can be set to true to
  disable SSL certificate verification. This should be used with care as it
  could allow an attacker to intercept your authentication token. If omitted,
  default value is `false`. Can also be specified with the
  `VSPHERE_ALLOW_UNVERIFIED_SSL` environment variable.
* `vim_keep_alive` - (Optional) Keep alive interval in minutes for the VIM
  session. Standard session timeout in vSphere is 30 minutes. This defaults to
  10 minutes to ensure that operations that take a longer than 30 minutes
  without API interaction do not result in a session timeout. Can also be
  specified with the `VSPHERE_VIM_KEEP_ALIVE` environment variable.
* `api_timeout` - (Optional) Sets the number of minutes to wait for operations
  to complete. The default timeout is 5 minutes. Can also be
  specified with the `VSPHERE_API_TIMEOUT` environment variable.

~> **NOTE:** Use of the `api_timeout` option to extend the timeout from the
default is recommended when creating virtual machines with large disks.

### Session Persistence Options

The provider also provides session persistence options that can be configured
below. These can help when using Terraform in a way where session limits could
be normally reached by creating a new session for every run, such as a large
amount of concurrent or consecutive Terraform runs in a short period of time.

~> **NOTE:** Session keys are as good as user credentials for as long as the
session is valid for - handle them with care and delete them when you know you
will no longer need them.

* `persist_session` - (Optional) Persist the SOAP and REST client sessions to
  disk. Default: `false`. Can also be specified by the
  `VSPHERE_PERSIST_SESSION` environment variable.
* `vim_session_path` - (Optional) The directory to save the VIM SOAP API
  session to. Default: `${HOME}/.govmomi/sessions`. Can also be specified by
  the `VSPHERE_VIM_SESSION_PATH` environment variable.
* `rest_session_path` - The directory to save the REST API session to.
  Default: `${HOME}/.govmomi/rest_sessions`. Can also be specified by the
  `VSPHERE_REST_SESSION_PATH` environment variable.

#### Session Interoperability for vmware/govc and the Provider

The session format used to save VIM SOAP sessions is the same used
with [`vmware/govc`][docs-govc]. If you use `govc` as part of your provisioning
process, Terraform will use the saved session if present and if
`persist_session` is enabled.

### Debugging Options

~> **NOTE:** The following options can leak sensitive data and should only be
enabled when instructed to do so by HashiCorp for the purposes of
troubleshooting issues with the provider, or when attempting to perform your
own troubleshooting. Use these option at your own risk and do not leave enabled!

* `client_debug` - (Optional) When `true`, the provider logs SOAP calls made to
  the vSphere API to disk.  The log files are logged to `${HOME}/.govmomi`.
  Can also be specified with the `VSPHERE_CLIENT_DEBUG` environment variable.
* `client_debug_path` - (Optional) Override the default log path. Can also
   be specified with the `VSPHERE_CLIENT_DEBUG_PATH` environment variable.
* `client_debug_path_run` - (Optional) A specific subdirectory in
  `client_debug_path` to use for debugging calls for this particular Terraform
  configuration. All data in this directory is removed at the start of the
  Terraform run. Can also be specified with the `VSPHERE_CLIENT_DEBUG_PATH_RUN`
  environment variable.

## Notes on Required Privileges

When using a non-administrator account to perform provider operations, consider
that most Terraform resources perform operations in a CRUD-like fashion and
require both read and write privileges to the resources they are managing. Make
sure that the user has appropriate read-write access to the resources you need
to work with. Read-only access should be sufficient when only using data
sources on some features. You can read more about vSphere permissions and user
management [here][vsphere-docs-user-management].

[vsphere-docs-user-management]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-security-8-0/vsphere-permissions-and-user-management-tasks.html

There are a some notable exceptions to keep in mind when setting up a restricted
provisioning user:

### Tags

The provider will always attempt to read [tags][vsphere-docs-tags] from a
resource, even if you do not have any tags defined. Ensure that your user has
access to read tags.

[vsphere-docs-tags]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-tags-and-attributes.html

### Events

The provider will attempt to read event data from vSphere to check for certain
events, such as, virtual machine customization or power events. Ensure that the
user account used for Terraform has the privilege to be able to read event data.

### Storage

The provider implementation requires the ability to read storage profiles
from vSphere for some resource and data source operations. Ensure that the
user account used for Terraform is provided the Profile-driven Storage > View
(`StorageProfile.View`) privilege to be able to read the available storage
policies.

### Virtual Machine

The provider implementation requires the ability to set a default swap
placement policy on a virtual machine resource. Ensure that the user account
used for Terraform is provided the Virtual Machine > Change Configuration >
Change Swapfile Placement (`VirtualMachine.Config.SwapPlacement`) privilege.

## Use of Managed Object References by the Provider

Unlike the vSphere client, many resources managed by the provider
use managed object IDs (or UUIDs when provided and practical) when referring
to placement parameters and upstream resources. This provides a stable
interface for providing necessary data to downstream resources, in addition to
minimizing the issues that can arise from the flexibility in how an individual
object's name or inventory path can be supplied.

There are several data sources (such as
[`vsphere_datacenter`][tf-vsphere-datacenter],
[`vsphere_host`][tf-vsphere-host],
[`vsphere_resource_pool`][tf-vsphere-resource-pool],
[`vsphere_datastore`][tf-vsphere-datastore], and
[`vsphere_network`][tf-vsphere-network]) that assist with searching for a
specific resource. For usage details on a specific data source, look for a
link in the provider documentation. In addition, most vSphere
resources in Terraform supply the managed object ID (or UUID, when it makes
more sense) as the `id` attribute, which can be supplied to downstream
resources that should depend on the parent.

[tf-vsphere-host]: /docs/providers/vsphere/d/host.html

### Locating Managed Object IDs

There are certain points in time that you may need to locate the managed object
ID of a specific vSphere resource yourself. A couple of methods are documented
below.

#### Using `govc`

[`govc`][docs-govc] is an vSphere CLI built on [govmomi][docs-govmomi], the
vSphere Go SDK. It has a robust inventory browser command that can also be used
to list managed object IDs.

To get all the necessary data in a single output, use `govc ls -l -i PATH`.

Sample output is below:

```shell
$ govc ls -l -i /dc-01/vm
VirtualMachine:vm-123 /dc-01/vm/foobar
Folder:group-v234 /dc-01/vm/subfolder
```

To do a reverse search, supply the `-L` switch:

```shell
$ govc ls -i -l -L VirtualMachine:vm-123
VirtualMachine:vm-123 /dc-01/vm/foo
```

For details on setting up `govc`, see the [GitHub project][docs-govc].

[docs-govc]: https://github.com/vmware/govmomi/tree/main/govc
[docs-govmomi]: https://github.com/vmware/govmomi

#### Using the vSphere Managed Object Browser (MOB)

The Managed Object Browser (MOB) allows one to browse the entire vSphere
inventory as it's presented to the API. It's normally accessed using
`https://<vcenter_fqdn>/mob`.

~> **NOTE:** The MOB also offers API method invocation capabilities, and for
security reasons should be used sparingly. Modern vSphere installations may
have the MOB disabled by default, at the very least on ESXi systems. For more
information on current security best practices related to the MOB on ESXi,
click [here][vsphere-docs-esxi-mob].

[vsphere-docs-esxi-mob]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-security-8-0/securing-esxi-hosts/general-security-recommendations/disable-the-managed-object-browser.html

## Bug Reports and Contributing

For more information how how to submit bug reports, feature requests, or
details on how to make your own contributions to the provider, see the vSphere
provider [project page][tf-vsphere-project-page].

[tf-vsphere-project-page]: https://github.com/vmware/terraform-provider-vsphere
