// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/vmware/terraform-provider-vsphere/v2/vsphere"
)

func main() {
	var debugMode bool
	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		ProviderFunc: vsphere.Provider,
	}

	if debugMode {
		opts.Debug = true
		opts.ProviderAddr = "vmware/vsphere"
	}

	plugin.Serve(opts)
}
