// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereGOSC_basic(t *testing.T) {
	goscName := acctest.RandomWithPrefix("lin")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereGOSCConfig(goscName),
				Check:  resource.TestCheckResourceAttr("data.vsphere_guest_os_customization.gosc1", "id", goscName),
			},
		},
	})
}

func testAccDataSourceVSphereGOSCConfig(goscName string) string {
	return fmt.Sprintf(`
resource "vsphere_guest_os_customization" "source" {
  name = %q
  type = "Linux"
  spec {
    linux_options {
      domain    = "example.com"
      host_name = "linux"
    }
  }
}
data "vsphere_guest_os_customization" "gosc1" {
  name = vsphere_guest_os_customization.source.id
}
`, goscName)
}
