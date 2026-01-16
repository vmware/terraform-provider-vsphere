// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereContentLibraryItem_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereContentLibraryItemConfig() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = data.vsphere_datastore.rootds1.name
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_test"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "TinyVM"
  library_id  = vsphere_content_library.library.id
  type        = "ova"
  file_url    = var.file
}

data "vsphere_content_library_item" "item" {
  name       = vsphere_content_library_item.item.name
  library_id = vsphere_content_library.library.id
  type       = "ovf"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
		testhelper.TestOva,
	)
}
