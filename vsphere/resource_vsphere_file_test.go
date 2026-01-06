// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

// TestAccResourceVSphereFile_basic verifies the basic functionality of the resource.
func TestAccResourceVSphereFile_basic(t *testing.T) {
	testFileData := []byte("test file data")
	testFile := "/tmp/tf_test.txt"
	err := os.WriteFile(testFile, testFileData, 0600)
	if err != nil {
		t.Errorf("error %s", err)
		return
	}

	datacenter := os.Getenv("TF_VAR_VSPHERE_DATACENTER")
	datastore := os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")
	testMethod := "basic"
	resourceName := "vsphere_file." + testMethod
	destinationFile := "tf_file_test.txt"
	sourceFile := testFile

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileConfig,
					testMethod,
					datacenter,
					datastore,
					sourceFile,
					destinationFile,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists(resourceName, destinationFile, true),
					resource.TestCheckResourceAttr(resourceName, "destination_file", destinationFile),
				),
			},
		},
	})
	_ = os.Remove(testFile)
}

// TestAccResourceVSphereFile_uploadWithCreateDirectories verifies uploading files with nested directories.
// creation.
func TestAccResourceVSphereFile_uploadWithCreateDirectories(t *testing.T) {
	testFileData := []byte("test file data")
	testFile := "/tmp/tf_test.txt"
	err := os.WriteFile(testFile, testFileData, 0600)
	if err != nil {
		t.Errorf("error %s", err)
		return
	}

	datacenter := os.Getenv("TF_VAR_VSPHERE_DATACENTER")
	datastore := os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")
	fileInNestedFolder := "fileInNestedFolder"
	fileInNestedFolderResourceName := "vsphere_file." + fileInNestedFolder
	destinationFileInNestedFolderPath := "/folder1/folder2/folder3/tf_file_test.txt"
	fileInRootFolder := "fileInRootFolder"
	fileInRootFolderResourceName := "vsphere_file." + fileInRootFolder
	destinationFileInRootFolderPath := "tf_file_test.txt"
	sourceFile := testFile

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileCreateFolderConfig,
					fileInNestedFolder,
					datacenter,
					datastore,
					sourceFile,
					destinationFileInNestedFolderPath,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists(
						fileInNestedFolderResourceName, destinationFileInNestedFolderPath, true),
					resource.TestCheckResourceAttr(
						fileInNestedFolderResourceName, "destination_file", destinationFileInNestedFolderPath),
				),
			},
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileCreateFolderConfig,
					fileInRootFolder,
					datacenter,
					datastore,
					sourceFile,
					destinationFileInRootFolderPath,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists(
						fileInRootFolderResourceName, destinationFileInRootFolderPath, true),
					resource.TestCheckResourceAttr(
						fileInRootFolderResourceName, "destination_file", destinationFileInRootFolderPath),
				),
			},
		},
	})
	_ = os.Remove(testFile)
}

// TestAccResourceVSphereFile_basicUploadAndCopy verifies uploading and copying files.
func TestAccResourceVSphereFile_basicUploadAndCopy(t *testing.T) {
	testAccSkipUnstable(t)
	testFileData := []byte("test file data")
	sourceFile := "/tmp/tf_test.txt"
	uploadResourceName := "myfileupload"
	copyResourceName := "myfilecopy"
	sourceDatacenter := os.Getenv("TF_VAR_VSPHERE_DATACENTER")
	datacenter := sourceDatacenter
	sourceDatastore := os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")
	datastore := sourceDatastore
	destinationFile := "tf_file_test.txt"
	sourceFileCopy := "${vsphere_file." + uploadResourceName + ".destination_file}"
	destinationFileCopy := "tf_file_test_copy.txt"

	err := os.WriteFile(sourceFile, testFileData, 0600)
	if err != nil {
		t.Errorf("error %s", err)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_NFS_DS_NAME"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileCopyConfig,
					uploadResourceName,
					datacenter,
					datastore,
					sourceFile,
					destinationFile,
					copyResourceName,
					datacenter,
					datacenter,
					datastore,
					datastore,
					sourceFileCopy,
					destinationFileCopy,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists("vsphere_file."+uploadResourceName, destinationFile, true),
					testAccCheckVSphereFileExists("vsphere_file."+copyResourceName, destinationFileCopy, true),
					resource.TestCheckResourceAttr("vsphere_file."+uploadResourceName, "destination_file", destinationFile),
					resource.TestCheckResourceAttr("vsphere_file."+copyResourceName, "destination_file", destinationFileCopy),
				),
			},
		},
	})
	_ = os.Remove(sourceFile)
}

// TestAccResourceVSphereFile_renamePostCreation verifies the renaming of a resource during creation and update phases.
func TestAccResourceVSphereFile_renamePostCreation(t *testing.T) {
	testAccSkipUnstable(t)
	testFileData := []byte("test file data")
	testFile := "/tmp/tf_test.txt"
	err := os.WriteFile(testFile, testFileData, 0600)
	if err != nil {
		t.Errorf("error %s", err)
		return
	}

	datacenter := os.Getenv("TF_VAR_VSPHERE_DATACENTER")
	datastore := os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")
	testMethod := "create_upgrade"
	resourceName := "vsphere_file." + testMethod
	destinationFile := "tf_test_file.txt"
	destinationFileMoved := "tf_test_file_moved.txt"
	sourceFile := testFile

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_NFS_DS_NAME"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileConfig,
					testMethod,
					datacenter,
					datastore,
					sourceFile,
					destinationFile,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists(resourceName, destinationFile, true),
					testAccCheckVSphereFileExists(resourceName, destinationFileMoved, false),
					resource.TestCheckResourceAttr(resourceName, "destination_file", destinationFile),
				),
			},
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileConfig,
					testMethod,
					datacenter,
					datastore,
					sourceFile,
					destinationFileMoved,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists(resourceName, destinationFile, false),
					testAccCheckVSphereFileExists(resourceName, destinationFileMoved, true),
					resource.TestCheckResourceAttr(resourceName, "destination_file", destinationFileMoved),
				),
			},
		},
	})
	_ = os.Remove(testFile)
}

// TestAccResourceVSphereFile_uploadAndCopyAndUpdate verifies uploading, copying, and updating files.
func TestAccResourceVSphereFile_uploadAndCopyAndUpdate(t *testing.T) {
	testAccSkipUnstable(t)
	testFileData := []byte("test file data")
	sourceFile := "/tmp/tf_test.txt"
	uploadResourceName := "myfileupload"
	copyResourceName := "myfilecopy"
	sourceDatacenter := os.Getenv("TF_VAR_VSPHERE_DATACENTER")
	datacenter := sourceDatacenter
	sourceDatastore := os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME")
	datastore := sourceDatastore
	destinationFile := "tf_file_test.txt"
	sourceFileCopy := "${vsphere_file." + uploadResourceName + ".destination_file}"
	destinationFileCopy := "tf_file_test_copy.txt"
	destinationFileMoved := "tf_test_file_moved.txt"

	err := os.WriteFile(sourceFile, testFileData, 0600)
	if err != nil {
		t.Errorf("error %s", err)
		return
	}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"TF_VAR_VSPHERE_DATACENTER", "TF_VAR_VSPHERE_NFS_DS_NAME"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVSphereFileDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileCopyConfig,
					uploadResourceName,
					datacenter,
					datastore,
					sourceFile,
					destinationFile,
					copyResourceName,
					datacenter,
					datacenter,
					datastore,
					datastore,
					sourceFileCopy,
					destinationFileCopy,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists("vsphere_file."+uploadResourceName, destinationFile, true),
					testAccCheckVSphereFileExists("vsphere_file."+copyResourceName, destinationFileCopy, true),
					resource.TestCheckResourceAttr("vsphere_file."+uploadResourceName, "destination_file", destinationFile),
					resource.TestCheckResourceAttr("vsphere_file."+copyResourceName, "destination_file", destinationFileCopy),
				),
			},
			{
				Config: fmt.Sprintf(
					testAccCheckVSphereFileCopyConfig,
					uploadResourceName,
					datacenter,
					datastore,
					sourceFile,
					destinationFile,
					copyResourceName,
					datacenter,
					datacenter,
					datastore,
					datastore,
					sourceFileCopy,
					destinationFileMoved,
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVSphereFileExists("vsphere_file."+uploadResourceName, destinationFile, true),
					testAccCheckVSphereFileExists("vsphere_file."+copyResourceName, destinationFileCopy, false),
					testAccCheckVSphereFileExists("vsphere_file."+copyResourceName, destinationFileMoved, true),
					resource.TestCheckResourceAttr("vsphere_file."+uploadResourceName, "destination_file", destinationFile),
					resource.TestCheckResourceAttr("vsphere_file."+copyResourceName, "destination_file", destinationFileMoved),
				),
			},
		},
	})
	_ = os.Remove(sourceFile)
}

// testAccCheckVSphereFileDestroy verifies deleting files.
func testAccCheckVSphereFileDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*Client).vimClient
	finder := find.NewFinder(client.Client, true)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_file" {
			continue
		}

		dc, err := finder.Datacenter(context.TODO(), rs.Primary.Attributes["datacenter"])
		if err != nil {
			return fmt.Errorf("error %s", err)
		}

		finder = finder.SetDatacenter(dc)

		ds, err := getDatastore(finder, rs.Primary.Attributes["datastore"])
		if err != nil {
			return fmt.Errorf("error %s", err)
		}

		_, err = ds.Stat(context.TODO(), rs.Primary.Attributes["destination_file"])
		if err != nil {
			var notFoundError object.DatastoreNoSuchFileError
			switch {
			case errors.As(err, &notFoundError):
				return nil
			default:
				return err
			}
		} else {
			return fmt.Errorf("file %s still exists", rs.Primary.Attributes["destination_file"])
		}
	}

	return nil
}

// testAccCheckVSphereFileExists verifies the existence or non-existence of a specified file.
func testAccCheckVSphereFileExists(n string, df string, exists bool) resource.TestCheckFunc {
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

		dc, err := finder.Datacenter(context.TODO(), rs.Primary.Attributes["datacenter"])
		if err != nil {
			return fmt.Errorf("error %s", err)
		}
		finder = finder.SetDatacenter(dc)

		ds, err := getDatastore(finder, rs.Primary.Attributes["datastore"])
		if err != nil {
			return fmt.Errorf("error %s", err)
		}

		_, err = ds.Stat(context.TODO(), df)
		if err != nil {
			var notFoundError *object.DatastoreNoSuchFileError
			switch {
			case errors.As(err, &notFoundError):
				if exists {
					return fmt.Errorf("file does not exist: %s", notFoundError.Error())
				}
				return nil
			default:
				return err
			}
		}
		return nil
	}
}

// testAccCheckVSphereFileConfig defines a configuration for creating or managing the resource.
const testAccCheckVSphereFileConfig = `
resource "vsphere_file" "%s" {
  datacenter       = "%s"
  datastore        = "%s"
  source_file      = "%s"
  destination_file = "%s"
}
`

// testAccCheckVSphereFileCopyConfig defines a configuration for uploading and copying the resource.
const testAccCheckVSphereFileCopyConfig = `
resource "vsphere_file" "%s" {
  datacenter       = "%s"
  datastore        = "%s"
  source_file      = "%s"
  destination_file = "%s"
}
resource "vsphere_file" "%s" {
  source_datacenter = "%s"
  datacenter        = "%s"
  source_datastore  = "%s"
  datastore         = "%s"
  source_file       = "%s"
  destination_file  = "%s"
}
`

// testAccCheckVSphereFileCreateFolderConfig defines a configuration for testing file uploads with directory creation.
const testAccCheckVSphereFileCreateFolderConfig = `
resource "vsphere_file" "%s" {
  datacenter         = "%s"
  datastore          = "%s"
  source_file        = "%s"
  destination_file   = "%s"
  create_directories = true
}
`
