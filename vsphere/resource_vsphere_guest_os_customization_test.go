// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/guestoscustomizations"
)

func TestAccResourceVSpherGOSC_windows_basic(t *testing.T) {
	goscName := acctest.RandomWithPrefix("win")
	goscResourceName := acctest.RandomWithPrefix("gosc")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccGOSCExists(goscResourceName, goscName, false),
		Steps: []resource.TestStep{
			{
				Config: testAccGOSCWindows(goscResourceName, goscName),
				Check:  testAccGOSCExists(goscResourceName, goscName, true),
			},
		},
	})
}

func TestAccResourceVSpherGOSC_windows_workGroup(t *testing.T) {
	goscName := acctest.RandomWithPrefix("win")
	goscResourceName := acctest.RandomWithPrefix("gosc")
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccGOSCExists(goscResourceName, goscName, false),
		Steps: []resource.TestStep{
			{
				Config: testAccGOSCWindowsAllPropsWorkGroup(goscResourceName, goscName),
				Check:  testAccGOSCExists(goscResourceName, goscName, true),
			},
		},
	})
}

func TestAccResourceVSpherGOSC_linux(t *testing.T) {
	goscName := acctest.RandomWithPrefix("lin")
	goscResourceName := acctest.RandomWithPrefix("gosc")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccGOSCExists(goscResourceName, goscName, false),
		Steps: []resource.TestStep{
			{
				Config: testAccGOSCLinux(goscResourceName, goscName),
				Check:  testAccGOSCExists(goscResourceName, goscName, true),
			},
		},
	})
}

func TestAccResourceVSpherGOSC_sysprep(t *testing.T) {
	goscName := acctest.RandomWithPrefix("lin")
	goscResourceName := acctest.RandomWithPrefix("gosc")
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccGOSCExists(goscResourceName, goscName, false),
		Steps: []resource.TestStep{
			{
				Config: testAccGOSCWindowsPrep(goscResourceName, goscName),
				Check:  testAccGOSCExists(goscResourceName, goscName, true),
			},
		},
	})
}

func testAccGOSCExists(resourceName string, goscName string, expectToExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource := fmt.Sprintf("vsphere_guest_os_customization.%s", resourceName)
		vars, err := testClientVariablesForResource(s, resource)
		if err != nil {
			return err
		}
		_, err = guestoscustomizations.FromName(vars.client, goscName)
		if err != nil && expectToExist {
			return err
		}

		return nil
	}
}

func testAccGOSCWindows(resourceName string, goscName string) string {
	return fmt.Sprintf(`
		resource "vsphere_guest_os_customization" %q {
			name = %q
			type = "Windows"
			spec {
				windows_options {
					computer_name = "windows"
				}
			}
		}
	`,
		resourceName,
		goscName,
	)
}

func testAccGOSCWindowsAllPropsWorkGroup(resourceName string, goscName string) string {
	return fmt.Sprintf(`
		resource "vsphere_guest_os_customization" %q {
			name = %q
			type = "Windows"
			spec {
				windows_options {
					run_once_command_list = ["command-1", "command-2"]
					computer_name = "windows"
					auto_logon = false
					auto_logon_count = 0
					admin_password = "VMware1!"
					time_zone = 004 #(GMT-08:00) Pacific Time (US and Canada); Tijuana
					workgroup = "workgroup"
				}
			}
		}
	`,
		resourceName,
		goscName,
	)
}

func testAccGOSCWindowsPrep(resourceName string, goscName string) string {
	return fmt.Sprintf(`
		resource "vsphere_guest_os_customization" %q {
			name = %q
			type = "Windows"
			spec {
				windows_sysprep_text = "Test sysprep text"
			}
		}
	`,
		resourceName,
		goscName,
	)
}

func testAccGOSCLinux(resourceName string, goscName string) string {
	return fmt.Sprintf(`
		resource "vsphere_guest_os_customization" %q {
			name = %q
			type = "Linux"
			spec {
				linux_options {
					domain = "example.com"
					host_name = "linux"
				}
				network_interface {}
			}
		}
	`,
		resourceName,
		goscName,
	)
}
