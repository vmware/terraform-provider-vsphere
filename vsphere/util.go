// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

func stringInSlice(s string, list []string) bool {
	for _, i := range list {
		if i == s {
			return true
		}
	}
	return false
}

func Strptr(s string) *string {
	return &s
}

func Intptr(i int) *int {
	return &i
}

func Boolptr(b bool) *bool {
	return &b
}
