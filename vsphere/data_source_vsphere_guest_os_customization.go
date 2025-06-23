// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/guestoscustomizations"
)

func dataSourceVSphereGuestOSCustomization() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereGuestCustomizationRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the customization specification is the unique identifier per vCenter Server instance.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "TThe type of customization specification: One among: Windows, Linux.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
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
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Container object for the guest operating system properties to be customized.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_server_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A list of DNS servers for a virtual network adapter with a static IP address.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"dns_suffix_list": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A list of DNS search domains to add to the DNS configuration on the virtual machine.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"linux_options": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A list of configuration options specific to Linux.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The domain name for this virtual machine.",
									},
									"host_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The hostname for this virtual machine.",
									},
									"hw_clock_utc": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Specifies whether or not the hardware clock should be in UTC or not.",
									},
									"script_text": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The customization script to run before and or after guest customization.",
									},
									"time_zone": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Set the time zone on the guest operating system. For a list of the acceptable values for Linux customization specifications, see [List of Time Zone Database Zones](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) on Wikipedia.",
									},
								},
							},
						},
						"windows_options": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A list of configuration options specific to Windows.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"run_once_command_list": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "A list of commands to run at first user logon, after guest customization.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									// CustomizationGuiUnattended
									"auto_logon": {
										Type:        schema.TypeBool,
										Computed:    true,
										Description: "Specifies whether or not the guest operating system automatically logs on as Administrator.",
									},
									"auto_logon_count": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "Specifies how many times the guest operating system should auto-logon the Administrator account when `auto_logon` is `true`.",
									},
									"admin_password": {
										Type:        schema.TypeString,
										Computed:    true,
										Sensitive:   true,
										Description: "The new administrator password for this virtual machine.",
									},
									"time_zone": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The new time zone for the virtual machine. This is a sysprep-dictated timezone code.",
									},

									// CustomizationIdentification
									"join_domain": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The Active Directory domain for the virtual machine to join.",
									},
									"domain_ou": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The MachineObjectOU which specifies the full LDAP path name of the OU to which the virtual machine belongs.",
									},
									"domain_admin_user": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The user account of the domain administrator used to join this virtual machine to the domain.",
									},
									"domain_admin_password": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "The user account used to join this virtual machine to the Active Directory domain.",
									},
									"workgroup": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The workgroup for this virtual machine if not joining an Active Directory domain.",
									},
									"computer_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The hostname for this virtual machine.",
									},
									"full_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The full name of the user of this virtual machine.",
									},
									"organization_name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The organization name this virtual machine is being installed for.",
									},
									"product_key": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The product key for this virtual machine.",
									},
								},
							},
						},
						"windows_sysprep_text": {
							Type:        schema.TypeString,
							Computed:    true,
							Sensitive:   true,
							Description: "Use this option to specify use of a Windows Sysprep file.",
						},
						"network_interface": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "A specification of network interface configuration options.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"dns_server_list": {
										Type:        schema.TypeList,
										Computed:    true,
										Description: "Network-interface specific DNS settings for Windows operating systems. Ignored on Linux.",
										Elem:        &schema.Schema{Type: schema.TypeString},
									},
									"dns_domain": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "A DNS search domain to add to the DNS configuration on the virtual machine.",
									},
									"ipv4_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The IPv4 address assigned to this network adapter. If left blank, DHCP is used.",
									},
									"ipv4_netmask": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The IPv4 CIDR netmask for the supplied IP address. Ignored if DHCP is selected.",
									},
									"ipv6_address": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The IPv6 address assigned to this network adapter. If left blank, default auto-configuration is used.",
									},
									"ipv6_netmask": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The IPv6 CIDR netmask for the supplied IP address. Ignored if auto-configuration is selected.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereGuestCustomizationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	name := d.Get("name").(string)
	specItem, err := guestoscustomizations.FromName(client, name)
	if err != nil {
		return err
	}

	d.SetId(name)

	return guestoscustomizations.FlattenGuestOsCustomizationSpec(d, specItem, client)
}
