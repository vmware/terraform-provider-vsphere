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
	"github.com/vmware/govmomi/vapi/vcenter/consumptiondomains/associations"
	"github.com/vmware/govmomi/vapi/vcenter/consumptiondomains/zones"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/zone"
)

func resourceVSphereZone() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSphereZoneCreate,
		ReadContext:   resourceVSphereZoneRead,
		UpdateContext: resourceVSphereZoneUpdate,
		DeleteContext: resourceVSphereZoneDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vSphere Zone",
				Required:    true,
				ForceNew:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the vSphere Zone",
				Optional:    true,
				ForceNew:    true,
			},
			"cluster_ids": {
				Type:        schema.TypeSet,
				Description: "The identifiers of any clusters associated with the vSphere Zone",
				Optional:    true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVSphereZoneCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	zm := zones.NewManager(c)

	zoneName := d.Get("name").(string)
	tflog.Debug(ctx, fmt.Sprintf("creating a specification for vSphere Zone %s", zoneName))
	spec := zones.CreateSpec{
		Zone:        zoneName,
		Description: d.Get("description").(string),
	}

	tflog.Debug(ctx, fmt.Sprintf("creating vSphere Zone %s", zoneName))
	if _, err := zm.CreateZone(spec); err != nil {
		return diag.FromErr(err)
	}

	am := associations.NewManager(c)

	for _, clusterID := range d.Get("cluster_ids").(*schema.Set).List() {
		tflog.Debug(ctx, fmt.Sprintf("creating an association for vSphere Zone %s and compute cluster %s", zoneName, clusterID))
		if err := am.AddAssociations(zoneName, clusterID.(string)); err != nil {
			// don't leave partially created zone
			// destroy if association fails
			_ = zm.DeleteZone(zoneName)
			return diag.FromErr(err)
		}
	}

	return resourceVSphereZoneRead(ctx, d, meta)
}

func resourceVSphereZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return zone.VSphereZoneRead(ctx, meta.(*Client).restClient, d)
}

func resourceVSphereZoneUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	am := associations.NewManager(c)

	oldRaw, newRaw := d.GetChange("cluster_ids")
	oldSet := oldRaw.(*schema.Set)
	newSet := newRaw.(*schema.Set)

	intersection := oldSet.Intersection(newSet)
	toRemove := oldSet.Difference(intersection)
	toAdd := newSet.Difference(intersection)

	zoneName := d.Get("name").(string)
	for _, clusterID := range toRemove.List() {
		tflog.Debug(ctx, fmt.Sprintf("removing association of vSphere Zone %s with cluster %s", zoneName, clusterID))
		if err := am.RemoveAssociations(zoneName, clusterID.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	for _, clusterID := range toAdd.List() {
		tflog.Debug(ctx, fmt.Sprintf("creating association of vSphere Zone %s with cluster %s", zoneName, clusterID))
		if err := am.AddAssociations(zoneName, clusterID.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceVSphereZoneRead(ctx, d, meta)
}

func resourceVSphereZoneDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	zm := zones.NewManager(c)

	name := d.Get("name").(string)
	tflog.Debug(ctx, fmt.Sprintf("deleting vSphere Zone %s", name))
	if err := zm.DeleteZone(name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
