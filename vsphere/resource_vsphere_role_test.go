// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const RoleResource = "role1"
const Privilege1 = "Alarm.Acknowledge"
const Privilege2 = "Alarm.Create"
const Privilege3 = "Datacenter.Create"
const Privilege4 = "Datacenter.Move"

func TestAccResourceVsphereRole_createRole(t *testing.T) {
	roleName := "terraform_role" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccResourceVsphereRoleCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereRoleConfigBasic(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVsphereRoleCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.0", Privilege1),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.1", Privilege2),
				),
			},
			{
				ResourceName:      "vsphere_role." + RoleResource,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceVsphereRole_addPrivileges(t *testing.T) {
	roleName := "terraform_role" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccResourceVsphereRoleCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereRoleConfigBasic(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVsphereRoleCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.0", Privilege1),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.1", Privilege2),
				),
			},
			{
				Config: testAccResourceVsphereRoleConfigAdditionalPrivileges(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVsphereRoleCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.0", Privilege1),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.1", Privilege2),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.2", Privilege3),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.3", Privilege4),
				),
			},
		},
	})
}

func TestAccResourceVsphereRole_removePrivileges(t *testing.T) {
	roleName := "terraform_role" + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccResourceVsphereRoleCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereRoleConfigAdditionalPrivileges(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVsphereRoleCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.0", Privilege1),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.1", Privilege2),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.2", Privilege3),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.3", Privilege4),
				),
			},
			{
				Config: testAccResourceVsphereRoleConfigAdditionalPrivileges(roleName),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVsphereRoleCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "name", roleName),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.0", Privilege1),
					resource.TestCheckResourceAttr("vsphere_role."+RoleResource, "role_privileges.1", Privilege2),
				),
			},
		},
	})
}

func TestAccResourceVsphereRole_importSystemRoleShouldError(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:            testAccResourceVsphereRoleConfigSystemRole(),
				ResourceName:      "vsphere_role." + RoleResource,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     NoAccessRoleID,
				ExpectError:       regexp.MustCompile(fmt.Sprintf("error specified role with id %s is a system role. System roles are not supported for this operation", NoAccessRoleID)),
			},
		},
	})
}

func testAccResourceVsphereRoleCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereRole(s, RoleResource)
		if err != nil {
			if strings.Contains(err.Error(), "role not found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected role to be missing")
		}
		return nil
	}
}

func testAccResourceVsphereRoleConfigBasic(roleName string) string {
	return fmt.Sprintf(`
resource "vsphere_role" "%s" {
  name            = "%s"
  role_privileges = ["%s", "%s"]
}
`, RoleResource,
		roleName,
		Privilege1,
		Privilege2,
	)
}

func testAccResourceVsphereRoleConfigAdditionalPrivileges(roleName string) string {
	return fmt.Sprintf(`
resource "vsphere_role" "%s" {
  name            = "%s"
  role_privileges = ["%s", "%s", "%s", "%s"]
}
`, RoleResource,
		roleName,
		Privilege1,
		Privilege2,
		Privilege3,
		Privilege4,
	)
}

func testAccResourceVsphereRoleConfigSystemRole() string {
	return fmt.Sprintf(`
resource "vsphere_role" "%s" {
  name            = "NoAccess"
  role_privileges = []
}
`, RoleResource)
}
