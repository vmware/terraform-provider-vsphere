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

func TestAccDataSourceVSphereSupervisor_name(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereSupervisorConfigName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "id"),
					resource.TestCheckResourceAttr("data.vsphere_supervisor.supervisor", "name", os.Getenv("TF_VAR_SUPERVISOR")),
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "config_status"),
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "kubernetes_status"),
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "cpu_usage"),
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "memory_usage"),
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "storage_usage"),
					resource.TestCheckResourceAttrSet("data.vsphere_supervisor.supervisor", "zones.#"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereSupervisor_id(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereSupervisorConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_supervisor.by_id",
						"id",
						"data.vsphere_supervisor.supervisor",
						"id",
					),
					resource.TestCheckResourceAttr("data.vsphere_supervisor.by_id", "name", os.Getenv("TF_VAR_SUPERVISOR")),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereSupervisor_cluster(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereSupervisorConfigCluster(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_supervisor.by_cluster",
						"id",
						"data.vsphere_supervisor.supervisor",
						"id",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereSupervisorConfigName() string {
	return fmt.Sprintf(`
data vsphere_supervisor "supervisor" {
  name = "%s"
}
`, os.Getenv("TF_VAR_SUPERVISOR"))
}

func testAccDataSourceVSphereSupervisorConfigID() string {
	return fmt.Sprintf(`
%s

data vsphere_supervisor "by_id" {
  id = data.vsphere_supervisor.supervisor.id
}
`, testAccDataSourceVSphereSupervisorConfigName())
}

func testAccDataSourceVSphereSupervisorConfigCluster() string {
	return fmt.Sprintf(`
%s

data vsphere_supervisor "by_cluster" {
  cluster = data.vsphere_supervisor.supervisor.zones.0.clusters.0
}
`, testAccDataSourceVSphereSupervisorConfigName())
}
