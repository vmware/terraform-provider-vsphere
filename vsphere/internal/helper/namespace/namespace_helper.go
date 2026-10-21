// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package namespace

import (
	"errors"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

const (
	networkProviderNsxVpc  = "NSX_VPC"
	networkProviderVsphere = "VSPHERE_NETWORK"
)

func ValidateVersionRequirements(version viapi.VSphereVersion, d *schema.ResourceDiff) error {
	log.Printf("[DEBUG] namespace.ValidateVersionRequirements: Validating version requirements against vCenter %d.%d.%d", version.Major, version.Minor, version.Patch)
	var errs []error
	require := func(feature string, major, minor, patch int) {
		required := viapi.VSphereVersion{Product: version.Product, Major: major, Minor: minor, Patch: patch}
		if !version.AtLeast(required) {
			errs = append(errs, fmt.Errorf("%s requires vCenter %d.%d.%d or later (connected to %d.%d.%d)",
				feature, major, minor, patch, version.Major, version.Minor, version.Patch))
		}
	}

	if d.Get("zone").(*schema.Set).Len() > 0 {
		require("zone", 9, 0, 0)
	}

	if vpcNetworks := d.Get("vpc_network").([]interface{}); len(vpcNetworks) > 0 {
		require("vpc_network", 9, 0, 0)

		usesSharedSubnetPath, usesSharedSubnetDefaults := false, false
		for _, subnet := range sharedSubnetsFromVpcNetworks(vpcNetworks) {
			if subnet["path"].(string) != "" {
				usesSharedSubnetPath = true
			}
			if subnet["pod_default"].(bool) || subnet["vm_default"].(bool) {
				usesSharedSubnetDefaults = true
			}
		}
		if usesSharedSubnetPath {
			require("vpc_network.shared_subnet.path", 9, 1, 0)
		}
		if usesSharedSubnetDefaults {
			require("vpc_network.shared_subnet pod_default and vm_default", 9, 1, 1)
		}
	}

	if vsphereNetworks := d.Get("vsphere_network").([]interface{}); len(vsphereNetworks) > 0 {
		require("vsphere_network", 9, 1, 0)
		if len(errs) == 0 {
			if err := validateDefaultWorkloadNetwork(d); err != nil {
				errs = append(errs, err)
			}
		}
	}

	return errors.Join(errs...)
}

func validateDefaultWorkloadNetwork(d *schema.ResourceDiff) error {
	if !d.NewValueKnown("vsphere_network.0.networks") || !d.NewValueKnown("vsphere_network.0.default_workload_network") {
		log.Printf("[DEBUG] namespacehelper.validateDefaultWorkloadNetwork: Networks are not known yet, skipping validation")
		return nil
	}
	defaultNetwork := d.Get("vsphere_network.0.default_workload_network").(string)
	networks, _ := d.Get("vsphere_network.0.networks").(*schema.Set)
	if networks == nil || !networks.Contains(defaultNetwork) {
		return fmt.Errorf("vsphere_network.networks must contain the default_workload_network %q", defaultNetwork)
	}
	return nil
}

func sharedSubnetsFromVpcNetworks(vpcNetworks []interface{}) []map[string]interface{} {
	if len(vpcNetworks) == 0 {
		return nil
	}
	vpcNetwork, _ := vpcNetworks[0].(map[string]interface{})
	subnets, _ := vpcNetwork["shared_subnet"].(*schema.Set)
	if subnets == nil {
		return nil
	}
	result := make([]map[string]interface{}, 0, subnets.Len())
	log.Printf("[DEBUG] namespace.sharedSubnetsFromVpcNetworks: Found %d shared subnets", subnets.Len())
	for _, subnet := range subnets.List() {
		result = append(result, subnet.(map[string]interface{}))
	}
	return result
}

func ExpandStorageSpecs(d *schema.ResourceData) []namespace.StorageSpec {
	storage := d.Get("storage_policy").(*schema.Set).List()
	log.Printf("[DEBUG] namespace.ExpandStorageSpecs: Expanding %d storage policies", len(storage))
	specs := make([]namespace.StorageSpec, 0, len(storage))
	for _, item := range storage {
		entry := item.(map[string]interface{})
		spec := namespace.StorageSpec{Policy: entry["policy"].(string)}
		if limit := int64(entry["limit"].(int)); limit != 0 {
			spec.Limit = &limit
		}
		specs = append(specs, spec)
	}
	return specs
}

func FlattenStorageSpecs(specs []namespace.StorageSpec) []interface{} {
	log.Printf("[DEBUG] namespace.FlattenStorageSpecs: Flattening %d storage policies", len(specs))
	result := make([]interface{}, 0, len(specs))
	for _, spec := range specs {
		limit := 0
		if spec.Limit != nil {
			limit = int(*spec.Limit)
		}
		result = append(result, map[string]interface{}{
			"policy": spec.Policy,
			"limit":  limit,
		})
	}
	return result
}

func ExpandAccessList(d *schema.ResourceData) []namespace.Access {
	entries := d.Get("access").(*schema.Set).List()
	log.Printf("[DEBUG] namespace.ExpandAccessList: Expanding %d access entries", len(entries))
	accessList := make([]namespace.Access, 0, len(entries))
	for _, item := range entries {
		entry := item.(map[string]interface{})
		access := namespace.Access{
			SubjectType: namespace.AccessSubjectType(entry["subject_type"].(string)),
			Subject:     entry["subject"].(string),
			Domain:      entry["domain"].(string),
			Role:        namespace.AccessRole(entry["role"].(string)),
		}
		if identityProvider := entry["identity_provider"].(string); identityProvider != "" {
			access.IdentityProvider = &identityProvider
		}
		accessList = append(accessList, access)
	}
	return accessList
}

func FlattenAccessList(accessList []namespace.Access) []interface{} {
	log.Printf("[DEBUG] namespace.FlattenAccessList: Flattening %d access entries", len(accessList))
	result := make([]interface{}, 0, len(accessList))
	for _, access := range accessList {
		identityProvider := ""
		if access.IdentityProvider != nil {
			identityProvider = *access.IdentityProvider
		}
		result = append(result, map[string]interface{}{
			"subject_type":      string(access.SubjectType),
			"subject":           access.Subject,
			"domain":            access.Domain,
			"role":              string(access.Role),
			"identity_provider": identityProvider,
		})
	}
	return result
}

func ExpandZones(d *schema.ResourceData) []namespace.ZoneSpec {
	entries := d.Get("zone").(*schema.Set).List()
	log.Printf("[DEBUG] namespace.ExpandZones: Expanding %d zones", len(entries))
	zones := make([]namespace.ZoneSpec, 0, len(entries))
	for _, item := range entries {
		entry := item.(map[string]interface{})
		zone := namespace.ZoneSpec{Name: entry["name"].(string)}
		if cpuLimit := int64(entry["cpu_limit"].(int)); cpuLimit != 0 {
			zone.CpuLimitMHz = &cpuLimit
		}
		if cpuReservation := int64(entry["cpu_reservation"].(int)); cpuReservation != 0 {
			zone.CpuReservationMHz = &cpuReservation
		}
		if memoryLimit := int64(entry["memory_limit"].(int)); memoryLimit != 0 {
			zone.MemoryLimitMiB = &memoryLimit
		}
		if memoryReservation := int64(entry["memory_reservation"].(int)); memoryReservation != 0 {
			zone.MemoryReservationMiB = &memoryReservation
		}
		if disallow := entry["disallow_unreserved_direct_path_usage"].(bool); disallow {
			zone.DisallowUnreservedDirectPathUsage = &disallow
		}
		if reservations := expandVMReservations(entry["vm_reservation"]); len(reservations) > 0 {
			zone.VmReservations = &reservations
		}
		zones = append(zones, zone)
	}
	return zones
}

func FlattenZones(d *schema.ResourceData, zones []namespace.ZoneInfo) []interface{} {
	log.Printf("[DEBUG] namespace.FlattenZones: Flattening %d zones", len(zones))
	priorByName := make(map[string]map[string]interface{})
	if prior, ok := d.Get("zone").(*schema.Set); ok {
		for _, item := range prior.List() {
			entry := item.(map[string]interface{})
			priorByName[entry["name"].(string)] = entry
		}
	}

	var zoneSettingKeys = []string{
		"cpu_limit",
		"cpu_reservation",
		"memory_limit",
		"memory_reservation",
		"disallow_unreserved_direct_path_usage",
	}

	result := make([]interface{}, 0, len(zones))
	for _, zone := range zones {
		if zone.MarkedForRemoval {
			log.Printf("[DEBUG] namespace.FlattenZones: Skipping zone %s, which is marked for removal", zone.Zone)
			continue
		}
		flattened := map[string]interface{}{"name": zone.Zone}
		if prior, found := priorByName[zone.Zone]; found {
			for _, key := range zoneSettingKeys {
				flattened[key] = prior[key]
			}
			if reservations, ok := prior["vm_reservation"].(*schema.Set); ok {
				flattened["vm_reservation"] = reservations.List()
			}
		}
		result = append(result, flattened)
	}
	return result
}

func expandVMReservations(value interface{}) []namespace.VirtualMachineClassAllocationSpec {
	set, _ := value.(*schema.Set)
	if set == nil {
		return nil
	}
	reservations := make([]namespace.VirtualMachineClassAllocationSpec, 0, set.Len())
	log.Printf("[DEBUG] namespace.expandVMReservations: Expanding %d VM reservations", set.Len())
	for _, item := range set.List() {
		entry := item.(map[string]interface{})
		reservations = append(reservations, namespace.VirtualMachineClassAllocationSpec{
			ReservedVmClass: entry["reserved_vm_class"].(string),
			Count:           int64(entry["count"].(int)),
		})
	}
	return reservations
}

func FlattenZoneInfos(zones []namespace.ZoneInfo) []interface{} {
	log.Printf("[DEBUG] namespace.FlattenZoneInfos: Flattening %d zones", len(zones))
	result := make([]interface{}, 0, len(zones))
	for _, zone := range zones {
		result = append(result, map[string]interface{}{
			"name": zone.Zone,
			"type": zone.Type,
		})
	}
	return result
}

func ExpandNetworkSpecCreate(d *schema.ResourceData) *namespace.NetworkConfigCreateSpec {
	if len(d.Get("vsphere_network").([]interface{})) > 0 {
		return expandVsphereNetworkCreate(d)
	}
	if vpcNetworks := d.Get("vpc_network").([]interface{}); len(vpcNetworks) > 0 {
		if vpcNetwork, ok := vpcNetworks[0].(map[string]interface{}); ok {
			return expandVpcNetworkCreate(vpcNetwork)
		}
	}
	log.Printf("[DEBUG] namespace.ExpandNetworkSpecCreate: No network configuration specified, the Supervisor default is used")
	return nil
}

func expandVsphereNetworkCreate(d *schema.ResourceData) *namespace.NetworkConfigCreateSpec {
	defaultNetwork := d.Get("vsphere_network.0.default_workload_network").(string)
	log.Printf("[DEBUG] namespacehelper.expandVsphereNetworkCreate: Expanding vSphere network configuration with default workload network %s", defaultNetwork)
	return &namespace.NetworkConfigCreateSpec{
		NetworkProvider: networkProviderVsphere,
		VsphereNetwork: &namespace.VsphereNetworkConfigCreateSpec{
			DefaultWorkloadNetwork: &defaultNetwork,
		},
	}
}

func expandVpcNetworkCreate(vpcNetwork map[string]interface{}) *namespace.NetworkConfigCreateSpec {
	vpcPath := vpcNetwork["vpc"].(string)
	vpc := &namespace.VpcNetworkCreateSpec{Vpc: &vpcPath}
	log.Printf("[DEBUG] namespace.expandVpcNetworkCreate: Expanding NSX VPC network configuration")
	if subnets := expandSharedSubnets(vpcNetwork["shared_subnet"]); len(subnets) > 0 {
		vpc.SharedSubnets = &subnets
	}

	return &namespace.NetworkConfigCreateSpec{
		NetworkProvider: networkProviderNsxVpc,
		VpcNetwork:      vpc,
	}
}

func ExpandNetworks(d *schema.ResourceData) []string {
	if len(d.Get("vsphere_network").([]interface{})) == 0 {
		return nil
	}
	defaultNetwork := d.Get("vsphere_network.0.default_workload_network").(string)
	networks := []string{defaultNetwork}
	for _, network := range d.Get("vsphere_network.0.networks").(*schema.Set).List() {
		if network.(string) != defaultNetwork {
			networks = append(networks, network.(string))
		}
	}
	log.Printf("[DEBUG] namespacehelper.ExpandNetworks: Expanded %d networks with default workload network %s", len(networks), defaultNetwork)
	return networks
}

func ExpandNetworkSpecUpdate(d *schema.ResourceData) *namespace.NetworkConfigUpdateSpec {
	if d.HasChange("vsphere_network.0.networks") && len(d.Get("vsphere_network").([]interface{})) > 0 {
		return expandVsphereNetworkUpdate(d)
	}
	vpcNetworks := d.Get("vpc_network").([]interface{})
	if len(vpcNetworks) == 0 || !d.HasChange("vpc_network") {
		log.Printf("[DEBUG] namespace.ExpandNetworkSpecUpdate: No NSX VPC network changes")
		return nil
	}
	vpcNetwork, ok := vpcNetworks[0].(map[string]interface{})
	if !ok {
		return nil
	}

	vpc := &namespace.VpcNetworkUpdateSpec{}
	log.Printf("[DEBUG] namespace.ExpandNetworkSpecUpdate: Expanding NSX VPC network changes")
	subnets := expandSharedSubnets(vpcNetwork["shared_subnet"])
	vpc.SharedSubnets = &subnets

	return &namespace.NetworkConfigUpdateSpec{
		NetworkProvider: networkProviderNsxVpc,
		VpcConfig:       vpc,
	}
}

func expandVsphereNetworkUpdate(d *schema.ResourceData) *namespace.NetworkConfigUpdateSpec {
	oldValue, newValue := d.GetChange("vsphere_network.0.networks")
	oldNetworks, _ := oldValue.(*schema.Set)
	newNetworks, _ := newValue.(*schema.Set)
	if oldNetworks == nil {
		oldNetworks = &schema.Set{F: schema.HashString}
	}
	if newNetworks == nil {
		newNetworks = &schema.Set{F: schema.HashString}
	}
	log.Printf("[DEBUG] namespacehelper.expandVsphereNetworkUpdate: Adding %d and removing %d networks", newNetworks.Difference(oldNetworks).Len(), oldNetworks.Difference(newNetworks).Len())

	return &namespace.NetworkConfigUpdateSpec{
		NetworkProvider: networkProviderVsphere,
		VsphereNetworkConfig: &namespace.VsphereNetworkConfigUpdateSpec{
			NetworksToAdd:    structure.SliceInterfacesToStrings(newNetworks.Difference(oldNetworks).List()),
			NetworksToRemove: structure.SliceInterfacesToStrings(oldNetworks.Difference(newNetworks).List()),
		},
	}
}

func expandSharedSubnets(value interface{}) []namespace.SharedSubnet {
	set, _ := value.(*schema.Set)
	if set == nil {
		return []namespace.SharedSubnet{}
	}
	subnets := make([]namespace.SharedSubnet, 0, set.Len())
	log.Printf("[DEBUG] namespace.expandSharedSubnets: Expanding %d shared subnets", set.Len())
	for _, item := range set.List() {
		entry := item.(map[string]interface{})
		subnet := namespace.SharedSubnet{Path: entry["path"].(string)}
		if podDefault := entry["pod_default"].(bool); podDefault {
			subnet.PodDefault = &podDefault
		}
		if vmDefault := entry["vm_default"].(bool); vmDefault {
			subnet.VmDefault = &vmDefault
		}
		subnets = append(subnets, subnet)
	}
	return subnets
}

func FlattenVpcNetwork(info *namespace.NetworkConfigInfo) []interface{} {
	if info == nil || info.VpcNetwork == nil {
		return nil
	}
	vpc := info.VpcNetwork
	log.Printf("[DEBUG] namespace.FlattenVpcNetwork: Flattening NSX VPC network configuration for VPC %s", vpc.Vpc)

	subnets := make([]interface{}, 0, len(vpc.SharedSubnets))
	for _, subnet := range vpc.SharedSubnets {
		flattened := map[string]interface{}{"path": subnet.Path}
		if subnet.PodDefault != nil {
			flattened["pod_default"] = *subnet.PodDefault
		}
		if subnet.VmDefault != nil {
			flattened["vm_default"] = *subnet.VmDefault
		}
		subnets = append(subnets, flattened)
	}

	vpcNetwork := map[string]interface{}{
		"vpc":           vpc.Vpc,
		"shared_subnet": subnets,
	}
	return []interface{}{vpcNetwork}
}

func FlattenVsphereNetwork(info *namespace.NetworkConfigInfo) []interface{} {
	if info == nil || info.VsphereNetwork == nil {
		return nil
	}

	networks := make([]interface{}, 0, len(info.VsphereNetwork.Networks))
	log.Printf("[DEBUG] namespacehelper.FlattenVsphereNetwork: Flattening %d networks", len(info.VsphereNetwork.Networks))
	for _, network := range info.VsphereNetwork.Networks {
		if network.ConfigStatus == namespace.NetworkConfigStatusRemoving {
			log.Printf("[DEBUG] namespacehelper.FlattenVsphereNetwork: Skipping network %s, which is being removed", network.Network)
			continue
		}
		networks = append(networks, network.Network)
	}

	return []interface{}{map[string]interface{}{
		"default_workload_network": info.VsphereNetwork.DefaultWorkloadNetwork,
		"networks":                 networks,
	}}
}
