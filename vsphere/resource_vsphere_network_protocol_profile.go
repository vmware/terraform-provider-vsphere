// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/ippool"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

const resourceVSphereNetworkProtocolProfileName = "vsphere_network_protocol_profile"

func resourceVSphereNetworkProtocolProfile() *schema.Resource {
	return &schema.Resource{
		Create: resourceVSphereNetworkProtocolProfileCreate,
		Read:   resourceVSphereNetworkProtocolProfileRead,
		Update: resourceVSphereNetworkProtocolProfileUpdate,
		Delete: resourceVSphereNetworkProtocolProfileDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVSphereNetworkProtocolProfileImport,
		},
		CustomizeDiff: resourceVSphereNetworkProtocolProfileCustomizeDiff,
		Schema: map[string]*schema.Schema{
			"datacenter_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The managed object ID of the datacenter this network protocol profile is associated with.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the network protocol profile.",
			},
			"network_ids": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "The managed object IDs of the networks (standard port groups, distributed port groups, or opaque networks) associated with this network protocol profile.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"dns_domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The DNS domain to use for this network protocol profile, for example \"example.com\".",
			},
			"dns_search_path": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The DNS search path to use for this network protocol profile, for example \"eng.example.com;example.com\".",
			},
			"host_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The prefix to use when generating host names for this network protocol profile.",
			},
			"http_proxy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The HTTP proxy to use on this network, in the form of a host and port, for example \"proxy.example.com:3128\".",
			},
			"ipv4": networkProtocolProfileIPConfigSchema("IPv4"),
			"ipv6": networkProtocolProfileIPConfigSchema("IPv6"),
		},
	}
}

func networkProtocolProfileIPConfigSchema(family string) *schema.Schema {
	return &schema.Schema{
		Type:         schema.TypeList,
		Optional:     true,
		MaxItems:     1,
		AtLeastOneOf: []string{"ipv4", "ipv6"},
		Description:  fmt.Sprintf("The %s configuration for this network protocol profile.", family),
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"subnet": {
					Type:        schema.TypeString,
					Required:    true,
					Description: fmt.Sprintf("The %s address of the subnet, for example \"10.10.10.0\".", family),
				},
				"netmask": {
					Type:        schema.TypeString,
					Required:    true,
					Description: fmt.Sprintf("The %s netmask of the subnet, for example \"255.255.255.0\".", family),
				},
				"gateway": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: fmt.Sprintf("The %s gateway of the subnet.", family),
				},
				"range": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The range(s) of addresses available for allocation, specified as one or more comma-separated \"<start-address>#<count>\" pairs, for example \"10.10.10.2#250\".",
				},
				"dns_servers": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "The DNS server addresses to use for this network protocol profile.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"dhcp_available": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Whether a DHCP server is available on this network.",
				},
				"enabled": {
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					Description: "Whether addresses can be allocated from this range.",
				},
				"available_addresses": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The number of addresses available for allocation from this range.",
				},
				"allocated_addresses": {
					Type:        schema.TypeInt,
					Computed:    true,
					Description: "The number of addresses currently allocated from this range.",
				},
			},
		},
	}
}

// networkIDSource is implemented by both *schema.ResourceData and
// *schema.ResourceDiff, allowing validateNetworkAssociations to run both at
// plan time (CustomizeDiff) and at apply time (Create/Update).
type networkIDSource interface {
	GetOk(key string) (interface{}, bool)
	Id() string
}

// validateNetworkAssociations ensures that none of the networks configured
// in network_ids are already associated with a different network protocol
// profile. vSphere does not reject this: it silently moves the network's
// association away from its current profile, which would surprise anyone
// relying on that existing configuration.
func validateNetworkAssociations(client *govmomi.Client, dc types.ManagedObjectReference, d networkIDSource) error {
	v, ok := d.GetOk("network_ids")
	if !ok {
		return nil
	}
	set, ok := v.(*schema.Set)
	if !ok || set.Len() == 0 {
		return nil
	}

	var selfID int32 = -1
	if id := d.Id(); id != "" {
		parsed, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return err
		}
		selfID = int32(parsed)
	}

	ids := make([]string, 0, set.Len())
	for _, raw := range set.List() {
		ids = append(ids, raw.(string))
	}

	conflicts, err := ippool.NetworkConflicts(client, dc, ids, selfID)
	if err != nil {
		return err
	}
	if len(conflicts) == 0 {
		return nil
	}

	msgs := make([]string, 0, len(conflicts))
	for _, c := range conflicts {
		msgs = append(msgs, fmt.Sprintf("  - network %q is already associated with network protocol profile %q (id %d)", c.NetworkID, c.PoolName, c.PoolID))
	}
	return fmt.Errorf("cannot associate network(s) that are already assigned to another network protocol profile, as this would silently move them:\n%s", strings.Join(msgs, "\n"))
}

// resourceVSphereNetworkProtocolProfileCustomizeDiff surfaces
// network association conflicts as a plan-time error rather than letting
// vSphere silently reassign the network at apply time.
func resourceVSphereNetworkProtocolProfileCustomizeDiff(_ context.Context, d *schema.ResourceDiff, meta interface{}) error {
	client := meta.(*Client).vimClient

	if !structure.ValuesAvailable("", []string{"datacenter_id", "network_ids"}, d) {
		return nil
	}

	dc, err := datacenterFromID(client, d.Get("datacenter_id").(string))
	if err != nil {
		return err
	}

	return validateNetworkAssociations(client, dc.Reference(), d)
}

func resourceVSphereNetworkProtocolProfileCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	dc, err := datacenterFromID(client, d.Get("datacenter_id").(string))
	if err != nil {
		return err
	}

	if err := validateNetworkAssociations(client, dc.Reference(), d); err != nil {
		return err
	}

	pool, err := expandIPPool(client, d)
	if err != nil {
		return err
	}

	id, err := ippool.Create(client, dc.Reference(), *pool)
	if err != nil {
		return fmt.Errorf("error creating network protocol profile: %s", err)
	}

	d.SetId(strconv.Itoa(int(id)))
	return resourceVSphereNetworkProtocolProfileRead(d, meta)
}

func resourceVSphereNetworkProtocolProfileRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	dc, err := datacenterFromID(client, d.Get("datacenter_id").(string))
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	id, err := strconv.ParseInt(d.Id(), 10, 32)
	if err != nil {
		return err
	}

	pool, err := ippool.FromID(client, dc.Reference(), int32(id))
	if err != nil {
		return err
	}
	if pool == nil {
		d.SetId("")
		return nil
	}

	return flattenIPPool(d, pool)
}

func resourceVSphereNetworkProtocolProfileUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	dc, err := datacenterFromID(client, d.Get("datacenter_id").(string))
	if err != nil {
		return err
	}

	if err := validateNetworkAssociations(client, dc.Reference(), d); err != nil {
		return err
	}

	pool, err := expandIPPool(client, d)
	if err != nil {
		return err
	}

	if err := ippool.Update(client, dc.Reference(), *pool); err != nil {
		return fmt.Errorf("error updating network protocol profile: %s", err)
	}

	return resourceVSphereNetworkProtocolProfileRead(d, meta)
}

func resourceVSphereNetworkProtocolProfileDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	dc, err := datacenterFromID(client, d.Get("datacenter_id").(string))
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return err
	}

	id, err := strconv.ParseInt(d.Id(), 10, 32)
	if err != nil {
		return err
	}

	if err := ippool.Delete(client, dc.Reference(), int32(id)); err != nil {
		return fmt.Errorf("error deleting network protocol profile: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVSphereNetworkProtocolProfileImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client := meta.(*Client).vimClient

	parts := strings.SplitN(d.Id(), ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil, fmt.Errorf("expected import ID in the format <datacenter-id>:<network-protocol-profile-id-or-name>, got: %s", d.Id())
	}

	dc, err := datacenterFromID(client, parts[0])
	if err != nil {
		return nil, err
	}

	var pool *types.IpPool
	if id, parseErr := strconv.ParseInt(parts[1], 10, 32); parseErr == nil {
		pool, err = ippool.FromID(client, dc.Reference(), int32(id))
	} else {
		pool, err = ippool.FromName(client, dc.Reference(), parts[1])
	}
	if err != nil {
		return nil, err
	}
	if pool == nil {
		return nil, fmt.Errorf("could not find network protocol profile %q in datacenter %s", parts[1], parts[0])
	}

	_ = d.Set("datacenter_id", dc.Reference().Value)
	d.SetId(strconv.Itoa(int(pool.Id)))

	return []*schema.ResourceData{d}, nil
}

// expandIPPool builds a types.IpPool from the resource's current schema
// data. If the resource already has an ID (i.e. this is an update), it is
// carried over so the API call knows which pool to modify.
func expandIPPool(client *govmomi.Client, d *schema.ResourceData) (*types.IpPool, error) {
	pool := &types.IpPool{
		Name:          d.Get("name").(string),
		DnsDomain:     d.Get("dns_domain").(string),
		DnsSearchPath: d.Get("dns_search_path").(string),
		HostPrefix:    d.Get("host_prefix").(string),
		HttpProxy:     d.Get("http_proxy").(string),
	}

	if v, ok := d.GetOk("ipv4"); ok {
		if l := v.([]interface{}); len(l) > 0 {
			pool.Ipv4Config = ippool.ExpandIPPoolConfigInfo(l[0].(map[string]interface{}))
		}
	}
	if v, ok := d.GetOk("ipv6"); ok {
		if l := v.([]interface{}); len(l) > 0 {
			pool.Ipv6Config = ippool.ExpandIPPoolConfigInfo(l[0].(map[string]interface{}))
		}
	}

	if v, ok := d.GetOk("network_ids"); ok {
		ids := structure.SliceInterfacesToStrings(v.(*schema.Set).List())
		assocs, err := ippool.ExpandNetworkAssociations(client, ids)
		if err != nil {
			return nil, err
		}
		pool.NetworkAssociation = assocs
	}

	if id := d.Id(); id != "" {
		parsed, err := strconv.ParseInt(id, 10, 32)
		if err != nil {
			return nil, err
		}
		pool.Id = int32(parsed)
	}

	return pool, nil
}

// flattenIPPool writes the attributes of pool into the resource's schema
// data.
func flattenIPPool(d *schema.ResourceData, pool *types.IpPool) error {
	_ = d.Set("name", pool.Name)
	_ = d.Set("dns_domain", pool.DnsDomain)
	_ = d.Set("dns_search_path", pool.DnsSearchPath)
	_ = d.Set("host_prefix", pool.HostPrefix)
	_ = d.Set("http_proxy", pool.HttpProxy)
	_ = d.Set("network_ids", ippool.FlattenNetworkAssociations(pool.NetworkAssociation))

	if pool.Ipv4Config != nil && pool.Ipv4Config.SubnetAddress != "" {
		_ = d.Set("ipv4", ippool.FlattenIPPoolConfigInfo(pool.Ipv4Config, pool.AvailableIpv4Addresses, pool.AllocatedIpv4Addresses))
	}
	if pool.Ipv6Config != nil && pool.Ipv6Config.SubnetAddress != "" {
		_ = d.Set("ipv6", ippool.FlattenIPPoolConfigInfo(pool.Ipv6Config, pool.AvailableIpv6Addresses, pool.AllocatedIpv6Addresses))
	}

	return nil
}
