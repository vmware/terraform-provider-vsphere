// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/ssohelper"
)

func dataSourceVSphereSSOGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereSSOGroupRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the vCenter Single Sign-On group to look up.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The identity source domain the group belongs to. Defaults to the local (system) domain.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the group.",
			},
			"member_user": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The user members of the group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The username of the member.",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identity source domain the member belongs to.",
						},
					},
				},
			},
			"member_group": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The nested group members of the group.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the nested group.",
						},
						"domain": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identity source domain the nested group belongs to.",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereSSOGroupRead(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	// Default to the local (system) domain when none is supplied.
	domain := d.Get("domain").(string)
	if domain == "" {
		domain = client.Domain
	}
	lookup := ssohelper.ID(name, domain)
	log.Printf("[DEBUG] vsphere_sso_group (data source): reading group %q", lookup)

	group, err := client.FindGroup(ctx, lookup)
	if err != nil {
		return fmt.Errorf("error reading SSO group %q: %s", lookup, err)
	}
	if group == nil {
		return fmt.Errorf("SSO group %q not found", lookup)
	}

	groupID := ssohelper.ID(group.Id.Name, group.Id.Domain)
	_ = d.Set("domain", group.Id.Domain)
	_ = d.Set("description", group.Details.Description)

	memberUsers, memberGroups, err := readSSOGroupMembers(ctx, client, groupID)
	if err != nil {
		return err
	}
	if err := d.Set("member_user", memberUsers); err != nil {
		return err
	}
	if err := d.Set("member_group", memberGroups); err != nil {
		return err
	}

	d.SetId(groupID)
	return nil
}
