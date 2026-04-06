// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const NoAccessRoleDescription = "No access"
const NoAccessRoleName = "NoAccess"
const NoAccessRoleID = "-5"

func TestAccDataSourceVSphereRole_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereRoleConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "id",
						"vsphere_role.test-role", "id",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "name",
						"vsphere_role.test-role", "name",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.0",
						"vsphere_role.test-role", "role_privileges.0",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.1",
						"vsphere_role.test-role", "role_privileges.1",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.2",
						"vsphere_role.test-role", "role_privileges.2",
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_role.role1", "role_privileges.3",
						"vsphere_role.test-role", "role_privileges.3",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereRole_systemRoleData(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereRoleSystemRoleConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_role.role1", "name", NoAccessRoleName),
					resource.TestCheckResourceAttr("data.vsphere_role.role1", "id", NoAccessRoleID),
					resource.TestCheckResourceAttr("data.vsphere_role.role1", "role_privileges.#", "0")),
			},
		},
	})
}

func testAccDataSourceVSphereRoleConfig() string {
	return fmt.Sprintf(`
resource "vsphere_role" test-role {
  name            = "terraform-test-role1"
  role_privileges = ["%s", "%s", "%s", "%s"]
}

data "vsphere_role" "role1" {
  label = vsphere_role.test-role.label
}
`, Privilege1,
		Privilege2,
		Privilege3,
		Privilege4,
	)
}

func testAccDataSourceVSphereRoleSystemRoleConfig() string {
	return fmt.Sprintf(`
data "vsphere_role" "role1" {
  label = "%s"
}
`,
		NoAccessRoleDescription,
	)
}
