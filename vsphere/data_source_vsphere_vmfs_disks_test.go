// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereVmfsDisks_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVmfsDisksConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputBool("found", "true"),
				),
			},
			{
				Config: testAccDataSourceVSphereVmfsDisksInfoConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputBool("found", "true"),
				),
			},
		},
	})
}

// testCheckOutputBool checks an output in the Terraform configuration
func testCheckOutputBool(name string, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Value.(string) != value {
			return fmt.Errorf(
				"output '%s': expected %#v, got %#v",
				name,
				value,
				rs)
		}

		return nil
	}
}

func testAccDataSourceVSphereVmfsDisksConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

data "vsphere_vmfs_disks" "available" {
  host_system_id = data.vsphere_host.esxi_host.id
  rescan         = true
}

output "found" {
  value = length(data.vsphere_vmfs_disks.available.disks) >= 1 ? "true" : "false"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"),
	)
}

func testAccDataSourceVSphereVmfsDisksInfoConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

data "vsphere_vmfs_disks" "available" {
  host_system_id = "${data.vsphere_host.esxi_host.id}"
  rescan         = true
}

output "found" {
  value = "${length(data.vsphere_vmfs_disks.available.disk_details) >= 1 ? "true" : "false" }"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI1"),
	)
}
