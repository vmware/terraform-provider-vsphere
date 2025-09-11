// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package alarm

import (
	"context"
	"fmt"
	"reflect"
	"unicode"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/alarm"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/provider"
)

type VsphereAdvancedActionName string

const (
	VsphereAdvancedActionNameEnterMaintenance = VsphereAdvancedActionName("EnterMaintenanceMode_Task")
	VsphereAdvancedActionNameExitMaintenance  = VsphereAdvancedActionName("ExitMaintenanceMode_Task")
	VsphereAdvancedActionNameEnterStandby     = VsphereAdvancedActionName("PowerDownHostToStandBy_Task")
	VsphereAdvancedActionNameExitStandby      = VsphereAdvancedActionName("PowerUpHostFromStandBy_Task")
	VsphereAdvancedActionNameRebootHost       = VsphereAdvancedActionName("RebootHost_Task")
	VsphereAdvancedActionNameShutdown         = VsphereAdvancedActionName("ShutdownHost_Task")
)

// Return the object reference from an object type and ID
func FindEntity(client *govmomi.Client, objType string, id string) (object.Reference, error) {
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  objType,
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return finder.ObjectReference(ctx, ref)
}

// Retrieve alarm on the given object
func getAlarms(client *govmomi.Client, entity object.Reference) ([]mo.Alarm, error) {
	m, err := alarm.GetManager(client.Client)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return m.GetAlarm(ctx, entity)
}

// FromID locates an alarm on a given entity by its managed object reference ID.
func FromID(client *govmomi.Client, id string, entity object.Reference) (*mo.Alarm, error) {
	alarms, err := getAlarms(client, entity)
	if err != nil {
		return nil, err
	}
	for _, al := range alarms {
		if al.Reference().Value == id {
			return &al, nil
		}
	}

	return nil, fmt.Errorf("alarm %s not found", id)
}

// FromName locates and alarm on a given entity from its name
func FromName(client *govmomi.Client, name string, entity object.Reference) (*mo.Alarm, error) {
	alarms, err := getAlarms(client, entity)
	if err != nil {
		return nil, err
	}
	for _, al := range alarms {
		if al.Info.Name == name {
			return &al, nil
		}
	}

	return nil, fmt.Errorf("alarm %s not found", name)
}

func UcFirst(s string) string {
	r := []rune(s)
	return string(append([]rune{unicode.ToUpper(r[0])}, r[1:]...))
}

// Store all kinds of expressions
type Expressions struct {
	EventExpressions  []map[string]any
	StateExpressions  []map[string]any
	MetricExpressions []map[string]any
}

// Craft expressions tfstate
func GetExpressions(b []types.BaseAlarmExpression) (Expressions, error) {
	expressions := Expressions{
		EventExpressions:  []map[string]any{},
		StateExpressions:  []map[string]any{},
		MetricExpressions: []map[string]any{},
	}
	for _, e := range b {
		switch subExp := e.(type) {
		case *types.EventAlarmExpression:
			comps := make([]map[string]any, len(subExp.Comparisons))
			for i, c := range subExp.Comparisons {
				comps[i] = map[string]any{
					"attribute_name": c.AttributeName,
					"operator":       c.Operator,
					"value":          c.Value,
				}
			}
			expressions.EventExpressions = append(expressions.EventExpressions, map[string]any{
				"event_type":    subExp.EventType,
				"event_type_id": subExp.EventTypeId,
				"object_type":   subExp.ObjectType,
				"comparison":    comps,
				"status":        subExp.Status,
			})
		case *types.StateAlarmExpression:
			expressions.StateExpressions = append(expressions.StateExpressions, map[string]any{
				"operator":    subExp.Operator,
				"object_type": subExp.Type,
				"state_path":  subExp.StatePath,
				"yellow":      subExp.Yellow,
				"red":         subExp.Red,
			})
		case *types.MetricAlarmExpression:
			expressions.MetricExpressions = append(expressions.MetricExpressions, map[string]any{
				"operator":          subExp.Operator,
				"object_type":       subExp.Type,
				"metric_counter_id": subExp.Metric.CounterId,
				"metric_instance":   subExp.Metric.Instance,
				"yellow":            subExp.Yellow,
				"red":               subExp.Red,
				"red_interval":      subExp.RedInterval,
				"yellow_interval":   subExp.YellowInterval,
			})
		default:
			return expressions, fmt.Errorf("unknown expression type: %s", reflect.TypeOf(subExp))
		}
	}
	return expressions, nil
}

// Store all kinds of actions
type Actions struct {
	EmailAction    []map[string]any
	SnmpAction     []map[string]any
	AdvancedAction []map[string]any
}

// Craft actions tfstate
func GetAlarmActions(b []types.BaseAlarmAction) (Actions, error) {
	actions := Actions{
		EmailAction:    []map[string]any{},
		SnmpAction:     []map[string]any{},
		AdvancedAction: []map[string]any{},
	}
	for _, a := range b {
		switch action := a.(type) {
		case *types.AlarmTriggeringAction:
			switch at := action.Action.(type) {
			case *types.SendEmailAction:
				actions.EmailAction = append(actions.EmailAction, map[string]any{
					"to":          at.ToList,
					"cc":          at.CcList,
					"subject":     at.Subject,
					"body":        at.Body,
					"final_state": action.TransitionSpecs[0].FinalState,
					"start_state": action.TransitionSpecs[0].StartState,
					"repeat":      action.TransitionSpecs[0].Repeats,
				})
			case *types.SendSNMPAction:
				actions.SnmpAction = append(actions.SnmpAction, map[string]any{
					"final_state": action.TransitionSpecs[0].FinalState,
					"start_state": action.TransitionSpecs[0].StartState,
					"repeat":      action.TransitionSpecs[0].Repeats,
				})
			case *types.MethodAction:
				actions.AdvancedAction = append(actions.AdvancedAction, map[string]any{
					"name":        at.Name,
					"final_state": action.TransitionSpecs[0].FinalState,
					"start_state": action.TransitionSpecs[0].StartState,
					"repeat":      action.TransitionSpecs[0].Repeats,
				})
			}
		default:
			return actions, fmt.Errorf("unknown expression type: %s", reflect.TypeOf(a))
		}
	}
	return actions, nil
}

func getStatusFromString(s string) (types.ManagedEntityStatus, error) {
	switch s {
	case "red":
		return types.ManagedEntityStatusRed, nil
	case "yellow":
		return types.ManagedEntityStatusYellow, nil
	case "green":
		return types.ManagedEntityStatusGreen, nil
	case "gray":
		return types.ManagedEntityStatusGray, nil
	}
	return "", fmt.Errorf("unknown status: %s", s)
}

func getMetricOperatorFromString(s string) (types.MetricAlarmOperator, error) {
	switch s {
	case "isAbove":
		return types.MetricAlarmOperatorIsAbove, nil
	case "isBelow":
		return types.MetricAlarmOperatorIsBelow, nil
	}
	return "", fmt.Errorf("unknown metric operator: %s", s)
}

func getStateAlarmOperatorFromString(s string) (types.StateAlarmOperator, error) {
	switch s {
	case "isEqual":
		return types.StateAlarmOperatorIsEqual, nil
	case "isUnequal":
		return types.StateAlarmOperatorIsUnequal, nil
	}
	return "", fmt.Errorf("unknown state alarm operator: %s", s)
}

func GetAlarmSpec(d *schema.ResourceData) (*types.AlarmSpec, error) {
	// Expressions
	fromExpressions, err := GetBaseExpressions(d)
	if err != nil {
		return nil, fmt.Errorf("failed to compute expressions: %s", err)
	}

	if len(fromExpressions) == 0 {
		return nil, fmt.Errorf("an alarm must be contain at least one expression")
	}
	var expressions types.BaseAlarmExpression
	switch d.Get("expression_operator").(string) {
	case "or":
		expressions = &types.OrAlarmExpression{
			Expression: fromExpressions,
		}
	case "and":
		expressions = &types.AndAlarmExpression{
			Expression: fromExpressions,
		}
	}

	// Actions
	actions, err := GetBaseActions(d)
	if err != nil {
		return nil, fmt.Errorf("failed to compute actions: %s", err)
	}

	// Final specs
	return &types.AlarmSpec{
		Name:            d.Get("name").(string),
		Description:     d.Get("description").(string),
		Enabled:         d.Get("enabled").(bool),
		SystemName:      "",
		Expression:      expressions,
		Action:          actions,
		ActionFrequency: 0,
		Setting: &types.AlarmSetting{
			ToleranceRange:     0,
			ReportingFrequency: 300,
		},
	}, nil
}

// Generate transitionSpec from schema
func getActionTransitionSpecs(d *schema.ResourceData, path string) (*types.AlarmTriggeringActionTransitionSpec, error) {
	startState, err := getStatusFromString(d.Get(fmt.Sprintf("%s.start_state", path)).(string))
	if err != nil {
		return nil, err
	}
	finalState, err := getStatusFromString(d.Get(fmt.Sprintf("%s.final_state", path)).(string))
	if err != nil {
		return nil, err
	}
	return &types.AlarmTriggeringActionTransitionSpec{
		StartState: startState,
		FinalState: finalState,
		Repeats:    d.Get(fmt.Sprintf("%s.repeat", path)).(bool),
	}, nil
}

// Generate action specs from schema
func GetBaseActions(d *schema.ResourceData) (types.BaseAlarmAction, error) {
	var actions = &types.GroupAlarmAction{
		Action: []types.BaseAlarmAction{},
	}
	// Email actions
	for i := range d.Get("email_action").([]any) {
		path := fmt.Sprintf("email_action.%d", i)
		transitionSpecs, err := getActionTransitionSpecs(d, path)
		if err != nil {
			return nil, err
		}
		actions.Action = append(actions.Action, &types.AlarmTriggeringAction{
			Action: &types.SendEmailAction{
				Subject: d.Get(fmt.Sprintf("%s.subject", path)).(string),
				ToList:  d.Get(fmt.Sprintf("%s.to", path)).(string),
				CcList:  d.Get(fmt.Sprintf("%s.cc", path)).(string),
				Body:    d.Get(fmt.Sprintf("%s.body", path)).(string),
			},
			TransitionSpecs: []types.AlarmTriggeringActionTransitionSpec{*transitionSpecs},
		})
	}

	// Snmp actions
	for i := range d.Get("snmp_action").([]any) {
		path := fmt.Sprintf("snmp_action.%d", i)
		transitionSpecs, err := getActionTransitionSpecs(d, path)
		if err != nil {
			return nil, err
		}
		actions.Action = append(actions.Action, &types.AlarmTriggeringAction{
			Action:          &types.SendSNMPAction{},
			TransitionSpecs: []types.AlarmTriggeringActionTransitionSpec{*transitionSpecs},
		})
	}

	// Advanced actions
	for i := range d.Get("advanced_action").([]any) {
		path := fmt.Sprintf("advanced_action.%d", i)
		transitionSpecs, err := getActionTransitionSpecs(d, path)
		if err != nil {
			return nil, err
		}

		actionName := d.Get(fmt.Sprintf("%s.name", path)).(string)
		var args []types.MethodActionArgument

		// Hardcoded default task parameters
		switch actionName {
		case string(VsphereAdvancedActionNameEnterStandby):
			args = []types.MethodActionArgument{{Value: int32(0)}, {Value: false}}
		case string(VsphereAdvancedActionNameExitStandby):
			args = []types.MethodActionArgument{{Value: int32(0)}}
		case string(VsphereAdvancedActionNameEnterMaintenance):
			args = []types.MethodActionArgument{{Value: int32(0)}, {Value: false}}
		case string(VsphereAdvancedActionNameExitMaintenance):
			args = []types.MethodActionArgument{{Value: int32(0)}}
		case string(VsphereAdvancedActionNameRebootHost):
			args = []types.MethodActionArgument{{Value: false}}
		case string(VsphereAdvancedActionNameShutdown):
			args = []types.MethodActionArgument{{Value: false}}
		}

		actions.Action = append(actions.Action, &types.AlarmTriggeringAction{
			Action: &types.MethodAction{
				Name:     actionName,
				Argument: args,
			},
			TransitionSpecs: []types.AlarmTriggeringActionTransitionSpec{*transitionSpecs},
		})
	}

	// GroupAlarmAction cannot be empty
	if len(actions.Action) == 0 {
		return nil, nil
	}
	return actions, nil
}

// Generate expression specs from schema
func GetBaseExpressions(d *schema.ResourceData) ([]types.BaseAlarmExpression, error) {
	var fromExpressions []types.BaseAlarmExpression
	// Event alarm expressions
	for i := range d.Get("event_expression").([]any) {
		path := fmt.Sprintf("event_expression.%d", i)
		status, err := getStatusFromString(d.Get(fmt.Sprintf("%s.status", path)).(string))
		if err != nil {
			return nil, fmt.Errorf("alarm event_expression error: %s", err)
		}
		comparisons := make([]types.EventAlarmExpressionComparison, len(d.Get(fmt.Sprintf("%s.comparison", path)).([]any)))
		for j := range d.Get(fmt.Sprintf("%s.comparison", path)).([]any) {
			comparisons[j] = types.EventAlarmExpressionComparison{
				AttributeName: d.Get(fmt.Sprintf("%s.comparison.%d.attribute_name", path, j)).(string),
				Operator:      d.Get(fmt.Sprintf("%s.comparison.%d.operator", path, j)).(string),
				Value:         d.Get(fmt.Sprintf("%s.comparison.%d.value", path, j)).(string),
			}
		}

		fromExpressions = append(fromExpressions, &types.EventAlarmExpression{
			ObjectType:  d.Get(fmt.Sprintf("%s.object_type", path)).(string),
			EventType:   d.Get(fmt.Sprintf("%s.event_type", path)).(string),
			Status:      status,
			EventTypeId: d.Get(fmt.Sprintf("%s.event_type_id", path)).(string),
			Comparisons: comparisons,
		})
	}

	// State alarm expressions
	for i := range d.Get("state_expression").([]any) {
		op, err := getStateAlarmOperatorFromString(d.Get(fmt.Sprintf("state_expression.%d.operator", i)).(string))
		if err != nil {
			return nil, err
		}
		fromExpressions = append(fromExpressions, &types.StateAlarmExpression{
			Operator:  op,
			Type:      d.Get(fmt.Sprintf("state_expression.%d.object_type", i)).(string),
			StatePath: d.Get(fmt.Sprintf("state_expression.%d.state_path", i)).(string),
			Yellow:    d.Get(fmt.Sprintf("state_expression.%d.yellow", i)).(string),
			Red:       d.Get(fmt.Sprintf("state_expression.%d.red", i)).(string),
		})
	}

	// Metric alarm expressions
	for i := range d.Get("metric_expression").([]any) {
		op, err := getMetricOperatorFromString(d.Get(fmt.Sprintf("metric_expression.%d.operator", i)).(string))
		if err != nil {
			return nil, err
		}
		fromExpressions = append(fromExpressions, &types.MetricAlarmExpression{
			Operator: op,
			Type:     d.Get(fmt.Sprintf("metric_expression.%d.object_type", i)).(string),
			Metric: types.PerfMetricId{
				CounterId: int32(d.Get(fmt.Sprintf("metric_expression.%d.metric_counter_id", i)).(int)),
				Instance:  d.Get(fmt.Sprintf("metric_expression.%d.metric_instance", i)).(string),
			},
			Yellow:         int32(d.Get(fmt.Sprintf("metric_expression.%d.yellow", i)).(int)),
			YellowInterval: int32(d.Get(fmt.Sprintf("metric_expression.%d.yellow_interval", i)).(int)),
			Red:            int32(d.Get(fmt.Sprintf("metric_expression.%d.red", i)).(int)),
			RedInterval:    int32(d.Get(fmt.Sprintf("metric_expression.%d.red_interval", i)).(int)),
		})
	}
	return fromExpressions, nil
}
