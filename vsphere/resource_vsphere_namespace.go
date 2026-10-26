// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/govmomi/vapi/rest"
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
				ForceNew:    true,
			},
			"supervisor": {
				Type:        schema.TypeString,
				Description: "The Supervisor which the vSphere Namespace belongs to",
				Required:    true,
				ForceNew:    true,
			},
			"vm_service": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_libraries": {
							Type:        schema.TypeSet,
							Description: "The content libraries associated with the vSphere Namespace",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"vm_classes": {
							Type:        schema.TypeSet,
							Description: "The VM classes associated with the vSphere Namespace",
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"storage_policies": {
				Type:        schema.TypeSet,
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
		Namespace:     d.Get("name").(string),
		Supervisor:    d.Get("supervisor").(string),
		StorageSpecs:  getStorageSpecs(d),
		VmServiceSpec: getVMServiceSpec(d),
	}

	if err := m.CreateNamespaceV2(ctx, spec); err != nil {
		return diag.FromErr(err)
	}

	if diags := waitForNamespaceCreation(ctx, m, spec.Namespace); diags != nil {
		log.Printf("[DEBUG] Namespace %q did not reach status 'Running'", spec.Namespace)
		return append(diags, deleteNamespace(ctx, m, spec.Namespace)...)
	}

	return resourceVSphereNamespaceRead(ctx, d, meta)
}

func resourceVSphereNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	name := d.Get("name").(string)
	data, err := m.GetNamespaceV2(ctx, name)
	if err != nil {
		if rest.IsStatusError(err, http.StatusNotFound) {
			log.Printf("[DEBUG] Namespace %s was not found", name)
			d.SetId("")
			return nil
		}
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

	if storageSpecs := getStorageSpecs(d); storageSpecs != nil {
		spec.StorageSpecs = *storageSpecs
	}

	if vmServiceSpec := getVMServiceSpec(d); vmServiceSpec != nil {
		spec.VmServiceSpec = *vmServiceSpec
	}

	return diag.FromErr(m.UpdateNamespace(ctx, name, spec))
}

func resourceVSphereNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return deleteNamespace(ctx, namespace.NewManager(meta.(*Client).restClient), d.Id())
}

func deleteNamespace(ctx context.Context, m *namespace.Manager, name string) diag.Diagnostics {
	if err := m.DeleteNamespace(ctx, name); err != nil {
		return diag.FromErr(err)
	}

	return waitForNamespaceDeletion(ctx, m, name)
}

func waitForNamespaceDeletion(ctx context.Context, m *namespace.Manager, name string) diag.Diagnostics {
	tickerCtx, cancel := context.WithTimeout(ctx, time.Minute*time.Duration(1))
	defer cancel()
	ticker := time.NewTicker(time.Second * time.Duration(5))
	defer ticker.Stop()

	for {
		select {
		case <-tickerCtx.Done():
			return diag.Errorf("timed out waiting for namespace to be deleted: %s", name)
		case <-ticker.C:
			_, err := m.GetNamespaceV2(ctx, name)
			if rest.IsStatusError(err, http.StatusNotFound) {
				return nil
			}
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
}

func waitForNamespaceCreation(ctx context.Context, m *namespace.Manager, name string) diag.Diagnostics {
	tickerCtx, cancel := context.WithTimeout(ctx, time.Minute*time.Duration(5))
	defer cancel()
	ticker := time.NewTicker(time.Second * time.Duration(15))
	defer ticker.Stop()

	for {
		select {
		case <-tickerCtx.Done():
			timeoutDiags := diag.Errorf("timed out waiting for namespace %s to reach RUNNING state", name)
			data, err := m.GetNamespaceV2(ctx, name)
			if err != nil {
				return timeoutDiags
			}
			return append(timeoutDiags, collectNamespaceDiagnostics(data)...)
		case <-ticker.C:
			data, err := m.GetNamespaceV2(ctx, name)
			if err != nil {
				return diag.FromErr(err)
			}

			switch data.ConfigStatus {
			case namespace.NamespaceInstanceConfigStatusRunning:
				return nil
			case namespace.NamespaceInstanceConfigStatusError:
				diags := collectNamespaceDiagnostics(data)
				return append(diag.Errorf("namespace %s entered ERROR state", name), diags...)
			}
		}
	}
}

func getStorageSpecs(d *schema.ResourceData) *[]namespace.StorageSpec {
	storagePolicies := d.Get("storage_policies").(*schema.Set).List()
	if len(storagePolicies) == 0 {
		return nil
	}

	storageSpecs := make([]namespace.StorageSpec, len(storagePolicies))
	for i, storagePolicy := range storagePolicies {
		storageSpecs[i] = namespace.StorageSpec{Policy: storagePolicy.(string)}
	}

	return &storageSpecs
}

func getVMServiceSpec(d *schema.ResourceData) *namespace.VmServiceSpec {
	vmService := d.Get("vm_service").([]interface{})
	if len(vmService) == 0 || vmService[0] == nil {
		return nil
	}

	vmServiceData := vmService[0].(map[string]interface{})
	return &namespace.VmServiceSpec{
		VmClasses:        structure.SliceInterfacesToStrings(vmServiceData["vm_classes"].(*schema.Set).List()),
		ContentLibraries: structure.SliceInterfacesToStrings(vmServiceData["content_libraries"].(*schema.Set).List()),
	}
}

func collectNamespaceDiagnostics(n namespace.NamespaceInstanceInfoV2) diag.Diagnostics {
	var d diag.Diagnostics
	if len(n.Messages) > 0 {
		for _, m := range n.Messages {
			d = append(d, diag.Errorf("%s", m.Details.DefaultMessage)...)
		}
	}

	return d
}
