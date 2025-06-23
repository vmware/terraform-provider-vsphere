// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/datastore"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/storagepod"
)

func dataSourceVSphereDatastoreStats() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereDatastoreStatsRead,

		Schema: map[string]*schema.Schema{
			"datacenter_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the datacenter to get datastores from.",
				Required:    true,
			},
			"free_space": {
				Type:        schema.TypeMap,
				Description: "The free space of the datastores.",
				Optional:    true,
			},
			"capacity": {
				Type:        schema.TypeMap,
				Description: "The capacity of the datastores.",
				Optional:    true,
			},
		},
	}
}

func dataSourceVSphereDatastoreStatsRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Client).vimClient
	var dc *object.Datacenter
	if dcID, ok := d.GetOk("datacenter_id"); ok {
		var err error
		dc, err = datacenterFromID(client, dcID.(string))
		if err != nil {
			return fmt.Errorf("cannot locate datacenter: %s", err)
		}
	}

	dss, err := datastore.List(client)
	if err != nil {
		return fmt.Errorf("error listing datastores: %s", err)
	}

	storagePods, err := storagepod.List(client)
	if err != nil {
		return fmt.Errorf("error retrieving storage pods: %s", err)
	}
	for s := range storagePods {
		childDatastores, err := storagePods[s].Children(ctx)
		if err != nil {
			return fmt.Errorf("error retrieving datastores in datastore cluster: %s", err)
		}
		for c := range childDatastores {
			ds, err := datastore.FromID(client, childDatastores[c].Reference().Value)
			if err != nil {
				log.Printf("[WARN] Skipping datastore with ID %s: %s", childDatastores[c].Reference().Value, err)
				continue
			}
			if ds != nil {
				dss = append(dss, ds)
			}
		}
	}

	processedAny := false

	for i := range dss {
		ds, err := datastore.FromPath(client, dss[i].Name(), dc)
		if err != nil {
			log.Printf("[WARN] Skipping inaccessible datastore %q: %s\n", dss[i].Name(), err)
			continue
		}
		if ds == nil {
			log.Printf("[WARN] Datastore object is nil for %q, skipping\n", dss[i].Name())
			continue
		}

		props, err := datastore.Properties(ds)
		if err != nil {
			log.Printf("[WARN] Skipping datastore %q with inaccessible properties: %s\n", ds.Reference().Value, err)
			continue
		}
		if props == nil {
			log.Printf("[WARN] Properties are nil for datastore %q, skipping\n", ds.Reference().Value)
			continue
		}

		capacityMap := d.Get("capacity").(map[string]interface{})
		capacityMap[dss[i].Name()] = fmt.Sprintf("%v", props.Summary.Capacity)
		_ = d.Set("capacity", capacityMap)

		fr := d.Get("free_space").(map[string]interface{})
		fr[dss[i].Name()] = fmt.Sprintf("%v", props.Summary.FreeSpace)
		_ = d.Set("free_space", fr)

		processedAny = true
	}

	if !processedAny {
		return fmt.Errorf("failed to process any datastores, all were inaccessible")
	}

	d.SetId(fmt.Sprintf("%s_stats", dc.Reference().Value))
	return nil
}
