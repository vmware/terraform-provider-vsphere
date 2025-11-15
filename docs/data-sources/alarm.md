---
subcategory: "Inventory"
page_title: "VMware vSphere: vsphere_alarm"
sidebar_current: "docs-vsphere-data-source-alarm"
description: |-
  Provides a vSphere alarm data source. Returns attributes of a
  vSphere alarm.
---

# vsphere_compute_cluster_host_group

The `vsphere_alarm` data source can be used to retrieve the property of a given alarm.

## Example Usage

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_alarm" "alarm" {
  entity_type = "Datacenter"
  entity_id   = data.vsphere_datacenter.dc.id
  name        = "example"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the host group.
* `entity_id` - (Required) The [managed object reference ID][docs-about-morefs] of the entity the alarm will be created in.
* `entity_type` - (Required) The type of the entity the alarm will be created in.

## Attribute Reference
* `id`: The [managed object reference ID][docs-about-morefs] of the alarm.
* `enabled` - Whether or not the alarm is enabled.
* `expression_operator` - The logical link between expressions.
* `event_expression` - The event expressions of the alarm.
* `metric_expression` - The metric expressions of the alarm.
* `state_expression` - The state expressions of the alarm.
* `advanced_action` - The advanced actions of the alarm.
* `email_action` - The email alarm actions of the alarm.
* `snmp_action` - The snmp alarm actions of the alarm.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider
