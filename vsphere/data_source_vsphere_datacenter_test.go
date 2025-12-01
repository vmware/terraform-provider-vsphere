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
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

var testAccDataSourceVSphereDatacenterExpectedRegexp = regexp.MustCompile("^datacenter-")

func TestAccDataSourceVSphereDatacenter_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatacenterConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_datacenter.dc",
						"id",
						testAccDataSourceVSphereDatacenterExpectedRegexp,
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatacenter_defaultDatacenter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatacenterConfigDefault,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_datacenter.dc",
						"id",
						testAccDataSourceVSphereDatacenterExpectedRegexp,
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereDatacenter_getVirtualMachines(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatacenterConfigGetVirtualMachines(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_datacenter.dc",
						"id",
						testAccDataSourceVSphereDatacenterExpectedRegexp,
					),
					testCheckOutputBool("found_virtual_machines", "true"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatacenterConfig() string {
	return fmt.Sprintf(`
data "vsphere_datacenter" "dc" {
  name = "%s"
}
`, os.Getenv("TF_VAR_VSPHERE_DATACENTER"))
}

const testAccDataSourceVSphereDatacenterConfigDefault = `
data "vsphere_datacenter" "dc" {}
`

func testAccDataSourceVSphereDatacenterConfigGetVirtualMachines() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "srcvm" {
  name             = "acc-test-vm"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
  datastore_id     = data.vsphere_datastore.rootds1.id
  num_cpus         = 1
  memory           = 1024
  guest_id         = "otherLinux64Guest"
  network_interface {
    network_id = data.vsphere_network.network1.id
  }
  disk {
    label = "disk0"
    size  = 1
    io_reservation = 1
  }
  wait_for_guest_ip_timeout  = 0
  wait_for_guest_net_timeout = 0
}

data "vsphere_datacenter" "dc" {
  name = "%s"
  depends_on = [vsphere_virtual_machine.srcvm]
}

output "found_virtual_machines" {
  value = length(data.vsphere_datacenter.dc.virtual_machines) >= 1 ? "true" : "false"
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_DATACENTER"),
	)
}
