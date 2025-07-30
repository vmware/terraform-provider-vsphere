// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/cis/tasks"
	"github.com/vmware/govmomi/vapi/esx/settings/clusters/configuration"
	"github.com/vmware/govmomi/vapi/esx/settings/clusters/configuration/drafts"
	"github.com/vmware/govmomi/vapi/esx/settings/clusters/enablement"
)

func resourceVSphereConfigProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSphereConfigProfileCreate,
		ReadContext:   resourceVSphereConfigProfileRead,
		UpdateContext: resourceVSphereConfigProfileUpdate,
		DeleteContext: resourceVSphereConfigProfileDelete,
		Schema: map[string]*schema.Schema{
			"reference_host_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"config"},
				Description:   "The identifier of the host to use as a source of the configuration.",
			},
			"config": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{"reference_host_id"},
				Description:      "The configuration json.",
				DiffSuppressFunc: configDiffSuppressFunc,
			},
			"schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration schema.",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the cluster that will be configured.",
			},
		},
	}
}

func resourceVSphereConfigProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	referenceHostId := d.Get("reference_host_id").(string)
	config := d.Get("config").(string)

	if referenceHostId != "" && config != "" {
		return diag.FromErr(fmt.Errorf("cannot specify both `reference_host_id` and `config`"))
	}

	clusterId := d.Get("cluster_id").(string)

	client := meta.(*Client).restClient
	m := enablement.NewManager(client)
	tm := tasks.NewManager(client)

	tflog.Debug(ctx, fmt.Sprintf("running eligibility checks on cluster: %s", clusterId))
	if taskId, err := m.CheckEligibility(clusterId); err != nil {
		return diag.FromErr(fmt.Errorf("failed to run eligibility check: %s", err))
	} else if _, err := tm.WaitForCompletion(ctx, taskId); err != nil {
		return diag.FromErr(err)
	}

	if referenceHostId != "" {
		tflog.Debug(ctx, fmt.Sprintf("importing configuration from reference host: %s", referenceHostId))
		if taskId, err := m.ImportFromReferenceHost(clusterId, referenceHostId); err != nil {
			return diag.FromErr(fmt.Errorf("failed to import configuration from reference host: %s", err))
		} else if _, err := tm.WaitForCompletion(ctx, taskId); err != nil {
			return diag.FromErr(err)
		}
	} else {
		tflog.Debug(ctx, "using configuration json")
		spec := enablement.FileSpec{Config: config}
		if _, err := m.ImportFromFile(clusterId, spec); err != nil {
			return diag.FromErr(fmt.Errorf("failed to import configuration: %s", err))
		}
	}

	tflog.Debug(ctx, "validating imported configuration")
	if taskId, err := m.ValidateConfiguration(clusterId); err != nil {
		return diag.FromErr(fmt.Errorf("failed to validate configuration: %s", err))
	} else if _, err := tm.WaitForCompletion(ctx, taskId); err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, fmt.Sprintf("running pre-checks on cluster: %s", clusterId))
	if taskId, err := m.RunPrecheck(clusterId); err != nil {
		return diag.FromErr(fmt.Errorf("failed to run precheck: %s", err))
	} else if _, err := tm.WaitForCompletion(ctx, taskId); err != nil {
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, fmt.Sprintf("transitioning cluster %s to configuration profiles", clusterId))
	if taskId, err := m.EnableClusterConfiguration(clusterId); err != nil {
		return diag.FromErr(fmt.Errorf("failed to enable cluster configuration: %s", err))
	} else if _, err := tm.WaitForCompletion(ctx, taskId); err != nil {
		return diag.FromErr(err)
	}

	return resourceVSphereConfigProfileRead(ctx, d, meta)
}

func resourceVSphereConfigProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).restClient
	m := configuration.NewManager(client)

	clusterId := d.Get("cluster_id").(string)

	d.SetId(fmt.Sprintf("config_profile_%s", clusterId))

	tflog.Debug(ctx, fmt.Sprintf("reading configuration for cluster: %s", clusterId))
	config, err := m.GetConfiguration(clusterId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve cluster configuration: %s", err))
	}

	_ = d.Set("config", config.Config)

	tflog.Debug(ctx, fmt.Sprintf("reading configuration schema for cluster: %s", clusterId))
	configSchema, err := m.GetSchema(clusterId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve configuration schema: %s", err))
	}

	_ = d.Set("schema", configSchema.Schema)

	return nil
}

func resourceVSphereConfigProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterId := d.Get("cluster_id").(string)

	client := meta.(*Client).restClient

	m := drafts.NewManager(client)

	tflog.Debug(ctx, fmt.Sprintf("looking for pending configuration draft on cluster: %s", clusterId))
	draftsList, err := m.ListDrafts(clusterId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list drafts: %s", err))
	}

	if len(draftsList) > 0 {
		// there is only one active draft
		for draftId, _ := range draftsList {
			tflog.Debug(ctx, fmt.Sprintf("deleting pending configuration draft: %s", draftId))
			if err := m.DeleteDraft(clusterId, draftId); err != nil {
				return diag.FromErr(fmt.Errorf("failed to delete draft: %s", err))
			}
		}
	}

	var createSpec drafts.CreateSpec

	if d.HasChange("reference_host_id") {
		referenceHostId := d.Get("reference_host_id").(string)
		tflog.Debug(ctx, fmt.Sprintf("updating cluster configuration using reference host: %s", referenceHostId))
		createSpec.ReferenceHost = referenceHostId
	} else {
		tflog.Debug(ctx, "updating cluster configuration")
		createSpec.Config = d.Get("config").(string)
	}

	tflog.Debug(ctx, "creating a new draft")
	draftId, err := m.CreateDraft(clusterId, createSpec)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create draft: %s", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("running pre-checks for draft: %s", draftId))
	taskId, err := m.Precheck(clusterId, draftId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to trigger precheck: %s", err))
	}

	if _, err := tasks.NewManager(client).WaitForCompletion(ctx, taskId); err != nil {
		return diag.FromErr(fmt.Errorf("precheck failed: %s", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("applying draft: %s", draftId))
	res, err := m.ApplyDraft(clusterId, draftId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to apply draft: %s", err))
	}

	if _, err := tasks.NewManager(client).WaitForCompletion(ctx, res.ApplyTask); err != nil {
		return diag.FromErr(fmt.Errorf("failed to apply draft: %s", err))
	}

	return resourceVSphereConfigProfileRead(ctx, d, meta)
}

func resourceVSphereConfigProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Can't go back after management via config profiles is enabled
	return nil
}

func configDiffSuppressFunc(_, old, new string, _ *schema.ResourceData) bool {
	var oldMap map[string]interface{}

	if err := json.Unmarshal([]byte(old), &oldMap); err != nil {
		return false
	}

	var newMap map[string]interface{}
	if err := json.Unmarshal([]byte(new), &newMap); err != nil {
		return false
	}

	delete(oldMap, "metadata")
	delete(newMap, "metadata")

	return reflect.DeepEqual(oldMap, newMap)
}
