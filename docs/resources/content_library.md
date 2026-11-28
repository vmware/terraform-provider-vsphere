---
subcategory: "Virtual Machine"
page_title: "VMware vSphere: vsphere_content_library"
sidebar_current: "docs-vsphere-resource-content-library"
description: |-
  Provides a vSphere content library. Content libraries allow you to manage and share virtual machines, vApp templates, and other types of files. Content libraries enable you to share content across vCenter Server instances in the same or different locations.
---

# vsphere_content_library

The `vsphere_content_library` resource can be used to manage content libraries.

> **NOTE:** This resource requires a vCenter Server instance and is not available on direct ESXi host connections.

## Example Usage

The following example creates a publishing content library using the datastore named `publisher-datastore` as the storage backing.

[tf-vsphere-vm-resource]: /docs/providers/vsphere/r/virtual_machine.html

```hcl
data "vsphere_datacenter" "datacenter_a" {
  name = "dc-01-a"
}

data "vsphere_datastore" "publisher_datastore" {
  name          = "publisher-datastore"
  datacenter_id = data.vsphere_datacenter.datacenter_a.id
}

resource "vsphere_content_library" "publisher_content_library" {
  name            = "Publisher Content Library"
  description     = "A publishing content library."
  storage_backing = [data.vsphere_datastore.publisher_datastore.id]
}
```

The next example creates a subscribed content library using the URL of the publisher content library as the source and the datastore named `subscriber-datastore` as the storage backing.

```hcl
data "vsphere_datacenter" "datacenter_b" {
  name = "dc-01-b"
}

data "vsphere_datastore" "subscriber_datastore" {
  name          = "subscriber-datastore"
  datacenter_id = data.vsphere_datacenter.datacenter_b.id
}

resource "vsphere_content_library" "subscriber_content_library" {
  name            = "Subscriber Content Library"
  description     = "A subscribing content library."
  storage_backing = [data.vsphere_datastore.subscriber_datastore.id]
  subscription {
    subscription_url = "https://vc-01-a.example.com:443/cls/vcsp/lib/f42a4b25-844a-44ec-9063-a3a5e9cc88c7/lib.json"
    automatic_sync   = true
    on_demand        = false
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the content library.
* `description` - (Optional) A description for the content library.
* `storage_backing` - (Required) The [managed object reference ID][docs-about-morefs] of the datastore on which to store the content library items.
* `publication` - (Optional) Options to publish a local content library.
  * `authentication_method` - (Optional) Method to authenticate users. Must be `NONE` or `BASIC`.
  * `username` - (Optional) Username used by subscribers to authenticate. Currently can only be `vcsp`.
  * `password` - (Optional) Password used by subscribers to authenticate.
  * `published` - (Optional) Publish the content library. Default `false`.
* `subscription` - (Optional) Options subscribe to a published content library.
  * `subscription_url` - (Required) URL of the published content library.
  * `authentication_method` - (Optional) Authentication method to connect ro a published content library. Must be `NONE` or `BASIC`.
  * `username` - (Optional) Username used for authentication.
  * `password` - (Optional) Password used for authentication.
  * `automatic_sync` - (Optional) Enable automatic synchronization with the published library. Default `false`.
  * `on_demand` - (Optional) Download the library from a content only when needed. Default `true`.
  * `ssl_thumbprint` - (Optional) The SHA-1 thumbprint of the SSL certificate presented by the published content library, in the form `AA:BB:...`. Required when the certificate is not trusted by vCenter, such as a self-signed certificate or one replaced by an SSL inspection proxy. If not set, the thumbprint reported by the vCenter is used.

~> **NOTE:** The thumbprint can be retrieved with `openssl s_client -connect <host>:443 </dev/null | openssl x509 -noout -fingerprint -sha1`.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider

## Attribute Reference

* `id` The [managed object reference ID][docs-about-morefs] of the content library.
* `subscription`
  * `publish_url` - The URL of the published content library.

## Importing

An existing content library can be [imported][docs-import] into this resource by supplying the content library ID. For example:

[docs-import]: https://developer.hashicorp.com/terraform/cli/import

```shell
terraform import vsphere_content_library publisher_content_library f42a4b25-844a-44ec-9063-a3a5e9cc88c7
```
