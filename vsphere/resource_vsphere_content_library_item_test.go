// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereContentLibraryItem_localOva(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryItemCheckExists(false),
		Steps: []resource.TestStep{
			{
				PreConfig: testAccResourceVSphereContentLibraryItemGetOva,
				Config:    testaccresourcevspherecontentlibraryitemconfigLocalova(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "description", regexp.MustCompile("TestAcc Description"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "name", regexp.MustCompile("testacc-item"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "type", regexp.MustCompile("ova"),
					),
					testAccResourceVSphereContentLibraryItemDescription(regexp.MustCompile("TestAcc Description")),
					testAccResourceVSphereContentLibraryItemName(regexp.MustCompile("testacc-item")),
					testAccResourceVSphereContentLibraryItemType(regexp.MustCompile("ova")),
					testAccResourceVSphereContentLibraryItemDestroyFile("./testdata/test.ova"),
				),
			},
		},
	})
}

func TestAccResourceVSphereContentLibraryItem_remoteOvf(t *testing.T) {
	// Disabled because it requires vCenter to trust the remote server hosting the OVF
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryItemCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherecontentlibraryitemconfigRemoteovf(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "description", regexp.MustCompile("TestAcc Description"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "name", regexp.MustCompile("testacc-item"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "type", regexp.MustCompile("ovf"),
					),
					testAccResourceVSphereContentLibraryItemDescription(regexp.MustCompile("TestAcc Description")),
					testAccResourceVSphereContentLibraryItemName(regexp.MustCompile("testacc-item")),
					testAccResourceVSphereContentLibraryItemType(regexp.MustCompile("ovf")),
				),
			},
			{
				ResourceName:      "vsphere_content_library_item.item",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"file_url",
				},
				Config: testaccresourcevspherecontentlibraryitemconfigRemoteovf(),
			},
		},
	})
}

func TestAccResourceVSphereContentLibraryItem_remoteOva(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereContentLibraryItemCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testaccresourcevspherecontentlibraryitemconfigRemoteova(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "id", regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "description", regexp.MustCompile("TestAcc Description"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "name", regexp.MustCompile("testacc-item"),
					),
					resource.TestMatchResourceAttr(
						"vsphere_content_library_item.item", "type", regexp.MustCompile("ovf"),
					),
					testAccResourceVSphereContentLibraryItemDescription(regexp.MustCompile("TestAcc Description")),
					testAccResourceVSphereContentLibraryItemName(regexp.MustCompile("testacc-item")),
					testAccResourceVSphereContentLibraryItemType(regexp.MustCompile("ovf")),
				),
			},
			{
				ResourceName:      "vsphere_content_library_item.item",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"file_url",
				},
				Config: testaccresourcevspherecontentlibraryitemconfigRemoteova(),
			},
		},
	})
}

func testAccResourceVSphereContentLibraryItemGetOva() {
	_ = testAccResourceVSphereContentLibraryItemGetFile(testhelper.TestOva, "./testdata/test.ova")
}

func testAccResourceVSphereContentLibraryItemGetFile(url, file string) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	out, err := os.Create(file)
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	_, err = io.Copy(out, resp.Body)
	return err
}

func testAccResourceVSphereContentLibraryItemDestroyFile(file string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		_ = os.Remove(file)
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemDescription(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			return err
		}
		if !expected.MatchString(*library.Description) {
			return fmt.Errorf("content library item description does not match. expected: %s, got %v", expected.String(), library.Description)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemName(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Name) {
			return fmt.Errorf("content library item name does not match. expected: %s, got %s", expected.String(), library.Name)
		}
		return nil
	}
}

func testAccResourceVSphereContentLibraryItemType(expected *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		library, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			return err
		}
		if !expected.MatchString(library.Type) {
			return fmt.Errorf("content library item type does not match. expected: %s, got %s", expected.String(), library.Type)
		}
		return nil
	}
}

func testaccresourcevspherecontentlibraryitemconfigLocalova() string {
	return fmt.Sprintf(`
%s

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "testacc-item"
  description = "TestAcc Description"
  library_id  = vsphere_content_library.library.id
  type        = "ova"
  file_url    = "./testdata/test.ova"
}
`, testaccresourcevspherecontentlibraryitemconfigBase(),
	)
}

func testaccresourcevspherecontentlibraryitemconfigRemoteovf() string {
	return fmt.Sprintf(`
%s

variable "file" {
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "testacc-item"
  description = "TestAcc Description"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}
`, testaccresourcevspherecontentlibraryitemconfigBase(),
		testhelper.TestOvf,
	)
}

func testaccresourcevspherecontentlibraryitemconfigRemoteova() string {
	return fmt.Sprintf(`
%s

variable "file" {
  default = "%s"
}

resource "vsphere_content_library" "library" {
  name            = "testacc_content_library"
  storage_backing = [data.vsphere_datastore.rootds1.id]
  description     = "Library Description"
}

resource "vsphere_content_library_item" "item" {
  name        = "testacc-item"
  description = "TestAcc Description"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}
`, testaccresourcevspherecontentlibraryitemconfigBase(),
		testhelper.TestOva,
	)
}

func testaccresourcevspherecontentlibraryitemconfigBase() string {
	return testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1())
}

func testAccResourceVSphereContentLibraryItemCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetContentLibraryItem(s, "item")
		if err != nil {
			missingState, _ := regexp.MatchString("not found in state", err.Error())
			missingVSphere, _ := regexp.MatchString("404 Not Found", err.Error())
			if missingState && !expected || missingVSphere && !expected {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected Content Library item to be missing")
		}
		return nil
	}
}
