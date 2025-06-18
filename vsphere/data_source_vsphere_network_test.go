// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereNetwork_dvsPortgroup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNetworkConfigDVSPortgroup(false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_network.net", "type", "DistributedVirtualPortgroup"),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_network.net", "id",
						"vsphere_distributed_port_group.pg", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereNetwork_withTimeout(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNetworkConfigDVSPortgroup(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_network.net", "type", "DistributedVirtualPortgroup"),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_network.net", "id",
						"vsphere_distributed_port_group.pg", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereNetwork_absolutePathNoDatacenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccSkipIfEsxi(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNetworkConfigDVSPortgroupAbsolute(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_network.net", "type", "DistributedVirtualPortgroup"),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_network.net", "id",
						"vsphere_distributed_port_group.pg", "id",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereNetwork_hostPortgroups(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNetworkConfigHostPortgroup(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_network.vmnet", "type", "Network"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereNetworkConfigDVSPortgroup(withTimeout bool) string {
	additionalConfig := ""
	if withTimeout {
		additionalConfig = `retry_timeout = 180`
	}
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = vsphere_distributed_virtual_switch.dvs.id
}

data "vsphere_network" "net" {
  name                            = vsphere_distributed_port_group.pg.name
  datacenter_id                   = data.vsphere_datacenter.rootdc1.id
  distributed_virtual_switch_uuid = vsphere_distributed_virtual_switch.dvs.id
  %s
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()), additionalConfig)
}

func testAccDataSourceVSphereNetworkConfigDVSPortgroupAbsolute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_distributed_port_group" "pg" {
  name                            = "terraform-test-pg"
  distributed_virtual_switch_uuid = vsphere_distributed_virtual_switch.dvs.id
}

data "vsphere_network" "net" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/network/${vsphere_distributed_port_group.pg.name}"
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccDataSourceVSphereNetworkConfigHostPortgroup() string {
	return fmt.Sprintf(`
%s
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootVMNet()),
	)
}

func testAccDataSourceVSphereNetworkConfigVPCNetwork() string {
	return fmt.Sprintf(`
%s
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataVPCNetwork(), testhelper.ConfigDataRootVMNet()),
	)
}

func TestAccDataSourceVSphereNetwork_vpcNetwork(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			if os.Getenv("TF_VAR_VSPHERE_VPC_SUBNET") == "" || os.Getenv("TF_VAR_VSPHERE_VPC_ID") == "" {
				t.Skipf("This test requires VPC settings, please set TF_VAR_VSPHERE_VPC_SUBNET and TF_VAR_VSPHERE_VPC_ID environment variables")
			}
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNetworkConfigVPCNetwork(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_network.vmnet", "type", "Network"),
				),
			},
		},
	})
}
