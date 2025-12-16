// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
)

func dataSourceVSphereNamespace() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVSphereNamespaceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vSphere Namespace",
				Required:    true,
			},
			"config_status": {
				Type:        schema.TypeString,
				Description: "The configuration status of the vSphere Namespace",
				Computed:    true,
			},
			"supervisor": {
				Type:        schema.TypeString,
				Description: "The Supervisor which the vSphere Namespace belongs to",
				Computed:    true,
			},
			"cpu_usage": {
				Type:        schema.TypeInt,
				Description: "The CPU usage of the vSphere Namespace in megabytes",
				Computed:    true,
			},
			"memory_usage": {
				Type:        schema.TypeInt,
				Description: "The memory usage of the vSphere Namespace in megabytes",
				Computed:    true,
			},
			"storage_usage": {
				Type:        schema.TypeInt,
				Description: "The storage usage of the vSphere Namespace in megabytes",
				Computed:    true,
			},
			"vm_service": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_libraries": {
							Type:        schema.TypeList,
							Description: "The content libraries associated with the vSphere Namespace",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"vm_classes": {
							Type:        schema.TypeList,
							Description: "The VM classes associated with the vSphere Namespace",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"storage_policies": {
				Type:        schema.TypeList,
				Description: "The storage policies associated with the vSphere Namespace",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVSphereNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	name := d.Get("name").(string)
	data, err := m.GetNamespace(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("config_status", data.ConfigStatus)

	supervisorID, err := getSupervisorID(ctx, m, data.ClusterId)
	if err != nil {
		return diag.FromErr(err)
	}
	_ = d.Set("supervisor", supervisorID)

	flattenNamespaceInfo(d, data)

	d.SetId(name)

	return nil
}

func flattenNamespaceInfo(d *schema.ResourceData, info namespace.NamespacesInstanceInfo) {
	_ = d.Set("cpu_usage", info.Stats.CpuUsed)
	_ = d.Set("memory_usage", info.Stats.MemoryUsed)
	_ = d.Set("storage_usage", info.Stats.StorageUsed)

	vmService := make(map[string]interface{})
	vmService["content_libraries"] = info.VmServiceSpec.ContentLibraries
	vmService["vm_classes"] = info.VmServiceSpec.VmClasses
	_ = d.Set("vm_service", []map[string]interface{}{vmService})

	storagePolicies := make([]string, len(info.StorageSpecs))
	for i, storageSpec := range info.StorageSpecs {
		storagePolicies[i] = storageSpec.Policy
	}
	_ = d.Set("storage_policies", storagePolicies)
}

func getSupervisorID(ctx context.Context, m *namespace.Manager, clusterID string) (string, error) {
	summaries, err := m.GetSupervisorSummaries(ctx)
	if err != nil {
		return "", err
	}

	for _, s := range summaries.Items {
		topology, err := m.GetSupervisorTopology(ctx, s.Supervisor)
		if err != nil {
			return "", err
		}

		if isClusterInSupervisor(topology, clusterID) {
			return s.Supervisor, nil
		}
	}

	return "", fmt.Errorf("could not find supervisor for cluster %s", clusterID)
}

func isClusterInSupervisor(topologies []namespace.SupervisorTopologyInfo, clusterID string) bool {
	for _, topology := range topologies {
		for _, cluster := range topology.Clusters {
			if cluster == clusterID {
				return true
			}
		}
	}

	return false
}
