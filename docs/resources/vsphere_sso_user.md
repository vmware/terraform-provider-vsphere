---
subcategory: "Security"
page_title: "VMware vSphere: SSO User"
sidebar_current: "docs-vsphere-resource-sso-user"
description: |-
  Provides CRUD operations on a vCenter Single Sign-On local user.
---

# vsphere_sso_user

The `vsphere_sso_user` resource can be used to create and manage local users in
the vCenter Single Sign-On local (system) domain.

~> **NOTE:** Users can only be *created* in the local domain. A user from an
external identity source is managed by that identity provider. It can be
imported and referenced (for example, added to a group), but it cannot be
created, modified, or deleted through this resource.

~> **NOTE:** When `domain` is omitted it defaults to the
resolved local domain.

~> **NOTE:** The connecting user must hold vCenter Single Sign-On administrator
privileges.

## Example Usage

```hcl
resource "vsphere_sso_user" "example" {
  name          = "custom.user"
  password      = "P@ssw0rd123!"
  first_name    = "Custom"
  last_name     = "User"
  email_address = "custom.user@domain.local"
  description   = "Managed by Terraform"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The username of the user. Forces a new resource if
  changed.
* `password` - (Required) The password for the user. This value is write-only
  it cannot be read back from vCenter, so no drift is detected on it.
* `domain` - (Optional) The identity source domain the user belongs to. Defaults
  to the local (system) domain. Forces a new resource if changed.
* `first_name` - (Optional) The first name of the user.
* `last_name` - (Optional) The last name of the user.
* `email_address` - (Optional) The email address of the user.
* `description` - (Optional) A description of the user.

## Attribute Reference

* `id` - The identifier of the user, in the form `name@domain`.

## Importing

An existing user can be imported into this resource by supplying its
`name@domain` identifier. An example is below:

```shell
terraform import vsphere_sso_user.example custom.user@vsphere.local
```

~> **NOTE:** The `password` cannot be read from vCenter, so it is not populated
by an import.
