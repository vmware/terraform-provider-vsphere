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

func TestAccDataSourceVSphereComputeClusterHostGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereComputeClusterHostGroupConfig(2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_compute_cluster_host_group.test", "host_system_ids",
						"vsphere_compute_cluster_host_group.cluster_host_group", "host_system_ids",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereComputeClusterHostGroupConfig(count int) string {
	return fmt.Sprintf(`
variable hosts {
  default = [ %q, %q ]
}

%s

data "vsphere_host" "hosts" {
  count         = %d
  name          = var.hosts[count.index]
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_compute_cluster_host_group" "cluster_host_group" {
  name               = "terraform-test-cluster-group"
  compute_cluster_id = data.vsphere_compute_cluster.rootcompute_cluster1.id
  host_system_ids    = data.vsphere_host.hosts.*.id
}

data "vsphere_compute_cluster_host_group" "test" {
  name               = "terraform-test-cluster-group"
  compute_cluster_id = data.vsphere_compute_cluster.rootcompute_cluster1.id
}
`,
		os.Getenv("TF_VAR_VSPHERE_ESXI1"),
		os.Getenv("TF_VAR_VSPHERE_ESXI2"),
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
		count)
}
