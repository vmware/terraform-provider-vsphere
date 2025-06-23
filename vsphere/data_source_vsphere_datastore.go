// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/datastore"
)

func dataSourceVSphereDatastore() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereDatastoreRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The name or path of the datastore.",
				Required:    true,
			},
			"datacenter_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the datacenter the datastore is in. This is not required when using ESXi directly, or if there is only one datacenter in your infrastructure.",
				Optional:    true,
			},
			"stats": {
				Type:        schema.TypeMap,
				Description: "The usage stats of the datastore, include total capacity and free space in bytes.",
				Optional:    true,
			},
		},
	}
}

func dataSourceVSphereDatastoreRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	name := d.Get("name").(string)
	var dc *object.Datacenter
	if dcID, ok := d.GetOk("datacenter_id"); ok {
		var err error
		dc, err = datacenterFromID(client, dcID.(string))
		if err != nil {
			return fmt.Errorf("cannot locate datacenter: %s", err)
		}
	}
	ds, err := datastore.FromPath(client, name, dc)
	if err != nil {
		return fmt.Errorf("error fetching datastore: %s", err)
	}

	d.SetId(ds.Reference().Value)
	props, err := datastore.Properties(ds)
	if err != nil {
		return fmt.Errorf("error getting properties for datastore ID %q: %s", ds.Reference().Value, err)
	}
	_ = d.Set("stats", map[string]string{"capacity": fmt.Sprintf("%v", props.Summary.Capacity), "free": fmt.Sprintf("%v", props.Summary.FreeSpace)})
	return nil
}
