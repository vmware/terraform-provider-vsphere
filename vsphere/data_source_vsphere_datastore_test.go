// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereDatastore_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore.datastore_data", "id",
						"data.vsphere_datastore.rootds1", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatastore_noDatacenterAndAbsolutePath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore.datastore_data", "id",
						"data.vsphere_datastore.rootds1", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatastore_getStats(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreConfigGetStats(),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputBool("found_stats", "true"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatastoreConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datastore" "datastore_data" {
  name          = data.vsphere_datastore.rootds1.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
	)
}

func testAccDataSourceVSphereDatastoreConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

data "vsphere_datastore" "datastore_data" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/datastore/${data.vsphere_datastore.rootds1.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
	)
}

func testAccDataSourceVSphereDatastoreConfigGetStats() string {
	return fmt.Sprintf(`
%s

data "vsphere_datastore" "datastore_data" {
  name          = data.vsphere_datastore.rootds1.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

output "found_stats" {
  value = length(data.vsphere_datastore.datastore_data.stats) >= 1 ? "true" : "false"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
	)
}
