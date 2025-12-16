// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceVSphereSupervisorV2_singleZone(t *testing.T) {
	// Cannot run on the basic test setup
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCheckEnvVariables(t, []string{
				"TF_VAR_VSPHERE_CLUSTER",
				"TF_VAR_STORAGE_POLICY",
				"TF_VAR_FLOATING_IP",
				"TF_VAR_CONTROL_PLANE_NETWORK",
				"TF_VAR_WORKLOAD_NETWORK",
				"TF_VAR_NTP_SERVER",
				"TF_VAR_EDGE_OVERLAY_NETWORK",
				"TF_VAR_EDGE_MANAGEMENT_NETWORK",
			})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereSupervisorV2ConfigSingleZone(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_supervisor_v2.supervisor", "id"),
				),
			},
		},
	})
}

func TestAccResourceVSphereSupervisorV2_multiZone(t *testing.T) {
	// Cannot run on the basic test setup
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCheckEnvVariables(t, []string{
				"TF_VAR_STORAGE_POLICY",
				"TF_VAR_FLOATING_IP",
				"TF_VAR_CONTROL_PLANE_NETWORK",
				"TF_VAR_WORKLOAD_NETWORK",
				"TF_VAR_NTP_SERVER",
				"TF_VAR_EDGE_OVERLAY_NETWORK",
				"TF_VAR_EDGE_MANAGEMENT_NETWORK",
				"TF_VAR_HAPROXY_HOST",
				"TF_VAR_HAPROXY_PORT",
				"TF_VAR_HAPROXY_USER",
				"TF_VAR_HAPROXY_PASS",
				"TF_VAR_HAPROXY_CA_CHAIN",
			})
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereSupervisorV2ConfigMultiZone(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vsphere_supervisor_v2.supervisor", "id"),
				),
			},
		},
	})
}

func testAccVSphereSupervisorV2ConfigSingleZone() string {
	return fmt.Sprintf(`

variable "control_plane_network" {
  default = "%s"
}

variable "ntp_server" {
  default = "%s"
}

variable "edge_overlay_network" {
  default = "%s"
}

variable "edge_management_network" {
  default = "%s"
}

resource "vsphere_supervisor_v2" "supervisor" {
  cluster         = "%s"
  name            = "supervisor"

  control_plane {
    size           = "SMALL"
    count          = 1
    storage_policy = "%s"

    network {
      network     = var.control_plane_network
      floating_ip = "%s"

      backing {
        network = var.control_plane_network
      }

      services {
        ntp {
          servers = [var.ntp_server]
        }
      }
    }
  }

  workloads {
    network {
      network = "primary"

      vsphere {
        dvpg = "%s"
      }

      services {
        ntp {
          servers = [var.ntp_server]
        }
        dns {
          servers        = ["192.19.189.10"]
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
      id = "flb-1"

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
              network = var.edge_overlay_network
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
              network = var.edge_management_network
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
`, os.Getenv("TF_VAR_CONTROL_PLANE_NETWORK"),
		os.Getenv("TF_VAR_NTP_SERVER"),
		os.Getenv("TF_VAR_EDGE_OVERLAY_NETWORK"),
		os.Getenv("TF_VAR_EDGE_MANAGEMENT_NETWORK"),
		os.Getenv("TF_VAR_VSPHERE_CLUSTER"),
		os.Getenv("TF_VAR_STORAGE_POLICY"),
		os.Getenv("TF_VAR_FLOATING_IP"),
		os.Getenv("TF_VAR_WORKLOAD_NETWORK"))
}

func testAccVSphereSupervisorV2ConfigMultiZone() string {
	return fmt.Sprintf(`

variable "control_plane_network" {
  default = "%s"
}

variable "ntp_server" {
  default = "%s"
}

resource "vsphere_supervisor_v2" "supervisor" {
  zones           = ["zone-1", "zone-2", "zone-3"]
  name            = "supervisor"

  control_plane {
    size           = "SMALL"
    count          = 3
    storage_policy = "%s"

    network {
      network     = var.control_plane_network
      floating_ip = "%s"

      backing {
        network = var.control_plane_network
      }

      services {
        ntp {
          servers = [var.ntp_server]
        }
      }
    }
  }

  workloads {
    network {
      network = "primary"

      vsphere {
        dvpg = "%s"
      }

      services {
        ntp {
          servers = [var.ntp_server]
        }
        dns {
          servers        = ["192.19.189.10"]
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

      lb_address_range {
        address = "192.168.130.0"
        count   = 5120
      }

      haproxy {
        server {
          host = "%s"
          port = %s
        }
        
        username = "%s"
        password = "%s"
        ca_chain = "%s"
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
`, os.Getenv("TF_VAR_CONTROL_PLANE_NETWORK"),
		os.Getenv("TF_VAR_NTP_SERVER"),
		os.Getenv("TF_VAR_STORAGE_POLICY"),
		os.Getenv("TF_VAR_FLOATING_IP"),
		os.Getenv("TF_VAR_WORKLOAD_NETWORK"),
		os.Getenv("TF_VAR_HAPROXY_HOST"),
		os.Getenv("TF_VAR_HAPROXY_PORT"),
		os.Getenv("TF_VAR_HAPROXY_USER"),
		os.Getenv("TF_VAR_HAPROXY_PASS"),
		os.Getenv("TF_VAR_HAPROXY_CA_CHAIN"))
}
