// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FrameworkProvider is the plugin-framework half of the muxed provider.
type FrameworkProvider struct{}

// NewFrameworkProvider returns a new plugin-framework provider instance.
func NewFrameworkProvider() provider.Provider {
	return &FrameworkProvider{}
}

// Metadata implements provider.Provider.
func (p *FrameworkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vsphere"
}

// Schema implements provider.Provider.
func (p *FrameworkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"user": schema.StringAttribute{
				Optional:    true,
				Description: "The user name for vSphere API operations. Can be set with VSPHERE_USER.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "The user password for vSphere API operations. Can be set with VSPHERE_PASSWORD.",
			},
			"vsphere_server": schema.StringAttribute{
				Optional:    true,
				Description: "The vSphere Server name for vSphere API operations. Can be set with VSPHERE_SERVER.",
			},
			"allow_unverified_ssl": schema.BoolAttribute{
				Optional:    true,
				Description: "If set, VMware vSphere client will permit unverifiable SSL certificates. Can be set with VSPHERE_ALLOW_UNVERIFIED_SSL.",
			},
			"vcenter_server": schema.StringAttribute{
				Optional:           true,
				DeprecationMessage: "This field has been renamed to vsphere_server.",
				Description:        "Deprecated; use vsphere_server. Can be set with VSPHERE_VCENTER.",
			},
			"client_debug": schema.BoolAttribute{
				Optional:    true,
				Description: "govmomi debug. Can be set with VSPHERE_CLIENT_DEBUG.",
			},
			"client_debug_path_run": schema.StringAttribute{
				Optional:    true,
				Description: "govmomi debug path for a single run. Can be set with VSPHERE_CLIENT_DEBUG_PATH_RUN.",
			},
			"client_debug_path": schema.StringAttribute{
				Optional:    true,
				Description: "govmomi debug path for debug. Can be set with VSPHERE_CLIENT_DEBUG_PATH.",
			},
			"persist_session": schema.BoolAttribute{
				Optional:    true,
				Description: "Persist vSphere client sessions to disk. Can be set with VSPHERE_PERSIST_SESSION.",
			},
			"vim_session_path": schema.StringAttribute{
				Optional:    true,
				Description: "The directory to save vSphere SOAP API sessions to. Can be set with VSPHERE_VIM_SESSION_PATH.",
			},
			"rest_session_path": schema.StringAttribute{
				Optional:    true,
				Description: "The directory to save vSphere REST API sessions to. Can be set with VSPHERE_REST_SESSION_PATH.",
			},
			"vim_keep_alive": schema.Int64Attribute{
				Optional:    true,
				Description: "Keep alive interval for the VIM session in minutes. Can be set with VSPHERE_VIM_KEEP_ALIVE.",
			},
			"api_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "API timeout in minutes (Default: 5). Can be set with VSPHERE_API_TIMEOUT.",
			},
		},
	}
}

// Configure implements provider.Provider.
func (p *FrameworkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var c struct {
		User               types.String `tfsdk:"user"`
		Password           types.String `tfsdk:"password"`
		VSphereServer      types.String `tfsdk:"vsphere_server"`
		AllowUnverifiedSSL types.Bool   `tfsdk:"allow_unverified_ssl"`
		VCenterServer      types.String `tfsdk:"vcenter_server"`
		ClientDebug        types.Bool   `tfsdk:"client_debug"`
		ClientDebugPathRun types.String `tfsdk:"client_debug_path_run"`
		ClientDebugPath    types.String `tfsdk:"client_debug_path"`
		PersistSession     types.Bool   `tfsdk:"persist_session"`
		VimSessionPath     types.String `tfsdk:"vim_session_path"`
		RestSessionPath    types.String `tfsdk:"rest_session_path"`
		VimKeepAlive       types.Int64  `tfsdk:"vim_keep_alive"`
		APITimeout         types.Int64  `tfsdk:"api_timeout"`
	}

	resp.Diagnostics.Append(req.Config.Get(ctx, &c)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := ProviderFrameworkConfig{
		User:               frameworkStringValueOrNil(c.User),
		Password:           frameworkStringValueOrNil(c.Password),
		VSphereServer:      frameworkStringValueOrNil(c.VSphereServer),
		VCenterServer:      frameworkStringValueOrNil(c.VCenterServer),
		AllowUnverifiedSSL: frameworkBoolValueOrNil(c.AllowUnverifiedSSL),
		ClientDebug:        frameworkBoolValueOrNil(c.ClientDebug),
		ClientDebugPathRun: frameworkStringValueOrNil(c.ClientDebugPathRun),
		ClientDebugPath:    frameworkStringValueOrNil(c.ClientDebugPath),
		PersistSession:     frameworkBoolValueOrNil(c.PersistSession),
		VimSessionPath:     frameworkStringValueOrNil(c.VimSessionPath),
		RestSessionPath:    frameworkStringValueOrNil(c.RestSessionPath),
		VimKeepAlive:       frameworkIntValueOrNil(c.VimKeepAlive),
		APITimeoutMins:     frameworkIntValueOrNil(c.APITimeout),
	}

	settings, err := ProviderSettingsFromFramework(in)
	if err != nil {
		resp.Diagnostics.AddError("Invalid provider configuration", err.Error())
		return
	}

	SetDefaultAPITimeoutFromProviderSettings(settings)

	cfg, err := NewConfigFromProviderSettings(settings)
	if err != nil {
		resp.Diagnostics.AddError("Invalid provider configuration", err.Error())
		return
	}

	client, err := cfg.Client()
	if err != nil {
		resp.Diagnostics.AddError("Error creating vSphere client", err.Error())
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

// Resources implements provider.Provider.
func (p *FrameworkProvider) Resources(context.Context) []func() resource.Resource {
	return nil
}

// DataSources implements provider.Provider.
func (p *FrameworkProvider) DataSources(context.Context) []func() datasource.DataSource {
	return nil
}

func frameworkStringValueOrNil(v types.String) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

func frameworkBoolValueOrNil(v types.Bool) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

func frameworkIntValueOrNil(v types.Int64) *int {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := int(v.ValueInt64())
	return &i
}
