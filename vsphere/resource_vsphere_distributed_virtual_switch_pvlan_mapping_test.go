// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
)

func TestAccResourceVSphereDistributedVirtualSwitchPvlanMapping_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchPvlanMappingExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchPvlanMappingConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchPvlanMappingExists(true),
				),
			},
		},
	})
}

func testAccResourceVSphereDistributedVirtualSwitchPvlanMappingConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch_pvlan_mapping" "mapping" {
  distributed_virtual_switch_id = vsphere_distributed_virtual_switch.dvs.id

  primary_vlan_id   = 1005
  secondary_vlan_id = 1005
  pvlan_type        = "promiscuous"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name                        = "testacc-dvs2"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  ignore_other_pvlan_mappings = true

  host {
    host_system_id = data.vsphere_host.roothost2.id
    devices        = ["%s"]
  }
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost2()),
		testhelper.HostNic1,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchPvlanMappingExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			if viapi.IsAnyNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}

			return err
		}

		mappingToSearchFor, err := testClientVariablesForResource(s, "vsphere_distributed_virtual_switch_pvlan_mapping.mapping")
		if err != nil {
			return fmt.Errorf("could not find pvlan mapping resource: %s", err)
		}

		primaryVlanIDInt64, err := strconv.ParseInt(mappingToSearchFor.resourceAttributes["primary_vlan_id"], 10, 32)
		if err != nil {
			return err
		}
		primaryVlanID := int32(primaryVlanIDInt64)

		secondaryVlanIDInt64, err := strconv.ParseInt(mappingToSearchFor.resourceAttributes["secondary_vlan_id"], 10, 32)
		if err != nil {
			return err
		}
		secondaryVlanID := int32(secondaryVlanIDInt64)

		pvlanType := mappingToSearchFor.resourceAttributes["pvlan_type"]

		for _, mapping := range props.Config.(*types.VMwareDVSConfigInfo).PvlanConfig {
			if mapping.PrimaryVlanId == primaryVlanID && mapping.SecondaryVlanId == secondaryVlanID && mapping.PvlanType == pvlanType {
				if !expected {
					return fmt.Errorf("found PVLAN mapping when not expecting to")
				}
				return nil
			}
		}

		if expected {
			return fmt.Errorf("could not find PVLAN mapping")
		}

		return nil
	}
}
