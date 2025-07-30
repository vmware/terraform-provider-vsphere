package configprofile

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/govmomi/vapi/esx/settings/clusters/configuration"
	"github.com/vmware/govmomi/vapi/rest"
)

func ReadConfigProfile(ctx context.Context, client *rest.Client, d *schema.ResourceData) diag.Diagnostics {
	m := configuration.NewManager(client)

	clusterId := d.Get("cluster_id").(string)

	d.SetId(fmt.Sprintf("config_profile_%s", clusterId))

	tflog.Debug(ctx, fmt.Sprintf("reading configuration for cluster: %s", clusterId))
	config, err := m.GetConfiguration(clusterId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve cluster configuration: %s", err))
	}

	_ = d.Set("config", config.Config)

	tflog.Debug(ctx, fmt.Sprintf("reading configuration schema for cluster: %s", clusterId))
	configSchema, err := m.GetSchema(clusterId)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to retrieve configuration schema: %s", err))
	}

	_ = d.Set("schema", configSchema.Schema)

	return nil
}
