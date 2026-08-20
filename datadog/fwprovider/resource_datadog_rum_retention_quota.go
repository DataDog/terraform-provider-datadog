package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure      = &rumRetentionQuotaResource{}
	_ resource.ResourceWithImportState    = &rumRetentionQuotaResource{}
	_ resource.ResourceWithValidateConfig = &rumRetentionQuotaResource{}
)

// This endpoint is not yet part of the public datadog-api-client-go SDK, so requests
// are issued via utils.SendRequest against the raw path instead of a generated API method.
// The API supports a generic scope_type/scope_id pair, but only "application" is a valid
// scope_type today, so it's hardcoded here rather than exposed in the schema.
const rumRetentionQuotaPath = "/api/v2/rum/config/retention-quota/application/%s"

type rumRetentionQuotaResource struct {
	Api  *datadog.APIClient
	Auth context.Context
}

type rumRetentionQuotaModel struct {
	ID            types.String                    `tfsdk:"id"`
	ApplicationID types.String                    `tfsdk:"application_id"`
	Mode          types.String                    `tfsdk:"mode"`
	Custom        *rumRetentionQuotaCustomModel   `tfsdk:"custom"`
	Adaptive      *rumRetentionQuotaAdaptiveModel `tfsdk:"adaptive"`
}

type rumRetentionQuotaCustomModel struct {
	WindowType         types.String `tfsdk:"window_type"`
	SessionLimit       types.Int64  `tfsdk:"session_limit"`
	DailyResetTime     types.String `tfsdk:"daily_reset_time"`
	DailyResetTimezone types.String `tfsdk:"daily_reset_timezone"`
	QuotaReachedAction types.String `tfsdk:"quota_reached_action"`
}

type rumRetentionQuotaAdaptiveModel struct {
	MaxRetentionRate types.Float64 `tfsdk:"max_retention_rate"`
}

type rumRetentionQuotaCustomAttributes struct {
	WindowType         string `json:"window_type"`
	SessionLimit       int64  `json:"session_limit"`
	DailyResetTime     string `json:"daily_reset_time"`
	DailyResetTimezone string `json:"daily_reset_timezone"`
	QuotaReachedAction string `json:"quota_reached_action"`
}

type rumRetentionQuotaAdaptiveAttributes struct {
	MaxRetentionRate float64 `json:"max_retention_rate"`
}

type rumRetentionQuotaAttributes struct {
	Mode     string                               `json:"mode"`
	Custom   *rumRetentionQuotaCustomAttributes   `json:"custom,omitempty"`
	Adaptive *rumRetentionQuotaAdaptiveAttributes `json:"adaptive,omitempty"`
}

type rumRetentionQuotaRequestData struct {
	ID         string                       `json:"id"`
	Type       string                       `json:"type"`
	Attributes *rumRetentionQuotaAttributes `json:"attributes"`
}

type rumRetentionQuotaRequest struct {
	Data rumRetentionQuotaRequestData `json:"data"`
}

type rumRetentionQuotaResponseData struct {
	ID         string                      `json:"id"`
	Type       string                      `json:"type"`
	Attributes rumRetentionQuotaAttributes `json:"attributes"`
}

type rumRetentionQuotaResponse struct {
	Data rumRetentionQuotaResponseData `json:"data"`
}

func NewRumRetentionQuotaResource() resource.Resource {
	return &rumRetentionQuotaResource{}
}

func (r *rumRetentionQuotaResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.HttpClient
	r.Auth = providerData.Auth
}

func (r *rumRetentionQuotaResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "rum_retention_quota"
}

func (r *rumRetentionQuotaResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RumRetentionQuota resource. This can be used to create and manage Datadog rum_retention_quota.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "RUM application ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mode": schema.StringAttribute{
				Description: "The retention quota mode.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("custom", "adaptive"),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"custom": schema.SingleNestedBlock{
				Description: "Custom retention quota configuration. Required when `mode` is `custom`.",
				Attributes: map[string]schema.Attribute{
					"window_type": schema.StringAttribute{
						Description: "The window over which the quota resets.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("daily"),
						},
					},
					"session_limit": schema.Int64Attribute{
						Description: "The maximum number of sessions to retain within the window.",
						Required:    true,
					},
					"daily_reset_time": schema.StringAttribute{
						Description: "The time of day the quota resets, in `HH:MM` format.",
						Required:    true,
					},
					"daily_reset_timezone": schema.StringAttribute{
						Description: "The UTC offset for `daily_reset_time`, in `±HH:MM` format.",
						Required:    true,
					},
					"quota_reached_action": schema.StringAttribute{
						Description: "The action taken once the quota is reached.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("stop", "slowdown"),
						},
					},
				},
			},
			"adaptive": schema.SingleNestedBlock{
				Description: "Adaptive retention quota configuration. Required when `mode` is `adaptive`.",
				Attributes: map[string]schema.Attribute{
					"max_retention_rate": schema.Float64Attribute{
						Description: "The maximum share of sessions to retain, in the range `(0, 1]`.",
						Required:    true,
					},
				},
			},
		},
	}
}

func (r *rumRetentionQuotaResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var state rumRetentionQuotaModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if state.Mode.IsNull() || state.Mode.IsUnknown() {
		return
	}

	switch state.Mode.ValueString() {
	case "custom":
		if state.Custom == nil {
			response.Diagnostics.AddAttributeError(path.Root("custom"), "Missing Attribute Configuration", "the `custom` block is required when `mode` is `custom`")
		}
		if state.Adaptive != nil {
			response.Diagnostics.AddAttributeError(path.Root("adaptive"), "Conflicting Attribute Configuration", "the `adaptive` block must not be set when `mode` is `custom`")
		}
	case "adaptive":
		if state.Adaptive == nil {
			response.Diagnostics.AddAttributeError(path.Root("adaptive"), "Missing Attribute Configuration", "the `adaptive` block is required when `mode` is `adaptive`")
		}
		if state.Custom != nil {
			response.Diagnostics.AddAttributeError(path.Root("custom"), "Conflicting Attribute Configuration", "the `custom` block must not be set when `mode` is `adaptive`")
		}
	}
}

func (r *rumRetentionQuotaResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), request.ID)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("application_id"), request.ID)...)
}

func (r *rumRetentionQuotaResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state rumRetentionQuotaModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	requestPath := fmt.Sprintf(rumRetentionQuotaPath, state.ApplicationID.ValueString())
	respBytes, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodGet, requestPath, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving RumRetentionQuota"))
		return
	}

	var resp rumRetentionQuotaResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		response.Diagnostics.AddError("error parsing RumRetentionQuota response", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionQuotaResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state rumRetentionQuotaModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildRequestBody(&state)
	requestPath := fmt.Sprintf(rumRetentionQuotaPath, state.ApplicationID.ValueString())

	respBytes, _, err := utils.SendRequest(r.Auth, r.Api, http.MethodPut, requestPath, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating RumRetentionQuota"))
		return
	}

	var resp rumRetentionQuotaResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		response.Diagnostics.AddError("error parsing RumRetentionQuota response", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionQuotaResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state rumRetentionQuotaModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildRequestBody(&state)
	requestPath := fmt.Sprintf(rumRetentionQuotaPath, state.ApplicationID.ValueString())

	respBytes, _, err := utils.SendRequest(r.Auth, r.Api, http.MethodPut, requestPath, body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating RumRetentionQuota"))
		return
	}

	var resp rumRetentionQuotaResponse
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		response.Diagnostics.AddError("error parsing RumRetentionQuota response", err.Error())
		return
	}
	r.updateState(&state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionQuotaResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state rumRetentionQuotaModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	requestPath := fmt.Sprintf(rumRetentionQuotaPath, state.ApplicationID.ValueString())
	_, httpResp, err := utils.SendRequest(r.Auth, r.Api, http.MethodDelete, requestPath, nil)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting RumRetentionQuota"))
		return
	}
}

func (r *rumRetentionQuotaResource) updateState(state *rumRetentionQuotaModel, resp *rumRetentionQuotaResponse) {
	data := resp.Data
	attributes := data.Attributes

	state.ID = types.StringValue(data.ID)
	state.Mode = types.StringValue(attributes.Mode)

	if attributes.Custom != nil {
		state.Custom = &rumRetentionQuotaCustomModel{
			WindowType:         types.StringValue(attributes.Custom.WindowType),
			SessionLimit:       types.Int64Value(attributes.Custom.SessionLimit),
			DailyResetTime:     types.StringValue(attributes.Custom.DailyResetTime),
			DailyResetTimezone: types.StringValue(attributes.Custom.DailyResetTimezone),
			QuotaReachedAction: types.StringValue(attributes.Custom.QuotaReachedAction),
		}
	} else {
		state.Custom = nil
	}

	if attributes.Adaptive != nil {
		state.Adaptive = &rumRetentionQuotaAdaptiveModel{
			MaxRetentionRate: types.Float64Value(attributes.Adaptive.MaxRetentionRate),
		}
	} else {
		state.Adaptive = nil
	}
}

func (r *rumRetentionQuotaResource) buildRequestBody(state *rumRetentionQuotaModel) *rumRetentionQuotaRequest {
	attributes := &rumRetentionQuotaAttributes{
		Mode: state.Mode.ValueString(),
	}

	if state.Custom != nil {
		attributes.Custom = &rumRetentionQuotaCustomAttributes{
			WindowType:         state.Custom.WindowType.ValueString(),
			SessionLimit:       state.Custom.SessionLimit.ValueInt64(),
			DailyResetTime:     state.Custom.DailyResetTime.ValueString(),
			DailyResetTimezone: state.Custom.DailyResetTimezone.ValueString(),
			QuotaReachedAction: state.Custom.QuotaReachedAction.ValueString(),
		}
	}

	if state.Adaptive != nil {
		attributes.Adaptive = &rumRetentionQuotaAdaptiveAttributes{
			MaxRetentionRate: state.Adaptive.MaxRetentionRate.ValueFloat64(),
		}
	}

	return &rumRetentionQuotaRequest{
		Data: rumRetentionQuotaRequestData{
			ID:         state.ApplicationID.ValueString(),
			Type:       "rum_quota_config",
			Attributes: attributes,
		},
	}
}
