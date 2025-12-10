// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/supervisor"
)

func resourceVsphereSupervisorV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVsphereSupervisorV2Create,
		ReadContext:   resourceVsphereSupervisorV2Read,
		DeleteContext: resourceVsphereSupervisorV2Delete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The name of the Supervisor cluster.",
				ValidateFunc: validation.StringIsNotEmpty,
				ForceNew:     true,
			},
			"cluster": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "The name of the compute cluster to enable the Supervisor on. Use this property if you want to create a single zone deployment.",
				ConflictsWith: []string{"zones"},
				ValidateFunc:  validation.StringIsNotEmpty,
				ForceNew:      true,
			},
			"zones": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "A list of vSphere Zones to enable the Supervisor on.",
				ConflictsWith: []string{"cluster"},
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				ForceNew: true,
			},
			"control_plane": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The configuration for the control plane VM(s).",
				MaxItems:    1,
				Elem:        controlPlaneSchema(),
				ForceNew:    true,
			},
			"workloads": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The configuration for the Supervisor workloads.",
				MaxItems:    1,
				Elem:        workloadsSchema(),
				ForceNew:    true,
			},
		},
	}
}

func controlPlaneSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"count": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: `The number of control plane VMs to deploy.`,
			},
			"size": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The size preset for the control plane VM(s).",
				ValidateFunc: validation.StringInSlice([]string{"TINY", "SMALL", "MEDIUM", "LARGE"}, false),
			},
			"storage_policy": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The storage policy for the control plane VM(s).",
			},
			"network": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The network configuration for the control plane VM(s).",
				Elem:        controlPlaneNetworkSchema(),
			},
		},
	}
}

func controlPlaneNetworkSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The network identifier for the management network.",
			},
			"floating_ip": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Floating IP address.",
			},
			"backing": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Backing network configuration.",
				Elem:        controlPlaneNetworkBackingSchema(),
			},
			"services": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Network services (e.g DNS, NTP) configuration.",
				Elem:        networkServicesSchema(),
			},
			"ip_management": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "IP Management configuration.",
				Elem:        ipManagementSchema(),
			},
			"proxy": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Proxy server configuration.",
				Elem:        controlPlaneNetworkProxySchema(),
			},
		},
	}
}

func controlPlaneNetworkBackingSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The Managed Object ID of the Network object.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"segments": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The backing network segment.",
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The DNS configuration.",
				Elem:        dnsSchema(),
			},
			"ntp": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The NTP configuration.",
				Elem:        ntpSchema(),
			},
		},
	}
}

func ipManagementSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"dhcp_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether to use DHCP or not.",
			},
			"gateway_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The IP address of the network gateway.",
				ValidateFunc: validation.IsCIDR,
			},
			"ip_assignment": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "IP assignment configuration.",
				Elem:        ipAssignmentSchema(),
			},
		},
	}
}

func ipAssignmentSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"assignee": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The type of the assignee.",
				ValidateFunc: validation.StringInSlice([]string{"POD", "NODE", "SERVICE"}, false),
			},
			"range": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The available IP addresses that can be consumed by Supervisor to run the cluster.",
				Elem:        ipRangeSchema(),
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
				Description:  "The starting IP address of the range.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"count": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The number of IP addresses in the range.",
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
				Description:  "The source of the proxy settings.",
				ValidateFunc: validation.StringInSlice([]string{"VC_INHERITED", "CLUSTER_CONFIGURED", "NONE"}, false),
			},
			"http_config": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "HTTP proxy configuration. This can be used if `CLUSTER_CONFIGURED` is specified for `settings_source`.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"https_config": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "HTTPS proxy configuration. This can be used if `CLUSTER_CONFIGURED` is specified for `settings_source`",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"tls_root_ca_bundle": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Proxy TLS root CA bundle which will be used to verify the proxy's certificates. Every certificate in the bundle is expected to be in PEM format. This can be used if `CLUSTER_CONFIGURED` is specified for `settings_source`",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"no_proxy_config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of addresses that should be accessed directly. This can be used if `CLUSTER_CONFIGURED` is specified for `settings_source`",
				Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
			},
		},
	}
}

func workloadsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"network": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The primary workload network configuration. Workloads will communicate with each other and will reach external networks over this network.",
				Elem:        workloadNetworkSchema(),
			},
			"edge": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Edge configuration",
				Elem:        edgeSchema(),
			},
			"kube_api_server_options": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Kubernetes API Server options",
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for storing and pulling images into the cluster.",
				Elem:        imagesSchema(),
			},
			"storage": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Persistent storage configuration.",
				Elem:        storageSchema(),
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
				Description:  "A unique identifier for the workload network.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"vsphere": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for vSphere network backing.",
				Elem:        workloadNetworkVSphereSchema(),
			},
			"nsx": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for NSX-T backing.",
				Elem:        workloadNetworkNSXSchema(),
			},
			"nsx_vpc": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for NSX VPC backing.",
				Elem:        workloadNetworkNSXVPCSchema(),
			},
			"services": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Network services (e.g DNS, NTP) configuration.",
				Elem:        networkServicesSchema(),
			},
			"ip_management": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "IP Management configuration.",
				Elem:        ipManagementSchema(),
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
				Description:  "The identifier of the Distributed Virtual Portgroup.",
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
				Description:  "The identifier of the vSphere Distributed Switch.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"namespace_subnet_prefix": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The size of the subnet reserved for namespace segments.",
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
				Description:  "The NSX Project for VPCs in the Supervisor, including the System VPC, and Supervisor Services VPC.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"vpc_connectivity_profile": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The identifier of the VPC Connectivity Profile.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"default_private_cidr": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Specifies CIDR blocks from which private subnets are allocated.",
				Elem:        ipv4CidrSchema(),
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
				Description:  "The starting IPv4 address of the CIDR block.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"prefix": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The number of addresses in the CIDR block.",
			},
		},
	}
}

func ntpSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"servers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The list of NTP servers.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringIsNotEmpty,
				},
			},
		},
	}
}

func dnsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"servers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The list of DNS servers.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.IsIPv4Address,
				},
			},
			"search_domains": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The list of search domains.",
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
				Description:  "The unique identifier of this edge.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"lb_address_range": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of addresses that a load balancer can consume to publish Kubernetes services.",
				Elem:        ipRangeSchema(),
			},
			"foundation": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Configuration for the vSphere Foundation Load Balancer.",
				MaxItems:    1,
				Elem:        edgeFoundationSchema(),
			},
			"haproxy": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for the HAProxy Load Balancer.",
				Elem:        haProxySchema(),
			},
			"nsx": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for the NSX Load Balancer.",
				Elem:        nsxSchema(),
			},
			"nsx_advanced": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for the NSX Advanced Load Balancer.",
				Elem:        nsxAdvancedSchema(),
			},
		},
	}
}

func edgeFoundationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"deployment_target": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "The configuration for the Load Balancer placement.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Configures the availability level for the load balancer.",
							ValidateFunc: validation.StringInSlice([]string{"ACTIVE_PASSIVE", "SINGLE_NODE"}, false),
						},
						"zones": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "The list of zones to deploy onto.",
							Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
						},
						"deployment_size": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Determines the CPU/memory resource size of the load balancer deployment.",
							ValidateFunc: validation.StringInSlice([]string{"SMALL", "MEDIUM", "LARGE", "X_LARGE"}, false),
						},
						"storage_policy": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "Storage Policy containing datastores hosting the load balancer nodes.",
							ValidateFunc: validation.StringIsNotEmpty,
						},
					},
				},
			},
			"interface": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Configuration for the Load Balancer network interfaces.",
				Elem:        edgeFoundationNetworkInterfaceSchema(),
			},
			"network_services": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for the Load Balancer network services.",
				Elem:        edgeNetworkServicesSchema(),
			},
		},
	}
}

func edgeFoundationNetworkInterfaceSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"personas": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Determines the type of traffic that passes through a network interface.",
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{"MANAGEMENT", "WORKLOAD", "FRONTEND"}, false),
				},
			},
			"network": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Network configuration for this interface.",
				Elem:        edgeFoundationNetworkInterfaceNetworkSchema(),
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
				Description:  "The type of network interface.",
				ValidateFunc: validation.StringInSlice([]string{"SUPERVISOR_MANAGEMENT", "PRIMARY_WORKLOAD", "DVPG"}, false),
			},
			"dvpg_network": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Identifier of the Distributed Virtual Portgroup.",
				Elem:        edgeFoundationNetworkInterfaceNetworkDvpgSchema(),
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
				Description:  "The display name of the Supervisor workload network.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"network": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The identifier of the Distributed Virtual Portgroup.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ipam": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "IP Address management scheme for this network.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ip_config": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Static IP Configuration for this network.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_range": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "IP range configuration.",
							Elem:        ipRangeSchema(),
						},
						"gateway": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Gateway address.",
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
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "DNS configuration.",
				Elem:        dnsSchema(),
			},
			"ntp": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "NTP configuration.",
				Elem:        ntpSchema(),
			},
			"syslog": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Remote log forwarding configuration.",
				Elem:        syslogSchema(),
			},
		},
	}
}

func haProxySchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The address for the data plane API server.",
				Elem:        edgeServerSchema(),
			},
			"username": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The username for the data plane API server.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The password for the data plane API server.",
			},
			"ca_chain": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The CA chain for the data plane API server.",
			},
		},
	}
}

func nsxSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"edge_cluster": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The identifier of the edge cluster.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"load_balancer_size": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The size of the load balancer node.",
				ValidateFunc: validation.StringInSlice([]string{"SMALL", "MEDIUM", "LARGE"}, false),
			},
			"t0_gateway": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Tier-0 gateway ID for the namespaces configuration.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"routing_mode": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Routing mode.",
				ValidateFunc: validation.StringInSlice([]string{"ROUTED", "NAT"}, false),
			},
			"default_ingress_tls_certificate": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The default certificate that is served on Ingress services, when another certificate is not presented.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"egress_ip_range": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "An IP Range from which NSX assigns IP addresses used for performing SNAT from container IPs to external IPs.",
				Elem:        ipRangeSchema(),
			},
		},
	}
}

func nsxAdvancedSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"server": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The address for the AVI controller.",
				Elem:        edgeServerSchema(),
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Username",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Password",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ca_chain": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Certificate authority chain.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cloud_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Cloud Name.",
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
				Description:  "The IP address of the host.",
				ValidateFunc: validation.IsIPv4Address,
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The port of the host.",
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
				Description:  "FQDN or IP address of the remote syslog server taking the form protocol://hostname|ipv4|ipv6[:port]. The syslog protocol defaults to tcp.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cert_authority_pem": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Certificate authority PEM.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func imagesSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"registry": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Configuration for the container image registry endpoint.",
				Elem:        registrySchema(),
			},
			"repository": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The default container image repository to use when the Kubernetes Pod configuration does not specify it.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"kubernetes_content_library": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The identifier of the Content Library which holds the VM Images for vSphere Kubernetes Service.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"content_library": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Content library associated with the Supervisor.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_library": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Content library identifier.",
							ValidateFunc: validation.StringIsNotEmpty,
						},
						"supervisor_services": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "A list of Supervisor Service IDs that are currently making use of the Content Library.",
							Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
						},
						"resource_naming_strategy": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "The resource naming strategy that is used to generate the Kubernetes resource names for images from this Content Library.",
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
				Description:  "The IP address of the image registry.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"port": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The port of the image registry.",
			},
			"username": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The username of the image registry.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"password": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The password of the image registry.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"ca_chain": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The certificate authority chain of the image registry.",
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
				Description:  "The storage policy associated with ephemeral disks of all the Kubernetes Pod VMs in the cluster.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"image_storage_policy": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The specification required to configure storage used for Pod VM container images.",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"cloud_native_file_volume": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Specifies the Cloud Native Storage file volume.",
				MaxItems:    1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vsan_clusters": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "A list of cluster identifiers.",
							Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringIsNotEmpty},
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
		id, err := supervisor.EnableSupervisorMultiZone(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(id)
	} else if cluster != "" {
		id, err := supervisor.EnableSupervisorSingleZone(ctx, d, m)
		if err != nil {
			return diag.FromErr(err)
		}

		d.SetId(id)
	} else {
		return diag.Errorf("either 'zones' or 'cluster' must be specified")
	}

	return supervisor.WaitForSupervisorEnable(ctx, d, m)
}

func resourceVsphereSupervisorV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	_, err := m.GetSupervisorSummary(ctx, d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVsphereSupervisorV2Delete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(*Client).restClient
	m := namespace.NewManager(c)

	topology, err := m.GetSupervisorTopology(ctx, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	id := getClusterID(ctx, topology)
	if id == "" {
		return diag.FromErr(fmt.Errorf("no clusters found for supervisor %s", d.Id()))
	}

	if err := m.DisableCluster(ctx, id); err != nil {
		return diag.FromErr(err)
	}

	return supervisor.WaitForSupervisorDisable(ctx, d, m)
}

func getClusterID(ctx context.Context, topology []namespace.SupervisorTopologyInfo) string {
	if len(topology) == 0 {
		tflog.Debug(ctx, "supervisor topology is empty")
	}

	for _, t := range topology {
		for _, cluster := range t.Clusters {
			return cluster
		}
	}

	return ""
}
