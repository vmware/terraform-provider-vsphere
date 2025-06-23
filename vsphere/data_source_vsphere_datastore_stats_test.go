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

func TestAccDataSourceVSphereDatastoreStats_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereDatastoreStatsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckOutputBool("found_free_space", "true"),
					testCheckOutputBool("found_capacity", "true"),
					testCheckOutputBool("free_values_exist", "true"),
					testCheckOutputBool("capacity_values_exist", "true"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereDatastoreStatsConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_datastore_stats" "datastore_stats" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

output "found_free_space" {
  value = length(data.vsphere_datastore_stats.datastore_stats.free_space) >= 1 ? "true" : "false"
}

output "found_capacity" {
  value = length(data.vsphere_datastore_stats.datastore_stats.capacity) >= 1 ? "true" : "false"
}

output "free_values_exist" {
  value = alltrue([
    for free in values(data.vsphere_datastore_stats.datastore_stats.free_space) :
    free >= 1
  ])
}

output "capacity_values_exist" {
  value = alltrue([
    for free in values(data.vsphere_datastore_stats.datastore_stats.capacity) :
    free >= 1
  ])
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1()))
}
