// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
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
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccDataSourceVSphereDatastorePreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore.datastore_data", "id",
						"vsphere_nas_datastore.ds1", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatastore_getStats(t *testing.T) {
	testAccSkipUnstable(t)
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
					resource.TestCheckResourceAttr(
						"data.vsphere_datastore.datastore_data", "name", os.Getenv("VSPHERE_DATASTORE_NAME"),
					),

					testCheckOutputBool("found_stats", "true"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatastorePreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_nas_datastore acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_nas_datastore acceptance tests")
	}
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
  name = "/${data.vsphere_datacenter.rootdc1.name}/datastore/${vsphere_nas_datastore.ds1.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigResDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccDataSourceVSphereDatastoreConfigGetStats() string {
	return fmt.Sprintf(`
variable "datastore_name" {
  default = "%s"
}
variable "datacenter_id" {
  default = "%s"
}
data "vsphere_datastore" "datastore_data" {
  name          = var.datastore_name
  datacenter_id = var.datacenter_id
}

output "found_stats" {
  value = length(data.vsphere_datastore.datastore_data.stats) >= 1 ? "true" : "false"
}
`, os.Getenv("VSPHERE_DATASTORE_NAME"), os.Getenv("VSPHERE_DATACENTER"),
	)
}
