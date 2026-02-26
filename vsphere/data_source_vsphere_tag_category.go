// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereTagCategory() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereTagCategoryRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The display name of the category.",
				Optional:    true,
				ExactlyOneOf: []string{
					"id",
					"name",
				},
			},
			"id": {
				Type:        schema.TypeString,
				Description: "The unique identifier of the category.",
				Optional:    true,
				ExactlyOneOf: []string{
					"id",
					"name",
				},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the category.",
			},
			"cardinality": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The associated cardinality of the category. Can be one of SINGLE (object can only be assigned one tag in this category) or MULTIPLE (object can be assigned multiple tags in this category).",
			},
			"associable_types": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Object types to which this category's tags can be attached.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVSphereTagCategoryRead(d *schema.ResourceData, meta interface{}) error {
	tm, err := meta.(*Client).TagsManager()
	if err != nil {
		return err
	}

	if id, ok := d.GetOk("id"); ok {
		d.SetId(id.(string))
		return resourceVSphereTagCategoryRead(d, meta)
	}
	name := d.Get("name").(string)
	if name == "" {
		return fmt.Errorf("either id or name must be provided")
	}

	id, err := tagCategoryByName(tm, d.Get("name").(string))
	if err != nil {
		return err
	}

	d.SetId(id)
	return resourceVSphereTagCategoryRead(d, meta)
}
