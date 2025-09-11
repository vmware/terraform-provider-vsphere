// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/alarm"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	helper "github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/alarm"
)

var allowedAlarmStatus = []string{
	string(types.ManagedEntityStatusRed),
	string(types.ManagedEntityStatusYellow),
	string(types.ManagedEntityStatusGreen),
	string(types.ManagedEntityStatusGray),
}

var allowedMetricExpressionOperators = []string{
	string(types.MetricAlarmOperatorIsAbove),
	string(types.MetricAlarmOperatorIsBelow),
}

var allowedStateExpressionOperators = []string{
	string(types.StateAlarmOperatorIsEqual),
	string(types.StateAlarmOperatorIsUnequal),
}

var allowAdvancedActionName = []string{
	string(helper.VsphereAdvancedActionNameEnterMaintenance),
	string(helper.VsphereAdvancedActionNameExitMaintenance),
	string(helper.VsphereAdvancedActionNameEnterStandby),
	string(helper.VsphereAdvancedActionNameExitStandby),
	string(helper.VsphereAdvancedActionNameRebootHost),
	string(helper.VsphereAdvancedActionNameShutdown),
}

var allowedEventExpressionComparisonOperators = []string{
	string(types.EventAlarmExpressionComparisonOperatorEquals),
	string(types.EventAlarmExpressionComparisonOperatorNotEqualTo),
	string(types.EventAlarmExpressionComparisonOperatorDoesNotEndWith),
	string(types.EventAlarmExpressionComparisonOperatorDoesNotStartWith),
	string(types.EventAlarmExpressionComparisonOperatorEndsWith),
	string(types.EventAlarmExpressionComparisonOperatorStartsWith),
}

func resourceVSphereAlarm() *schema.Resource {
	return &schema.Resource{
		Create:        resourceVSphereAlarmCreate,
		Read:          resourceVSphereAlarmRead,
		Update:        resourceVSphereAlarmUpdate,
		Delete:        resourceVSphereAlarmDelete,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the alarm.",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the alarm.",
				Required:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether or not the alarm is enabled.",
				Default:     true,
				Optional:    true,
			},
			"entity_id": {
				Type:         schema.TypeString,
				Description:  "The managed object id or uuid of the entity.",
				ValidateFunc: validation.NoZeroValues,
				Required:     true,
				ForceNew:     true,
			},
			"entity_type": {
				Type:         schema.TypeString,
				Description:  "The entity managed object type.",
				ValidateFunc: validation.NoZeroValues,
				StateFunc:    func(i any) string { return helper.UcFirst(i.(string)) },
				Required:     true,
				ForceNew:     true,
			},
			"expression_operator": {
				Type:        schema.TypeString,
				Description: "The logical operator between alarm expressions.",
				Optional:    true,
				Default:     "or",
				ValidateFunc: validation.StringInSlice(
					[]string{"or", "and"},
					false,
				),
			},
			// expressions are split by type to handle all kind of different parameters
			"event_expression": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Create an event expression that could trigger the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "Event",
							Description: "Type of Event (vim.event.Event).",
							StateFunc:   func(i any) string { return helper.UcFirst(i.(string)) },
						},
						"event_type_id": {
							Type:        schema.TypeString,
							Description: "Name of the event (vim.event).",
							Required:    true,
						},
						"object_type": {
							Type:        schema.TypeString,
							Description: "Type of object where the event applies on.",
							Required:    true,
							StateFunc:   func(i any) string { return helper.UcFirst(i.(string)) },
						},
						"status": {
							Type:         schema.TypeString,
							Description:  "Alarm status once triggered.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"comparison": {
							Type:        schema.TypeList,
							Description: "Additional check that allows adding threshold on the given object attribute.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attribute_name": {
										Required:    true,
										Description: "Name of the attribute to compare.",
										Type:        schema.TypeString,
									},
									"operator": {
										Required:     true,
										Description:  "Comparison operator.",
										Type:         schema.TypeString,
										ValidateFunc: validation.StringInSlice(allowedEventExpressionComparisonOperators, false),
									},
									"value": {
										Required:    true,
										Description: "Value to compare.",
										Type:        schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
			"state_expression": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Create a state expression that could trigger the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operator": {
							Type:         schema.TypeString,
							Description:  "Check if state is equal or unequal.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedStateExpressionOperators, false),
						},
						"object_type": {
							Type:        schema.TypeString,
							Description: "Type of object where the event applies on, ie: HostSystem.",
							Required:    true,
						},
						"state_path": {
							Type:        schema.TypeString,
							Description: "State path: ie. runtime.connectionState.",
							Required:    true,
						},
						"yellow": {
							Type:        schema.TypeString,
							Description: "State value to trigger warning alarm.",
							Optional:    true,
						},
						"red": {
							Type:        schema.TypeString,
							Description: "State value to trigger critical alarm.",
							Optional:    true,
						},
					},
				},
			},
			"metric_expression": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Create a metric expression that could trigger the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"operator": {
							Type:         schema.TypeString,
							Description:  "Whether the metric is below or above the given threshold.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedMetricExpressionOperators, false),
						},
						"object_type": {
							Type:        schema.TypeString,
							Description: "Type of object of the metric, ie: HostSystem.",
							Required:    true,
						},
						"metric_counter_id": {
							Type:        schema.TypeInt,
							Description: "ID of the metric.",
							Required:    true,
						},
						"metric_instance": {
							Type:        schema.TypeString,
							Description: "",
							Optional:    true,
						},
						"yellow": {
							Type:         schema.TypeInt,
							Description:  "Warning threshold, for percentage, 9900 is 99%.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 10000),
						},
						"red": {
							Type:         schema.TypeInt,
							Description:  "Critical threshold, for percentage, 9900 is 99%.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 10000),
						},
						"yellow_interval": {
							Type:         schema.TypeInt,
							Description:  "Amount of seconds the threshold must be crossed to trigger the warning alarm.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 3600),
						},
						"red_interval": {
							Type:         schema.TypeInt,
							Description:  "Amount of seconds the threshold must be crossed to trigger the critical alarm.",
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 3600),
						},
					},
				},
			},
			// actions are also split by type to handles all kind of different parameters
			"snmp_action": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The SNMP action to define in the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_state": {
							Type:         schema.TypeString,
							Description:  "Triggers the action only for this initial state.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"final_state": {
							Type:         schema.TypeString,
							Description:  "Triggers the action only for this final state.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"repeat": {
							Type:        schema.TypeBool,
							Description: "Whether or not the action should be repeated.",
							Default:     false,
							Optional:    true,
						},
					},
				},
			},
			"email_action": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The email action to define in the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_state": {
							Type:         schema.TypeString,
							Description:  "Triggers the action only for this initial state.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"final_state": {
							Type:         schema.TypeString,
							Description:  "Triggers the action only for this final state.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"repeat": {
							Type:        schema.TypeBool,
							Description: "Whether or not the action should be repeated.",
							Default:     false,
							Optional:    true,
						},
						"to": {
							Type:        schema.TypeString,
							Description: "Email destination.",
							Optional:    true,
						},
						"cc": {
							Type:        schema.TypeString,
							Description: "Email destination cc.",
							Optional:    true,
						},
						"subject": {
							Type:        schema.TypeString,
							Description: "Email subject.",
							Optional:    true,
						},
						"body": {
							Type:        schema.TypeString,
							Description: "Email body.",
							Optional:    true,
						},
					},
				},
			},
			"advanced_action": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The advanced action to define in the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_state": {
							Type:         schema.TypeString,
							Description:  "Triggers the action only for this initial state.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"final_state": {
							Type:         schema.TypeString,
							Description:  "Triggers the action only for this final state.",
							Required:     true,
							ValidateFunc: validation.StringInSlice(allowedAlarmStatus, false),
						},
						"repeat": {
							Type:        schema.TypeBool,
							Description: "Whether or not the action should be repeated.",
							Default:     false,
							Optional:    true,
						},
						"name": {
							Optional:     true,
							Description:  "Name of the action to perform.",
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice(allowAdvancedActionName, false),
						},
					},
				},
			},
		},
	}
}

func resourceVSphereAlarmCreate(d *schema.ResourceData, meta any) error {
	client := meta.(*Client).vimClient
	m, err := alarm.GetManager(client.Client)
	if err != nil {
		return err
	}

	// Getting the object the alarm will be created on
	entity, err := helper.FindEntity(client, d.Get("entity_type").(string), d.Get("entity_id").(string))
	if err != nil {
		return fmt.Errorf("alarm entity error: %s", err)
	}

	alarmSpec, err := helper.GetAlarmSpec(d)
	if err != nil {
		return fmt.Errorf("failed to generate alarm spec: %s", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	ref, err := m.CreateAlarm(ctx, entity, alarmSpec)
	if err != nil {
		return fmt.Errorf("failed to create new alarm: %s", err)
	}
	d.SetId(ref.Reference().Value)
	return resourceVSphereAlarmRead(d, meta)
}

func resourceVSphereAlarmRead(d *schema.ResourceData, meta any) error {
	client := meta.(*Client).vimClient

	entity, err := helper.FindEntity(client, d.Get("entity_type").(string), d.Get("entity_id").(string))
	if err != nil {
		return fmt.Errorf("alarm entity error: %s", err)
	}

	al, err := helper.FromID(client, d.Id(), entity)
	if err != nil {
		return fmt.Errorf("cannot locate alarm: %s", err)
	}

	_ = d.Set("name", al.Info.Name)
	_ = d.Set("description", al.Info.Description)
	_ = d.Set("entity_type", al.Info.Entity.Type)
	_ = d.Set("entity_id", al.Info.Entity.Value)

	// Manage alarm expressions
	var expressions helper.Expressions
	switch exp := al.Info.Expression.(type) {
	case *types.OrAlarmExpression:
		_ = d.Set("expression_operator", "or")
		expressions, err = helper.GetExpressions(exp.Expression)
		if err != nil {
			return err
		}
	case *types.AndAlarmExpression:
		_ = d.Set("expression_operator", "and")
		expressions, err = helper.GetExpressions(exp.Expression)
		if err != nil {
			return err
		}
	}
	if len(expressions.EventExpressions) > 0 {
		_ = d.Set("event_expression", expressions.EventExpressions)
	}
	if len(expressions.StateExpressions) > 0 {
		_ = d.Set("state_expression", expressions.StateExpressions)
	}
	if len(expressions.MetricExpressions) > 0 {
		_ = d.Set("metric_expression", expressions.MetricExpressions)
	}

	// Manage actions
	if al.Info.Action != nil {
		var actions helper.Actions
		switch alarmAction := al.Info.Action.(type) {
		case *types.GroupAlarmAction:
			actions, err = helper.GetAlarmActions(alarmAction.Action)
		case *types.AlarmAction:
			actions, err = helper.GetAlarmActions([]types.BaseAlarmAction{alarmAction})
		default:
			return fmt.Errorf("unmanaged alarm action type: %s", reflect.TypeOf(alarmAction))
		}
		if err != nil {
			return err
		}
		if len(actions.EmailAction) > 0 {
			_ = d.Set("email_action", actions.EmailAction)
		}
		if len(actions.SnmpAction) > 0 {
			_ = d.Set("snmp_action", actions.SnmpAction)
		}
		if len(actions.AdvancedAction) > 0 {
			_ = d.Set("advanced_action", actions.AdvancedAction)
		}

	}
	return nil
}

func resourceVSphereAlarmUpdate(d *schema.ResourceData, meta any) error {
	client := meta.(*Client).vimClient

	entity, err := helper.FindEntity(client, d.Get("entity_type").(string), d.Get("entity_id").(string))
	if err != nil {
		return fmt.Errorf("alarm entity error: %s", err)
	}

	al, err := helper.FromID(client, d.Id(), entity)
	if err != nil {
		return fmt.Errorf("cannot locate alarm: %s", err)
	}

	alarmSpec, err := helper.GetAlarmSpec(d)
	if err != nil {
		return fmt.Errorf("failed to generate alarm spec: %s", err)
	}

	tctx, tcancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer tcancel()
	_, err = methods.ReconfigureAlarm(tctx, client.RoundTripper, &types.ReconfigureAlarm{
		This: al.Self,
		Spec: alarmSpec,
	})
	if err != nil {
		return fmt.Errorf("failed to reconfigure alarm: %s", err)
	}
	return resourceVSphereAlarmRead(d, meta)
}

func resourceVSphereAlarmDelete(d *schema.ResourceData, meta any) error {
	client := meta.(*Client).vimClient

	var ref types.ManagedObjectReference

	if !ref.FromString(d.Id()) {
		ref.Type = "Alarm"
		ref.Value = d.Id()
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	_, err := methods.RemoveAlarm(ctx, client.Client, &types.RemoveAlarm{
		This: ref,
	})
	if err != nil {
		return fmt.Errorf("cannot delete alarm: %s", err)
	}
	d.SetId("")
	return nil
}
