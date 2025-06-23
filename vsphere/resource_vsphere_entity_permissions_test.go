// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
)

const EntityPermissionResource = "entity_permission1"

func TestAccResourcevsphereEntityPermissions_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceEntityPermissionsCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVsphereEntityPermissionsConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceEntityPermissionsCheckExists(true),
					resource.TestCheckResourceAttrSet("vsphere_entity_permissions."+EntityPermissionResource, "permissions.0.user_or_group"),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+EntityPermissionResource, "permissions.0.propagate", "true"),
					resource.TestCheckResourceAttr("vsphere_entity_permissions."+EntityPermissionResource, "permissions.0.is_group", "true"),
				),
			},
		},
	})
}

func testAccResourceEntityPermissionsCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetVsphereEntityPermission(s, EntityPermissionResource)
		if err != nil {
			if strings.Contains(err.Error(), "permissions not found") && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected permissions to be missing")
		}
		return nil
	}
}

func testAccResourceVsphereEntityPermissionsConfigBasic() string {
	return fmt.Sprintf(`
%s

data "vsphere_role" "role1" {
  label = "Administrator"
}

resource vsphere_entity_permissions "%s" {
  entity_id   = data.vsphere_datacenter.rootdc1.id
  entity_type = "Datacenter"
  permissions {
    user_or_group = "%s"
    propagate     = true
    is_group      = true
    role_id       = data.vsphere_role.role1.id
  }
}
`, testhelper.ConfigDataRootDC1(),
		EntityPermissionResource,
		"root",
	)
}
