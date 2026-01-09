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
	"github.com/vmware/govmomi/vapi/esx/settings/clusters/configuration/drafts"
	"github.com/vmware/govmomi/vapi/esx/settings/clusters/enablement"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/configprofile"
)

func resourceVSphereConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSphereConfigurationProfileCreate,
		ReadContext:   resourceVSphereConfigurationProfileRead,
		UpdateContext: resourceVSphereConfigurationProfileUpdate,
		DeleteContext: resourceVSphereConfigurationProfileDelete,
		Schema: map[string]*schema.Schema{
			"reference_host_id": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"configuration"},
				Description:   "The identifier of the host to use as a source of the configuration.",
			},
			"configuration": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{"reference_host_id"},
				Description:      "The configuration JSON.",
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

func resourceVSphereConfigurationProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	referenceHostID := d.Get("reference_host_id").(string)
	config := d.Get("configuration").(string)

	if referenceHostID != "" && config != "" {
		return diag.FromErr(fmt.Errorf("cannot specify both `reference_host_id` and `configuration`"))
	}

	clusterID := d.Get("cluster_id").(string)

	client := meta.(*Client).restClient
	m := enablement.NewManager(client)
	tm := tasks.NewManager(client)

	tflog.Debug(ctx, fmt.Sprintf("running eligibility checks on cluster: %s", clusterID))

	statusInfo, err := m.GetClusterConfigurationStatus(clusterID)
	if err != nil {
		return diag.FromErr(err)
	}

	if statusInfo.Status == "NOT_STARTED" {
		if taskID, err := m.CheckEligibility(clusterID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to run eligibility check: %s", err))
		} else if _, err := tm.WaitForCompletion(ctx, taskID); err != nil {
			return diag.FromErr(err)
		}

		if referenceHostID != "" {
			tflog.Debug(ctx, fmt.Sprintf("importing configuration from reference host: %s", referenceHostID))
			if taskID, err := m.ImportFromReferenceHost(clusterID, referenceHostID); err != nil {
				return diag.FromErr(fmt.Errorf("failed to import configuration from reference host: %s", err))
			} else if _, err := tm.WaitForCompletion(ctx, taskID); err != nil {
				return diag.FromErr(err)
			}
		} else {
			tflog.Debug(ctx, "using configuration json")
			spec := enablement.FileSpec{Config: config}
			if _, err := m.ImportFromFile(clusterID, spec); err != nil {
				return diag.FromErr(fmt.Errorf("failed to import configuration: %s", err))
			}
		}

		tflog.Debug(ctx, "validating imported configuration")
		if taskID, err := m.ValidateConfiguration(clusterID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to validate configuration: %s", err))
		} else if _, err := tm.WaitForCompletion(ctx, taskID); err != nil {
			return diag.FromErr(err)
		}

		tflog.Debug(ctx, fmt.Sprintf("running pre-checks on cluster: %s", clusterID))
		if taskID, err := m.RunPrecheck(clusterID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to run precheck: %s", err))
		} else if _, err := tm.WaitForCompletion(ctx, taskID); err != nil {
			return diag.FromErr(err)
		}

		tflog.Debug(ctx, fmt.Sprintf("transitioning cluster %s to configuration profiles", clusterID))
		if taskID, err := m.EnableClusterConfiguration(clusterID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to enable cluster configuration: %s", err))
		} else if _, err := tm.WaitForCompletion(ctx, taskID); err != nil {
			return diag.FromErr(err)
		}

		return resourceVSphereConfigurationProfileRead(ctx, d, meta)
	}

	// The target cluster is already using configuration profiles
	// Defer to the update routine
	return resourceVSphereConfigurationProfileUpdate(ctx, d, meta)
}

func resourceVSphereConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).restClient
	return configprofile.ReadConfigProfile(ctx, client, d)
}

func resourceVSphereConfigurationProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clusterID := d.Get("cluster_id").(string)

	client := meta.(*Client).restClient

	m := drafts.NewManager(client)

	tflog.Debug(ctx, fmt.Sprintf("looking for pending configuration draft on cluster: %s", clusterID))
	draftsList, err := m.ListDrafts(clusterID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to list drafts: %s", err))
	}

	if len(draftsList) > 0 {
		// there is only one active draft per user
		for draftID := range draftsList {
			tflog.Debug(ctx, fmt.Sprintf("deleting pending configuration draft: %s", draftID))
			if err := m.DeleteDraft(clusterID, draftID); err != nil {
				return diag.FromErr(fmt.Errorf("failed to delete draft: %s", err))
			}
		}
	}

	var createSpec drafts.CreateSpec

	if configuration := d.Get("configuration"); configuration != "" {
		tflog.Debug(ctx, "updating cluster configuration")
		createSpec.Config = configuration.(string)
	}

	tflog.Debug(ctx, "creating a new draft")
	draftID, err := m.CreateDraft(clusterID, createSpec)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to create draft: %s", err))
	}

	if referenceHostID := d.Get("reference_host_id"); referenceHostID != "" {
		tflog.Debug(ctx, fmt.Sprintf("updating cluster configuration using reference host: %s", referenceHostID.(string)))
		importSpec := drafts.ImportSpec{
			Host: referenceHostID.(string),
		}

		taskID, err := m.ImportFromHost(clusterID, draftID, importSpec)
		if err != nil {
			return diag.FromErr(fmt.Errorf("failed to import configuration: %s", err))
		}

		if _, err := tasks.NewManager(client).WaitForCompletion(ctx, taskID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to import configuration: %s", err))
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("running pre-checks for draft: %s", draftID))
	taskID, err := m.Precheck(clusterID, draftID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to trigger precheck: %s", err))
	}

	if _, err := tasks.NewManager(client).WaitForCompletion(ctx, taskID); err != nil {
		return diag.FromErr(fmt.Errorf("precheck failed: %s", err))
	}

	tflog.Debug(ctx, fmt.Sprintf("applying draft: %s", draftID))
	res, err := m.ApplyDraft(clusterID, draftID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to apply draft: %s", err))
	}

	if _, err := tasks.NewManager(client).WaitForCompletion(ctx, res.ApplyTask); err != nil {
		return diag.FromErr(fmt.Errorf("failed to apply draft: %s", err))
	}

	return resourceVSphereConfigurationProfileRead(ctx, d, meta)
}

func resourceVSphereConfigurationProfileDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// Can't go back after management via config profiles is enabled
	return nil
}

func configDiffSuppressFunc(_, oldVal, newVal string, _ *schema.ResourceData) bool {
	var oldMap map[string]interface{}

	if err := json.Unmarshal([]byte(oldVal), &oldMap); err != nil {
		return false
	}

	var newMap map[string]interface{}
	if err := json.Unmarshal([]byte(newVal), &newMap); err != nil {
		return false
	}

	delete(oldMap, "metadata")
	delete(newMap, "metadata")

	return reflect.DeepEqual(oldMap, newMap)
}
