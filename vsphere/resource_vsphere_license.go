// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/license"
	"github.com/vmware/govmomi/vim25/types"
	helper "github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/license"
)

var (
	// ErrKeyNotFound is an error primarily thrown by the Read method of the resource.
	ErrKeyNotFound = errors.New("the license key was not found")
	// ErrKeyNotDeleted is an error which occurs when a license key that is being removed.
	ErrKeyNotDeleted = errors.New("the license key was not deleted")
)

func resourceVSphereLicense() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,

		CreateContext: resourceVSphereLicenseCreate,
		ReadContext:   resourceVSphereLicenseRead,
		UpdateContext: resourceVSphereLicenseUpdate,
		DeleteContext: resourceVSphereLicenseDelete,

		Schema: map[string]*schema.Schema{
			"license_key": {
				Type:        schema.TypeString,
				Description: "The license key value.",
				Required:    true,
				ForceNew:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: "A map of labels to be applied to the license key.",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"edition_key": {
				Type:        schema.TypeString,
				Description: "The product edition of the license key.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "The display name for the license key.",
				Computed:    true,
			},
			"total": {
				Type:        schema.TypeInt,
				Description: "The total number of units contained in the license key.",
				Computed:    true,
			},
			"used": {
				Type:        schema.TypeInt,
				Description: "The number of units assigned to this license key.",
				Computed:    true,
			},
		},
	}
}

// resourceVSphereLicenseCreate creates a new license using the provided license key and optional labels.
func resourceVSphereLicenseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).vimClient
	manager := license.NewManager(client.Client)

	key := d.Get("license_key").(string)

	helper.MaskedLicenseKeyLogOperation(ctx, "Creating license resource", key, nil)

	var labelMap map[string]interface{}
	if labels, ok := d.GetOk("labels"); ok {
		labelMap = labels.(map[string]interface{})
		tflog.Debug(ctx, "Found labels for license", map[string]interface{}{
			"labelCount": len(labelMap),
		})
	}

	var info types.LicenseManagerLicenseInfo
	var err error
	switch t := client.ServiceContent.About.ApiType; t {
	case "HostAgent":
		if len(labelMap) != 0 {
			tflog.Error(ctx, "Labels are not allowed for unmanaged ESX hosts")
			return diag.FromErr(errors.New("labels are not allowed for unmanaged ESX hosts"))
		}
		tflog.Debug(ctx, "Updating license for ESX host.")
		info, err = manager.Update(ctx, key, nil)

	case "VirtualCenter":
		tflog.Debug(ctx, "Adding license to vCenter instance.")
		info, err = manager.Add(ctx, key, nil)
		if err != nil {
			tflog.Error(ctx, "Failed to add license to vCenter instance", map[string]interface{}{
				"error": err.Error(),
			})
			return diag.FromErr(err)
		}
		tflog.Debug(ctx, "License added successfully, updating labels if any.")
		err = helper.UpdateLabels(ctx, manager, key, labelMap)

	default:
		tflog.Error(ctx, "Unsupported API type", map[string]interface{}{"apiType": t})
		return diag.FromErr(fmt.Errorf("unsupported ApiType: %s", t))
	}

	if err != nil {
		tflog.Error(ctx, "Error creating license", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	if err = helper.DiagnosticError(ctx, info); err != nil {
		tflog.Error(ctx, "License diagnostic error", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(err)
	}

	d.SetId(info.LicenseKey)
	tflog.Debug(ctx, "License created successfully, proceeding to Read")

	return resourceVSphereLicenseRead(ctx, d, meta)
}

// resourceVSphereLicenseRead retrieves license information and populates the resource data with its attributes.
func resourceVSphereLicenseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Client).vimClient
	manager := license.NewManager(client.Client)
	licenseKey := d.Id()

	tflog.Debug(ctx, "Reading license")

	info := helper.GetLicenseInfoFromKey(ctx, licenseKey, manager)
	if info == nil {
		tflog.Warn(ctx, "license not found, removing from state")
		d.SetId("")
		return diags
	}

	tflog.Debug(ctx, "Found license, setting attributes")

	if err := d.Set("labels", helper.KeyValuesToMap(ctx, info.Labels)); err != nil {
		tflog.Error(ctx, "Failed to set 'labels' attribute", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(fmt.Errorf("error setting license labels: %w", err))
	}
	if err := d.Set("license_key", licenseKey); err != nil {
		tflog.Error(ctx, "Failed to set 'license_key' attribute", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(fmt.Errorf("error setting license key attribute: %w", err))
	}
	if err := d.Set("edition_key", info.EditionKey); err != nil {
		tflog.Error(ctx, "Failed to set 'edition_key' attribute", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(fmt.Errorf("error setting license edition_key: %w", err))
	}
	if err := d.Set("name", info.Name); err != nil {
		tflog.Error(ctx, "Failed to set 'name' attribute", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(fmt.Errorf("error setting license name: %w", err))
	}
	if err := d.Set("total", info.Total); err != nil {
		tflog.Error(ctx, "Failed to set 'total' attribute", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(fmt.Errorf("error setting license total: %w", err))
	}
	if err := d.Set("used", info.Used); err != nil {
		tflog.Error(ctx, "Failed to set 'used' attribute", map[string]interface{}{"error": err.Error()})
		return diag.FromErr(fmt.Errorf("error setting license used: %w", err))
	}

	tflog.Debug(ctx, "Successfully finished reading vSphere license")
	return diags
}

// resourceVSphereLicenseUpdate updates an existing license by modifying its labels or attributes as needed.
func resourceVSphereLicenseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).vimClient
	manager := license.NewManager(client.Client)
	resourceID := d.Id()

	helper.MaskedLicenseKeyLogOperation(ctx, "Updating license resource", resourceID, nil)

	if d.HasChange("labels") {
		tflog.Debug(ctx, "Labels have changed, updating.")

		if !helper.KeyExists(ctx, resourceID, manager) {
			tflog.Error(ctx, "vSphere license key specified by resource ID not found during update")
			return diag.Errorf("license key not found on vSphere, cannot update labels")
		}

		labelMap := d.Get("labels").(map[string]interface{})

		err := helper.UpdateLabels(ctx, manager, resourceID, labelMap)
		if err != nil {
			tflog.Error(ctx, "Failed to update license labels", map[string]interface{}{
				"error": err.Error(),
			})
			return diag.FromErr(fmt.Errorf("error updating labels for license resource: %w", err))
		}
		tflog.Debug(ctx, "Successfully updated labels")
	} else {
		tflog.Debug(ctx, "No change detected in labels")
	}

	tflog.Debug(ctx, "Update actions complete, proceeding to Read")

	return resourceVSphereLicenseRead(ctx, d, meta)
}

// resourceVSphereLicenseDelete removes a license key from the license manager, performing validation post-deletion.
func resourceVSphereLicenseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Client).vimClient
	manager := license.NewManager(client.Client)
	licenseKey := d.Id()

	tflog.Debug(ctx, "Deleting vSphere license resource", map[string]interface{}{"resourceId": licenseKey})

	if !helper.KeyExists(ctx, licenseKey, manager) {
		tflog.Warn(ctx, "License key not found during delete operation, assuming already deleted", map[string]interface{}{"resourceId": licenseKey})
		d.SetId("")
		return diags
	}

	tflog.Debug(ctx, "License found, proceeding with removal via API", map[string]interface{}{"resourceId": licenseKey})

	err := manager.Remove(ctx, licenseKey)
	if err != nil {
		tflog.Error(ctx, "Failed to remove license via API", map[string]interface{}{
			"resourceId": licenseKey,
			"error":      err.Error(),
		})
		return diag.FromErr(fmt.Errorf("error removing license %s: %w", licenseKey, err))
	}

	if helper.KeyExists(ctx, licenseKey, manager) {
		tflog.Error(ctx, "License key still exists after deletion attempt", map[string]interface{}{"resourceId": licenseKey})
		return diag.FromErr(ErrKeyNotDeleted)
	}

	d.SetId("")
	tflog.Debug(ctx, "Successfully deleted vSphere license resource", map[string]interface{}{"resourceId": licenseKey})
	return diags
}
