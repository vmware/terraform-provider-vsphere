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

func TestAccDataSourceVSphereComputeCluster_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereComputeClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "id",
						"data.vsphere_compute_cluster.rootcompute_cluster1", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "resource_pool_id",
						"data.vsphere_compute_cluster.rootcompute_cluster1", "resource_pool_id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereComputeCluster_absolutePathNoDatacenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereComputeClusterConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "id",
						"vsphere_compute_cluster.compute_cluster", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster.compute_cluster_data", "resource_pool_id",
						"vsphere_compute_cluster.compute_cluster", "resource_pool_id",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereComputeClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

data "vsphere_compute_cluster" "compute_cluster_data" {
  name          = "%s"
  datacenter_id = data.vsphere_compute_cluster.rootcompute_cluster1.datacenter_id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
		os.Getenv("TF_VAR_VSPHERE_CLUSTER"),
	)
}

func testAccDataSourceVSphereComputeClusterConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-datastore-cluster"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

data "vsphere_compute_cluster" "compute_cluster_data" {
  name          = "/${data.vsphere_datacenter.rootdc1.name}/host/${vsphere_compute_cluster.compute_cluster.name}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}
