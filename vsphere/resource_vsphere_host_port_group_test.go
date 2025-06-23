// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereHostPortGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostPortGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostPortGroup_complexWithOverrides(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostPortGroupConfigWithOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
					testAccResourceVSphereHostPortGroupCheckVlan(1000),
					testAccResourceVSphereHostPortGroupCheckEffectiveActive([]string{testhelper.HostNic1}),
					testAccResourceVSphereHostPortGroupCheckEffectiveStandby([]string{testhelper.HostNic2}),
					testAccResourceVSphereHostPortGroupCheckEffectivePromisc(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostPortGroup_basicToComplex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostPortGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostPortGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
				),
			},
			{
				Config: testAccResourceVSphereHostPortGroupConfigWithOverrides(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostPortGroupExists(true),
					testAccResourceVSphereHostPortGroupCheckVlan(1000),
					testAccResourceVSphereHostPortGroupCheckEffectiveActive([]string{testhelper.HostNic1}),
					testAccResourceVSphereHostPortGroupCheckEffectiveStandby([]string{testhelper.HostNic2}),
					testAccResourceVSphereHostPortGroupCheckEffectivePromisc(true),
				),
			},
		},
	})
}

func testAccResourceVSphereHostPortGroupExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		name := "PGTerraformTest"
		id := "pg"
		_, err := testGetPortGroup(s, id)
		if err != nil {
			if err.Error() == fmt.Sprintf("could not find port group %s", name) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if expected == false {
			return fmt.Errorf("expected port group %s to still be missing", name)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckVlan(expected int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := pg.Spec.VlanId
		if expected != actual {
			return fmt.Errorf("expected VLAN ID to be %d, got %d", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckEffectiveActive(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := pg.ComputedPolicy.NicTeaming.NicOrder.ActiveNic
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected effective active NICs to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckEffectiveStandby(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := pg.ComputedPolicy.NicTeaming.NicOrder.StandbyNic
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected effective standby NICs to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupCheckEffectivePromisc(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		id := "pg"
		pg, err := testGetPortGroup(s, id)
		if err != nil {
			return err
		}
		actual := *pg.ComputedPolicy.Security.AllowPromiscuous
		if expected != actual {
			return fmt.Errorf("expected effective allow promiscuous to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereHostPortGroupConfig() string {
	return fmt.Sprintf(`
variable "host_nic1" {
  default = "%s"
}

%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_host_virtual_switch" "switch" {
  name           = "vSwitchTerraformTest2"
  host_system_id = data.vsphere_host.esxi_host.id

  network_adapters = [var.host_nic1]
  active_nics      = [var.host_nic1]
  standby_nics     = []
}

resource "vsphere_host_port_group" "pg" {
  name                = "PGTerraformTest"
  host_system_id      = data.vsphere_host.esxi_host.id
  virtual_switch_name = vsphere_host_virtual_switch.switch.name
}
`, testhelper.HostNic1,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}

func testAccResourceVSphereHostPortGroupConfigWithOverrides() string {
	return fmt.Sprintf(`
variable "host_nic1" {
  default = "%s"
}

variable "host_nic2" {
  default = "%s"
}

%s

data "vsphere_host" "esxi_host" {
  name          = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_host_virtual_switch" "switch" {
  name           = "vSwitchTerraformTest2"
  host_system_id = data.vsphere_host.esxi_host.id

  network_adapters  = [var.host_nic1, var.host_nic2]
  active_nics       = [var.host_nic1]
  standby_nics      = [var.host_nic2]
  allow_promiscuous = false
}

resource "vsphere_host_port_group" "pg" {
  name                = "PGTerraformTest"
  host_system_id      = data.vsphere_host.esxi_host.id
  virtual_switch_name = vsphere_host_virtual_switch.switch.name

  vlan_id           = 1000
  active_nics       = [var.host_nic1]
  standby_nics      = [var.host_nic2]
  allow_promiscuous = true
}
`, testhelper.HostNic1,
		testhelper.HostNic2,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}
