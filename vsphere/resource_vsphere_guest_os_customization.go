// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/guestoscustomizations"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/provider"
)

func resourceVSphereGuestOsCustomization() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereGuestOsCustomizationCreate,
		Read:   resourceVSphereGuestOsCustomizationRead,
		Update: resourceVSphereGuestOsCustomizationUpdate,
		Delete: resourceVSphereGuestOsCustomizationDelete,
		Schema: getSchema(),
	}
}

func resourceVSphereGuestOsCustomizationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	specItem, err := guestoscustomizations.FromName(client, d.Id())
	if err != nil {
		return err
	}

	return guestoscustomizations.FlattenGuestOsCustomizationSpec(d, specItem, client)
}

func resourceVSphereGuestOsCustomizationCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Beginning creation of customization specification %s", d.Get("name"))
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	csm := object.NewCustomizationSpecManager(client.Client)

	spec, err := guestoscustomizations.ExpandGuestOsCustomizationSpec(d, client)
	if err != nil {
		log.Printf("[ERROR] Error creating customization specification %s expansion: %s", d.Get("name"), err.Error())
		return err
	}
	log.Printf("[DEBUG] Successfully expanded customization specification %s", d.Get("name"))

	err = csm.CreateCustomizationSpec(ctx, *spec)
	if err == nil {
		log.Printf("[DEBUG] Successfully created customization specification %s", d.Get("name"))
		d.SetId(spec.Info.Name)
		return resourceVSphereGuestOsCustomizationRead(d, meta)
	}

	log.Printf("[ERROR] Error creating customization specification %s:  %s ", d.Get("name"), err.Error())

	return err
}

func resourceVSphereGuestOsCustomizationUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating customization specification %s", d.Get("name"))
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	csm := object.NewCustomizationSpecManager(client.Client)

	oldName, newName := d.GetChange("name")
	if oldName != newName {
		log.Printf("[DEBUG] Renaming customization specification name %s to %s", oldName, newName)
		err := csm.RenameCustomizationSpec(ctx, oldName.(string), newName.(string))
		if err != nil {
			log.Printf("[ERROR] Renaming customization specification %s to %s: %s", oldName, newName, err.Error())
			return err
		}

		log.Printf("[DEBUG] Successfully renamed customization specification %s to %s. Reseting the ID", oldName, newName)
		d.SetId(newName.(string))
	}

	spec, err := guestoscustomizations.ExpandGuestOsCustomizationSpec(d, client)
	if err != nil {
		log.Printf("[ERROR] Error expanding the customization specification %s: %s ", d.Get("name"), err.Error())
		return err
	}

	log.Printf("[DEBUG] Updating customization specification %s", d.Get("name").(string))
	err = csm.OverwriteCustomizationSpec(ctx, *spec)
	if err != nil {
		log.Printf("[ERROR] Error updating customization specification %s: %s", d.Get("name").(string), err.Error())
		return err
	}

	log.Printf("[DEBUG] Successfully updated customization specification %s", d.Get("name").(string))

	return nil
}

func resourceVSphereGuestOsCustomizationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	csm := object.NewCustomizationSpecManager(client.Client)
	return csm.DeleteCustomizationSpec(ctx, d.Id())
}

func getSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the customization specification is the unique identifier per vCenter Server instance.",
		},
		"type": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "The type of customization specification: One among: Windows, Linux.",
			ValidateFunc: validation.StringInSlice([]string{guestoscustomizations.GuestOsCustomizationTypeLinux, guestoscustomizations.GuestOsCustomizationTypeWindows}, false),
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The description for the customization specification.",
		},
		"last_update_time": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The time of last modification to the customization specification.",
		},
		"change_version": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The number of last changed version to the customization specification.",
		},
		"spec": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Required: true,
			Elem: &schema.Resource{
				Schema: guestoscustomizations.SpecSchema(false),
			},
		},
	}
}
