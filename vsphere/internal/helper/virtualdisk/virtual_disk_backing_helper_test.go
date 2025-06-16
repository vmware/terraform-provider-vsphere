// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package virtualdisk

import (
	"reflect"
	"testing"

	"github.com/vmware/govmomi/vim25/types"
)

func TestVirtualDiskToSchemaPropsMap_WithNilBacking(t *testing.T) {
	result := ToSchemaPropsMap(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty map for nil backing, got map with %d elements", len(result))
	}
}

func TestVirtualDiskToSchemaPropsMap_WithDiskBacking(t *testing.T) {
	// Create backing object for testing
	diskBacking := types.VirtualDiskFlatVer2BackingInfo{
		VirtualDeviceFileBackingInfo: types.VirtualDeviceFileBackingInfo{
			FileName: "[datastore1] vm/disk.vmdk",
		},
		DiskMode:        "persistent",
		ThinProvisioned: types.NewBool(true),
	}
	result := ToSchemaPropsMap(&diskBacking)

	// Verify expected fields exist in the result
	expectedFields := []string{"DiskMode", "ThinProvisioned"}
	for _, field := range expectedFields {
		if _, exists := result[field]; !exists {
			t.Fatalf("expected field %s to be present in result map", field)
		}
	}

	// Verify field values
	if result["VirtualDeviceFileBackingInfo"].(types.VirtualDeviceFileBackingInfo).FileName != "[datastore1] vm/disk.vmdk" {
		t.Fatalf("expected FileName to be '[datastore1] vm/disk.vmdk', got %v", result["FileName"])
	}

	if result["DiskMode"] != "persistent" {
		t.Fatalf("expected DiskMode to be 'persistent', got %v", result["DiskMode"])
	}

	if !reflect.DeepEqual(result["ThinProvisioned"], types.NewBool(true)) {
		t.Fatalf("expected ThinProvisioned to be true, got %v", result["ThinProvisioned"])
	}
}

func TestVirtualDiskToSchemaPropsMap_WithEmptyBacking(t *testing.T) {
	// Create empty backing object
	emptyBacking := &types.VirtualDeviceFileBackingInfo{}

	result := ToSchemaPropsMap(emptyBacking)

	// Should have some fields but they might be zero values
	if len(result) == 0 {
		t.Fatalf("expected non-empty map for empty backing, got empty map")
	}
}
