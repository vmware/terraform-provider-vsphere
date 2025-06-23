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
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereHost_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host.host",
						"id",
						testAccDataSourceVSphereHostExpectedRegexp(),
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereHost_defaultHost(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccSkipIfNotEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereHostConfigDefault(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_host.host",
						"id",
						testAccDataSourceVSphereHostExpectedRegexp(),
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereHostExpectedRegexp() *regexp.Regexp {
	return regexp.MustCompile("^host-")
}

func testAccDataSourceVSphereHostConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "host" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`, testhelper.ConfigDataRootDC1(), os.Getenv("TF_VAR_VSPHERE_ESXI1"))
}

func testAccDataSourceVSphereHostConfigDefault() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "host" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}`, testhelper.ConfigDataRootDC1())
}
