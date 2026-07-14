// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package ssohelper

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/ssoadmin/types"
)

// ID builds the "name@domain" identifier used as the Terraform resource ID for
// SSO principals that can belong to any identity source.
func ID(name, domain string) string {
	return name + "@" + domain
}

// ParseID splits a "name@domain" identifier into its components.
func ParseID(id string) (name, domain string, err error) {
	parts := strings.SplitN(id, "@", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid SSO ID %q: expected \"name@domain\"", id)
	}
	return parts[0], parts[1], nil
}

// ExpandMembers converts a set of {name, domain} blocks into the slice of
// PrincipalId values expected by the ssoadmin membership methods. Members (users
// or nested groups) may come from any identity source, so the domain is taken
// from each block.
func ExpandMembers(set *schema.Set) []types.PrincipalId {
	if set == nil || set.Len() == 0 {
		return nil
	}
	ids := make([]types.PrincipalId, 0, set.Len())
	for _, raw := range set.List() {
		m := raw.(map[string]interface{})
		ids = append(ids, types.PrincipalId{
			Name:   m["name"].(string),
			Domain: m["domain"].(string),
		})
	}
	return ids
}

// FlattenMembers converts a slice of PrincipalId values into the format expected
// by a schema.TypeSet or schema.TypeList of {name, domain} blocks.
func FlattenMembers(ids []types.PrincipalId) []interface{} {
	out := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		out = append(out, map[string]interface{}{
			"name":   id.Name,
			"domain": id.Domain,
		})
	}
	return out
}
