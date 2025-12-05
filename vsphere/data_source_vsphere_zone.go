// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/zone"
)

func dataSourceVSphereZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSphereZoneRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vSphere Zone",
				Required:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the vSphere Zone",
				Computed:    true,
			},
			"cluster_ids": {
				Type:        schema.TypeSet,
				Description: "The identifiers of any clusters associated with the vSphere Zone",
				Computed:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceVSphereZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return zone.VSphereZoneRead(ctx, meta.(*Client).restClient, d)
}
