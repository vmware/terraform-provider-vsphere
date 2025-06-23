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

func TestAccDataSourceVSphereDistributedVirtualSwitch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.0",
						testhelper.HostNic1,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.1",
						testhelper.HostNic2,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_distributed_virtual_switch.dvs-data", "id",
						"vsphere_distributed_virtual_switch.dvs", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDistributedVirtualSwitch_absolutePathNoDatacenterSpecified(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDistributedVirtualSwitchConfigAbsolute(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.#",
						"2",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.0",
						testhelper.HostNic1,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_distributed_virtual_switch.dvs-data",
						"uplinks.1",
						testhelper.HostNic2,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_distributed_virtual_switch.dvs-data", "id",
						"vsphere_distributed_virtual_switch.dvs", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDistributedVirtualSwitch_CreatePortgroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDistributedVirtualSwitchConfigWithPortgroup(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"active_uplinks.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"standby_uplinks.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"active_uplinks.0",
						testhelper.HostNic1,
					),
					resource.TestCheckResourceAttr(
						"vsphere_distributed_port_group.pg",
						"standby_uplinks.0",
						testhelper.HostNic2,
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDistributedVirtualSwitchConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  uplinks       = ["%s", "%s"]
}

data "vsphere_distributed_virtual_switch" "dvs-data" {
  name          = vsphere_distributed_virtual_switch.dvs.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testhelper.HostNic1,
		testhelper.HostNic2,
	)
}

func testAccDataSourceVSphereDistributedVirtualSwitchConfigWithPortgroup() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  uplinks       = ["%s", "%s"]
}

data "vsphere_distributed_virtual_switch" "dvs-data" {
  name          = vsphere_distributed_virtual_switch.dvs.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = data.vsphere_distributed_virtual_switch.dvs-data.id

  active_uplinks  = [data.vsphere_distributed_virtual_switch.dvs-data.uplinks[0]]
  standby_uplinks = [data.vsphere_distributed_virtual_switch.dvs-data.uplinks[1]]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testhelper.HostNic1,
		testhelper.HostNic2,
	)
}

func testAccDataSourceVSphereDistributedVirtualSwitchConfigAbsolute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  uplinks       = ["%s", "%s"]
}

data "vsphere_distributed_virtual_switch" "dvs-data" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/network/${vsphere_distributed_virtual_switch.dvs.name}"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		testhelper.HostNic1,
		testhelper.HostNic2,
	)
}
