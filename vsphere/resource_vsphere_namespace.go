// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
)

func resourceVSphereNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVSphereNamespaceCreate,
		ReadContext:   resourceVSphereNamespaceRead,
		UpdateContext: resourceVSphereNamespaceUpdate,
		DeleteContext: resourceVSphereNamespaceDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the vSphere Namespace",
				Required:    true,
			},
			"supervisor": {
				Type:        schema.TypeString,
				Description: "The Supervisor which the vSphere Namespace belongs to",
				Required:    true,
			},
			"vm_service": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_libraries": {
							Type:        schema.TypeList,
							Description: "The content libraries associated with the vSphere Namespace",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"vm_classes": {
							Type:        schema.TypeList,
							Description: "The VM classes associated with the vSphere Namespace",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"storage_policies": {
				Type:        schema.TypeList,
				Description: "The storage policies associated with the vSphere Namespace",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceVSphereNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	spec := namespace.NamespaceInstanceCreateSpecV2{
		Namespace:  d.Get("name").(string),
		Supervisor: d.Get("supervisor").(string),
	}

	if storagePolicies := d.Get("storage_policies").([]interface{}); len(storagePolicies) > 0 {
		storageSpecs := make([]namespace.StorageSpec, len(storagePolicies))
		for i, storagePolicy := range storagePolicies {
			storageSpecs[i] = namespace.StorageSpec{Policy: storagePolicy.(string)}
		}

		spec.StorageSpecs = &storageSpecs
	}

	if vmService := d.Get("vm_service").([]interface{}); len(vmService) > 0 {
		vmServiceData := vmService[0].(map[string]interface{})
		spec.VmServiceSpec = &namespace.VmServiceSpec{
			VmClasses:        structure.SliceInterfacesToStrings(vmServiceData["vm_classes"].([]interface{})),
			ContentLibraries: structure.SliceInterfacesToStrings(vmServiceData["content_libraries"].([]interface{})),
		}
	}

	if err := m.CreateNamespaceV2(ctx, spec); err != nil {
		return diag.FromErr(err)
	}

	if err := waitForNamespaceCreation(ctx, d, meta); err != nil {
		return err
	}

	return resourceVSphereNamespaceRead(ctx, d, meta)
}

func resourceVSphereNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	name := d.Get("name").(string)
	data, err := m.GetNamespaceV2(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	var vmServiceAttr []map[string]interface{}
	if len(data.VmServiceSpec.ContentLibraries) > 0 || len(data.VmServiceSpec.VmClasses) > 0 {
		vmService := make(map[string]interface{})
		vmService["content_libraries"] = data.VmServiceSpec.ContentLibraries
		vmService["vm_classes"] = data.VmServiceSpec.VmClasses
		vmServiceAttr = append(vmServiceAttr, vmService)
	}
	_ = d.Set("vm_service", vmServiceAttr)

	storagePolicies := make([]string, len(data.StorageSpecs))
	for i, storageSpec := range data.StorageSpecs {
		storagePolicies[i] = storageSpec.Policy
	}
	_ = d.Set("storage_policies", storagePolicies)

	d.SetId(name)

	return nil
}

func resourceVSphereNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	name := d.Get("name").(string)
	spec := namespace.NamespacesInstanceUpdateSpec{}

	if storagePolicies := d.Get("storage_policies").([]interface{}); len(storagePolicies) > 0 {
		storageSpecs := make([]namespace.StorageSpec, len(storagePolicies))
		for i, storagePolicy := range storagePolicies {
			storageSpecs[i] = namespace.StorageSpec{Policy: storagePolicy.(string)}
		}

		spec.StorageSpecs = storageSpecs
	}

	if vmService := d.Get("vm_service").([]interface{}); len(vmService) > 0 {
		vmServiceData := vmService[0].(map[string]interface{})
		spec.VmServiceSpec = namespace.VmServiceSpec{
			VmClasses:        structure.SliceInterfacesToStrings(vmServiceData["vm_classes"].([]interface{})),
			ContentLibraries: structure.SliceInterfacesToStrings(vmServiceData["content_libraries"].([]interface{})),
		}
	}

	return diag.FromErr(m.UpdateNamespace(ctx, name, spec))
}

func resourceVSphereNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	return diag.FromErr(m.DeleteNamespace(ctx, d.Id()))
}

func waitForNamespaceCreation(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	name := d.Get("name").(string)

	tickerCtx, cancel := context.WithTimeout(ctx, time.Minute*time.Duration(2))
	defer cancel()
	ticker := time.NewTicker(time.Second * time.Duration(15))

	for {
		select {
		case <-tickerCtx.Done():
			return diag.Errorf("failed to create namespace: %s", name)
		case <-ticker.C:
			data, err := m.GetNamespaceV2(ctx, name)
			if err != nil {
				return diag.FromErr(err)
			}

			if data.ConfigStatus == "RUNNING" {
				return nil
			}
		}
	}
}
