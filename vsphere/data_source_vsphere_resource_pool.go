// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/resourcepool"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
)

// dataSourceVSphereResourcePool defines a data source for retrieving information about a vSphere resource pool.
func dataSourceVSphereResourcePool() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereResourcePoolRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name or path of the resource pool. When used with parent_resource_pool_id, this must be a simple name of a child resource pool, not a path.",
				Optional:    true,
			},
			"datacenter_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the datacenter the resource pool is in. This is not required when using ESXi directly, or if there is only one datacenter in your infrastructure.",
				Optional:    true,
			},
			"parent_resource_pool_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the parent resource pool.",
				Optional:    true,
			},
		},
	}
}

// dataSourceVSphereResourcePoolRead reads details of a vSphere resource pool based on its name and parent identifiers.
func dataSourceVSphereResourcePoolRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	finder := find.NewFinder(client.Client, true)

	name, nameOk := d.Get("name").(string)
	dcID, dcOk := d.GetOk("datacenter_id")
	parentID, parentOk := d.GetOk("parent_resource_pool_id")

	var dc *object.Datacenter
	if dcOk {
		var err error
		dc, err = datacenterFromID(client, dcID.(string))
		if err != nil {
			return fmt.Errorf("cannot locate datacenter %q: %s", dcID.(string), err)
		}
		finder.SetDatacenter(dc)
	} else {
		if err := viapi.ValidateVirtualCenter(client); err == nil {
			defaultDc, err := finder.DefaultDatacenter(context.TODO())
			if err != nil {
				return fmt.Errorf("failed to get default datacenter: %w", err)
			}
			finder.SetDatacenter(defaultDc)
		}
	}

	var rp *object.ResourcePool
	var err error

	if parentOk {
		if !nameOk || name == "" {
			return fmt.Errorf("argument 'name' is required when 'parent_resource_pool_id' is specified")
		}

		var err error
		rp, err = resourcepool.FromParentAndName(client, parentID.(string), name)
		if err != nil {
			return err
		}

	} else {
		if !nameOk || name == "" {
			return fmt.Errorf("argument 'name' is required when 'parent_resource_pool_id' is not specified")
		}

		rp, err = resourcepool.FromPathOrDefault(client, name, dc)
		if err != nil {
			return fmt.Errorf("error fetching resource pool by path %q: %s", name, err)
		}
	}

	d.SetId(rp.Reference().Value)

	return nil
}
