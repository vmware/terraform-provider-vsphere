// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package namespace

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
)

func ExpandStorageSpecs(d *schema.ResourceData) []namespace.StorageSpec {
	storage := d.Get("storage_policy").(*schema.Set).List()
	log.Printf("[DEBUG] namespace.ExpandStorageSpecs: Expanding %d storage policies", len(storage))
	specs := make([]namespace.StorageSpec, 0, len(storage))
	for _, item := range storage {
		entry := item.(map[string]interface{})
		spec := namespace.StorageSpec{Policy: entry["policy"].(string)}
		if limit := int64(entry["limit"].(int)); limit != 0 {
			spec.Limit = &limit
		}
		specs = append(specs, spec)
	}
	return specs
}

func FlattenStorageSpecs(specs []namespace.StorageSpec) []interface{} {
	log.Printf("[DEBUG] namespace.FlattenStorageSpecs: Flattening %d storage policies", len(specs))
	result := make([]interface{}, 0, len(specs))
	for _, spec := range specs {
		limit := 0
		if spec.Limit != nil {
			limit = int(*spec.Limit)
		}
		result = append(result, map[string]interface{}{
			"policy": spec.Policy,
			"limit":  limit,
		})
	}
	return result
}
