package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &governanceControlResource{}
	_ resource.ResourceWithImportState = &governanceControlResource{}
)

type governanceControlResource struct {
	Api  *datadogV2.GovernanceControlsApi
	Auth context.Context
}

type governanceControlModel struct {
	ID                     types.String         `tfsdk:"id"`
	DetectionType          types.String         `tfsdk:"detection_type"`
	Name                   types.String         `tfsdk:"name"`
	DetectionFrequency     types.String         `tfsdk:"detection_frequency"`
	DetectionParameters    jsontypes.Normalized `tfsdk:"detection_parameters"`
	MitigationType         types.String         `tfsdk:"mitigation_type"`
	MitigationParameters   jsontypes.Normalized `tfsdk:"mitigation_parameters"`
	NotificationType       types.String         `tfsdk:"notification_type"`
	NotificationFrequency  types.String         `tfsdk:"notification_frequency"`
	NotificationParameters jsontypes.Normalized `tfsdk:"notification_parameters"`
}

func NewGovernanceControlResource() resource.Resource {
	return &governanceControlResource{}
}

func (r *governanceControlResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetGovernanceControlsApiV2()
	r.Auth = providerData.Auth
}

func (r *governanceControlResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "governance_control"
}

func (r *governanceControlResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Governance Control resource. This can be used to configure built-in Governance Console controls, such as their detection, mitigation, and notification settings. Controls are built into Datadog: this resource configures an existing control rather than creating one, and removing it from Terraform only removes it from state.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"detection_type": schema.StringAttribute{
				Required:    true,
				Description: "The detection type that uniquely identifies the control, for example `unused_api_keys`.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Human-readable name of the control.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"detection_frequency": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "How often detections are evaluated for the control.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"detection_parameters": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Detection parameters for the control, as a JSON-encoded map of parameter names to their configured values.",
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mitigation_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The mitigation type configured for the control. Empty when not configured.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mitigation_parameters": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Mitigation parameters for the control, as a JSON-encoded map of parameter names to their configured values.",
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notification_type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The notification type configured for the control. Empty when not configured.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notification_frequency": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The notification frequency configured for the control. Empty when not configured.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notification_parameters": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Notification parameters for the control, as a JSON-encoded map of parameter names to their configured values.",
				CustomType:  jsontypes.NormalizedType{},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *governanceControlResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *governanceControlResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state governanceControlModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, httpResp, err := r.Api.GetGovernanceControl(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving governance control"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("datadog_governance_control: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(&state, &resp); err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating state from governance control response"))
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *governanceControlResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state governanceControlModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	r.upsertGovernanceControl(&state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *governanceControlResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state governanceControlModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	r.upsertGovernanceControl(&state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *governanceControlResource) Delete(_ context.Context, _ resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddWarning("resource cannot be deleted", "governance controls are built into Datadog and cannot be deleted; the control is only removed from Terraform state")
}

// upsertGovernanceControl backs both Create and Update: controls are built into
// Datadog, so "creating" the resource means configuring the existing control.
func (r *governanceControlResource) upsertGovernanceControl(state *governanceControlModel, diags *diag.Diagnostics) {
	req := r.buildGovernanceControlUpdateRequest(state, diags)
	if diags.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateGovernanceControl(r.Auth, state.DetectionType.ValueString(), *req)
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error updating governance control"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("datadog_governance_control: response contains unparsedObject", err.Error())
		return
	}

	if err := r.updateState(state, &resp); err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error updating state from governance control response"))
	}
}

func (r *governanceControlResource) buildGovernanceControlUpdateRequest(state *governanceControlModel, diags *diag.Diagnostics) *datadogV2.GovernanceControlUpdateRequest {
	attributes := datadogV2.NewGovernanceControlUpdateAttributesWithDefaults()

	if !state.Name.IsNull() && !state.Name.IsUnknown() {
		attributes.SetName(state.Name.ValueString())
	}
	if !state.DetectionFrequency.IsNull() && !state.DetectionFrequency.IsUnknown() {
		attributes.SetDetectionFrequency(state.DetectionFrequency.ValueString())
	}
	if !state.MitigationType.IsNull() && !state.MitigationType.IsUnknown() {
		attributes.SetMitigationType(state.MitigationType.ValueString())
	}
	if !state.NotificationType.IsNull() && !state.NotificationType.IsUnknown() {
		attributes.SetNotificationType(state.NotificationType.ValueString())
	}
	if !state.NotificationFrequency.IsNull() && !state.NotificationFrequency.IsUnknown() {
		attributes.SetNotificationFrequency(state.NotificationFrequency.ValueString())
	}

	for _, param := range []struct {
		value  jsontypes.Normalized
		name   string
		setter func(map[string]interface{})
	}{
		{state.DetectionParameters, "detection_parameters", attributes.SetDetectionParameters},
		{state.MitigationParameters, "mitigation_parameters", attributes.SetMitigationParameters},
		{state.NotificationParameters, "notification_parameters", attributes.SetNotificationParameters},
	} {
		if param.value.IsNull() || param.value.IsUnknown() {
			continue
		}
		var decoded map[string]interface{}
		if err := json.Unmarshal([]byte(param.value.ValueString()), &decoded); err != nil {
			diags.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("%s must be a valid JSON object", param.name)))
			return nil
		}
		param.setter(decoded)
	}

	req := datadogV2.NewGovernanceControlUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewGovernanceControlUpdateDataWithDefaults()
	req.Data.SetAttributes(*attributes)
	return req
}

func (r *governanceControlResource) updateState(state *governanceControlModel, resp *datadogV2.GovernanceControlResponse) error {
	attributes := resp.Data.GetAttributes()

	state.ID = types.StringValue(attributes.GetDetectionType())
	state.DetectionType = types.StringValue(attributes.GetDetectionType())
	state.Name = types.StringValue(attributes.GetName())
	state.DetectionFrequency = types.StringValue(attributes.GetDetectionFrequency())
	state.MitigationType = types.StringValue(attributes.GetMitigationType())
	state.NotificationType = types.StringValue(attributes.GetNotificationType())
	state.NotificationFrequency = types.StringValue(attributes.GetNotificationFrequency())

	for _, param := range []struct {
		value  map[string]interface{}
		target *jsontypes.Normalized
	}{
		{attributes.GetDetectionParameters(), &state.DetectionParameters},
		{attributes.GetMitigationParameters(), &state.MitigationParameters},
		{attributes.GetNotificationParameters(), &state.NotificationParameters},
	} {
		encoded, err := json.Marshal(param.value)
		if err != nil {
			return fmt.Errorf("error marshaling governance control parameters: %w", err)
		}
		*param.target = jsontypes.NewNormalizedValue(string(encoded))
	}

	return nil
}
