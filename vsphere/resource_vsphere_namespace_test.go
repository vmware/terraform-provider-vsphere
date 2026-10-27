// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereNamespace_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR", "TF_VAR_STORAGE_POLICY"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNamespaceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
				),
			},
		},
	})
}

func TestAccResourceVSphereNamespace_storagePolicy(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR", "TF_VAR_STORAGE_POLICY"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNamespaceConfigStoragePolicy(0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
					resource.TestCheckResourceAttr("vsphere_namespace.namespace", "storage_policy.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("vsphere_namespace.namespace", "storage_policy.*.policy", "data.vsphere_storage_policy.image_policy", "id"),
					resource.TestCheckNoResourceAttr("vsphere_namespace.namespace", "storage_policy.0.limit"),
				),
			},
			{
				Config: testAccResourceVSphereNamespaceConfigStoragePolicy(1024),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vsphere_namespace.namespace", "storage_policy.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vsphere_namespace.namespace", "storage_policy.*", map[string]string{"limit": "1024"}),
				),
			},
			{
				Config: testAccResourceVSphereNamespaceConfigStoragePolicy(0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vsphere_namespace.namespace", "storage_policy.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("vsphere_namespace.namespace", "storage_policy.*.policy", "data.vsphere_storage_policy.image_policy", "id"),
					resource.TestCheckNoResourceAttr("vsphere_namespace.namespace", "storage_policy.0.limit"),
				),
			},
		},
	})
}

func TestAccResourceVSphereNamespace_vmService(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR", "TF_VAR_STORAGE_POLICY"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNamespaceConfigVMService(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "storage_policy.#"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "vm_service.0.vm_classes.#"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "vm_service.0.content_libraries.#"),
				),
			},
		},
	})
}

func TestAccResourceVSphereNamespace_update(t *testing.T) {
	// consecutively apply some of the creation configurations to verify the work via the update path too
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_SUPERVISOR", "TF_VAR_STORAGE_POLICY"})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereNamespaceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
				),
			},
			{
				Config: testAccResourceVSphereNamespaceConfigStoragePolicy(0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "storage_policy.#"),
				),
			},
			{
				Config: testAccResourceVSphereNamespaceConfigVMService(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "id"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "supervisor"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "storage_policy.#"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "vm_service.0.vm_classes.#"),
					resource.TestCheckResourceAttrSet("vsphere_namespace.namespace", "vm_service.0.content_libraries.#"),
				),
			},
		},
	})
}

func testAccResourceVSphereNamespaceConfigBasic() string {
	return fmt.Sprintf(`
%s

resource vsphere_namespace "namespace" {
  name       = "test-acc-namespace"
  supervisor = data.vsphere_supervisor.supervisor.id
}
`, testAccDataSourceVSphereSupervisorConfigName())
}

func testAccResourceVSphereNamespaceConfigStoragePolicy(limit int) string {
	limitAttr := ""
	if limit > 0 {
		limitAttr = fmt.Sprintf("limit  = %d", limit)
	}
	return fmt.Sprintf(`
%s

data vsphere_storage_policy image_policy {
  name = "%s"
}

resource vsphere_namespace "namespace" {
  name       = "test-acc-namespace"
  supervisor = data.vsphere_supervisor.supervisor.id

  storage_policy {
    policy = data.vsphere_storage_policy.image_policy.id
    %s
  }
}
`,
		testAccDataSourceVSphereSupervisorConfigName(),
		os.Getenv("TF_VAR_STORAGE_POLICY"),
		limitAttr)
}

func testAccResourceVSphereNamespaceConfigVMService() string {
	return fmt.Sprintf(`
%s

data vsphere_storage_policy image_policy {
  name = "%s"
}

resource "vsphere_virtual_machine_class" "vm_class" {
  name               = "acc-test-vmclass"
  cpus               = 4
  memory             = 4096
  memory_reservation = 100
}

resource "vsphere_content_library" "library_published" {
  name            = "testacc_published"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  publication {
    authentication_method = "BASIC"
    username              = "vcsp"
    password              = "Password123!"
    published             = true
  }
}

resource "vsphere_content_library" "library_subscribed" {
  name            = "testacc_subscribed"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  subscription {
    authentication_method = "BASIC"
    username              = "vcsp"
    password              = "Password123!"
    subscription_url      = vsphere_content_library.library_published.publication.0.publish_url
  }
}

resource vsphere_namespace "namespace" {
  name       = "test-acc-namespace"
  supervisor = data.vsphere_supervisor.supervisor.id

  vm_service {
    content_libraries = [vsphere_content_library.library_subscribed.id]
    vm_classes = [vsphere_virtual_machine_class.vm_class.id]
  }

  storage_policy {
    policy = data.vsphere_storage_policy.image_policy.id
  }
}
`,
		testhelper.CombineConfigs(
			testAccResourceVSphereNamespaceConfigBase(),
			testAccDataSourceVSphereSupervisorConfigName()),
		os.Getenv("TF_VAR_STORAGE_POLICY"))
}

func testAccResourceVSphereNamespaceConfigBase() string {
	return testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1())
}
