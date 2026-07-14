// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereSSOGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereSSOGroupConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_sso_group.acc-test-group", "id"),
					resource.TestCheckResourceAttrPair("data.vsphere_sso_group.acc-test-group", "description", "vsphere_sso_group.acc-test-group", "description"),
					resource.TestCheckResourceAttr("data.vsphere_sso_group.acc-test-group", "member_user.#", "1"),
					resource.TestCheckResourceAttr("data.vsphere_sso_group.acc-test-group", "member_user.0.name", testAccSSOUserName),
				),
			},
		},
	})
}

func testAccDataSourceVSphereSSOGroupConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_sso_group" acc-test-group {
  name = vsphere_sso_group.acc-test-group.name
}
`, testAccResourceVSphereSSOGroupConfigBasic())
}
