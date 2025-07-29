// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceVSphereConfigProfile(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereConfigProfileConfig(),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func testAccResourceVSphereConfigProfileConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_config_profile" "profile" {
  #reference_host_id = data.vsphere_host.roothost2.id
  #reference_host_id = "host-10"
  #cluster_id = data.vsphere_compute_cluster.rootcompute_cluster1.id
  cluster_id = "domain-c85"
  config = file("~/git/terraform-provider-vsphere/config.json")
}
`,
		"")
	//testhelper.CombineConfigs(
	//	testhelper.ConfigDataRootDC1(),
	//	testhelper.ConfigDataRootComputeCluster1(),
	//	testhelper.ConfigDataRootHost2()))
}
