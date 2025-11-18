// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereZone_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereZoneConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("data.vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("data.vsphere_zone.zone1", "description", testZoneDescription),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereZone_associations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereZoneConfigAssociations(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("data.vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("data.vsphere_zone.zone1", "description", testZoneDescription),
					testCheckOutputBool("association_created", "true"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereZoneConfigBasic() string {
	return fmt.Sprintf(`
locals {
  zoneName        = "%s"
  zoneDescription = "%s"
}

resource "vsphere_zone" "zone1" {
  name          = local.zoneName
  description   = local.zoneDescription
}

data "vsphere_zone" "zone1" {
  name          = vsphere_zone.zone1.name
}
`,
		testZoneName,
		testZoneDescription,
	)
}

func testAccDataSourceVSphereZoneConfigAssociations() string {
	return fmt.Sprintf(`
%s

locals {
  zoneName        = "%s"
  zoneDescription = "%s"
}

resource "vsphere_compute_cluster" cluster1 {
  name            = "cluster-1"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_zone" "zone1" {
  name          = local.zoneName
  description   = local.zoneDescription
  cluster_ids = [vsphere_compute_cluster.cluster1.id]
}

data "vsphere_zone" "zone1" {
  name          = vsphere_zone.zone1.name
}

output "association_created" {
  value = contains(vsphere_zone.zone1.cluster_ids, vsphere_compute_cluster.cluster1.id)
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
		),
		testZoneName,
		testZoneDescription,
	)
}
