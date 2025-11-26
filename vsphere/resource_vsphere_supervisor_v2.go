// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
)

func resourceVsphereSupervisorV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVsphereSupervisorV2Create,
		ReadContext:   resourceVsphereSupervisorV2Read,
		UpdateContext: resourceVsphereSupervisorV2Update,
		DeleteContext: resourceVsphereSupervisorV2Delete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cluster": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"zones"},
				ValidateFunc:  validation.StringIsNotEmpty,
			},
			"zones": {
				Type:          schema.TypeList,
				Optional:      true,
				ConflictsWith: []string{"cluster"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
			"control_plane": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     controlPlaneSchema(),
			},
			"workloads": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     workloadsSchema(),
			},
		},
	}
}

func controlPlaneSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"count": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"size": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"TINY", "SMALL", "MEDIUM", "LARGE"}, false),
			},
			"storage_policy": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"login_banner": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"network": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     controlPlaneNetworkSchema(),
			},
		},
	}
}

func controlPlaneNetworkSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"floating_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"backing": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     controlPlaneNetworkBackingSchema(),
			},
			"services": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     networkServicesSchema(),
			},
			"ip_management": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     ipManagementSchema(),
			},
			"proxy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     controlPlaneNetworkProxySchema(),
			},
		},
	}
}

func controlPlaneNetworkBackingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"backing": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"network": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"segments": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func networkServicesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dns": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     dnsSchema(),
			},
			"ntp": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     ntpSchema(),
			},
		},
	}
}

func ipManagementSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dhcp_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"gateway_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"ip_assignment": {
				Type: schema.TypeList,
				Elem: ipAssignmentSchema(),
			},
		},
	}
}

func ipAssignmentSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"assignee": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"range": {
				Type:     schema.TypeString,
				Optional: true,
				Elem:     ipRangeSchema(),
			},
		},
	}
}

func ipRangeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.IsIPv4Address,
			},
			"count": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func controlPlaneNetworkProxySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"settings_source": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"http_config": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"https_config": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"tls_root_ca_bundle": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"no_proxy_config": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
			},
		},
	}
}

func workloadsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     workloadNetworkSchema(),
			},
			"edge": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     edgeSchema(),
			},
			"kube_api_server_options": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"certificate_dns_names": {
										Type:     schema.TypeList,
										Required: true,
										Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
									},
								},
							},
						},
					},
				},
			},
			"images": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     imagesSchema(),
			},
			"storage": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     storageSchema(),
			},
		},
	}
}

func workloadNetworkSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"network_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"NSXT", "NSX_VPC", "VSPHERE"}, false),
			},
			"vsphere": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem:     workloadNetworkVSphereSchema(),
			},
			"nsx": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem:     workloadNetworkNSXSchema(),
			},
			"nsx_vpc": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem:     workloadNetworkNSXVPCSchema(),
			},
			"services": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem:     networkServicesSchema(),
			},
			"ip_management": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Elem:     ipManagementSchema(),
			},
		},
	}
}

func workloadNetworkVSphereSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dvpg": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func workloadNetworkNSXSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dvs": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_subnet_prefix": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func workloadNetworkNSXVPCSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"nsx_project": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"vpc_connectivity_profile": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"default_private_cidr": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     ipv4CidrSchema(),
			},
		},
	}
}

func ipv4CidrSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"address": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"prefix": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func ntpSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"servers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
			},
		},
	}
}

func dnsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"servers": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
			},
			"search_domains": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func edgeSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"lb_address_range": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     ipRangeSchema(),
			},
			"provider": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"foundation": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem:     edgeFoundationSchema(),
			},
			"haproxy": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     haProxySchema(),
			},
			"advanced_lb": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     nsxAdvancedLBSchema(),
			},
		},
	}
}

func edgeFoundationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"deployment_target": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"zones": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
						},
						"deployment_size": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"SMALL", "MEDIUM", "LARGE", "X_LARGE"}, false),
						},
						"storage_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
			"interface": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     edgeFoundationNetworkInterfaceSchema(),
			},
			"network_services": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     edgeNetworkServicesSchema(),
			},
		},
	}
}

func edgeFoundationNetworkInterfaceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"personas": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
			},
			"network": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     edgeFoundationNetworkInterfaceNetworkSchema(),
			},
		},
	}
}

func edgeFoundationNetworkInterfaceNetworkSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"dvpg_network": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     edgeFoundationNetworkInterfaceNetworkDvpgSchema(),
			},
		},
	}
}

func edgeFoundationNetworkInterfaceNetworkDvpgSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"network": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ipam": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ip_config": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_ranges": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     ipRangeSchema(),
						},
						"gateway": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
					},
				},
			},
		},
	}
}

func edgeNetworkServicesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dns": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     dnsSchema(),
			},
			"ntp": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     ntpSchema(),
			},
			"syslog": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem:     syslogSchema(),
			},
		},
	}
}

func haProxySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     edgeServerSchema(),
			},
			"username": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ca_chain": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func nsxAdvancedLBSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     edgeServerSchema(),
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ca_chain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cloud_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func edgeServerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func syslogSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"endpoint": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cert_authority_pem": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func imagesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"registry": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem:     registrySchema(),
			},
			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"kubernetes_content_library": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"content_library": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_library": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"supervisor_services": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
						},
						"resource_naming_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
		},
	}
}

func registrySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"port": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"certificate_chain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func storageSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"ephemeral_storage_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"image_storage_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cloud_native_file_volume": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vsan_clusters": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
						},
					},
				},
			},
		},
	}
}

func resourceVsphereSupervisorV2Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	zones := d.Get("zones").([]interface{})
	cluster := d.Get("cluster").(string)

	if len(zones) > 0 {
		if id, err := enableSupervisorMultiZone(ctx, d, m); err != nil {
			return diag.FromErr(err)
		} else {
			d.SetId(id)
		}
	} else if cluster != "" {
		if id, err := enableSupervisorSingleZone(ctx, d, m); err != nil {
			return diag.FromErr(err)
		} else {
			d.SetId(id)
		}
	} else {
		return diag.Errorf("either 'zones' or 'cluster' must be specified")
	}

	if err := waitForSupervisorEnable(m, d); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVsphereSupervisorV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := resourceVsphereSupervisorRead(d, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVsphereSupervisorV2Update(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := resourceVsphereSupervisorUpdate(d, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVsphereSupervisorV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if err := resourceVsphereSupervisorDelete(d, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func enableSupervisorSingleZone(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) (string, error) {
	cluster := d.Get("cluster").(string)
	name := d.Get("name").(string)
	spec := &namespace.EnableOnComputeClusterSpec{
		Name:         name,
		ControlPlane: buildControlPlane(d),
		Workloads:    buildWorkloads(d),
	}

	return m.EnableOnComputeCluster(ctx, cluster, spec)
}

func enableSupervisorMultiZone(ctx context.Context, d *schema.ResourceData, m *namespace.Manager) (string, error) {
	name := d.Get("name").(string)
	zones := d.Get("zones").([]interface{})
	spec := namespace.EnableOnZonesSpec{
		Name:         name,
		Zones:        structure.SliceInterfacesToStrings(zones),
		ControlPlane: buildControlPlane(d),
		Workloads:    buildWorkloads(d),
	}

	return m.EnableOnZones(ctx, &spec)
}

func buildControlPlane(d *schema.ResourceData) namespace.ControlPlane {
	controlPlaneProperty := d.Get("control_plane").([]interface{})
	controlPlaneData := controlPlaneProperty[0].(map[string]interface{})

	result := namespace.ControlPlane{
		Network: buildControlPlaneNetwork(controlPlaneData),
	}

	if count := controlPlaneData["count"].(int); count > 0 {
		result.Count = &count
	}

	if size := controlPlaneData["size"].(string); size != "" {
		result.Size = &size
	}

	if storagePolicy := controlPlaneData["storage_policy"].(string); storagePolicy != "" {
		result.StoragePolicy = &storagePolicy
	}

	if loginBanner := controlPlaneData["login_banner"].(string); loginBanner != "" {
		result.LoginBanner = &loginBanner
	}

	return result
}

func buildWorkloads(d *schema.ResourceData) namespace.Workloads {
	network := d.Get("network").([]interface{})
	edge := d.Get("edge").([]interface{})
	kubeAPIServerOptions := d.Get("kube_api_server_options").([]interface{})
	result := namespace.Workloads{
		Network:              buildWorkloadNetwork(network[0].(map[string]interface{})),
		Edge:                 buildEdge(edge[0].(map[string]interface{})),
		KubeAPIServerOptions: buildKubeAPIServerOptions(kubeAPIServerOptions[0].(map[string]interface{})),
	}

	if images := d.Get("images").([]interface{}); len(images) > 0 {
		value := buildImages(images[0].(map[string]interface{}))
		result.Images = &value
	}

	if storage := d.Get("storage").([]interface{}); len(storage) > 0 {
		value := buildStorage(storage[0].(map[string]interface{}))
		result.Storage = &value
	}

	return result
}

func buildWorkloadNetwork(workloadNetworkData map[string]interface{}) namespace.WorkloadNetwork {
	result := namespace.WorkloadNetwork{
		NetworkType: workloadNetworkData["network_type"].(string),
	}

	if network := workloadNetworkData["network"].(string); network != "" {
		result.Network = &network
	}

	if vsphere := workloadNetworkData["vsphere"].([]interface{}); len(vsphere) > 0 {
		vsphereData := vsphere[0].(map[string]interface{})
		result.VSphere = &namespace.NetworkVSphere{
			DVPG: vsphereData["dvpg"].(string),
		}
	}

	if nsx := workloadNetworkData["nsx"].([]interface{}); len(nsx) > 0 {
		nsxData := nsx[0].(map[string]interface{})
		result.NSX = &namespace.NetworkNSX{
			DVS:                   nsxData["dvs"].(string),
			NamespaceSubnetPrefix: Intptr(nsxData["namespace_subnet_prefix"].(int)),
		}
	}

	if nsxVpc := workloadNetworkData["nsx_vpc"].([]interface{}); len(nsxVpc) > 0 {
		nsxVpcData := nsxVpc[0].(map[string]interface{})
		result.NSXVPC = &namespace.NetworkVPC{
			NSXProject:             Strptr(nsxVpcData["nsx_project"].(string)),
			VPCConnectivityProfile: Strptr(nsxVpcData["vpc_connectivity_profile"].(string)),
		}

		if defaultPrivateCidrs := nsxVpcData["default_private_cidrs"].([]interface{}); len(defaultPrivateCidrs) > 0 {
			result.NSXVPC.DefaultPrivateCIDRs = make([]namespace.Ipv4Cidr, len(defaultPrivateCidrs))
			for i, defaultPrivateCidr := range defaultPrivateCidrs {
				data := defaultPrivateCidr.(map[string]interface{})
				result.NSXVPC.DefaultPrivateCIDRs[i] = namespace.Ipv4Cidr{
					Address: data["address"].(string),
					Prefix:  data["prefix"].(int),
				}
			}
		}
	}

	if services := workloadNetworkData["services"].([]interface{}); len(services) > 0 {
		result.Services = buildControlPlaneNetworkServices(services[0].(map[string]interface{}))
	}

	if ipManagement := workloadNetworkData["ip_management"].([]interface{}); len(ipManagement) > 0 {
		result.IPManagement = buildControlPlaneNetworkIPManagement(ipManagement[0].(map[string]interface{}))
	}

	return result
}

func buildEdge(edgeData map[string]interface{}) namespace.Edge {
	result := namespace.Edge{}

	if id := edgeData["id"].(string); id != "" {
		result.ID = Strptr(id)
	}

	if provider := edgeData["provider"].(string); provider != "" {
		result.ID = Strptr(provider)
	}

	if lbAddressRange := edgeData["lb_address_range"].([]interface{}); len(lbAddressRange) > 0 {
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

	if foundation := edgeData["foundation"].([]interface{}); len(foundation) > 0 {
		value := buildFoundation(foundation[0].(map[string]interface{}))
		result.Foundation = &value
	}

	if foundation := edgeData["foundation"].([]interface{}); len(foundation) > 0 {
		value := buildFoundation(foundation[0].(map[string]interface{}))
		result.Foundation = &value
	}

	if haproxy := edgeData["haproxy"].([]interface{}); len(haproxy) > 0 {
		value := buildEdgeHAProxy(haproxy[0].(map[string]interface{}))
		result.HAProxy = &value
	}

	if advancedLb := edgeData["advanced_lb"].([]interface{}); len(advancedLb) > 0 {
		value := buildEdgeAdvancedLB(advancedLb[0].(map[string]interface{}))
		result.NSXAdvanced = &value
	}

	return result
}

func buildEdgeHAProxy(haproxyData map[string]interface{}) namespace.HAProxy {
	serversData := haproxyData["servers"].([]interface{})
	servers := make([]namespace.EdgeServer, len(serversData))
	for i, v := range serversData {
		servers[i] = buildEdgeServer(v.(map[string]interface{}))
	}

	return namespace.HAProxy{
		Username:                  haproxyData["username"].(string),
		Password:                  haproxyData["password"].(string),
		CertificateAuthorityChain: haproxyData["ca_chain"].(string),
		Servers:                   servers,
	}
}

func buildEdgeAdvancedLB(advancedLBData map[string]interface{}) namespace.NSXAdvancedLBConfig {
	serversData := advancedLBData["servers"].([]interface{})
	serverData := serversData[0].(map[string]interface{})

	result := namespace.NSXAdvancedLBConfig{
		Username:                  advancedLBData["username"].(string),
		Password:                  advancedLBData["password"].(string),
		CertificateAuthorityChain: advancedLBData["ca_chain"].(string),
		Server:                    buildEdgeServer(serverData),
	}

	if cloudName := advancedLBData["cloud_name"].(string); cloudName != "" {
		result.CloudName = Strptr(cloudName)
	}

	return result
}

func buildEdgeServer(edgeServerData map[string]interface{}) namespace.EdgeServer {
	return namespace.EdgeServer{
		Host: edgeServerData["host"].(string),
		Port: edgeServerData["port"].(int),
	}
}

func buildFoundation(foundationData map[string]interface{}) namespace.VSphereFoundationConfig {
	result := namespace.VSphereFoundationConfig{}

	if deploymentTarget := foundationData["deployment_target"].([]interface{}); len(deploymentTarget) > 0 {
		deploymentTargetData := deploymentTarget[0].(map[string]interface{})
		result.DeploymentTarget = &namespace.DeploymentTarget{}

		if zones := deploymentTargetData["zones"].([]interface{}); len(zones) > 0 {
			value := structure.SliceInterfacesToStrings(zones)
			result.DeploymentTarget.Zones = &value
		}

		if availability := deploymentTargetData["availability"].(string); availability != "" {
			result.DeploymentTarget.Availability = &availability
		}

		if deploymentSize := deploymentTargetData["deployment_size"].(string); deploymentSize != "" {
			result.DeploymentTarget.DeploymentSize = &deploymentSize
		}

		if storagePolicy := deploymentTargetData["storage_policy"].(string); storagePolicy != "" {
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

	if networkServices := foundationData["network_services"].([]interface{}); len(networkServices) > 0 {
		value := buildEdgeNetworkServices(networkServices[0].(map[string]interface{}))
		result.NetworkServices = &value
	}

	return result
}

func buildEdgeNetworkServices(networkServicesData map[string]interface{}) namespace.EdgeNetworkServices {
	result := namespace.EdgeNetworkServices{}

	if dns := networkServicesData["dns"].([]interface{}); len(dns) > 0 {
		dnsData := dns[0].(map[string]interface{})
		result.DNS = &namespace.DNS{
			Servers:       dnsData["servers"].([]string),
			SearchDomains: dnsData["search_domains"].([]string),
		}
	}

	if ntp := networkServicesData["ntp"].([]interface{}); len(ntp) > 0 {
		ntpData := ntp[0].(map[string]interface{})
		result.NTP = &namespace.NTP{
			Servers: ntpData["servers"].([]string),
		}
	}

	if syslog := networkServicesData["syslog"].([]interface{}); len(syslog) > 0 {
		syslogData := syslog[0].(map[string]interface{})
		result.Syslog = &namespace.Syslog{}

		if endpoint := syslogData["endpoint"].(string); endpoint != "" {
			result.Syslog.Endpoint = &endpoint
		}

		if certAuthPem := syslogData["cert_authority_pem"].(string); certAuthPem != "" {
			result.Syslog.CertificateAuthorityPEM = &certAuthPem
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

	if dvpgNetwork := networkData["dvpg_network"].([]interface{}); len(dvpgNetwork) > 0 {
		value := buildDvpgNetwork(dvpgNetwork[0].(map[string]interface{}))
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

	if ipConfig := dvpgNetwork["ip_config"].([]interface{}); len(ipConfig) > 0 {
		ipConfigData := ipConfig[0].(map[string]interface{})
		result.IPConfig = &namespace.IPConfig{
			Gateway: ipConfigData["gateway"].(string),
		}

		if ipRanges := ipConfigData["ip_ranges"].([]interface{}); len(ipRanges) > 0 {
			result.IPConfig.IPRanges = make([]namespace.IPRange, len(ipRanges))

			for i, ipRange := range ipRanges {
				data := ipRange.(map[string]interface{})
				result.IPConfig.IPRanges[i] = namespace.IPRange{
					Address: data["address"].(string),
					Count:   data["count"].(int),
				}
			}
		}
	}

	return result
}

func buildControlPlaneNetwork(controlPlaneData map[string]interface{}) namespace.ControlPlaneNetwork {
	backingProperty := controlPlaneData["backing"].([]interface{})
	backingData := backingProperty[0].(map[string]interface{})
	result := namespace.ControlPlaneNetwork{
		Backing: buildControlPlaneNetworkBacking(backingData),
	}

	if network := controlPlaneData["network"].(string); network != "" {
		result.Network = &network
	}

	if floatingIp := controlPlaneData["floating_ip"].(string); floatingIp != "" {
		result.FloatingIPAddress = &floatingIp
	}

	if services := controlPlaneData["services"].([]interface{}); len(services) > 0 {
		result.Services = buildControlPlaneNetworkServices(services[0].(map[string]interface{}))
	}

	if ipManagement := controlPlaneData["ip_management"].([]interface{}); len(ipManagement) > 0 {
		result.IPManagement = buildControlPlaneNetworkIPManagement(ipManagement[0].(map[string]interface{}))
	}

	if proxy := controlPlaneData["proxy"].([]interface{}); len(proxy) > 0 {
		proxyData := proxy[0].(map[string]interface{})
		result.Proxy = &namespace.Proxy{
			ProxySettingsSource: proxyData["settings_source"].(string),
		}

		if httpConfig := proxyData["http_config"].(string); httpConfig != "" {
			result.Proxy.HTTPProxyConfig = &httpConfig
		}

		if httpsConfig := proxyData["httpс_config"].(string); httpsConfig != "" {
			result.Proxy.HTTPSProxyConfig = &httpsConfig
		}

		if tlsBundle := proxyData["tls_root_ca_bundle"].(string); tlsBundle != "" {
			result.Proxy.TLSRootCABundle = &tlsBundle
		}

		if noProxyConf := proxyData["no_proxy_config"].([]interface{}); len(noProxyConf) > 0 {
			value := structure.SliceInterfacesToStrings(noProxyConf)
			result.Proxy.NoProxyConfig = &value
		}
	}

	return result
}

func buildControlPlaneNetworkBacking(backingData map[string]interface{}) namespace.Backing {
	result := namespace.Backing{
		Backing: backingData["backing"].(string),
	}

	if network := backingData["network"].(string); network != "" {
		result.Network = &network
	}

	if segments := backingData["segments"].([]interface{}); len(segments) > 0 {
		result.NetworkSegment = &namespace.NetworkSegment{
			Networks: structure.SliceInterfacesToStrings(segments),
		}
	}

	return result
}

func buildControlPlaneNetworkServices(servicesData map[string]interface{}) *namespace.Services {
	result := namespace.Services{}

	if dns := servicesData["dns"].([]interface{}); len(dns) > 0 {
		dnsData := dns[0].(map[string]interface{})
		result.DNS = &namespace.DNS{
			Servers:       dnsData["servers"].([]string),
			SearchDomains: dnsData["search_domains"].([]string),
		}
	}

	if ntp := servicesData["ntp"].([]interface{}); len(ntp) > 0 {
		ntpData := ntp[0].(map[string]interface{})
		result.NTP = &namespace.NTP{
			Servers: ntpData["servers"].([]string),
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
			value[i] = buildIpAssignment(ipAssignmentData)
		}

		result.IPAssignments = &value
	}

	return &result
}

func buildIpAssignment(ipManagementData map[string]interface{}) namespace.IPAssignment {
	rangesData := ipManagementData["ranges"].([]interface{})
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

func buildKubeAPIServerOptions(kubeApiServerData map[string]interface{}) namespace.KubeAPIServerOptions {
	result := namespace.KubeAPIServerOptions{}

	if security := kubeApiServerData["security"].([]interface{}); len(security) > 0 {
		securityData := security[0].(map[string]interface{})
		result.Security = &namespace.KubeAPIServerSecurity{
			CertificateDNSNames: structure.SliceInterfacesToStrings(securityData["certificate_dns_names"].([]interface{})),
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

	if cloudNativeFileVolumes := storageData["cloud_native_file_volume"].([]interface{}); len(cloudNativeFileVolumes) > 0 {
		data := cloudNativeFileVolumes[0].(map[string]interface{})
		result.CloudNativeFileVolume = &namespace.CloudNativeFileVolume{
			VSANClusters: structure.SliceInterfacesToStrings(data["vsan_clusters"].([]interface{})),
		}
	}

	return result
}
