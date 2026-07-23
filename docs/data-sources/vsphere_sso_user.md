---
subcategory: "Security"
page_title: "VMware vSphere: SSO User"
sidebar_current: "docs-vsphere-data-source-sso-user"
description: |-
  Provides a vCenter Single Sign-On user data source.
---

# vsphere_sso_user

The `vsphere_sso_user` data source can be used to look up a vCenter Single
Sign-On user by its username and domain, including users from external identity
sources.

## Example Usage

```hcl
# A user in the local (system) domain.
data "vsphere_sso_user" "local" {
  name = "local.user"
}

# A user in an external identity source.
data "vsphere_sso_user" "external" {
  name   = "john.doe"
  domain = "example.com"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The username of the user to look up.
* `domain` - (Optional) The identity source domain the user belongs to. Defaults
  to the local (system) domain.

## Attribute Reference

* `id` - The identifier of the user, in the form `name@domain`.
* `domain` - The identity source domain the user belongs to.
* `first_name` - The first name of the user.
* `last_name` - The last name of the user.
* `email_address` - The email address of the user.
* `description` - The description of the user.
