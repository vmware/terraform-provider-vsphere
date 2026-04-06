// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereCustomAttribute_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereCustomAttributeConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_custom_attribute.testacc-attribute-data",
						"name",
						testAccDataSourceVSphereCustomAttributeConfigName,
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_custom_attribute.testacc-attribute-data",
						"managed_object_type",
						testAccDataSourceVSphereCustomAttributeConfigType,
					),
					resource.TestCheckResourceAttrPair(
						"data.vsphere_custom_attribute.testacc-attribute-data", "id",
						"vsphere_custom_attribute.testacc-attribute", "id",
					),
				),
			},
		},
	})
}

const testAccDataSourceVSphereCustomAttributeConfigName = "testacc-attribute"
const testAccDataSourceVSphereCustomAttributeConfigType = "VirtualMachine"

func testAccDataSourceVSphereCustomAttributeConfig() string {
	return fmt.Sprintf(`
variable "attribute_name" {
  default = "%s"
}

variable "attribute_type" {
  default = "%s"
}

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = var.attribute_name
  managed_object_type = var.attribute_type
}

data "vsphere_custom_attribute" "testacc-attribute-data" {
  name = vsphere_custom_attribute.testacc-attribute.name
}
`,
		testAccDataSourceVSphereCustomAttributeConfigName,
		testAccDataSourceVSphereCustomAttributeConfigType,
	)
}
