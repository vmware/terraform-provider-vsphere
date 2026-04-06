// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/vmware/govmomi/vim25/types"
)

const testAccTerraformRequiredVersionActions = `
terraform {
  required_version = ">= 1.14.0"
}
`

func TestAccActionVSphereVirtualMachinePower_offOn(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.RequireAbove(version.Must(version.NewVersion("1.14.0"))),
		},
		CheckDestroy: testAccResourceVSphereVirtualMachineCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccActionVSphereVirtualMachinePowerConfigVMOnly(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccVirtualMachineRuntimePowerState(types.VirtualMachinePowerStatePoweredOn),
				),
			},
			{
				Config: testAccActionVSphereVirtualMachinePowerConfigPowerOff(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccVirtualMachineRuntimePowerState(types.VirtualMachinePowerStatePoweredOff),
				),
			},
			{
				Config: testAccActionVSphereVirtualMachinePowerConfigPowerOffAndOn(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereVirtualMachineCheckExists(true),
					testAccVirtualMachineRuntimePowerState(types.VirtualMachinePowerStatePoweredOn),
				),
			},
		},
	})
}

func testAccVirtualMachineRuntimePowerState(want types.VirtualMachinePowerState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetVirtualMachineProperties(s, "vm")
		if err != nil {
			return err
		}
		if props.Runtime.PowerState != want {
			return fmt.Errorf("expected VM runtime power state %s, got %s", want, props.Runtime.PowerState)
		}
		return nil
	}
}

func testAccActionVSphereVirtualMachinePowerConfigVMOnly() string {
	return fmt.Sprintf(`
%s

%s
`, testAccTerraformRequiredVersionActions, testAccResourceVSphereVirtualMachineConfigBasic())
}

func testAccActionVSphereVirtualMachinePowerConfigPowerOff() string {
	return fmt.Sprintf(`
%s

%s

action "vsphere_virtual_machine_power" "off" {
  config {
    uuid                  = vsphere_virtual_machine.vm.id
    power_state           = "off"
    force_power_off       = true
    shutdown_wait_timeout = 1
  }
}

resource "terraform_data" "power_off" {
  depends_on = [vsphere_virtual_machine.vm]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.vsphere_virtual_machine_power.off]
    }
  }
}
`, testAccTerraformRequiredVersionActions, testAccResourceVSphereVirtualMachineConfigBasic())
}

func testAccActionVSphereVirtualMachinePowerConfigPowerOffAndOn() string {
	return fmt.Sprintf(`
%s

%s

action "vsphere_virtual_machine_power" "off" {
  config {
    uuid                  = vsphere_virtual_machine.vm.id
    power_state           = "off"
    force_power_off       = true
    shutdown_wait_timeout = 1
  }
}

resource "terraform_data" "power_off" {
  depends_on = [vsphere_virtual_machine.vm]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.vsphere_virtual_machine_power.off]
    }
  }
}

action "vsphere_virtual_machine_power" "on" {
  config {
    uuid        = vsphere_virtual_machine.vm.id
    power_state = "on"
  }
}

resource "terraform_data" "power_on" {
  depends_on = [vsphere_virtual_machine.vm, terraform_data.power_off]

  lifecycle {
    action_trigger {
      events  = [after_create]
      actions = [action.vsphere_virtual_machine_power.on]
    }
  }
}
`, testAccTerraformRequiredVersionActions, testAccResourceVSphereVirtualMachineConfigBasic())
}
