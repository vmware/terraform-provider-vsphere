// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"crypto/sha256"
	"fmt"
	"log"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/hostsystem"
)

func dataSourceVSphereHostVGpuProfile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostVGpuProfileRead,
		Schema: map[string]*schema.Schema{
			"host_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Managed Object ID of the host system.",
			},
			"name_regex": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A regular expression used to match the vGPU Profile on the host.",
			},
			"vgpu_profiles": {
				Type:        schema.TypeList,
				Description: "List of vGPU profiles available via the host.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vgpu": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of a particular VGPU available as a shared GPU device (vGPU profile).",
						},
						"disk_snapshot_supported": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether the GPU plugin on this host is capable of disk-only snapshots when VM is not powered off.",
						},
						"memory_snapshot_supported": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether the GPU plugin on this host is capable of memory snapshots.",
						},
						"suspend_supported": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether the GPU plugin on this host is capable of suspend-resume.",
						},
						"migrate_supported": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Indicates whether the GPU plugin on this host is capable of migration.",
						},
					},
				},
			},
		},
	}
}

func dataSourceVSphereHostVGpuProfileRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] DataHostVGpuProfile: Beginning vGPU Profile lookup on %s", d.Get("host_id").(string))

	client := meta.(*Client).vimClient

	host, err := hostsystem.FromID(client, d.Get("host_id").(string))
	if err != nil {
		return err
	}

	hprops, err := hostsystem.Properties(host)
	if err != nil {
		return err
	}

	// Create unique ID based on the host_id
	idsum := sha256.New()
	if _, err := fmt.Fprintf(idsum, "%#v", d.Get("host_id").(string)); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%x", idsum.Sum(nil)))

	log.Printf("[DEBUG] DataHostVGpuProfile: Looking for available vGPU profiles")

	// Retrieve the SharedGpuCapabilities from host properties
	vgpusRaw := hprops.Config.SharedGpuCapabilities

	// If searching for a specific vGPU profile (by name)
	name, ok := d.GetOk("name_regex")
	if (ok) && (name.(string) != "") {
		return searchVGpuProfileByName(d, vgpusRaw, name.(string))
	}

	// Loop over all vGPU profiles on host
	vgpus := make([]interface{}, len(vgpusRaw))
	for i, v := range vgpusRaw {
		log.Printf("[DEBUG] DataHostVGpuProfile: Host %s has vGPU profile %s", d.Get("host_id").(string), v.Vgpu)
		vgpu := map[string]interface{}{
			"vgpu":                      v.Vgpu,
			"disk_snapshot_supported":   v.DiskSnapshotSupported,
			"memory_snapshot_supported": v.MemorySnapshotSupported,
			"suspend_supported":         v.SuspendSupported,
			"migrate_supported":         v.MigrateSupported,
		}
		vgpus[i] = vgpu
	}

	// Set the `vgpu_profile` output to all vGPU profiles
	if err := d.Set("vgpu_profiles", vgpus); err != nil {
		return err
	}

	log.Printf("[DEBUG] DataHostVGpuProfile: Identified %d vGPU profiles available on %s", len(vgpus), d.Get("host_id").(string))

	return nil
}

func searchVGpuProfileByName(d *schema.ResourceData, vgpusRaw []types.HostSharedGpuCapabilities, name string) error {
	log.Printf("[DEBUG] DataHostVGpuProfile: Selecting devices which match name regex")

	vgpus := make([]interface{}, 0, len(vgpusRaw))

	re, err := regexp.Compile(name)
	if err != nil {
		return err
	}

	// Loop over all vGPU profile and attempt to match by name
	for _, v := range vgpusRaw {
		if re.Match([]byte(v.Vgpu)) {
			// Identified matching vGPU profile
			log.Printf("[DEBUG] DataHostVGpuProfile: Host %s has vGPU profile %s", d.Get("host_id").(string), v.Vgpu)
			vgpu := map[string]interface{}{
				"vgpu":                      v.Vgpu,
				"disk_snapshot_supported":   v.DiskSnapshotSupported,
				"memory_snapshot_supported": v.MemorySnapshotSupported,
				"suspend_supported":         v.SuspendSupported,
				"migrate_supported":         v.MigrateSupported,
			}
			vgpus = append(vgpus, vgpu)
		}
	}

	if len(vgpus) == 0 {
		log.Printf("[DEBUG] DataHostVGpuProfile: Host %s does not support vGPU profile name regex [%s]", d.Get("host_id").(string), name)
	}

	// Set the `vgpu_profile` output to located vGPU.
	// If a matching vGPU profile is not found, return null
	if err := d.Set("vgpu_profiles", vgpus); err != nil {
		return err
	}

	return nil
}
