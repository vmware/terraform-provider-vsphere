// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var testAccDataSourceVSphereAlarmExpectedRegexp = regexp.MustCompile("^alarm-")

func TestAccDataSourceVSphereAlarm_EventExpWithActions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereAlarmEventExpWithActions(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_alarm.alarm",
						"id",
						testAccDataSourceVSphereAlarmExpectedRegexp,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_alarm.alarm", "id",
						"vsphere_alarm.alarm", "id",
					),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "description", "when lacp is broken"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "event_expression.0.event_type_id", "esx.problem.net.lacp.lag.transition.down"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "event_expression.0.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "event_expression.0.status", "red"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "event_expression.0.comparison.0.attribute_name", "datacenter.name"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "event_expression.0.comparison.0.operator", "equals"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "event_expression.0.comparison.0.value", "dc_forbidden"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "advanced_action.0.name", "EnterMaintenanceMode_Task"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "advanced_action.0.start_state", "green"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "advanced_action.0.final_state", "red"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "email_action.0.start_state", "green"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "email_action.0.final_state", "red"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "email_action.0.to", "foo@example.com"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "email_action.0.cc", "bar@example.com"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "email_action.0.subject", "LAG down on host"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereAlarmEventExpWithActions() string {
	return fmt.Sprintf(`
data "vsphere_datacenter" "dc" {
  name = "%s"
}

resource "vsphere_alarm" "alarm" {
  name        = "test_alarm"
  description = "when lacp is broken"
  entity_id   = data.vsphere_datacenter.dc.id
  entity_type = "Datacenter"

  event_expression {
    event_type_id = "esx.problem.net.lacp.lag.transition.down"
    object_type   = "HostSystem"
    status        = "red"
	comparison {
	  attribute_name = "datacenter.name"
	  operator       = "equals"
	  value          = "dc_forbidden"
	}
  }

  email_action {
    start_state = "green"
    final_state = "red"
    to          = "foo@example.com"
	cc          = "bar@example.com"
	subject     = "LAG down on host"
  }

  advanced_action {
    start_state = "green"
    final_state = "red"
    name        = "EnterMaintenanceMode_Task"
  }
}

data "vsphere_alarm" "alarm" {
  entity_type  = "Datacenter"
  entity_id    = data.vsphere_datacenter.dc.id
  name         = vsphere_alarm.alarm.name
}
`, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
}

func TestAccDataSourceVSphereAlarm_StateExpWithActions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereAlarmStateExp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_alarm.alarm",
						"id",
						testAccDataSourceVSphereAlarmExpectedRegexp,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_alarm.alarm", "id",
						"vsphere_alarm.alarm", "id",
					),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "state_expression.0.operator", "isEqual"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "state_expression.0.state_path", "runtime.connectionState"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "state_expression.0.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "state_expression.0.yellow", "disconnected"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereAlarmStateExp() string {
	return fmt.Sprintf(`
data "vsphere_datacenter" "dc" {
  name = "%s"
}

resource "vsphere_alarm" "alarm" {
  name        = "test_alarm"
  description = "host disconnected"
  entity_id   = data.vsphere_datacenter.dc.id
  entity_type = "Datacenter"

  state_expression {
    operator    = "isEqual"
    state_path  = "runtime.connectionState"
    object_type = "HostSystem"
	yellow      = "disconnected"
  }
}

data "vsphere_alarm" "alarm" {
  entity_type  = "Datacenter"
  entity_id    = data.vsphere_datacenter.dc.id
  name         = vsphere_alarm.alarm.name
}
`, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
}

func TestAccDataSourceVSphereAlarm_MetricExpWithActions(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereAlarmMetricExp(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_alarm.alarm",
						"id",
						testAccDataSourceVSphereAlarmExpectedRegexp,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_alarm.alarm", "id",
						"vsphere_alarm.alarm", "id",
					),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "metric_expression.0.metric_counter_id", "2"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "metric_expression.0.operator", "isAbove"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "metric_expression.0.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "metric_expression.0.yellow", "9900"),
					resource.TestCheckResourceAttr("data.vsphere_alarm.alarm", "metric_expression.0.yellow_interval", "300"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereAlarmMetricExp() string {
	return fmt.Sprintf(`
data "vsphere_datacenter" "dc" {
  name = "%s"
}

resource "vsphere_alarm" "alarm" {
  name        = "test_alarm"
  description = "host disconnected"
  entity_id   = data.vsphere_datacenter.dc.id
  entity_type = "Datacenter"

  metric_expression {
    metric_counter_id = 2
    operator          = "isAbove"
    object_type       = "HostSystem"
	yellow            = 9900
	yellow_interval   = 300
  }
}

data "vsphere_alarm" "alarm" {
  entity_type  = "Datacenter"
  entity_id    = data.vsphere_datacenter.dc.id
  name         = vsphere_alarm.alarm.name
}
`, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
}
