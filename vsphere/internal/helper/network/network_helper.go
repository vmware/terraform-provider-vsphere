// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package network

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/provider"
)

var Type = []string{
	"Network",
	"DistributedVirtualPortgroup",
	"OpaqueNetwork",
}

type NotFoundError struct {
	Name string
	ID   string
}

func (e NotFoundError) Error() string {
	if len(e.ID) > 0 {
		return fmt.Sprintf("Network with ID %s not found", e.ID)
	}

	return fmt.Sprintf("Network %s not found", e.Name)
}

// FromPath loads a network via its path.
//
// A network is a usually one of three kinds of networks: a DVS port group, a
// host port group, or a "opaque" network, provided externally from something
// like NSX. All three of these can be used as a backing for a virtual ethernet
// card, which is usually what these helpers are used with.
//
// Datacenter is optional here - if not provided, it's expected that the path
// is sufficient enough for finder to determine the datacenter required.
func FromPath(client *govmomi.Client, name string, dc *object.Datacenter) (object.NetworkReference, error) {
	finder := find.NewFinder(client.Client, false)
	if dc != nil {
		finder.SetDatacenter(dc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return finder.Network(ctx, name)
}

func FromNameAndDVSUuid(client *govmomi.Client, name string, dc *object.Datacenter, dvsUUID string) (object.NetworkReference, error) {
	finder := find.NewFinder(client.Client, false)
	if dc != nil {
		finder.SetDatacenter(dc)
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	networks, err := finder.NetworkList(ctx, name)
	if err != nil {
		return nil, err
	}
	if len(networks) == 0 {
		return nil, NotFoundError{Name: name}
	}

	switch {
	case len(networks) == 1 && dvsUUID == "":
		return networks[0], nil
	case len(networks) > 1 && dvsUUID == "":
		return nil, fmt.Errorf("path '%s' resolves to multiple %ss, Please specify", name, "network")
	case dvsUUID != "":
		dvsObj, err := dvsFromUUID(client, dvsUUID)
		if err != nil {
			return nil, err
		}
		dvsMoid := dvsObj.Reference().Value
		for _, network := range networks {
			if network.Reference().Type == "DistributedVirtualPortgroup" {
				dvPortGroup := object.NewDistributedVirtualPortgroup(client.Client, network.Reference())

				var dvPortGroupObj mo.DistributedVirtualPortgroup

				err = dvPortGroup.Properties(ctx, dvPortGroup.Reference(), []string{"config"}, &dvPortGroupObj)
				if err != nil {
					return nil, err
				}

				if dvPortGroupObj.Config.DistributedVirtualSwitch != nil &&
					dvsMoid == dvPortGroupObj.Config.DistributedVirtualSwitch.Value {
					return dvPortGroup, nil
				}
			}
		}
		return nil, fmt.Errorf("error while getting Network with name %s and Distributed virtual switch %s", name, dvsUUID)
	}
	return nil, NotFoundError{Name: name}
}

func FromNameAndVPCId(client *govmomi.Client, name string, dc *object.Datacenter, vpcID string) (object.NetworkReference, error) {
	ctx := context.TODO()
	finder := find.NewFinder(client.Client, true)

	// Set the datacenter
	if dc != nil {
		finder.SetDatacenter(dc)
	}

	// Find the network by name
	networks, err := finder.NetworkList(ctx, name)
	if err != nil {
		return nil, NotFoundError{Name: name}
	}
	// Filter networks by additional attributes
	for _, network := range networks {
		path := network.GetInventoryPath()
		pathSplit := strings.Split(path, "/")

		vpcIndex := slices.Index(pathSplit, "Virtual Private Clouds")
		// Path format is /<datacenter>/network/Virtual Private Clouds/<VPC>/<subnet> ... And could contain folders as well
		if vpcIndex != -1 && len(pathSplit) > vpcIndex && pathSplit[vpcIndex+1] == vpcID {

			if network.Reference().Type == "DistributedVirtualPortgroup" {
				dvPortGroup := object.NewDistributedVirtualPortgroup(client.Client, network.Reference())
				return dvPortGroup, nil
			}

			if netObj, ok := network.(*object.Network); ok {
				if networkName, err := netObj.ObjectName(ctx); err == nil && networkName == name {
					return network, nil
				}
			}
		}
	}

	return nil, NotFoundError{Name: name}
}

func List(client *govmomi.Client) ([]*object.VmwareDistributedVirtualSwitch, error) {
	return getSwitches(client, "/*")
}

func getSwitches(client *govmomi.Client, path string) ([]*object.VmwareDistributedVirtualSwitch, error) {
	ctx := context.TODO()
	var nets []*object.VmwareDistributedVirtualSwitch
	finder := find.NewFinder(client.Client, false)
	es, err := finder.ManagedObjectListChildren(ctx, path+"/*", "dvs", "folder")
	if err != nil {
		return nil, err
	}
	for _, id := range es {
		switch {
		case id.Object.Reference().Type == "VmwareDistributedVirtualSwitch":
			net, err := dvsFromMOID(client, id.Object.Reference().Value)
			if err != nil {
				return nil, err
			}
			nets = append(nets, net)
		case id.Object.Reference().Type == "Folder":
			newDSs, err := getSwitches(client, id.Path)
			if err != nil {
				return nil, err
			}
			nets = append(nets, newDSs...)
		default:
			continue
		}
	}
	return nets, nil
}

// FromID loads a network via its managed object reference ID.
func FromID(client *govmomi.Client, id string) (object.NetworkReference, error) {
	// I'm not too sure if a more efficient method to do this exists, but if this
	// becomes a pain point we might want to change this logic a bit.
	//
	// This is pretty much the direct example from
	// github.com/vmware/govmomi/examples/networks/main.go.
	m := view.NewManager(client.Client)

	vctx, vcancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer vcancel()
	v, err := m.CreateContainerView(vctx, client.ServiceContent.RootFolder, []string{"Network"}, true)
	if err != nil {
		return nil, err
	}

	defer func() {
		dctx, dcancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
		defer dcancel()
		_ = v.Destroy(dctx)
	}()

	var networks []mo.Network
	err = v.Retrieve(vctx, []string{"Network"}, []string{"name"}, &networks)
	if err != nil {
		return nil, err
	}

	for _, net := range networks {
		ref := net.Reference()
		if ref.Value == id {
			finder := find.NewFinder(client.Client, false)
			fctx, fcancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)

			nref, err := finder.ObjectReference(fctx, ref)
			if err != nil {
				fcancel()
				return nil, err
			}

			fcancel()

			// Should be safe to return here, as we have already asserted that this type
			// should be a NetworkReference by using ContainerView.
			return nref.(object.NetworkReference), nil
		}
	}

	return nil, NotFoundError{ID: id}
}

func dvsFromMOID(client *govmomi.Client, id string) (*object.VmwareDistributedVirtualSwitch, error) {
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "VmwareDistributedVirtualSwitch",
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	ds, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}
	// Should be safe to return here. If our reference returned here and is not a
	// VmwareDistributedVirtualSwitch, then we have bigger problems and to be
	// honest we should be panicking anyway.
	return ds.(*object.VmwareDistributedVirtualSwitch), nil
}
func dvsFromUUID(client *govmomi.Client, uuid string) (*object.VmwareDistributedVirtualSwitch, error) {
	dvsm := types.ManagedObjectReference{Type: "DistributedVirtualSwitchManager", Value: "DVSManager"}
	req := &types.QueryDvsByUuid{
		This: dvsm,
		Uuid: uuid,
	}
	resp, err := methods.QueryDvsByUuid(context.TODO(), client, req)
	if err != nil {
		return nil, err
	}

	return dvsFromMOID(client, resp.Returnval.Reference().Value)
}

// FromName fetches a network by name and applies additional filters.
func FromName(client *vim25.Client, name string, dc *object.Datacenter, filters map[string]string) (object.NetworkReference, error) {
	ctx := context.TODO()
	finder := find.NewFinder(client, true)

	// Set the datacenter
	if dc != nil {
		finder.SetDatacenter(dc)
	}

	// Find the network by name
	networks, err := finder.NetworkList(ctx, name)
	if err != nil {
		return nil, NotFoundError{Name: name}
	}

	// If multiple networks are found and no filters are specified, return an error
	if len(networks) > 1 && len(filters) == 0 {
		return nil, fmt.Errorf("multiple networks found with the name '%s'. Please specify a filter to narrow down the results", name)
	}

	// Filter networks by additional attributes
	for _, network := range networks {
		match := true
		for key, value := range filters {
			switch key {
			case "name":
				netObj, ok := network.(*object.Network)
				if !ok {
					match = false
					break
				}
				networkName, err := netObj.ObjectName(ctx)
				if err != nil || networkName != value {
					match = false
				}
			case "network_type":
				if network.Reference().Type != value {
					match = false
				}
			}
			if !match {
				break
			}
		}
		if match {
			return network, nil
		}
	}

	return nil, NotFoundError{Name: name}
}
