// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/hostsystem"
)

// hostStorageSystemFromHostSystemID locates a HostStorageSystem from a
// specified HostSystem managed object ID.
func hostStorageSystemFromHostSystemID(client *govmomi.Client, hsID string) (*object.HostStorageSystem, error) {
	hs, err := hostsystem.FromID(client, hsID)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
	defer cancel()
	return hs.ConfigManager().StorageSystem(ctx)
}
