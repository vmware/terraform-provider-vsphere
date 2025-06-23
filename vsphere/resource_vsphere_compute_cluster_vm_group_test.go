// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/structure"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/virtualmachine"
)

func TestAccResourceVSphereComputeClusterVMGroup_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterVMGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterVMGroupConfig(2),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMGroupExists(true),
					testAccResourceVSphereComputeClusterVMGroupMatchMembership(),
				),
			},
			{
				ResourceName:      "vsphere_compute_cluster_vm_group.cluster_vm_group",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cluster, err := testGetComputeClusterFromDataSource(s, "rootcompute_cluster1")
					if err != nil {
						return "", err
					}

					rs, ok := s.RootModule().Resources["vsphere_compute_cluster_vm_group.cluster_vm_group"]
					if !ok {
						return "", errors.New("no resource at address vsphere_compute_cluster_vm_group.cluster_vm_group")
					}
					name, ok := rs.Primary.Attributes["name"]
					if !ok {
						return "", errors.New("vsphere_compute_cluster_vm_group.cluster_vm_group has no name attribute")
					}

					m := make(map[string]string)
					m["compute_cluster_path"] = cluster.InventoryPath
					m["name"] = name
					b, err := json.Marshal(m)
					if err != nil {
						return "", err
					}

					return string(b), nil
				},
				Config: testAccResourceVSphereComputeClusterVMGroupConfig(1),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMGroupExists(true),
					testAccResourceVSphereComputeClusterVMGroupMatchMembership(),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeClusterVMGroup_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterVMGroupExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterVMGroupConfig(2),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMGroupExists(true),
					testAccResourceVSphereComputeClusterVMGroupMatchMembership(),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterVMGroupConfig(3),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterVMGroupExists(true),
					testAccResourceVSphereComputeClusterVMGroupMatchMembership(),
				),
			},
		},
	})
}

func testAccResourceVSphereComputeClusterVMGroupExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		info, err := testGetComputeClusterVMGroup(s, "cluster_vm_group")
		if err != nil {
			if expected == false {
				if viapi.IsManagedObjectNotFoundError(err) {
					// This is not necessarily a missing group, but more than likely a
					// missing cluster, which happens during destroy as the dependent
					// resources will be missing as well, so want to treat this as a
					// deleted override as well.
					return nil
				}
			}
			return err
		}

		switch {
		case info == nil && !expected:
			// Expected missing
			return nil
		case info == nil && expected:
			// Expected to exist
			return errors.New("cluster VM group missing when expected to exist")
		case !expected:
			return errors.New("cluster VM group still present when expected to be missing")
		}

		return nil
	}
}

func testAccResourceVSphereComputeClusterVMGroupMatchMembership() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actual, err := testGetComputeClusterVMGroup(s, "cluster_vm_group")
		if err != nil {
			return err
		}

		if actual == nil {
			return errors.New("cluster VM group missing")
		}

		vms, err := testAccResourceVSphereComputeClusterVMGroupMatchMembershipVMIDs(s)
		if err != nil {
			return err
		}

		expectedSort := structure.MoRefSorter(vms)
		sort.Sort(expectedSort)

		expected := &types.ClusterVmGroup{
			ClusterGroupInfo: types.ClusterGroupInfo{
				Name:        actual.Name,
				UserCreated: actual.UserCreated,
			},
			Vm: []types.ManagedObjectReference(expectedSort),
		}

		actualSort := structure.MoRefSorter(actual.Vm)
		sort.Sort(actualSort)
		actual.Vm = actualSort

		if !reflect.DeepEqual(expected, actual) {
			return spew.Errorf("expected %#v got %#v", expected, actual)
		}

		return nil
	}
}

func testAccResourceVSphereComputeClusterVMGroupMatchMembershipVMIDs(s *terraform.State) ([]types.ManagedObjectReference, error) {
	var ids []string
	if rs, ok := s.RootModule().Resources["vsphere_virtual_machine.vm"]; ok {
		ids = []string{rs.Primary.ID}
	} else {
		ids = testAccResourceVSphereComputeClusterVMGroupGetMultiple(s)
	}

	results, err := virtualmachine.MOIDsForUUIDs(testAccProvider.Meta().(*Client).vimClient, ids)
	if err != nil {
		return nil, err
	}
	return results.ManagedObjectReferences(), nil
}

func testAccResourceVSphereComputeClusterVMGroupGetMultiple(s *terraform.State) []string {
	var i int
	var ids []string
	for {
		rs, ok := s.RootModule().Resources[fmt.Sprintf("vsphere_virtual_machine.vm.%d", i)]
		if !ok {
			break
		}
		ids = append(ids, rs.Primary.ID)
		i++
	}
	return ids
}

func testAccResourceVSphereComputeClusterVMGroupConfig(count int) string {
	return fmt.Sprintf(`
%s

variable "vm_count" {
  default = "%d"
}

resource "vsphere_virtual_machine" "vm" {
  count            = var.vm_count
  name             = "terraform-test-${count.index}"
  resource_pool_id = data.vsphere_compute_cluster.rootcompute_cluster1.resource_pool_id
  datastore_id     = data.vsphere_datastore.rootds1.id

  num_cpus = 2
  memory   = 2048
  guest_id = "other3xLinuxGuest"

  wait_for_guest_net_timeout = 0

  network_interface {
    network_id = data.vsphere_network.network1.id
  }

  disk {
    label = "disk0"
    size  = 1
    io_reservation = 1 
  }
}

resource "vsphere_compute_cluster_vm_group" "cluster_vm_group" {
  name                = "terraform-test-cluster-group"
  compute_cluster_id  = data.vsphere_compute_cluster.rootcompute_cluster1.id
  virtual_machine_ids = vsphere_virtual_machine.vm.*.id
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigDataRootDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
		testhelper.ConfigDataRootPortGroup1()),
		count,
	)
}
