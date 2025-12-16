// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereNamespace_basic(t *testing.T) {
	// cannot run on the standard testbed
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNamespaceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "config_status"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "cpu_usage"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "memory_usage"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "storage_usage"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "vm_service.0.content_libraries.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "vm_service.0.vm_classes.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "storage_policies.#"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereNamespaceConfigBasic() string {
	return `
data vsphere_namespace "namespace" {
  name = "test-acc-namespace"
}
`
}
