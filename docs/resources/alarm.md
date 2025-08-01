---
subcategory: "Inventory"
page_title: "VMware vSphere: vsphere_alarm"
sidebar_current: "docs-vsphere-resource-inventory-alarm"
description: |-
  Provides a VMware vSphere alarm resource. This can be used deployed on all kinds of vSphere inventory objects.
---

# vsphere_alarm

Provides a VMware vSphere alarm resource. This can be used deployed on all kinds of vSphere inventory objects.

## Example Usages

**Create warning alarm with an email action on a datacenter when on a host is disconnected:**

```hcl
resource "vsphere_alarm" "host_disconnected_warning" {
  name        = "Host disconnected"
  description = "Triggers a warning when a host is disconnected"
  entity_type = "Datacenter"
  entity_id   = data.vsphere_datacenter.datacenter.id

  state_expression {
    operator    = "isEqual"
    state_path  = "runtime.connectionState"
    object_type = "HostSystem"
    yellow      = "disconnected"
  }

  email_action {
    to          = "foo@example.com"
    subject     = "Host disconnected"
    start_state = "green"
    final_state = "yellow"
  }
}
```

**Create critical alarm when a host CPU usage is above 95% for 5min:**

```hcl
resource "vsphere_alarm" "host_disconnected_warning" {
  name        = "Host CPU usage"
  description = "Triggers a critical when a host CPU usage is too high"
  entity_type = "Datacenter"
  entity_id   = data.vsphere_datacenter.datacenter.id

  metric_expression {
    metric_counter_id = 2 # hosted counter ID for CPU usage
    operator          = "isAbove"
    object_type       = "HostSystem"
    red               = 9500
    red_interval      = 300
  }
}
```

**Create alarm to set a host in maintenance when a lacp down event occurs, and remove the maintenance when the lacp is up again:**

```hcl
resource "vsphere_alarm" "lacp_down_on_host" {
  name        = "LACP down on host"
  description = "Set the host is maintenance if lacp goes down"
  entity_type = "Datacenter"
  entity_id   = data.vsphere_datacenter.datacenter.id

  # alert
  event_expression {
    event_type_id = "esx.problem.net.lacp.lag.transition.down"
    object_type   = "HostSystem"
    status        = "red"
  }

  # recovery
  event_expression {
    event_type_id = "esx.problem.net.lacp.lag.transition.up"
    object_type   = "HostSystem"
    status        = "green"
  }

  # enter maintenance mode if alert is triggered
  advanced_action {
    start_state = "green"
    final_state = "red"
    name        = "EnterMaintenanceMode_Task"
  }

  # remove maintenance mode on alert² recovery
  advanced_action {
    start_state = "red"
    final_state = "green"
    name        = "ExitMaintenanceMode_Task"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the alarm. This name needs to be unique
  within the vCenter. Forces a new resource if changed.
* `description` - (Optional) The alarm description.
* `enabled` - (Optional) Whether or not the alarm is enabled.
* `entity_id` - (Required) The [managed object reference ID][docs-about-morefs] of the entity the alarm will be created in.
* `entity_type` - (Required) The type of the entity the alarm will be created in.
* `expression_operator` - (Optional) The logical link between expressions.
* `event_expression` - (Optional) Alarm trigger expressions based on events.
* `metric_expression` - (Optional) Alarm trigger expressions based on metric values.
* `state_expression` - (Optional) Alarm trigger expressions based on object state changes.
* `advanced_action` - (Optional) Advanced alarm action to trigger depending on the alarm state, such as entering maintenance mode.
* `email_action` - (Optional) Email alarm action to trigger depending on the alarm state.
* `snmp_action` - (Optional) Snmp alarm action to trigger depending on the alarm state.

## Attribute Reference

* `id` - The ID of the alarm.

## Importing

Importing vSphere alarm is not managed.
