// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereHostThumbprint_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostThumbprintConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_host_thumbprint.thumb", "id", regexp.MustCompile("([0-9A-F]{2}:)+[0-9A-F]{2}")),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostThumbprintConfig() string {
	return fmt.Sprintf(`
data "vsphere_host_thumbprint" "thumb" {
  address  = "%s"
  insecure = true
}`,
		os.Getenv("TF_VAR_VSPHERE_ESXI4"),
	)
}
