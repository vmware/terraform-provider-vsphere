// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/configprofile"
)

func dataSourceVSphereConfigurationProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSphereConfigurationProfileRead,
		Schema: map[string]*schema.Schema{
			"configuration": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration json.",
			},
			"schema": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The configuration schema.",
			},
			"cluster_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The identifier of the cluster.",
			},
		},
	}
}

func dataSourceVSphereConfigurationProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*Client).restClient
	return configprofile.ReadConfigProfile(ctx, client, d)
}
