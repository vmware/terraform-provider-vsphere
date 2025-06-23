// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/structure"
)

func resourceVsphereSupervisor() *schema.Resource {
	return &schema.Resource{
		Create: resourceVsphereSupervisorCreate,
		Read:   resourceVsphereSupervisorRead,
		Update: resourceVsphereSupervisorUpdate,
		Delete: resourceVsphereSupervisorDelete,
		Schema: map[string]*schema.Schema{
			"cluster": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ID of the vSphere cluster on which workload management will be enabled.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"storage_policy": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of a storage policy associated with the datastore where the container images will be stored.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"management_network": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The configuration for the management network which the control plane VMs will be connected to.",
				MaxItems:    1,
				Elem:        mgmtNetworkSchema(),
			},
			"content_library": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ID of the subscribed content library.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"main_dns": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of DNS servers to use on the Kubernetes API server.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"main_ntp": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of NTP servers to use on the Kubernetes API server.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"worker_ntp": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of NTP servers to use on the worker nodes.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"worker_dns": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of DNS servers to use on the worker nodes.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"edge_cluster": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ID of the NSX Edge Cluster.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"dvs_uuid": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The UUID (not ID) of the distributed switch.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"sizing_hint": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Size of the Kubernetes API server.",
				ValidateFunc: validation.StringInSlice([]string{"TINY", "SMALL", "MEDIUM", "LARGE"}, false),
			},
			"egress_cidr": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "CIDR blocks from which NSX assigns IP addresses used for performing SNAT from container IPs to external IPs.",
				Elem:        cidrSchema(),
			},
			"ingress_cidr": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "CIDR blocks from which NSX assigns IP addresses for Kubernetes Ingresses and Kubernetes Services of type LoadBalancer.",
				Elem:        cidrSchema(),
			},
			"pod_cidr": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "CIDR blocks from which Kubernetes allocates pod IP addresses. Minimum subnet size is 23.",
				Elem:        cidrSchema(),
			},
			"service_cidr": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "CIDR block from which Kubernetes allocates service cluster IP addresses.",
				MaxItems:    1,
				MinItems:    1,
				Elem:        cidrSchema(),
			},
			"search_domains": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "List of DNS search domains.",
				MaxItems:    1,
				MinItems:    1,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
			"namespace": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of namespaces associated with the cluster.",
				Elem:        namespaceSchema(),
			},
		},
	}
}

func mgmtNetworkSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "ID of the network. (e.g. a distributed port group).",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"starting_address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Starting address of the management network range.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"subnet_mask": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Subnet mask.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"gateway": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Gateway IP address.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"address_count": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Number of addresses to allocate. Starts from 'starting_address'",
				ValidateFunc: validation.IntAtLeast(1),
			},
		},
	}
}

func cidrSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Network address.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"prefix": {
				Type:         schema.TypeInt,
				Required:     true,
				Description:  "Subnet prefix.",
				ValidateFunc: validation.IntBetween(0, 32),
			},
		},
	}
}

func namespaceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the namespace.",
			},
			"content_libraries": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of content libraries.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vm_classes": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of virtual machine classes.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVsphereSupervisorCreate(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	clusterID := d.Get("cluster").(string)

	spec := buildClusterEnableSpec(d)

	if err := m.EnableCluster(context.Background(), clusterID, spec); err != nil {
		return err
	}

	d.SetId(clusterID)

	if err := waitForSupervisorEnable(m, d); err != nil {
		return err
	}

	namespaces := d.Get("namespace").(*schema.Set).List()

	for _, ns := range namespaces {
		nsData := ns.(map[string]interface{})

		if err := createNamespace(m, nsData, d.Id()); err != nil {
			return err
		}
	}

	return nil
}

func resourceVsphereSupervisorRead(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	cluster := getClusterByID(m, d.Id())

	if cluster == nil {
		return fmt.Errorf("could not find cluster %s", cluster.ID)
	}

	// To fully implement "read" and allow this resource to be imported
	// the following API has to be added to govmomi.
	// https://developer.broadcom.com/xapis/vsphere-automation-api/latest/vcenter/api/vcenter/namespace-management/clusters/cluster/get/
	return nil
}

func resourceVsphereSupervisorUpdate(d *schema.ResourceData, meta interface{}) error {
	if d.HasChange("namespace") {
		c := meta.(*Client).restClient
		m := namespace.NewManager(c)

		oldRaw, newRaw := d.GetChange("namespace")
		oldSet := oldRaw.(*schema.Set)
		newSet := newRaw.(*schema.Set)

		// these haven't changed, we don't need to do anything with them
		intersection := oldSet.Intersection(newSet)
		oldNamespaces := getNamespacesMap(oldSet.Difference(intersection).List())
		newNamespaces := getNamespacesMap(newSet.Difference(intersection).List())

		for k, v := range oldNamespaces {
			if _, found := newNamespaces[k]; found {
				// Namespace still exists but has changed
				if err := updateNamespace(m, v.(map[string]interface{})); err != nil {
					return err
				}
			} else {
				// Namespace no longer exists
				if err := deleteNamespace(m, v.(map[string]interface{})); err != nil {
					return err
				}
			}
		}

		for k, v := range newNamespaces {
			if _, found := oldNamespaces[k]; !found {
				// This is a new namespace
				if err := createNamespace(m, v.(map[string]interface{}), d.Id()); err != nil {
					return err
				}
			}
		}

		return nil
	}

	return fmt.Errorf("only updates to namespaces are supported")
}

func resourceVsphereSupervisorDelete(d *schema.ResourceData, meta interface{}) error {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	if err := m.DisableCluster(context.Background(), d.Id()); err != nil {
		return err
	}

	return waitForSupervisorDisable(m, d)
}

func buildClusterEnableSpec(d *schema.ResourceData) *namespace.EnableClusterSpec {
	egressCidrs := d.Get("egress_cidr").([]interface{})
	ingressCidrs := d.Get("ingress_cidr").([]interface{})
	podCidrs := d.Get("pod_cidr").([]interface{})
	ncpNetworkSpec := namespace.NcpClusterNetworkSpec{
		NsxEdgeCluster:           d.Get("edge_cluster").(string),
		ClusterDistributedSwitch: d.Get("dvs_uuid").(string),
		EgressCidrs:              getCidrs(egressCidrs),
		IngressCidrs:             getCidrs(ingressCidrs),
		PodCidrs:                 getCidrs(podCidrs),
	}

	contentLib := d.Get("content_library").(string)
	mainDNS := d.Get("main_dns").([]interface{})
	workerDNS := d.Get("worker_dns").([]interface{})
	mainNtp := d.Get("main_ntp").([]interface{})
	workerNtp := d.Get("worker_ntp").([]interface{})
	dnsSearchDomains := d.Get("search_domains").([]interface{})
	storagePolicy := d.Get("storage_policy").(string)
	serviceCidrs := d.Get("service_cidr").([]interface{})

	spec := &namespace.EnableClusterSpec{
		EphemeralStoragePolicy: storagePolicy,
		SizeHint:               getSizingHint(d.Get("sizing_hint")),
		ServiceCidr:            &getCidrs(serviceCidrs)[0],
		// Only NSX-T backing is supported for now
		NetworkProvider:                        &namespace.NsxtContainerPluginNetworkProvider,
		MasterStoragePolicy:                    storagePolicy,
		MasterManagementNetwork:                getMgmtNetwork(d),
		ImageStorage:                           namespace.ImageStorageSpec{StoragePolicy: storagePolicy},
		NcpClusterNetworkSpec:                  &ncpNetworkSpec,
		MasterDNS:                              structure.SliceInterfacesToStrings(mainDNS),
		WorkerDNS:                              structure.SliceInterfacesToStrings(workerDNS),
		MasterNTPServers:                       structure.SliceInterfacesToStrings(mainNtp),
		WorkloadNTPServers:                     structure.SliceInterfacesToStrings(workerNtp),
		DefaultKubernetesServiceContentLibrary: contentLib,
		MasterDNSSearchDomains:                 structure.SliceInterfacesToStrings(dnsSearchDomains),
	}
	return spec
}

func getMgmtNetwork(d *schema.ResourceData) *namespace.MasterManagementNetwork {
	mgmtNetworkProperty := d.Get("management_network").([]interface{})
	mgmtNetworkData := mgmtNetworkProperty[0].(map[string]interface{})
	return &namespace.MasterManagementNetwork{
		Mode:    &namespace.StaticRangeIpAssignmentMode,
		Network: mgmtNetworkData["network"].(string),
		AddressRange: &namespace.AddressRange{
			SubnetMask:      mgmtNetworkData["subnet_mask"].(string),
			StartingAddress: mgmtNetworkData["starting_address"].(string),
			Gateway:         mgmtNetworkData["gateway"].(string),
			AddressCount:    mgmtNetworkData["address_count"].(int),
		},
	}
}

func getCidrs(data []interface{}) []namespace.Cidr {
	result := make([]namespace.Cidr, len(data))
	for i, cidrData := range data {
		cidr := cidrData.(map[string]interface{})
		result[i] = namespace.Cidr{
			Address: cidr["address"].(string),
			Prefix:  cidr["prefix"].(int),
		}
	}
	return result
}

func getSizingHint(data interface{}) *namespace.SizingHint {
	switch data {
	case "TINY":
		return &namespace.TinySizingHint
	case "SMALL":
		return &namespace.SmallSizingHint
	case "MEDIUM":
		return &namespace.MediumSizingHint
	case "LARGE":
		return &namespace.LargeSizingHint
	}

	return &namespace.UndefinedSizingHint
}

func waitForSupervisorEnable(m *namespace.Manager, d *schema.ResourceData) error {
	ticker := time.NewTicker(time.Minute * time.Duration(1))
	failureCount := 0

	for {
		select {
		case <-context.Background().Done():
		case <-ticker.C:
			cluster := getClusterByID(m, d.Id())

			if cluster == nil {
				return fmt.Errorf("could not find cluster %s", cluster.ID)
			}

			if namespace.RunningConfigStatus == *cluster.ConfigStatus {
				return nil
			}
			if namespace.ErrorConfigStatus == *cluster.ConfigStatus {
				// The supervisor sometimes reports errors but manages to recover
				// We will only give up if we get several consecutive errors
				if failureCount > 3 {
					return fmt.Errorf("could not enable supervisor on cluster %s", cluster.ID)
				}
				failureCount++
			}
			if namespace.ConfiguringConfigStatus == *cluster.ConfigStatus {
				// Reset error counter
				failureCount = 0
			}
		}
	}
}

func waitForSupervisorDisable(m *namespace.Manager, d *schema.ResourceData) error {
	ticker := time.NewTicker(time.Minute * time.Duration(1))

	for {
		select {
		case <-context.Background().Done():
		case <-ticker.C:
			cluster := getClusterByID(m, d.Id())

			if cluster == nil {
				return nil
			}

			if namespace.ErrorConfigStatus == *cluster.ConfigStatus {
				return fmt.Errorf("could not disable supervisor on cluster %s", cluster.ID)
			}
		}
	}
}

func getClusterByID(m *namespace.Manager, id string) *namespace.ClusterSummary {
	clusters, err := m.ListClusters(context.Background())

	if err != nil {
		return nil
	}

	for _, cluster := range clusters {
		if id == cluster.ID {
			return &cluster
		}
	}

	return nil
}

func getNamespacesMap(namespaces []interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for _, n := range namespaces {
		name := n.(map[string]interface{})["name"].(string)
		result[name] = n
	}

	return result
}

func createNamespace(m *namespace.Manager, nsData map[string]interface{}, cluster string) error {
	namespaceSpec := namespace.NamespacesInstanceCreateSpec{
		Namespace:     nsData["name"].(string),
		Cluster:       cluster,
		VmServiceSpec: namespace.VmServiceSpec{},
	}

	if contentLibs, contains := nsData["content_libraries"]; contains {
		namespaceSpec.VmServiceSpec.ContentLibraries = structure.SliceInterfacesToStrings(contentLibs.([]interface{}))
	}

	if vmClasses, contains := nsData["vm_classes"]; contains {
		namespaceSpec.VmServiceSpec.VmClasses = structure.SliceInterfacesToStrings(vmClasses.([]interface{}))
	}

	return m.CreateNamespace(context.Background(), namespaceSpec)
}

func updateNamespace(m *namespace.Manager, nsData map[string]interface{}) error {
	spec := namespace.NamespacesInstanceUpdateSpec{
		VmServiceSpec: namespace.VmServiceSpec{},
	}

	if contentLibs, contains := nsData["content_libraries"]; contains {
		spec.VmServiceSpec.ContentLibraries = structure.SliceInterfacesToStrings(contentLibs.([]interface{}))
	}

	if vmClasses, contains := nsData["vm_classes"]; contains {
		spec.VmServiceSpec.VmClasses = structure.SliceInterfacesToStrings(vmClasses.([]interface{}))
	}

	return m.UpdateNamespace(context.Background(), nsData["name"].(string), spec)
}

func deleteNamespace(m *namespace.Manager, nsData map[string]interface{}) error {
	return m.DeleteNamespace(context.Background(), nsData["name"].(string))
}
