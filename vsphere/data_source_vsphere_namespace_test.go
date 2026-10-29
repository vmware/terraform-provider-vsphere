// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDataSourceVSphereNamespace_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR"})
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
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "storage_policy.#"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereNamespace_accessList(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNamespaceConfigFor(testAccResourceVSphereNamespaceConfigAccess("EDIT")),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_namespace.namespace", "access.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.vsphere_namespace.namespace", "access.*", map[string]string{
						"subject_type": "USER",
						"subject":      "acc-namespace-user",
						"role":         "EDIT",
					}),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereNamespace_zones(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR", "TF_VAR_NAMESPACE_ZONE"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereNamespaceConfigFor(testAccResourceVSphereNamespaceConfigZones()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_namespace.namespace", "zone.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.vsphere_namespace.namespace", "zone.*", map[string]string{
						"name": os.Getenv("TF_VAR_NAMESPACE_ZONE"),
					}),
					resource.TestCheckResourceAttrSet("data.vsphere_namespace.namespace", "zone.0.type"),
				),
			},
		},
	})
}

func testAccDataSourceVSphereNamespaceConfigFor(namespaceConfig string) string {
	return fmt.Sprintf(`
%s

data vsphere_namespace "namespace" {
  name = vsphere_namespace.namespace.name
}
`, namespaceConfig)
}

func testAccDataSourceVSphereNamespaceConfigBasic() string {
	return fmt.Sprintf(`
%s

data vsphere_namespace "namespace" {
  name = vsphere_namespace.namespace.name
}
`, testAccResourceVSphereNamespaceConfigVMService())
}
