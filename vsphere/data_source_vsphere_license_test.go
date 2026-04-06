// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereLicense_basic(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereLicensePreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereLicenseConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_license.license",
						"id",
						os.Getenv("TF_VAR_VSPHERE_LICENSE"),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereLicensePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_LICENSE") == "" {
		t.Skip("set TF_VAR_VSPHERE_LICENSE to run vsphere_license acceptance tests")
	}
}

func testAccDataSourceVSphereLicenseConfig() string {
	return fmt.Sprintf(`
data "vsphere_license" "license" {
  license_key = "%s"
}
`, os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}
