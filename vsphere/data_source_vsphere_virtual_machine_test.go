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

func TestAccDataSourceVSphereVirtualMachine_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVirtualMachineConfig(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_virtual_machine.vm",
						"id",
						regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "scsi_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "memory_reservation_locked_to_max"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_socket"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_numa_node"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "firmware"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "hardware_version"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.eagerly_scrub"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.thin_provisioned"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.unit_number"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interface_types.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.adapter_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_limit"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_reservation"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_level"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_count"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.mac_address"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.network_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "instance_uuid"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVirtualMachine_noDatacenterAndAbsolutePath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVirtualMachineConfigAbsolutePath(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_virtual_machine.vm",
						"id",
						regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "scsi_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_socket"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_numa_node"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "firmware"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "hardware_version"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.eagerly_scrub"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.thin_provisioned"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.unit_number"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interface_types.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.adapter_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_limit"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_reservation"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_level"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_count"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.mac_address"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.network_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "instance_uuid"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVirtualMachine_uuid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVirtualMachineConfigUUID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_virtual_machine.vm",
						"id",
						regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "scsi_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_socket"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_numa_node"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "firmware"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "hardware_version"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.eagerly_scrub"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.thin_provisioned"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.unit_number"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interface_types.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.adapter_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_limit"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_reservation"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_level"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_count"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.mac_address"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.network_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "uuid"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "instance_uuid"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVirtualMachine_moid(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereVirtualMachineConfigMOID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.vsphere_virtual_machine.vm",
						"id",
						regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "scsi_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_socket"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_numa_node"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "firmware"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "hardware_version"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.size"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.eagerly_scrub"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.thin_provisioned"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.unit_number"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.label"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interface_types.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.#"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.adapter_type"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_limit"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_reservation"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_level"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_count"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.mac_address"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.network_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "instance_uuid"),
					resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "moid"),
				),
			},
		},
	})
}

func TestAccDataSourceVSphereVirtualMachine_nameAndFolder(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{{
			Config: testAccDataSourceVirtualMachineFolder(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestMatchResourceAttr(
					"data.vsphere_virtual_machine.vm",
					"id",
					regexp.MustCompile("^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$")),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "guest_id"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "scsi_type"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "memory"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cpus"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_socket"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "num_cores_per_numa_node"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "firmware"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "hardware_version"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.#"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.size"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.eagerly_scrub"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.thin_provisioned"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.unit_number"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "disks.0.label"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interface_types.#"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.#"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.adapter_type"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_limit"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_reservation"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_level"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.bandwidth_share_count"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.mac_address"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "network_interfaces.0.network_id"),
				resource.TestCheckResourceAttrSet("data.vsphere_virtual_machine.vm", "instance_uuid")),
		}},
	})
}

func testAccDataSourceVSphereVirtualMachineConfigUUID() string {
	return fmt.Sprintf(`
%s

data "vsphere_virtual_machine" "vm" {
  uuid = vsphere_virtual_machine.srcvm.uuid
}
`,
		testAccDataSourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccDataSourceVSphereVirtualMachineConfigMOID() string {
	return fmt.Sprintf(`
%s

data "vsphere_virtual_machine" "vm" {
  moid = vsphere_virtual_machine.srcvm.moid
}
`,
		testAccDataSourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccDataSourceVSphereVirtualMachineConfig() string {
	return fmt.Sprintf(`
%s

data "vsphere_virtual_machine" "vm" {
  name          = vsphere_virtual_machine.srcvm.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}
`,
		testAccDataSourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccDataSourceVSphereVirtualMachineConfigAbsolutePath() string {
	return fmt.Sprintf(`
%s

data "vsphere_virtual_machine" "vm" {
  name = "/${data.vsphere_datacenter.rootdc1.name}/vm/${vsphere_virtual_machine.srcvm.name}"
}
`,
		testAccDataSourceVSphereVirtualMachineConfigBase(),
	)
}

func testAccDataSourceVirtualMachineFolder() string {
	return fmt.Sprintf(`
%s

resource "vsphere_folder" "new_vm_folder" {
  path          = "new-vm-folder"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  type          = "vm"
}

resource "vsphere_virtual_machine" "srcvm" {
  name             = "acc-test-vm"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
  datastore_id     = data.vsphere_datastore.rootds1.id
  folder           = vsphere_folder.new_vm_folder.path
  num_cpus         = 1
  memory           = 1024
  guest_id         = "otherLinux64Guest"
  network_interface {
    network_id = data.vsphere_network.network1.id
  }
  disk {
    label = "disk0"
    size  = 1
    io_reservation = 1
  }
  wait_for_guest_ip_timeout  = 0
  wait_for_guest_net_timeout = 0
}

data vsphere_virtual_machine "vm" {
  name          = vsphere_virtual_machine.srcvm.name
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  folder        = vsphere_folder.new_vm_folder.path
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccDataSourceVSphereVirtualMachineConfigBase() string {
	return fmt.Sprintf(`
%s

resource "vsphere_virtual_machine" "srcvm" {
  name             = "acc-test-vm"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
  datastore_id     = data.vsphere_datastore.rootds1.id
  num_cpus         = 1
  memory           = 1024
  guest_id         = "otherLinux64Guest"
  network_interface {
    network_id = data.vsphere_network.network1.id
  }
  disk {
    label = "disk0"
    size  = 1
    io_reservation = 1
  }
  wait_for_guest_ip_timeout  = 0
  wait_for_guest_net_timeout = 0
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootPortGroup1()),
	)
}
