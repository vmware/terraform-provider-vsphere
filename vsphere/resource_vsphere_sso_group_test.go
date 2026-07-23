// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	ssoadmintypes "github.com/vmware/govmomi/ssoadmin/types"
)

func TestAccResourceVSphereSSOGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereSSOGroupCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereSSOGroupConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereSSOGroupCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_sso_group.acc-test-group", "name", testAccSSOGroupName),
					resource.TestCheckResourceAttr("vsphere_sso_group.acc-test-group", "description", "Managed by Terraform acceptance test"),
					resource.TestCheckResourceAttr("vsphere_sso_group.acc-test-group", "member_user.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vsphere_sso_group.acc-test-group", "member_user.*", map[string]string{
						"name": testAccSSOUserName,
					}),
				),
			},
			{
				ResourceName:      "vsphere_sso_group.acc-test-group",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testGetSSOGroup(s *terraform.State, resourceAddr string) (*ssoadmintypes.AdminGroup, error) {
	vars, err := testClientVariablesForResource(s, resourceAddr)
	if err != nil {
		return nil, err
	}
	client, err := testAccProvider.Meta().(*Client).SSOAdminClient(context.Background())
	if err != nil {
		return nil, err
	}
	return client.FindGroup(context.Background(), vars.resourceID)
}

func testAccResourceVSphereSSOGroupCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		group, err := testGetSSOGroup(s, "vsphere_sso_group.acc-test-group")
		if err != nil {
			return err
		}
		if group == nil && expected {
			return errors.New("expected SSO group to exist")
		}
		if group != nil && !expected {
			return errors.New("expected SSO group to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereSSOGroupConfigBasic() string {
	return fmt.Sprintf(`
resource "vsphere_sso_user" "acc-test-user" {
  name     = %q
  password = %q
}

resource "vsphere_sso_group" "acc-test-group" {
  name        = %q
  description = "Managed by Terraform acceptance test"

  member_user {
    name   = vsphere_sso_user.acc-test-user.name
    domain = vsphere_sso_user.acc-test-user.domain
  }
}
`, testAccSSOUserName, testAccSSOUserPassword, testAccSSOGroupName)
}
