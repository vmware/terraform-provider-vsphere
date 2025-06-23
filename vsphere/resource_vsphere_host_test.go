// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/hostsystem"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/testhelper"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere/internal/helper/viapi"
)

var fallbackNtpServers = []string{
	"0.pool.ntp.org",
	"1.pool.ntp.org",
	"2.pool.ntp.org",
	"3.pool.ntp.org",
}

func TestAccResourceVSphereHost_basic(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereHostConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
			{
				ResourceName: "vsphere_host.h1",
				Config:       testaccvspherehostconfigImport(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceVSphereHost_rootFolder(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigRootfolder(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_connection(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigConnection(false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostConnected("vsphere_host.h1", false),
				),
			},
			{
				Config: testaccvspherehostconfigConnection(true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostConnected("vsphere_host.h1", true),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_maintenance(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigMaintenance(true),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostMaintenanceState("vsphere_host.h1", true),
				),
			},
			{
				Config: testaccvspherehostconfigMaintenance(false),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostMaintenanceState("vsphere_host.h1", false),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_lockdown(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigLockdown("strict"),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostLockdownState("vsphere_host.h1", "strict"),
				),
			},
			{
				Config: testaccvspherehostconfigLockdown("normal"),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostLockdownState("vsphere_host.h1", "normal"),
				),
			},
			{
				Config: testaccvspherehostconfigLockdown("disabled"),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccVSphereHostLockdownState("vsphere_host.h1", "disabled"),
				),
			},
		},
	})
}

func TestAccResourceVSphereHost_lockdown_invalid(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testaccvspherehostconfigLockdown("invalidvalue"),
				ExpectError: regexp.MustCompile(`be one of \[disabled normal strict\], got invalidvalue`),
			},
		},
	})
}

func TestAccResourceVSphereHost_emptyLicense(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testaccvspherehostconfigEmptylicense(),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
				),
			},
		},
	})
}

func testAccVSphereHostExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostExists(client, hostID)
		if err != nil {
			return err
		}

		if !res {
			return fmt.Errorf("host with ID %s not found", hostID)
		}

		return nil
	}
}

func testAccVSphereHostConnected(name string, shouldBeConnected bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostConnected(client, hostID)
		if err != nil {
			return err
		}

		if res != shouldBeConnected {
			return fmt.Errorf("host with ID %s connection: %t, expected %t", hostID, res, shouldBeConnected)
		}

		return nil
	}
}

func testAccVSphereHostMaintenanceState(name string, inMaintenance bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostInMaintenance(client, hostID)
		if err != nil {
			return err
		}

		if res != inMaintenance {
			return fmt.Errorf("host with ID %s in maintenance: %t, expected %t", hostID, res, inMaintenance)
		}

		return nil
	}
}

func testAccVSphereHostLockdownState(name string, lockdown string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]

		if !ok {
			return fmt.Errorf("%s key not found on the server", name)
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := checkHostLockdown(client, hostID, lockdown)
		if err != nil {
			return err
		}

		if !res {
			return fmt.Errorf("host with ID %s not in desired lockdown state. current state: %s", hostID, lockdown)
		}

		return nil
	}
}

func testAccVSphereHostDestroy(s *terraform.State) error {
	message := ""
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vsphere_host" {
			continue
		}
		hostID := rs.Primary.ID
		client := testAccProvider.Meta().(*Client).vimClient
		res, err := hostExists(client, hostID)
		if err != nil {
			return err
		}

		if res {
			message += fmt.Sprintf("Host with ID %s was found", hostID)
		}
	}
	if message != "" {
		return errors.New(message)
	}
	return nil
}

func hostExists(client *govmomi.Client, hostID string) (bool, error) {
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		if viapi.IsManagedObjectNotFoundError(err) {
			return false, nil
		}
		return false, err
	}

	if hs.Reference().Value != hostID {
		return false, nil
	}
	return true, nil
}

func hostConnected(client *govmomi.Client, hostID string) (bool, error) {
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return false, err
	}

	connectionState, err := hostsystem.GetConnectionState(hs)
	if err != nil {
		return false, err
	}
	return connectionState == types.HostSystemConnectionStateConnected, nil
}

func hostInMaintenance(client *govmomi.Client, hostID string) (bool, error) {
	hs, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return false, err
	}

	maintenanceState, err := hostsystem.HostInMaintenance(hs)
	if err != nil {
		return false, err
	}
	return maintenanceState, nil
}

func checkHostLockdown(client *govmomi.Client, hostID, lockdownMode string) (bool, error) {
	host, err := hostsystem.FromID(client, hostID)
	if err != nil {
		return false, err
	}
	hostProps, err := hostsystem.Properties(host)
	if err != nil {
		return false, err
	}

	lockdownModes := map[types.HostLockdownMode]string{
		types.HostLockdownModeLockdownDisabled: "disabled",
		types.HostLockdownModeLockdownNormal:   "normal",
		types.HostLockdownModeLockdownStrict:   "strict",
	}

	modeString, ok := lockdownModes[hostProps.Config.LockdownMode]
	if !ok {
		return false, fmt.Errorf("unknown lockdown mode found: %s", hostProps.Config.LockdownMode)
	}

	return modeString == lockdownMode, nil
}

func TestAccResourceVSphereHostNtpService(t *testing.T) {
	configs := []NtpdServiceConfig{
		{Enabled: true, Policy: "on"},
		{Enabled: true, Policy: "off"},
		{Enabled: true, Policy: "automatic"},
		{Enabled: false, Policy: "on"},
		{Enabled: false, Policy: "off"},
		{Enabled: false, Policy: "automatic"},
	}

	for _, config := range configs {
		t.Run(fmt.Sprintf("Enabled=%t,Policy=%s", config.Enabled, config.Policy), func(t *testing.T) {
			testAccSkipUnstable(t)
			resource.Test(t, resource.TestCase{
				PreCheck:     func() { testAccPreCheck(t) },
				Providers:    testAccProviders,
				CheckDestroy: testAccVSphereHostDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccVSphereHostConfigNtpdServices(config),
						Check: resource.ComposeTestCheckFunc(
							testAccCheckVSphereHostNTPServiceState("vsphere_host.h1", config),
						),
					},
				},
			})
		})
	}
}

func testAccCheckVSphereHostNTPServiceState(resourceName string, config NtpdServiceConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		actualEnabled := rs.Primary.Attributes["services.0.ntpd.0.enabled"]
		actualPolicy := rs.Primary.Attributes["services.0.ntpd.0.policy"]

		expectedEnabled := strconv.FormatBool(config.Enabled)
		expectedPolicy := config.Policy

		if actualEnabled != expectedEnabled {
			return fmt.Errorf("expected NTPD service enabled state: %s, got: %s", expectedEnabled, actualEnabled)
		}
		if actualPolicy != expectedPolicy {
			return fmt.Errorf("expected NTPD policy: %s, got: %s", expectedPolicy, actualPolicy)
		}

		return nil
	}
}

func TestAccResourceVSphereHostNTPServers(t *testing.T) {
	ntpServers := os.Getenv("NTP_SERVERS")
	if ntpServers == "" {
		t.Log("NTP_SERVERS environment variable is not set, using fallback value")
		ntpServers = strings.Join(fallbackNtpServers, ",")
	}

	// Split the environment variable into a slice of strings
	ntpServersSlice := strings.Split(ntpServers, ",")

	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCheckEnvVariables(t, []string{"ESX_HOSTNAME", "ESX_USERNAME", "ESX_PASSWORD"})
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccVSphereHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVSphereHostConfigNtpServers(ntpServersSlice),
				Check: resource.ComposeTestCheckFunc(
					testAccVSphereHostExists("vsphere_host.h1"),
					testAccCheckVSphereHostNTPServers("vsphere_host.h1", ntpServersSlice),
				),
			},
		},
	})
}

func testAccCheckVSphereHostNTPServers(resourceName string, expectedServers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		log.Printf("Resource attributes: %+v", rs.Primary.Attributes)

		// Retrieve the number of NTP servers
		ntpServersCount, err := strconv.Atoi(rs.Primary.Attributes["services.0.ntpd.0.ntp_servers.#"])
		if err != nil {
			return fmt.Errorf("error converting ntp_servers count: %s", err)
		}

		// Collect actual NTP servers
		var actualServers []string
		for i := 0; i < ntpServersCount; i++ {
			server := rs.Primary.Attributes[fmt.Sprintf("services.0.ntpd.0.ntp_servers.%d", i)]
			actualServers = append(actualServers, server)
		}

		expectedServersStr := strings.Join(expectedServers, ",")
		actualServersStr := strings.Join(actualServers, ",")

		if actualServersStr != expectedServersStr {
			return fmt.Errorf("expected NTP servers: %s, got: %s", expectedServersStr, actualServersStr)
		}

		return nil
	}
}

func testAccVSphereHostConfig() string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}

	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  # Makes sense to update
	  license = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testaccvspherehostconfigRootfolder() string {
	return fmt.Sprintf(`
	%s

	data "vsphere_host_thumbprint" "thumbprint1" {
	  address = "%s"
	  insecure = true
	}

	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "root"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.thumbprint1.id

	  datacenter = data.vsphere_datacenter.rootdc1.id

	  lifecycle {
		  ignore_changes = ["services"]
	  }

	}
	`, testhelper.ConfigDataRootDC1(),
		os.Getenv("TF_VAR_VSPHERE_ESXI4"), // for thumbprint retrieval
		os.Getenv("TF_VAR_VSPHERE_ESXI4"), // for connection config
		os.Getenv("TF_VAR_VSPHERE_ESXI4_PASSWORD"))
}

func testaccvspherehostconfigEmptylicense() string {
	return fmt.Sprintf(`
	%s
	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  # Makes sense to update
	  datacenter = data.vsphere_datacenter.rootdc1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
	)
}

func testaccvspherehostconfigImport() string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}

	resource "vsphere_host" "h1" {
	  # Useful only for connection
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  # Makes sense to update
	  license = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"))
}

func testaccvspherehostconfigConnection(connection bool) string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}

	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  license = "%s"
	  connected = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"),
		strconv.FormatBool(connection))
}

func testaccvspherehostconfigMaintenance(maintenance bool) string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}

	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  license = "%s"
	  connected = "true"
	  maintenance = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"),
		strconv.FormatBool(maintenance))
}

func testaccvspherehostconfigLockdown(lockdown string) string {
	return fmt.Sprintf(`
	%s

	resource "vsphere_compute_cluster" "c1" {
	  name = "%s"
	  datacenter_id = data.vsphere_datacenter.rootdc1.id
	}

	resource "vsphere_host" "h1" {
	  hostname = "%s"
	  username = "%s"
	  password = "%s"
	  thumbprint = data.vsphere_host_thumbprint.id

	  license = "%s"
	  connected = "true"
	  maintenance = "false"
	  lockdown = "%s"
	  cluster = vsphere_compute_cluster.c1.id
	}
	`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		os.Getenv("TF_VAR_VSPHERE_LICENSE"),
		lockdown)
}

type NtpdServiceConfig struct {
	Enabled bool
	Policy  string
}

func testAccVSphereHostConfigNtpdServices(config NtpdServiceConfig) string {
	return fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "c1" {
  name = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

data "vsphere_host_thumbprint" "thumbprint" {
    address = "%s"
    insecure = true
}

resource "vsphere_host" "h1" {
    hostname = "%s"
    username = "%s"
    password = "%s"
    thumbprint = data.vsphere_host_thumbprint.thumbprint.id
    services {
        ntpd {
            enabled = "%s"
            policy  = "%s"
        }
    }
    cluster = vsphere_compute_cluster.c1.id
}
`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		strconv.FormatBool(config.Enabled),
		config.Policy)
}

func testAccVSphereHostConfigNtpServers(ntpServers []string) string {
	// Convert the NTP servers slice to a string suitable for interpolation in the Terraform configuration
	serversStr := strings.Join(ntpServers, "\", \"")
	config := fmt.Sprintf(`
%s

resource "vsphere_compute_cluster" "c1" {
  name = "%s"
  datacenter_id = data.vsphere_datacenter.rootdc1.id
}

data "vsphere_host_thumbprint" "thumbprint" {
    address = "%s"
    insecure = true
}

resource "vsphere_host" "h1" {
    hostname = "%s"
    username = "%s"
    password = "%s"
    thumbprint = data.vsphere_host_thumbprint.thumbprint.id
    services {
        ntpd {
             ntp_servers = ["%s"]
             enabled = true
             policy = "on"
        }
    }
    cluster = vsphere_compute_cluster.c1.id
}`, testhelper.ConfigDataRootDC1(),
		"TestCluster",
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_HOSTNAME"),
		os.Getenv("ESX_USERNAME"),
		os.Getenv("ESX_PASSWORD"),
		serversStr)
	log.Printf("Generated Terraform configuration: %s", config)
	return config
}
