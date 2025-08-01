// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vim25/types"
	helper "github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/alarm"
)

func dataSourceVSphereAlarm() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereAlarmRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the alarm.",
				Required:    true,
			},
			"entity_type": {
				Type:        schema.TypeString,
				Description: "The entity managed object type.",
				Required:    true,
			},
			"entity_id": {
				Type:        schema.TypeString,
				Description: "The managed object id or uuid of the entity.",
				Required:    true,
			},
			// computed
			"description": {
				Type:        schema.TypeString,
				Description: "Description of the alarm.",
				Computed:    true,
			},
			"enabled": {
				Type:        schema.TypeBool,
				Description: "Whether or not the alarm is enabled.",
				Computed:    true,
			},
			"expression_operator": {
				Type:        schema.TypeString,
				Description: "The logical operator between alarm expressions.",
				Computed:    true,
			},
			"event_expression": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Create an event expression that could trigger the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of Event (vim.event.Event).",
						},
						"event_type_id": {
							Type:        schema.TypeString,
							Description: "Name of the event (vim.event).",
							Computed:    true,
						},
						"object_type": {
							Type:        schema.TypeString,
							Description: "Type of object where the event applies on.",
							Computed:    true,
						},
						"status": {
							Type:        schema.TypeString,
							Description: "Alarm status once triggered.",
							Computed:    true,
						},
						"comparison": {
							Type:        schema.TypeList,
							Description: "Additional check that allows adding threshold on the given object attribute.",
							Optional:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attribute_name": {
										Computed:    true,
										Description: "Name of the attribute to compare.",
										Type:        schema.TypeString,
									},
									"operator": {
										Computed:    true,
										Description: "Comparison operator.",
										Type:        schema.TypeString,
									},
									"value": {
										Computed:    true,
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
							Type:        schema.TypeString,
							Description: "Check if state is equal or unequal.",
							Computed:    true,
						},
						"object_type": {
							Type:        schema.TypeString,
							Description: "Type of object where the event applies on, ie: HostSystem.",
							Computed:    true,
						},
						"state_path": {
							Type:        schema.TypeString,
							Description: "State path: ie. runtime.connectionState.",
							Computed:    true,
						},
						"yellow": {
							Type:        schema.TypeString,
							Description: "State value to trigger warning alarm.",
							Computed:    true,
						},
						"red": {
							Type:        schema.TypeString,
							Description: "State value to trigger critical alarm.",
							Computed:    true,
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
							Type:        schema.TypeString,
							Description: "Whether the metric is below or above the given threshold.",
							Computed:    true,
						},
						"object_type": {
							Type:        schema.TypeString,
							Description: "Type of object of the metric, ie: HostSystem.",
							Computed:    true,
						},
						"metric_counter_id": {
							Type:        schema.TypeInt,
							Description: "ID of the metric.",
							Computed:    true,
						},
						"metric_instance": {
							Type:        schema.TypeString,
							Description: "",
							Computed:    true,
						},
						"yellow": {
							Type:        schema.TypeInt,
							Description: "Warning threshold, for percentage, 9900 is 99%.",
							Computed:    true,
						},
						"red": {
							Type:        schema.TypeInt,
							Description: "Critical threshold, for percentage, 9900 is 99%.",
							Computed:    true,
						},
						"yellow_interval": {
							Type:        schema.TypeInt,
							Description: "Amount of seconds the threshold must be crossed to trigger the warning alarm.",
							Computed:    true,
						},
						"red_interval": {
							Type:        schema.TypeInt,
							Description: "Amount of seconds the threshold must be crossed to trigger the critical alarm.",
							Computed:    true,
						},
					},
				},
			},
			"snmp_action": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The SNMP action to define in the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_state": {
							Type:        schema.TypeString,
							Description: "Triggers the action only for this initial state.",
							Computed:    true,
						},
						"final_state": {
							Type:        schema.TypeString,
							Description: "Triggers the action only for this final state.",
							Computed:    true,
						},
						"repeat": {
							Type:        schema.TypeBool,
							Description: "Whether or not the action should be repeated.",
							Computed:    true,
						},
					},
				},
			},
			"email_action": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The email action to define in the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_state": {
							Type:        schema.TypeString,
							Description: "Triggers the action only for this initial state.",
							Computed:    true,
						},
						"final_state": {
							Type:        schema.TypeString,
							Description: "Triggers the action only for this final state.",
							Computed:    true,
						},
						"repeat": {
							Type:        schema.TypeBool,
							Description: "Whether or not the action should be repeated.",
							Computed:    true,
						},
						"to": {
							Type:        schema.TypeString,
							Description: "Email destination.",
							Computed:    true,
						},
						"cc": {
							Type:        schema.TypeString,
							Description: "Email destination cc.",
							Computed:    true,
						},
						"subject": {
							Type:        schema.TypeString,
							Description: "Email subject.",
							Computed:    true,
						},
						"body": {
							Type:        schema.TypeString,
							Description: "Email body.",
							Computed:    true,
						},
					},
				},
			},
			"advanced_action": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The advanced action to define in the alarm.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_state": {
							Type:        schema.TypeString,
							Description: "Triggers the action only for this initial state.",
							Computed:    true,
						},
						"final_state": {
							Type:        schema.TypeString,
							Description: "Triggers the action only for this final state.",
							Computed:    true,
						},
						"repeat": {
							Type:        schema.TypeBool,
							Description: "Whether or not the action should be repeated.",
							Computed:    true,
						},
						"name": {
							Computed:    true,
							Description: "Name of the action to perform.",
							Type:        schema.TypeString,
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereAlarmRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	entity, err := helper.FindEntity(client, d.Get("entity_type").(string), d.Get("entity_id").(string))
	if err != nil {
		return fmt.Errorf("alarm entity error: %s", err)
	}

	al, err := helper.FromName(client, d.Get("name").(string), entity)
	if err != nil {
		return fmt.Errorf("cannot locate alarm: %s", err)
	}

	d.SetId(al.Reference().Value)
	_ = d.Set("description", al.Info.Description)

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
