// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package guestoscustomizations

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/provider"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

const (
	GuestOsCustomizationTypeWindows = "Windows"
	GuestOsCustomizationTypeLinux   = "Linux"
	// GuestOsCustomizationHostNameFixed user enters a host name
	GuestOsCustomizationHostNameFixed = "fixed"

	GuestOsCustomizationHostNamePrefixed = "prefixed"
	GuestOsCustomizationHostNameUnknown  = "unknown"

	GuestOsCustomizationHostNameVMname = "VMname"

	schemaPrefixVMClone = "clone.0.customize.0."

	schemaPrefixGOSC = "spec.0."
)

func netifKey(key string, n int, prefix string) string {
	netifKeyPrefix := prefix + "network_interface"
	return fmt.Sprintf("%s.%d.%s", netifKeyPrefix, n, key)
}

// matchGateway take an IP, mask, and gateway, and checks to see if the gateway
// is reachable from the IP address.
func matchGateway(a string, m int, g string) bool {
	ip := net.ParseIP(a)
	gw := net.ParseIP(g)
	var mask net.IPMask
	if ip.To4() != nil {
		mask = net.CIDRMask(m, 32)
	} else {
		mask = net.CIDRMask(m, 128)
	}
	if ip.Mask(mask).Equal(gw.Mask(mask)) {
		return true
	}
	return false
}

func v4CIDRMaskToDotted(mask int) string {
	m := net.CIDRMask(mask, 32)
	a := int(m[0])
	b := int(m[1])
	c := int(m[2])
	d := int(m[3])
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

type HostName struct {
	Type  string
	Value string
}

func SpecSchema(isVM bool) map[string]*schema.Schema {
	prefix := getSchemaPrefix(isVM)
	return map[string]*schema.Schema{
		// CustomizationGlobalIPSettings
		"dns_server_list": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "The list of DNS servers for a virtual network adapter with a static IP address.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},
		"dns_suffix_list": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A list of DNS search domains to add to the DNS configuration on the virtual machine.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},

		// CustomizationLinuxPrep
		"linux_options": {
			Type:          schema.TypeList,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{prefix + "windows_options", prefix + "windows_sysprep_text"},
			Description:   "A list of configuration options specific to Linux virtual machines.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"domain": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The domain name for this virtual machine.",
				},
				"host_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The hostname for this virtual machine.",
				},
				"hw_clock_utc": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Specifies whether or not the hardware clock should be in UTC or not.",
				},
				"script_text": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The customization script to run before and or after guest customization",
				},
				"time_zone": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Customize the time zone on the VM. This should be a time zone-style entry, like America/Los_Angeles.",
					ValidateFunc: validation.StringMatch(
						regexp.MustCompile("^[-+/_a-zA-Z0-9]+$"),
						"must be similar to America/Los_Angeles or other Linux/Unix TZ format",
					),
				},
			}},
		},

		// CustomizationSysprep
		"windows_options": {
			Type:          schema.TypeList,
			Optional:      true,
			MaxItems:      1,
			ConflictsWith: []string{prefix + "linux_options", prefix + "windows_sysprep_text"},
			Description:   "A list of configuration options specific to Windows virtual machines.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				// CustomizationGuiRunOnce
				"run_once_command_list": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "A list of commands to run at first user logon, after guest customization.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				// CustomizationGuiUnattended
				"auto_logon": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Specifies whether or not the VM automatically logs on as Administrator.",
				},
				"auto_logon_count": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     1,
					Description: "Specifies how many times the VM should auto-logon the Administrator account when auto_logon is true.",
				},
				"admin_password": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Ephemeral:   true,
					Description: "The new administrator password for this virtual machine.",
				},
				"time_zone": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     85,
					Description: "The new time zone for the virtual machine. This is a sysprep-dictated timezone code.",
				},

				// CustomizationIdentification
				"join_domain": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{prefix + "windows_options.0.workgroup"},
					Description:   "The domain that the virtual machine should join.",
					RequiredWith:  []string{prefix + "windows_options.0.domain_admin_user", prefix + "windows_options.0.domain_admin_password"},
				},
				"domain_ou": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{prefix + "windows_options.0.workgroup"},
					Description:   "The MachineObjectOU which specifies the full LDAP path name of the OU to which the virtual machine belongs.",
					RequiredWith:  []string{prefix + "windows_options.0.join_domain"},
				},
				"domain_admin_user": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{prefix + "windows_options.0.workgroup"},
					Description:   "The user account of the domain administrator used to join this virtual machine to the domain.",
					RequiredWith:  []string{prefix + "windows_options.0.join_domain"},
				},
				"domain_admin_password": {
					Type:          schema.TypeString,
					Optional:      true,
					Sensitive:     true,
					Ephemeral:     true,
					ConflictsWith: []string{prefix + "windows_options.0.workgroup"},
					Description:   "The password of the domain administrator used to join this virtual machine to the domain.",
					RequiredWith:  []string{prefix + "windows_options.0.join_domain"},
				},
				"workgroup": {
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{prefix + "windows_options.0.join_domain"},
					Description:   "The workgroup for this virtual machine if not joining a domain.",
				},

				// CustomizationUserData
				"computer_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The host name for this virtual machine.",
				},
				"full_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "Administrator",
					Description: "The full name of the user of this virtual machine.",
				},
				"organization_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "Managed by Terraform",
					Description: "The organization name this virtual machine is being installed for.",
				},
				"product_key": {
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					Description: "The product key for this virtual machine.",
				},
			}},
		},

		// CustomizationSysprepText
		"windows_sysprep_text": {
			Type:          schema.TypeString,
			Optional:      true,
			Sensitive:     true,
			ConflictsWith: []string{prefix + "linux_options", prefix + "windows_options"},
			Description:   "Use this option to specify a windows sysprep file directly.",
		},

		// CustomizationIPSettings
		"network_interface": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "A specification of network interface configuration options.",
			Elem: &schema.Resource{Schema: map[string]*schema.Schema{
				"dns_server_list": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Network-interface specific DNS settings for Windows operating systems. Ignored on Linux.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"dns_domain": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "A DNS search domain to add to the DNS configuration on the virtual machine.",
				},
				"ipv4_address": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The IPv4 address assigned to this network adapter. If left blank, DHCP is used.",
				},
				"ipv4_netmask": {
					Type:         schema.TypeInt,
					Optional:     true,
					Description:  "The IPv4 CIDR netmask for the supplied IP address. Ignored if DHCP is selected.",
					ValidateFunc: validation.IntAtMost(32),
				},
				"ipv6_address": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The IPv6 address assigned to this network adapter. If left blank, default auto-configuration is used.",
				},
				"ipv6_netmask": {
					Type:         schema.TypeInt,
					Optional:     true,
					Description:  "The IPv6 CIDR netmask for the supplied IP address. Ignored if auto-configuration is selected.",
					ValidateFunc: validation.IntAtMost(128),
				},
			}},
		},

		// Base-level settings
		"ipv4_gateway": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The IPv4 default gateway when using network_interface customization on the virtual machine. This address must be local to a static IPv4 address configured in an interface sub-resource.",
		},
		"ipv6_gateway": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The IPv6 default gateway when using network_interface customization on the virtual machine. This address must be local to a static IPv4 address configured in an interface sub-resource.",
		},
	}
}

func FromName(client *govmomi.Client, name string) (*types.CustomizationSpecItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()

	csm := object.NewCustomizationSpecManager(client.Client)
	return csm.GetCustomizationSpec(ctx, name)
}

func FlattenGuestOsCustomizationSpec(d *schema.ResourceData, specItem *types.CustomizationSpecItem, client *govmomi.Client) error {
	_ = d.Set("type", specItem.Info.Type)
	_ = d.Set("description", specItem.Info.Description)
	_ = d.Set("last_update_time", specItem.Info.LastUpdateTime.String())
	_ = d.Set("change_version", specItem.Info.ChangeVersion)

	specData := make(map[string]interface{})
	specData["dns_server_list"] = specItem.Spec.GlobalIPSettings.DnsServerList
	specData["dns_suffix_list"] = specItem.Spec.GlobalIPSettings.DnsSuffixList

	switch specItem.Info.Type {
	case GuestOsCustomizationTypeLinux:
		linuxPrep := specItem.Spec.Identity.(*types.CustomizationLinuxPrep)
		linuxOptions, err := flattenLinuxOptions(linuxPrep)
		if err != nil {
			return err
		}

		specData["linux_options"] = linuxOptions
	case GuestOsCustomizationTypeWindows:
		sysprepText := flattenSysprepText(specItem.Spec.Identity)
		if len(sysprepText) > 0 {
			specData["windows_sysprep_text"] = sysprepText
		} else {
			specItemWinOptions := specItem.Spec.Identity.(*types.CustomizationSysprep)
			version := viapi.ParseVersionFromClient(client)
			windowsOptions, err := flattenWindowsOptions(specItemWinOptions, version)
			if err != nil {
				return err
			}

			specData["windows_options"] = windowsOptions
		}

	}

	var networkInterfaces []map[string]interface{}
	for _, networkAdapterMapping := range specItem.Spec.NicSettingMap {
		data := make(map[string]interface{})
		data["dns_server_list"] = networkAdapterMapping.Adapter.DnsServerList
		data["dns_domain"] = networkAdapterMapping.Adapter.DnsDomain
		data["ipv4_address"] = ""
		if ipAddress, ok := networkAdapterMapping.Adapter.Ip.(*types.CustomizationFixedIp); ok {
			data["ipv4_address"] = ipAddress.IpAddress
		}
		if len(networkAdapterMapping.Adapter.SubnetMask) > 0 {
			ip := net.ParseIP(networkAdapterMapping.Adapter.SubnetMask)
			if ip != nil {
				addr := ip.To4()
				mask, _ := net.IPv4Mask(addr[0], addr[1], addr[2], addr[3]).Size()
				data["ipv4_netmask"] = mask
			}
		}

		if networkAdapterMapping.Adapter.IpV6Spec != nil {
			ipV6IP, ok := networkAdapterMapping.Adapter.IpV6Spec.Ip[0].(*types.CustomizationFixedIpV6)
			if ok {
				data["ipv6_address"] = ipV6IP.IpAddress
				data["ipv6_netmask"] = ipV6IP.SubnetMask
			}
		}

		networkInterfaces = append(networkInterfaces, data)
	}
	specData["network_interface"] = networkInterfaces
	spec := []map[string]interface{}{specData}
	_ = d.Set("spec", spec)

	return nil
}

func IsSpecOsApplicableToVMOs(vmOsFamily types.VirtualMachineGuestOsFamily, specType string) bool {
	if specType == GuestOsCustomizationTypeWindows && vmOsFamily == types.VirtualMachineGuestOsFamilyWindowsGuest {
		return true
	}

	if specType == GuestOsCustomizationTypeLinux && vmOsFamily == types.VirtualMachineGuestOsFamilyLinuxGuest {
		return true
	}

	return false
}

func ExpandGuestOsCustomizationSpec(d *schema.ResourceData, client *govmomi.Client) (*types.CustomizationSpecItem, error) {
	osType := d.Get("type").(string)
	osFamily := types.VirtualMachineGuestOsFamilyLinuxGuest
	if osType == GuestOsCustomizationTypeWindows {
		osFamily = types.VirtualMachineGuestOsFamilyWindowsGuest
	}

	version := viapi.ParseVersionFromClient(client)

	return &types.CustomizationSpecItem{
		Info: types.CustomizationSpecInfo{
			Name:        d.Get("name").(string),
			Type:        osType,
			Description: d.Get("description").(string),
		},
		Spec: ExpandCustomizationSpec(d, string(osFamily), false, version),
	}, nil
}

// ValidateCustomizationSpec checks the validity of the supplied customization
// spec. It should be called during diff customization to veto invalid configs.
func ValidateCustomizationSpec(d *schema.ResourceDiff, family string, isVM bool) error {
	prefix := getSchemaPrefix(isVM)
	// Validate that the proper section exists for OS family suboptions.
	linuxExists := len(d.Get(prefix+"linux_options").([]interface{})) > 0 || !structure.ValuesAvailable(prefix+"linux_options.", []string{"host_name", "domain"}, d)
	windowsExists := len(d.Get(prefix+"windows_options").([]interface{})) > 0 || !structure.ValuesAvailable(prefix+"windows_options.", []string{"computer_name"}, d)
	sysprepExists := d.Get(prefix+"windows_sysprep_text").(string) != "" || !structure.ValuesAvailable(prefix, []string{"windows_sysprep_text"}, d)
	switch {
	case family == string(types.VirtualMachineGuestOsFamilyLinuxGuest) && !linuxExists:
		return errors.New("linux_options must exist in VM customization options for Linux operating systems")
	case family == string(types.VirtualMachineGuestOsFamilyWindowsGuest) && !windowsExists && !sysprepExists:
		return errors.New("one of windows_options or windows_sysprep_text must exist in VM customization options for Windows operating systems")
	}
	return nil
}

func flattenWindowsOptions(customizationPrep *types.CustomizationSysprep, version viapi.VSphereVersion) ([]map[string]interface{}, error) {
	winOptionsData := make(map[string]interface{})
	if customizationPrep.GuiRunOnce != nil {
		winOptionsData["run_once_command_list"] = customizationPrep.GuiRunOnce.CommandList
	}
	winOptionsData["auto_logon"] = customizationPrep.GuiUnattended.AutoLogon
	winOptionsData["auto_logon_count"] = customizationPrep.GuiUnattended.AutoLogonCount
	winOptionsData["time_zone"] = customizationPrep.GuiUnattended.TimeZone
	winOptionsData["domain_admin_user"] = customizationPrep.Identification.DomainAdmin
	winOptionsData["join_domain"] = customizationPrep.Identification.JoinDomain

	// Minimum Supported Version: 8.0.2
	if version.AtLeast(viapi.VSphereVersion{Product: version.Product, Major: 8, Minor: 0, Patch: 2}) {
		winOptionsData["domain_ou"] = customizationPrep.Identification.DomainOU
	}

	winOptionsData["workgroup"] = customizationPrep.Identification.JoinWorkgroup
	hostName, err := flattenHostName(customizationPrep.UserData.ComputerName)
	if err != nil {
		return nil, err
	}
	winOptionsData["computer_name"] = hostName.Value
	winOptionsData["full_name"] = customizationPrep.UserData.FullName
	winOptionsData["organization_name"] = customizationPrep.UserData.OrgName
	winOptionsData["product_key"] = customizationPrep.UserData.ProductId

	return []map[string]interface{}{winOptionsData}, nil
}

func flattenLinuxOptions(customizationPrep *types.CustomizationLinuxPrep) ([]map[string]interface{}, error) {
	linuxOptionsData := make(map[string]interface{})
	linuxOptionsData["domain"] = customizationPrep.Domain
	hostName, err := flattenHostName(customizationPrep.HostName)
	if err != nil {
		return nil, err
	}

	linuxOptionsData["host_name"] = hostName.Value

	linuxOptionsData["hw_clock_utc"] = customizationPrep.HwClockUTC
	linuxOptionsData["script_text"] = customizationPrep.ScriptText
	linuxOptionsData["time_zone"] = customizationPrep.TimeZone

	return []map[string]interface{}{linuxOptionsData}, nil
}

func flattenSysprepText(identity types.BaseCustomizationIdentitySettings) string {
	sysprep, ok := identity.(*types.CustomizationSysprepText)
	if ok {
		return sysprep.Value
	}
	return ""
}

func flattenHostName(hostName types.BaseCustomizationName) (HostName, error) {
	if name, ok := hostName.(*types.CustomizationFixedName); ok {
		return HostName{
			Type:  GuestOsCustomizationHostNameFixed,
			Value: name.Name,
		}, nil
	}

	if name, ok := hostName.(*types.CustomizationPrefixName); ok {
		return HostName{
			Type:  GuestOsCustomizationHostNamePrefixed,
			Value: name.Base,
		}, nil
	}

	if _, ok := hostName.(*types.CustomizationVirtualMachineName); ok {
		return HostName{
			Type: GuestOsCustomizationHostNameVMname,
		}, nil
	}

	if _, ok := hostName.(*types.CustomizationUnknownName); ok {
		return HostName{
			Type: GuestOsCustomizationHostNameUnknown,
		}, nil
	}

	return HostName{}, errors.New("unknown linux host name type")
}

// ExpandCustomizationSpec reads certain ResourceData keys and
// returns a CustomizationSpec.
func ExpandCustomizationSpec(d *schema.ResourceData, family string, isVM bool, version viapi.VSphereVersion) types.CustomizationSpec {
	prefix := getSchemaPrefix(isVM)
	obj := types.CustomizationSpec{
		Identity:         expandBaseCustomizationIdentitySettings(d, family, prefix, version),
		GlobalIPSettings: expandCustomizationGlobalIPSettings(d, prefix),
		NicSettingMap:    expandSliceOfCustomizationAdapterMapping(d, prefix),
	}
	return obj
}

// expandBaseCustomizationIdentitySettings returns a
// BaseCustomizationIdentitySettings, depending on what is defined.
//
// Only one of the three types of identity settings can be specified: Linux
// settings (from linux_options), Windows settings (from windows_options), and
// the raw Windows sysprep file (via windows_sysprep_text).
func expandBaseCustomizationIdentitySettings(d *schema.ResourceData, family string, prefix string, version viapi.VSphereVersion) types.BaseCustomizationIdentitySettings {
	var obj types.BaseCustomizationIdentitySettings
	windowsExists := len(d.Get(prefix+"windows_options").([]interface{})) > 0
	sysprepExists := len(d.Get(prefix+"windows_sysprep_text").(string)) > 0
	switch {
	case family == string(types.VirtualMachineGuestOsFamilyLinuxGuest):
		linuxKeyPrefix := prefix + "linux_options.0."
		obj = expandCustomizationLinuxPrep(d, linuxKeyPrefix)
	case family == string(types.VirtualMachineGuestOsFamilyWindowsGuest) && windowsExists:
		windowsKeyPrefix := prefix + "windows_options.0."
		obj = expandCustomizationSysprep(d, windowsKeyPrefix, version)
	case family == string(types.VirtualMachineGuestOsFamilyWindowsGuest) && sysprepExists:
		obj = &types.CustomizationSysprepText{
			Value: d.Get(prefix + "windows_sysprep_text").(string),
		}
	default:
		obj = &types.CustomizationIdentitySettings{}
	}
	return obj
}

// expandCustomizationLinuxPrep reads certain ResourceData keys and
// returns a CustomizationLinuxPrep.
func expandCustomizationLinuxPrep(d *schema.ResourceData, prefix string) *types.CustomizationLinuxPrep {

	obj := &types.CustomizationLinuxPrep{
		HostName: &types.CustomizationFixedName{
			Name: d.Get(prefix + "host_name").(string),
		},
		Domain:     d.Get(prefix + "domain").(string),
		TimeZone:   d.Get(prefix + "time_zone").(string),
		ScriptText: d.Get(prefix + "script_text").(string),
		HwClockUTC: structure.GetBoolPtr(d, prefix+"hw_clock_utc"),
	}
	return obj
}

// expandCustomizationSysprep reads certain ResourceData keys and
// returns a CustomizationSysprep.
func expandCustomizationSysprep(d *schema.ResourceData, prefix string, version viapi.VSphereVersion) *types.CustomizationSysprep {
	obj := &types.CustomizationSysprep{
		GuiUnattended:  expandCustomizationGuiUnattended(d, prefix),
		UserData:       expandCustomizationUserData(d, prefix),
		GuiRunOnce:     expandCustomizationGuiRunOnce(d, prefix),
		Identification: expandCustomizationIdentification(d, prefix, version),
	}
	return obj
}

// expandCustomizationGuiRunOnce reads certain ResourceData keys and
// returns a CustomizationGuiRunOnce.
func expandCustomizationGuiRunOnce(d *schema.ResourceData, prefix string) *types.CustomizationGuiRunOnce {
	obj := &types.CustomizationGuiRunOnce{
		CommandList: structure.SliceInterfacesToStrings(d.Get(prefix + "run_once_command_list").([]interface{})),
	}
	if len(obj.CommandList) < 1 {
		return nil
	}
	return obj
}

// expandCustomizationGuiUnattended reads certain ResourceData keys and
// returns a CustomizationGuiUnattended.
func expandCustomizationGuiUnattended(d *schema.ResourceData, prefix string) types.CustomizationGuiUnattended {
	obj := types.CustomizationGuiUnattended{
		TimeZone:       int32(d.Get(prefix + "time_zone").(int)),
		AutoLogon:      d.Get(prefix + "auto_logon").(bool),
		AutoLogonCount: int32(d.Get(prefix + "auto_logon_count").(int)),
	}
	if v, ok := d.GetOk(prefix + "admin_password"); ok {
		obj.Password = &types.CustomizationPassword{
			Value:     v.(string),
			PlainText: true,
		}
	}

	return obj
}

// expandCustomizationIdentification reads certain ResourceData keys and
// returns a CustomizationIdentification.
func expandCustomizationIdentification(d *schema.ResourceData, prefix string, version viapi.VSphereVersion) types.CustomizationIdentification {
	obj := types.CustomizationIdentification{
		JoinWorkgroup: d.Get(prefix + "workgroup").(string),
		JoinDomain:    d.Get(prefix + "join_domain").(string),
		DomainAdmin:   d.Get(prefix + "domain_admin_user").(string),
	}

	// Minimum Supported Version: 8.0.2
	if version.AtLeast(viapi.VSphereVersion{Product: version.Product, Major: 8, Minor: 0, Patch: 2}) {
		obj.DomainOU = d.Get(prefix + "domain_ou").(string)
	}

	if v, ok := d.GetOk(prefix + "domain_admin_password"); ok {
		obj.DomainAdminPassword = &types.CustomizationPassword{
			Value:     v.(string),
			PlainText: true,
		}
	}
	return obj
}

// expandCustomizationUserData reads certain ResourceData keys and
// returns a CustomizationUserData.
func expandCustomizationUserData(d *schema.ResourceData, prefix string) types.CustomizationUserData {
	obj := types.CustomizationUserData{
		FullName: d.Get(prefix + "full_name").(string),
		OrgName:  d.Get(prefix + "organization_name").(string),
		ComputerName: &types.CustomizationFixedName{
			Name: d.Get(prefix + "computer_name").(string),
		},
		ProductId: d.Get(prefix + "product_key").(string),
	}
	return obj
}

// expandCustomizationGlobalIPSettings reads certain ResourceData keys and
// returns a CustomizationGlobalIPSettings.
func expandCustomizationGlobalIPSettings(d *schema.ResourceData, prefix string) types.CustomizationGlobalIPSettings {
	obj := types.CustomizationGlobalIPSettings{
		DnsSuffixList: structure.SliceInterfacesToStrings(d.Get(prefix + "dns_suffix_list").([]interface{})),
		DnsServerList: structure.SliceInterfacesToStrings(d.Get(prefix + "dns_server_list").([]interface{})),
	}
	return obj
}

// expandSliceOfCustomizationAdapterMapping reads certain ResourceData keys and
// returns a CustomizationAdapterMapping slice.
func expandSliceOfCustomizationAdapterMapping(d *schema.ResourceData, prefix string) []types.CustomizationAdapterMapping {
	s := d.Get(prefix + "network_interface").([]interface{})
	if len(s) < 1 {
		return nil
	}
	result := make([]types.CustomizationAdapterMapping, len(s))
	for i := range s {
		adapter := expandCustomizationIPSettings(d, i, prefix)
		obj := types.CustomizationAdapterMapping{
			Adapter: adapter,
		}
		result[i] = obj
	}
	return result
}

// expandCustomizationIPSettings reads certain ResourceData keys and
// returns a CustomizationIPSettings.
func expandCustomizationIPSettings(d *schema.ResourceData, n int, prefix string) types.CustomizationIPSettings {
	v4addr, v4addrOk := d.GetOk(netifKey("ipv4_address", n, prefix))
	v4mask := d.Get(netifKey("ipv4_netmask", n, prefix)).(int)
	v4gw, v4gwOk := d.Get(prefix + "ipv4_gateway").(string)
	var obj types.CustomizationIPSettings
	switch {
	case v4addrOk:
		obj.Ip = &types.CustomizationFixedIp{
			IpAddress: v4addr.(string),
		}
		obj.SubnetMask = v4CIDRMaskToDotted(v4mask)
		// Check for the gateway
		if v4gwOk && matchGateway(v4addr.(string), v4mask, v4gw) {
			obj.Gateway = []string{v4gw}
		}
	default:
		obj.Ip = &types.CustomizationDhcpIpGenerator{}
	}
	obj.DnsServerList = structure.SliceInterfacesToStrings(d.Get(netifKey("dns_server_list", n, prefix)).([]interface{}))
	obj.DnsDomain = d.Get(netifKey("dns_domain", n, prefix)).(string)
	obj.IpV6Spec = expandCustomizationIPSettingsIPV6AddressSpec(d, n, prefix)
	return obj
}

// expandCustomizationIPSettingsIPV6AddressSpec reads certain ResourceData keys and
// returns a CustomizationIPSettingsIpV6AddressSpec.
func expandCustomizationIPSettingsIPV6AddressSpec(d *schema.ResourceData, n int, prefix string) *types.CustomizationIPSettingsIpV6AddressSpec {
	v, ok := d.GetOk(netifKey("ipv6_address", n, prefix))
	if !ok {
		return nil
	}
	addr := v.(string)
	mask := d.Get(netifKey("ipv6_netmask", n, prefix)).(int)
	gw, gwOk := d.Get(prefix + "ipv6_gateway").(string)
	obj := &types.CustomizationIPSettingsIpV6AddressSpec{
		Ip: []types.BaseCustomizationIpV6Generator{
			&types.CustomizationFixedIpV6{
				IpAddress:  addr,
				SubnetMask: int32(mask),
			},
		},
	}
	if gwOk && matchGateway(addr, mask, gw) {
		obj.Gateway = []string{gw}
	}
	return obj
}

func getSchemaPrefix(inVMClone bool) string {
	if inVMClone {
		return schemaPrefixVMClone
	}

	return schemaPrefixGOSC
}
