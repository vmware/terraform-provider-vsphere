// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

const (
	testZoneName               = "zone-1"
	testZoneDescription        = "description-1"
	testZoneNameUpdated        = "zone-2"
	testZoneDescriptionUpdated = "description-2"
	testClusterName1           = "cluster-1"
	testClusterName2           = "cluster-2"
)

func TestAccResourceVSphereZone_createBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereZoneConfigCreateBasic(),
				Check: resource.ComposeTestCheckFunc(
					checkZoneExists("zone1", true),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "description", testZoneDescription),
				),
			},
		},
		CheckDestroy: checkZoneExists("zone1", false),
	})
}

func TestAccResourceVSphereZone_createAssociations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereZoneConfigCreateAssociations(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "description", testZoneDescription),
					testCheckOutputBool("association_created", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereZone_updateBasic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// initial create
				Config: testAccResourceVSphereZoneConfigCreateBasic(),
			},
			{
				// update name, keep description
				Config: testAccResourceVSphereZoneConfigUpdateBasic(testZoneNameUpdated, testZoneDescription),
			},
			{
				// update description, keep name
				Config: testAccResourceVSphereZoneConfigUpdateBasic(testZoneNameUpdated, testZoneDescriptionUpdated),
			},
			{
				// update both name and description
				Config: testAccResourceVSphereZoneConfigUpdateBasic(testZoneName, testZoneDescription),
			},
		},
	})
}

func TestAccResourceVSphereZone_updateAddAssociation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// initial create
				Config: testAccDataSourceVSphereZoneConfigBasic(),
			},
			{
				// create cluster and add association
				Config: testAccResourceVSphereZoneConfigCreateAssociations(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "description", testZoneDescription),
					testCheckOutputBool("association_created", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereZone_updateRemoveAssociation(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// initial create
				Config: testAccResourceVSphereZoneConfigCreateAssociations(),
			},
			{
				// remove association
				Config: testAccDataSourceVSphereZoneConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "description", testZoneDescription),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "cluster_ids.#", "0"),
				),
			},
		},
	})
}

func TestAccResourceVSphereZone_updateAssociations(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				// initial create
				Config: testAccResourceVSphereZoneConfigUpdateAssociationsStep1(),
			},
			{
				// update the configuration to perform an addition and a removal of an association
				Config: testAccResourceVSphereZoneConfigUpdateAssociationsStep2(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "id", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "name", testZoneName),
					resource.TestCheckResourceAttr("vsphere_zone.zone1", "description", testZoneDescription),
					testCheckOutputBool("association_created", "true"),
				),
			},
		},
	})
}

func testAccResourceVSphereZoneConfigCreateBasic() string {
	return fmt.Sprintf(`
resource "vsphere_zone" "zone1" {
  name          = "%s"
  description   = "%s"
}
`,
		testZoneName,
		testZoneDescription,
	)
}

func testAccResourceVSphereZoneConfigCreateAssociations() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" cluster1 {
  name            = "cluster-1"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_zone" "zone1" {
  name          = "%s"
  description   = "%s"
  cluster_ids = [vsphere_compute_cluster.cluster1.id]
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

func testAccResourceVSphereZoneConfigUpdateBasic(name, description string) string {
	return fmt.Sprintf(`
resource "vsphere_zone" "zone1" {
  name          = "%s"
  description   = "%s"
}
`,
		name,
		description,
	)
}

func testAccResourceVSphereZoneConfigUpdateAssociationsStep1() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" cluster1 {
  name            = "%s"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_compute_cluster" cluster2 {
  name            = "%s"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_zone" "zone1" {
  name          = "%s"
  description   = "%s"
  cluster_ids = [vsphere_compute_cluster.cluster1.id]
}

output "association_created" {
  value = contains(vsphere_zone.zone1.cluster_ids, vsphere_compute_cluster.cluster1.id)
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
		),
		testClusterName1,
		testClusterName2,
		testZoneName,
		testZoneDescription,
	)
}

func testAccResourceVSphereZoneConfigUpdateAssociationsStep2() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" cluster1 {
  name            = "%s"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_compute_cluster" cluster2 {
  name            = "%s"
  datacenter_id   = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_zone" "zone1" {
  name          = "%s"
  description   = "%s"
  cluster_ids = [vsphere_compute_cluster.cluster2.id]
}

output "association_created" {
  value = contains(vsphere_zone.zone1.cluster_ids, vsphere_compute_cluster.cluster2.id)
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
		),
		testClusterName1,
		testClusterName2,
		testZoneName,
		testZoneDescription,
	)
}

func checkZoneExists(zone string, shouldExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name, err := testGetVSphereZone(s, zone)
		if err != nil && shouldExist {
			return err
		} else if err == nil && !shouldExist {
			return fmt.Errorf("zone %s should not exist", name)
		}

		return nil
	}
}
