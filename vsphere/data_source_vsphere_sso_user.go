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

func dataSourceVSphereSSOUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereSSOUserRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username of the vCenter Single Sign-On user to look up.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The identity source domain the user belongs to. Defaults to the local (system) domain. Set this to look up a user from an external identity source.",
			},
			"first_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The first name of the user.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The last name of the user.",
			},
			"email_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the user.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the user.",
			},
		},
	}
}

func dataSourceVSphereSSOUserRead(d *schema.ResourceData, meta interface{}) error {
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
	log.Printf("[DEBUG] vsphere_sso_user (data source): reading user %q", lookup)

	user, err := client.FindPersonUser(ctx, lookup)
	if err != nil {
		return fmt.Errorf("error reading SSO user %q: %s", lookup, err)
	}
	if user == nil {
		return fmt.Errorf("SSO user %q not found", lookup)
	}

	_ = d.Set("name", user.Id.Name)
	_ = d.Set("domain", user.Id.Domain)
	_ = d.Set("first_name", user.Details.FirstName)
	_ = d.Set("last_name", user.Details.LastName)
	_ = d.Set("email_address", user.Details.EmailAddress)
	_ = d.Set("description", user.Details.Description)

	d.SetId(ssohelper.ID(user.Id.Name, user.Id.Domain))
	return nil
}
