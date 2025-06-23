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
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereOfflineSoftwareDepot_basic(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereOfflineSoftwareDepotPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testhelper.ConfigDataSoftwareDepot(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereOfflineSoftwareDepotCheckFunc(),
				),
			},
		},
	})
}

func testAccResourceVSphereOfflineSoftwareDepotCheckFunc() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		tVars, err := testClientVariablesForResource(s, "vsphere_offline_software_depot.depot")
		if err != nil {
			return err
		}

		location := tVars.resourceAttributes["location"]
		expected := os.Getenv("TF_VAR_VSPHERE_SOFTWARE_DEPOT_LOCATION")
		if location != expected {
			return fmt.Errorf("depot location is incorrect. Expected %s but got %s", expected, location)
		}

		return nil
	}
}

func testAccResourceVSphereOfflineSoftwareDepotPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_SOFTWARE_DEPOT_LOCATION") == "" {
		t.Skip("set TF_VAR_VSPHERE_SOFTWARE_DEPOT_LOCATION to run vsphere_offline_software_depot acceptance tests")
	}
}
