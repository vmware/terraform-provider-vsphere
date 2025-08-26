// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"errors"
	"fmt"
	"path"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/find"
)

const testAccCheckVSphereDatacenterResourceName = "vsphere_datacenter.testDC"

func testAccCheckVSphereDatacenterConfig(name string) string {
	return fmt.Sprintf(`
resource "vsphere_datacenter" "testDC" {
  name = "%s"
}
`, name)
}

func testAccCheckVSphereDatacenterConfigSubfolder(name, folder string) string {
	return fmt.Sprintf(`
resource "vsphere_folder" "folder" {
  path = "%s"
  type = "datacenter"
}

resource "vsphere_datacenter" "testDC" {
  name   = "%s"
  folder = vsphere_folder.folder.path
}
`, folder, name)
}

const testAccCheckVSphereDatacenterConfigTags = `
resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datacenter",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = vsphere_tag_category.testacc-category.id
}

resource "vsphere_datacenter" "testDC" {
  name = "testDC"
  tags = [vsphere_tag.testacc-tag.id]
}
`

const testAccCheckVSphereDatacenterConfigMultiTags = `
variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "Datacenter",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = vsphere_tag_category.testacc-category.id
}

resource "vsphere_tag" "testacc-tags-alt" {
  count       = length(var.extra_tags)
  name        = var.extra_tags[count.index]
  category_id = vsphere_tag_category.testacc-category.id
}

resource "vsphere_datacenter" "testDC" {
  name = "testDC"
  tags = vsphere_tag.testacc-tags-alt.*.id
}
`

const testAccCheckVSphereDatacenterConfigCustomAttributes = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Datacenter"
}

locals {
  dc_attrs = {
    vsphere_custom_attribute.testacc-attribute.id = "value"
  }
}

resource "vsphere_datacenter" "testDC" {
  name              = "testDC"
  custom_attributes = local.dc_attrs
}
`
const testAccCheckVSphereDatacenterConfigMultiCustomAttributes = `
resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "Datacenter"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "Datacenter"
}

locals {
  dc_attrs = {
    vsphere_custom_attribute.testacc-attribute.id   = "value"
    vsphere_custom_attribute.testacc-attribute-2.id = "value-2"
  }
}

resource "vsphere_datacenter" "testDC" {
  name              = "testDC"
  custom_attributes = local.dc_attrs
}
`

// Create a datacenter on the root folder
func TestAccResourceVSphereDatacenter_createOnRootFolder(t *testing.T) {
	name := "testDC"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfig(name),
				Check:  resource.ComposeTestCheckFunc(testAccCheckVSphereDatacenterExists(testAccCheckVSphereDatacenterResourceName, true)),
			},
			{
				ResourceName:        testAccCheckVSphereDatacenterResourceName,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "/",
				ImportStateId:       name,
			},
		},
	})
}

// Create a datacenter on a subfolder
func TestAccResourceVSphereDatacenter_createOnSubfolder(t *testing.T) {
	dcFolder := "dc-folder"
	name := "testDC"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigSubfolder(name, dcFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
				),
			},
			{
				ResourceName:        testAccCheckVSphereDatacenterResourceName,
				ImportState:         true,
				ImportStateVerify:   true,
				ImportStateIdPrefix: "/",
				ImportStateId:       path.Join(dcFolder, name),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_singleTag(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_modifyTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccCheckVSphereDatacenterConfigMultiTags,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_singleCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigCustomAttributes,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterHasCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereDatacenter_modifyCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereDatacenterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckVSphereDatacenterConfigCustomAttributes,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterHasCustomAttributes(),
				),
			},
			{
				Config: testAccCheckVSphereDatacenterConfigMultiCustomAttributes,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereDatacenterExists(
						testAccCheckVSphereDatacenterResourceName,
						true,
					),
					testAccResourceVSphereDatacenterHasCustomAttributes(),
				),
			},
		},
	})
}

func testAccCheckVSphereDatacenterDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).vimClient
	finder := find.NewFinder(client.Client, true)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_datacenter" {
			continue
		}

		datacenterPath := rs.Primary.Attributes["name"]
		if _, ok := rs.Primary.Attributes["folder"]; ok {
			datacenterPath + = rs.Primary.Attributes["folder"] + "/" + datacenterPath
		}
		_, err := finder.Datacenter(context.TODO(), datacenterPath)
		if err != nil {
			var notFoundError *find.NotFoundError
			switch {
			case errors.As(err, &notFoundError):
				return nil
			default:
				return err
			}
		} else {
			return fmt.Errorf("datacenter '%s' still exists", datacenterPath)
		}
	}

	return nil
}

func testAccCheckVSphereDatacenterExists(n string, exists bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*Client).vimClient
		finder := find.NewFinder(client.Client, true)

		datacenterPath := rs.Primary.Attributes["name"]
		if _, ok := rs.Primary.Attributes["folder"]; ok {
			datacenterPath = rs.Primary.Attributes["folder"] + "/" + datacenterPath
		}
		_, err := finder.Datacenter(context.TODO(), datacenterPath)
		if err != nil {
			var notFoundError *find.NotFoundError
			switch {
			case errors.As(err, &notFoundError):
				if exists {
					return fmt.Errorf("datacenter does not exist: %s", notFoundError.Error())
				}
				return nil
			default:
				return err
			}
		}
		return nil
	}
}

// testAccResourceVSphereDatacenterCheckTags is a check to ensure that any tags
// that have been created with supplied resource name have been attached to the
// datacenter.
func testAccResourceVSphereDatacenterCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vars, err := testClientVariablesForResource(s, "vsphere_datacenter.testDC")
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}

		finder := find.NewFinder(vars.client.Client, true)

		datacenterPath := vars.resourceAttributes["name"]
		if _, ok := vars.resourceAttributes["folder"]; ok {
			datacenterPath = vars.resourceAttributes["folder"] + "/" + datacenterPath
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()
		dc, err := finder.Datacenter(ctx, datacenterPath)
		if err != nil {
			return err
		}

		return testObjectHasTags(s, tagsClient, dc, tagResName)
	}
}

func testAccResourceVSphereDatacenterHasCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ca, err := testGetDatacenterCustomAttributes(s, "testDC")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_datacenter", "testDC", ca.Entity())
	}
}
