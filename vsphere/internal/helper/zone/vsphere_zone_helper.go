// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package zone

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter/consumptiondomains/associations"
	"github.com/vmware/govmomi/vapi/vcenter/consumptiondomains/zones"
)

func DataSourceVSphereZoneRead(ctx context.Context, c *rest.Client, d *schema.ResourceData) diag.Diagnostics {
	zm := zones.NewManager(c)

	zoneName := d.Get("name").(string)

	d.SetId(zoneName)

	tflog.Debug(ctx, fmt.Sprintf("reading configuration for vsphere zone: %s", zoneName))
	zone, err := zm.GetZone(zoneName)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("description", zone.Description)

	am := associations.NewManager(c)

	tflog.Debug(ctx, fmt.Sprintf("reading associations for vsphere zone: %s", zoneName))
	clusterIDs, err := am.GetAssociations(zoneName)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("cluster_ids", clusterIDs)

	return nil
}
