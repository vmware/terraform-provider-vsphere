// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccDataSourceVSphereContentLibraryItems_nameAndNameRegexConflict(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceVSphereContentLibraryItemsNameAndRegexConfig(),
				ExpectError: regexp.MustCompile(`"name_regex": conflicts with name`),
			},
		},
	})
}

func TestAccDataSourceVSphereContentLibraryItems_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemsConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.items", "items.#", "1",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.items", "items.0.name", "TinyVM",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.items", "items.0.type", "ova",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereContentLibraryItems_allItems(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemsAllConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.vsphere_content_library_items.all", "items.#",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereContentLibraryItems_nameRegex(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemsRegexConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.regex", "items.#", "1",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.regex", "items.0.name", "TinyVM",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereContentLibraryItems_sortByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemsSortByNameConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.sorted_asc", "items.#", "2",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.sorted_asc", "items.0.name", "AlphaVM",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.sorted_asc", "items.1.name", "ZetaVM",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.sorted_desc", "items.0.name", "ZetaVM",
					),
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.sorted_desc", "items.1.name", "AlphaVM",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereContentLibraryItems_limit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemsLimitConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vsphere_content_library_items.limited", "items.#", "1",
					),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereContentLibraryItems_timestamps(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereContentLibraryItemsAllConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(
						"data.vsphere_content_library_items.all", "items.0.creation_time",
					),
					resource.TestCheckResourceAttrSet(
						"data.vsphere_content_library_items.all", "items.0.last_modified_time",
					),
				),
			},
		},
	})
}

func testAccDataSourceVSphereContentLibraryItemsConfig() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_items_test"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name       = "TinyVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

data "vsphere_content_library_items" "items" {
  library_id = vsphere_content_library.library.id
  name       = vsphere_content_library_item.item.name
  type       = "ova"

  depends_on = [vsphere_content_library_item.item]
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
		testhelper.TestOva,
	)
}

func testAccDataSourceVSphereContentLibraryItemsAllConfig() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_items_all_test"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name       = "TinyVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

data "vsphere_content_library_items" "all" {
  library_id = vsphere_content_library.library.id

  depends_on = [vsphere_content_library_item.item]
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
		testhelper.TestOva,
	)
}

func testAccDataSourceVSphereContentLibraryItemsRegexConfig() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_items_regex_test"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name       = "TinyVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

data "vsphere_content_library_items" "regex" {
  library_id = vsphere_content_library.library.id
  name_regex = "^Tiny"

  depends_on = [vsphere_content_library_item.item]
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
		testhelper.TestOva,
	)
}

func testAccDataSourceVSphereContentLibraryItemsSortByNameConfig() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_items_sort_test"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "alpha" {
  name       = "AlphaVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

resource "vsphere_content_library_item" "zeta" {
  name       = "ZetaVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

data "vsphere_content_library_items" "sorted_asc" {
  library_id = vsphere_content_library.library.id
  sort_by    = "name"
  sort_order = "asc"

  depends_on = [vsphere_content_library_item.alpha, vsphere_content_library_item.zeta]
}

data "vsphere_content_library_items" "sorted_desc" {
  library_id = vsphere_content_library.library.id
  sort_by    = "name"
  sort_order = "desc"

  depends_on = [vsphere_content_library_item.alpha, vsphere_content_library_item.zeta]
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
		testhelper.TestOva,
	)
}

func testAccDataSourceVSphereContentLibraryItemsLimitConfig() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "ContentLibrary_items_limit_test"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "alpha" {
  name       = "AlphaVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

resource "vsphere_content_library_item" "zeta" {
  name       = "ZetaVM"
  library_id = vsphere_content_library.library.id
  type       = "ova"
  file_url   = var.file
}

data "vsphere_content_library_items" "limited" {
  library_id = vsphere_content_library.library.id
  sort_by    = "name"
  sort_order = "asc"
  limit      = 1

  depends_on = [vsphere_content_library_item.alpha, vsphere_content_library_item.zeta]
}
`, testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1()),
		testhelper.TestOva,
	)
}

func testAccDataSourceVSphereContentLibraryItemsNameAndRegexConfig() string {
	return `
data "vsphere_content_library_items" "conflict" {
  library_id = "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  name       = "TinyVM"
  name_regex = "^Tiny"
}
`
}
