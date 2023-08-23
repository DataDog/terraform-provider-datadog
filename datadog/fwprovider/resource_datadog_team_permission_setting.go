package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure = &teamPermissionSettingResource{}
)

type teamPermissionSettingResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamPermissionSettingModel struct {
	ID     types.String `tfsdk:"id"`
	TeamId types.String `tfsdk:"team_id"`
	Action types.String `tfsdk:"action"`
	Value  types.String `tfsdk:"value"`
}

func NewTeamPermissionSettingResource() resource.Resource {
	return &teamPermissionSettingResource{}
}

func (r *teamPermissionSettingResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamPermissionSettingResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_permission_setting"
}

func (r *teamPermissionSettingResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog TeamPermissionSetting resource. This can be used to manage Datadog team_permission_setting.",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the team the team permission setting is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"action": schema.StringAttribute{
				Required:    true,
				Description: "The identifier for the action.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingSerializerActionFromValue),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Description: "The action value.",
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewTeamPermissionSettingValueFromValue),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *teamPermissionSettingResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamPermissionSettingModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	permissions, httpresp, err := r.Api.GetTeamPermissionSettings(r.Auth, state.TeamId.ValueString())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error getting team permission setting"))
		return
	}

	found := false
	for _, permission := range permissions.Data {
		if permission.Id == state.ID.ValueString() {
			r.updateState(ctx, &state, &permission)
			found = true
		}
	}

	if !found {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("error getting team permission setting with id %s", state.ID.ValueString())))
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamPermissionSettingResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamPermissionSettingModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	reqBody := r.buildTeamMembershipPermissionSettingRequestBody(&state)

	resp, _, err := r.Api.UpdateTeamPermissionSetting(r.Auth, state.TeamId.ValueString(), state.Action.ValueString(), *reqBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating team permission setting"))
		return
	}

	r.updateState(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamPermissionSettingResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state teamPermissionSettingModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	reqBody := r.buildTeamMembershipPermissionSettingRequestBody(&state)

	resp, _, err := r.Api.UpdateTeamPermissionSetting(r.Auth, state.TeamId.ValueString(), state.Action.ValueString(), *reqBody)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating team permission setting"))
		return
	}

	r.updateState(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamPermissionSettingResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddWarning("resource cannot be deleted", "team permission setting cannot be deleted and is only removed from terraform state")
}

func (r *teamPermissionSettingResource) updateState(ctx context.Context, state *teamPermissionSettingModel, resp *datadogV2.TeamPermissionSetting) {
	state.ID = types.StringValue(resp.GetId())
	state.Action = types.StringValue(string(resp.Attributes.GetAction()))
	state.Value = types.StringValue(string(resp.Attributes.GetValue()))

}

func (r *teamPermissionSettingResource) buildTeamMembershipPermissionSettingRequestBody(state *teamPermissionSettingModel) *datadogV2.TeamPermissionSettingUpdateRequest {
	attributes := datadogV2.NewTeamPermissionSettingUpdateAttributesWithDefaults()

	attributes.SetValue(datadogV2.TeamPermissionSettingValue(state.Value.ValueString()))

	req := datadogV2.NewTeamPermissionSettingUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewTeamPermissionSettingUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}
