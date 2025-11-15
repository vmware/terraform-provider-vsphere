// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

const AlarmResource = "testacc"

func testAccResourceAlarmCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereAlarm(s, AlarmResource)
		if err != nil {
			if strings.Contains(err.Error(), "not found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected permissions to be missing")
		}
		return nil
	}
}

func TestAccResourceVsphereAlarm_basicEvent(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceAlarmCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereAlarmConfigBasicEvent(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceAlarmCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "name", "lacp down"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "description", "when lacp is broken"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "expression_operator", "or"),
					resource.TestCheckResourceAttrSet("vsphere_alarm."+AlarmResource, "entity_id"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "entity_type", "Datacenter"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.event_type_id", "esx.problem.net.lacp.lag.transition.down"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.event_type", "Event"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.status", "red"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.comparison.0.attribute_name", "datacenter.name"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.comparison.0.operator", "equals"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.0.comparison.0.value", "dc_forbidden"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.1.event_type_id", "esx.problem.net.lacp.lag.transition.up"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.1.event_type", "Event"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.1.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "event_expression.1.status", "green"),
				),
			},
		},
	})
}

func testAccResourceVsphereAlarmConfigBasicEvent() string {
	return fmt.Sprintf(`
%s

resource vsphere_alarm "%s" {
  name        = "lacp down"
  description = "when lacp is broken"
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"

  event_expression {
    event_type_id  = "esx.problem.net.lacp.lag.transition.down"
    object_type    = "HostSystem"
    status         = "red"
	comparison {
	  attribute_name = "datacenter.name"
	  operator       = "equals"
	  value          = "dc_forbidden"
	}
  }

  event_expression {
    event_type_id  = "esx.problem.net.lacp.lag.transition.up"
    object_type    = "HostSystem"
    status         = "green"
  }
}
`, testhelper.ConfigDataRootDC1(), AlarmResource)
}

func TestAccResourceVsphereAlarm_snmpAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceAlarmCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereAlarmConfigSnmpAction(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceAlarmCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "snmp_action.0.start_state", "green"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "snmp_action.0.final_state", "red"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "snmp_action.0.repeat", "false"),
				),
			},
		},
	})
}

func testAccResourceVsphereAlarmConfigSnmpAction() string {
	return fmt.Sprintf(`
%s

resource vsphere_alarm "%s" {
  name        = "lacp down"
  description = "when lacp is broken"
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"

  event_expression {
    event_type_id  = "esx.problem.net.lacp.lag.transition.down"
    object_type    = "HostSystem"
    status         = "red"
  }

  snmp_action {
    start_state = "green"
    final_state = "red"
  }
}
`, testhelper.ConfigDataRootDC1(), AlarmResource)
}

func TestAccResourceVsphereAlarm_advancedAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceAlarmCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereAlarmConfigAdvancedAction(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceAlarmCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "advanced_action.0.name", "EnterMaintenanceMode_Task"),
				),
			},
		},
	})
}

func testAccResourceVsphereAlarmConfigAdvancedAction() string {
	return fmt.Sprintf(`
%s

resource vsphere_alarm "%s" {
  name        = "lacp down"
  description = "when lacp is broken"
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"

  event_expression {
    event_type_id = "esx.problem.net.lacp.lag.transition.down"
    object_type   = "HostSystem"
    status        = "red"
  }

  advanced_action {
    start_state = "green"
    final_state = "red"
    name        = "EnterMaintenanceMode_Task"
  }
}
`, testhelper.ConfigDataRootDC1(), AlarmResource)
}

func TestAccResourceVsphereAlarm_emailAction(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceAlarmCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereAlarmConfigEmailAction(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceAlarmCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "email_action.0.to", "foo@example.com"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "email_action.0.cc", "bar@example.com"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "email_action.0.subject", "LAG down on host"),
				),
			},
		},
	})
}

func testAccResourceVsphereAlarmConfigEmailAction() string {
	return fmt.Sprintf(`
%s

resource vsphere_alarm "%s" {
  name        = "LAG down"
  description = "when lag is broken"
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"

  event_expression {
    event_type_id = "esx.problem.net.lacp.lag.transition.down"
    object_type   = "HostSystem"
    status        = "red"
  }

  email_action {
    start_state = "green"
    final_state = "red"
    to          = "foo@example.com"
	cc          = "bar@example.com"
	subject     = "LAG down on host"
  }
}
`, testhelper.ConfigDataRootDC1(), AlarmResource)
}
func TestAccResourceVsphereAlarm_metricExpression(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceAlarmCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereAlarmConfigMetricExpression(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceAlarmCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "metric_expression.0.operator", "isAbove"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "metric_expression.0.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "metric_expression.0.yellow", "9900"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "metric_expression.0.yellow_interval", "300"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "metric_expression.0.metric_counter_id", "2")),
			},
		},
	})
}

func testAccResourceVsphereAlarmConfigMetricExpression() string {
	return fmt.Sprintf(`
%s

resource vsphere_alarm "%s" {
  name        = "Host CPU above 90 percent"
  description = "This is a test alarm"
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"

  metric_expression {
    metric_counter_id = 2
    operator          = "isAbove"
    object_type       = "HostSystem"
	yellow            = 9900
	yellow_interval   = 300
  }
}
`, testhelper.ConfigDataRootDC1(), AlarmResource)
}

func TestAccResourceVsphereAlarm_stateExpression(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceAlarmCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereAlarmConfigStateExpression(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceAlarmCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "state_expression.0.operator", "isEqual"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "state_expression.0.object_type", "HostSystem"),
					resource.TestCheckResourceAttr("vsphere_alarm."+AlarmResource, "state_expression.0.yellow", "disconnected"),
				),
			},
		},
	})
}

func testAccResourceVsphereAlarmConfigStateExpression() string {
	return fmt.Sprintf(`
%s

resource vsphere_alarm "%s" {
  name        = "Host disconnected"
  description = "This is a test alarm"
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"

  state_expression {
    operator    = "isEqual"
    state_path  = "runtime.connectionState"
    object_type = "HostSystem"
	yellow      = "disconnected"
  }
}
`, testhelper.ConfigDataRootDC1(), AlarmResource)
}
