// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package resourcepool

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/computeresource"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/provider"
)

// List retrieves all resource pools.
func List(client *govmomi.Client) ([]*object.ResourcePool, error) {
	return FromPath(client, "/*")
}

// FromPathOrDefault retrieves a resource pool using its supplied path.
func FromPathOrDefault(client *govmomi.Client, name string, dc *object.Datacenter) (*object.ResourcePool, error) {
	finder := find.NewFinder(client.Client, false)

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	t := client.ServiceContent.About.ApiType
	switch t {
	case "HostAgent":
		ddc, err := finder.DefaultDatacenter(ctx)
		if err != nil {
			return nil, err
		}
		finder.SetDatacenter(ddc)
		return finder.DefaultResourcePool(ctx)
	case "VirtualCenter":
		if dc != nil {
			finder.SetDatacenter(dc)
		}
		if name != "" {
			return finder.ResourcePool(ctx, name)
		}
		return finder.DefaultResourcePool(ctx)
	}
	return nil, fmt.Errorf("unsupported ApiType: %s", t)
}

// FromParentAndName retrieves a resource pool by its name and the ID of its parent resource pool.
func FromParentAndName(client *govmomi.Client, parentID string, name string) (*object.ResourcePool, error) {
	if strings.Contains(name, "/") {
		return nil, fmt.Errorf("argument 'name' cannot be a path when 'parent_resource_pool_id' is specified, use the simple resource pool name")
	}

	finder := find.NewFinder(client.Client, false)
	parentRef := types.ManagedObjectReference{
		Type:  "ResourcePool",
		Value: parentID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	parentObj, err := finder.ObjectReference(ctx, parentRef)
	defer cancel()
	if err != nil {
		return nil, fmt.Errorf("could not find parent resource pool with ID %q: %w", parentID, err)
	}

	parentRP, ok := parentObj.(*object.ResourcePool)
	if !ok {
		return nil, fmt.Errorf("object with ID %q is not a ResourcePool", parentID)
	}

	ctx, cancel = context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	var parentMo mo.ResourcePool
	err = parentRP.Properties(ctx, parentRP.Reference(), []string{"resourcePool"}, &parentMo)
	defer cancel()
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for parent resource pool %q: %w", parentID, err)
	}

	var errorMessages []string
	for _, childRef := range parentMo.ResourcePool {
		ctx, cancel = context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
		var childMo mo.ResourcePool
		childRP := object.NewResourcePool(client.Client, childRef)
		err = childRP.Properties(ctx, childRef, []string{"name"}, &childMo)

		if err != nil {
			cancel()
			errorMessages = append(errorMessages, fmt.Sprintf("could not get properties for child resource pool %s: %s", childRef.Value, err))
			continue
		}

		if childMo.Name == name {
			cancel()
			return childRP, nil
		}

		cancel()
	}

	if len(errorMessages) > 0 {
		var msg strings.Builder

		_, err := fmt.Fprintf(&msg, "resource pool %q not found under parent resource pool %q. Errors encountered during search:", name, parentID)
		if err != nil {
			return nil, fmt.Errorf("error building error message: %w", err)
		}

		for i, errMsg := range errorMessages {
			if i < 5 {
				_, err := fmt.Fprintf(&msg, "\n- %s", errMsg)
				if err != nil {
					return nil, fmt.Errorf("error building error message: %w", err)
				}
			} else {
				_, err := fmt.Fprintf(&msg, "\n- and %d more errors...", len(errorMessages)-5)
				if err != nil {
					return nil, fmt.Errorf("error building error message: %w", err)
				}
				break
			}
		}

		return nil, fmt.Errorf("%s", msg.String())
	}

	return nil, fmt.Errorf("resource pool %q not found under parent resource pool %q", name, parentID)
}

// FromPath retrieves all resource pools recursively from the specified inventory path.
func FromPath(client *govmomi.Client, path string) ([]*object.ResourcePool, error) {
	ctx := context.TODO()
	var rps []*object.ResourcePool
	finder := find.NewFinder(client.Client, false)
	es, err := finder.ManagedObjectListChildren(ctx, path+"/*", "pool", "folder")
	if err != nil {
		return nil, err
	}
	for _, id := range es {
		if id.Object.Reference().Type == "ResourcePool" {
			ds, err := FromID(client, id.Object.Reference().Value)
			if err != nil {
				return nil, err
			}
			rps = append(rps, ds)
		}
		if id.Object.Reference().Type == "Folder" || id.Object.Reference().Type == "ClusterComputeResource" || id.Object.Reference().Type == "ResourcePool" {
			newRPs, err := FromPath(client, id.Path)
			if err != nil {
				return nil, err
			}
			rps = append(rps, newRPs...)
		}
	}
	return rps, nil
}

// FromID retrieves a resource pool by its managed object reference ID.
func FromID(client *govmomi.Client, id string) (*object.ResourcePool, error) {
	log.Printf("[DEBUG] Locating resource pool with ID %s", id)
	finder := find.NewFinder(client.Client, false)

	ref := types.ManagedObjectReference{
		Type:  "ResourcePool",
		Value: id,
	}

	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	obj, err := finder.ObjectReference(ctx, ref)
	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Resource pool found: %s", obj.Reference().Value)
	return obj.(*object.ResourcePool), nil
}

// Properties retrieves the resource pool managed object from its higher-level object.
func Properties(obj *object.ResourcePool) (*mo.ResourcePool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	var props mo.ResourcePool
	if err := obj.Properties(ctx, obj.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

// ValidateHost verifies if the specified host is a member of the given resource pool.
func ValidateHost(client *govmomi.Client, pool *object.ResourcePool, host *object.HostSystem) error {
	if host == nil {
		// Nothing to validate here, move along
		log.Printf("[DEBUG] ValidateHost: no host supplied, nothing to do")
		return nil
	}
	log.Printf("[DEBUG] Validating that host %q is a member of resource pool %q", host.Reference().Value, pool.Reference().Value)
	pprops, err := Properties(pool)
	if err != nil {
		return err
	}
	cprops, err := computeresource.BasePropertiesFromReference(client, pprops.Owner)
	if err != nil {
		return err
	}
	for _, href := range cprops.Host {
		if href.Value == host.Reference().Value {
			log.Printf("[DEBUG] Validated that host %q is a member of resource pool %q.", host.Reference().Value, pool.Reference().Value)
			return nil
		}
	}
	return fmt.Errorf("host ID %q is not a member of resource pool %q", host.Reference().Value, pool.Reference().Value)
}

// DefaultDevices retrieves the default virtual device list for a given resource pool and guest OS type.
func DefaultDevices(client *govmomi.Client, pool *object.ResourcePool, guest string) (object.VirtualDeviceList, error) {
	log.Printf("[DEBUG] Fetching default device list for resource pool %q for OS type %q", pool.Reference().Value, guest)
	pprops, err := Properties(pool)
	if err != nil {
		return nil, err
	}
	return computeresource.DefaultDevicesFromReference(client, pprops.Owner, guest)
}

// OSFamily determines the operating system family for a given guest ID based on the resource pool and hardware version.
func OSFamily(client *govmomi.Client, pool *object.ResourcePool, guest string, hardwareVersion int) (string, error) {
	log.Printf("[DEBUG] Looking for OS family for guest ID %q", guest)
	pprops, err := Properties(pool)
	if err != nil {
		return "", err
	}
	return computeresource.OSFamily(client, pprops.Owner, guest, hardwareVersion)
}

// Create creates a resource pool.
func Create(rp *object.ResourcePool, name string, spec *types.ResourceConfigSpec) (*object.ResourcePool, error) {
	log.Printf("[DEBUG] Creating resource pool %q", fmt.Sprintf("%s/%s", rp.InventoryPath, name))
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	nrp, err := rp.Create(ctx, name, *spec)
	if err != nil {
		return nil, err
	}
	return nrp, nil
}

// Update updates a resource pool.
func Update(rp *object.ResourcePool, name string, spec *types.ResourceConfigSpec) error {
	log.Printf("[DEBUG] Updating resource pool %q", rp.InventoryPath)
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	return rp.UpdateConfig(ctx, name, spec)
}

// Delete destroys a resource pool.
func Delete(rp *object.ResourcePool) error {
	log.Printf("[DEBUG] Deleting resource pool %q", rp.InventoryPath)
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	task, err := rp.Destroy(ctx)
	if err != nil {
		return err
	}
	return task.WaitEx(ctx)
}

// MoveIntoResourcePool moves a virtual machine, resource pool, or vApp into the specified resource pool.
func MoveIntoResourcePool(p *object.ResourcePool, c types.ManagedObjectReference) error {
	req := types.MoveIntoResourcePool{
		This: p.Reference(),
		List: []types.ManagedObjectReference{c},
	}
	ctx, cancel := context.WithTimeout(context.Background(), provider.DefaultAPITimeout)
	defer cancel()
	_, err := methods.MoveIntoResourcePool(ctx, p.Client(), &req)
	return err
}

// HasChildren checks to see if a resource pool has any child items and returns true if that is the case.
// This is useful when checking to see if a resource pool is safe to delete. Destroying a resource pool destroys all
// children if at all possible, so extra verification is necessary to prevent accidental removal.
func HasChildren(rp *object.ResourcePool) (bool, error) {
	props, err := Properties(rp)
	if err != nil {
		return false, err
	}
	if len(props.Vm) > 0 || len(props.ResourcePool) > 0 {
		return true, nil
	}
	return false, nil
}
