package fwprovider

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ resource.ResourceWithConfigure = &bitsAIMonitorAutomationResource{}
var _ resource.ResourceWithImportState = &bitsAIMonitorAutomationResource{}

type monitorAutomationAPI interface {
	GetMonitorAutomation(context.Context, int64) (datadogV2.MonitorAutomationResponse, *http.Response, error)
	UpdateMonitorAutomation(context.Context, int64, datadogV2.MonitorAutomationRequest) (datadogV2.MonitorAutomationResponse, *http.Response, error)
}

type bitsAIMonitorAutomationResource struct {
	api  monitorAutomationAPI
	auth context.Context
}

type bitsAIMonitorAutomationModel struct {
	ID        types.String `tfsdk:"id"`
	MonitorID types.String `tfsdk:"monitor_id"`
	Enabled   types.Bool   `tfsdk:"enabled"`
}

func NewBitsAIMonitorAutomationResource() resource.Resource {
	return &bitsAIMonitorAutomationResource{}
}

func (r *bitsAIMonitorAutomationResource) Metadata(_ context.Context, _ resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "bits_ai_monitor_automation"
}

func (r *bitsAIMonitorAutomationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Bits automatic investigations for an existing monitor. Requires access to the monitor and the Bits monitor automation API. Read requires monitors_read and bits_investigations_read; writes also require monitors_write and bits_investigations_write. Destroying this resource disables automatic investigations without deleting the monitor.",
		Attributes: map[string]schema.Attribute{
			"id":         schema.StringAttribute{Computed: true, Description: "The monitor ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}},
			"monitor_id": schema.StringAttribute{Required: true, Description: "ID of the existing monitor to configure. Manage each monitor's automatic investigation setting with only one resource.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"enabled":    schema.BoolAttribute{Required: true, Description: "Whether Bits automatically investigates alerts from the monitor. Set false to disable investigations."},
		},
	}
}

func (r *bitsAIMonitorAutomationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*FrameworkProvider)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data", fmt.Sprintf("Expected *FrameworkProvider, got %T", req.ProviderData))
		return
	}
	client := data.DatadogApiInstances.HttpClient
	client.GetConfig().SetUnstableOperationEnabled("v2.GetMonitorAutomation", true)
	client.GetConfig().SetUnstableOperationEnabled("v2.UpdateMonitorAutomation", true)
	r.api = datadogV2.NewBitsAIApi(client)
	r.auth = data.Auth
}

func monitorAutomationID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("monitor_id must be a positive integer")
	}
	return id, nil
}

func (r *bitsAIMonitorAutomationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bitsAIMonitorAutomationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := monitorAutomationID(state.MonitorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid monitor ID", err.Error())
		return
	}
	result, httpResp, err := r.api.GetMonitorAutomation(r.auth, id)
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading monitor automation"))
		return
	}
	if err := utils.CheckForUnparsed(result); err != nil {
		resp.Diagnostics.AddError("Invalid API response", err.Error())
		return
	}
	state.Enabled = types.BoolValue(result.Data.Attributes.Enabled)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bitsAIMonitorAutomationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bitsAIMonitorAutomationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	applied, diags := r.apply(ctx, &plan)
	// Preserve the ID after a successful write even if waiting for read visibility fails.
	if applied {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
	resp.Diagnostics.Append(diags...)
}

func (r *bitsAIMonitorAutomationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bitsAIMonitorAutomationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	applied, diags := r.apply(ctx, &plan)
	if applied {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}
	resp.Diagnostics.Append(diags...)
}

func monitorAutomationBody(enabled bool) datadogV2.MonitorAutomationRequest {
	return datadogV2.MonitorAutomationRequest{Data: datadogV2.MonitorAutomationRequestData{
		Type:       datadogV2.MONITORAUTOMATIONTYPE_MONITOR_AUTOMATION,
		Attributes: datadogV2.MonitorAutomationAttributes{Enabled: enabled},
	}}
}

// waitMonitorAutomation retries only explicit eventual-consistency conditions.
// Permission errors and other failures are surfaced immediately.
func waitMonitorAutomation(ctx context.Context, check func() (bool, error)) error {
	for {
		done, err := check()
		if err != nil || done {
			return err
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (r *bitsAIMonitorAutomationResource) apply(ctx context.Context, plan *bitsAIMonitorAutomationModel) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	id, err := monitorAutomationID(plan.MonitorID.ValueString())
	if err != nil {
		diags.AddError("Invalid monitor ID", err.Error())
		return false, diags
	}
	// The authenticated context carries provider credentials; tie it to a bounded
	// operation and the caller's cancellation without dropping authentication.
	authCtx, cancel := context.WithTimeout(r.auth, 2*time.Minute)
	stop := context.AfterFunc(ctx, cancel)
	defer stop()
	defer cancel()
	enabled := plan.Enabled.ValueBool()
	err = waitMonitorAutomation(authCtx, func() (bool, error) {
		_, httpResp, err := r.api.UpdateMonitorAutomation(authCtx, id, monitorAutomationBody(enabled))
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		return err == nil, err
	})
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error updating monitor automation"))
		return false, diags
	}
	plan.ID = plan.MonitorID
	err = waitMonitorAutomation(authCtx, func() (bool, error) {
		result, httpResp, err := r.api.GetMonitorAutomation(authCtx, id)
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		if err := utils.CheckForUnparsed(result); err != nil {
			return false, err
		}
		return result.Data.Attributes.Enabled == enabled, nil
	})
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "monitor automation was updated but its read state did not converge"))
	}
	return true, diags
}

func (r *bitsAIMonitorAutomationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bitsAIMonitorAutomationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, err := monitorAutomationID(state.MonitorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Invalid monitor ID", err.Error())
		return
	}
	_, httpResp, err := r.api.UpdateMonitorAutomation(r.auth, id, monitorAutomationBody(false))
	if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
		return
	}
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error disabling monitor automation"))
	}
}

func (r *bitsAIMonitorAutomationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if _, err := monitorAutomationID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid monitor ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("monitor_id"), req.ID)...)
}
