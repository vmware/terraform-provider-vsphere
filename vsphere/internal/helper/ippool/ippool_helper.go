// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

// Package ippool provides helpers for managing vSphere IP Pools, which are
// exposed to users in the vSphere Client as "Network Protocol Profiles".
package ippool

import (
	"context"
	"errors"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/network"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/provider"
)

// ErrNotVirtualCenter is returned when the connected endpoint does not expose
// an IP Pool Manager, which is the case for direct ESXi host connections.
var ErrNotVirtualCenter = errors.New("network protocol profiles (IP pools) are only supported when connected to vCenter Server")

// manager returns the managed object reference of the IP Pool Manager
// singleton for the connected vCenter Server.
func manager(client *govmomi.Client) (*types.ManagedObjectReference, error) {
	if client.ServiceContent.IpPoolManager == nil {
		return nil, ErrNotVirtualCenter
	}
	return client.ServiceContent.IpPoolManager, nil
}

// List returns all IP pools defined on the supplied datacenter.
func List(client *govmomi.Client, dc types.ManagedObjectReference) ([]types.IpPool, error) {
	m, err := manager(client)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	resp, err := methods.QueryIpPools(ctx, client.Client, &types.QueryIpPools{
		This: *m,
		Dc:   dc,
	})
	if err != nil {
		return nil, err
	}
	return resp.Returnval, nil
}

// FromID locates an IP pool by its numeric ID within the supplied
// datacenter. A nil result (with a nil error) is returned if no pool with
// the given ID exists.
func FromID(client *govmomi.Client, dc types.ManagedObjectReference, id int32) (*types.IpPool, error) {
	pools, err := List(client, dc)
	if err != nil {
		return nil, err
	}
	for i := range pools {
		if pools[i].Id == id {
			return &pools[i], nil
		}
	}
	return nil, nil
}

// FromName locates an IP pool by its name within the supplied datacenter. A
// nil result (with a nil error) is returned if no pool with the given name
// exists.
func FromName(client *govmomi.Client, dc types.ManagedObjectReference, name string) (*types.IpPool, error) {
	pools, err := List(client, dc)
	if err != nil {
		return nil, err
	}
	for i := range pools {
		if pools[i].Name == name {
			return &pools[i], nil
		}
	}
	return nil, nil
}

// Create creates a new IP pool on the supplied datacenter and returns its
// generated ID.
func Create(client *govmomi.Client, dc types.ManagedObjectReference, pool types.IpPool) (int32, error) {
	m, err := manager(client)
	if err != nil {
		return 0, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	resp, err := methods.CreateIpPool(ctx, client.Client, &types.CreateIpPool{
		This: *m,
		Dc:   dc,
		Pool: pool,
	})
	if err != nil {
		return 0, err
	}
	return resp.Returnval, nil
}

// Update reconfigures an existing IP pool. pool.Id must be set to the ID of
// the pool being updated.
func Update(client *govmomi.Client, dc types.ManagedObjectReference, pool types.IpPool) error {
	m, err := manager(client)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	_, err = methods.UpdateIpPool(ctx, client.Client, &types.UpdateIpPool{
		This: *m,
		Dc:   dc,
		Pool: pool,
	})
	return err
}

// Delete destroys the IP pool with the given ID. force controls whether the
// pool is destroyed even if it still has allocated addresses.
func Delete(client *govmomi.Client, dc types.ManagedObjectReference, id int32, force bool) error {
	m, err := manager(client)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	_, err = methods.DestroyIpPool(ctx, client.Client, &types.DestroyIpPool{
		This:  *m,
		Dc:    dc,
		Id:    id,
		Force: force,
	})
	return err
}

// NetworkConflict describes a network that is already associated with a
// network protocol profile other than the one being configured.
type NetworkConflict struct {
	NetworkID string
	PoolID    int32
	PoolName  string
}

// NetworkConflicts checks the given datacenter for any existing IP pools
// (other than excludePoolID) that already have one of networkIDs
// associated with them. Associating a network with more than one IP pool
// silently moves it away from its current pool, so callers should surface
// any conflicts as an error rather than letting the API perform the move.
// Pass excludePoolID as -1 when checking a not-yet-created pool.
func NetworkConflicts(client *govmomi.Client, dc types.ManagedObjectReference, networkIDs []string, excludePoolID int32) ([]NetworkConflict, error) {
	pools, err := List(client, dc)
	if err != nil {
		return nil, err
	}

	wanted := make(map[string]bool, len(networkIDs))
	for _, id := range networkIDs {
		wanted[id] = true
	}

	var conflicts []NetworkConflict
	for _, pool := range pools {
		if pool.Id == excludePoolID {
			continue
		}
		for _, assoc := range pool.NetworkAssociation {
			if assoc.Network == nil || !wanted[assoc.Network.Value] {
				continue
			}
			conflicts = append(conflicts, NetworkConflict{
				NetworkID: assoc.Network.Value,
				PoolID:    pool.Id,
				PoolName:  pool.Name,
			})
		}
	}
	return conflicts, nil
}

// ExpandNetworkAssociations resolves a list of network managed object IDs
// (standard port groups, distributed port groups, or opaque networks) into
// the association entries expected by the IP pool API.
func ExpandNetworkAssociations(client *govmomi.Client, ids []string) ([]types.IpPoolAssociation, error) {
	assocs := make([]types.IpPoolAssociation, 0, len(ids))
	for _, id := range ids {
		net, err := network.FromID(client, id)
		if err != nil {
			return nil, err
		}
		ref := net.Reference()
		assocs = append(assocs, types.IpPoolAssociation{
			Network: &ref,
		})
	}
	return assocs, nil
}

// FlattenNetworkAssociations returns the network IDs associated with an IP
// pool, suitable for storing in Terraform state.
func FlattenNetworkAssociations(assocs []types.IpPoolAssociation) []interface{} {
	ids := make([]interface{}, 0, len(assocs))
	for _, a := range assocs {
		if a.Network != nil {
			ids = append(ids, a.Network.Value)
		}
	}
	return ids
}

// ExpandIPPoolConfigInfo builds the IPv4/IPv6 configuration portion of an IP
// pool from its Terraform schema representation.
func ExpandIPPoolConfigInfo(m map[string]interface{}) *types.IpPoolIpPoolConfigInfo {
	cfg := &types.IpPoolIpPoolConfigInfo{
		SubnetAddress: m["subnet"].(string),
		Netmask:       m["netmask"].(string),
		Gateway:       m["gateway"].(string),
		Range:         m["range"].(string),
	}

	if v, ok := m["dns_servers"].([]interface{}); ok && len(v) > 0 {
		dns := make([]string, len(v))
		for i, raw := range v {
			dns[i] = raw.(string)
		}
		cfg.Dns = dns
	}

	dhcp := m["dhcp_available"].(bool)
	cfg.DhcpServerAvailable = &dhcp

	enabled := m["enabled"].(bool)
	cfg.IpPoolEnabled = &enabled

	return cfg
}

// FlattenIPPoolConfigInfo converts an IPv4/IPv6 configuration, along with
// its address usage counters, into the Terraform schema representation.
func FlattenIPPoolConfigInfo(cfg *types.IpPoolIpPoolConfigInfo, available, allocated int32) []interface{} {
	if cfg == nil {
		return nil
	}

	m := map[string]interface{}{
		"subnet":              cfg.SubnetAddress,
		"netmask":             cfg.Netmask,
		"gateway":             cfg.Gateway,
		"range":               cfg.Range,
		"dns_servers":         cfg.Dns,
		"available_addresses": available,
		"allocated_addresses": allocated,
	}

	if cfg.DhcpServerAvailable != nil {
		m["dhcp_available"] = *cfg.DhcpServerAvailable
	}
	if cfg.IpPoolEnabled != nil {
		m["enabled"] = *cfg.IpPoolEnabled
	}

	return []interface{}{m}
}
