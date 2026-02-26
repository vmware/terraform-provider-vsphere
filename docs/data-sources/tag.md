---
subcategory: "Inventory"
page_title: "VMware vSphere: vsphere_tag"
sidebar_current: "docs-vsphere-data-source-tag-data-source"
description: |-
  Provides a vSphere tag data source.
  This can be used to reference tags not managed in Terraform.
---

# vsphere_tag

The `vsphere_tag` data source can be used to reference tags that are not managed
by Terraform. Its attributes are exactly the same as the
[`vsphere_tag` resource][resource-tag], and, like importing, the data source
uses a name and category as search criteria. The `id` and other attributes are
populated with the data found by the search.

[resource-tag]: /docs/providers/vsphere/r/tag.html

~> **NOTE:** Tagging is not supported on direct ESXi hosts connections and
requires vCenter Server.

## Example Usage

### Lookup by Name and Category

```hcl
data "vsphere_tag_category" "category" {
  name = "example-category"
}

data "vsphere_tag" "tag" {
  name        = "example-tag"
  category_id = data.vsphere_tag_category.category.id
}
```
### Lookup by ID (new)

```hcl
data "vsphere_tag" "by_id" {
  id = "urn:vmomi:InventoryServiceTag:xxxx"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The unique identifier of the tag. If specified, `name` and
  `category_id` must not be set.

* `name` - (Optional) The name of the tag. If specified, `category_id` must also
  be provided.

* `category_id` - (Optional) The ID of the tag category in which the tag is
  located. Required when `name` is used.

## Attribute Reference

In addition to the `id` being exported, all of the fields that are available in
the [`vsphere_tag` resource][resource-tag] are also populated.
