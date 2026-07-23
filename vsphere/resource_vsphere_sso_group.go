// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/ssoadmin"
	"github.com/vmware/govmomi/ssoadmin/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/ssohelper"
)

func ssoPrincipalMemberResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the principal (user or group)",
			},
			"domain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The domain of the principal (user or group)",
			},
		},
	}
}

func resourceVSphereSSOGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereSSOGroupCreate,
		Read:   resourceVSphereSSOGroupRead,
		Update: resourceVSphereSSOGroupUpdate,
		Delete: resourceVSphereSSOGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the group. Groups are created in the local (system) domain.",
			},
			"domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The identity source domain the group belongs to (the local/system domain).",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the group.",
			},
			"member_user": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The set of users that are members of this group. Members may come from any identity source, so each entry requires both a name and a domain. This is authoritative: users not listed here are removed from the group.",
				Elem:        ssoPrincipalMemberResource(),
			},
			"member_group": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The set of groups that are nested members of this group. Members may come from any identity source, so each entry requires both a name and a domain. This is authoritative: groups not listed here are removed from the group.",
				Elem:        ssoPrincipalMemberResource(),
			},
		},
	}
}

func resourceVSphereSSOGroupCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	log.Printf("[DEBUG] vsphere_sso_group: creating group %q", name)

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	details := types.AdminGroupDetails{Description: d.Get("description").(string)}
	if err := client.CreateGroup(ctx, name, details); err != nil {
		return fmt.Errorf("error creating SSO group %q: %s", name, err)
	}

	// Groups are created in the local (system) domain. The ID is qualified with
	// the domain so it is unique across identity sources.
	d.SetId(ssohelper.ID(name, client.Domain))

	if users := ssohelper.ExpandMembers(d.Get("member_user").(*schema.Set)); len(users) > 0 {
		if err := client.AddUsersToGroup(ctx, name, users...); err != nil {
			return fmt.Errorf("error adding user members to SSO group %q: %s", name, err)
		}
	}
	if groups := ssohelper.ExpandMembers(d.Get("member_group").(*schema.Set)); len(groups) > 0 {
		if err := client.AddGroupsToGroup(ctx, name, groups...); err != nil {
			return fmt.Errorf("error adding group members to SSO group %q: %s", name, err)
		}
	}

	return resourceVSphereSSOGroupRead(d, meta)
}

func resourceVSphereSSOGroupRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] vsphere_sso_group: reading group %q", d.Id())

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	group, err := client.FindGroup(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("error reading SSO group %q: %s", d.Id(), err)
	}
	if group == nil {
		log.Printf("[DEBUG] vsphere_sso_group: group %q no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	_ = d.Set("name", group.Id.Name)
	_ = d.Set("domain", group.Id.Domain)
	_ = d.Set("description", group.Details.Description)

	memberUsers, memberGroups, err := readSSOGroupMembers(ctx, client, d.Id())
	if err != nil {
		return err
	}
	_ = d.Set("member_user", memberUsers)
	_ = d.Set("member_group", memberGroups)

	return nil
}

func resourceVSphereSSOGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	name, _, err := ssohelper.ParseID(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] vsphere_sso_group: updating group %q", d.Id())

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	if d.HasChange("description") {
		details := types.AdminGroupDetails{Description: d.Get("description").(string)}
		if err := client.UpdateGroup(ctx, name, details); err != nil {
			return fmt.Errorf("error updating SSO group %q: %s", name, err)
		}
	}

	if d.HasChange("member_user") {
		oldRaw, newRaw := d.GetChange("member_user")
		oldUsers := ssohelper.ExpandMembers(oldRaw.(*schema.Set))
		newUsers := ssohelper.ExpandMembers(newRaw.(*schema.Set))
		if err := reconcileSSOGroupUsers(ctx, client, name, oldUsers, newUsers); err != nil {
			return err
		}
	}

	if d.HasChange("member_group") {
		oldRaw, newRaw := d.GetChange("member_group")
		oldGroups := ssohelper.ExpandMembers(oldRaw.(*schema.Set))
		newGroups := ssohelper.ExpandMembers(newRaw.(*schema.Set))
		if err := reconcileSSOGroupGroups(ctx, client, name, oldGroups, newGroups); err != nil {
			return err
		}
	}

	return resourceVSphereSSOGroupRead(d, meta)
}

func resourceVSphereSSOGroupDelete(d *schema.ResourceData, meta interface{}) error {
	name, _, err := ssohelper.ParseID(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] vsphere_sso_group: deleting group %q", d.Id())

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	if err := client.DeletePrincipal(ctx, name); err != nil {
		return fmt.Errorf("error deleting SSO group %q: %s", name, err)
	}

	d.SetId("")
	return nil
}

// readSSOGroupMembers returns the current user and nested-group members of the
// group identified by groupID ("name@domain"), each flattened as {name, domain}
// blocks.
func readSSOGroupMembers(ctx context.Context, client *ssoadmin.Client, groupID string) (users, groups []interface{}, err error) {
	memberUsers, err := client.FindUsersInGroup(ctx, groupID, "")
	if err != nil {
		return nil, nil, fmt.Errorf("error reading user members of SSO group %q: %s", groupID, err)
	}
	userIDs := make([]types.PrincipalId, 0, len(memberUsers))
	for _, u := range memberUsers {
		userIDs = append(userIDs, u.Id)
	}

	memberGroups, err := client.FindGroupsInGroup(ctx, groupID, "")
	if err != nil {
		return nil, nil, fmt.Errorf("error reading group members of SSO group %q: %s", groupID, err)
	}
	groupIDs := make([]types.PrincipalId, 0, len(memberGroups))
	for _, g := range memberGroups {
		groupIDs = append(groupIDs, g.Id)
	}

	return ssohelper.FlattenMembers(userIDs), ssohelper.FlattenMembers(groupIDs), nil
}

// reconcileSSOGroupUsers adds and removes user members so the group's user
// membership matches newPrincipals.
func reconcileSSOGroupUsers(ctx context.Context, client *ssoadmin.Client, name string, oldPrincipals, newPrincipals []types.PrincipalId) error {
	if add := subtractSSOPrincipalIDs(newPrincipals, oldPrincipals); len(add) > 0 {
		if err := client.AddUsersToGroup(ctx, name, add...); err != nil {
			return fmt.Errorf("error adding user members to SSO group %q: %s", name, err)
		}
	}
	if remove := subtractSSOPrincipalIDs(oldPrincipals, newPrincipals); len(remove) > 0 {
		if err := client.RemoveUsersFromGroup(ctx, name, remove...); err != nil {
			return fmt.Errorf("error removing user members from SSO group %q: %s", name, err)
		}
	}
	return nil
}

// reconcileSSOGroupGroups adds and removes nested group members so the group's
// group membership matches newPrincipals. Removal uses RemoveUsersFromGroup,
// which the ssoadmin API uses for both users and groups (there is no separate
// call).
func reconcileSSOGroupGroups(ctx context.Context, client *ssoadmin.Client, name string, oldPrincipals, newPrincipals []types.PrincipalId) error {
	if add := subtractSSOPrincipalIDs(newPrincipals, oldPrincipals); len(add) > 0 {
		if err := client.AddGroupsToGroup(ctx, name, add...); err != nil {
			return fmt.Errorf("error adding group members to SSO group %q: %s", name, err)
		}
	}
	if remove := subtractSSOPrincipalIDs(oldPrincipals, newPrincipals); len(remove) > 0 {
		if err := client.RemoveUsersFromGroup(ctx, name, remove...); err != nil {
			return fmt.Errorf("error removing group members from SSO group %q: %s", name, err)
		}
	}
	return nil
}

// subtractSSOPrincipalIDs returns the principals present in a but not in b.
func subtractSSOPrincipalIDs(a, b []types.PrincipalId) []types.PrincipalId {
	inB := make(map[types.PrincipalId]struct{}, len(b))
	for _, id := range b {
		inB[id] = struct{}{}
	}
	var out []types.PrincipalId
	for _, id := range a {
		if _, ok := inB[id]; !ok {
			out = append(out, id)
		}
	}
	return out
}
