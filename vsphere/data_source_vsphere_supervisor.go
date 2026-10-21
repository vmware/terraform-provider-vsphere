// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/supervisor"
)

func dataSourceVSphereSupervisor() *schema.Resource {
	lookupKeys := []string{"id", "name", "cluster"}

	return &schema.Resource{
		ReadContext: dataSourceVSphereSupervisorRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Description:  "The ID of the Supervisor",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: lookupKeys,
			},
			"name": {
				Type:         schema.TypeString,
				Description:  "The name of the Supervisor",
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: lookupKeys,
			},
			"cluster": {
				Type:         schema.TypeString,
				Description:  "The ID of a compute cluster managed by the Supervisor",
				Optional:     true,
				ExactlyOneOf: lookupKeys,
			},
			"config_status": {
				Type:        schema.TypeString,
				Description: "The configuration status of the Supervisor",
				Computed:    true,
			},
			"kubernetes_status": {
				Type:        schema.TypeString,
				Description: "The Kubernetes status of the Supervisor",
				Computed:    true,
			},
			"cpu_usage": {
				Type:        schema.TypeInt,
				Description: "The CPU usage of the Supervisor in megahertz",
				Computed:    true,
			},
			"memory_usage": {
				Type:        schema.TypeInt,
				Description: "The memory usage of the Supervisor in mebibytes",
				Computed:    true,
			},
			"storage_usage": {
				Type:        schema.TypeInt,
				Description: "The storage usage of the Supervisor in mebibytes",
				Computed:    true,
			},
			"zones": {
				Type:        schema.TypeList,
				Description: "The vSphere Zones and the compute clusters of the Supervisor",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"zone": {
							Type:        schema.TypeString,
							Description: "The ID of the vSphere Zone",
							Computed:    true,
						},
						"clusters": {
							Type:        schema.TypeList,
							Description: "The IDs of the compute clusters in the vSphere Zone",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"workload_network": {
				Type:        schema.TypeList,
				Description: "The workload networks of the Supervisor",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:        schema.TypeString,
							Description: "The name of the Supervisor network",
							Computed:    true,
						},
						"network_provider": {
							Type:        schema.TypeString,
							Description: "The network provider of the Supervisor network",
							Computed:    true,
						},
						"portgroup": {
							Type:        schema.TypeString,
							Description: "The ID of the distributed port group backing the Supervisor network",
							Computed:    true,
						},
						"ip_assignment_mode": {
							Type:        schema.TypeString,
							Description: "The IP assignment mode of the Supervisor network",
							Computed:    true,
						},
						"gateway": {
							Type:        schema.TypeString,
							Description: "The gateway of the Supervisor network",
							Computed:    true,
						},
						"subnet_mask": {
							Type:        schema.TypeString,
							Description: "The subnet mask of the Supervisor network",
							Computed:    true,
						},
						"namespaces": {
							Type:        schema.TypeList,
							Description: "The vSphere Namespaces that use the Supervisor network",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereSupervisorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	id, err := resolveSupervisorID(ctx, d, m)
	if err != nil {
		return diag.FromErr(err)
	}

	summary, err := m.GetSupervisorSummary(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	topology, err := m.GetSupervisorTopology(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	var networks []namespace.NetworkInfo
	if clusterID := supervisor.GetClusterID(ctx, topology); clusterID != "" {
		networks, err = m.ListClusterNetworks(ctx, clusterID)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	supervisor.FlattenSupervisorSummary(d, summary)
	supervisor.FlattenSupervisorTopology(d, topology)
	supervisor.FlattenClusterNetworks(d, networks)
	d.SetId(id)

	return nil
}

func resolveSupervisorID(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) (string, error) {
	if id := d.Get("id").(string); id != "" {
		return id, nil
	}

	if name := d.Get("name").(string); name != "" {
		return supervisor.GetSupervisorIDByName(ctx, m, name)
	}

	return supervisor.GetSupervisorIDByClusterID(ctx, m, d.Get("cluster").(string))
}
