// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package supervisor

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
)

func EnableSupervisorSingleZone(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) (string, error) {
	cluster := d.Get("cluster").(string)
	name := d.Get("name").(string)

	cp, err := buildControlPlane(d)
	if err != nil {
		return "", err
	}

	wld, err := buildWorkloads(d)
	if err != nil {
		return "", err
	}

	spec := &namespace.EnableOnComputeClusterSpec{
		Name:         name,
		ControlPlane: cp,
		Workloads:    wld,
	}

	return m.EnableOnComputeCluster(ctx, cluster, spec)
}

func EnableSupervisorMultiZone(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) (string, error) {
	name := d.Get("name").(string)
	zones := d.Get("zones").([]interface{})

	cp, err := buildControlPlane(d)
	if err != nil {
		return "", err
	}

	wld, err := buildWorkloads(d)
	if err != nil {
		return "", err
	}

	spec := namespace.EnableOnZonesSpec{
		Name:         name,
		Zones:        structure.SliceInterfacesToStrings(zones),
		ControlPlane: cp,
		Workloads:    wld,
	}

	return m.EnableOnZones(ctx, &spec)
}

func WaitForSupervisorEnable(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) diag.Diagnostics {
	ticker := time.NewTicker(time.Minute * time.Duration(1))
	failureCount := 0

	for {
		select {
		case <-ctx.Done():
		case <-ticker.C:
			info, err := m.GetSupervisorSummary(ctx, d.Id())

			if err != nil {
				return diag.FromErr(fmt.Errorf("could not find supervisor %s, %s", d.Id(), err))
			}

			if namespace.RunningConfigStatus == info.ConfigStatus {
				return nil
			}
			if namespace.ErrorConfigStatus == info.ConfigStatus {
				// The supervisor sometimes reports errors but manages to recover
				// We will only give up after a recovery tolerance interval of 5 minutes
				if failureCount > 4 {
					return diag.FromErr(fmt.Errorf("could not enable supervisor %s", d.Id()))
				}
				failureCount++
			}
			if namespace.ConfiguringConfigStatus == info.ConfigStatus {
				// Reset error counter
				failureCount = 0
			}
		}
	}
}

func WaitForSupervisorDisable(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) diag.Diagnostics {
	ticker := time.NewTicker(time.Minute * time.Duration(1))

	for {
		select {
		case <-ctx.Done():
		case <-ticker.C:
			info, err := m.GetSupervisorSummary(ctx, d.Id())

			if err != nil {
				// Supervisor not found, we're done
				return nil
			}

			if namespace.ErrorConfigStatus == info.ConfigStatus {
				return diag.FromErr(fmt.Errorf("could not disable supervisor %s", d.Id()))
			}
		}
	}
}

func buildControlPlane(d *schema.ResourceData) (namespace.ControlPlane, error) {
	controlPlaneProperty := d.Get("control_plane").([]interface{})
	controlPlaneData := controlPlaneProperty[0].(map[string]interface{})
	networkData := controlPlaneData["network"].([]interface{})

	result := namespace.ControlPlane{}

	network, err := buildControlPlaneNetwork(networkData[0].(map[string]interface{}))
	if err != nil {
		return result, err
	}

	result.Network = network

	if count := controlPlaneData["count"].(int); count > 0 {
		result.Count = &count
	}

	if size := controlPlaneData["size"].(string); size != "" {
		result.Size = &size
	}

	if storagePolicy := controlPlaneData["storage_policy"].(string); storagePolicy != "" {
		result.StoragePolicy = &storagePolicy
	}

	return result, nil
}

func buildWorkloads(d *schema.ResourceData) (namespace.Workloads, error) {
	workloads := d.Get("workloads").([]interface{})
	workloadsData := workloads[0].(map[string]interface{})
	network := workloadsData["network"].([]interface{})
	edge := workloadsData["edge"].([]interface{})
	kubeAPIServerOptions := workloadsData["kube_api_server_options"].([]interface{})

	result := namespace.Workloads{
		KubeAPIServerOptions: buildKubeAPIServerOptions(kubeAPIServerOptions[0].(map[string]interface{})),
	}

	var err error
	result.Network, err = buildWorkloadNetwork(network[0].(map[string]interface{}))
	if err != nil {
		return result, err
	}

	result.Edge, err = buildEdge(edge[0].(map[string]interface{}))
	if err != nil {
		return result, err
	}

	if images, found := getOptionalNestedAttribute(d.Get("images")); found {
		value := buildImages(images)
		result.Images = &value
	}

	if storage, found := getOptionalNestedAttribute(d.Get("storage")); found {
		value := buildStorage(storage)
		result.Storage = &value
	}

	return result, nil
}

func buildWorkloadNetwork(workloadNetworkData map[string]interface{}) (namespace.WorkloadNetwork, error) {
	result := namespace.WorkloadNetwork{}

	if network := workloadNetworkData["network"].(string); network != "" {
		result.Network = &network
	}

	numNetworks := 0

	if vsphere, found := getOptionalNestedAttribute(workloadNetworkData["vsphere"]); found {
		result.NetworkType = "VSPHERE"
		result.VSphere = &namespace.NetworkVSphere{
			DVPG: vsphere["dvpg"].(string),
		}
		numNetworks++
	}

	if nsx, found := getOptionalNestedAttribute(workloadNetworkData["nsx"]); found {
		result.NetworkType = "NSXT"
		result.NSX = &namespace.NetworkNSX{
			DVS:                   nsx["dvs"].(string),
			NamespaceSubnetPrefix: Intptr(nsx["namespace_subnet_prefix"].(int)),
		}
		numNetworks++
	}

	if nsxVpc, found := getOptionalNestedAttribute(workloadNetworkData["nsx_vpc"]); found {
		result.NetworkType = "NSX_VPC"
		result.NSXVPC = &namespace.NetworkVPC{
			NSXProject:             Strptr(nsxVpc["nsx_project"].(string)),
			VPCConnectivityProfile: Strptr(nsxVpc["vpc_connectivity_profile"].(string)),
		}

		if defaultPrivateCidrs := nsxVpc["default_private_cidr"].([]interface{}); len(defaultPrivateCidrs) > 0 {
			result.NSXVPC.DefaultPrivateCIDRs = make([]namespace.Ipv4Cidr, len(defaultPrivateCidrs))
			for i, defaultPrivateCidr := range defaultPrivateCidrs {
				data := defaultPrivateCidr.(map[string]interface{})
				result.NSXVPC.DefaultPrivateCIDRs[i] = namespace.Ipv4Cidr{
					Address: data["address"].(string),
					Prefix:  data["prefix"].(int),
				}
			}
		}
		numNetworks++
	}

	var err error
	if numNetworks > 1 {
		err = fmt.Errorf("workload can only have one type of network")
	}

	if services, found := getOptionalNestedAttribute(workloadNetworkData["services"]); found {
		result.Services = buildControlPlaneNetworkServices(services)
	}

	if ipManagement, found := getOptionalNestedAttribute(workloadNetworkData["ip_management"]); found {
		result.IPManagement = buildControlPlaneNetworkIPManagement(ipManagement)
	}

	return result, err
}

func buildEdge(edgeData map[string]interface{}) (namespace.Edge, error) {
	result := namespace.Edge{}

	if id := edgeData["id"].(string); id != "" {
		result.ID = Strptr(id)
	}

	if lbAddressRangeData := edgeData["lb_address_range"]; lbAddressRangeData != nil {
		if lbAddressRange := lbAddressRangeData.([]interface{}); len(lbAddressRange) > 0 {
			ranges := make([]namespace.IPRange, len(lbAddressRange))

			for i, v := range lbAddressRange {
				data := v.(map[string]interface{})
				ranges[i] = namespace.IPRange{
					Address: data["address"].(string),
					Count:   data["count"].(int),
				}
			}

			result.LoadBalancerAddressRanges = &ranges
		}
	}

	numLoadBalancers := 0
	if foundation, found := getOptionalNestedAttribute(edgeData["foundation"]); found {
		result.Provider = Strptr("VSPHERE_FOUNDATION")
		value := buildFoundation(foundation)
		result.Foundation = &value
		numLoadBalancers++
	}

	if haproxy, found := getOptionalNestedAttribute(edgeData["haproxy"]); found {
		result.Provider = Strptr("HAPROXY")
		value := buildEdgeHAProxy(haproxy)
		result.HAProxy = &value
		numLoadBalancers++
	}

	if nsx, found := getOptionalNestedAttribute(edgeData["nsx"]); found {
		result.Provider = Strptr("NSX")
		value := buildNsx(nsx)
		result.NSX = &value
		numLoadBalancers++
	}

	if nsxAdvanced, found := getOptionalNestedAttribute(edgeData["nsx_advanced"]); found {
		result.Provider = Strptr("NSX_ADVANCED")
		value := buildNsxAdvanced(nsxAdvanced)
		result.NSXAdvanced = &value
		numLoadBalancers++
	}

	var err error
	if numLoadBalancers > 1 {
		err = fmt.Errorf("edge can only have one type of load balancer")
	}

	return result, err
}

func buildEdgeHAProxy(haproxyData map[string]interface{}) namespace.HAProxy {
	serversData := haproxyData["server"].([]interface{})
	servers := make([]namespace.EdgeServer, len(serversData))
	for i, v := range serversData {
		vd := v.(map[string]interface{})
		servers[i] = namespace.EdgeServer{
			Host: vd["host"].(string),
			Port: vd["port"].(int),
		}
	}

	return namespace.HAProxy{
		Username:                  haproxyData["username"].(string),
		Password:                  haproxyData["password"].(string),
		CertificateAuthorityChain: haproxyData["ca_chain"].(string),
		Servers:                   servers,
	}
}

func buildNsx(nsxData map[string]interface{}) namespace.EdgeNSX {
	result := namespace.EdgeNSX{}

	if edgeCluster := nsxData["edge_cluster"].(string); edgeCluster != "" {
		result.EdgeClusterID = Strptr(edgeCluster)
	}

	if lbSize := nsxData["load_balancer_size"].(string); lbSize != "" {
		result.LoadBalancerSize = Strptr(lbSize)
	}

	if routingMode := nsxData["routing_mode"].(string); routingMode != "" {
		result.RoutingMode = Strptr(routingMode)
	}

	if t0Gateway := nsxData["t0_gateway"].(string); t0Gateway != "" {
		result.T0Gateway = Strptr(t0Gateway)
	}

	if tlsCert := nsxData["default_ingress_tls_certificate"].(string); tlsCert != "" {
		result.DefaultIngressTLSCertificate = Strptr(tlsCert)
	}

	if ipRanges := nsxData["ip_range"]; ipRanges != nil {
		rawData := ipRanges.([]interface{})
		value := make([]namespace.IPRange, len(rawData))

		for i, ipRange := range rawData {
			data := ipRange.(map[string]interface{})
			value[i] = namespace.IPRange{
				Address: data["address"].(string),
				Count:   data["count"].(int),
			}
		}

		result.EgressIPRanges = &value
	}

	return result
}

func buildNsxAdvanced(advancedLBData map[string]interface{}) namespace.NSXAdvancedLBConfig {
	result := namespace.NSXAdvancedLBConfig{
		Username:                  advancedLBData["username"].(string),
		Password:                  advancedLBData["password"].(string),
		CertificateAuthorityChain: advancedLBData["ca_chain"].(string),
		Server: namespace.EdgeServer{
			Host: advancedLBData["host"].(string),
			Port: advancedLBData["port"].(int),
		},
	}

	if cloudName := advancedLBData["cloud_name"].(string); cloudName != "" {
		result.CloudName = Strptr(cloudName)
	}

	return result
}

func buildFoundation(foundationData map[string]interface{}) namespace.VSphereFoundationConfig {
	result := namespace.VSphereFoundationConfig{}

	if deploymentTarget, found := getOptionalNestedAttribute(foundationData["deployment_target"]); found {
		result.DeploymentTarget = &namespace.DeploymentTarget{}

		if zones := deploymentTarget["zones"].([]interface{}); len(zones) > 0 {
			value := structure.SliceInterfacesToStrings(zones)
			result.DeploymentTarget.Zones = &value
		}

		if availability := deploymentTarget["availability"].(string); availability != "" {
			result.DeploymentTarget.Availability = &availability
		}

		if deploymentSize := deploymentTarget["deployment_size"].(string); deploymentSize != "" {
			result.DeploymentTarget.DeploymentSize = &deploymentSize
		}

		if storagePolicy := deploymentTarget["storage_policy"].(string); storagePolicy != "" {
			result.DeploymentTarget.StoragePolicy = &storagePolicy
		}
	}

	if foundationInterfaces := foundationData["interface"].([]interface{}); len(foundationInterfaces) > 0 {
		value := make([]namespace.NetworkInterface, len(foundationInterfaces))

		for i, v := range foundationInterfaces {
			data := v.(map[string]interface{})
			value[i] = buildInterface(data)
		}

		result.Interfaces = &value
	}

	if networkServices, found := getOptionalNestedAttribute(foundationData["network_services"]); found {
		value := buildEdgeNetworkServices(networkServices)
		result.NetworkServices = &value
	}

	return result
}

func buildEdgeNetworkServices(networkServicesData map[string]interface{}) namespace.EdgeNetworkServices {
	result := namespace.EdgeNetworkServices{}

	if dns, found := getOptionalNestedAttribute(networkServicesData["dns"]); found {
		result.DNS = &namespace.DNS{
			Servers:       structure.SliceInterfacesToStrings(dns["servers"].([]interface{})),
			SearchDomains: structure.SliceInterfacesToStrings(dns["search_domains"].([]interface{})),
		}
	}

	if ntp, found := getOptionalNestedAttribute(networkServicesData["ntp"]); found {
		result.NTP = &namespace.NTP{
			Servers: structure.SliceInterfacesToStrings(ntp["servers"].([]interface{})),
		}
	}

	if syslog, found := getOptionalNestedAttribute(networkServicesData["syslog"]); found {
		result.Syslog = &namespace.Syslog{}

		if endpoint := syslog["endpoint"].(string); endpoint != "" {
			result.Syslog.Endpoint = &endpoint
		}

		if caCert := syslog["ca_cert"].(string); caCert != "" {
			result.Syslog.CertificateAuthorityPEM = &caCert
		}
	}

	return result
}

func buildInterface(interfaceData map[string]interface{}) namespace.NetworkInterface {
	networkData := interfaceData["network"].([]interface{})
	return namespace.NetworkInterface{
		Personas: structure.SliceInterfacesToStrings(interfaceData["personas"].([]interface{})),
		Network:  buildInterfaceNetwork(networkData[0].(map[string]interface{})),
	}
}

func buildInterfaceNetwork(networkData map[string]interface{}) namespace.NetworkInterfaceNetwork {
	result := namespace.NetworkInterfaceNetwork{
		NetworkType: networkData["network_type"].(string),
	}

	if dvpgNetwork, found := getOptionalNestedAttribute(networkData["dvpg_network"]); found {
		value := buildDvpgNetwork(dvpgNetwork)
		result.DVPGNetwork = &value
	}

	return result
}

func buildDvpgNetwork(dvpgNetwork map[string]interface{}) namespace.DVPGNetwork {
	result := namespace.DVPGNetwork{
		Name:    dvpgNetwork["name"].(string),
		Network: dvpgNetwork["network"].(string),
		IPAM:    dvpgNetwork["ipam"].(string),
	}

	if ipConfig, found := getOptionalNestedAttribute(dvpgNetwork["ip_config"]); found {
		result.IPConfig = &namespace.IPConfig{
			Gateway: ipConfig["gateway"].(string),
		}

		if ipRanges := ipConfig["ip_range"]; ipRanges != nil {
			data := ipRanges.([]interface{})
			result.IPConfig.IPRanges = make([]namespace.IPRange, len(data))

			for i, ipRange := range data {
				dataMap := ipRange.(map[string]interface{})
				result.IPConfig.IPRanges[i] = namespace.IPRange{
					Address: dataMap["address"].(string),
					Count:   dataMap["count"].(int),
				}
			}
		}
	}

	return result
}

func buildControlPlaneNetwork(controlPlaneNetworkData map[string]interface{}) (namespace.ControlPlaneNetwork, error) {
	backingProperty := controlPlaneNetworkData["backing"].([]interface{})
	backingData := backingProperty[0].(map[string]interface{})
	backing, err := buildControlPlaneNetworkBacking(backingData)
	if err != nil {
		return namespace.ControlPlaneNetwork{}, err
	}

	result := namespace.ControlPlaneNetwork{
		Backing: backing,
	}

	if network := controlPlaneNetworkData["network"].(string); network != "" {
		result.Network = &network
	}

	if floatingIP := controlPlaneNetworkData["floating_ip"].(string); floatingIP != "" {
		result.FloatingIPAddress = &floatingIP
	}

	if services, found := getOptionalNestedAttribute(controlPlaneNetworkData["services"]); found {
		result.Services = buildControlPlaneNetworkServices(services)
	}

	if ipManagement, found := getOptionalNestedAttribute(controlPlaneNetworkData["ip_management"]); found {
		result.IPManagement = buildControlPlaneNetworkIPManagement(ipManagement)
	}

	if proxy, found := getOptionalNestedAttribute(controlPlaneNetworkData["proxy"]); found {
		result.Proxy = &namespace.Proxy{
			ProxySettingsSource: proxy["settings_source"].(string),
		}

		isClusterConfigured := result.Proxy.ProxySettingsSource == "CLUSTER_CONFIGURED"

		if httpConfig := proxy["http_config"].(string); httpConfig != "" {
			if isClusterConfigured {
				return result, fmt.Errorf("`http_config` cannot be specified if `settings_source` is `CLUSTER_CONFIGURED`")
			}
			result.Proxy.HTTPProxyConfig = &httpConfig
		}

		if httpsConfig := proxy["https_config"].(string); httpsConfig != "" {
			if isClusterConfigured {
				return result, fmt.Errorf("`https_config` cannot be specified if `settings_source` is `CLUSTER_CONFIGURED`")
			}
			result.Proxy.HTTPSProxyConfig = &httpsConfig
		}

		if tlsBundle := proxy["tls_root_ca_bundle"].(string); tlsBundle != "" {
			if isClusterConfigured {
				return result, fmt.Errorf("`tls_root_ca_bundle` cannot be specified if `settings_source` is `CLUSTER_CONFIGURED`")
			}
			result.Proxy.TLSRootCABundle = &tlsBundle
		}

		if noProxyConf := proxy["no_proxy_config"]; noProxyConf != nil {
			if isClusterConfigured {
				return result, fmt.Errorf("`no_proxy_config` cannot be specified if `settings_source` is `CLUSTER_CONFIGURED`")
			}
			value := structure.SliceInterfacesToStrings(noProxyConf.([]interface{}))
			result.Proxy.NoProxyConfig = &value
		}
	}

	return result, nil
}

func buildControlPlaneNetworkBacking(backingData map[string]interface{}) (namespace.Backing, error) {
	result := namespace.Backing{}
	numBackings := 0

	if network := backingData["network"].(string); network != "" {
		result.Network = &network
		result.Backing = "NETWORK"
		numBackings++
	}

	if segments := backingData["segments"].([]interface{}); len(segments) > 0 {
		result.NetworkSegment = &namespace.NetworkSegment{
			Networks: structure.SliceInterfacesToStrings(segments),
		}
		result.Backing = "NETWORK_SEGMENT"
		numBackings++
	}

	var err error
	if numBackings > 1 {
		err = fmt.Errorf("control plane configuration cannot specify both `network` and `segments`")
	}

	return result, err
}

func buildControlPlaneNetworkServices(servicesData map[string]interface{}) *namespace.Services {
	result := namespace.Services{}

	if dns, found := getOptionalNestedAttribute(servicesData["dns"]); found {
		result.DNS = &namespace.DNS{
			Servers:       structure.SliceInterfacesToStrings(dns["servers"].([]interface{})),
			SearchDomains: structure.SliceInterfacesToStrings(dns["search_domains"].([]interface{})),
		}
	}

	if ntp, found := getOptionalNestedAttribute(servicesData["ntp"]); found {
		result.NTP = &namespace.NTP{
			Servers: structure.SliceInterfacesToStrings(ntp["servers"].([]interface{})),
		}
	}

	return &result
}

func buildControlPlaneNetworkIPManagement(ipManagementData map[string]interface{}) *namespace.IPManagement {
	result := namespace.IPManagement{
		DHCPEnabled: Boolptr(ipManagementData["dhcp_enabled"].(bool)),
	}

	if gatewayAddress := ipManagementData["gateway_address"].(string); gatewayAddress != "" {
		result.GatewayAddress = Strptr(gatewayAddress)
	}

	if ipAssignments := ipManagementData["ip_assignment"].([]interface{}); len(ipAssignments) > 0 {
		value := make([]namespace.IPAssignment, len(ipAssignments))

		for i, ipAssignment := range ipAssignments {
			ipAssignmentData := ipAssignment.(map[string]interface{})
			value[i] = buildIPAssignment(ipAssignmentData)
		}

		result.IPAssignments = &value
	}

	return &result
}

func buildIPAssignment(ipManagementData map[string]interface{}) namespace.IPAssignment {
	rangesData := ipManagementData["range"].([]interface{})
	result := namespace.IPAssignment{
		Assignee: Strptr(ipManagementData["assignee"].(string)),
		Ranges:   make([]namespace.IPRange, len(rangesData)),
	}

	for i, ipRange := range rangesData {
		ipRangeData := ipRange.(map[string]interface{})
		result.Ranges[i] = namespace.IPRange{
			Count:   ipRangeData["count"].(int),
			Address: ipRangeData["address"].(string),
		}
	}

	return result
}

func buildKubeAPIServerOptions(kubeAPIServerData map[string]interface{}) namespace.KubeAPIServerOptions {
	result := namespace.KubeAPIServerOptions{}

	if security, found := getOptionalNestedAttribute(kubeAPIServerData["security"]); found {
		result.Security = &namespace.KubeAPIServerSecurity{
			CertificateDNSNames: structure.SliceInterfacesToStrings(security["certificate_dns_names"].([]interface{})),
		}
	}

	return result
}

func buildImages(imagesData map[string]interface{}) namespace.Images {
	registryData := imagesData["registry"].([]interface{})
	contentLibraryData := imagesData["content_library"].([]interface{})
	contentLibraries := make([]namespace.ContentLibrary, len(contentLibraryData))

	for i, v := range contentLibraryData {
		contentLibraries[i] = buildContentLibrary(v.(map[string]interface{}))
	}

	return namespace.Images{
		Registry:                 buildRegistry(registryData[0].(map[string]interface{})),
		Repository:               imagesData["repository"].(string),
		KubernetesContentLibrary: imagesData["kubernetes_content_library"].(string),
		ContentLibraries:         contentLibraries,
	}
}

func buildContentLibrary(libraryData map[string]interface{}) namespace.ContentLibrary {
	result := namespace.ContentLibrary{
		ContentLibrary: libraryData["content_library"].(string),
	}

	if supervisorServices := libraryData["supervisor_services"].([]interface{}); len(supervisorServices) > 0 {
		value := structure.SliceInterfacesToStrings(supervisorServices)
		result.SupervisorServices = &value
	}

	if resourceNamingStrategy := libraryData["resource_naming_strategy"].(string); resourceNamingStrategy != "" {
		result.ResourceNamingStrategy = &resourceNamingStrategy
	}

	return result
}

func buildRegistry(registryData map[string]interface{}) namespace.Registry {
	return namespace.Registry{
		Hostname:         registryData["hostname"].(string),
		Port:             registryData["port"].(int),
		Username:         registryData["username"].(string),
		Password:         registryData["password"].(string),
		CertificateChain: registryData["certificate_chain"].(string),
	}
}

func buildStorage(storageData map[string]interface{}) namespace.WorkloadsStorageConfig {
	result := namespace.WorkloadsStorageConfig{}

	if ephemeralPolicy := storageData["ephemeral_storage_policy"].(string); ephemeralPolicy != "" {
		result.EphemeralStoragePolicy = &ephemeralPolicy
	}

	if imagePolicy := storageData["image_storage_policy"].(string); imagePolicy != "" {
		result.ImageStoragePolicy = &imagePolicy
	}

	if cloudNativeFileVolumes, found := getOptionalNestedAttribute(storageData["cloud_native_file_volume"]); found {
		result.CloudNativeFileVolume = &namespace.CloudNativeFileVolume{
			VSANClusters: structure.SliceInterfacesToStrings(cloudNativeFileVolumes["vsan_clusters"].([]interface{})),
		}
	}

	return result
}

func getOptionalNestedAttribute(rawValue interface{}) (map[string]interface{}, bool) {
	if rawValue == nil {
		return nil, false
	}

	rawAttr := rawValue.([]interface{})
	if len(rawAttr) == 0 {
		return nil, false
	}

	return rawAttr[0].(map[string]interface{}), true
}

func Strptr(s string) *string {
	return &s
}

func Intptr(i int) *int {
	return &i
}

func Boolptr(b bool) *bool {
	return &b
}
