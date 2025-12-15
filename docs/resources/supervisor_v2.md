---
subcategory: "Workload Management"
page_title: "VMware vSphere: vsphere_supervisor_v2"
sidebar_current: "docs-vsphere-resource-vsphere-supervisor-v2"
description: |-
  Provides a vSphere Supervisor resource.
---

# vsphere_supervisor_v2

Provides a resource for configuring vSphere Supervisor.

~> **NOTE:** Some attributes are only available in vSphere 9, consult the product documentation if you want to use this with vSphere 8.

~> **NOTE:** Update and Import operations are not supported yet and will be added in a future release.

To configure a single-zone Supervisor you must set the `cluster` attribute.
Its value should be the Managed Object identifier of the compute cluster you wish to deploy on.
This identifier corresponds to the `id` attribute of `d/compute_cluster` and `r/compute_cluster`.

```hcl
data "vsphere_datacenter" "datacenter" {
  name = "dc-01"
}

data "vsphere_compute_cluster" "compute_cluster" {
  name          = "cluster-01"
  datacenter_id = data.vsphere_datacenter.datacenter.id
}

resource "vsphere_supervisor_v2" "supervisor" {
  cluster = data.vsphere_compute_cluster.compute_cluster.id
  name    = "supervisor"
  # ...
}
```

To configure a multi-zone Supervisor you must set the `zones` attribute.
A standard stretched Supervisor is deployed on 3 vSphere zones and their identifiers can be obtained from
`d/vsphere_zone` or `r/vsphere_zone`.

```hcl
data "vsphere_zone" "zone1" {
  name = "zone-1"
}

data "vsphere_zone" "zone2" {
  name = "zone-2"
}

data "vsphere_zone" "zone3" {
  name = "zone-3"
}

resource "vsphere_supervisor_v2" "supervisor" {
  zones = [
    data.vsphere_zone.zone1.id,
    data.vsphere_zone.zone2.id,
    data.vsphere_zone.zone3.id
  ]
  name = "supervisor"
  # ...
}
```

The two deployment modes are not interchangeable - you cannot stretch a single-zone deployment into a three-zone one or shrink
a three-zone Supervisor into a single-zone. The `cluster` and `zones` attribute are marked as conflicting and
the provider will not allow you to specify both at the same time.
Apart from these two attributes the rest of the schema for this resource is identical for both modes.

The resource requires you to provide control plane and workload configurations as mandatory nested blocks
but most of their inner attributes are optional and can be specified based on the requirements of your
deployment.

The schema closely follows the structure of the API payload with some simplifications where possible (_e.g._, you do not need to 
provide the network type of your workload network, it is assumed based on the network type of the configuration provided.)

## Example Usages

### Enable Supervisor on a single Compute Cluster

```hcl

resource "vsphere_supervisor_v2" "supervisor" {
  cluster = "domain-c52"
  name    = "supervisor"

  control_plane {
    size           = "SMALL"
    count          = 1
    storage_policy = "a07f430a-b98a-4389-836d-d301e87d1531"

    network {
      network     = "network-47"
      floating_ip = "10.11.12.13"

      backing {
        network = "network-47"
      }

      services {
        ntp {
          servers = ["ntp1.mycompany.local"]
        }
      }
    }
  }

  workloads {
    network {
      network      = "primary"
      network_type = "VSPHERE"

      vsphere {
        dvpg = "dvportgroup-66"
      }

      services {
        ntp {
          servers = ["ntp1.example.com"]
        }
        dns {
          servers = ["192.19.189.10"]
          search_domains = [
            "domain-1.test",
            "wcp.integration.test",
            "xn--80akhbyknj4f",
          ]
        }
      }

      ip_management {
        dhcp_enabled    = false
        gateway_address = "192.168.1.1/16"

        ip_assignment {
          assignee = "SERVICE"
          range {
            address = "172.24.0.0"
            count   = 65536
          }
        }

        ip_assignment {
          assignee = "NODE"
          range {
            address = "192.168.128.0"
            count   = 256
          }
        }
      }
    }

    edge {
      id       = "flb-1"
      provider = "VSPHERE_FOUNDATION"

      lb_address_range {
        address = "172.16.0.200"
        count   = 54
      }

      foundation {
        deployment_target {
          availability = "SINGLE_NODE"
        }

        interface {
          personas = ["FRONTEND"]
          network {
            network_type = "DVPG"
            dvpg_network {
              name    = "network-1"
              network = "dvportgroup-62"
              ipam    = "STATIC"

              ip_config {
                gateway = "172.16.0.1/16"
                ip_range {
                  address = "172.16.0.2"
                  count   = 196
                }
              }
            }
          }
        }

        interface {
          personas = ["MANAGEMENT"]
          network {
            network_type = "DVPG"
            dvpg_network {
              name    = "flb-mgmt"
              network = "dvportgroup-64"
              ipam    = "STATIC"

              ip_config {
                gateway = "172.25.0.1/16"
                ip_range {
                  address = "172.25.0.2"
                  count   = 196
                }
              }
            }
          }
        }

        interface {
          personas = ["WORKLOAD"]
          network {
            network_type = "PRIMARY_WORKLOAD"
          }
        }
      }
    }

    kube_api_server_options {
      security {
        certificate_dns_names = [
          "domain-1.test",
          "wcp.integration.test",
          "xn--80akhbyknj4f",
        ]
      }
    }
  }
}
```

### Enable Supervisor on 3 vSphere Zones

```hcl
resource "vsphere_supervisor_v2" "supervisor" {
  zones = ["zone-1", "zone-2", "zone-3"]
  name  = "supervisor"

  control_plane {
    size           = "SMALL"
    count          = 3
    storage_policy = "a07f430a-b98a-4389-836d-d301e87d1531"

    network {
      network     = "network-47"
      floating_ip = "10.11.12.13"

      backing {
        network = "network-47"
      }

      services {
        ntp {
          servers = ["ntp1.mycompany.local"]
        }
      }
    }
  }

  workloads {
    network {
      network      = "primary"
      network_type = "VSPHERE"

      vsphere {
        dvpg = "%s"
      }

      services {
        ntp {
          servers = ["ntp1.example.com"]
        }
        dns {
          servers = ["192.19.189.10"]
          search_domains = [
            "domain-1.test",
            "wcp.integration.test",
            "xn--80akhbyknj4f",
          ]
        }
      }

      ip_management {
        dhcp_enabled    = false
        gateway_address = "192.168.1.1/16"

        ip_assignment {
          assignee = "SERVICE"
          range {
            address = "172.24.0.0"
            count   = 65536
          }
        }

        ip_assignment {
          assignee = "NODE"
          range {
            address = "192.168.128.0"
            count   = 256
          }
        }
      }
    }

    edge {
      id       = "haproxy"
      provider = "HAPROXY"

      lb_address_range {
        address = "192.168.130.0"
        count   = 5120
      }

      haproxy {
        server {
          host = "192.168.100.110"
          port = 5556
        }
      }
    }

    kube_api_server_options {
      security {
        certificate_dns_names = [
          "domain-1.test",
          "wcp.integration.test",
          "xn--80akhbyknj4f",
        ]
      }
    }
  }
}
```

## Argument Reference
The following arguments are supported:

* `name` - (Required) The name of the Supervisor cluster.
* `control_plane` - (Required) The configuration for the control plane VM(s). See [control_plane](#nestedblock--control-plane).
* `workloads` - (Required) The configuration for the Supervisor workloads. See [workloads](#nestedblock--workloads).
* `cluster` - (Optional) The name of the compute cluster to enable the Supervisor on. Use this property if you want to create a single zone deployment. Conflicts with `zones`.
* `zones` - (Optional) A list of vSphere Zones to enable the Supervisor on. Conflicts with `cluster`.

<a id="nestedblock--control-plane"></a>
### Nested schema for `control_plane`
The `control_plane` block configures the management layer of the Supervisor.

* `network` - (Required) The network configuration for the control plane VM(s). See [control_plane.network](#nestedblock--control-plane-network).
* `count` - (Optional) The number of control plane VMs to deploy. If specified, this should be greater or equal to the number of zones backing the Supervisor - 1 for single zone deployments and 3 for stretched Supervisors.
* `size` - (Optional) The size preset for the control plane VM(s). Allowed values are `TINY`, `SMALL`, `MEDIUM`, and `LARGE`.
* `storage_policy` - (Optional) The storage policy for the control plane VM(s).

<a id="nestedblock--control-plane-network"></a>
### Nested schema for `control_plane.network`

* `backing` - (Required) Backing network configuration. See [backing](#nestedblock--backing).
* `network` - (Optional) The network identifier for the management network.
* `floating_ip` - (Optional) Floating IP address.
* `services` - (Optional) Network services (_e.g._, DNS, NTP) configuration.
* * `dns` - (Optional) The DNS configuration.
* * * `servers` - (Required) The list of DNS servers.
* * * `search_domains` - (Required) The list of search domains.
* * `ntp` - (Optional) The NTP configuration.
* * * `servers` - (Required) The list of NTP servers.
* `ip_management` - (Optional) IP Management configuration. See [ip_management](#nestedblock--ip-management).
* `proxy` - (Optional) Proxy server configuration. See [proxy](#nestedblock--proxy).

<a id="nestedblock--workloads"></a>
### Nested schema for `workloads`

The workloads block configures the workload network, storage, and image registry settings.

* `network` - (Required) The primary workload network configuration. Workloads will communicate with each other and will reach external networks over this network. See [workloads.network](#nestedblock--workloads-network).
* `edge` - (Required) Edge configuration. See [workloads.edge](#nestedblock--workloads-edge).
* `kube_api_server_options` - (Required) Kubernetes API Server options. See [workloads.kube_api_server_options](#nestedblock--workloads-kube-api-server).
* `images` - (Optional) Configuration for storing and pulling images into the cluster. See [workloads.images](#nestedblock--workloads-images).
* `storage` - (Optional) Persistent storage configuration. See [workloads.storage](#nestedblock--workloads-storage).

<a id="nestedblock--workloads-network"></a>
### Nested schema for `workloads.network`

* `dvs` - (Required) The identifier of the vSphere Distributed Switch.
* `dvpg` - (Required) The identifier of the Distributed Virtual Portgroup.
* `default_private_cidr` - (Required) Specifies CIDR blocks from which private subnets are allocated. See [cidr](#nestedblock--cidr).
* `network` - (Optional) A unique identifier for the workload network.
* `services` - (Optional) Network services (_e.g._, DNS, NTP) configuration.
* * `dns` - (Optional) The DNS configuration.
* * * `servers` - (Required) The list of DNS servers.
* * * `search_domains` - (Required) The list of search domains.
* * `ntp` - (Optional) The NTP configuration.
* * * `servers` - (Required) The list of NTP servers.
* `ip_management` - (Optional) IP Management configuration. See [ip-management](#nestedblock--ip-management).
* `vsphere` - (Optional) Configuration for vSphere network backing. Conflicts with `nsx` and `nsx_vpc`.
* `nsx` - (Optional) Configuration for NSX backing. Conflicts with `vsphere` and `nsx_vpc`.
* `namespace_subnet_prefix` - (Optional) The size of the subnet reserved for namespace segments.
* `nsx_vpc` - (Optional) Configuration for NSX VPC backing. Conflicts with `vsphere` and `nsx`.
* `nsx_project` - (Optional) The NSX Project for VPCs in the Supervisor, including the System VPC, and Supervisor Services VPC.
* `vpc_connectivity_profile` - (Optional) The identifier of the VPC Connectivity Profile.

<a id="nestedblock--workloads-edge"></a>
### Nested schema for `workloads.edge`

The edge block configures the load balancer settings.

* `id` - (Optional) The unique identifier of this edge.
* `lb_address_range` - (Optional) The list of addresses that a load balancer can consume to publish Kubernetes services. See [ip_range](#nestedblock--ip-range).
* `foundation` - (Optional) Configuration for the vSphere Foundation Load Balancer. Conflicts with `haproxy`, `nsx`, `nsx_advanced`. See [workloads.edge.foundation](#nestedblock--workloads-edge-foundation).
* `haproxy` - (Optional) Configuration for the HAProxy Load Balancer. Conflicts with `foundation`, `nsx`, `nsx_advanced`. See [workloads.edge.haproxy](#nestedblock--workloads-edge-haproxy).
* `nsx` - (Optional) Configuration for the NSX Load Balancer. Conflicts with `haproxy`, `foundation`, `nsx_advanced`. See [workloads.edge.nsx](#nestedblock--workloads-edge-nsx).
* `nsx_advanced` - (Optional) Configuration for the NSX Advanced Load Balancer. Conflicts with haproxy, nsx, foundation. See [workloads.edge.nsxadvanced](#nestedblock--workloads-edge-nsxadvanced).

<a id="nestedblock--workloads-edge-foundation"></a>
### Nested schema for `workloads.edge.foundation`

* `deployment_target` - (Optional) The configuration for the Load Balancer placement. Includes `availability`, `zones`, `deployment_size`, and `storage_policy`.
* `interface` - (Optional) Configuration for the Load Balancer network interfaces. Includes personas and network.
* `network_services` - (Optional) Configuration for the Load Balancer network services.
* * `dns` - (Optional) The DNS configuration.
* * * `servers` - (Required) The list of DNS servers.
* * * `search_domains` - (Required) The list of search domains.
* * `ntp` - (Optional) The NTP configuration.
* * * `servers` - (Required) The list of NTP servers.
* * `syslog` - (Optional) Remote log forwarding configuration.
* * * `endpoint` - (Optional) FQDN or IP address of the remote syslog server.
* * * `ca_cert` - (Optional) Certificate authority certificate in PEM format.

<a id="nestedblock--workloads-edge-haproxy"></a>
### Nested schema for `workloads.edge.haproxy`

* `server` - (Required) The address for the data plane API server.
* * `host` - (Required) The IP address of the host.
* * `port` - (Required) The port of the host.
* `username` - (Required) Username.
* `password` - (Required) Password.
* `ca_chain` - (Required) The certificate authority chain.

<a id="nestedblock--workloads-edge-nsx"></a>
### Nested schema for `workloads.edge.nsx`

* `edge_cluster` - (Optional) The identifier of the edge cluster.
* `load_balancer_size` - (Optional) The size of the load balancer node. Allowed values are `SMALL`, `MEDIUM`, `LARGE`.
* `t0_gateway` - (Optional) Tier-0 gateway ID for the namespaces configuration.
* `routing_mode` - (Optional) Routing mode. Allowed values are `ROUTED`, `NAT`.
* `default_ingress_tls_certificate` - (Optional) The default certificate that is served on Ingress services, when another certificate is not presented.
* `egress_ip_range` - (Optional) An IP Range from which NSX assigns IP addresses used for performing SNAT from container IPs to external IPs. [ip_range](#nestedblock--ip-range).

<a id="nestedblock--workloads-edge-nsxadvanced"></a>
### Nested schema for `workloads.edge.nsxadvanced`

* `host` - (Required) The IP address of the AVI controller.
* `port` - (Required) The port of the AVI controller.
* `username` - (Required) Username.
* `password` - (Required) Password.
* `ca_chain` - (Required) Certificate authority chain.
* `cloud_name` - (Optional) Cloud Name.

<a id="nestedblock--workloads-kube-api-server"></a>
### Nested schema for `workloads.kube_api_server_options`

* `certificate_dns_names` - (Required) List of DNS names to include in the certificate.
* `security` - (Optional) Security configuration.

<a id="nestedblock--workloads-images"></a>
### Nested schema for `workloads.images`

* `registry` - (Required) Configuration for the container image registry endpoint. See [workloads.images.registry](#nestedblock--workloads.images.registry).
* `repository` - (Required) The default container image repository to use when the Kubernetes Pod configuration does not specify it.
* `kubernetes_content_library` - (Required) The identifier of the Content Library which holds the VM Images for vSphere Kubernetes Service.
* `content_library` - (Required) Content library associated with the Supervisor.
* * `content_library` - (Required) Content library identifier.
* * `supervisor_services` - (Optional) A list of Supervisor Service IDs that are currently making use of the Content Library.
* * `resource_naming_strategy` - (Optional) The resource naming strategy that is used to generate the Kubernetes resource names for images from this Content Library.

<a id="nestedblock--workloads-storage"></a>
### Nested schema for `workloads.storage`

* `vsan_clusters` - (Required) A list of cluster identifiers.
* `ephemeral_storage_policy` - (Optional) The storage policy associated with ephemeral disks of all the Kubernetes Pod VMs in the cluster.
* `image_storage_policy` - (Optional) The specification required to configure storage used for Pod VM container images.
* `cloud_native_file_volume` - (Optional) Specifies the Cloud Native Storage file volume.

<a id="nestedblock--backing"></a>
### Nested schema for `backing`

* `network` - (Optional) The Managed Object ID of the Network object. Conflicts with `segments`.
* `segments` - (Optional) The backing network segment. Conflicts with network.

<a id="nestedblock--ip-management"></a>
### Nested schema for `ip_management`

* `dhcp_enabled` - (Optional) Whether to use DHCP or not.
* `gateway_address` - (Optional) The IP address of the network gateway.
* `ip_assignment` - (Optional) IP assignment configuration.
* `assignee` - (Optional) The type of the assignee. Allowed values: `POD`, `NODE`, `SERVICE`.
* `range` - (Optional) The available IP addresses that can be consumed by Supervisor to run the cluster. See [ip_range](#nestedblock--ip-range").

<a id="nestedblock--ip-range"></a>
### Nested schema for `ip_range`

* `address` - (Required) The starting IP address of the range.
* `count` - (Required) The number of IP addresses in the range.

<a id="nestedblock--cidr"></a>
### Nested schema for `cidr`

* `address` - (Required) The starting IPv4 address of the CIDR block.
* `prefix` - (Required) The number of addresses in the CIDR block.

<a id="nestedblock--proxy"></a>
### Nested schema for `proxy`

* `settings_source` - (Required) The source of the proxy settings. Allowed values are `VC_INHERITED`, `CLUSTER_CONFIGURED`, `NONE`.
* `http_config` - (Optional) HTTP proxy configuration. Can be used if CLUSTER_CONFIGURED.
* `https_config` - (Optional) HTTPS proxy configuration. Can be used if CLUSTER_CONFIGURED.
* `tls_root_ca_bundle` - (Optional) Proxy TLS root CA bundle which will be used to verify the proxy's certificates.
* `no_proxy_config` - (Optional) List of addresses that should be accessed directly.

<a id="nestedblock--workloads.images.registry"></a>
### Nested schema for `workloads.images.registry`

* `hostname` - (Required) The IP address of the image registry.
* `port` - (Required) The port of the image registry.
* `username` - (Required) The username of the image registry.
* `password` - (Required) The password of the image registry.
* `ca_chain` - (Required) The certificate authority chain of the image registry.
