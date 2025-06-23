// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereDatastoreCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore_cluster.datastore_cluster_data", "id",
						"vsphere_datastore_cluster.datastore_cluster", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatastoreCluster_absolutePathNoDatacenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreClusterConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore_cluster.datastore_cluster_data", "id",
						"vsphere_datastore_cluster.datastore_cluster", "id",
					),
				),
			},
		},
	})
}
func TestAccDataSourceVSphereDatastoreCluster_getDatastores(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreClusterGetDatastores(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_datastore_cluster.datastore_cluster_data", "id",
						"vsphere_datastore_cluster.datastore_cluster", "id",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatastoreClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  sdrs_io_load_balance_enabled = false
}

data "vsphere_datastore_cluster" "datastore_cluster_data" {
  name          = vsphere_datastore_cluster.datastore_cluster.name
  datacenter_id = vsphere_datastore_cluster.datastore_cluster.datacenter_id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1()),
	)
}

func testAccDataSourceVSphereDatastoreClusterConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  sdrs_io_load_balance_enabled = false
}

data "vsphere_datastore_cluster" "datastore_cluster_data" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/datastore/${vsphere_datastore_cluster.datastore_cluster.name}"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1()),
	)
}

func testAccDataSourceVSphereDatastoreClusterGetDatastores() string {
	return fmt.Sprintf(`
%s

resource "vsphere_datastore_cluster" "datastore_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  sdrs_io_load_balance_enabled = false
}

data "vsphere_datastore_cluster" "datastore_cluster_data" {
  name          = vsphere_datastore_cluster.datastore_cluster.name
  datacenter_id = vsphere_datastore_cluster.datastore_cluster.datacenter_id
}

output "found_datastores" {
  value = length(data.vsphere_datastore_cluster.datastore_cluster_data.datastores) >= 1 ? "true" : "false"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1()),
	)
}
