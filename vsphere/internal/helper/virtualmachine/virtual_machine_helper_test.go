// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package virtualmachine

import (
	"testing"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type stubReference types.ManagedObjectReference

func (s stubReference) Reference() types.ManagedObjectReference {
	return types.ManagedObjectReference(s)
}

func TestPickUUIDSearchResult(t *testing.T) {
	t.Parallel()

	uuid := "4238c0b4-0000-0000-0000-000000000001"
	one := stubReference{Type: "VirtualMachine", Value: "vm-1"}
	two := stubReference{Type: "VirtualMachine", Value: "vm-2"}

	t.Run("none", func(t *testing.T) {
		_, err := pickUUIDSearchResult(uuid, nil)
		if !IsUUIDNotFoundError(err) {
			t.Fatalf("expected UUIDNotFoundError, got %v", err)
		}
	})

	t.Run("one", func(t *testing.T) {
		got, err := pickUUIDSearchResult(uuid, []object.Reference{one})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Reference().Value != "vm-1" {
			t.Fatalf("got %s, want vm-1", got.Reference().Value)
		}
	})

	t.Run("multiple", func(t *testing.T) {
		_, err := pickUUIDSearchResult(uuid, []object.Reference{one, two})
		if err == nil {
			t.Fatal("expected error for duplicate UUIDs")
		}
		if IsUUIDNotFoundError(err) {
			t.Fatalf("did not expect UUIDNotFoundError: %v", err)
		}
		want := `multiple virtual machines with UUID "4238c0b4-0000-0000-0000-000000000001" found; set datacenter_id to scope the search to a specific datacenter`
		if err.Error() != want {
			t.Fatalf("got %q, want %q", err.Error(), want)
		}
	})
}
