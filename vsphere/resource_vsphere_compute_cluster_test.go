// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/folder"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/hostsystem"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
)

const (
	testAccResourceVSphereComputeClusterNameStandard = "testacc-compute-cluster"
	testAccResourceVSphereComputeClusterNameRenamed  = "testacc-compute-cluster-renamed"
	testAccResourceVSphereComputeClusterFolder       = "compute-cluster-folder-test"
)

func TestAccResourceVSphereComputeCluster_basic(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(false),
				),
			},
			{
				ResourceName:      "vsphere_compute_cluster.compute_cluster",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					cluster, err := testGetComputeCluster(s, "compute_cluster", resourceVSphereComputeClusterName)
					if err != nil {
						return "", err
					}
					return cluster.InventoryPath, nil
				},
				ImportStateVerifyIgnore: []string{"force_evacuate_on_destroy"},
				Config:                  testAccResourceVSphereComputeClusterConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_haAdmissionControlPolicyDisabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigHAAdmissionControlPolicyDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(false),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_drsHAEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigDRSHABasic(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckHAEnabled(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vlcm(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVlcm(""),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigVlcm(testAccResourceVSphereComputeClusterImageConfig()),
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigVlcm(""),
				Check:  resource.ComposeTestCheckFunc(),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanDedupEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANDedupEnabledCompressEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_dedup_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_compression_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanCompressionEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANCompressionEnabledOnly(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_dedup_enabled", "false"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_compression_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanPerfEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANPerfEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_performance_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanPerfVerboseEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANPerfVerboseEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_performance_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_verbose_mode_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanPerfVerboseDiagnosticEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANPerfVerboseDiagnosticEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_performance_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_verbose_mode_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_network_diagnostic_mode_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanUnmapEnabledwithVsanEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANUnmapEnabledwithVsanEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_unmap_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanUnmapDisabledwithVsanDisabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANUnmapEnabledwithVsanEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_unmap_enabled", "true"),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANUnmapEnabledwithVsanDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "false"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_unmap_enabled", "true"),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANUnmapDisabledwithVsanDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "false"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_unmap_enabled", "false"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanDITEncryption(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{

				Config: testAccResourceVSphereComputeClusterConfigVSANDITEncryptionEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_dit_encryption_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_dit_rekey_interval", "1800"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanEsaEnabled(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
			testAccResourceVSphereComputeClusterVSANEsaPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigVSANEsaEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_enabled", "true"),
					resource.TestCheckResourceAttr("vsphere_compute_cluster.compute_cluster", "vsan_esa_enabled", "true"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_faultDomain(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigFaultDomains(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckTypeSetElemAttrPair(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_fault_domains.*.fault_domain.*.host_ids.*",
						"data.vsphere_host.roothost3",
						"id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_fault_domains.*.fault_domain.*.host_ids.*",
						"data.vsphere_host.roothost4",
						"id",
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_fault_domains.*.fault_domain.*",
						map[string]string{
							"name": "fd1",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_fault_domains.*.fault_domain.*",
						map[string]string{
							"name": "fd2",
						},
					),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_vsanStretchedCluster(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
			testAccResourceVSphereComputeClusterVSANStretchedClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterStretchedClusterEnabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					resource.TestCheckTypeSetElemAttrPair(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_stretched_cluster.*.preferred_fault_domain_host_ids.*",
						"data.vsphere_host.roothost1",
						"id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_stretched_cluster.*.secondary_fault_domain_host_ids.*",
						"data.vsphere_host.roothost2",
						"id",
					),
					resource.TestCheckTypeSetElemAttrPair(
						"vsphere_compute_cluster.compute_cluster",
						"vsan_stretched_cluster.*.witness_node",
						"data.vsphere_host.roothost3",
						"id",
					),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterStretchedClusterDisabled(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_explicitFailoverHost(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigDRSHABasicExplicitFailoverHost(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckDRSEnabled(true),
					testAccResourceVSphereComputeClusterCheckHAEnabled(true),
					testAccResourceVSphereComputeClusterCheckAdmissionControlMode(clusterAdmissionControlTypeFailoverHosts),
					testAccResourceVSphereComputeClusterCheckAdmissionControlFailoverHost(os.Getenv("TF_VAR_VSPHERE_ESXI3")),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_rename(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigWithName(testAccResourceVSphereComputeClusterNameStandard),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckName(testAccResourceVSphereComputeClusterNameStandard),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigWithName(testAccResourceVSphereComputeClusterNameRenamed),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckName(testAccResourceVSphereComputeClusterNameRenamed),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_inFolder(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigWithFolder(testAccResourceVSphereComputeClusterFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterMatchInventoryPath(testAccResourceVSphereComputeClusterFolder),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_moveToFolder(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterMatchInventoryPath(""),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigWithFolder(testAccResourceVSphereComputeClusterFolder),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterMatchInventoryPath(testAccResourceVSphereComputeClusterFolder),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_singleTag(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tag"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_multipleTags(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_switchTags(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tag"),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiTag(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckTags("testacc-tags-alt"),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_singleCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_multipleCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func TestAccResourceVSphereComputeCluster_switchCustomAttribute(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccResourceVSphereComputeClusterPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccResourceVSphereComputeClusterCheckExists(false),
		Steps: []resource.TestStep{
			{
				Config: testAccResourceVSphereComputeClusterConfigSingleCustomAttribute(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
			{
				Config: testAccResourceVSphereComputeClusterConfigMultiCustomAttributes(),
				Check: resource.ComposeTestCheckFunc(
					testAccResourceVSphereComputeClusterCheckExists(true),
					testAccResourceVSphereComputeClusterCheckCustomAttributes(),
				),
			},
		},
	})
}

func testAccResourceVSphereComputeClusterPreCheck(t *testing.T) {
	if os.Getenv("TF_VAR_VSPHERE_DATACENTER") == "" {
		t.Skip("set TF_VAR_VSPHERE_DATACENTER to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI3") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI3 to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_ESXI4") == "" {
		t.Skip("set TF_VAR_VSPHERE_ESXI4 to run vsphere_compute_cluster acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_PG_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_PG_NAME to run vsphere_virtual_machine acceptance tests")
	}
	if os.Getenv("TF_VAR_VSPHERE_NFS_DS_NAME") == "" {
		t.Skip("set TF_VAR_VSPHERE_NFS_DS_NAME to run vsphere_virtual_machine acceptance tests")
	}
}

func testAccResourceVSphereComputeClusterVSANEsaPreCheck(t *testing.T) {
	meta, err := testAccProviderMeta(t)
	if err != nil {
		t.Skip("can not get meta")
	}
	client, err := resourceVSphereComputeClusterClient(meta)
	if err != nil {
		t.Skip("can not get client")
	}

	// Minimum Supported Version: 8.0.0
	if version := viapi.ParseVersionFromClient(client); !version.AtLeast(viapi.VSphereVersion{Product: version.Product, Major: 8, Minor: 0}) {
		t.Skip("vSAN ESA acceptance test should be run on vSphere 8.0 or higher")
	}
}

func testAccResourceVSphereComputeClusterVSANStretchedClusterPreCheck(t *testing.T) {
	if os.Getenv("TF_VSPHERE_VSAN_HOST_1") == "" {
		t.Skip("set TF_VSPHERE_VSAN_HOST_1 to run vsphere_compute_cluster stretched cluster acceptance tests")
	}
	if os.Getenv("TF_VSPHERE_VSAN_HOST_2") == "" {
		t.Skip("set TF_VSPHERE_VSAN_HOST_2 to run vsphere_compute_cluster stretched cluster acceptance tests")
	}
	if os.Getenv("TF_VSPHERE_VSAN_WITNESS_HOST") == "" {
		t.Skip("set TF_VSPHERE_VSAN_WITNESS_HOST to run vsphere_compute_cluster stretched cluster acceptance tests")
	}
}

func testAccResourceVSphereComputeClusterCheckExists(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := testGetComputeCluster(s, "compute_cluster", resourceVSphereComputeClusterName)
		if err != nil {
			if viapi.IsManagedObjectNotFoundError(err) && expected == false {
				// Expected missing
				return nil
			}
			return err
		}
		if !expected {
			return errors.New("expected compute cluster to be missing")
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckDRSEnabled(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		actual := *props.ConfigurationEx.(*types.ClusterConfigInfoEx).DrsConfig.Enabled
		if expected != actual {
			return fmt.Errorf("expected enabled to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckHAEnabled(expected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		actual := *props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.Enabled
		if expected != actual {
			return fmt.Errorf("expected enabled to be %t, got %t", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckAdmissionControlMode(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}

		var actual string
		switch props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.AdmissionControlPolicy.(type) {
		case *types.ClusterFailoverResourcesAdmissionControlPolicy:
			actual = clusterAdmissionControlTypeResourcePercentage
		case *types.ClusterFailoverLevelAdmissionControlPolicy:
			actual = clusterAdmissionControlTypeSlotPolicy
		case *types.ClusterFailoverHostAdmissionControlPolicy:
			actual = clusterAdmissionControlTypeFailoverHosts
		default:
			actual = clusterAdmissionControlTypeDisabled
		}
		if expected != actual {
			return fmt.Errorf("expected admission control policy to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckAdmissionControlFailoverHost(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}

		failoverHostsPolicy, ok := props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.AdmissionControlPolicy.(*types.ClusterFailoverHostAdmissionControlPolicy)
		if !ok {
			return fmt.Errorf(
				"admission control policy is not *types.ClusterFailoverHostAdmissionControlPolicy (actual: %T)",
				props.ConfigurationEx.(*types.ClusterConfigInfoEx).DasConfig.AdmissionControlPolicy,
			)
		}

		// We just test the first host. The fixture this check is designed to be
		// used with currently only sets one failover host.
		if len(failoverHostsPolicy.FailoverHosts) < 1 {
			return errors.New("no failover hosts")
		}

		client := testAccProvider.Meta().(*Client).vimClient
		hs, err := hostsystem.FromID(client, failoverHostsPolicy.FailoverHosts[0].Value)
		if err != nil {
			return err
		}

		actual := hs.Name()
		if expected != actual {
			return fmt.Errorf("expected failover host name to be %s, got %s", expected, actual)
		}

		if *failoverHostsPolicy.ResourceReductionToToleratePercent != 0 {
			return fmt.Errorf("expected ha_admission_control_performance_tolerance be 0, got %d", failoverHostsPolicy.ResourceReductionToToleratePercent)
		}

		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckName(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testGetComputeCluster(s, "compute_cluster", resourceVSphereComputeClusterName)
		if err != nil {
			return err
		}
		actual := cluster.Name()
		if expected != actual {
			return fmt.Errorf("expected name to be %q, got %q", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterMatchInventoryPath(expected string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testGetComputeCluster(s, "compute_cluster", resourceVSphereComputeClusterName)
		if err != nil {
			return err
		}

		expected, err = folder.RootPathParticleHost.PathFromNewRoot(cluster.InventoryPath, folder.RootPathParticleHost, expected)
		actual := path.Dir(cluster.InventoryPath)
		if err != nil {
			return fmt.Errorf("bad: %s", err)
		}
		if expected != actual {
			return fmt.Errorf("expected path to be %s, got %s", expected, actual)
		}
		return nil
	}
}

func testAccResourceVSphereComputeClusterCheckTags(tagResName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		cluster, err := testGetComputeCluster(s, "compute_cluster", resourceVSphereComputeClusterName)
		if err != nil {
			return err
		}
		tagsClient, err := testAccProvider.Meta().(*Client).TagsManager()
		if err != nil {
			return err
		}
		return testObjectHasTags(s, tagsClient, cluster, tagResName)
	}
}

func testAccResourceVSphereComputeClusterCheckCustomAttributes() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		props, err := testGetComputeClusterProperties(s, "compute_cluster")
		if err != nil {
			return err
		}
		return testResourceHasCustomAttributeValues(s, "vsphere_compute_cluster", "compute_cluster", props.Entity())
	}
}

func testAccResourceVSphereComputeClusterConfigEmpty() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigHAAdmissionControlPolicyDisabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids             = [data.vsphere_host.roothost3.id]
  ha_enabled                  = true
  ha_admission_control_policy = "disabled"

  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANDedupEnabledCompressEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_dedup_enabled = true
  vsan_compression_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANCompressionEnabledOnly() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_dedup_enabled = false
  vsan_compression_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANPerfEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_performance_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANPerfVerboseEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_performance_enabled = true
  vsan_verbose_mode_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANPerfVerboseDiagnosticEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_performance_enabled = true
  vsan_verbose_mode_enabled = true
  vsan_network_diagnostic_mode_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANDITEncryptionEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled                = true
  vsan_dit_encryption_enabled = true
  vsan_dit_rekey_interval     = 1800
  force_evacuate_on_destroy   = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANUnmapEnabledwithVsanEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_unmap_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANUnmapEnabledwithVsanDisabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = false
  vsan_unmap_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANUnmapDisabledwithVsanDisabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = false
  vsan_unmap_enabled = false
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVSANEsaEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  vsan_enabled = true
  vsan_esa_enabled = true
  vsan_unmap_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigFaultDomains() string {
	return fmt.Sprintf(`
%s
resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]
  vsan_enabled = true
  vsan_fault_domains {
    fault_domain {
      name = "fd1"
      host_ids = [data.vsphere_host.roothost3.id]
    }
    fault_domain {
      name = "fd2"
      host_ids = [data.vsphere_host.roothost4.id]
    }
  }
  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
		),
	)
}

func testAccResourceVSphereComputeClusterStretchedClusterEnabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  vsan_enabled = true
  vsan_stretched_cluster {
    preferred_fault_domain_host_ids = [data.vsphere_host.roothost1.id]
    secondary_fault_domain_host_ids = [data.vsphere_host.roothost2.id]
    witness_node = data.vsphere_host.roothost3.id
  }
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataVsanHost1(),
			testhelper.ConfigDataVsanHost2(),
			testhelper.ConfigDataVsanWitnessHost(),
		),
	)
}

func testAccResourceVSphereComputeClusterStretchedClusterDisabled() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name                        = "testacc-compute-cluster"
  datacenter_id               = data.vsphere_datacenter.rootdc1.id
  host_system_ids             = [data.vsphere_host.roothost1.id, data.vsphere_host.roothost2.id]

  vsan_enabled = true
  force_evacuate_on_destroy = true
}

`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataVsanHost1(),
			testhelper.ConfigDataVsanHost2(),
			testhelper.ConfigDataVsanWitnessHost(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigBasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [ data.vsphere_host.roothost3.id ]

  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost3()))
}

func testAccResourceVSphereComputeClusterConfigDRSHABasic() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [data.vsphere_host.roothost3.id]

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled = true

	force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigVlcm(imageConfig string) string {
	return fmt.Sprintf(`
data "vsphere_host_base_images" "base_images" {}

%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [data.vsphere_host.roothost3.id]

  force_evacuate_on_destroy = true

  %s
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootHost2(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataSoftwareDepot(),
		),
		imageConfig,
	)
}

func testAccResourceVSphereComputeClusterImageConfig() string {
	return `
host_image {
  esx_version = "${data.vsphere_host_base_images.base_images.version.0}"
  component {
    key = vsphere_offline_software_depot.depot.component.0.key
    version = vsphere_offline_software_depot.depot.component.0.version.0
  }
}
`
}

func testAccResourceVSphereComputeClusterConfigDRSHABasicExplicitFailoverHost() string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name            = "testacc-compute-cluster"
  datacenter_id   = "${data.vsphere_datacenter.rootdc1.id}"
  host_system_ids = [data.vsphere_host.roothost3.id, data.vsphere_host.roothost4.id]

  drs_enabled          = true
  drs_automation_level = "fullyAutomated"

  ha_enabled                                    = true
  ha_admission_control_policy                   = "failoverHosts"
  ha_admission_control_failover_host_system_ids = [data.vsphere_host.roothost3.id]
  ha_admission_control_performance_tolerance    = 0

  force_evacuate_on_destroy = true
}
`,
		testhelper.CombineConfigs(
			testhelper.ConfigDataRootDC1(),
			testhelper.ConfigDataRootPortGroup1(),
			testhelper.ConfigDataRootHost3(),
			testhelper.ConfigDataRootHost4(),
			testhelper.ConfigDataRootComputeCluster1(),
			testhelper.ConfigDataRootDS1(),
			testhelper.ConfigDataRootVMNet(),
		),
	)
}

func testAccResourceVSphereComputeClusterConfigWithName(name string) string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "%s"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		name,
	)
}

func testAccResourceVSphereComputeClusterConfigWithFolder(f string) string {
	return fmt.Sprintf(`
%s

variable "folder" {
  default = "%s"
}

resource "vsphere_folder" "compute_cluster_folder" {
  path          = "${var.folder}"
  type          = "host"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"
  folder        = "${vsphere_folder.compute_cluster_folder.path}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
		f,
	)
}

func testAccResourceVSphereComputeClusterConfigSingleTag() string {
	return fmt.Sprintf(`
%s

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "ClusterComputeResource",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = [
    "${vsphere_tag.testacc-tag.id}",
  ]
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigMultiTag() string {
	return fmt.Sprintf(`
%s

variable "extra_tags" {
  default = [
    "terraform-test-thing1",
    "terraform-test-thing2",
  ]
}

resource "vsphere_tag_category" "testacc-category" {
  name        = "testacc-tag-category"
  cardinality = "MULTIPLE"

  associable_types = [
    "ClusterComputeResource",
  ]
}

resource "vsphere_tag" "testacc-tag" {
  name        = "testacc-tag"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_tag" "testacc-tags-alt" {
  count       = "${length(var.extra_tags)}"
  name        = "${var.extra_tags[count.index]}"
  category_id = "${vsphere_tag_category.testacc-category.id}"
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  tags = "${vsphere_tag.testacc-tags-alt.*.id}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigSingleCustomAttribute() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "ClusterComputeResource"
}

locals {
  attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
  }
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}

func testAccResourceVSphereComputeClusterConfigMultiCustomAttributes() string {
	return fmt.Sprintf(`
%s

resource "vsphere_custom_attribute" "testacc-attribute" {
  name                = "testacc-attribute"
  managed_object_type = "ClusterComputeResource"
}

resource "vsphere_custom_attribute" "testacc-attribute-2" {
  name                = "testacc-attribute-2"
  managed_object_type = "ClusterComputeResource"
}

locals {
  attrs = {
    "${vsphere_custom_attribute.testacc-attribute.id}" = "value"
    "${vsphere_custom_attribute.testacc-attribute-2.id}" = "value-2"
  }
}

resource "vsphere_compute_cluster" "compute_cluster" {
  name          = "testacc-compute-cluster"
  datacenter_id = "${data.vsphere_datacenter.rootdc1.id}"

  custom_attributes = "${local.attrs}"
}
`,
		testhelper.CombineConfigs(testhelper.ConfigDataRootDC1(), testhelper.ConfigDataRootPortGroup1()),
	)
}
