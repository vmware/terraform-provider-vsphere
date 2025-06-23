// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/folder"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
)

const (
	testAccResourceVSphereDistributedVirtualSwitchUpperVersion = "8.0.0"
	testAccResourceVSphereDistributedVirtualSwitchLowerVersion = "7.0.0"
)

func TestAccResourceVSphereDistributedVirtualSwitch_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
				),
			},
			{
				ResourceName:            "vsphere_distributed_virtual_switch.dvs",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"vlan_range", "ignore_other_pvlan_mappings"},
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					dvs, err := testGetDVS(s, "dvs")
					if err != nil {
						return "", err
					}
					return dvs.InventoryPath, nil
				},
				Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_noHosts(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigNoHosts(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_removeNIC(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
				),
			},
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleNIC(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_standbyWithExplicitFailoverOrder(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasUplinks([]string{testhelper.HostNic0, testhelper.HostNic1}),
					testAccResourceVSphereDistributedVirtualSwitchHasActiveUplinks([]string{testhelper.HostNic0}),
					testAccResourceVSphereDistributedVirtualSwitchHasStandbyUplinks([]string{testhelper.HostNic1}),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_basicToStandbyWithFailover(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
				),
			},
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasUplinks([]string{testhelper.HostNic0, testhelper.HostNic1}),
					testAccResourceVSphereDistributedVirtualSwitchHasActiveUplinks([]string{testhelper.HostNic0}),
					testAccResourceVSphereDistributedVirtualSwitchHasStandbyUplinks([]string{testhelper.HostNic1}),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_upgradeVersion(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigStaticVersion(testAccResourceVSphereDistributedVirtualSwitchLowerVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasVersion(testAccResourceVSphereDistributedVirtualSwitchLowerVersion),
				),
			},
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigStaticVersion(testAccResourceVSphereDistributedVirtualSwitchUpperVersion),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasVersion(testAccResourceVSphereDistributedVirtualSwitchUpperVersion),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_networkResourceControl(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigNetworkResourceControl(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlEnabled(),
					testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlVersion(
						types.DistributedVirtualSwitchNetworkResourceControlVersionVersion3,
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_explicitUplinks(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigUplinks(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasUplinks([]string{testhelper.HostNic0, testhelper.HostNic1}),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_modifyUplinks(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasUplinks(
						[]string{
							"uplink1",
							"uplink2",
							"uplink3",
							"uplink4",
						},
					),
				),
			},
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasUplinks(
						[]string{
							testhelper.HostNic0,
							testhelper.HostNic1,
						},
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_inFolder(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigInFolder(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchMatchInventoryPath("tf-network-folder"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_singleTag(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_modifyTags(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_netflow(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigNetflow(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasNetflow(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_vlanRanges(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigMultiVlanRange(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(1000, 1999),
					testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(3000, 3999),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_singleCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDistributedVirtualSwitch_multiCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereDistributedVirtualSwitchPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereDistributedVirtualSwitchExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchCheckCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereDistributedVirtualSwitchConfigMultiCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereDistributedVirtualSwitchExists(true),
					testAccResourceVSphereDistributedVirtualSwitchCheckCustomAttributes(),
				),
			},
		},
	})
}

func testAccResourceVSphereDistributedVirtualSwitchPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_host_virtual_switch acceptance tests")
	}
}

func testAccResourceVSphereDistributedVirtualSwitchExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVS(s, "dvs")
		if err != nil {
			if viapi.IsAnyNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected DVS %s to be missing", dvs.Reference().Value)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasVersion(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actual := props.Summary.ProductInfo.Version
		if expected != actual {
			return fmt.Errorf("expected version to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlEnabled() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actual := props.Config.(*types.VMwareDVSConfigInfo).NetworkResourceManagementEnabled
		if actual == nil || !*actual {
			return errors.New("expected network resource control to be enabled")
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasNetworkResourceControlVersion(expected types.DistributedVirtualSwitchNetworkResourceControlVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actual := props.Config.(*types.VMwareDVSConfigInfo).NetworkResourceControlVersion
		if string(expected) != actual {
			return fmt.Errorf("expected network resource control version to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasUplinks(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		policy := props.Config.(*types.VMwareDVSConfigInfo).UplinkPortPolicy.(*types.DVSNameArrayUplinkPortPolicy)
		actual := policy.UplinkPortName
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected uplinks to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasActiveUplinks(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		pc := props.Config.(*types.VMwareDVSConfigInfo).DefaultPortConfig.(*types.VMwareDVSPortSetting)
		actual := pc.UplinkTeamingPolicy.UplinkPortOrder.ActiveUplinkPort
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected active uplinks to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasStandbyUplinks(expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		pc := props.Config.(*types.VMwareDVSConfigInfo).DefaultPortConfig.(*types.VMwareDVSPortSetting)
		actual := pc.UplinkTeamingPolicy.UplinkPortOrder.StandbyUplinkPort
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("expected standby uplinks to be %#v, got %#v", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasNetflow() resource.TestCheckFunc {
	expectedIPv4Addr := "10.0.0.100"
	expectedIpfixConfig := &types.VMwareIpfixConfig{
		CollectorIpAddress:  "10.0.0.10",
		CollectorPort:       9000,
		ObservationDomainId: 1000,
		ActiveFlowTimeout:   90,
		IdleFlowTimeout:     20,
		SamplingRate:        10,
		InternalFlowsOnly:   true,
	}

	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		actualIPv4Addr := props.Config.(*types.VMwareDVSConfigInfo).SwitchIpAddress
		actualIpfixConfig := props.Config.(*types.VMwareDVSConfigInfo).IpfixConfig

		if expectedIPv4Addr != actualIPv4Addr {
			return fmt.Errorf("expected switch IP to be %s, got %s", expectedIPv4Addr, actualIPv4Addr)
		}
		if !reflect.DeepEqual(expectedIpfixConfig, actualIpfixConfig) {
			return fmt.Errorf("expected netflow config to be %#v, got %#v", expectedIpfixConfig, actualIpfixConfig)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchHasVlanRange(emin, emax int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		pc := props.Config.(*types.VMwareDVSConfigInfo).DefaultPortConfig.(*types.VMwareDVSPortSetting)
		ranges := pc.Vlan.(*types.VmwareDistributedVirtualSwitchTrunkVlanSpec).VlanId
		var found bool
		for _, rng := range ranges {
			if rng.Start == emin && rng.End == emax {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("could not find start %d and end %d in %#v", emin, emax, ranges)
		}
		return nil
	}
}

func testAccResourceVSphereDistributedVirtualSwitchMatchInventoryPath(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVS(s, "dvs")
		if err != nil {
			return err
		}

		expected, err = folder.RootPathParticleNetwork.PathFromNewRoot(dvs.InventoryPath, folder.RootPathParticleNetwork, expected)
		actual := path.Dir(dvs.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

// testAccResourceVSphereDistributedVirtualSwitchCheckTags is a check to ensure that any tags
// that have been created with the supplied resource name have been attached to
// the folder.
func testAccResourceVSphereDistributedVirtualSwitchCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		dvs, err := testGetDVS(s, "dvs")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, dvs, tagResName)
	}
}

func testAccResourceVSphereDistributedVirtualSwitchCheckCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetDVSProperties(s, "dvs")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_distributed_virtual_switch", "dvs", props.Entity())
	}
}

func testAccResourceVSphereDistributedVirtualSwitchConfig() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  host {
    host_system_id = data.vsphere_host.roothost3.id
    devices = ["%s"]
  }
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost3()),
		testhelper.HostNic1,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigStaticVersion(version string) string {
	return fmt.Sprintf(`
%s

variable "network_interfaces" {
  default = [
    "%s",
  ]
}

variable "dvs_version" {
  default = "%s"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  version       = "${var.dvs_version}"

  host {
    host_system_id = "${data.vsphere_host.roothost1.id}"
    devices = "${var.network_interfaces}"
  }
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost1()),
		testhelper.HostNic1,
		version,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigSingleNIC() string {
	return fmt.Sprintf(`
%s

variable "network_interfaces" {
  default = [
    "%s",
  ]
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  host {
    host_system_id = "${data.vsphere_host.roothost1.id}"
    devices = "${var.network_interfaces}"
  }

  host {
    host_system_id = "${data.vsphere_host.roothost2.id}"
    devices = "${var.network_interfaces}"
  }
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
		),
		testhelper.HostNic1,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigNetworkResourceControl() string {
	return fmt.Sprintf(`
%s


variable "network_interfaces" {
  default = [
    "%s",
  ]
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  network_resource_control_enabled = true
  network_resource_control_version = "version3"

  host {
    host_system_id = "${data.vsphere_host.roothost1.id}"
    devices = "${var.network_interfaces}"
  }

  host {
    host_system_id = "${data.vsphere_host.roothost2.id}"
    devices = "${var.network_interfaces}"
  }
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		testhelper.HostNic1,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigUplinks() string {
	return fmt.Sprintf(`
%s

variable "network_interfaces" {
  default = [
    "%s",
    "%s"
  ]
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  uplinks = var.network_interfaces

  host {
    host_system_id = "${data.vsphere_host.roothost1.id}"
    devices = "${var.network_interfaces}"
  }

  host {
    host_system_id = "${data.vsphere_host.roothost2.id}"
    devices = "${var.network_interfaces}"
  }
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		testhelper.HostNic0,
		testhelper.HostNic1,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigStandbyLink() string {
	return fmt.Sprintf(`
%s

variable "network_interfaces" {
  default = [
    "%s",
	"%s"
  ]
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  uplinks         = var.network_interfaces
  active_uplinks  = [var.network_interfaces.0]
  standby_uplinks = [var.network_interfaces.1]

  host {
    host_system_id = "${data.vsphere_host.roothost1.id}"
    devices = "${var.network_interfaces}"
  }

  host {
    host_system_id = "${data.vsphere_host.roothost2.id}"
    devices = "${var.network_interfaces}"
  }
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2()),
		testhelper.HostNic0,
		testhelper.HostNic1,
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigNoHosts() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigInFolder() string {
	return fmt.Sprintf(`
%s

resource "vsphere_folder" "folder" {
  path          = "tf-network-folder"
  type          = "network"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  folder        = "${vsphere_folder.folder.path}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigSingleTag() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VmwareDistributedVirtualSwitch",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  tags          = ["${vsphere_tag.testacc-tag.id}"]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigMultiTag() string {
	return fmt.Sprintf(`
%s

variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "VmwareDistributedVirtualSwitch",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_tag" "testacc-tags-alt" {
  count       = "${length(var.extra_tags)}"
  name        = "${var.extra_tags[count.index]}"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  tags          = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigNetflow() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  ipv4_address                  = "10.0.0.100"
  netflow_enabled               = true
  netflow_active_flow_timeout   = 90
  netflow_collector_ip_address  = "10.0.0.10"
  netflow_collector_port        = 9000
  netflow_idle_flow_timeout     = 20
  netflow_internal_flows_only   = true
  netflow_observation_domain_id = 1000
  netflow_sampling_rate         = 10
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigMultiVlanRange() string {
	return fmt.Sprintf(`
%s

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  vlan_range {
    min_vlan = 1000
    max_vlan = 1999
  }

  vlan_range {
    min_vlan = 3000
    max_vlan = 3999
  }
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigSingleCustomAttribute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "VmwareDistributedVirtualSwitch"
}

locals {
  vs_attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.vs_attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereDistributedVirtualSwitchConfigMultiCustomAttribute() string {
	return fmt.Sprintf(`
%s

variable "custom_attrs" {
  default = [
    "testacc-attribute-1",
    "terraform-test-attriubute-2"
  ]
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "VmwareDistributedVirtualSwitch"
}

resource "vsphere_custom_attribute" "testacc-attribute-alt" {
  count               = "${length(var.custom_attrs)}"
  name                = "${var.custom_attrs[count.index]}"
  managed_object_type = "VmwareDistributedVirtualSwitch"
}

locals {
  vs_attrs = {
    "${vsphere_custom_attribute.testacc-attribute-alt.0.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-alt.1.id}" = "value-2"
  }
}

resource "vsphere_distributed_virtual_switch" "dvs" {
  name          = "testacc-dvs1"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.vs_attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
