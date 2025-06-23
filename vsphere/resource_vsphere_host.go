// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"bytes"
	"context"

	// TODO: Transition to crypto/sha256 in next major release.
	"crypto/sha1" //nolint
	"crypto/tls"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/license"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	gtask "github.com/vmware/govmomi/task"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/clustercomputeresource"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/customattribute"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/hostsystem"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/provider"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
)

const defaultHostPort = "443"

var servicesPolicyAllowedValues = []string{
	string(types.HostServicePolicyOff),
	string(types.HostServicePolicyOn),
	string(types.HostServicePolicyAutomatic),
}

func resourceVsphereHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceVsphereHostCreate,
		Read:   resourceVsphereHostRead,
		Update: resourceVsphereHostUpdate,
		Delete: resourceVsphereHostDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"datacenter": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "ID of the vSphere datacenter the host will belong to.",
			},
			"cluster": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of the vSphere cluster the host will belong to.",
				ConflictsWith: []string{"cluster_managed"},
			},
			"cluster_managed": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   "Must be set if host is a member of a managed compute_cluster resource.",
				ConflictsWith: []string{"cluster"},
			},
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "FQDN or IP address of the host.",
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Username of the administration account of the host.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Password of the administration account of the host.",
				Sensitive:   true,
			},
			"thumbprint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Host's certificate SHA-1 thumbprint. If not set then the CA that signed the host's certificate must be trusted.",
			},
			"license": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "License key that will be applied to this host.",
			},
			"force": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Force add the host to the vSphere inventory even if it's already managed by a different vCenter Server instance.",
				Default:     false,
			},
			"connected": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set the state of the host. If set to false then the host will be asked to disconnect.",
				Default:     true,
			},
			"maintenance": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set the host's maintenance mode. Default is false",
				Default:     false,
			},
			"lockdown": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Set the host's lockdown status. Default is disabled. Valid options are 'disabled', 'normal', 'strict'",
				Default:      "disabled",
				ValidateFunc: validation.StringInSlice([]string{"disabled", "normal", "strict"}, true),
			},
			"services": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ntpd": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:        schema.TypeBool,
										Optional:    true,
										Description: "Whether the NTP service is enabled. Default is false.",
									},
									"policy": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(servicesPolicyAllowedValues, false),
										Description:  "The policy for the NTP service. Valid values are 'Start and stop with host', 'Start and stop manually', 'Start and stop with port usage'.",
									},
									"ntp_servers": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},

			// Tagging
			vSphereTagAttributeKey: tagsSchema(),

			// Custom Attributes
			customattribute.ConfigKey: customattribute.ConfigSchema(),
		},
	}
}

func resourceVsphereHostCreate(d *schema.ResourceData, meta interface{}) error {
	err := validateFields(d)
	if err != nil {
		return err
	}

	client := meta.(*Client).vimClient

	hcs, err := buildHostConnectSpec(d)
	if err != nil {
		return fmt.Errorf("failed to build host connect spec: %v", err)
	}

	licenseKey := d.Get("license").(string)

	if licenseKey != "" {
		licFound, err := licenseExists(client.Client, licenseKey)
		if err != nil {
			return fmt.Errorf("error while looking for license key. Error: %s", err)
		}

		if !licFound {
			return fmt.Errorf("license key supplied (%s) did not match against known license keys", licenseKey)
		}
	}

	var connectedState bool
	val := d.Get("connected")
	if val == nil {
		connectedState = true
	} else {
		connectedState = val.(bool)
	}

	var task *object.Task
	clusterID := d.Get("cluster").(string)
	if clusterID != "" {
		ccr, err := clustercomputeresource.FromID(client, clusterID)
		if err != nil {
			return fmt.Errorf("error while searching cluster %s. Error: %s", clusterID, err)
		}

		task, err = ccr.AddHost(context.TODO(), hcs, connectedState, &licenseKey, nil)
		if err != nil {
			return fmt.Errorf("error while adding host with hostname %s to cluster %s.  Error: %s", d.Get("hostname").(string), clusterID, err)
		}
	} else {
		dcID := d.Get("datacenter").(string)
		dc, err := datacenterFromID(client, dcID)
		if err != nil {
			return fmt.Errorf("error while retrieving datacenter object for datacenter: %s. Error: %s", dcID, err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultAPITimeout)
		defer cancel()
		var dcProps mo.Datacenter
		if err := dc.Properties(ctx, dc.Reference(), nil, &dcProps); err != nil {
			return fmt.Errorf("error while retrieving properties for datacenter %s. Error: %s", dcID, err)
		}

		hostFolder := object.NewFolder(client.Client, dcProps.HostFolder)
		task, err = hostFolder.AddStandaloneHost(context.TODO(), hcs, connectedState, &licenseKey, nil)
		if err != nil {
			return fmt.Errorf("error while adding standalone host %s. Error: %s", hcs.HostName, err)
		}
	}

	p := property.DefaultCollector(client.Client)
	res, err := gtask.WaitEx(context.TODO(), task.Reference(), p, nil)
	if err != nil {
		return fmt.Errorf("host addition failed. %s", err)
	}
	taskResult := res.Result

	var hostID string
	taskResultType := taskResult.(types.ManagedObjectReference).Type
	switch taskResultType {
	case "ComputeResource":
		computeResource := object.NewComputeResource(client.Client, taskResult.(types.ManagedObjectReference))
		crHosts, err := computeResource.Hosts(context.TODO())
		if err != nil {
			return fmt.Errorf("failed to retrieve created computeResource Hosts. Error: %s", err)
		}
		hostID = crHosts[0].Reference().Value
		log.Printf("[DEBUG] standalone hostID: %s", hostID)
	case "HostSystem":
		hostID = taskResult.(types.ManagedObjectReference).Value
	default:
		return fmt.Errorf("unexpected task result type encountered. Got %s while waiting ComputeResourceType or Hostsystem", taskResultType)
	}
	log.Printf("[DEBUG] Host added with ID %s", hostID)
	d.SetId(hostID)

	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("failed while retrieving host object for host %s. Error: %s", hostID, err)
	}

	// Load the tags client to validate the vCenter Sever connection before
	// attempting to proceed if tags have been defined.
	tagsClient, err := tagsManagerIfDefined(d, meta)
	if err != nil {
		return err
	}

	// Verify the vCenter Server connection before
	// attempting to proceed if custom attributes have been defined.
	attrsProcessor, err := customattribute.GetDiffProcessorIfAttributesDefined(client, d)
	if err != nil {
		return err
	}

	// Apply tags
	if tagsClient != nil {
		if err := processTagDiff(tagsClient, d, host); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	// Apply custom attributes
	if attrsProcessor != nil {
		if err := attrsProcessor.ProcessDiff(host); err != nil {
			return err
		}
	}

	lockdownModeString := d.Get("lockdown").(string)
	lockdownMode, err := hostLockdownType(lockdownModeString)
	if err != nil {
		return err
	}

	if connectedState {
		hostProps, err := hostsystem.Properties(host)
		if err != nil {
			return fmt.Errorf("error while retrieving properties for host %s. Error: %s", hostID, err)
		}

		hamRef := hostProps.ConfigManager.HostAccessManager.Reference()
		ham := NewHostAccessManager(client.Client, hamRef)
		err = ham.ChangeLockdownMode(context.TODO(), lockdownMode)
		if err != nil {
			return fmt.Errorf("error while changing lockdown mode for host %s. Error: %s", hostID, err)
		}
	}

	maintenanceMode := d.Get("maintenance").(bool)
	if maintenanceMode {
		err = hostsystem.EnterMaintenanceMode(host, provider.DefaultAPITimeout, true)
	} else {
		err = hostsystem.ExitMaintenanceMode(host, provider.DefaultAPITimeout)
	}
	if err != nil {
		return fmt.Errorf("error while toggling maintenance mode for host %s. Error: %s", hostID, err)
	}

	mutableKeys := map[string]func(*schema.ResourceData, interface{}, interface{}, interface{}) error{
		"services": resourceVSphereHostUpdateServices,
	}
	for k, v := range mutableKeys {
		log.Printf("[DEBUG] Checking if key %s changed", k)
		if !d.HasChange(k) {
			continue
		}
		log.Printf("[DEBUG] Key %s has change, processing", k)
		old, newVal := d.GetChange(k)
		err := v(d, meta, old, newVal)
		if err != nil {
			return fmt.Errorf("error while updating %s: %s", k, err)
		}
	}

	return resourceVsphereHostRead(d, meta)
}

func resourceVsphereHostRead(d *schema.ResourceData, meta interface{}) error {
	// NOTE: Destroying the host without telling vsphere about it will result in us not
	// knowing that the host does not exist any more.

	// Look for host
	client := meta.(*Client).vimClient
	hostID := d.Id()

	// Find host and get reference to it.
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error while searching host %s. Error: %s ", hostID, err)
	}

	ctx := context.TODO()
	serviceKey := "ntpd"
	ntpServers, err := readHostNtpServerConfig(ctx, client, hs)
	if err != nil {
		return fmt.Errorf("error while reading NTP configuration for host: %s", err)
	}

	policyConfig, err := readHostServicePolicy(ctx, client, hs, serviceKey)
	if err != nil {
		return fmt.Errorf("error while reading policy configuration for host: %s", err)
	}

	serviceEnabled, err := readHostServiceStatus(ctx, client, hs, serviceKey)
	if err != nil {
		return fmt.Errorf("error while reading service status for host: %s", err)
	}

	ntpdService := map[string]interface{}{
		"ntpd": []interface{}{
			map[string]interface{}{
				"enabled":     serviceEnabled,
				"policy":      policyConfig,
				"ntp_servers": ntpServers,
			},
		},
	}

	// Set this structure under the "services" key in the resource data
	if err := d.Set("services", []interface{}{ntpdService}); err != nil {
		return fmt.Errorf("error setting services: %s", err)
	}

	maintenanceState, err := hostsystem.HostInMaintenance(hs)
	if err != nil {
		return fmt.Errorf("error while checking maintenance status for host %s. Error: %s", hostID, err)
	}
	_ = d.Set("maintenance", maintenanceState)

	// Retrieve host's properties.
	log.Printf("[DEBUG] Got host %s", hs.String())
	host, err := hostsystem.Properties(hs)
	if err != nil {
		return fmt.Errorf("error while retrieving properties for host %s. Error: %s", hostID, err)
	}

	if host.Parent != nil && host.Parent.Type == "ClusterComputeResource" && !d.Get("cluster_managed").(bool) {
		_ = d.Set("cluster", host.Parent.Value)
	} else {
		_ = d.Set("cluster", "")
	}

	connectionState, err := hostsystem.GetConnectionState(hs)
	if err != nil {
		return fmt.Errorf("error while getting connection state for host %s. Error: %s", hostID, err)
	}

	if connectionState == types.HostSystemConnectionStateDisconnected {
		// Config and LicenseManager cannot be used while the host is
		// disconnected.
		_ = d.Set("connected", false)
		return nil
	}
	_ = d.Set("connected", true)

	lockdownMode, err := hostLockdownString(host.Config.LockdownMode)
	if err != nil {
		return err
	}

	log.Printf("Setting lockdown to %s", lockdownMode)
	_ = d.Set("lockdown", lockdownMode)

	licenseKey := d.Get("license").(string)
	if licenseKey != "" {
		licFound, err := isLicenseAssigned(client.Client, hostID, licenseKey)
		if err != nil {
			return fmt.Errorf("error while checking license assignment for host %s. Error: %s", hostID, err)
		}

		if !licFound {
			_ = d.Set("license", "")
		}
	}

	// Read tags
	if tagsClient, _ := meta.(*Client).TagsManager(); tagsClient != nil {
		if err := readTagsForResource(tagsClient, host, d); err != nil {
			return fmt.Errorf("error reading tags: %s", err)
		}
	}

	// Read custom attributes
	if customattribute.IsSupported(client) {
		moHost, err := hostsystem.Properties(hs)
		if err != nil {
			return err
		}
		customattribute.ReadFromResource(moHost.Entity(), d)
	}

	return nil
}

func resourceVsphereHostUpdate(d *schema.ResourceData, meta interface{}) error {
	err := validateFields(d)
	if err != nil {
		return err
	}

	client := meta.(*Client).vimClient

	tagsClient, err := tagsManagerIfDefined(d, meta)
	if err != nil {
		return err
	}

	attrsProcessor, err := customattribute.GetDiffProcessorIfAttributesDefined(client, d)
	if err != nil {
		return err
	}

	// First let's establish where we are and where we want to go
	var desiredConnectionState bool
	if d.HasChange("connected") {
		_, newVal := d.GetChange("connected")
		desiredConnectionState = newVal.(bool)
	} else {
		desiredConnectionState = d.Get("connected").(bool)
	}

	hostID := d.Id()
	hostObject, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem object for host ID %s. Error: %s", hostID, err)
	}

	actualConnectionState, err := hostsystem.GetConnectionState(hostObject)
	if err != nil {
		return fmt.Errorf("error while retrieving connection state for host %s. Error: %s", hostID, err)
	}

	// Have there been any changes that warrant a reconnect?
	reconnect := false
	connectionKeys := []string{"hostname", "username", "password", "thumbprint"}
	for _, k := range connectionKeys {
		if d.HasChange(k) {
			reconnect = true
			break
		}
	}

	// Decide if we're going to reconnect or not
	reconnectNeeded, err := shouldReconnect(d, meta, actualConnectionState, desiredConnectionState, reconnect)
	if err != nil {
		return err
	}

	switch reconnectNeeded {
	case 1:
		err := resourceVSphereHostReconnect(d, meta)
		if err != nil {
			return fmt.Errorf("error while reconnecting host %s. Error: %s", hostID, err)
		}
	case -1:
		err := resourceVSphereHostDisconnect(d, meta)
		if err != nil {
			return fmt.Errorf("error while disconnecting host %s. Error: %s", hostID, err)
		}
	case 0:
		break
	}

	mutableKeys := map[string]func(*schema.ResourceData, interface{}, interface{}, interface{}) error{
		"license":     resourceVSphereHostUpdateLicense,
		"cluster":     resourceVSphereHostUpdateCluster,
		"maintenance": resourceVSphereHostUpdateMaintenanceMode,
		"lockdown":    resourceVSphereHostUpdateLockdownMode,
		"thumbprint":  resourceVSphereHostUpdateThumbprint,
		"services":    resourceVSphereHostUpdateServices,
	}
	for k, v := range mutableKeys {
		log.Printf("[DEBUG] Checking if key %s changed", k)
		if !d.HasChange(k) {
			continue
		}
		log.Printf("[DEBUG] Key %s has change, processing", k)
		old, newVal := d.GetChange(k)
		err := v(d, meta, old, newVal)
		if err != nil {
			return fmt.Errorf("error while updating %s: %s", k, err)
		}
	}

	// Apply tags
	if tagsClient != nil {
		if err := processTagDiff(tagsClient, d, hostObject); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	// Apply custom attributes
	if attrsProcessor != nil {
		if err := attrsProcessor.ProcessDiff(hostObject); err != nil {
			return err
		}
	}

	return resourceVsphereHostRead(d, meta)
}

func resourceVsphereHostDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Id()

	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem object for host ID %s. Error: %s", hostID, err)
	}

	connectionState, err := hostsystem.GetConnectionState(hs)
	if err != nil {
		return fmt.Errorf("error while retrieving connection state for host %s. Error: %s", hostID, err)
	}

	if connectionState != types.HostSystemConnectionStateDisconnected {
		// We cannot put a disconnected server in maintenance mode.
		err = resourceVSphereHostDisconnect(d, meta)
		if err != nil {
			return fmt.Errorf("error while disconnecting host: %s", err.Error())
		}
	}

	hostProps, err := hostsystem.Properties(hs)
	if err != nil {
		return fmt.Errorf("error while retrieving properties fort host %s. Error: %s", hostID, err)
	}

	// If this is a standalone host we need to destroy the ComputeResource object
	// and not the Hostsystem itself.
	var task *object.Task
	if hostProps.Parent.Type == "ComputeResource" {
		cr := object.NewComputeResource(client.Client, *hostProps.Parent)
		task, err = cr.Destroy(context.TODO())
		if err != nil {
			return fmt.Errorf("error while submitting destroy task for compute resource %s. Error: %s", hostProps.Parent.Value, err)
		}
	} else {
		task, err = hs.Destroy(context.TODO())
		if err != nil {
			return fmt.Errorf("error while submitting destroy task for host system %s. Error: %s", hostProps.Parent.Value, err)
		}
	}
	p := property.DefaultCollector(client.Client)
	_, err = gtask.WaitEx(context.TODO(), task.Reference(), p, nil)
	if err != nil {
		return fmt.Errorf("error while waiting for host (%s) to be removed: %s", hostID, err)
	}
	return nil
}

func resourceVSphereHostUpdateLockdownMode(d *schema.ResourceData, meta, _, newVal interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Id()
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem object for host ID %s. Error: %s", hostID, err)
	}
	lockdownModeString := newVal.(string)
	lockdownMode, err := hostLockdownType(lockdownModeString)
	if err != nil {
		return err
	}

	var hostProps mo.HostSystem
	err = host.Properties(context.TODO(), host.ConfigManager().Reference(), []string{"configManager.hostAccessManager"}, &hostProps)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem properties for host ID %s. Error: %s", hostID, err)
	}

	hamRef := hostProps.ConfigManager.HostAccessManager.Reference()
	ham := NewHostAccessManager(client.Client, hamRef)
	err = ham.ChangeLockdownMode(context.TODO(), lockdownMode)
	if err != nil {
		return fmt.Errorf("error while changing lonckdown mode for host ID %s to %s. Error: %s", hostID, lockdownMode, err)
	}

	return nil
}

func resourceVSphereHostUpdateMaintenanceMode(d *schema.ResourceData, meta, _, newVal interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Id()

	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem object for host ID %s. Error: %s", hostID, err)
	}

	maintenanceMode := newVal.(bool)
	if maintenanceMode {
		err = hostsystem.EnterMaintenanceMode(host, provider.DefaultAPITimeout, true)
	} else {
		err = hostsystem.ExitMaintenanceMode(host, provider.DefaultAPITimeout)
	}
	if err != nil {
		return fmt.Errorf("error while toggling maintenance mode for host %s. Error: %s", host.Name(), err)
	}
	return nil
}

func resourceVSphereHostUpdateLicense(d *schema.ResourceData, meta, _, newVal interface{}) error {
	client := meta.(*Client).vimClient
	lm := license.NewManager(client.Client)
	lam, err := lm.AssignmentManager(context.TODO())
	if err != nil {
		return fmt.Errorf("error while accessing License Assignment Manager endpoint. Error: %s", err)
	}
	_, err = lam.Update(context.TODO(), d.Id(), newVal.(string), "")
	if err != nil {
		return fmt.Errorf("error while updating license. error: %s", err)
	}
	return nil
}

func resourceVSphereHostUpdateCluster(d *schema.ResourceData, meta, _, newVal interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Id()
	newClusterID := newVal.(string)

	newCluster, err := clustercomputeresource.FromID(client, newClusterID)
	if err != nil {
		return fmt.Errorf("error while searching newVal cluster %s. Error: %s", newClusterID, err)
	}

	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem object for host ID %s. Error: %s", hostID, err)
	}

	err = hostsystem.EnterMaintenanceMode(hs, provider.DefaultAPITimeout, true)
	if err != nil {
		return fmt.Errorf("error while putting host to maintenance mode: %s", err.Error())
	}

	task, err := newCluster.MoveInto(context.TODO(), hs)
	if err != nil {
		return fmt.Errorf("error while moving HostSystem with ID %s to new cluster. Error: %s", hostID, err)
	}
	p := property.DefaultCollector(client.Client)
	_, err = gtask.WaitEx(context.TODO(), task.Reference(), p, nil)
	if err != nil {
		return fmt.Errorf("error while moving host to new cluster (%s): %s", newClusterID, err)
	}

	err = hostsystem.ExitMaintenanceMode(hs, provider.DefaultAPITimeout)
	if err != nil {
		return fmt.Errorf("error while taking host out of maintenance mode: %s", err.Error())
	}

	return nil
}

func resourceVSphereHostUpdateThumbprint(d *schema.ResourceData, meta, _, _ interface{}) error {
	return resourceVSphereHostReconnect(d, meta)
}

func resourceVSphereHostReconnect(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Id()
	client := meta.(*Client).vimClient
	host := object.NewHostSystem(client.Client, types.ManagedObjectReference{Type: "HostSystem", Value: d.Id()})
	hcs, err := buildHostConnectSpec(d)
	if err != nil {
		return fmt.Errorf("failed to build host connect spec: %v", err)
	}

	task, err := host.Reconnect(context.TODO(), &hcs, nil)
	if err != nil {
		return fmt.Errorf("error while reconnecting host with ID %s. Error: %s", hostID, err)
	}

	p := property.DefaultCollector(client.Client)
	_, err = gtask.WaitEx(context.TODO(), task.Reference(), p, nil)
	if err != nil {
		return fmt.Errorf("error while reconnecting host(%s): %s", hostID, err)
	}

	maintenanceState, err := hostsystem.HostInMaintenance(host)
	if err != nil {
		return fmt.Errorf("error while retrieving host maintenance status for host %s. Error: %s", host.Name(), err)
	}

	maintenanceConfig := d.Get("maintenance").(bool)
	if maintenanceState && !maintenanceConfig {
		err := hostsystem.ExitMaintenanceMode(host, provider.DefaultAPITimeout)
		if err != nil {
			return fmt.Errorf("error while taking host %s out of maintenance mode. Error: %s", host.Name(), err)
		}
	}
	return nil
}

func resourceVSphereHostDisconnect(d *schema.ResourceData, meta interface{}) error {
	hostID := d.Id()
	client := meta.(*Client).vimClient
	host := object.NewHostSystem(client.Client, types.ManagedObjectReference{Type: "HostSystem", Value: d.Id()})
	task, err := host.Disconnect(context.TODO())
	if err != nil {
		return fmt.Errorf("error while disconnecting host %s. Error: %s", host.Name(), err)
	}

	p := property.DefaultCollector(client.Client)
	_, err = gtask.WaitEx(context.TODO(), task.Reference(), p, nil)
	if err != nil {
		return fmt.Errorf("error while disconnecting host(%s): %s", hostID, err)
	}
	return nil
}

func shouldReconnect(_ *schema.ResourceData, _ interface{}, actual types.HostSystemConnectionState, desired, shouldReconnect bool) (int, error) {
	log.Printf("[DEBUG] Figuring out if we need to do something about the host's connection")

	// desired state is connected and one of the connectionKeys has changed
	if shouldReconnect && desired {
		log.Printf("[DEBUG] Desired state is connected and one of the settings relevant to the connection changed. Reconnecting")
		return 1, nil
	}

	// desired state is connected and actual state is disconnected
	if desired && (actual != types.HostSystemConnectionStateConnected) {
		log.Printf("[DEBUG] Desired state is connected but host is not connected. Reconnecting")
		return 1, nil
	}

	// desired state is connected and actual state is connected (or host is missing heartbeats) and
	// none of the connectionKeys have changed.
	if desired && (actual != types.HostSystemConnectionStateDisconnected) && !shouldReconnect {
		log.Printf("[DEBUG] Desired state is connected and host is connected and no changes in config. Noop")
		return 0, nil
	}

	// desired state is disconnected and host is disconnected
	if !desired && (actual == types.HostSystemConnectionStateDisconnected) {
		log.Printf("[DEBUG] Desired state is disconnected and host is disconnected")
		return 0, nil
	}

	if !desired && (actual != types.HostSystemConnectionStateDisconnected) {
		log.Printf("[DEBUG] Desired state is disconnected but host is not disconnected. Disconnecting")
		return -1, nil
	}

	log.Printf("[DEBUG] Unexpected combination of desired and actual states, not sure how to handle. Please submit a bug report.")
	return 255, fmt.Errorf("unexpected combination of connection states")
}

func hostLockdownType(lockdownMode string) (types.HostLockdownMode, error) {
	lockdownModes := map[string]types.HostLockdownMode{
		"disabled": types.HostLockdownModeLockdownDisabled,
		"normal":   types.HostLockdownModeLockdownNormal,
		"strict":   types.HostLockdownModeLockdownStrict,
	}

	log.Printf("Looking for mode %s in lockdown modes %#v", lockdownMode, lockdownModes)
	if modeString, ok := lockdownModes[lockdownMode]; ok {
		log.Printf("Found match for %s. Returning %s.", lockdownMode, modeString)
		return modeString, nil
	}
	return "", fmt.Errorf("unknown Lockdown mode encountered")
}

func hostLockdownString(lockdownMode types.HostLockdownMode) (string, error) {
	lockdownModes := map[types.HostLockdownMode]string{
		types.HostLockdownModeLockdownDisabled: "disabled",
		types.HostLockdownModeLockdownNormal:   "normal",
		types.HostLockdownModeLockdownStrict:   "strict",
	}

	log.Printf("Looking for mode %s in lockdown modes %#v", lockdownMode, lockdownModes)
	if modeString, ok := lockdownModes[lockdownMode]; ok {
		log.Printf("Found match for %s. Returning %s.", lockdownMode, modeString)
		return modeString, nil
	}
	return "", fmt.Errorf("unknown Lockdown mode encountered")
}

func buildHostConnectSpec(d *schema.ResourceData) (types.HostConnectSpec, error) {
	thumbprint := d.Get("thumbprint").(string)
	hostname := d.Get("hostname").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	log.Printf("Building HostConnectSpec for host: %s", hostname)
	// Retrieve the actual thumbprint from the ESXi host.
	if thumbprint != "" {
		actualThumbprint, err := getHostThumbprint(d)
		if err != nil {
			return types.HostConnectSpec{}, fmt.Errorf("error retrieving host thumbprint: %s", err)
		}

		// Compare the the provided and returned thumbprints.
		if thumbprint != actualThumbprint {
			return types.HostConnectSpec{}, fmt.Errorf("thumbprint mismatch: expected %s, got %s", thumbprint, actualThumbprint)
		}
	}

	hcs := types.HostConnectSpec{
		HostName:      hostname,
		UserName:      username,
		Password:      password,
		SslThumbprint: thumbprint,
		Force:         d.Get("force").(bool),
	}
	return hcs, nil
}

func getHostThumbprint(d *schema.ResourceData) (string, error) {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2 or higher.
	}

	// Check the hostname.
	address, ok := d.Get("hostname").(string)
	if !ok {
		return "", fmt.Errorf("hostname field is not a string or is nil")
	}

	// Default port for HTTPS.
	port := defaultHostPort
	if p, ok := d.GetOk("port"); ok {
		if portStr, ok := p.(string); ok {
			port = portStr
		} else {
			return "", fmt.Errorf("port field is not a string")
		}
	}

	// Check if allow_unverified_ssl is true. If so, skip the verification.
	// Otherwise, use the default value of false.
	if thumbprint, ok := d.Get("thumbprint").(string); ok && thumbprint != "" {
		return thumbprint, nil
	}

	config.InsecureSkipVerify = false

	if insecure, ok := d.GetOk("allow_unverified_ssl"); ok {
		if insecureBool, typeOk := insecure.(bool); typeOk {
			config.InsecureSkipVerify = insecureBool
		}
	}

	conn, err := tls.Dial("tcp", address+":"+port, config)
	if err != nil {
		return "", fmt.Errorf("error dialing TLS connection: %w", err)
	}
	defer conn.Close()

	cert := conn.ConnectionState().PeerCertificates[0]
	fingerprint := sha1.Sum(cert.Raw)

	var buf bytes.Buffer
	for i, f := range fingerprint {
		if i > 0 {
			_, _ = fmt.Fprintf(&buf, ":")
		}
		_, _ = fmt.Fprintf(&buf, "%02X", f)
	}
	return buf.String(), nil
}

func isLicenseAssigned(client *vim25.Client, hostID, licenseKey string) (bool, error) {
	ctx := context.TODO()
	lm := license.NewManager(client)
	am, err := lm.AssignmentManager(ctx)
	if err != nil {
		return false, err
	}

	licenses, err := am.QueryAssigned(ctx, hostID)
	if err != nil {
		return false, err
	}

	licFound := false
	for _, lic := range licenses {
		if licenseKey == lic.AssignedLicense.LicenseKey {
			licFound = true
			break
		}
	}
	return licFound, nil
}

func licenseExists(client *vim25.Client, licenseKey string) (bool, error) {
	ctx := context.TODO()
	lm := license.NewManager(client)
	ll, err := lm.List(ctx)
	if err != nil {
		return false, err
	}

	licFound := false
	for _, l := range ll {
		if l.LicenseKey == licenseKey {
			licFound = true
			break
		}
	}
	return licFound, nil
}

// Make sure input makes sense
func validateFields(d *schema.ResourceData) error {
	_, dcSet := d.GetOk("datacenter")
	_, clusterSet := d.GetOk("cluster")
	if dcSet && clusterSet {
		return fmt.Errorf("datacenter and cluster arguments are mutually exclusive")
	}
	return nil
}

type HostAccessManager struct {
	object.Common
}

func NewHostAccessManager(c *vim25.Client, ref types.ManagedObjectReference) *HostAccessManager {
	return &HostAccessManager{
		Common: object.NewCommon(c, ref),
	}
}

func (h HostAccessManager) ChangeLockdownMode(ctx context.Context, mode types.HostLockdownMode) error {
	req := types.ChangeLockdownMode{
		This: h.Reference(),
		Mode: mode,
	}
	_, err := methods.ChangeLockdownMode(ctx, h.Client(), &req)
	return err
}

func resourceVSphereHostUpdateServices(d *schema.ResourceData, meta interface{}, _, _ interface{}) error {
	client := meta.(*Client).vimClient
	hostID := d.Id()
	hostObject, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return fmt.Errorf("error while retrieving HostSystem object for host ID %s. Error: %s", hostID, err)
	}

	servicesSet, ok := d.Get("services").(*schema.Set)
	if !ok {
		return fmt.Errorf("error reading 'services': type assertion to *schema.Set failed")
	}

	updatedServices := make([]interface{}, 0) // Prepare to collect updated services configurations

	services := servicesSet.List()
	for _, service := range services {
		serviceMap := service.(map[string]interface{})
		updatedServiceMap := make(map[string]interface{}) // Prepare to collect updated service configuration

		if ntpd, ok := serviceMap["ntpd"]; ok {
			ntpdConfig := ntpd.([]interface{})[0].(map[string]interface{})
			updatedNtpdConfig := make(map[string]interface{}) // Copy ntpdConfig if needed before modifications

			// Start the NTP service if enabled
			if enabled, ok := ntpdConfig["enabled"].(bool); ok && enabled {
				err := StartHostService(context.Background(), client, hostObject, "ntpd")
				if err != nil {
					return fmt.Errorf("failed to start NTP service on host %s: %v", hostID, err)
				}
			}

			// Update NTP servers
			interfaces := ntpdConfig["ntp_servers"].([]interface{})
			newServers := make([]string, len(interfaces))
			for i, server := range interfaces {
				newServers[i] = server.(string)
			}

			err := changeHostNtpServers(context.Background(), hostObject, newServers)
			if err != nil {
				return fmt.Errorf("error while updating NTP servers for host %s. Error: %s", hostID, err)
			}

			// Update the ntpdConfig map before setting it
			updatedNtpdConfig["enabled"] = ntpdConfig["enabled"]
			updatedNtpdConfig["ntp_servers"] = newServers // Updated servers
			updatedNtpdConfig["policy"] = ntpdConfig["policy"]

			// Update the NTP service policy if applicable
			if policy, ok := ntpdConfig["policy"].(string); ok {
				err := UpdateHostServicePolicy(context.Background(), client, hostObject, "ntpd", policy)
				if err != nil {
					return fmt.Errorf("failed to update NTP service policy on host %s: %v", hostID, err)
				}
			}

			updatedServiceMap["ntpd"] = []interface{}{updatedNtpdConfig}
		}
		// Handle other services similarly and add to updatedServiceMap as needed

		updatedServices = append(updatedServices, updatedServiceMap) // Add the updated service map to the collection
	}

	// After processing all services, set the entire updated services configuration
	if err := d.Set("services", updatedServices); err != nil {
		return fmt.Errorf("error setting updated services configuration: %s", err)
	}

	return nil
}

func changeHostNtpServers(ctx context.Context, host *object.HostSystem, servers []string) error {
	s, err := host.ConfigManager().DateTimeSystem(ctx)
	if err != nil {
		return err
	}

	ntpConfig := types.HostNtpConfig{
		Server: servers,
	}

	dateTimeConfig := types.HostDateTimeConfig{
		NtpConfig: &ntpConfig,
	}

	return s.UpdateConfig(ctx, dateTimeConfig)
}

func readHostNtpServerConfig(ctx context.Context, _ *govmomi.Client, hostObject *object.HostSystem) ([]string, error) {
	// Retrieve the host's configuration
	var hostSystem mo.HostSystem
	err := hostObject.Properties(ctx, hostObject.Reference(), []string{"config.dateTimeInfo"}, &hostSystem)
	if err != nil {
		return nil, fmt.Errorf("failed to get host configuration: %v", err)
	}

	// Check if the dateTimeInfo configuration is available
	if hostSystem.Config == nil || hostSystem.Config.DateTimeInfo == nil {
		return nil, fmt.Errorf("dateTimeInfo configuration is not available on host")
	}

	// Check if the NTP configuration is available
	if hostSystem.Config.DateTimeInfo.NtpConfig == nil {
		return nil, fmt.Errorf("NTP configuration is not available for the host")
	}

	// Log the NTP servers found (optional)
	fmt.Printf("NTP Servers for host: %v\n", hostSystem.Config.DateTimeInfo.NtpConfig.Server)

	// Return the NTP servers
	return hostSystem.Config.DateTimeInfo.NtpConfig.Server, nil
}

// StartHostService starts a specified service on a host.
func StartHostService(ctx context.Context, _ *govmomi.Client, hostObject *object.HostSystem, serviceKey string) error {
	// Retrieve the host's service system
	serviceSystem, err := hostObject.ConfigManager().ServiceSystem(ctx)
	if err != nil {
		return fmt.Errorf("failed to get host service system: %v", err)
	}

	// Directly attempt to start the service without listing all services first
	err = serviceSystem.Start(ctx, serviceKey)
	if err != nil {
		return fmt.Errorf("failed to start service %s: %v", serviceKey, err)
	}
	fmt.Printf("Service %s started successfully on host\n", serviceKey)
	return nil
}

// UpdateHostServicePolicy updates the policy of a specified service on a host.
func UpdateHostServicePolicy(ctx context.Context, _ *govmomi.Client, hostObject *object.HostSystem, serviceKey, policy string) error {
	// Retrieve the host's service system
	serviceSystem, err := hostObject.ConfigManager().ServiceSystem(ctx)
	if err != nil {
		return fmt.Errorf("failed to get host service system: %v", err)
	}

	err = serviceSystem.UpdatePolicy(ctx, serviceKey, policy)
	if err != nil {
		return fmt.Errorf("failed to update policy for service %s: %v", serviceKey, err)
	}
	fmt.Printf("Policy for service %s updated successfully on host\n", serviceKey)
	return nil
}

// ReadHostServicePolicy reads the policy of a specified service on a host.
func readHostServicePolicy(ctx context.Context, _ *govmomi.Client, hostObject *object.HostSystem, serviceKey string) (string, error) {
	// Retrieve the host's configuration
	var hostSystem mo.HostSystem
	err := hostObject.Properties(ctx, hostObject.Reference(), []string{"config.service"}, &hostSystem)
	if err != nil {
		return "", fmt.Errorf("failed to get host configuration: %v", err)
	}

	// Check if the service configuration is available
	if hostSystem.Config == nil || hostSystem.Config.Service == nil {
		return "", fmt.Errorf("service configuration is not available on host")
	}

	// Iterate over the services to find the specified service and its policy
	for _, service := range hostSystem.Config.Service.Service {
		if service.Key == serviceKey {
			// Policy found for the specified service
			fmt.Printf("Policy for service %s is %s\n", serviceKey, service.Policy) // Optional: log the found policy
			return service.Policy, nil
		}
	}

	return "", fmt.Errorf("service %s not found on host", serviceKey)
}

func readHostServiceStatus(ctx context.Context, _ *govmomi.Client, hostObject *object.HostSystem, serviceKey string) (bool, error) {
	// Retrieve the host's configuration
	var hostSystem mo.HostSystem
	err := hostObject.Properties(ctx, hostObject.Reference(), []string{"config.service"}, &hostSystem)
	if err != nil {
		return false, fmt.Errorf("failed to get host configuration: %v", err)
	}

	// Check if the service configuration is available
	if hostSystem.Config == nil || hostSystem.Config.Service == nil {
		return false, fmt.Errorf("service configuration is not available on host")
	}

	// Debug: Log the number of services found
	fmt.Printf("Number of services found: %d\n", len(hostSystem.Config.Service.Service))

	// Iterate over the services to find the NTP service
	for _, service := range hostSystem.Config.Service.Service {
		// Debug: Log each service key encountered
		fmt.Printf("Checking service: %s\n", service.Key)

		if service.Key == serviceKey {
			fmt.Printf("Enabled for service %s is %t\n", serviceKey, service.Running) // Log the running status
			return service.Running, nil
		}
	}

	return false, fmt.Errorf("NTP service not found on host")
}
