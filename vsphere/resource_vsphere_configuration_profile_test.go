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

func TestAccResourceVSphereConfigProfile(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereConfigProfileConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_configuration_profile.profile1", "configuration"),
					resource.TestCheckResourceAttrSet("vsphere_configuration_profile.profile1", "schema"),
					resource.TestCheckResourceAttrSet("vsphere_configuration_profile.profile2", "configuration"),
					resource.TestCheckResourceAttrSet("vsphere_configuration_profile.profile2", "schema"),
				),
			},
		},
	})
}

func testAccResourceVSphereConfigProfileConfig() string {
	return fmt.Sprintf(`
%s

locals {
  esx4_hostname = "%s"
  esx4_password = "%s"
}

data "vsphere_host_thumbprint" "thumbprint" {
  address  = local.esx4_hostname
  insecure = true
}

resource "vsphere_host" "host4" {
  hostname   = local.esx4_hostname
  username   = "root"
  password   = local.esx4_password
  thumbprint = data.vsphere_host_thumbprint.thumbprint.id

  datacenter = data.vsphere_datacenter.rootdc1.id

  lifecycle {
    ignore_changes = ["services", "cluster"]
  }
}

resource "vsphere_compute_cluster" "cluster1" {
  name                      = "cluster1"
  datacenter_id             = data.vsphere_datacenter.rootdc1.id
  host_system_ids           = [vsphere_host.host4.id]
  force_evacuate_on_destroy = true
}

resource "vsphere_compute_cluster" "cluster2" {
  name          = "cluster2"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

resource "vsphere_configuration_profile" "profile1" {
  reference_host_id = vsphere_host.host4.id
  cluster_id        = vsphere_compute_cluster.cluster1.id
}

resource "vsphere_configuration_profile" "profile2" {
  cluster_id    = vsphere_compute_cluster.cluster2.id
  configuration = vsphere_configuration_profile.profile1.configuration
}
`,
		testhelper.ConfigDataRootDC1(),
		os.Getenv("TF_VAR_VSPHERE_ESXI4"),
		os.Getenv("TF_VAR_VSPHERE_ESXI4_PASSWORD"))
}
