# <!-- markdownlint-disable first-line-h1 no-inline-html -->

## v2.14.1

> Release Date: 2025-07-21

BUG FIX:

- `r/virtual_machine`: Update static mac when reconfiguring VM. (#2561)
- `r/virtual_machine`: Server fault when cloning VM on DRS-enabled cluster. (#2560)

DOCUMENTATION:

- `r/virtual_machine`: Fix broken link to “Specifying the Root Resource Pool”. (#2553)

## v2.14.0

> Release Date: 2025-06-30

FEATURES:

- `r/virtual_machine`: Add/remove TPM module when editing and cloning r/virtual_machine. (#2542)
- `r/virtual_machine`: Attach sparse `.vmdk` when creating virtual machine. (#2383)
- `d/network`: Add VPC awareness to network data source. (#2547)

IMPROVEMENTS:

- `d/resource_pool`: Add `parent_resource_pool_id`. (#2460)

BUG FIX:

- `d/datastore_stats`: Skip inaccessible datastores. (#2497)
- `r/virtual_disk`: Simultaneous attempts to create parent directories. (#2366)
- `r/virtual_machine`: Apply top-rated SDRS placement recommendation. (#2550)

CHORE:

- `technical-debt`: Code quality improvements including resolving lint warnings, correcting error handling, simplifying control flow, improving naming consistency, cleaning code formatting, and refactoring for clearer and more maintainable code. (#2488, #2489, #2490, #2491, #2492, #2493, #2494, #2503, #2504, #2505, #2506, #2507, #2509, #2510, #2511, #2512, #2516, #2517, #2526, #2528, #2533, #2534, #2535, #2536)
- `technical-debt`: Use of deprecated `GetOkExists`. (#2482)
- `technical-debt`: Exported name starts with package. (#2530)
- `technical-debt`: Direct comparison of errors. (#2529)
- `technical-debt`: Remove legacy state migration. (#2524)
- `technical-debt`: Remove network interface resource allocation gate. (#2522)
- `technical-debt`: Replace type assertion on errors. (#2527)
- `acceptance-tests`: Stabilization and refactoring of existing tests. (#2500, #2501, #2502, #2521, #2525, #2531, #2540)
- `provider`: `d.Set` is not checked. (#2496)
- `provider`: Remove legacy min versions. (#2515)
- `provider`: Update codeowners. (#2518)
- `provider`: Set golangci-lint version to v2.1.6. (#2551)
- `provider`: Updated `CodelyTV/pr-size-labeler` to 1.10.3. (#2537)
- `provider`: Updated `github.com/hashicorp/terraform-plugin-testing` to 1.13.2. (#2546)
- `provider`: Updated `github.com/vmware/govmomi` to 0.51.0. (#2539)
- `provider`: Updated `github.com/cloudflare/circl` to 1.6.1. (#2545)

## v2.13.0

> Release Date: 2025-05-12

IMPROVEMENTS:

- `r/entity_permission`: Changes to `entity_id` and `entity_type` will now force resource re-creation. (#2387)
- `r/host_virtual_switch`: Added a diff suppression to ignore ordering changes for `network_adapters`. (#2388)
- `r/host_virtual_switch`: Resolved panic when an operation is called with a `nil` result. (#2445)
- `d/guest_os_customization`: Added missing schema options. (#2447)
- `r/resource_pool`: Updated to reconcile the scale descendants shares setting for a resource pool. (#2429)
- `r/virtual_machine`: Resolved panic when both parent vApp and resource pool are `nil`. (#2444)
- `r/license`/`d/license`: Refactored logging implementation and masked sensitive data. (#2436)

DOCUMENTATION:

- `d/folder`/`r/folder`: Updated documentation. (#2459)
- `r/license`/`d/license`: Updated documentation. (#2435)
- `r/virtual_machine`: Updated documentation. (#2480)
- `d/ovf_vm_template`: Updated documentation. (#2480)

CHORE:

- `provider`: Updated `vmware/govmomi` to 0.50.0. (#2382)
- `provider`: Migrated provider testing from `hashicorp/terraform-plugin-sdk` to `hashicorp/terraform-plugin-testing`. (#2381)
- `workflows`: Updated, added, and removed GitHub Actions workflows. (#2390, #2391, #2392, #2393, #2395, #2437, #2449, #2450, #2456, #2463, #2464, #2465, #2466, #2467, #2468)
- `technical-debt`: Code quality improvements including resolving lint warnings, correcting error handling, simplifying control flow, improving naming consistency, cleaning code formatting, and refactoring for clearer and more maintainable code. (#2415, #2416, #2417, #2418, #2419, #2420, #2421, #2422, #2423, #2424, #2425, #2426, #2427, #2428, #2430, #2431, #2432, #2433, #2434, #2439, #2440, #2441)

## v2.12.0

> Release Date: 2025-04-22

IMPROVEMENTS:

- `d/network`: Added `retry_timeout` and `retry_interval` attributes that enable retry if network is not found. (#2349)
- `r/host_port_group`: Fixed `null` reference panic during resource import. (#2377)
- `r/vmfs_datastore`: Fixed error during resource deletion. (#2360)

CHORE:

- `provider`: Updated `go` to 1.23.8. (#2372)
- `provider`: Updated `golang.org/x/net` to 0.38.0. (#2358, #2378)
- `provider`: Updated `golang.org/x/text` to 0.22.0. (#2350)
- `provider`: Updated `golang.org/x/sys` to 0.30.0. (#2350)
- `provider`: Updated `golang.org/x/sync` to 0.11.0. (#2350)
- `provider`: Updated `golang.org/x/crypto` to 0.35.0. (#2350)
- `provider`: Updated `github.com/hashicorp/terraform-plugin-sdk/v2` to 0.1.2. (#2346)
- `provider`: Updated `github.com/hashicorp/yamux` to 2.36.1. (#2334)
- `provider`: Updated `vmware/govmomi` to 0.49.0. (#2342, #2357)

DOCUMENTATION:

- Migrated documentation from legacy path. (#2370)

## v2.11.1

> Release Date: 2025-02-03

IMPROVEMENTS:

- `r/supervisor`: Added support for `main_ntp` and `worker_ntp`. (#2326)
- Updates the session file formatter string to `"%064x"` from `"%040x"` to match `vmware/govc` v0.48.0 and later. (#2329)

CHORE:

- `provider`: Updated `vmware/govmomi` to 0.48.0. (#2325, #2329)

## v2.11.0

> Release Date: 2025-01-21

IMPROVEMENTS:

- `r/virtual_machine`: Added support for NVMe controllers. (#2321)
- `r/distributed_virtual_switch`: Added support for vSphere distributed switch version `8.0.3` in vSphere 8.0 U3. (#2306)
- `r/distributed_virtual_switch_pvlan_mapping`: New resource added to support management of individual PVLAN mapping records on a distributed switch. (#2291)

DOCUMENTATION:

- Updated documentation links to `techdocs.broadcom.com`, as needed. (#2322)

CHORE:

- `provider`: Updated `go` to 1.22.8. (#2289)
- `provider`: Updated `golang.org/x/net` to 0.33.0. (#2319)
- `provider`: Updated `golang.org/x/crypto` to 0.31.0. (#2318)

## v2.10.0

> Release Date: 2024-10-16

FEATURES:

- `d/network`: Adds ability to add `filter` to find port groups based on network type of standard virtual port group, distributed virtual port group, or network port group. (#2281)
- `r/virtual_machine`: Adds ability to add a virtual Trusted Platform Module (`vtpm`) to virtual machine on creation or clone. (#2279)
- `d/virtual_machine`: Adds ability to read the configuration of a virtual Trusted Platform Module (`vtpm`) on virtual machine; will return `true` or `false` based on the configuration. (#2279)
- `d/datastore_stats`: Adds ability to return all data stores, both local and under a datastore cluster, in the datastore list. (#2273)
- `d/datasource_cluster`: Adds ability to return datastore names from a datastore cluster. (#2274)
- `d/datacenter`: Adds ability to return list of virtual machine names for the specified datacenter. (#2276)

IMPROVEMENTS:

- `r/virtual_machine`: Documentation Updated. (#2285)

CHORE:

- `provider`: Updated `vmware/govmomi` to 0.44.1. (#2282)

## v2.9.3

> Release Date: 2024-10-08

BUG FIX:

- `r/tag_category`: Updates resource not to `ForceNew` for cardinality. This will allow the `tag_category` to be updated. (#2263)
- `r/host`: Updates resource to check thumbprint of the ESXI host thumbprint before adding the host to a cluster or vCenter Server. (#2266)

DOCUMENTATION:

- `r/resource_pool`: Updates to include steps to create resource pool on standalone ESXi hosts. (#2264)
- `r/virtual_machine`: Updates to fix examples of `disk0` to reflect that during import the disk gets set back to default like `Hard Disk 1`. (#2272)

## v2.9.2

> Release Date: 2024-09-16

BUG FIX:

- `r/compute_cluster_vm_group`: Updates resource to allow for additional virtual machines to be added or removed from a VM Group. Must be run in conjunction with an import. (#2260)

FEATURES:

- `r/tag`: Adds a format validation for `category_id`. (#2261)

## v2.9.1

> Release Date: 2024-09-09

BUG FIX:

- `r/resource_pool`: Removes the default setting for `scale_descendants_shares` to allow for inheritance from the parent resource pool. (#2255)

DOCUMENTATION:

- `r/virtual_machine`: Updates to clarify assignment of `network_interface` resources. (#2256)
- `r/host`: Updates to clarify import of `hosts`. (#2257)
- `r/compute_cluster`: Updates to clarify import of `compute_cluster` resources. (#2257)
- `r/virtual_machine`: Updates to clarify the `vm` path in the import of `virtual_machine` resources. (#2257)

## v2.9.0

> Release Date: 2024-09-03

FEATURES:

- `r/host`: Added support for ntpd for services on `r/host`. This allows for ntpd service settings and policy to be added to a host resource and future expansion of additional services. (#2232)

CHORE:

- `provider`: Updated `vmware/govmomi` to 0.42.0. (#2248)
- `provider`: Updated `go` to 1.22.6 (#2247)

## v2.8.3

> Release Date: 2024-08-13

BUG FIX:

- `r/virtual_machine`: Fixed virtual machine reconfiguration with multiple PCI passthrough devices. (#2236)
- `r/virtual_machine`: Fixed inability to apply a default gateway on more than one network adapter. (#2235)

FEATURES:

- `r/virtual_disk`: Allows the increasing of the size of virtual disks. Reductions in size are not supported by vSphere and are allowed. (#2244)

CHORES:

- `provider`: Updated `vmware/govmomi` to 0.39.0. (#2240)

## v2.8.2

> Release Date: 2024-06-28

BUG FIX:

- `r/file`: Updated to ensure that incoming file names with special characters (`+`, specifically) retain their original name when uploaded. (#2217)
- `r/virtual_machine`: Updated `searchPath` to use `path` instead of `filepath` since this is a vSphere inventory path (_e.g._, `\Datacenter\vm\<vm_name>`), not a directory path. (#2216)
- `r/virtual_machine`: Removed the default values for `ept_rvi_mode` and `hv_mode` from the virtual machine configuration. (#2230)
- `r/virtual_machine`: Fixed overflow for the disk sub-resource when running a 32-bit version of the provider. Modified the call to `GiBToByte` by passing the parameter as `int64` which forces the function to go through the 64-bit case. (#2200)

FEATURES:

- `d/virtual_machine`: Added support for `instance_uuid`. (#2198)

DOCUMENTATION:

- `d/ovf_vm_template` and `r/virtual_machine`: Updated to use a Ubuntu Server cloud image since the nested ESXi OVA images are no longer available for direct download from Flings. (#2215)
- `r/virtual_machine`: Updated to denote support and limitations for options. (#2218)
- `r/virtual_machine`: Added examples for the use of guest customization specifications. (#2218)
- Removed deprecated interpolation syntax where it is no longer required. (#2220)
- Updated examples to use the correct syntax, formatting, and alignment with other examples in the docs. (#2222)
- Updated all links in documentation, as necessary. (#2212)

CHORE:

- `provider`: Updated `vmware/govmomi` to 0.38.0. (#2229)
- `provider`: Updated `hashicorp/terraform-plugin-sdk` to 2.34.0. (#2201)
- `provider`: Added version tracking (`// Minimum Supported Version: x.y.z`) where there is a version restriction. This is preparation for removing checks and to support only 7.0 and later per the product lifecycle. (#2213)

## v2.8.1

> Release Date: 2024-05-08

BUG FIX:

- `r/virtual_machine`: Reverts removing the default values for `ept_rvi_mode` and `hv_mode` from the virtual machine configuration. (#2194)

## v2.8.0

> Release Date: 2024-05-07

BUG FIX:

- `r/virtual_machine`: Removed the default values for `ept_rvi_mode` and `hv_mode` from the virtual machine configuration. (#2172)
- `r/virtual_machine`: Fixed issue when network interfaces, created by Docker, with the same `deviceConfigId` causes an unexpected output. (#2121)

FEATURES:

- `r/virtual_machine`: Added support for specifying a `datastore_cluster_id` when cloning from a vSphere content library. (#2061)
- `r/guest_os_customization`: Added support for `domain_ou` for Windows customizations added in vSphere 8.0.2. (#2181)
- Added resources for vSphere workload management. (#2791)
  - Enable workload management on a cluster.
  - Create custom namespaces and VM classes.
  - Choose a content library.
  - Configure passthrough devices for VM classes (e.g. vGPU).
- `r/offline_software_depot`: Added resource to the provider for offline software depots. Support for online depots can be added at a later time. Only depots with source type "PULL" are supported. This is intentional and aims to discourage the use of the deprecated VUM functionality. (#2143)
- `d/host_base_images`: Added data source to the provider for base images. Declaring this data source allows users to retrieve the full list of available ESXi versions for their environment. (#2143)
- `r/compute_cluster`: Added property that serves as an entry point for the vLCM configuration. Allows selection of a base image and a list of custom components from a depot. Configuring this property for the first time also enables vLCM on the cluster. (#2143)

DOCUMENTATION:

- `folder`: Added clarification for storage folders instead of datastore folders. (#2183)
- `r/virtual_machine`: Corrected resource and data source anchor links intended for `virtual_machine#virtual-machine-customizations`. (#2182)

CHORES:

- `provider`: Updated to allow the use of a SHA256 thumbprint when connecting to Center Server. Support for SHA256 was added to `vmware/govmomi` 0.36.1. (#2184)
- `provider`: Updated `hashicorp/terraform-plugin-sdk` to 2.33.0. (#2137)
- `provider`: Updated `vmware/govmomi` to 0.37.1. (#2174)
- `provider`: Updated `golang.org/x/net` to 0.23.0. (#2173)
- `provider`: Updated `golang.org/protobuf` to 1.33.0. (#2155)

## v2.7.0

> Release Date: 2024-03-06

BUG FIX:

- `r/virtual_machine`: Fixed support for SR-IOV passthrough virtual machine network adapters. (#2133)
- `r/virtual_machine`: Unifies `disk.keep_on_remove` with default and `disk.label` with the correct one assigned to the virtual machine disk during import. If the datastore for a virtual machine is part of a datastore cluster the `datastore_cluster_id` attribute is filled during import. (#2127)
- `r/virtual_machine`: Changed the default value for `sync_time_with_host` in `r/virtual_machine` to `true` to align with default value provided by the UI. (#2120)
- `r/virtual_machine`: Added the virtual machine folder in the search for virtual machine criteria when deploying from an OVF/OVA scenario. Allows virtual machines with same names in different virtual machine folders to be distinguished as different managed entities. (#2118)
- `r/virtual_disk`: Fixed import to use the correct `vmdk_path`. (#1762)

FEATURES:

- `r/virtual_machine`: Added support for `memory_reservation_locked_to_max` property. If set true, memory resource reservation for the virtual machine will always be equal to the virtual machine's memory size. (#2093)
- `d/host_vgpu_profile`: Added data source to the provider to query and return available vGPU profiles for an ESXi host. (#2048)
- `d/datastore_stats`: Added datastore stats to report total capacity and free space of datastores. (#1896)
- `d/datastore`: Added stats to report total capacity and free space of a single datastore. (#1896)

DOCUMENTATION:

- Updated `install.md` to use `unzip` for Linux and macOS examples. (#2105)

CHORES:

- `provider`: Updated `vmware/govmomi` to 0.35.0. (#2132)
- `provider`: Updated `hashicorp/terraform-plugin-sdk` to 2.32.0. (#2125)
- `provider`: Updated `go` to 1.22.0 (#2139)

## v2.6.1

> Release Date: 2023-12-11

BUG FIX:

- `r/guest_os_customization`: Fixed incorrect path for `RequiredWith` and `ConflictsWith` attribute identifiers for `windows_options`. (#2083)
- `r/virtual_machine`: Fixed error setting SR-IOV (`sriov`) network interface address. (#2081)

## v2.6.0

> Release Date: 2023-11-29

BUG FIX:

- `r/virtual_machine`: Fixed upload error when deploying an OVF/OVA directly to an ESXi host. (#1813)

FEATURES:

- `r/compute_cluster`: Added support for vSAN Express Storage Architecture in vSphere 8.0. (#1874)
- `r/compute_cluster`: Added support for vSAN stretched clusters. (#1885)
- `r/compute_cluster`: Added support for vSAN fault domains. (#1968)
- `r/guest_os_customization`: Added support for the customization specifications for guest operating systems. (#2053)
- `d/guest_os_customization`: Added support for the customization specifications for guest operating systems. (#2053)
- `r/virtual_machine`: Added support for the use of customization specifications for guest operating systems. (#2053)
- `r/virtual_machine`: Added support for the SR-IOV (`sriov`) network interface adapter type. (#2059, #1417)

## v2.5.1

> Release Date: 2023-10-12

BUG FIX:

- `r/virtual_machine`: Fixed cloning regression on datastore cluster. Restored behavior not to send relocate specs for the virtual disks when it is cloned on datastore cluster with exception when `datastore_id` is explicitly specified for the virtual disk. (#2037)
- `r/virtual_disk`: Fixed improper disk type handling forcing disks to be recreated. (#2033)

IMPROVEMENTS:

- `r/virtual_machine`: Allow hardware version up to 21. (#2038)

CHORES:

- `provider`: Updated `golang.org/x/net` from 0.13.0 to 0.17.0. (#2035)

## v2.5.0

> Release Date: 2023-10-09

BUG FIX:

- `r/virtual_machine`: Removes the validation for `eagerly_scrubbed` and `thin_provision` fields for a disk subresource so that `ignore_changes` fixed a deployment. (#2028)
- `r/virtual_machine`: Added a differential between the disk properties specified and those existing on the source virtual machine disk, allowing changes to be sent to the API for disk subresource. (#2028)

IMPROVEMENTS:

- `r/nic`: Documentation Updated. (#2017)

CHORES:

- `provider`: Updated `vmware/govmomi` to 0.31.0. (#2026)

## v2.4.3

> Release Date: 2023-09-08

BUG FIX:

- `r/virtual_machine`: Fixed hardware version conversion. (#2011)

CHORES:

- `provider`: Updated `hashicorp/terraform-plugin-sdk` v2.28.0. (#2002)
- `provider`: Updated `vmware/govmomi` to 0.30.7. (#1972)

## v2.4.2

> Release Date: 2023-08-21

BUG FIX:

- `r/virtual_machine`: Fixed hardware version error when cloning and/or configuring a VM/Template. (#1995)
- `r/virtual_machine`: Fixed invalid operation for device '0' when reconfiguring a VM. (#1996)

CHORES:

- `provider`: Updated `hashicorp/terraform-plugin-sdk` to 2.27.0. (#1937)
- `provider`: Updated `vmware/govmomi` to 0.30.7. (#1972)

## v2.4.1

> Release Date: 2023-06-26

BUG FIX:

- `r/compute_cluster`: Added version check for [vSphere 7.0.1 or later](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vsan.doc/GUID-9113BBD6-5428-4287-9F61-C8C3EE51E07E.html) when enabling vSAN HCI Mesh. (#1931)

## v2.4.0

> Release Date: 2023-05-05

FEATURES:

- `d/virtual_machine`: Added support for lookup by moid. (#1868)
- `r/vnic`: Added support for services on vmkernel adapter/vnic. (#1855)

BUG FIX:

- `r/nas_datastore`: Fixed issue mounting and/or unmounting NFS datastores when updating `host_system_ids` as a day-two operation. (#1860)
- `r/virtual_machine_storage_policy`: Updated the `resourceVMStoragePolicyDelete` method to check the response of `pbmClient.DeleteProfile()` API for errors. If a storage policy is in use and cannot be deleted, the destroy operation will fail and the storage policy will remain in the state. (#1863)
- `r/virtual_machine`: Fixed vSAN timeout. (#1864)

IMPROVEMENTS:

- `r/host`: Documentation Updated. (#1884)
- `r/vnic`: Fixed tests. (#1887)

CHORES:

- `provider`: Updated `hashicorp/terraform-plugin-sdk` to 2.26.1. (#1862)
- `provider`: Updated `vmware/govmomi` to 0.30.4. (#1858)

## v2.3.1

> Release Date: 2023-02-08

**If you are using v2.3.0, please upgrade to this new version as soon as possible.**

BUG FIX:

- `r/compute_cluster`: Fixed panic when reading vSAN. (#1835)
- `r/file`: Fixed a provider crash by updating the `createDirectory` method to check if the provided file path has any parent folder(s). If no folders need to be created `FileManager.MakeDirectory` is not invoked. (#1866)

## v2.3.0

> Release Date: 2023-02-08

FEATURES:

- `r/virtual_machine`: Added support for the paravirtual RDMA (PVRDMA) `vmxnet3vrdma` network interface adapter type. (#1598)
- `r/virtual_machine`: Added support for an optional `extra_config_reboot_required` argument to `r/virtual_machine`. This argument allows you to configure if a virtual machine reboot is enforced when `extra_config` is changed. (#1603)
- `r/virtual_machine`: Added support for two (2) CD-ROMs attached to a virtual machine. (#1631)
- `r/compute_cluster`: Added support for vSAN compression and deduplication. (#1702)
- `r/compute_cluster`: Added support for vSAN performance services. (#1727)
- `r/compute_cluster`: Added support for vSAN unmap. (#1745)
- `r/compute_cluster`: Added support for vSAN HCI Mesh. (#1820)
- `r/compute_cluster`: Added support for vSAN Data-in-Transit Encryption. (#1820)
- `r/role`: Added support for import. (#1822)

BUG FIX:

- `r/datastore_cluster`: Fixed error parsing string as enum type for `sdrs_advanced_options`. (#1749)
- `provider`: Reverts a linting update from #1416 back to SHA1. SHA1 is used by `vmware/govmomi` for the session file. This will allow session reuse from govc. (#1808)
- `r/compute_cluster`: Fixed panic in vsan disk group. (#1820)
- `r/virtual_machine`: Updating the `datastore_id` will apply to disk sub-resources. (#1817)

IMPROVEMENTS:

- `r/distributed_virtual_switch`: Added support for vSphere distributed switch version `8.0.0` in vSphere 8.0. (#1767)
- `r/virtual_machine`: Enables virtual machine reconfiguration tasks to use the provider `api_timeout` setting. (#1645)
- `r/host`: Documentation Updated. (#1675)
- `r/host_virtual_switch`: Allows `standby_nics` on `r/host_virtual_switch` to be an optional attribute so `standby_nics = []` does not need to be defined when no standby NICs are required/available. (#1695)
- `r/compute_cluster_vm_anti_affinity_rule`: Documentation Updated. (#1700)
- `ovf_vm_template`: Documentation Updated. (#1792)

CHORES:

- `provider`: Updated `vmware/govmomi` to 0.29.0. (#1701)

## v2.2.0

> Release Date: 2022-06-16

BUG FIX:

- `r/virtual_machine`: Fixed ability to clone and import virtual machine resources with SATA and IDE controllers. (#1629)
- `r/dvs`: Prevent setting unsupported traffic classes. (#1633)
- `r/virtual_machine`: Fixed provider panic when a non-supported PCI device is added outside Terraform to a virtual machine. (#1627)
- `r/datacenter`: Updated `resourceVSphereDatacenterImport` to include the datacenter folder in which the datacenter object may exist. (#1607)
- `r/virtual_machine`: Fixed issue where PCI passthrough devices not applied during initial cloning. (#1625)
- `helper/content_library`: Fixed content library item local iso upload (#1665)

FEATURES:

- `r/host`: Added support for custom attributes. (#1619)
- `r/virtual_machine`: Added support for guest customization script for Linux guest operating systems. (#1621)
- `d/virtual_machine`: Added support for lookup by `uuid`. (#1650)
- `r/compute_cluster`: Added support for scalable shares. (#1634)
- `r/resource_pool`: Added support for scalable shares. (#1634)
- `d/compute_cluster_host_group`: Added support for a data source that can be used to read general attributes of a host group. (#1636)

IMPROVEMENTS:

- `r/resource_pool`: Documentation Updated. (#1620)
- `r/virtual_machine`: Documentation Updated. (#1630)
- `r/virtual_machine`: Added `tools_upgrade_policy` to list of params that trigger a reboot (#1644)
- `r/datastore_cluster`: Documentation Updated. (#1670)
- `provider`: Index page documentation Updated. (#1672)
- `r/host_port_group`: Documentation Updated. (#1671)
- `provider`: Updated Go to 1.18 and remove vendoring. (#1676)
- `provider`: Updated Terraform Plugin SDK to 2.17.0. (#1677)

## v2.1.1

> Release Date: 2022-03-08

**If you are using v2.1.0, please upgrade to this new version as soon as possible.**

BUG FIX:

- `r/compute_cluster`: Reverts (#1432) switching `vsan_disk_group` back to `TypeList`. Switching from `TypeList` to `TypeSet` is a sore spot when it comes to what is considered a breaking change to provider configuration. Generally we accept that users may use list indices within their config. When this attribute switched to `TypeSet` this caused a breaking change for configurations doing that, as `TypeSet` is indexed by a hash value that Terraform calculates. Furthermore other code around type assertions was not changed and this attribute actually crashed the provider in `v2.1.0`, we will address the now re-opened (#1205) in `v3.0.0` of the provider. (#1615)

FEATURES:

- `r/virtual_machine`: Added support to check the power state of the resource. (#1407)

## v2.1.0

> Release Date: 2022-02-28

BUG FIX:

- `r/compute_cluster`: Updated `ha_datastore_apd_response_delay` to the API default (180) for `vmTerminateDelayForAPDSec`. Previously set to 3 (minutes) however the codebase uses this value as seconds. Users who had the field left blank may see a warning about the state value drifting from 3 to 180, after applying this should go away. (#1542)
- `r/virtual_machine`: Don't read `storage_policy_id` if vCenter is not configured. This is not a scenario we test or support explicitly (#1408)
- `d/virtual_machine`: Fixed silent failure and add `default_ip_address` attribute. (#1532)
- `r/virtual_machine`: Fixed race condition by always forcing a new datastore id. (#1486)
- `r/virtual_machine`: Fixed default guest OS identifier. (#1543)
- `r/virtual_machine`: Updated `windows_options` to ensure all required options for domain join are provided (#1562)
- `r/virtual_machine`: Fixed migration of all disks and configuration files when the datastore_cluster_id is changed on the resource. (#1546)
- `r/file`: Fixed upload of VMDK to datastore on ESXi host. (#1409)
- `r/tag`: Fixed deletion detection in `tag` and `tag_category`. (#1579)
- `r/virtual_machine`: Sets `annotation` to optional + computed. (#1588)

FEATURES:

- `d/license`: New datasource can be used to read general attributes of a license. (#1580)
- `r/virtual_machine`: Added the `tools_upgrade_policy` argument to set the upgrade policy for VMware Tools. (#1506)

IMPROVEMENTS:

- `r/vapp_container`: Documentation Updated. (#1551)
- `r/computer_cluster_vm_affinity_rule`: Documentation Updated. (#1544)
- `r/computer_cluster_vm_anti_affinity_rule`: Documentation Updated. (#1544)
- `r/virtual_machine`: Documentation Updated. (#1513)
- `r/custom_attribute`: Documentation Updated. (#1508)
- `r/virtual_machine_storage_policy`: Documentation Updated. (#1541)
- `d/storage_policy`: Documentation Updated. (#1541)
- `r/distributed_virtual_switch`: Documentation Updated. (#1504)
- `d/distributed_virtual_switch`: Documentation Updated. (#1504)
- `r/distributed_virtual_switch`: Added support for dvs versions `6.6.0` and `7.0.3`. (#1501)
- `d/content_library_item`: Documentation Updated. (#1507)
- `r/ha_vm_override`: Added `disabled` option to `ha_vm_restart_priority`. (#1505)
- `r/virtual_disk`: Documentation Updated. (#1569)
- `r/virtual_machine`: Documentation Updated. (#1566)
- `r/content_library`: Documentation Updated. (#1577)
- `r/virtual_machine`: Updated the documentation with the conditions that causes the virtual machine to reboot on update. (#1522)
- `r/distributed_virtual_switch`: Devices argument is now optional (#1468)
- `r/virtual_host`: Added support add tags to hosts. (#1499)
- `r/content_library_item`: Documentation Updated. (#1586)
- `r/file`: Documentation Updated and deletion fix. (#1590)
- `r/virtual_machine`: Documentation Updated. (#1595)
- `r/tag_category`: Performs validation of associable types and update documentation around "All" which never worked. (#1602)

## v2.0.2

> Release Date: 2021-06-25

BUG FIX:

- `r/virtual_machine`: Fix logic bug that caused the provider to set unsupported fields when talking to Sphere 6.5. (#1430)
- `r/virtual_machine`: Fix resource diff bug where it was not possible to ignore changes to `cdrom` subresource. (#1433)

IMPROVEMENTS:

- `r/virtual_machine`: Support periodic time syncing for VMs on vSphere 7.0U1 onwards. (#1431)

## v2.0.1

> Release Date: 2021-06-09

BUG FIX:

- `r/virtual_machine`: Only set vvtd/vbs if vSphere version is newer than 6.5. (#1423)

## v2.0.0

> Release Date: 2021-06-02

BREAKING CHANGES:

- `provider`: Moving forward this provider will only work with Terraform version v0.12 and later.
- `r/virtual_machine`: [Deprecated attribute `name`](https://github.com/vmware/terraform-provider-vsphere/blob/main/CHANGELOG.md#130-january-26-2018) has been removed from the `disk` subresource.

BUG FIX:

- `d/ovf_datasource`: Fix validation error when importing OVF spec. (#1398)
- `r/virtual_machine`: Fix post-import VM regression. (#1361)
- `r/virtual_machine`: Round up when calculating disk capacity. (#1397)
- `r/vnic`: Fix default netstack name. (#1376)

IMPROVEMENTS:

- `provider`: Provider wide API timeout setting. (#1405)
- `provider`: Enable keepalive for REST API sessions. (#1301)
- `provider`: Upgrade Plugin SDK to 2.6.1 (#1379)
- `d/virtual_machine`: Added `network_interfaces` output. (#1274)
- `r/virtual_machine`: Allow non-configurable vApp properties to be set. (#1199)
- `r/virtual_machine`: Enable VBS (`vbsEnabled`) and I/O MMU (`vvtdEnabled`). (#1287)
- `r/virtual_machine`: Added `replace_trigger` to support replacement of vms based external changes such as cloud_init (#1360)

## v1.26.0

> Release Date: 2021-04-20

BUG FIX:

- Minor Fixed of issues that came up during testing against vSphere 7.0
- Change the way we set the timeout for maintenance mode (#1392)

IMPROVEMENTS:

- `provider`: vSphere 7 compatibility validation (#1381)
- `r/virtual_machine`: Allow hardware version up to 19 (#1391)

## v1.25.0

> Release Date: 2021-03-17

BUG FIX:

- `r/entity_permissions`: Sorting permission objects on user name/group name before storing. (#1311)
- `r/virtual_machine`: Limit netmask length for ipv4 and ipv6 netmask. (#1321)
- `r/virtual_machine`: Fix missing vApp properties. (#1322)

FEATURES:

- `d/ovf_vm_template`: Created data source OVF VM Template. This new data source allows `virtual_machine` to be created by its exported attributes. See PR for more details. (#1339)

IMPROVEMENTS:

- `r/distributed_virtual_switch`: Allow vSphere7. (#1363)
- `provider`: Bump Go to version 1.16. (#1365)

## v1.24.3

> Release Date: 2020-12-14

BUG FIX:

- `r/virtual_machine`: Support for no disks in config (#1241)
- `r/virtual_machine`: Make API timeout configurable when building VMs (#1278)

## v1.24.2

> Release Date: 2020-10-16

BUG FIX:

- `r/virtual_machine`: Prevent `guest_id` nil condition. (#1234)

## v1.24.1

> Release Date: 2020-10-07

IMPROVEMENTS:

- `d/content_library_item`: Add `type` to content library item data source. (#1184)
- `r/virtual_switch`: Fix port group resource to enable LACP in virtual switch only. (#1214)
- `r/distributed_port_group`: Import distributed port group using MOID. (#1208)
- `r/host_port_group`: Add support for importing. (#1194)
- `r/virtual_machine`: Allow more config options to be changed from OVF. (#1218)
- `r/virtual_machine`: Convert folder path to MOID. (#1207)

BUG FIX:

- `r/datastore_cluster`: Fix missing field in import. (#1203)
- `r/virtual_machine`: Change default OS method on bare VMs. (#1217)
- `r/virtual_machine`: Read virtual machine after clone and OVF/OVA deploy. (#1221)

## v1.24.0

> Release Date: 2020-09-02

BUG FIX:

- `r/virtual_machine`: Skip SCSI controller check when empty. (#1179)
- `r/virtual_machine`: Make storage_policy_id computed to prevent flapping when unset. (#1185)
- `r/virtual_machine`: Ignore nil objects in host network on read. (#1186)
- `r/virtual_machine`: Keep progress channel open when deploying an OVF. (#1187)
- `r/virtual_machine`: Set SCSI controller type to unknown when nil. (#1188)

IMPROVEMENTS:

- `r/content_library_item`: Add local upload, OVA, and vm-template sources. (#1196)
- `r/content_library`: Subscription and publication support. (#1197)
- `r/virtual_machine`: Content library vm-template, disk type, and vApp property support. (#1198)

## v1.23.0

> Release Date: 2020-08-21

BUG FIX:

- `r/vnic`: Fix missing fields on vnic import. (#1162)
- `r/virtual_machine`: Ignore thin_provisioned and eagerly_scrub during DiskPostCloneOperation. (#1161)
- `r/virtual_machine`: Fix SetHardwareOptions to fetch the hardware version from QueryConfigOption. (#1159)

IMPROVEMENTS:

- `r/virtual_machine`: Allow performing a linked-clone from a template. (#1158)
- `d/virtual_machine`: Merge the virtual machine configuration schema. (#1157)

## v1.22.0

> Release Date: 2020-08-07

FEATURES:

- `r/compute_cluster`: Basic vSAN support on compute clusters. (#1151)
- `r/role`: Resource and data source to create and manage vSphere roles. (#1144)
- `r/entity_permission`: Resource to create and manage vSphere permissions. (#1144)
- `d/entity_permission`: Data source to acquire ESXi host thumbprints. (#1142)

## v1.21.1

> Release Date: 2020-07-20

BUG FIX:

- `r/virtual_machine`: Set guest_id before customization. (#1139)

## v1.21.0

> Release Date: 2020-06-30

FEATURES:

- `r/virtual_machine`: Support for SATA and IDE disks. (#1118)

## v1.20.0

> Release Date: 2020-06-23

FEATURES:

- `r/virtual_machine`: Add support for OVA deployment. (#1105)

BUG FIX:

- `r/virtual_machine`: Delete disks on destroy when deployed from OVA/OVF. (#1106)
- `r/virtual_machine`: Skip PCI passthrough operations if there are no changes. (#1112)

## v1.19.0

> Release Date: 2020-06-16

FEATURES:

- `d/dynamic`: Data source which can be used to match any tagged managed object. (#1103)
- `r/virtual_machine_storage_policy_profile`: A resource for tag-based storage placement policies management. (#1094)
- `r/virtual_machine`: Add support for PCI passthrough devices on virtual machines. (#1099)
- `d/host_pci_device`: Data source which will locate the address of a PCI device on a host. (#1099)

## v1.18.3

> Release Date: 2020-06-01

IMPROVEMENTS:

- `r/custom_attribute`: Fix id in error message when category is missing. (#1088)
- `r/virtual_machine`: Add vApp properties with OVF deployment. (#1082)

## v1.18.2

> Release Date: 2020-05-22

IMPROVEMENTS:

- `r/host` & `r/compute_cluster`: Add arguments for specifying if cluster management should be handled in `host` or `compute_cluster` resource. (#1085)
- `r/virtual_machine`: Handle OVF argument validation during VM creation. (#1084)
- `r/host`: Disconnect rather than entering maintenance mode when deleting. (#1083)

## v1.18.1

> Release Date: 2020-05-12

BUG FIX:

- `r/virtual_machine`: Skip unexpected NIC entries. (#1067)
- Respect `session_persistence` for REST sessions. (#1077)

## v1.18.0

> Release Date: 2020-05-04

FEATURES:

- `r/virtual_machine`: Allow users to deploy OVF templates from both local system and remote URL. (#1052)

## v1.17.4

> Release Date: 2020-04-29

IMPROVEMENTS:

- `r/virtual_machine`: Mark `product_key` as sensitive. (#1045)
- `r/virtual_machine`: Increase max `hardware_version` for vSphere v7.0. (#1056)

BUG FIX:

- `r/virtual_machine`: Fix to disk bus sorting. (#1039)
- `r/virtual_machine`: Only include `hardware_version` in CreateSpecs. (#1055)

## v1.17.3

> Release Date: 2020-04-22

IMPROVEMENTS:

- Use built-in session persistence in `vmware/govmomi`. (#1050)

## v1.17.2

> Release Date: 2020-04-13

IMPROVEMENTS:

- `r/virtual_disk`: Support VMDK files. (#987)

BUG FIX:

- `r/virtual_machine`: Fix disk controller sorting. (#1032)

## v1.17.1

> Release Date: 2020-04-07

IMPROVEMENTS:

- `r/virtual_machine`: Add support for hardware version tracking and upgrading. (#1020)
- `d/network`: Handle cases of network port groups with same name using `distributed_virtual_switch_uuid`. (#1001)

BUG FIX:

- `r/virtual_machine`: Fix working with orphaned devices. (#1005)
- `r/virtual_machine`: Ignore `guest_id` with content library. (#1014)

## v1.17.0

> Release Date: 2020-03-23

FEATURES:

- **New Data Source:** `content_library` (#985)
- **New Data Source:** `content_library_item` (#985)
- **New Resource:** `content_library` (#985)
- **New Resource:** `content_library_item` (#985)

IMPROVEMENTS:

- `r/virtual_machine`: Add `poweron_timeout` option for the amount of time to give a VM to power on. (#990)

## v1.16.2

> Release Date: 2020-03-04

IMPROVEMENTS:

- `r/virtual_machine`: Optimize OS family query. (#959)
- Migrate provider to Terraform plugin SDK. (#982)

## v1.16.1

> Release Date: 2020-02-06

BUG FIX:

- `r/virtual_machine`: Set `storage_policy_id` based off of VM rather than template. (#970)

## v1.16.0

> Release Date: 2020-02-04

FEATURES:

- **New Data Source:** `storage_policy` (#881)

IMPROVEMENTS:

- Switch to `vmware/govmomi` REST client (#955)
- Add storage policy to `virtual_machine` resource. **Requires `profile-driven storage` privilege on vCenter Server for the Terraform provider user. (#881)

## v1.15.0

> Release Date: 2020-01-23

IMPROVEMENTS:

- `r/virtual_machine`: Do not throw error when disk path is not known yet. (#944)

BUG FIX:

- `r/virtual_machine`: Do not set `datastoreID` in `RelocateSpec` when `datastore_cluster` is set. (#933)
- `r/vapp_container`: Fix handling of child vApp containers. (#941)
- `r/virtual_disk`: Enforce .vmdk suffix on `vmdk_path`. (#942)

## v1.14.0

> Release Date: 2019-12-18

IMPROVEMENTS:

- `r/host`: Add details to error messages. (#850)
- `r/virtual_machine`: Pick default datastore for extra disks. (#897)
- `r/virtual_machine`: Extend `ignored_guest_ips` to support CIDR. (#841)

FEATURES:

- **New Resource:** `vsphere_vnic` (#876)

BUG FIX:

- `r/virtual_machine`: Allow blank networkID in order to support cloning into clusters that do not include the source network. (#787)
- `r/host`: Properly handle situation where NIC teaming policy is `nil`. (#889)
- Limit scope when listing network interfaces. (#840)
- `r/compute_cluster`: Set HA Admission Control Failure to `off` before deleting. (#891)
- `r/virtual_machine_snapshot`: Fix typo in error condition. (#906)
- `tags`: Return matched tag rather than last tag in list. (#910)
- `r/virtual_machine`: Unmount ISO when switching CDROM backends. (#920)
- `r/virtual_machine`: Migrate VM when moving to different root resource pool. (#931)

## v1.13.0

> Release Date: 2019-10-01

IMPROVEMENTS:

- Add `vim_keep_alive` which sets a keepalive interval for VIM session. (#792)
- `r/virtual_machine`: Mark `windows_sysprep_text` as sensitive. (#802)

FEATURES:

- **New Resource:** `vsphere_host` (#836)

BUG FIX:

- `r/virtual_machine`: Change the way we detect if a VM is in a vApp. (#825)
- Delete tags and tag_categories when they are removed. (#801)

## v1.12.0

> Release Date: 2019-06-19

IMPROVEMENTS:

- `r/virtual_machine`: Allow cloning of powered on virtual machines. (#785)
- Add keep alive timer for VIM sessions. (#792)

BUG FIX:

- `r/virtual_machine`: Ignore validation when interpolation is not available. (#784)
- `r/virtual_machine`: Only set vApp properties that are UserConfigurable. (#751)
- `r/virtual_machine`: Set `network_id` to empty string when cloning a `virtual_machine` to a cluster that is not part of source DVS. (#787)

## v1.11.0

> Release Date: 2019-05-09

IMPROVEMENTS:

- Add support for importing datacenters. (#737)
- Document max character limit on `run_once_command_list`. (#748)
- Add missing ENV variable checks for acceptance tests. (#758)
- Switch to Terraform 0.12 SDK which is required for Terraform 0.12 support. This is the first release to use the 0.12 SDK required for Terraform 0.12 support. Some provider behavior may have changed as a result of changes made by the new SDK version. (#760)

## v1.10.0

> Release Date: 2019-03-15

FEATURES:

- **New Data Source:** `vsphere_folder` (#709)

IMPROVEMENTS:

- Update tf-vsphere-devrc.mk.example to include all environment variables (#707)
- Add Go Modules support (#705)
- Fix assorted typos in documentation
- `r/virtual_machine`: Add support for using guest.ipAddress for older versions of VM Tools. (#684)

BUG FIX:

- `r/virtual_machine`: Do not set optional `ignored_guest_ips` on read (#726)

## v1.9.1

> Release Date: 2019-01-10

IMPROVEMENTS:

- `r/virtual_machine`: Increase logging after old config expansion during diff checking (#661)
- `r/virtual_machine`: Unlock `memory_reservation` from maximum when `memory_reservation` is not equal to `memory`. (#680)

BUG FIX:

- `r/virtual_machine`: Return zero instead of nil for memory allocation and reservation values (#655)
- Ignore nil interfaces when converting a slice of interfaces into a slice of strings. (#666)
- `r/virtual_machine`: Use schema for `properties` elem definition in `vapp` schema. (#678)

## v1.9.0

> Release Date: 2018-10-31

FEATURES:

- **New Resource:** `vsphere_vapp_entity` (#640)
- `r/host_virtual_switch`: Add support for importing (#625)

IMPROVEMENTS:

- `r/virtual_disk`: Update existing and add additional tests (#635)

BUG FIX:

- `r/virtual_disk`: Ignore "already exists" errors when creating directories on vSAN. (#639)
- Find tag changes when first tag is changed. (#632)
- `r/virtual_machine`: Do not ForceNew when clone `timeout` is changed. (#631)
- `r/virtual_machine_snapshot`: Raise error on snapshot create task error. (#628)

## v1.8.1

> Release Date: 2018-09-11

IMPROVEMENTS:

- `d/vapp_container`: Re-add `data_source_vapp_container`. (#617)

## v1.8.0

> Release Date: 2018-09-10

FEATURES:

- **New Data Source:** `vsphere_vapp_container` (#610)

BUG FIX:

- `r/virtual_machine`: Only relocate after create if `host_system_id` is set and does not match host the VM currently resides on. (#609)
- `r/compute_cluster`: Return empty policy instead of trying to read `nil` variable when `ha_admission_control_policy` is set to `disabled`. (#611)
- `r/virtual_machine`: Skip reading latency sensitivity parameters when LatencySensitivity is `nil`. (#612)
- `r/compute_cluster`: Unset ID when the resource is not found. (#613)
- `r/virtual_machine`: Skip OS specific customization checks when `resource_pool_id` is not set. (#614)

## v1.7.0

> Release Date: 2018-08-24

FEATURES:

- **New Resource:** `vsphere_vapp_container` (#566)
- `r/virtual_machine`: Added support for bus sharing on SCSI adapters. (#574)

IMPROVEMENTS:

- `r/datacenter`: Added `moid` to expose the managed object ID because the datacenter's name is currently being used as the `id`. (#575)
- `r/virtual_machine`: Check if relocation is necessary after creation. (#583)

BUG FIX:

- `r/virtual_machine`: The resource no longer attempts to set ResourceAllocation on virtual ethernet cards when the vSphere version is under 6.0. (#579)
- `r/resource_pool`: The read function is now called at the end of resource creation. (#560)
- Updated govmomi to 0.18. (#600)

## v1.6.0

> Release Date: 2018-05-31

FEATURES:

- **New Resource:** `vsphere_resource_pool` (#535)

IMPROVEMENTS:

- `d/host`: Now exports the `resource_pool_id` attribute, which points to the root resource pool of either the standalone host, or the cluster's root resource pool in the event the host is a member of a cluster. (#535)

BUG FIX:

- `r/virtual_machine`: Scenarios that force a new resource will no longer create diff mismatches when external disks are attached with the `attach` parameter. (#528)

## v1.5.0

> Release Date: 2018-05-11

FEATURES:

- **New Data Source:** `vsphere_compute_cluster` (#492)
- **New Resource:** `vsphere_compute_cluster` (#487)
- **New Resource:** `vsphere_drs_vm_override` (#498)
- **New Resource:** `vsphere_ha_vm_override` (#501)
- **New Resource:** `vsphere_dpm_host_override` (#503)
- **New Resource:** `vsphere_compute_cluster_vm_group` (#506)
- **New Resource:** `vsphere_compute_cluster_host_group` (#508)
- **New Resource:** `vsphere_compute_cluster_vm_host_rule` (#511)
- **New Resource:** `vsphere_compute_cluster_vm_dependency_rule` (#513)
- **New Resource:** `vsphere_compute_cluster_vm_affinity_rule` (#515)
- **New Resource:** `vsphere_compute_cluster_vm_anti-affinity_rule` (#515)
- **New Resource:** `vsphere_datastore_cluster_vm_anti-affinity_rule` (#520)

IMPROVEMENTS:

- `r/virtual_machine`: Exposed `latency_sensitivity`, which can be used to adjust the scheduling priority of the virtual machine for low-latency applications. (#490)
- `r/virtual_disk`: Introduced the `create_directories` setting, which tells this resource to create any parent directories in the VMDK path. (#512)

## v1.4.1

> Release Date: 2018-04-23

IMPROVEMENTS:

- `r/virtual_machine`: Introduced the `wait_for_guest_net_routable` setting, which controls whether or not the guest network waiter waits on an address that matches the virtual machine's configured default gateway. (#470)

BUG FIX:

- `r/virtual_machine`: The resource now correctly blocks `clone` workflows on direct ESXi connections, where cloning is not supported. (#476)
- `r/virtual_machine`: Corrected an issue that was preventing VMs from being migrated from one cluster to another. (#474)
- `r/virtual_machine`: Corrected an issue where changing datastore information and cloning/customization parameters (which forces a new resource) at the same time was creating a diff mismatch after destroying the old virtual machine. (#469)
- `r/virtual_machine`: Corrected a crash that can come up from an incomplete lookup of network information during network device management. (#456)
- `r/virtual_machine`: Corrected some issues where some post-clone configuration errors were leaving the resource half-completed and irrecoverable without direct modification of the state. (#467)
- `r/virtual_machine`: Corrected a crash that can come up when a retrieved virtual machine has no lower-level configuration object in the API. (#463)
- `r/virtual_machine`: Fixed an issue where disk sub-resource configurations were not being checked for newly created disks. (#481)

## v1.4.0

> Release Date: 2018-04-10

FEATURES:

- **New Resource:** `vsphere_storage_drs_vm_override` (#450)
- **New Resource:** `vsphere_datastore_cluster` (#436)
- **New Data Source:** `vsphere_datastore_cluster` (#437)

IMPROVEMENTS:

- The provider now has the ability to persist sessions to disk, which can help when running large amounts of consecutive or concurrent Terraform operations at once. See the [provider documentation](https://www.terraform.io/docs/providers/vsphere/index.html) for more details. (#422)
- `r/virtual_machine`: This resource now supports import of resources or migrations from legacy versions of the provider (provider version 0.4.2 or earlier) into configurations that have the `clone` block specified. See [Additional requirements and notes for importing](https://www.terraform.io/docs/providers/vsphere/r/virtual_machine.html#additional-requirements-and-notes-for-importing) in the resource documentation for more details. (#460)
- `r/virtual_machine`: Now supports datastore clusters. Virtual machines placed in a datastore cluster will use Storage DRS recommendations for initial placement, virtual disk creation, and migration between datastore clusters. Migrations made by Storage DRS outside of Terraform will no longer create diffs when datastore clusters are in use. (#447)
- `r/virtual_machine`: Added support for ISO transport of vApp properties. The resource should now behave better with virtual machines cloned from OVF/OVA templates that use the ISO transport to supply configuration settings. (#381)
- `r/virtual_machine`: Added support for client mapped CDROM devices. (#421)
- `r/virtual_machine`: Destroying a VM that currently has external disks attached should now function correctly and not give a duplicate UUID error. (#442)
- `r/nas_datastore`: Now supports datastore clusters. (#439)
- `r/vmfs_datastore`: Now supports datastore clusters. (#439)

## v1.3.3

> Release Date: 2018-03-01

IMPROVEMENTS:

- `r/virtual_machine`: The `moid` attribute has now been re-added to the resource, exporting the managed object ID of the virtual machine. (#390)

BUG FIX:

- `r/virtual_machine`: Fixed a crash scenario that can happen when a virtual machine is deployed to a cluster that does not have any hosts, or under certain circumstances such as an expired vCenter license. (#414)
- `r/virtual_machine`: Corrected an issue reading disk capacity values after a vCenter or ESXi upgrade. (#405)
- `r/virtual_machine`: Opaque networks, such as those coming from NSX, should now be able to be correctly added as networks for virtual machines. (#398)

## v1.3.2

> Release Date: 2018-02-07

BUG FIX:

- `r/virtual_machine`: Changed the update implemented in (#377) to use a local filter implementation. This corrects situations where virtual machines in inventory with orphaned or otherwise corrupt configurations were interfering with UUID searches, creating erroneous duplicate UUID errors. This fix applies to Sphere 6.0 and lower only. vSphere 6.5 was not affected. (#391)

## v1.3.1

> Release Date: 2018-02-01

BUG FIX:

- `r/virtual_machine`: Looking up templates by their UUID now functions correctly for vSphere 6.0 and earlier. (#377)

## v1.3.0

> Release Date: 2018-01-26

BREAKING CHANGES:

- The `virtual_machine` resource now has a new method of identifying virtual disk sub-resources, via the `label` attribute. This replaces the `name` attribute, which has now been marked as deprecated and will be removed in the next major version (2.0.0). Further to this, there is a `path` attribute that now must also be supplied for external disks. This has lifted several virtual disk-related cloning and migration restrictions, in addition to changing requirements for importing. See the [resource documentation](https://www.terraform.io/docs/providers/vsphere/r/virtual_machine.html) for usage details.

IMPROVEMENTS:

- `r/virtual_machine`: Fixed an issue where certain changes happening at the same time (such as a disk resize along with a change of SCSI type) were resulting in invalid device change operations. (#371)
- `r/virtual_machine`: Introduced the `label` argument, which allows one to address a virtual disk independent of its VMDK file name and position on the SCSI bus. (#363)
- `r/virtual_machine`: Introduced the `path` argument, which replaces the `name` attribute for supplying the path for externally attached disks supplied with `attach = true`, and is otherwise a computed attribute pointing to the current path of any specific virtual disk. (#363)
- `r/virtual_machine`: Introduced the `uuid` attribute, a new computed attribute that allows the tracking of a disk independent of its current position on the SCSI bus. This is used in all scenarios aside from freshly-created or added virtual disks. (#363)
- `r/virtual_machine`: The virtual disk `name` argument is now deprecated and will be removed from future releases. It no longer dictates the name of non-attached VMDK files and serves as an alias to the now-split `label` and `path` attributes. (#363)
- `r/virtual_machine`: Cloning no longer requires you to choose a disk label (name) that matches the name of the VM. (#363)
- `r/virtual_machine`: Storage vMotion can now be performed on renamed virtual machines. (#363)
- `r/virtual_machine`: Storage vMotion no longer cares what your disks are labeled (named), and will not block migrations based on the naming criteria added after 1.1.1. (#363)
- `r/virtual_machine`: Storage vMotion now works on linked clones. (#363)
- `r/virtual_machine`: The import restrictions for virtual disks have changed, and rather than ensuring that disk `name` arguments match a certain convention, `label` is now expected to match a convention of `diskN`, where N is the disk number, ordered by the disk's position on the SCSI bus. Importing to a configuration still using `name` to address disks is no longer supported. (#363)
- `r/virtual_machine`: Now supports setting vApp properties that usually come from an OVF/OVA template or virtual appliance. (#303)

## v1.2.0

> Release Date: 2018-01-11

FEATURES:

- **New Resource:** `vsphere_custom_attribute` (#229)
- **New Data Source:** `vsphere_custom_attribute` (#229)

IMPROVEMENTS:

- All vSphere provider resources that are capable of doing so now support custom attributes. Check the documentation of any specific resource for more details! (#229)
- `r/virtual_machine`: The resource will now disallow a disk's `name` coming from a value that is still unavailable at plan time (such as a computed value from a resource). (#329)

BUG FIX:

- `r/virtual_machine`: Fixed an issue that was causing crashes when working with virtual machines or templates when no network interface was occupying the first available device slot on the PCI bus. (#344)

## v1.1.1

> Release Date: 2017-12-14

IMPROVEMENTS:

- `r/virtual_machine`: Network interface resource allocation options are now restricted to Sphere 6.0 and higher, as they are unsupported on vSphere 5.5. (#322)
- `r/virtual_machine`: Resources that were deleted outside of Terraform will now be marked as gone in the state, causing them to be re-created during the next apply. (#321)
- `r/virtual_machine`: Added some restrictions to storage vMotion to cover some currently un-supported scenarios that were still allowed, leading to potentially dangerous situations or invalid post-application states. (#319)
- `r/virtual_machine`: The resource now treats disks that it does not recognize at a known device address as orphaned, and will set `keep_on_remove` to safely remove them. (#317)
- `r/virtual_machine`: The resource now attempts to detect unsafe disk deletion scenarios that can happen from the renaming of a virtual machine in situations where the VM and disk names may share a common variable. The provider will block such operations from proceeding. (#305)

## v1.1.0

> Release Date: 2017-12-07

BREAKING CHANGES:

- The `virtual_machine` _data source_ has a new sub-resource attribute for disk information, named `disks`. This takes the place of `disk_sizes`, which has been moved to a `size` attribute within this new sub-resource, and also contains information about the discovered disks' `eagerly_scrub` and `thin_provisioned` settings. This is to facilitate the ability to discover all settings that could cause issues when cloning virtual machines.

To transition to the new syntax, any `disk` sub-resource in a `vsphere_virtual_machine` resource that depends on a syntax such as:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  disk {
    name = "terraform-test.vmdk"
    size = data.vsphere_virtual_machine.template.disk_sizes[0]
  }
}
```

Should be changed to:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  disk {
    name = "terraform-test.vmdk"
    size = data.vsphere_virtual_machine.template.disks.0.size
  }
}
```

If you are using `linked_clone`, add the new settings for `eagerly_scrub` and `thin_provisioned`:

```hcl
resource "vsphere_virtual_machine" "vm" {
  # ... other configuration ...

  disk {
    name             = "terraform-test.vmdk"
    thin_provisioned = data.vsphere_virtual_machine.template.disks.0.thin_provisioned
  }
}
```

For a more complete example, see the [cloning and customization example](https://www.terraform.io/docs/providers/vsphere/r/virtual_machine.html#cloning-and-customization-example) in the documentation.

BUG FIX:

- `r/virtual_machine`: Fixed a bug with NIC device assignment logic that was causing a crash when adding more than 3 NICs to a VM. (#280)
- `r/virtual_machine`: CDROM devices on cloned virtual machines are now connected properly on power on. (#278)
- `r/virtual_machine`: Tightened the pre-clone checks for virtual disks to ensure that the size and disk types are the same between the template and the created virtual machine's configuration. (#277)

## v1.0.3

> Release Date: 2017-12-06

BUG FIX:

- `r/virtual_machine`: Fixed an issue in the post-clone process when a CDROM device exists in configuration. (#276)

## v1.0.2

> Release Date: 2017-12-05

BUG FIX:

- `r/virtual_machine`: Fixed issues related to correct processing VM templates with no network interfaces, or fewer network interfaces than the amount that will ultimately end up in configuration. (#269)
- `r/virtual_machine`: Version comparison logic now functions correctly to properly disable certain features when using older versions of vSphere. (#272)

## v1.0.1

> Release Date: 2017-12-02

BUG FIX:

- `r/virtual_machine`: Corrected an issue that was preventing the use of this resource on standalone ESXi. (#263)
- `d/resource_pool`: This data source now works as documented on standalone ESXi. (#263)

## v1.0.0

> Release Date: 2017-12-01

BREAKING CHANGES:

- The `virtual_machine` resource has received a major update and change to its interface. See the documentation for the resource for full details, including information on things to consider while migrating the new version of the resource.

FEATURES:

- **New Data Source:** `vsphere_resource_pool` (#244)
- **New Data Source:** `vsphere_datastore` (#244)
- **New Data Source:** `vsphere_virtual_machine` (#244)

IMPROVEMENTS:

- `r/virtual_machine`: The distinct VM workflows are now better defined: all cloning options are now contained within a `clone` sub-resource, with customization being a `customize` sub-resource off of that. Absence of the `clone` sub-resource means no cloning or customization will occur. (#244)
- `r/virtual_machine`: Nearly all customization options have now been exposed. Magic values such as hostname and DNS defaults have been removed, with some of these options now being required values depending on the OS being customized. (#244)
- `r/virtual_machine`: Device management workflows have been greatly improved, exposing more options and fixing several bugs. (#244)
- `r/virtual_machine`: Added support for CPU and memory hot-plug. Several other VM reconfiguration operations are also supported while the VM is powered on, guest type and VMware Tools permitting in some cases. (#244)
- `r/virtual_machine`: The resource now supports both host and storage vMotion. Virtual machines can now be moved between hosts, clusters, resource pools, and datastores. Individual disks can be pinned to a single datastore with a VM located in another. (#244)
- `r/virtual_machine`: The resource now supports import. (#244)
- `r/virtual_machine`: Several other minor improvements, see documentation for more details. (#244)

BUG FIX:

- `r/virtual_machine`: Several long-standing issues have been fixed, namely surrounding virtual disk and network device management. (#244)
- `r/host_virtual_switch`: This resource now correctly supports a configuration with no NICs. (#256)
- `d/network`: No longer restricted to being used on vCenter. (#248)

## v0.4.2

> Release Date: 2017-10-13

FEATURES:

- **New Data Source:** `vsphere_network` (#201)
- **New Data Source:** `vsphere_distributed_virtual_switch` (#170)
- **New Resource:** `vsphere_distributed_port_group` (#189)
- **New Resource:** `vsphere_distributed_virtual_switch` (#188)

IMPROVEMENTS:

- `r/virtual_machine`: The customization waiter is now tunable through the `wait_for_customization_timeout` argument. The timeout can be adjusted or the waiter can be disabled altogether. (#199)
- `r/virtual_machine`: `domain` now acts as a default for `dns_suffix` if the latter is not defined, setting the value in `domain` as a search domain in the customization specification. `vsphere.local` is not used as a last resort only. (#185)
- `r/virtual_machine`: Expose the `adapter_type` parameter to allow the control of the network interface type. This is currently restricted to `vmxnet3` and `e1000` but offers more control than what was available before, and more interface types will follow in later versions of the provider. (#193)

BUG FIX:

- `r/virtual_machine`: Fixed a regression with network discovery that was causing Terraform to crash while the VM was in a powered-off state. (#198)
- All resources that can use tags will now properly remove their tags completely (or remove any out-of-band added tags) when the `tags` argument is not present in configuration. (#196)

## v0.4.1

> Release Date: 2017-10-02

BUG FIX:

- `r/folder`: Migration of state from a version of this resource before v0.4.0 now works correctly. (#187)

## v0.4.0

> Release Date: 2017-09-29

BREAKING CHANGES:

- The `folder` resource has been re-written, and its configuration is significantly different. See the [resource documentation](https://www.terraform.io/docs/providers/vsphere/r/folder.html) for more details. Existing state will be migrated. (#179)

FEATURES:

- **New Data Source:** `vsphere_tag` (#171)
- **New Data Source:** `vsphere_tag_category` (#167)
- **New Resource:** `vsphere_tag` (#171)
- **New Resource:** `vsphere_tag_category` (#164)

IMPROVEMENTS:

- `r/folder`: You can now create any kind of folder with this resource, not just virtual machine folders. (#179)
- `r/folder`: Now supports tags. (#179)
- `r/folder`: Now supports import. (#179)
- `r/datacenter`: Tags can now be applied to datacenters. (#177)
- `r/nas_datastore`: Tags can now be applied to NAS datastores. (#176)
- `r/vmfs_datastore`: Tags can now be applied to VMFS datastores. (#176)
- `r/virtual_machine`: Tags can now be applied to virtual machines. (#175)
- `r/virtual_machine`: Adjusted the customization timeout to 10 minutes (#168)

BUG FIX:

- `r/virtual_machine`: This resource can now be used with networks with unescaped slashes in its network name. (#181)
- `r/virtual_machine`: Fixed a crash where virtual NICs were created with networks backed by a 3rd party hardware VDS. (#181)
- `r/virtual_machine`: Fixed crashes and spurious diffs that were caused by errors in the code that associates the default gateway with its correct network device during refresh. (#180)

## v0.3.0

> Release Date: 2017-09-14

BREAKING CHANGES:

- `virtual_machine` now waits on a _routable_ IP address by default, and does not wait when running `terraform plan`, `terraform refresh`, or `terraform destroy`. There is also now a timeout of 5 minutes, after which `terraform apply` will fail with an error. Note that the apply may not fail exactly on the 5-minute mark. The network waiter can be disabled completely by setting `wait_for_guest_net` to `false`. (#158)

FEATURES:

- **New Resource:** `vsphere_virtual_machine_snapshot` (#107)

IMPROVEMENTS:

- `r/virtual_machine`: Virtual machine power state is now enforced. Terraform will trigger a diff if the VM is powered off or suspended, and power it back on during the next apply. (#152)

BUG FIX:

- `r/virtual_machine`: Fixed customization behavior to watch customization events for success, rather than returning immediately when the `CustomizeVM` task returns. This is especially important during Windows customization where a large part of the customization task involves out-of-band configuration through Sysprep. (#158)

## v0.2.2

> Release Date: 2017-09-07

FEATURES:

- **New Resource:** `vsphere_nas_datastore` (#149)
- **New Resource:** `vsphere_vmfs_datastore` (#142)
- **New Data Source:** `vsphere_vmfs_disks` (#141)

## v0.2.1

> Release Date: 2017-08-31

FEATURES:

- **New Resource:** `vsphere_host_port_group` (#139)
- **New Resource:** `vsphere_host_virtual_switch` (#138)
- **New Data Source:** `vsphere_datacenter` (#144)
- **New Data Source:** `vsphere_host` (#146)

IMPROVEMENTS:

- `r/virtual_machine`: Allow customization of hostname (#79)

BUG FIX:

- `r/virtual_machine`: Fix IPv4 address mapping issues causing spurious diffs, in addition to IPv6 normalization issues that can lead to spurious diffs as well. (#128)

## v0.2.0

> Release Date: 2017-08-23

BREAKING CHANGES:

- `r/virtual_disk`: Default adapter type is now `lsiLogic`, changed from `ide`. (#94)

FEATURES:

- **New Resource:** `vsphere_datacenter` (#126)
- **New Resource:** `vsphere_license` (#110)

IMPROVEMENTS:

- `r/virtual_machine`: Add annotation argument (#111)

BUG FIX:

- Updated [govmomi](https://github.com/vmware/govmomi) to 0.15.0 (#114)
- Updated network interface discovery behavior in refresh. (#129). This fixed several reported bugs - see the PR for references!

## v0.1.0

> Release Date: 2017-06-20

NOTES:

- Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
