---
subcategory: "Host and Cluster Management"
page_title: "VMware vSphere: vsphere_compute_cluster"
sidebar_current: "docs-vsphere-resource-compute-compute-cluster"
description: |-
  Provides a vSphere cluster resource. This can be used to create and manage clusters of hosts.
---

# vsphere_compute_cluster

-> **A note on the naming of this resource:** VMware refers to clusters of
hosts in the UI and documentation as _clusters_, _HA clusters_, or _DRS
clusters_. All of these refer to the same kind of resource (with the latter two
referring to specific features of clustering). In Terraform, we use
`vsphere_compute_cluster` to differentiate host clusters from _datastore
clusters_, which are clusters of datastores that can be used to distribute load
and ensure fault tolerance via distribution of virtual machines. Datastore
clusters can also be managed through Terraform, via the
[`vsphere_datastore_cluster` resource][docs-r-vsphere-datastore-cluster].

[docs-r-vsphere-datastore-cluster]: /docs/providers/vsphere/r/datastore_cluster.html

The `vsphere_compute_cluster` resource can be used to create and manage
clusters of hosts allowing for resource control of compute resources, load
balancing through DRS, and high availability through vSphere HA.

For more information on vSphere clusters and DRS, see [this
page][ref-vsphere-drs-clusters]. For more information on vSphere HA, see [this
page][ref-vsphere-ha-clusters].

[ref-vsphere-drs-clusters]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-resource-management-8-0/creating-a-drs-cluster.html
[ref-vsphere-ha-clusters]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-availability.html

~> **NOTE:** This resource requires vCenter and is not available on
direct ESXi connections.

## Example Usage

The following example sets up a cluster and enables DRS and vSphere HA with the
default settings. The hosts have to exist already in vSphere and should not
already be members of clusters - it's best to add these as standalone hosts
before adding them to a cluster.

Note that the following example assumes each host has been configured correctly
according to the requirements of vSphere HA. For more information, click
[here][ref-vsphere-ha-checklist].

[ref-vsphere-ha-checklist]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-availability.html

```hcl
variable "datacenter" {
  default = "dc-01"
}

variable "hosts" {
  default = [
    "esxi-01.example.com",
    "esxi-02.example.com",
    "esxi-03.example.com",
  ]
}

data "vsphere_datacenter" "datacenter" {
  name = var.datacenter
}

data "vsphere_host" "host" {
  for_each      = toset(var.hosts)
  name          = each.value
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "terraform-compute-cluster-test"
  datacenter_id   = data.vsphere_datacenter.datacenter.id
  host_system_ids = [for host in data.vsphere_host.host : host.id]

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster.
* `datacenter_id` - (Required) The [managed object ID][docs-about-morefs] of
  the datacenter to create the cluster in. Forces a new resource if changed.
* `folder` - (Optional) The relative path to a folder to put this cluster in.
  This is a path relative to the datacenter you are deploying the cluster to.
  Example: for the `dc1` datacenter, and a provided `folder` of `foo/bar`,
  Terraform will place a cluster named `terraform-compute-cluster-test` in a
  host folder located at `/dc1/host/foo/bar`, with the final inventory path
  being `/dc1/host/foo/bar/terraform-datastore-cluster-test`.
* `tags` - (Optional) The IDs of any tags to attach to this resource. See
  [here][docs-applying-tags] for a reference on how to apply tags.

[docs-about-morefs]: /docs/providers/vsphere/index.html#use-of-managed-object-references-by-the-vsphere-provider
[docs-applying-tags]: /docs/providers/vsphere/r/tag.html#using-tags-in-a-supported-resource

* `custom_attributes` - (Optional) A map of custom attribute ids to attribute
  value strings to set for the datastore cluster. See
  [here][docs-setting-custom-attributes] for a reference on how to set values
  for custom attributes.

[docs-setting-custom-attributes]: /docs/providers/vsphere/r/custom_attribute.html#using-custom-attributes-in-a-supported-resource

~> **NOTE:** Custom attributes are not supported on direct ESXi host
connections and requires vCenter Server.

### Host Management Options

The following settings control cluster membership or tune how hosts are managed
within the cluster itself by Terraform.

* `host_system_ids` - (Optional) The [managed object IDs][docs-about-morefs] of
  the hosts to put in the cluster. Conflicts with: `host_managed`.
* `host_managed` - (Optional) Can be set to `true` if compute cluster
  membership will be managed through the `host` resource rather than the
  `compute_cluster` resource. Conflicts with: `host_system_ids`.
* `host_cluster_exit_timeout` - The timeout, in seconds, for each host maintenance
  mode operation when removing hosts from a cluster. Default: `3600` seconds (1 hour).
* `force_evacuate_on_destroy` - When destroying the resource, setting this to
  `true` will auto-remove any hosts that are currently a member of the cluster,
  as if they were removed by taking their entry out of `host_system_ids` (see
  [below](#how-terraform-removes-hosts-from-clusters)). This is an advanced
  option and should only be used for testing. Default: `false`.

~> **NOTE:** Do not set `force_evacuate_on_destroy` in production operation as
there are many pitfalls to its use when working with complex cluster
configurations. Depending on the virtual machines currently on the cluster, and
your DRS and HA settings, the full host evacuation may fail. Instead,
incrementally remove hosts from your configuration by adjusting the contents of
the `host_system_ids` attribute.

#### How Terraform Removes ESXi Hosts from Clusters

You can remove hosts from clusters by adjusting the
[`host_system_ids`](#host_system_ids) configuration setting and removing the
hosts in question. Hosts are removed sequentially, by placing them in
maintenance mode, _moving them_ to the root host folder in vSphere inventory,
and then taking the host out of maintenance mode. This process, if successful,
preserves the host in vSphere inventory as a standalone host.

Note that whether or not this operation succeeds as intended depends on your
DRS and high availability settings. To ensure as much as possible that this
operation will succeed, ensure that no HA configuration depends on the host
_before_ applying the host removal operation, as host membership operations are
processed before configuration is applied. If there are VMs on the host, set
your [`drs_automation_level`](#drs_automation_level) to `fullyAutomated` to
ensure that DRS can correctly evacuate the host before removal.

Note that all virtual machines are migrated as part of the maintenance mode
operation, including ones that are powered off or suspended. Ensure there is
enough capacity on your remaining hosts to accommodate the extra load.

### DRS Automation Options

The following options control the settings for DRS on the cluster.

* `drs_enabled` - (Optional) Enable DRS for this cluster. Default: `false`.
* `drs_automation_level` (Optional) The default automation level for all
  virtual machines in this cluster. Can be one of `manual`,
  `partiallyAutomated`, or `fullyAutomated`. Default: `manual`.
* `drs_migration_threshold` - (Optional) A value between `1` and `5` indicating
  the threshold of imbalance tolerated between hosts. A lower setting will
  tolerate more imbalance while a higher setting will tolerate less. Default:
  `3`.
* `drs_enable_vm_overrides` - (Optional) Allow individual DRS overrides to be
  set for virtual machines in the cluster. Default: `true`.
* `drs_enable_predictive_drs` - (Optional) When `true`, enables DRS to use data
  from [VMware Aria Operations][ref-vmware-aria-operations] to make proactive DRS
  recommendations. <sup>[\*](#vsphere-version-requirements)</sup>

[ref-vmware-aria-operations]: https://techdocs.broadcom.com/us/en/vmware-cis/aria.html

* `drs_scale_descendants_shares` - (Optional) Enable scalable shares for all
  resource pools in the cluster. Can be one of `disabled` or
  `scaleCpuAndMemoryShares`. Default: `disabled`.
* `drs_advanced_options` - (Optional) A key/value map that specifies advanced
  options for DRS and [DPM](#dpm-options).

#### DPM Options

The following settings control the [Distributed Power
Management][ref-vsphere-dpm] (DPM) settings for the cluster. DPM allows the
cluster to manage host capacity on-demand depending on the needs of the
cluster, powering on hosts when capacity is needed, and placing hosts in
standby when there is excess capacity in the cluster.

[ref-vsphere-dpm]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-resource-management-8-0/using-drs-clusters-to-manage-resources/managing-power-resources.html

* `dpm_enabled` - (Optional) Enable DPM support for DRS in this cluster.
  Requires [`drs_enabled`](#drs_enabled) to be `true` in order to be effective.
  Default: `false`.
* `dpm_automation_level` - (Optional) The automation level for host power
  operations in this cluster. Can be one of `manual` or `automated`. Default:
  `manual`.
* `dpm_threshold` - (Optional) A value between `1` and `5` indicating the
  threshold of load within the cluster that influences host power operations.
  This affects both power on and power off operations - a lower setting will
  tolerate more of a surplus/deficit than a higher setting. Default: `3`.

### vSphere HA Options

The following settings control the [vSphere HA][ref-vsphere-ha-clusters]
settings for the cluster.

~> **NOTE:** vSphere HA has a number of requirements that should be met to
ensure that any configured settings work correctly. For a full list, see the
[vSphere HA Checklist][ref-vsphere-ha-checklist].

* `ha_enabled` - (Optional) Enable vSphere HA for this cluster. Default:
  `false`.
* `ha_host_monitoring` - (Optional) Global setting that controls whether
  vSphere HA remediates virtual machines on host failure. Can be one of `enabled`
  or `disabled`. Default: `enabled`.
* `ha_vm_restart_priority` - (Optional) The default restart priority
  for affected virtual machines when vSphere detects a host failure. Can be one
  of `lowest`, `low`, `medium`, `high`, or `highest`. Default: `medium`.
* `ha_vm_dependency_restart_condition` - (Optional) The condition used to
  determine whether or not virtual machines in a certain restart priority class
  are online, allowing HA to move on to restarting virtual machines on the next
  priority. Can be one of `none`, `poweredOn`, `guestHbStatusGreen`, or
  `appHbStatusGreen`. The default is `none`, which means that a virtual machine
  is considered ready immediately after a host is found to start it on.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_vm_restart_additional_delay` - (Optional) Additional delay, in seconds,
  after ready condition is met. A VM is considered ready at this point.
  Default: `0` seconds (no delay). <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_vm_restart_timeout` - (Optional) The maximum time, in seconds,
  that vSphere HA will wait for virtual machines in one priority to be ready
  before proceeding with the next priority. Default: `600` seconds (10 minutes).
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_host_isolation_response` - (Optional) The action to take on virtual
  machines when a host has detected that it has been isolated from the rest of
  the cluster. Can be one of `none`, `powerOff`, or `shutdown`. Default:
  `none`.
* `ha_advanced_options` - (Optional) A key/value map that specifies advanced
  options for vSphere HA.

#### HA Virtual Machine Component Protection Settings

The following settings control Virtual Machine Component Protection (VMCP) in
vSphere HA. VMCP gives vSphere HA the ability to monitor a host for datastore
accessibility failures, and automate recovery for affected virtual machines.

-> **Note on terminology:** In VMCP, Permanent Device Loss (PDL), or a failure
where access to a specific disk device is not recoverable, is differentiated
from an All Paths Down (APD) failure, which is used to denote a transient
failure where disk device access may eventually return. Take note of this when
tuning these options.

* `ha_vm_component_protection` - (Optional) Controls vSphere VM component
  protection for virtual machines in this cluster. Can be one of `enabled` or
  `disabled`. Default: `enabled`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_datastore_pdl_response` - (Optional) Controls the action to take on
  virtual machines when the cluster has detected a permanent device loss to a
  relevant datastore. Can be one of `disabled`, `warning`, or
  `restartAggressive`. Default: `disabled`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_datastore_apd_response` - (Optional) Controls the action to take on
  virtual machines when the cluster has detected loss to all paths to a
  relevant datastore. Can be one of `disabled`, `warning`,
  `restartConservative`, or `restartAggressive`.  Default: `disabled`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_datastore_apd_recovery_action` - (Optional) Controls the action to take
  on virtual machines if an APD status on an affected datastore clears in the
  middle of an APD event. Can be one of `none` or `reset`. Default: `none`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_datastore_apd_response_delay` - (Optional) The time, in seconds,
  to wait after an APD timeout event to run the response action defined in
  [`ha_datastore_apd_response`](#ha_datastore_apd_response). Default: `180`
  seconds (3 minutes). <sup>[\*](#vsphere-version-requirements)</sup>

#### HA Virtual Machine and Application Monitoring Settings

The following settings illustrate the options that can be set to work with
virtual machine and application monitoring on vSphere HA.

* `ha_vm_monitoring` - (Optional) The type of virtual machine monitoring to use
  when HA is enabled in the cluster. Can be one of `vmMonitoringDisabled`,
  `vmMonitoringOnly`, or `vmAndAppMonitoring`. Default: `vmMonitoringDisabled`.
* `ha_vm_failure_interval` - (Optional) The time interval, in seconds, a heartbeat
  from a virtual machine is not received within this configured interval,
  the virtual machine is marked as failed. Default: `30` seconds.
* `ha_vm_minimum_uptime` - (Optional) The time, in seconds, that HA waits after
  powering on a virtual machine before monitoring for heartbeats. Default:
  `120` seconds (2 minutes).
* `ha_vm_maximum_resets` - (Optional) The maximum number of resets that HA will
  perform to a virtual machine when responding to a failure event. Default: `3`
* `ha_vm_maximum_failure_window` - (Optional) The time, in seconds, for the reset window in
  which [`ha_vm_maximum_resets`](#ha_vm_maximum_resets) can operate. When this
  window expires, no more resets are attempted regardless of the setting
  configured in `ha_vm_maximum_resets`. `-1` means no window, meaning an
  unlimited reset time is allotted. Default: `-1` (no window).

#### vSphere HA Admission Control Settings

The following settings control vSphere HA Admission Control, which controls
whether or not specific VM operations are permitted in the cluster in order to
protect the reliability of the cluster. Based on the constraints defined in
these settings, operations such as power on or migration operations may be
blocked to ensure that enough capacity remains to react to host failures.

#### Admission Control Modes

The [`ha_admission_control_policy`](#ha_admission_control_policy) parameter
controls the specific mode that Admission Control uses. What settings are
available depends on the admission control mode:

* **Cluster resource percentage**: This is the default admission control mode,
  and allows you to specify a percentage of the cluster's CPU and memory
  resources to reserve as spare capacity, or have these settings automatically
  determined by failure tolerance levels. To use, set
  [`ha_admission_control_policy`](#ha_admission_control_policy) to
  `resourcePercentage`.
* **Slot Policy (powered-on VMs)**: This allows the definition of a virtual
  machine "slot", which is a set amount of CPU and memory resources that should
  represent the size of an average virtual machine in the cluster. To use, set
  [`ha_admission_control_policy`](#ha_admission_control_policy) to
  `slotPolicy`.
* **Dedicated failover hosts**: This allows the reservation of dedicated
  failover hosts. Admission Control will block access to these hosts for normal
  operation to ensure that they are available for failover events. In the event
  that a dedicated host does not enough capacity, hosts that are not part of
  the dedicated pool will still be used for overflow if possible. To use, set
  [`ha_admission_control_policy`](#ha_admission_control_policy) to
  `failoverHosts`.

It is also possible to disable Admission Control by setting
[`ha_admission_control_policy`](#ha_admission_control_policy) to `disabled`,
however this is not recommended as it can lead to issues with cluster capacity,
and instability with vSphere HA.

* `ha_admission_control_policy` - (Optional) The type of admission control
  policy to use with vSphere HA. Can be one of `resourcePercentage`,
  `slotPolicy`, `failoverHosts`, or `disabled`. Default: `resourcePercentage`.

#### Common Admission Control Settings

The following settings are available for all Admission Control modes, but will
infer different meanings in each mode.

* `ha_admission_control_host_failure_tolerance` - (Optional) The maximum number
  of failed hosts that admission control tolerates when making decisions on
  whether to permit virtual machine operations. The maximum is one less than
  the number of hosts in the cluster. Default: `1`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_admission_control_performance_tolerance` - (Optional) The percentage of
  resource reduction that a cluster of virtual machines can tolerate in case of
  a failover. A value of 0 produces warnings only, whereas a value of 100
  disables the setting. Default: `100` (disabled).

#### Admission Control Settings for Resource Percentage Mode

The following settings control specific settings for Admission Control when
`resourcePercentage` is selected in
[`ha_admission_control_policy`](#ha_admission_control_policy).

* `ha_admission_control_resource_percentage_auto_compute` - (Optional)
  Automatically determine available resource percentages by subtracting the
  average number of host resources represented by the
  [`ha_admission_control_host_failure_tolerance`](#ha_admission_control_host_failure_tolerance)
  setting from the total amount of resources in the cluster. Disable to supply
  user-defined values. Default: `true`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `ha_admission_control_resource_percentage_cpu` - (Optional) Controls the
  user-defined percentage of CPU resources in the cluster to reserve for
  failover. Default: `100`.
* `ha_admission_control_resource_percentage_memory` - (Optional) Controls the
  user-defined percentage of memory resources in the cluster to reserve for
  failover. Default: `100`.

#### Admission Control Settings for Slot Policy Mode

The following settings control specific settings for Admission Control when
`slotPolicy` is selected in
[`ha_admission_control_policy`](#ha_admission_control_policy).

* `ha_admission_control_slot_policy_use_explicit_size` - (Optional) Controls
  whether or not you wish to supply explicit values to CPU and memory slot
  sizes. The default is `false`, which tells vSphere to gather a automatic
  average based on all powered-on virtual machines currently in the cluster.
* `ha_admission_control_slot_policy_explicit_cpu` - (Optional) Controls the
  user-defined CPU slot size, in MHz. Default: `32`.
* `ha_admission_control_slot_policy_explicit_memory` - (Optional) Controls the
  user-defined memory slot size, in MB. Default: `100`.

#### Admission Control Settings for Dedicated Failover Hosts Mode

The following settings control specific settings for Admission Control when
`failoverHosts` is selected in
[`ha_admission_control_policy`](#ha_admission_control_policy).

* `ha_admission_control_failover_host_system_ids` - (Optional) Defines the
  [managed object IDs][docs-about-morefs] of hosts to use as dedicated failover
  hosts. These hosts are kept as available as possible - admission control will
  block access to the host, and DRS will ignore the host when making
  recommendations.

#### vSphere HA Datastore Settings

vSphere HA uses datastore heartbeats to determine the health of a particular
host. Depending on how your datastores are configured, the settings below may
need to be altered to ensure that specific datastores are used over others.

If you require a user-defined list of datastores, ensure you select either
`userSelectedDs` (for user selected only) or `allFeasibleDsWithUserPreference`
(for automatic selection with preferred overrides) for the
[`ha_heartbeat_datastore_policy`](#ha_heartbeat_datastore_policy) setting.

* `ha_heartbeat_datastore_policy` - (Optional) The selection policy for HA
  heartbeat datastores. Can be one of `allFeasibleDs`, `userSelectedDs`, or
  `allFeasibleDsWithUserPreference`. Default:
  `allFeasibleDsWithUserPreference`.
* `ha_heartbeat_datastore_ids` - (Optional) The list of managed object IDs for
  preferred datastores to use for HA heartbeats. This setting is only useful
  when [`ha_heartbeat_datastore_policy`](#ha_heartbeat_datastore_policy) is set
  to either `userSelectedDs` or `allFeasibleDsWithUserPreference`.

#### Proactive HA Settings

The following settings pertain to [Proactive HA][ref-vsphere-proactive-ha], an
advanced feature of vSphere HA that allows the cluster to get data from
external providers and make decisions based on the data reported.

Working with Proactive HA is outside the scope of this document. For more
details, see the referenced link in the above paragraph.

[ref-vsphere-proactive-ha]: https://techdocs.broadcom.com/us/en/vmware-cis/vsphere/vsphere/8-0/vsphere-availability.html

* `proactive_ha_enabled` - (Optional) Enables Proactive HA. Default: `false`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `proactive_ha_automation_level` - (Optional) Determines how the host
  quarantine, maintenance mode, or virtual machine migration recommendations
  made by proactive HA are to be handled. Can be one of `Automated` or
  `Manual`. Default: `Manual`. <sup>[\*](#vsphere-version-requirements)</sup>
* `proactive_ha_moderate_remediation` - (Optional) The configured remediation
  for moderately degraded hosts. Can be one of `MaintenanceMode` or
  `QuarantineMode`. Note that this cannot be set to `MaintenanceMode` when
  [`proactive_ha_severe_remediation`](#proactive_ha_severe_remediation) is set
  to `QuarantineMode`. Default: `QuarantineMode`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `proactive_ha_severe_remediation` - (Optional) The configured remediation for
  severely degraded hosts. Can be one of `MaintenanceMode` or `QuarantineMode`.
  Note that this cannot be set to `QuarantineMode` when
  [`proactive_ha_moderate_remediation`](#proactive_ha_moderate_remediation) is
  set to `MaintenanceMode`. Default: `QuarantineMode`.
  <sup>[\*](#vsphere-version-requirements)</sup>
* `proactive_ha_provider_ids` - (Optional) The list of IDs for health update
  providers configured for this cluster.
  <sup>[\*](#vsphere-version-requirements)</sup>

### vSAN Settings

* `vsan_enabled` - (Optional) Enables vSAN on the cluster.
* `vsan_esa_enabled` - (Optional) Enables vSAN ESA on the cluster.
* `vsan_unmap_enabled` - (Optional) Enables vSAN unmap on the cluster.
  You must explicitly enable vSAN unmap when you enable vSAN ESA on the cluster.
* `vsan_dedup_enabled` - (Optional) Enables vSAN deduplication on the cluster.
  Cannot be independently set to `true`. When vSAN deduplication is enabled, vSAN
  compression must also be enabled.
* `vsan_compression_enabled` - (Optional) Enables vSAN compression on the
  cluster.
* `vsan_performance_enabled` - (Optional) Enables vSAN performance service on
  the cluster. Default: `true`.
* `vsan_verbose_mode_enabled` - (Optional) Enables verbose mode for vSAN
  performance service on the cluster.
* `vsan_network_diagnostic_mode_enabled` - (Optional) Enables network
  diagnostic mode for vSAN performance service on the cluster.
* `vsan_remote_datastore_ids` - (Optional) The remote vSAN datastore IDs to be
  mounted to this cluster. Conflicts with `vsan_dit_encryption_enabled` and
  `vsan_dit_rekey_interval`, i.e., vSAN HCI Mesh feature cannot be enabled with
  data-in-transit encryption feature at the same time.
* `vsan_dit_encryption_enabled` - (Optional) Enables vSAN data-in-transit
  encryption on the cluster. Conflicts with `vsan_remote_datastore_ids`, i.e.,
  vSAN data-in-transit feature cannot be enabled with the vSAN HCI Mesh feature
  at the same time.
* `vsan_dit_rekey_interval` - (Optional) Indicates the rekey interval in
  minutes for data-in-transit encryption. The valid rekey interval is 30 to
  10800 (feature defaults to 1440). Conflicts with `vsan_remote_datastore_ids`.
* `vsan_disk_group` - (Optional) Represents the configuration of a host disk
  group in the cluster.
  * `cache` - The canonical name of the disk to use for vSAN cache.
  * `storage` - An array of disk canonical names for vSAN storage.
* `vsan_fault_domains` - (Optional) Configurations of vSAN fault domains.
  * `fault_domain` - The configuration for single fault domain.
    * `name` - The name of fault domain.
    * `host_ids` - The managed object IDs of the hosts to put in the fault domain.
* `vsan_stretched_cluster` - (Optional) Configurations of vSAN stretched cluster.
  * `preferred_fault_domain_host_ids` - The managed object IDs of the hosts to put in the first fault domain.
  * `secondary_fault_domain_host_ids` - The managed object IDs of the hosts to put in the second fault domain.
  * `witness_node` - The managed object IDs of the host selected as witness node when enable stretched cluster.
  * `preferred_fault_domain_name` - (Optional) The name of first fault domain. Default is `Preferred`.
  * `secondary_fault_domain_name` - (Optional) The name of second fault domain. Default is `Secondary`.

~> **NOTE:** You must disable vSphere HA before you enable vSAN on the cluster.
You can enable or re-enable vSphere HA after vSAN is configured.

```hcl
resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "terraform-compute-cluster-test"
  datacenter_id   = data.vsphere_datacenter.datacenter.id
  host_system_ids = [data.vsphere_host.host.*.id]

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = false

  vsan_enabled                         = true
  vsan_esa_enabled                     = true
  vsan_dedup_enabled                   = true
  vsan_compression_enabled             = true
  vsan_performance_enabled             = true
  vsan_verbose_mode_enabled            = true
  vsan_network_diagnostic_mode_enabled = true
  vsan_unmap_enabled                   = true
  vsan_dit_encryption_enabled          = true
  vsan_dit_rekey_interval              = 1800
  vsan_disk_group {
    cache   = data.vsphere_vmfs_disks.cache_disks[0]
    storage = data.vsphere_vmfs_disks.storage_disks
  }
  vsan_fault_domains {
    fault_domain {
      name     = "fd1"
      host_ids = [data.vsphere_host.faultdomain1_hosts.*.id]
    }
    fault_domain {
      name     = "fd2"
      host_ids = [data.vsphere_host.faultdomain2_hosts.*.id]
    }
  }
  vsan_stretched_cluster {
    preferred_fault_domain_host_ids = [data.vsphere_host.preferred_fault_domain_host.*.id]
    secondary_fault_domain_host_ids = [data.vsphere_host.secondary_fault_domain_host.*.id]
    witness_node                    = data.vsphere_host.witness_host.id
  }
}
```

### vLCM Settings

After vLCM is enabled on a cluster it is not possible to disable it.

* `host_image` - (Optional) Enables vLCM on the cluster and applies an ESXi image with the provided configuration.
  * `esx_version` - The ESXi version which will be used as a base for the image. See [`host_base_images`][docs-d-host-base-images].
  * `component` - A custom component to add to the base image. TODO - add link to offline depot resource docs
    * `key` - The identifier of the component
    * `version` - The version of the component

### EVC Settings

CPU-level EVC is configurable via the provider.

* `evc_mode` - Enhanced vMotion Compatibility mode. Removing the value for this attribute disables EVC.

## Attribute Reference

The following attributes are exported:

* `id`: The [managed object ID][docs-about-morefs] of the cluster.
* `resource_pool_id` The [managed object ID][docs-about-morefs] of the primary
  resource pool for this cluster. This can be passed directly to the
  [`resource_pool_id`
  attribute][docs-r-vsphere-virtual-machine-resource-pool-id] of the
  [`vsphere_virtual_machine`][docs-r-vsphere-virtual-machine] resource.

[docs-r-vsphere-virtual-machine-resource-pool-id]: /docs/providers/vsphere/r/virtual_machine.html#resource_pool_id
[docs-r-vsphere-virtual-machine]: /docs/providers/vsphere/r/virtual_machine.html
[docs-d-host-base-images]: /docs/providers/vsphere/d/host_base_images.html

## Importing

An existing cluster can be [imported][docs-import] into this resource via the
path to the cluster, via the following command:

[docs-import]: https://developer.hashicorp.com/terraform/cli/import

```hcl
variable "datacenter" {
  default = "dc-01"
}

data "vsphere_datacenter" "datacenter" {
  name = var.datacenter
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}
```

~> **NOTE:** When you import a cluster, all managed settings are returned. Ensure all settings are set correctly in resource. For example:

```hcl
resource "vsphere_compute_cluster" "compute_cluster" {
  name                      = "cluster-01"
  datacenter_id             = data.vsphere_datacenter.datacenter.id
  vsan_enabled              = true
  vsan_performance_enabled  = true
  host_system_ids           = [for host in data.vsphere_host.host : host.id]
  dpm_automation_level      = "automated"
  drs_automation_level      = "fullyAutomated"
  drs_enabled               = true
  ha_datastore_apd_response = "restartConservative"
  ha_datastore_pdl_response = "restartAggressive"
}
```

All information will be added to the Terraform state after import.

```shell
terraform import vsphere_compute_cluster.compute_cluster /dc-01/host/cluster-01
```

The above would import the cluster named `cluster-01` that is located in
the `dc-01` datacenter.

## vSphere Version Requirements

Some settings in the `vsphere_compute_cluster` resource may require a
specific version of vSphere.

### Settings that Require vSphere 7.0 or higher

These settings require vSphere 7.0 or higher:

* [`drs_scale_descendants_shares`](#drs_scale_descendants_shares)

### Settings that Require vSphere 8.0 or higher

These settings require vSphere 8.0 or higher:

* [`vsan_esa_enabled`](#vsan_esa_enabled)
