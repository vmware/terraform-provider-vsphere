---
subcategory: "Virtual Machine"
page_title: "VMware vSphere: vsphere_content_library_items"
sidebar_current: "docs-vsphere-data-source-content-library-items"
description: |-
  Provides a VMware vSphere content library items data source.
---

# vsphere_content_library_items

The `vsphere_content_library_items` data source can be used to search for
items in a content library by name, regular expression, and/or type, and
returns the list of matching items.

~> **NOTE:** This resource requires vCenter and is not available on direct ESXi
host connections.

## Example Usage

### Search by name and type

```hcl
data "vsphere_content_library" "library" {
  name = "Content Library"
}

data "vsphere_content_library_items" "ovf_items" {
  library_id = data.vsphere_content_library.library.id
  name       = "ubuntu-server-lts"
  type       = "ovf"
}
```

### Search by name regex

```hcl
data "vsphere_content_library" "library" {
  name = "Content Library"
}

data "vsphere_content_library_items" "ubuntu" {
  library_id = data.vsphere_content_library.library.id
  name_regex = "^ubuntu-.*-lts$"
}
```

### Search all items of a given type

```hcl
data "vsphere_content_library" "library" {
  name = "Content Library"
}

data "vsphere_content_library_items" "all_ovf" {
  library_id = data.vsphere_content_library.library.id
  type       = "ovf"
}
```

### List all items in a library

```hcl
data "vsphere_content_library" "library" {
  name = "Content Library"
}

data "vsphere_content_library_items" "all" {
  library_id = data.vsphere_content_library.library.id
}
```

## Argument Reference

The following arguments are supported:

* `library_id` - (Required) The ID of the content library to search.
* `name` - (Optional) Filter items by exact name. Mutually exclusive with
  `name_regex`. If omitted, items of any name are returned.
* `name_regex` - (Optional) Filter items by name using a regular expression.
  Mutually exclusive with `name`. If omitted, items of any name are returned.
* `type` - (Optional) Filter items by type (e.g. `ovf`, `iso`, `vm-template`).
  If omitted, items of any type are returned.
* `sort_by` - (Optional) Sort items by field. Accepted values:
  `last_modified_time`, `creation_time`, `name`. Defaults to `last_modified_time`.
* `sort_order` - (Optional) Sort order. Accepted values: `asc`, `desc`.
  Defaults to `desc` (most recent first).
* `limit` - (Optional) Maximum number of items to return after sorting. If
  omitted, all matching items are returned.

## Attribute Reference

* `id` - A deterministic identifier derived from the search parameters.
* `items` - A list of content library items matching the search criteria. Each
  element contains:
  * `id` - The UUID of the content library item.
  * `name` - The name of the content library item.
  * `type` - The type of the content library item.
  * `creation_time` - The creation time of the item (RFC3339).
  * `last_modified_time` - The last modification time of the item (RFC3339).
