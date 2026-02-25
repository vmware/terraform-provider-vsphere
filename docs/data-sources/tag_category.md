---
subcategory: "Inventory"
page_title: "VMware vSphere: vsphere_tag_category"
sidebar_current: "docs-vsphere-data-source-tag-category"
description: |-
  Provides a vSphere tag category data source.
  This can be used to reference tag categories not managed in Terraform.
---

# vsphere_tag_category

The `vsphere_tag_category` data source can be used to reference tag categories
that are not managed by Terraform. Its attributes are the same as the
[`vsphere_tag_category` resource][resource-tag-category], and, like importing,
the data source uses a name and category as search criteria. The `id` and other
attributes are populated with the data found by the search.

[resource-tag-category]: /docs/providers/vsphere/r/tag_category.html

~> **NOTE:** Tagging is not supported on direct ESXi hosts connections and
requires vCenter Server.

## Example Usage

### Lookup by name (classic)

```hcl
data "vsphere_tag_category" "category" {
  name = "example-category"
}
```

### Lookup by ID (new)

```hcl
data "vsphere_tag_category" "by_id" {
  id = "urn:vmomi:InventoryServiceCategory:xxxx"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Optional) The unique identifier of the tag category. If specified,  `name` must not be set.

* `name` - (Optional) The name of the tag category. Required if `id` is not set.

## Attribute Reference

In addition to the `id` being exported, all of the fields that are available in
the [`vsphere_tag_category` resource][resource-tag-category] are also populated.
