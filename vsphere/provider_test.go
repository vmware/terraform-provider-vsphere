// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProtoV6ProviderFactories map[string]func() (tfprotov6.ProviderServer, error)

// testAccProvider is the SDKv2 provider used for schema-based helpers (e.g.
// TestResourceDataRaw) and InternalValidate; it is not used directly as a
// Terraform provider in acceptance tests (those use the muxed Proto v6 server).
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProtoV6ProviderFactories = muxedProtoV6Factories()
}

func muxedProtoV6Factories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"vsphere": func() (tfprotov6.ProviderServer, error) {
			ctx := context.Background()
			upgradedSdkServer, err := tf5to6server.UpgradeServer(ctx, testAccProvider.GRPCProvider)
			if err != nil {
				return nil, err
			}
			protos := []func() tfprotov6.ProviderServer{
				providerserver.NewProtocol6(NewFrameworkProvider()),
				func() tfprotov6.ProviderServer {
					return upgradedSdkServer
				},
			}
			muxServer, err := tf6muxserver.NewMuxServer(ctx, protos...)
			if err != nil {
				return nil, err
			}
			return muxServer.ProviderServer(), nil
		},
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestMuxedProtoV6ProviderFactory(t *testing.T) {
	t.Parallel()
	f := testAccProtoV6ProviderFactories["vsphere"]
	if f == nil {
		t.Fatal("missing vsphere factory")
	}
	srv, err := f()
	if err != nil {
		t.Fatalf("factory: %v", err)
	}
	if srv == nil {
		t.Fatal("nil ProviderServer")
	}
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("VSPHERE_USER"); v == "" {
		t.Fatal("VSPHERE_USER must be set for acceptance tests")
	}

	if v := os.Getenv("VSPHERE_PASSWORD"); v == "" {
		t.Fatal("VSPHERE_PASSWORD must be set for acceptance tests")
	}

	if v := os.Getenv("VSPHERE_SERVER"); v == "" {
		t.Fatal("VSPHERE_SERVER must be set for acceptance tests")
	}
}

func testAccSkipUnstable(t *testing.T) {
	if skip, _ := strconv.ParseBool(os.Getenv("TF_VAR_VSPHERE_SKIP_UNSTABLE_TESTS")); skip {
		t.Skip()
	}
}

func testAccCheckEnvVariables(t *testing.T, variableNames []string) {
	for _, name := range variableNames {
		if v := os.Getenv(name); v == "" {
			t.Skipf("%s must be set for this acceptance test", name)
		}
	}
}

// testAccProviderMeta returns a instantiated VSphereClient for this provider.
// It's useful in state migration tests where a provider connection is actually
// needed, and we don't want to go through the regular provider configure
// channels (so this function doesn't interfere with the muxed acceptance test
// factories).
//
// Note we lean on environment variables for most of the provider configuration
// here and this will fail if those are missing. A pre-check is not run.
func testAccProviderMeta(t *testing.T) (interface{}, error) {
	t.Helper()
	d := schema.TestResourceDataRaw(t, testAccProvider.Schema, make(map[string]interface{}))
	return providerConfigure(d)
}
