// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package virtualdevice

import (
	"testing"

	"github.com/vmware/govmomi/vim25/types"
)

func TestDiskCapacityInGiB(t *testing.T) {
	cases := []struct {
		name     string
		subject  *types.VirtualDisk
		expected int
	}{
		{
			name: "capacityInBytes - integer GiB",
			subject: &types.VirtualDisk{
				CapacityInBytes: 4294967296,
				CapacityInKB:    4194304,
			},
			expected: 4,
		},
		{
			name: "capacityInKB - integer GiB",
			subject: &types.VirtualDisk{
				CapacityInKB: 4194304,
			},
			expected: 4,
		},
		{
			name: "capacityInBytes - non-integer GiB",
			subject: &types.VirtualDisk{
				CapacityInBytes: 4294968320,
				CapacityInKB:    4194305,
			},
			expected: 5,
		},
		{
			name: "capacityInKB - non-integer GiB",
			subject: &types.VirtualDisk{
				CapacityInKB: 4194305,
			},
			expected: 5,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := diskCapacityInGiB(tc.subject)
			if tc.expected != actual {
				t.Fatalf("expected %d, got %d", tc.expected, actual)
			}
		})
	}
}

func TestScsiUsableUnitsPerController(t *testing.T) {
	cases := []struct {
		name     string
		scsiType string
		expected int
	}{
		{name: "pvscsi", scsiType: SubresourceControllerTypeParaVirtual, expected: 63},
		{name: "lsilogic", scsiType: SubresourceControllerTypeLsiLogic, expected: 15},
		{name: "lsilogic-sas", scsiType: SubresourceControllerTypeLsiLogicSAS, expected: 15},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := scsiUsableUnitsPerController(tc.scsiType)
			if tc.expected != actual {
				t.Fatalf("expected %d, got %d", tc.expected, actual)
			}
		})
	}
}
