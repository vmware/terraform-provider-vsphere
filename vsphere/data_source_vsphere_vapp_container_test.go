// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereVAppContainer_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: true,
				Config:             testAccDataSourceVSphereVAppContainerConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_vapp_container.container", "id", regexp.MustCompile("^resgroup-")),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVAppContainer_path(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				ExpectNonEmptyPlan: true,
				Config:             testAccDataSourceVSphereVAppContainerPathConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_vapp_container.container", "id", regexp.MustCompile("^resgroup-")),
				),
			},
		},
	})
}

func testAccDataSourceVSphereVAppContainerConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datacenter" "dc" {
  name = data.vsphere_datacenter.rootdc1.name
}

resource "vsphere_vapp_container" "vapp" {
  name                    = "vapp-test"
  parent_resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
}

data "vsphere_vapp_container" "container" {
  name          = vsphere_vapp_container.vapp.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,

		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}

func testAccDataSourceVSphereVAppContainerPathConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datacenter" "dc" {
  name = data.vsphere_datacenter.rootdc1.name
}

resource "vsphere_vapp_container" "vapp" {
  name                    = "vapp-test"
  parent_resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
}

data "vsphere_vapp_container" "container" {
  name          = "/${data.vsphere_datacenter.rootdc1.name}/vm/${vsphere_vapp_container.vapp.name}"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootComputeCluster1()),
	)
}
