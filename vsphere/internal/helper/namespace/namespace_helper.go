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
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/viapi"
)

func ValidateVersionRequirements(version viapi.VSphereVersion, d *schema.ResourceDiff) error {
	log.Printf("[DEBUG] namespacehelper.ValidateVersionRequirements: Validating version requirements against vCenter %d.%d.%d", version.Major, version.Minor, version.Patch)
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

	return errors.Join(errs...)
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
	log.Printf("[DEBUG] namespacehelper.ExpandAccessList: Expanding %d access entries", len(entries))
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
	log.Printf("[DEBUG] namespacehelper.FlattenAccessList: Flattening %d access entries", len(accessList))
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
	log.Printf("[DEBUG] namespacehelper.ExpandZones: Expanding %d zones", len(entries))
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
	log.Printf("[DEBUG] namespacehelper.FlattenZones: Flattening %d zones", len(zones))
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
			log.Printf("[DEBUG] namespacehelper.FlattenZones: Skipping zone %s, which is marked for removal", zone.Zone)
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
	log.Printf("[DEBUG] namespacehelper.expandVMReservations: Expanding %d VM reservations", set.Len())
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
	log.Printf("[DEBUG] namespacehelper.FlattenZoneInfos: Flattening %d zones", len(zones))
	result := make([]interface{}, 0, len(zones))
	for _, zone := range zones {
		result = append(result, map[string]interface{}{
			"name": zone.Zone,
			"type": zone.Type,
		})
	}
	return result
}
