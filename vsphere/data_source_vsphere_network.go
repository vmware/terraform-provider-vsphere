// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/network"
)

const (
	waitForNetworkPending   = "waitForNetworkPending"
	waitForNetworkCompleted = "waitForNetworkCompleted"
	waitForNetworkError     = "waitForNetworkError"
)

func dataSourceVSphereNetwork() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereNetworkRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Description:  "The name or path of the network.",
				Optional:     true,
				ExactlyOneOf: []string{"name", "vlan_id"},
			},
			"vlan_id": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "The VLAN ID of the network.",
				ExactlyOneOf: []string{"name", "vlan_id"},
			},
			"datacenter_id": {
				Type:        schema.TypeString,
				Description: "The managed object ID of the datacenter the network is in. This is required if the supplied path is not an absolute path containing a datacenter and there are multiple datacenters in your infrastructure.",
				Optional:    true,
			},
			"type": {
				Type:        schema.TypeString,
				Description: "The managed object type of the network.",
				Computed:    true,
			},
			"distributed_virtual_switch_uuid": {
				Type:        schema.TypeString,
				Description: "Id of the distributed virtual switch of which the port group is a part of",
				Optional:    true,
			},
			"vpc_project_id": {
				Type:        schema.TypeString,
				Description: "Id of the project object, which the VPC and network belongs to",
				Optional:    true,
			},
			"vpc_id": {
				Type:        schema.TypeString,
				Description: "Id of the VPC which the network belongs to",
				Optional:    true,
			},
			"filter": {
				Type:        schema.TypeSet,
				Description: "Apply a filter for the discovered network.",
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network_type": {
							Type:         schema.TypeString,
							Description:  "The type of the network (e.g., Network, DistributedVirtualPortgroup, OpaqueNetwork)",
							Optional:     true,
							ValidateFunc: validation.StringInSlice(network.Type, false),
						},
					},
				},
			},
			"retry_timeout": {
				Type:        schema.TypeInt,
				Description: "Timeout (in seconds) if network is not present yet",
				Optional:    true,
				Default:     0,
			},
			"retry_interval": {
				Type:        schema.TypeInt,
				Description: "Retry interval (in milliseconds) when probing the network",
				Optional:    true,
				Default:     500,
			},
		},
	}
}

type distributedPortGroupStructure struct {
	VLANID int
}

func expandDistributedPortGroupVlan(d *schema.ResourceData) *distributedPortGroupStructure {
	if v, ok := d.GetOk("vlan_id"); ok {
		return &distributedPortGroupStructure{
			VLANID: v.(int),
		}
	}
	return nil
}

func dataSourceVSphereNetworkRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient

	vlan := expandDistributedPortGroupVlan(d)
	name := d.Get("name").(string)
	dvSwitchUUID := d.Get("distributed_virtual_switch_uuid").(string)
	vpcID := d.Get("vpc_id").(string)
	vpcProjectID := d.Get("vpc_project_id").(string)
	var dc *object.Datacenter
	if dcID, ok := d.GetOk("datacenter_id"); ok {
		var err error
		dc, err = datacenterFromID(client, dcID.(string))
		if err != nil {
			return fmt.Errorf("cannot locate datacenter: %s", err)
		}
	}

	vimClient := client.Client

	// Read filter from the schema.
	filters := make(map[string]string)
	if v, ok := d.GetOk("filter"); ok {
		filterList := v.(*schema.Set).List()
		if len(filterList) > 0 {
			for key, value := range filterList[0].(map[string]interface{}) {
				filters[key] = value.(string)
			}
		}
	}

	readRetryFunc := func() (interface{}, string, error) {
		var net interface{}
		var err error
		if dvSwitchUUID != "" {
			// Handle distributed virtual switch port group
			net, err = network.FromNameAndDVSUuid(client, name, dc, dvSwitchUUID)
			if err != nil {
				var notFoundError *network.NotFoundError
				if errors.As(err, &notFoundError) {
					return struct{}{}, waitForNetworkPending, nil
				}

				return struct{}{}, waitForNetworkError, err
			}
			return net, waitForNetworkCompleted, nil
		} else if vpcID != "" {
			// Handle VPC network
			net, err = network.FromNameAndVPCId(client, name, dc, vpcProjectID, vpcID)
			if err != nil {
				var notFoundError *network.NotFoundError
				if errors.As(err, &notFoundError) {
					return struct{}{}, waitForNetworkPending, nil
				}

				return struct{}{}, waitForNetworkError, err
			}
			return net, waitForNetworkCompleted, nil
		}
		// Handle VLAN-based lookup (Distributed Virtual Port Groups only)
		if vlan != nil {
			ctx := context.Background()
			finder := find.NewFinder(vimClient, false)
			if dc != nil {
				finder.SetDatacenter(dc)
			}

			nets, err := finder.NetworkList(ctx, "*")
			if err != nil {
				return struct{}{}, waitForNetworkError, err
			}

			var matches []object.NetworkReference

			for _, n := range nets {
				dvpg, ok := n.(*object.DistributedVirtualPortgroup)
				if !ok {
					continue
				}

				var pg mo.DistributedVirtualPortgroup
				if err := dvpg.Properties(ctx, dvpg.Reference(), []string{"config.defaultPortConfig"}, &pg); err != nil {
					return struct{}{}, waitForNetworkError, err
				}

				cfg, ok := pg.Config.DefaultPortConfig.(*types.VMwareDVSPortSetting)
				if !ok || cfg.Vlan == nil {
					continue
				}

				vlanSpec, ok := cfg.Vlan.(*types.VmwareDistributedVirtualSwitchVlanIdSpec)
				if !ok {
					continue
				}

				if int(vlanSpec.VlanId) == vlan.VLANID {
					matches = append(matches, dvpg)
				}
			}

			if len(matches) == 0 {
				return struct{}{}, waitForNetworkPending, nil
			}

			if len(matches) > 1 {
				return struct{}{}, waitForNetworkError,
					fmt.Errorf("multiple distributed port groups found with vlan_id %d", vlan.VLANID)
			}

			return matches[0], waitForNetworkCompleted, nil
		}

		// Handle standard switch port group
		net, err = network.FromName(vimClient, name, dc, filters) // Pass the *vim25.Client
		if err != nil {
			var notFoundError *network.NotFoundError
			if errors.As(err, &notFoundError) {
				return struct{}{}, waitForNetworkPending, nil
			}
			return struct{}{}, waitForNetworkError, err
		}
		return net, waitForNetworkCompleted, nil

	}

	var net object.NetworkReference
	var netObj interface{}
	var err error
	var state string

	retryTimeout := d.Get("retry_timeout").(int)
	retryInterval := d.Get("retry_interval").(int)

	if retryTimeout == 0 {
		// no retry
		netObj, state, err = readRetryFunc()
	} else {

		deleteRetry := &resource.StateChangeConf{
			Pending:    []string{waitForNetworkPending},
			Target:     []string{waitForNetworkCompleted},
			Refresh:    readRetryFunc,
			Timeout:    time.Duration(retryTimeout) * time.Second,
			MinTimeout: time.Duration(retryInterval) * time.Millisecond,
		}

		netObj, err = deleteRetry.WaitForState()
	}

	if state == waitForNetworkPending {
		if vlan != nil {
			err = fmt.Errorf("network with vlan_id %d not found", vlan.VLANID)
		} else {
			err = fmt.Errorf("network %s not found", name)
		}
	}

	if err != nil {
		return err
	}
	net = netObj.(object.NetworkReference)

	d.SetId(net.Reference().Value)
	_ = d.Set("type", net.Reference().Type)
	return nil
}
