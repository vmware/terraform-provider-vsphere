// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereHostVirtualSwitch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
			{
				ResourceName:      "vsphere_host_virtual_switch.switch",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					vars, err := testClientVariablesForResource(s, fmt.Sprintf("vsphere_host_virtual_switch.%s", "switch"))
					if err != nil {
						return "", err
					}
					return vars.resourceID, err
				},
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_removeNIC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfigSingleNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_noNICs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfigNoNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
					testAccResourceVSphereHostVirtualSwitchNoBridge(),
				),
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_badActiveNICList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereHostVirtualSwitchConfigBadActive(),
				ExpectError: regexp.MustCompile(fmt.Sprintf("active NIC entry %q not present in network_adapters list", testhelper.HostNic1)),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_badStandbyNICList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config:      testAccResourceVSphereHostVirtualSwitchConfigBadStandby(),
				ExpectError: regexp.MustCompile(fmt.Sprintf("standby NIC entry %q not present in network_adapters list", testhelper.HostNic1)),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccResourceVSphereHostVirtualSwitch_removeAllNICs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereHostVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
				),
			},
			{
				Config: testAccResourceVSphereHostVirtualSwitchConfigNoNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereHostVirtualSwitchExists(true),
					testAccResourceVSphereHostVirtualSwitchNoBridge(),
				),
			},
		},
	})
}

func testAccResourceVSphereHostVirtualSwitchExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_host_virtual_switch.switch")
		if err != nil {
			if expected {
				return errors.New("vsphere_host_virtual_switch.switch not found in state")
			}
			return nil
		}

		hsID, name, err := splitHostVirtualSwitchID(vars.resourceID)
		if err != nil {
			return err
		}
		ns, err := hostNetworkSystemFromHostSystemID(vars.client, hsID)
		if err != nil {
			return fmt.Errorf("error loading host network system: %s", err)
		}

		_, err = hostVSwitchFromName(vars.client, ns, name)
		if err != nil {
			if err.Error() == fmt.Sprintf("could not find virtual switch %s", name) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected vSwitch %s to be missing", name)
		}
		return nil
	}
}

func testAccResourceVSphereHostVirtualSwitchNoBridge() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_host_virtual_switch.switch")
		if err != nil {
			return errors.New("vsphere_host_virtual_switch.switch not found in state")
		}

		hsID, name, err := splitHostVirtualSwitchID(vars.resourceID)
		if err != nil {
			return err
		}
		ns, err := hostNetworkSystemFromHostSystemID(vars.client, hsID)
		if err != nil {
			return fmt.Errorf("error loading host network system: %s", err)
		}

		sw, err := hostVSwitchFromName(vars.client, ns, name)
		if err != nil {
			return err
		}
		if sw.Spec.Bridge != nil {
			return fmt.Errorf("expected no bridge on switch, got %+v", sw.Spec.Bridge)
		}
		return nil
	}
}

func testAccResourceVSphereHostVirtualSwitchConfig() string {
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

  network_adapters = [var.host_nic1, var.host_nic2]

  active_nics  = [var.host_nic1]
  standby_nics = [var.host_nic2]
}
`, testhelper.HostNic1,
		testhelper.HostNic2,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}

func testAccResourceVSphereHostVirtualSwitchConfigSingleNIC() string {
	return fmt.Sprintf(`
variable "host_nic0" {
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

  network_adapters = [var.host_nic0]

  active_nics = [var.host_nic0]
}
`, testhelper.HostNic1,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}

func testAccResourceVSphereHostVirtualSwitchConfigNoNIC() string {
	return fmt.Sprintf(`
variable "host_nic0" {
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

  network_adapters = []

  active_nics  = []
  standby_nics = []
}
`, os.Getenv("TF_VAR_VSPHERE_ESXI_TRUNK_NIC"),
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}

func testAccResourceVSphereHostVirtualSwitchConfigBadActive() string {
	return fmt.Sprintf(`
variable "host_nic0" {
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

  network_adapters = []

  active_nics  = [var.host_nic0]
  standby_nics = []
}
`, testhelper.HostNic1,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}

func testAccResourceVSphereHostVirtualSwitchConfigBadStandby() string {
	return fmt.Sprintf(`
variable "host_nic0" {
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

  network_adapters = []

  active_nics  = []
  standby_nics = [var.host_nic0]
}
`, testhelper.HostNic1,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		os.Getenv("TF_VAR_VSPHERE_ESXI3"))
}
