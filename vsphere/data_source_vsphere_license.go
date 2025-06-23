// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/license"
	helper "github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/license"
)

// dataSourceVSphereLicense defines a data source for retrieving licenses.
func dataSourceVSphereLicense() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSphereLicenseRead,

		Schema: map[string]*schema.Schema{
			"license_key": {
				Type:        schema.TypeString,
				Description: "The license key value.",
				Required:    true,
			},
			"id": {
				Type:        schema.TypeString,
				Description: "The license key ID.",
				Computed:    true,
			},
			"labels": {
				Type:        schema.TypeMap,
				Description: "A map of labels applied to the license key.",
				Computed:    true,
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

// dataSourceVSphereLicenseRead retrieves and populates license details for a license key.
func dataSourceVSphereLicenseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := meta.(*Client).vimClient
	manager := license.NewManager(client.Client)
	licenseKey := d.Get("license_key").(string)

	helper.MaskedLicenseKeyLogOperation(ctx, "Reading license data source", licenseKey, nil)

	info := helper.GetLicenseInfoFromKey(ctx, licenseKey, manager)
	if info == nil {
		return diag.FromErr(ErrKeyNotFound)
	}

	tflog.Debug(ctx, "Setting license attributes in data source")

	if err := d.Set("labels", helper.KeyValuesToMap(ctx, info.Labels)); err != nil {
		tflog.Error(ctx, "Failed to set labels attribute", map[string]interface{}{
			"error": err.Error(),
		})
		return diag.FromErr(fmt.Errorf("error setting license labels: %w", err))
	}
	if err := d.Set("edition_key", info.EditionKey); err != nil {
		tflog.Error(ctx, "Failed to set edition_key attribute", map[string]interface{}{
			"error": err.Error(),
		})
		return diag.FromErr(fmt.Errorf("error setting license edition_key: %w", err))
	}
	if err := d.Set("name", info.Name); err != nil {
		tflog.Error(ctx, "Failed to set name attribute", map[string]interface{}{
			"error": err.Error(),
		})
		return diag.FromErr(fmt.Errorf("error setting license name: %w", err))
	}
	if err := d.Set("total", info.Total); err != nil {
		tflog.Error(ctx, "Failed to set total attribute", map[string]interface{}{
			"error": err.Error(),
		})
		return diag.FromErr(fmt.Errorf("error setting license total: %w", err))
	}
	if err := d.Set("used", info.Used); err != nil {
		tflog.Error(ctx, "Failed to set used attribute", map[string]interface{}{
			"error": err.Error(),
		})
		return diag.FromErr(fmt.Errorf("error setting license used: %w", err))
	}

	// Use the license key as the ID for consistent reference
	d.SetId(licenseKey)
	tflog.Debug(ctx, "Successfully read vSphere license data source")

	return diags
}
