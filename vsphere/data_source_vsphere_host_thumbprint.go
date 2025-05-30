// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"bytes"
	// TODO: Transition to crypto/sha256 in next major release.
	"crypto/sha1" //nolint
	"crypto/tls"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceVSphereHostThumbprint() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereHostThumbprintRead,
		Schema: map[string]*schema.Schema{
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The address of the ESXi to extract the thumbprint from.",
			},
			"port": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "443",
				Description: "The port to connect to on the ESXi host.",
			},
			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean that can be set to true to disable SSL certificate verification.",
			},
		},
	}
}

func dataSourceVSphereHostThumbprintRead(d *schema.ResourceData, _ interface{}) error {
	config := &tls.Config{
		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2 or higher.
	}
	config.InsecureSkipVerify = d.Get("insecure").(bool)
	conn, err := tls.Dial("tcp", d.Get("address").(string)+":"+d.Get("port").(string), config)
	if err != nil {
		return err
	}
	cert := conn.ConnectionState().PeerCertificates[0]
	fingerprint := sha1.Sum(cert.Raw)

	var buf bytes.Buffer
	for i, f := range fingerprint {
		if i > 0 {
			_, _ = fmt.Fprintf(&buf, ":")
		}
		_, _ = fmt.Fprintf(&buf, "%02X", f)
	}
	d.SetId(buf.String())
	return nil
}
