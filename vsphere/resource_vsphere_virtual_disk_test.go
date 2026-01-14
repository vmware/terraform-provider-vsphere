// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

func TestAccResourceVSphereVirtualDisk_basic(t *testing.T) {
	rString := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo", false),
		Steps: []resource.TestStep{
			{
				Config: testacccheckvspherevirtualdiskconfigBasic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualDisk_extend(t *testing.T) {
	rString := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo", false),
		Steps: []resource.TestStep{
			{
				Config: testacccheckvspherevirtualdiskconfigBasic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vsphere_virtual_disk.foo", "size", "1"),
				),
			},
			{
				Config: testacccheckvspherevirtualdiskconfigExtended(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vsphere_virtual_disk.foo", "size", "2"),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualDisk_multi(t *testing.T) {
	rString := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.0", false),
			testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.1", false),
			testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.2", false),
		),
		Steps: []resource.TestStep{
			{
				Config: testacccheckvspherevirtuadiskconfigMulti(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.0", true),
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.1", true),
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.2", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualDisk_multiWithParent(t *testing.T) {
	rString := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.0", false),
			testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.1", false),
			testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.2", false),
		),
		Steps: []resource.TestStep{
			{
				Config: testacccheckvspherevirtuadiskconfigMultiwithparent(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.0", true),
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.1", true),
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo.2", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereVirtualDisk_withParent(t *testing.T) {
	rString := acctest.RandString(5)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo", false),
		Steps: []resource.TestStep{
			{
				Config: testacccheckvspherevirtuadiskconfigWithparent(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereVirtualDiskExists("vsphere_virtual_disk.foo", true),
				),
			},
		},
	})
}

func testAccVSphereVirtualDiskExists(name string, expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}

		client := testAccProvider.Meta().(*Client).vimClient
		finder := find.NewFinder(client.Client, true)

		dc, err := finder.Datacenter(context.TODO(), rs.Primary.Attributes["datacenter"])
		if err != nil {
			return err
		}
		finder = finder.SetDatacenter(dc)

		ds, err := finder.Datastore(context.TODO(), rs.Primary.Attributes["datastore"])
		if err != nil {
			if err.Error() == fmt.Sprintf("datastore '%s' not found", rs.Primary.Attributes["datastore"]) {
				return nil
			}
			return err
		}

		_, err = ds.Stat(context.TODO(), rs.Primary.Attributes["vmdk_path"])
		if err != nil {
			if testAccCheckVSphereVirtualDiskIsFileNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return fmt.Errorf("expected virtual disk %s to be missing", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckVSphereVirtualDiskIsFileNotFoundError(err error) bool {
	if strings.HasPrefix(err.Error(), "cannot stat") && strings.HasSuffix(err.Error(), "No such file") {
		return true
	}
	return false
}

func testacccheckvspherevirtualdiskconfigBasic(rName string) string {
	return fmt.Sprintf(`
%s

variable "rstring" {
  default = "%s"
}

resource "vsphere_virtual_disk" "foo" {
  size         = 1
  vmdk_path    = "tfTestDisk-${var.rstring}.vmdk"
  adapter_type = "lsiLogic"
  type         = "thin"
  datacenter   = data.vsphere_datacenter.rootdc1.name
  datastore    = data.vsphere_datastore.rootds1.name
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1()),
		rName,
	)
}

func testacccheckvspherevirtualdiskconfigExtended(rName string) string {
	return fmt.Sprintf(`
%s

variable "rstring" {
  default = "%s"
}

resource "vsphere_virtual_disk" "foo" {
  size         = 2
  vmdk_path    = "tfTestDisk-${var.rstring}.vmdk"
  adapter_type = "lsiLogic"
  type         = "thin"
  datacenter   = data.vsphere_datacenter.rootdc1.name
  datastore    = data.vsphere_datastore.rootds1.name
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigDataRootDS1()),
		rName,
	)
}

func testacccheckvspherevirtuadiskconfigMulti(rName string) string {
	return fmt.Sprintf(`
%s

variable "rstring" {
  default = "%s"
}

data "vsphere_datastore" "ds" {
  name          = data.vsphere_datastore.rootds1.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_virtual_disk" "foo" {
  count        = 3
  size         = 1
  vmdk_path    = "tfTestDisk-${var.rstring}-${count.index}.vmdk"
  adapter_type = "lsiLogic"
  type         = "thin"
  datacenter   = data.vsphere_datacenter.rootdc1.name
  datastore    = data.vsphere_datastore.ds.name
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
		rName,
	)
}

func testacccheckvspherevirtuadiskconfigMultiwithparent(rName string) string {
	return fmt.Sprintf(`
%s

variable "rstring" {
  default = "%s"
}

data "vsphere_datastore" "ds" {
  name          = data.vsphere_datastore.rootds1.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_virtual_disk" "foo" {
  count              = 3
  size               = 1
  vmdk_path          = "tfTestParent/tfTestDisk-${var.rstring}-${count.index}.vmdk"
  adapter_type       = "lsiLogic"
  type               = "thin"
  datacenter         = data.vsphere_datacenter.rootdc1.name
  datastore          = data.vsphere_datastore.ds.name
  create_directories = true
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
		rName,
	)
}

func testacccheckvspherevirtuadiskconfigWithparent(rName string) string {
	return fmt.Sprintf(`
%s

variable "rstring" {
  default = "%s"
}

data "vsphere_datastore" "ds" {
  name          = data.vsphere_datastore.rootds1.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_virtual_disk" "foo" {
  size               = 1
  vmdk_path          = "tfTestParent-${var.rstring}/tfTestDisk-${var.rstring}.vmdk"
  adapter_type       = "lsiLogic"
  type               = "thin"
  datacenter         = data.vsphere_datacenter.rootdc1.name
  datastore          = data.vsphere_datastore.ds.name
  create_directories = true
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootHost1(), testhelper.ConfigDataRootHost2(), testhelper.ConfigDataRootDS1(), testhelper.ConfigDataRootComputeCluster1(), testhelper.ConfigResResourcePool1(), testhelper.ConfigDataRootPortGroup1()),
		rName,
	)
}
