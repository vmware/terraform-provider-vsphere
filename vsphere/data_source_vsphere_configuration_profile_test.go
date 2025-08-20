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

func TestAccDataSourceVSphereConfigProfile(t *testing.T) {
	// Run this test manually, do not include in automated testing
	t.Skipf("Skipped due to cleanup problems - https://github.com/vmware/terraform-provider-vsphere/issues/2543")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereConfigProfileConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_config_profile.profile2", "config"),
					resource.TestCheckResourceAttrSet("data.vsphere_config_profile.profile2", "schema"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereConfigProfileConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_host_thumbprint" "thumbprint" {
  address  = "%s"
  insecure = true
}

resource "vsphere_host" "h1" {
  hostname = "%s"
  username = "root"
  password = "%s"
  thumbprint = data.vsphere_host_thumbprint.thumbprint.id

  datacenter = data.vsphere_datacenter.rootdc1.id

  lifecycle {
    ignore_changes = ["services"]
  }
}

resource "vsphere_compute_cluster" "cluster1" {
  name            = "cluster1"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
  host_system_ids = [vsphere_host.h1.id]
}

resource "vsphere_configuration_profile" "profile1" {
  reference_host_id = vsphere_host.h1.id
  cluster_id = vsphere_compute_cluster.cluster1.id
}

data "vsphere_configuration_profile" "profile2" {
  cluster_id = vsphere_config_profile.profile1.cluster_id
}
`,
		testhelper.ConfigDataRootDC1(),
		os.Getenv("TF_VAR_VSPHERE_ESXI4"),
		os.Getenv("TF_VAR_VSPHERE_ESXI4"),
		os.Getenv("TF_VAR_VSPHERE_ESXI4_PASSWORD"))
}
