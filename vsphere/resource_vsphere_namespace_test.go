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

func TestAccResourceVSphereNamespace_basic(t *testing.T) {
	// cannot run on the standard testbed
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNamespaceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "storage_policies.#"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "vm_service.0.vm_classes.#"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "vm_service.0.content_libraries.#"),
				),
			},
		},
	})
}

func testAccResourceVSphereNamespaceConfigBasic() string {
	return fmt.Sprintf(`
data vsphere_storage_policy image_policy {
  name = "%s"
}

data vsphere_content_library subscribed_lib {
  name = "%s"
}

resource vsphere_namespace "namespace" {
  name       = "test-acc-namespace"
  supervisor = "ff69d7fb-4ad4-44a2-8d91-8b3bede80eaa"

  vm_service {
    content_libraries = [data.vsphere_content_library.subscribed_lib.id]
    vm_classes = ["%s"]
  }

  storage_policies = [data.vsphere_storage_policy.image_policy.id]
}
`,
		os.Getenv("TF_VAR_STORAGE_POLICY"),
		os.Getenv("TF_VAR_CONTENT_LIBRARY"),
		os.Getenv("TF_VAR_VM_CLASS"))
}
