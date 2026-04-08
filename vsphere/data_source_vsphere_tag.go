// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereTag() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereTagRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Description: "The display name of the tag.",
				Optional:    true,
				ExactlyOneOf: []string{
					"id",
					"name",
				},
			},
			"id": {
				Type:        schema.TypeString,
				Description: "The unique identifier of the tag.",
				Optional:    true,
				ExactlyOneOf: []string{
					"id",
					"name",
				},
			},
			"category_id": {
				Type:        schema.TypeString,
				Description: "The unique identifier of the parent category for this tag.",
				Optional:    true,
			},
			"description": {
				Type:        schema.TypeString,
				Description: "The description of the tag.",
				Computed:    true,
			},
		},
	}
}

func dataSourceVSphereTagRead(d *schema.ResourceData, meta interface{}) error {
	tm, err := meta.(*Client).TagsManager()
	if err != nil {
		return err
	}

	if tagID, ok := d.GetOk("id"); ok {
		d.SetId(tagID.(string))
		return resourceVSphereTagRead(d, meta)
	}

	name := d.Get("name").(string)
	categoryID := d.Get("category_id").(string)

	if name == "" {
		return fmt.Errorf("either id or name must be provided")
	}

	if categoryID == "" {
		return fmt.Errorf("category_id must be provided when using name lookup")
	}

	tagID, err := tagByName(tm, name, categoryID)
	if err != nil {
		return err
	}

	d.SetId(tagID)
	return resourceVSphereTagRead(d, meta)
}
