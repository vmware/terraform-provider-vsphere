// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

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
				Description: "The CPU usage of the vSphere Namespace in MHz",
				Computed:    true,
			},
			"memory_usage": {
				Type:        schema.TypeInt,
				Description: "The memory usage of the vSphere Namespace in MiB",
				Computed:    true,
			},
			"storage_usage": {
				Type:        schema.TypeInt,
				Description: "The storage usage of the vSphere Namespace in MiB",
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
	data, err := m.GetNamespaceV2(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	_ = d.Set("config_status", string(data.ConfigStatus))
	_ = d.Set("supervisor", data.Supervisor)
	_ = d.Set("cpu_usage", data.Stats.CPUUsed)
	_ = d.Set("memory_usage", data.Stats.MemoryUsed)
	_ = d.Set("storage_usage", data.Stats.StorageUsed)

	vmService := make(map[string]interface{})
	vmService["content_libraries"] = data.VmServiceSpec.ContentLibraries
	vmService["vm_classes"] = data.VmServiceSpec.VmClasses
	_ = d.Set("vm_service", []map[string]interface{}{vmService})

	storagePolicies := make([]string, len(data.StorageSpecs))
	for i, storageSpec := range data.StorageSpecs {
		storagePolicies[i] = storageSpec.Policy
	}
	_ = d.Set("storage_policies", storagePolicies)

	d.SetId(name)

	return nil
}
