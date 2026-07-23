---
subcategory: "Security"
page_title: "VMware vSphere: SSO Group"
sidebar_current: "docs-vsphere-data-source-sso-group"
description: |-
  Provides a vCenter Single Sign-On group data source, including its members.
---

# vsphere_sso_group

The `vsphere_sso_group` data source can be used to look up a vCenter Single
Sign-On group by its name and domain, along with its user and nested-group
members.

## Example Usage

```hcl
data "vsphere_sso_group" "example" {
  name = "engineering"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the group to look up.
* `domain` - (Optional) The identity source domain the group belongs to.
  Defaults to the local (system) domain.

## Attribute Reference

* `id` - The identifier of the group, in the form `name@domain`.
* `domain` - The identity source domain the group belongs to.
* `description` - The description of the group.
* `member_user` - The user members of the group. Each entry exports:
  * `name` - The username of the member.
  * `domain` - The identity source domain the member belongs to.
* `member_group` - The nested group members of the group. Each entry exports:
  * `name` - The name of the nested group.
  * `domain` - The identity source domain the nested group belongs to.
