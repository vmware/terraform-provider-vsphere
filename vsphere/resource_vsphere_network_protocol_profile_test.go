// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereDatacenterNetworkProtocolProfile_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDatacenterNetworkProtocolProfileExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDatacenterNetworkProtocolProfileConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatacenterNetworkProtocolProfileExists(true),
					testAccResourceVSphereDatacenterNetworkProtocolProfileHasName("testacc-profile"),
				),
			},
			{
				ResourceName:      "vsphere_datacenter_network_protocol_profile.profile",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					profile, err := testGetNetworkProtocolProfile(s, "profile")
					if err != nil {
						return "", err
					}
					if profile == nil {
						return "", errors.New("network protocol profile does not exist")
					}
					rs, ok := s.RootModule().Resources["vsphere_datacenter_network_protocol_profile.profile"]
					if !ok {
						return "", errors.New("resource not found in state")
					}
					return fmt.Sprintf("%s:%d", rs.Primary.Attributes["datacenter_id"], profile.Id), nil
				},
				Config: testAccResourceVSphereDatacenterNetworkProtocolProfileConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDatacenterNetworkProtocolProfileExists(true),
				),
			},
		},
	})
}

func testAccResourceVSphereDatacenterNetworkProtocolProfileExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := testGetNetworkProtocolProfile(s, "profile")
		if err != nil {
			return err
		}
		if profile == nil && expected {
			return errors.New("expected network protocol profile to exist")
		} else if profile != nil && !expected {
			return errors.New("expected network protocol profile to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereDatacenterNetworkProtocolProfileHasName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		profile, err := testGetNetworkProtocolProfile(s, "profile")
		if err != nil {
			return err
		}
		actual := profile.Name
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDatacenterNetworkProtocolProfileConfigBasic() string {
	return testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1(), `
resource "vsphere_datacenter_network_protocol_profile" "profile" {
  name          = "testacc-profile"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  network_ids   = [data.vsphere_network.network1.id]

  ipv4 {
    subnet  = "10.10.10.0"
    netmask = "255.255.255.0"
    gateway = "10.10.10.1"
    range   = "10.10.10.100#100"
  }
}
`)
}
