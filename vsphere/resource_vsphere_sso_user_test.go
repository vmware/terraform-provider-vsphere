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

const (
	testAccSSOUserName     = "acc-test-user"
	testAccSSOUserPassword = "AccTestP@ssw0rd!"
	testAccSSOGroupName    = "acc-test-group"
)

func TestAccResourceVSphereSSOUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereSSOUserCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereSSOUserConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereSSOUserCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_sso_user.acc-test-user", "name", testAccSSOUserName),
					resource.TestCheckResourceAttrSet("vsphere_sso_user.acc-test-user", "domain"),
					resource.TestCheckResourceAttr("vsphere_sso_user.acc-test-user", "first_name", "Acc"),
					resource.TestCheckResourceAttr("vsphere_sso_user.acc-test-user", "last_name", "Test"),
					resource.TestCheckResourceAttr("vsphere_sso_user.acc-test-user", "email_address", "acc-test@example.com"),
					resource.TestCheckResourceAttr("vsphere_sso_user.acc-test-user", "description", "Managed by Terraform acceptance test"),
				),
			},
			{
				ResourceName:            "vsphere_sso_user.acc-test-user",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"},
			},
		},
	})
}

func testGetSSOUser(s *terraform.State, resourceAddr string) (*ssoadmintypes.AdminPersonUser, error) {
	vars, err := testClientVariablesForResource(s, resourceAddr)
	if err != nil {
		return nil, err
	}
	client, err := testAccProvider.Meta().(*Client).SSOAdminClient(context.Background())
	if err != nil {
		return nil, err
	}
	return client.FindPersonUser(context.Background(), vars.resourceID)
}

func testAccResourceVSphereSSOUserCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		user, err := testGetSSOUser(s, "vsphere_sso_user.acc-test-user")
		if err != nil {
			return err
		}
		if user == nil && expected {
			return errors.New("expected SSO user to exist")
		}
		if user != nil && !expected {
			return errors.New("expected SSO user to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereSSOUserConfigBasic() string {
	return fmt.Sprintf(`
resource "vsphere_sso_user" "acc-test-user" {
  name          = %q
  password      = %q
  first_name    = "Acc"
  last_name     = "Test"
  email_address = "acc-test@example.com"
  description   = "Managed by Terraform acceptance test"
}
`, testAccSSOUserName, testAccSSOUserPassword)
}
