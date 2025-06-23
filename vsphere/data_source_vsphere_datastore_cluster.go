// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/datastore"
)

func dataSourceVSphereDatastoreCluster() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereDatastoreClusterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name or absolute path to the datastore cluster.",
			},
			"datacenter_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The managed object ID of the datacenter the cluster is located in. Not required if using an absolute path.",
			},
			"datastores": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The names of datastores included in the datastore cluster.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVSphereDatastoreClusterRead(d *schema.ResourceData, meta interface{}) error {
	ctx := context.Background()
	client := meta.(*Client).vimClient
	pod, err := resourceVSphereDatastoreClusterGetPodFromPath(meta, d.Get("name").(string), d.Get("datacenter_id").(string))
	if err != nil {
		return fmt.Errorf("error loading datastore cluster: %s", err)
	}
	d.SetId(pod.Reference().Value)
	var dsNames []string
	childDatastores, err := pod.Children(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving datastores in datastore cluster: %s", err)
	}
	for d := range childDatastores {
		ds, err := datastore.FromID(client, childDatastores[d].Reference().Value)
		if err != nil {
			return fmt.Errorf("error retrieving datastore: %s", err)
		}
		dsNames = append(dsNames, ds.Name())
	}
	err = d.Set("datastores", dsNames)
	if err != nil {
		return fmt.Errorf("cannot set datastores: %s", err)
	}
	return nil
}
