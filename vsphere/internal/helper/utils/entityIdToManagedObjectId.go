// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package utils

import (
	"context"
	"log"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/virtualmachine"
)

const VM = "VirtualMachine"
const DISTRIBUTEDVIRTUALSWITCH = "VmwareDistributedVirtualSwitch"

func GetMoid(client *govmomi.Client, entityType string, id string) (string, error) {
	switch entityType {
	case VM:
		vm, err := virtualmachine.FromUUID(client, id)
		if err != nil {
			log.Printf("unable to find VM object with uuid:%s, error %s,treating given id as managed object id", id, err)
			return id, nil
		}
		return vm.Reference().Value, nil
	case DISTRIBUTEDVIRTUALSWITCH:
		dvsm := types.ManagedObjectReference{Type: "DistributedVirtualSwitchManager", Value: "DVSManager"}
		req := &types.QueryDvsByUuid{
			This: dvsm,
			Uuid: id,
		}
		resp, err := methods.QueryDvsByUuid(context.TODO(), client, req)
		if err != nil {
			log.Printf("unable to find DVS object with uuid:%s, error %s, treating given id as managed object id", id, err)
			return id, nil
		}
		return resp.Returnval.Reference().Value, nil
	default:
		return id, nil
	}
}
