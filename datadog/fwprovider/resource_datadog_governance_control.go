package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
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
	Api  *datadogV2.GovernanceConsoleApi
	Auth context.Context
}

type governanceControlModel struct {
	ID                   types.String         `tfsdk:"id"`
	DetectionType        types.String         `tfsdk:"detection_type"`
	Name                 types.String         `tfsdk:"name"`
	DetectionParameters  jsontypes.Normalized `tfsdk:"detection_parameters"`
	MitigationType       types.String         `tfsdk:"mitigation_type"`
	MitigationParameters jsontypes.Normalized `tfsdk:"mitigation_parameters"`
	// NotificationSettings is a types.List (not a native Go slice) because it is Optional+Computed:
	// it can be Unknown during planning, and only attr.Value-backed types can represent that.
	NotificationSettings types.List `tfsdk:"notification_settings"`
}

// governanceControlNotificationTargetModel mirrors datadogV2.ControlNotificationTarget.
type governanceControlNotificationTargetModel struct {
	Type   types.String `tfsdk:"type"`
	Handle types.String `tfsdk:"handle"`
}

// governanceControlNotificationEventSettingModel mirrors datadogV2.ControlNotificationEventSetting.
// It is only ever accessed via NotificationSettings.ElementsAs/ListValueFrom, so its fields do not
// need to individually support Unknown the way the outer list does.
type governanceControlNotificationEventSettingModel struct {
	EventType types.String                               `tfsdk:"event_type"`
	Enabled   types.Bool                                 `tfsdk:"enabled"`
	Targets   []governanceControlNotificationTargetModel `tfsdk:"targets"`
}

var governanceControlNotificationTargetAttrTypes = map[string]attr.Type{
	"type":   types.StringType,
	"handle": types.StringType,
}

var governanceControlNotificationEventSettingAttrTypes = map[string]attr.Type{
	"event_type": types.StringType,
	"enabled":    types.BoolType,
	"targets":    types.ListType{ElemType: types.ObjectType{AttrTypes: governanceControlNotificationTargetAttrTypes}},
}

var governanceControlNotificationEventSettingObjectType = types.ObjectType{AttrTypes: governanceControlNotificationEventSettingAttrTypes}

func NewGovernanceControlResource() resource.Resource {
	return &governanceControlResource{}
}

func (r *governanceControlResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetGovernanceConsoleApiV2()
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
				Computed:    true,
				Description: "Human-readable name of the control.",
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
			"notification_settings": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The notification settings for the control, one entry per event type.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"event_type": schema.StringAttribute{
							Required:    true,
							Description: "The event type the notification settings apply to, such as `new_detection`.",
						},
						"enabled": schema.BoolAttribute{
							Required:    true,
							Description: "Whether notifications are enabled for this event type.",
						},
						"targets": schema.ListNestedAttribute{
							Required:    true,
							Description: "The destinations that receive notifications for this event type.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:    true,
										Description: "The type of notification target: `email`, `slack`, `at_mention`, or `case`.",
									},
									"handle": schema.StringAttribute{
										Required:    true,
										Description: "The handle of the notification target.",
									},
								},
							},
						},
					},
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

	notificationResp, notificationHTTPResp, err := r.Api.GetGovernanceControlNotificationSettings(r.Auth, state.ID.ValueString())
	if err != nil {
		if notificationHTTPResp != nil && notificationHTTPResp.StatusCode == http.StatusNotFound {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving governance control notification settings"))
		return
	}
	if err := utils.CheckForUnparsed(notificationResp); err != nil {
		response.Diagnostics.AddError("datadog_governance_control: response contains unparsedObject", err.Error())
		return
	}

	response.Diagnostics.Append(r.updateState(ctx, &state, &resp, &notificationResp)...)
	if response.Diagnostics.HasError() {
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
	r.upsertGovernanceControl(ctx, &state, &response.Diagnostics)
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
	r.upsertGovernanceControl(ctx, &state, &response.Diagnostics)
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
func (r *governanceControlResource) upsertGovernanceControl(ctx context.Context, state *governanceControlModel, diags *diag.Diagnostics) {
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

	notificationResp := r.upsertNotificationSettings(ctx, state, diags)
	if diags.HasError() {
		return
	}

	diags.Append(r.updateState(ctx, state, &resp, notificationResp)...)
}

func (r *governanceControlResource) buildGovernanceControlUpdateRequest(state *governanceControlModel, diags *diag.Diagnostics) *datadogV2.GovernanceControlUpdateRequest {
	attributes := datadogV2.NewGovernanceControlUpdateAttributesWithDefaults()

	if !state.MitigationType.IsNull() && !state.MitigationType.IsUnknown() {
		attributes.SetMitigationType(state.MitigationType.ValueString())
	}

	for _, param := range []struct {
		value  jsontypes.Normalized
		name   string
		setter func(map[string]interface{})
	}{
		{state.DetectionParameters, "detection_parameters", attributes.SetDetectionParameters},
		{state.MitigationParameters, "mitigation_parameters", attributes.SetMitigationParameters},
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
	req.Data.SetType(datadogV2.GOVERNANCECONTROLRESOURCETYPE_GOVERNANCE_CONTROL)
	req.Data.SetAttributes(*attributes)
	return req
}

// upsertNotificationSettings updates the control's notification settings when configured, then
// always re-fetches the current settings so state reflects the control's actual configuration.
func (r *governanceControlResource) upsertNotificationSettings(ctx context.Context, state *governanceControlModel, diags *diag.Diagnostics) *datadogV2.ControlNotificationSettingsResponse {
	if !state.NotificationSettings.IsNull() && !state.NotificationSettings.IsUnknown() {
		req, reqDiags := r.buildNotificationSettingsUpdateRequest(ctx, state)
		diags.Append(reqDiags...)
		if diags.HasError() {
			return nil
		}

		resp, _, err := r.Api.UpdateGovernanceControlNotificationSettings(r.Auth, state.DetectionType.ValueString(), *req)
		if err != nil {
			diags.Append(utils.FrameworkErrorDiag(err, "error updating governance control notification settings"))
			return nil
		}
		if err := utils.CheckForUnparsed(resp); err != nil {
			diags.AddError("datadog_governance_control: response contains unparsedObject", err.Error())
			return nil
		}
		return &resp
	}

	resp, _, err := r.Api.GetGovernanceControlNotificationSettings(r.Auth, state.DetectionType.ValueString())
	if err != nil {
		diags.Append(utils.FrameworkErrorDiag(err, "error retrieving governance control notification settings"))
		return nil
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		diags.AddError("datadog_governance_control: response contains unparsedObject", err.Error())
		return nil
	}
	return &resp
}

func (r *governanceControlResource) buildNotificationSettingsUpdateRequest(ctx context.Context, state *governanceControlModel) (*datadogV2.ControlNotificationSettingsUpdateRequest, diag.Diagnostics) {
	var settings []governanceControlNotificationEventSettingModel
	diags := state.NotificationSettings.ElementsAs(ctx, &settings, false)
	if diags.HasError() {
		return nil, diags
	}

	eventSettings := make([]datadogV2.ControlNotificationEventSetting, 0, len(settings))
	for _, es := range settings {
		targets := make([]datadogV2.ControlNotificationTarget, 0, len(es.Targets))
		for _, t := range es.Targets {
			targets = append(targets, datadogV2.ControlNotificationTarget{
				Type:   datadogV2.ControlNotificationTargetType(t.Type.ValueString()),
				Handle: t.Handle.ValueString(),
			})
		}
		eventSettings = append(eventSettings, datadogV2.ControlNotificationEventSetting{
			EventType: es.EventType.ValueString(),
			Enabled:   es.Enabled.ValueBool(),
			Targets:   targets,
		})
	}

	attributes := datadogV2.NewControlNotificationSettingsUpdateAttributesWithDefaults()
	attributes.SetEventSettings(eventSettings)

	req := datadogV2.NewControlNotificationSettingsUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewControlNotificationSettingsUpdateDataWithDefaults()
	req.Data.SetType(datadogV2.CONTROLNOTIFICATIONSETTINGSRESOURCETYPE_CONTROL_NOTIFICATION_SETTINGS)
	req.Data.SetAttributes(*attributes)
	return req, diags
}

func (r *governanceControlResource) updateState(ctx context.Context, state *governanceControlModel, resp *datadogV2.GovernanceControlResponse, notificationResp *datadogV2.ControlNotificationSettingsResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	attributes := resp.Data.GetAttributes()

	state.ID = types.StringValue(resp.Data.GetId())
	state.DetectionType = types.StringValue(resp.Data.GetId())
	state.Name = types.StringValue(attributes.GetName())
	state.MitigationType = types.StringValue(attributes.GetMitigationType())

	for _, param := range []struct {
		value  map[string]interface{}
		target *jsontypes.Normalized
	}{
		{attributes.GetDetectionParameters(), &state.DetectionParameters},
		{attributes.GetMitigationParameters(), &state.MitigationParameters},
	} {
		encoded, err := json.Marshal(param.value)
		if err != nil {
			diags.Append(utils.FrameworkErrorDiag(err, "error marshaling governance control parameters"))
			return diags
		}
		*param.target = jsontypes.NewNormalizedValue(string(encoded))
	}

	notificationAttributes := notificationResp.Data.GetAttributes()
	eventSettings := notificationAttributes.GetEventSettings()
	settings := make([]governanceControlNotificationEventSettingModel, 0, len(eventSettings))
	for _, es := range eventSettings {
		targets := make([]governanceControlNotificationTargetModel, 0, len(es.Targets))
		for _, t := range es.Targets {
			targets = append(targets, governanceControlNotificationTargetModel{
				Type:   types.StringValue(string(t.Type)),
				Handle: types.StringValue(t.Handle),
			})
		}
		settings = append(settings, governanceControlNotificationEventSettingModel{
			EventType: types.StringValue(es.EventType),
			Enabled:   types.BoolValue(es.Enabled),
			Targets:   targets,
		})
	}

	notificationSettings, notificationDiags := types.ListValueFrom(ctx, governanceControlNotificationEventSettingObjectType, settings)
	diags.Append(notificationDiags...)
	if diags.HasError() {
		return diags
	}
	state.NotificationSettings = notificationSettings

	return diags
}
