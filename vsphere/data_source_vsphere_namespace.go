// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
	namespacehelper "github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/namespace"
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
			"storage_policy": {
				Type:        schema.TypeList,
				Description: "A storage policy assigned to the vSphere Namespace, with its storage limit",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"policy": {
							Type:        schema.TypeString,
							Description: "The identifier of the storage policy",
							Computed:    true,
						},
						"limit": {
							Type:        schema.TypeInt,
							Description: "The maximum storage, in mebibytes (MiB), the vSphere Namespace can use for this policy. 0 means no limit",
							Computed:    true,
						},
					},
				},
			},
			"access": {
				Type:        schema.TypeSet,
				Description: "A principal with access to the vSphere Namespace",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"subject_type": {
							Type:        schema.TypeString,
							Description: "The type of the subject",
							Computed:    true,
						},
						"subject": {
							Type:        schema.TypeString,
							Description: "The name of the user or group, without the domain",
							Computed:    true,
						},
						"domain": {
							Type:        schema.TypeString,
							Description: "The domain of the identity source that owns the subject",
							Computed:    true,
						},
						"role": {
							Type:        schema.TypeString,
							Description: "The role of the subject in the vSphere Namespace",
							Computed:    true,
						},
						"identity_provider": {
							Type:        schema.TypeString,
							Description: "The identifier of the identity provider",
							Computed:    true,
						},
					},
				},
			},
			"zone": {
				Type:        schema.TypeSet,
				Description: "A vSphere Zone assigned to the vSphere Namespace",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Description: "The identifier of the vSphere Zone",
							Computed:    true,
						},
						"type": {
							Type:        schema.TypeString,
							Description: "The type of the vSphere Zone",
							Computed:    true,
						},
					},
				},
			},
			"vpc_network": {
				Type:        schema.TypeList,
				Description: "The NSX VPC network configuration",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc": {
							Type:        schema.TypeString,
							Description: "The NSX path of the VPC",
							Computed:    true,
						},
						"shared_subnet": {
							Type:        schema.TypeSet,
							Description: "An NSX subnet created outside the Supervisor that is associated with the vSphere Namespace",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"path": {
										Type:        schema.TypeString,
										Description: "The NSX path of the subnet",
										Computed:    true,
									},
									"pod_default": {
										Type:        schema.TypeBool,
										Description: "Whether the subnet is the default network for vSphere Pods",
										Computed:    true,
									},
									"vm_default": {
										Type:        schema.TypeBool,
										Description: "Whether the subnet is the default network for VMs",
										Computed:    true,
									},
								},
							},
						},
					},
				},
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

	_ = d.Set("storage_policy", namespacehelper.FlattenStorageSpecs(data.StorageSpecs))
	_ = d.Set("access", namespacehelper.FlattenAccessList(data.AccessList))
	_ = d.Set("zone", namespacehelper.FlattenZoneInfos(data.Zones))
	_ = d.Set("vpc_network", namespacehelper.FlattenVpcNetwork(data.NetworkSpec))

	d.SetId(name)

	return nil
}
