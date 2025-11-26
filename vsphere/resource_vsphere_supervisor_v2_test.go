// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/govmomi/vapi/namespace"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccResourceVSphereSupervisorV2_singleZone(t *testing.T) {
	// Cannot run on the basic test setup
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
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

resource "vsphere_supervisor_v2" "supervisor" {
  cluster         = "domain-c52"
}
`)
}

func testAccVSphereSupervisorV2ConfigMultiZone() string {
	return fmt.Sprintf(`

resource "vsphere_supervisor_v2" "supervisor" {
  cluster         = "domain-c52"
}
`)
}

// Mocks

func buildMockSpecSingleZone(d *schema.ResourceData) *namespace.EnableOnComputeClusterSpec {
	return &namespace.EnableOnComputeClusterSpec{
		Name: "supervisor",
		ControlPlane: namespace.ControlPlane{
			Network: namespace.ControlPlaneNetwork{
				Network: Strptr("dvportgroup-69"),
				Backing: namespace.Backing{
					Backing: "NETWORK",
					Network: Strptr("dvportgroup-69"),
				},
				Services: &namespace.Services{
					NTP: &namespace.NTP{
						Servers: []string{"ntp1.vcfd.broadcom.net"},
					},
				},
				FloatingIPAddress: Strptr("10.192.34.217"),
			},
			Size:          Strptr("SMALL"),
			StoragePolicy: Strptr("b84c5a2c-712a-4b14-8420-0e9a7503388c"),
			Count:         Intptr(1),
		},
		Workloads: namespace.Workloads{
			Network: namespace.WorkloadNetwork{
				Network:     Strptr("primary"),
				NetworkType: "VSPHERE",
				VSphere: &namespace.NetworkVSphere{
					DVPG: "dvportgroup-56",
				},
				Services: &namespace.Services{
					DNS: &namespace.DNS{
						Servers: []string{"192.19.189.10"},
						SearchDomains: []string{
							"domain-1.test",
							"wcp.integration.test",
							"xn--80akhbyknj4f",
						},
					},
					NTP: &namespace.NTP{
						Servers: []string{"ntp1.vcfd.broadcom.net"},
					},
				},
				IPManagement: &namespace.IPManagement{
					DHCPEnabled:    types.NewBool(false),
					GatewayAddress: Strptr("192.168.1.1/16"),
					IPAssignments: &[]namespace.IPAssignment{
						{
							Assignee: Strptr("SERVICE"),
							Ranges: []namespace.IPRange{
								{
									Address: "172.24.0.0",
									Count:   65536,
								},
							},
						},
						{
							Assignee: Strptr("NODE"),
							Ranges: []namespace.IPRange{
								{
									Address: "192.168.128.0",
									Count:   256,
								},
							},
						},
					},
				},
			},
			Edge: namespace.Edge{
				ID: Strptr("flb-1"),
				LoadBalancerAddressRanges: &[]namespace.IPRange{
					{
						Address: "172.16.0.200",
						Count:   54,
					},
				},
				Foundation: &namespace.VSphereFoundationConfig{
					DeploymentTarget: &namespace.DeploymentTarget{
						Zones:        &[]string{},
						Availability: Strptr("SINGLE_NODE"),
					},
					Interfaces: &[]namespace.NetworkInterface{
						{
							Personas: []string{"FRONTEND"},
							Network: namespace.NetworkInterfaceNetwork{
								NetworkType: "DVPG",
								DVPGNetwork: &namespace.DVPGNetwork{
									Name:    "network-1",
									Network: "dvportgroup-62",
									IPAM:    "STATIC",
									IPConfig: &namespace.IPConfig{
										IPRanges: []namespace.IPRange{
											{
												Address: "172.16.0.2",
												Count:   196,
											},
										},
										Gateway: "172.16.0.1/16",
									},
								},
							},
						},
						{
							Personas: []string{"MANAGEMENT"},
							Network: namespace.NetworkInterfaceNetwork{
								NetworkType: "DVPG",
								DVPGNetwork: &namespace.DVPGNetwork{
									Name:    "flb-mgmt",
									Network: "dvportgroup-64",
									IPAM:    "STATIC",
									IPConfig: &namespace.IPConfig{
										IPRanges: []namespace.IPRange{
											{
												Address: "172.25.0.2",
												Count:   196,
											},
										},
										Gateway: "172.25.0.1/16",
									},
								},
							},
						},
						{
							Personas: []string{"WORKLOAD"},
							Network: namespace.NetworkInterfaceNetwork{
								NetworkType: "PRIMARY_WORKLOAD",
							},
						},
					},
				},
				Provider: Strptr("VSPHERE_FOUNDATION"),
			},
			KubeAPIServerOptions: namespace.KubeAPIServerOptions{
				Security: &namespace.KubeAPIServerSecurity{
					CertificateDNSNames: []string{
						"domain-1.test",
						"wcp.integration.test",
						"xn--80akhbyknj4f",
					},
				},
			},
		},
	}
}

func buildMockSpecMultiZone(d *schema.ResourceData) *namespace.EnableOnZonesSpec {
	return &namespace.EnableOnZonesSpec{
		Zones: []string{"zone-1", "zone-2", "zone-3"},
		Name:  "supervisor",
		ControlPlane: namespace.ControlPlane{
			Network: namespace.ControlPlaneNetwork{
				Network: Strptr("network-47"),
				Backing: namespace.Backing{
					Backing: "NETWORK",
					Network: Strptr("network-47"),
				},
				Services: &namespace.Services{
					NTP: &namespace.NTP{
						Servers: []string{"ntp1.vcfd.broadcom.net"},
					},
				},
				FloatingIPAddress: Strptr("10.161.150.13"),
			},
			LoginBanner:   Strptr("Welcome to Supervisor on vSphere Zones!"),
			StoragePolicy: Strptr("ff84d20e-d62c-4a02-ad65-f85ff2121302"),
			Count:         Intptr(3),
		},
		Workloads: namespace.Workloads{
			Network: namespace.WorkloadNetwork{
				Network:     Strptr("primary"),
				NetworkType: "VSPHERE",
				VSphere: &namespace.NetworkVSphere{
					DVPG: "dvportgroup-60",
				},
				Services: &namespace.Services{
					DNS: &namespace.DNS{
						Servers: []string{"192.19.189.10", "192.19.189.20"},
						SearchDomains: []string{
							"domain-1.test",
							"wcp.integration.test",
							"xn--80akhbyknj4f",
						},
					},
					NTP: &namespace.NTP{
						Servers: []string{"ntp1.vcfd.broadcom.net"},
					},
				},
				IPManagement: &namespace.IPManagement{
					DHCPEnabled:    Boolptr(false),
					GatewayAddress: Strptr("192.168.1.1/16"),
					IPAssignments: &[]namespace.IPAssignment{
						{
							Assignee: Strptr("SERVICE"),
							Ranges: []namespace.IPRange{
								{Address: "172.24.0.0", Count: 65536},
							},
						},
						{
							Assignee: Strptr("NODE"),
							Ranges: []namespace.IPRange{
								{Address: "192.168.128.0", Count: 256},
							},
						},
					},
				},
			},
			Edge: namespace.Edge{
				ID: Strptr("haproxy"),
				LoadBalancerAddressRanges: &[]namespace.IPRange{
					{Address: "192.168.130.0", Count: 5120},
				},
				HAProxy: &namespace.HAProxy{
					Servers: []namespace.EdgeServer{
						{Host: "10.161.144.15", Port: 5556},
					},
					Username:                  "wcp",
					Password:                  "ca$hc0w",
					CertificateAuthorityChain: "-----BEGIN CERTIFICATE-----\nMIIDqTCCApGgAwIBAgIDAKkxMA0GCSqGSIb3DQEBCwUAMHUxCzAJBgNVBAYTAlVT\nMQswCQYDVQQIDAJDQTESMBAGA1UEBwwJUGFsbyBBbHRvMQ8wDQYDVQQKDAZWTXdh\ncmUxDDAKBgNVBAsMA1dDUDEmMCQGA1UEAwwdSEFQcm94eSBSb290IENBIENBIFhG\nRFhCNDlHT0EwHhcNMjUxMTI1MDk0NzE3WhcNMzAxMTI0MDk0NzE3WjB1MQswCQYD\nVQQGEwJVUzELMAkGA1UECAwCQ0ExEjAQBgNVBAcMCVBhbG8gQWx0bzEPMA0GA1UE\nCgwGVk13YXJlMQwwCgYDVQQLDANXQ1AxJjAkBgNVBAMMHUhBUHJveHkgUm9vdCBD\nQSBDQSBYRkRYQjQ5R09BMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\nm+BfkBhVoJyxUimFcaLOCtZZMhnEAIB4i5bls0KPxmTTfPvK5o3NYnlByZqN5vv+\nJwzphslS2BbTMC0Jg4nG0egv4Cus3bTFcCZVrqn7loZk6eqPXTFomEnFeyx06xm4\nLFAXegryfq/Bvm2RE2VZcFoW5mdvKuOllCagmfvufAY843zVCXPRbA50ghaYVsF9\nVdyKJ4P2KvmWuG23ABlRfHmMyLN1P58XhqR6EryMXhMFAx00fN/B8huTgbPoRtZd\n4TsbuSSkP0APNVQEZTHKDMjaayb/ePc/bFjjpuWhQL8SgXyCt575coMopYPvEqOz\nE9AY7QhqEncczx9LX79rxwIDAQABo0IwQDAPBgNVHRMBAf8EBTADAQH/MA4GA1Ud\nDwEB/wQEAwIBBjAdBgNVHQ4EFgQUuSg6qCVwMyYRqnXwZGDGNIjvGMMwDQYJKoZI\nhvcNAQELBQADggEBACPZiOba2O5vhAcwhe0TAnVNm1xSJN1Gce4p9v/oGISioYXH\nhRuY6ubP58OLinf5HPUGa9/Ss9pAUkyFbivrh2js29a6tZJWmxgiHWFXr4vnn0Ma\n8//INq8s1lEq+Mkrucd+fcOPpc9p1dw1SxfnHeLn1Q9JBYUtscoymyW1e1ljc727\n5g+O0XoKts5UgsFHnOuMBlOgIeYYj9FU6qQp2ASEoIGvJ/9Ph4nFlmetMG/TQhf1\ncCCCib3b5i70h7elwTlE830hKb0uH0fvcA2vHzDdwnKrOKTiRVuhSUXtm8FfvgWw\nbwSIO3Ryp2CBkyrpn0e7/SkyJnUEVt8yeXMz3J4=\n-----END CERTIFICATE-----\n",
				},
				Provider: Strptr("HAPROXY"),
			},
			KubeAPIServerOptions: namespace.KubeAPIServerOptions{
				Security: &namespace.KubeAPIServerSecurity{
					CertificateDNSNames: []string{
						"domain-1.test",
						"wcp.integration.test",
						"xn--80akhbyknj4f",
					},
				},
			},
		},
	}
}
