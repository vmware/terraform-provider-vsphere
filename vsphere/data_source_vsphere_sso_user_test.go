// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereSSOUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereSSOUserConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_sso_user.acc-test-user", "id"),
					resource.TestCheckResourceAttrPair("data.vsphere_sso_user.acc-test-user", "domain", "vsphere_sso_user.acc-test-user", "domain"),
					resource.TestCheckResourceAttrPair("data.vsphere_sso_user.acc-test-user", "first_name", "vsphere_sso_user.acc-test-user", "first_name"),
					resource.TestCheckResourceAttrPair("data.vsphere_sso_user.acc-test-user", "last_name", "vsphere_sso_user.acc-test-user", "last_name"),
					resource.TestCheckResourceAttrPair("data.vsphere_sso_user.acc-test-user", "email_address", "vsphere_sso_user.acc-test-user", "email_address"),
					resource.TestCheckResourceAttrPair("data.vsphere_sso_user.acc-test-user", "description", "vsphere_sso_user.acc-test-user", "description"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereSSOUserConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_sso_user" acc-test-user {
  name = vsphere_sso_user.acc-test-user.name
}
`, testAccResourceVSphereSSOUserConfigBasic())
}
