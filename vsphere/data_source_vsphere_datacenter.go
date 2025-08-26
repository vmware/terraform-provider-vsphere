// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
)

func dataSourceVSphereDatacenter() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereDatacenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name of the datacenter. This can be a name or path. Can be omitted if there is only one datacenter in your inventory.",
				Optional:    true,
			},
			"virtual_machines": {
				Type:        schema.TypeSet,
				Description: "List of all virtual machines included in the vSphere datacenter object.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVSphereDatacenterRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Client).vimClient
	datacenter := d.Get("name").(string)
	dc, err := getDatacenter(client, datacenter)
	if err != nil {
		return fmt.Errorf("error fetching datacenter: %s", err)
	}
	id := dc.Reference().Value
	d.SetId(id)
	finder := find.NewFinder(client.Client)
	finder.SetDatacenter(dc)
	viewManager := view.NewManager(client.Client)
	containerView, err := viewManager.CreateContainerView(ctx, dc.Reference(), []string{"VirtualMachine"}, true)
	if err != nil {
		return fmt.Errorf("error fetching datacenter: %s", err)
	}
	defer func() {
		if err := containerView.Destroy(ctx); err != nil {
			log.Printf("[WARN] Error destroying containerView during cleanup: %v", err)
		}
	}()
	var vms []mo.VirtualMachine
	err = containerView.Retrieve(ctx, []string{"VirtualMachine"}, []string{"name"}, &vms)
	if err != nil {
		return fmt.Errorf("error fetching virtual machines: %s", err)
	}
	var vmNames []string
	for v := range vms {
		vmNames = append(vmNames, vms[v].Name)
	}
	err = d.Set("virtual_machines", vmNames)
	if err != nil {
		return fmt.Errorf("error setting virtual_machines: %s", err)
	}
	return nil
}
