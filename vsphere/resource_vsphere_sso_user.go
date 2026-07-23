// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/ssoadmin/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/ssohelper"
)

func resourceVSphereSSOUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereSSOUserCreate,
		Read:   resourceVSphereSSOUserRead,
		Update: resourceVSphereSSOUserUpdate,
		Delete: resourceVSphereSSOUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The username of the user.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "The identity source domain the user belongs to. Defaults to the local (system) domain. Users can only be created in the local domain; a user from an external domain can be imported but not created or modified.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "The password for the user. This value is write-only; it cannot be read back from vCenter, so no drift is detected on it.",
			},
			"first_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The first name of the user.",
			},
			"last_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The last name of the user.",
			},
			"email_address": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The email address of the user.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of the user.",
			},
		},
	}
}

func resourceVSphereSSOUserCreate(d *schema.ResourceData, meta interface{}) error {
	name := d.Get("name").(string)
	log.Printf("[DEBUG] vsphere_sso_user: creating user %q", name)

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	// Users can only be created in the local (system) domain. A user in an
	// external domain is managed by its identity provider and can only be
	// imported.
	if domain := d.Get("domain").(string); domain != "" && domain != client.Domain {
		return fmt.Errorf("cannot create user in domain %q: users can only be created in the local domain %q (external users are managed by their identity provider and can only be imported)", domain, client.Domain)
	}

	details := types.AdminPersonDetails{
		Description:  d.Get("description").(string),
		EmailAddress: d.Get("email_address").(string),
		FirstName:    d.Get("first_name").(string),
		LastName:     d.Get("last_name").(string),
	}

	if err := client.CreatePersonUser(ctx, name, details, d.Get("password").(string)); err != nil {
		return fmt.Errorf("error creating SSO user %q: %s", name, err)
	}

	d.SetId(ssohelper.ID(name, client.Domain))
	return resourceVSphereSSOUserRead(d, meta)
}

func resourceVSphereSSOUserRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] vsphere_sso_user: reading user %q", d.Id())

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	user, err := client.FindPersonUser(ctx, d.Id())
	if err != nil {
		return fmt.Errorf("error reading SSO user %q: %s", d.Id(), err)
	}
	if user == nil {
		log.Printf("[DEBUG] vsphere_sso_user: user %q no longer exists", d.Id())
		d.SetId("")
		return nil
	}

	_ = d.Set("name", user.Id.Name)
	_ = d.Set("domain", user.Id.Domain)
	_ = d.Set("description", user.Details.Description)
	_ = d.Set("email_address", user.Details.EmailAddress)
	_ = d.Set("first_name", user.Details.FirstName)
	_ = d.Set("last_name", user.Details.LastName)
	// password is intentionally not set: it cannot be read back from vCenter.

	return nil
}

func resourceVSphereSSOUserUpdate(d *schema.ResourceData, meta interface{}) error {
	name, domain, err := ssohelper.ParseID(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] vsphere_sso_user: updating user %q", d.Id())

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	// Only local-domain users can be modified.
	if domain != client.Domain {
		return fmt.Errorf("cannot modify user %q: only users in the local domain %q can be modified", d.Id(), client.Domain)
	}

	if d.HasChanges("description", "email_address", "first_name", "last_name") {
		details := types.AdminPersonDetails{
			Description:  d.Get("description").(string),
			EmailAddress: d.Get("email_address").(string),
			FirstName:    d.Get("first_name").(string),
			LastName:     d.Get("last_name").(string),
		}
		if err := client.UpdatePersonUser(ctx, name, details); err != nil {
			return fmt.Errorf("error updating SSO user %q: %s", d.Id(), err)
		}
	}

	if d.HasChange("password") {
		if err := client.ResetPersonPassword(ctx, name, d.Get("password").(string)); err != nil {
			return fmt.Errorf("error resetting password for SSO user %q: %s", d.Id(), err)
		}
	}

	return resourceVSphereSSOUserRead(d, meta)
}

func resourceVSphereSSOUserDelete(d *schema.ResourceData, meta interface{}) error {
	name, domain, err := ssohelper.ParseID(d.Id())
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] vsphere_sso_user: deleting user %q", d.Id())

	ctx := context.Background()
	client, err := meta.(*Client).SSOAdminClient(ctx)
	if err != nil {
		return err
	}

	// Only local-domain users can be deleted.
	if domain != client.Domain {
		return fmt.Errorf("cannot delete user %q: only users in the local domain %q can be deleted", d.Id(), client.Domain)
	}

	if err := client.DeletePrincipal(ctx, name); err != nil {
		return fmt.Errorf("error deleting SSO user %q: %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}
