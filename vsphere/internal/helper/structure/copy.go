// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package structure

import (
	"fmt"

	"github.com/vmware/govmomi/vim25/types"
)

// CopyMap returns a shallow copy of a map[string]interface{}.
func CopyMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// CopyDeviceConfigSpec returns a shallow copy of a types.BaseVirtualDeviceConfigSpec.
func CopyDeviceConfigSpec(op types.BaseVirtualDeviceConfigSpec) types.BaseVirtualDeviceConfigSpec {
	switch v := op.(type) {
	case *types.VirtualDeviceConfigSpec:
		c := *v
		return &c
	default:
		panic(fmt.Sprintf("unsupported virtual device config spec type %T", op))
	}
}
