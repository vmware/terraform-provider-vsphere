// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/action/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	vimtypes "github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/virtualmachine"
)

type virtualMachinePowerAction struct {
	client *Client
}

var _ action.ActionWithConfigure = (*virtualMachinePowerAction)(nil)

// NewVirtualMachinePowerAction instantiates the vsphere_virtual_machine_power action.
func NewVirtualMachinePowerAction() action.Action {
	return &virtualMachinePowerAction{}
}

func (a *virtualMachinePowerAction) Metadata(_ context.Context, req action.MetadataRequest, resp *action.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machine_power"
}

func (a *virtualMachinePowerAction) Schema(_ context.Context, _ action.SchemaRequest, resp *action.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Sets the power state of an existing virtual machine identified by instance UUID.",
		Attributes: map[string]schema.Attribute{
			"uuid": schema.StringAttribute{
				Required:    true,
				Description: "The instance UUID of the virtual machine (same as vsphere_virtual_machine resource id).",
			},
			"power_state": schema.StringAttribute{
				Required:    true,
				Description: "Target power state: on, off, or suspended (matches vsphere_virtual_machine power_state values).",
			},
			"shutdown_wait_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Minutes to wait for guest shutdown when powering off. Default 3. Valid range 1–10.",
			},
			"force_power_off": schema.BoolAttribute{
				Optional:    true,
				Description: "If true, force power off when guest shutdown does not complete in time. Default true.",
			},
			"poweron_timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Seconds to retry power on. Default 300. Minimum 300.",
			},
		},
	}
}

func (a *virtualMachinePowerAction) Configure(_ context.Context, req action.ConfigureRequest, resp *action.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Action Configure Type",
			fmt.Sprintf("Expected *vsphere.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	a.client = client
}

type virtualMachinePowerModel struct {
	UUID                types.String `tfsdk:"uuid"`
	PowerState          types.String `tfsdk:"power_state"`
	ShutdownWaitTimeout types.Int64  `tfsdk:"shutdown_wait_timeout"`
	ForcePowerOff       types.Bool   `tfsdk:"force_power_off"`
	PoweronTimeout      types.Int64  `tfsdk:"poweron_timeout"`
}

func (a *virtualMachinePowerAction) Invoke(ctx context.Context, req action.InvokeRequest, resp *action.InvokeResponse) {
	var data virtualMachinePowerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if a.client == nil || a.client.vimClient == nil {
		resp.Diagnostics.AddError("Provider not configured", "The vSphere client is not available. Run apply with a configured provider.")
		return
	}

	want, ok := parseConfigPowerState(data.PowerState.ValueString())
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("power_state"),
			"Invalid power_state",
			`power_state must be one of: on, off, suspended.`,
		)
		return
	}

	shutdownMins := int64(3)
	if !data.ShutdownWaitTimeout.IsNull() && !data.ShutdownWaitTimeout.IsUnknown() {
		shutdownMins = data.ShutdownWaitTimeout.ValueInt64()
	}
	if shutdownMins < 1 || shutdownMins > 10 {
		resp.Diagnostics.AddAttributeError(
			path.Root("shutdown_wait_timeout"),
			"Invalid shutdown_wait_timeout",
			"shutdown_wait_timeout must be between 1 and 10 (minutes).",
		)
		return
	}

	forceOff := true
	if !data.ForcePowerOff.IsNull() && !data.ForcePowerOff.IsUnknown() {
		forceOff = data.ForcePowerOff.ValueBool()
	}

	powerOnSecs := int64(300)
	if !data.PoweronTimeout.IsNull() && !data.PoweronTimeout.IsUnknown() {
		powerOnSecs = data.PoweronTimeout.ValueInt64()
	}
	if powerOnSecs < 300 {
		resp.Diagnostics.AddAttributeError(
			path.Root("poweron_timeout"),
			"Invalid poweron_timeout",
			"poweron_timeout must be at least 300 (seconds).",
		)
		return
	}

	vm, err := virtualmachine.FromUUID(a.client.vimClient, data.UUID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Could not find virtual machine", err.Error())
		return
	}

	vprops, err := virtualmachine.Properties(vm)
	if err != nil {
		resp.Diagnostics.AddError("Error reading virtual machine", err.Error())
		return
	}

	if vprops.Runtime.PowerState == want {
		return
	}

	switch want {
	case vimtypes.VirtualMachinePowerStatePoweredOn:
		pTimeout := time.Duration(powerOnSecs) * time.Second
		if err := virtualmachine.PowerOn(vm, pTimeout); err != nil {
			resp.Diagnostics.AddError("Error powering on virtual machine", err.Error())
		}
	case vimtypes.VirtualMachinePowerStatePoweredOff:
		if err := virtualmachine.GracefulPowerOff(a.client.vimClient, vm, int(shutdownMins), forceOff); err != nil {
			resp.Diagnostics.AddError("Error powering off virtual machine", err.Error())
		}
	case vimtypes.VirtualMachinePowerStateSuspended:
		if err := virtualmachine.Suspend(vm); err != nil {
			resp.Diagnostics.AddError("Error suspending virtual machine", err.Error())
		}
	}
}

func parseConfigPowerState(s string) (vimtypes.VirtualMachinePowerState, bool) {
	switch s {
	case "on":
		return vimtypes.VirtualMachinePowerStatePoweredOn, true
	case "off":
		return vimtypes.VirtualMachinePowerStatePoweredOff, true
	case "suspended":
		return vimtypes.VirtualMachinePowerStateSuspended, true
	default:
		return "", false
	}
}
