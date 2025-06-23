// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/contentlibrary"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/provider"
)

func dataSourceVSphereContentLibrary() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereContentLibraryRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the content library.",
			},
		},
	}
}

func dataSourceVSphereContentLibraryRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).restClient
	lib, err := contentlibrary.FromName(c, d.Get("name").(string))
	if err != nil {
		return provider.Error(d.Get("name").(string), "dataSourceVSphereContentLibraryRead", err)
	}
	d.SetId(lib.ID)
	return nil
}
