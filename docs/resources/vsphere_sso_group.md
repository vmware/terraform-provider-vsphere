---
subcategory: "Security"
page_title: "VMware vSphere: SSO Group"
sidebar_current: "docs-vsphere-resource-sso-group"
description: |-
  Provides CRUD operations on a vCenter Single Sign-On group, including its membership.
---

# vsphere_sso_group

The `vsphere_sso_group` resource can be used to create and manage groups in the
vCenter Single Sign-On local (system) domain, including the users and nested
groups that belong to the group.

~> **NOTE:** Groups are always created in the local (system) domain. Group
membership, however, may include users from any identity source (see
`member_user`).

~> **NOTE:** Membership is authoritative. The principals listed in
`member_user` and `member_group` are the complete set of members.

~> **NOTE:** The connecting user must hold vCenter Single Sign-On administrator
privileges.

## Example Usage

This example creates a local user and a group that contains that local user, a
user from an external identity source, and a nested local group.

```hcl
resource "vsphere_sso_user" "example" {
  name     = "custom.user"
  password = "P@ssw0rd123!"
}

resource "vsphere_sso_group" "nested" {
  name = "engineering-leads"
}

resource "vsphere_sso_group" "example" {
  name        = "engineering"
  description = "Managed by Terraform"

  # A local user. The domain flows from the created resource.
  member_user {
    name   = vsphere_sso_user.example.name
    domain = vsphere_sso_user.example.domain
  }

  # A user from an external identity source.
  member_user {
    name   = "external.user"
    domain = "domain.local"
  }

  # A nested group in the local domain.
  member_group {
    name   = vsphere_sso_group.nested.name
    domain = vsphere_sso_group.nested.domain
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the group. Forces a new resource if changed.
* `description` - (Optional) A description of the group.
* `member_user` - (Optional) The set of users that are members of this group.
  Members may come from any identity source. Each `member_user` block supports
  the following:
  * `name` - (Required) The username of the member.
  * `domain` - (Required) The identity source domain the member belongs to.
* `member_group` - (Optional) The set of groups that are nested members of this
  group. Members may come from any identity source. Each `member_group` block
  supports the following:
  * `name` - (Required) The name of the nested group.
  * `domain` - (Required) The identity source domain the nested group belongs to.

## Attribute Reference

* `id` - The identifier of the group, in the form `name@domain`.
* `domain` - The identity source domain the group belongs to (the local/system
  domain).

## Importing

An existing group can be imported into this resource by supplying its
`name@domain` identifier. An example is below:

```shell
terraform import vsphere_sso_group.example engineering@vsphere.local
```
