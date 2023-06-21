package fwprovider

import (
	"context"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.ResourceWithConfigure   = &teamMembershipResource{}
	_ resource.ResourceWithImportState = &teamMembershipResource{}
)

type teamMembershipResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type TeamMembershipModel struct {
	ID     types.String `tfsdk:"id"`
	TeamId types.String `tfsdk:"team_id"`
	UserId types.String `tfsdk:"user_id"`
	Role   types.String `tfsdk:"role"`
}

func NewTeamMembershipResource() resource.Resource {
	return &teamMembershipResource{}
}

func (r *teamMembershipResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamMembershipResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_membership"
}

func (r *teamMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog TeamMembership resource. This can be used to create and manage Datadog team_membership.",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the team the team membership is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Optional:    true,
				Description: "The user's role within the team.",
				Validators: []validator.String{
					validators.NewEnumValidator[validator.String](datadogV2.NewUserTeamRoleFromValue),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *teamMembershipResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving team_id or user_id from given ID", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("team_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("user_id"), result[1])...)
}

func (r *teamMembershipResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()
	pageSize := int64(100)
	pageNumber := int64(0)

	var userTeams []datadogV2.UserTeam
	for {
		resp, _, err := r.Api.GetTeamMemberships(r.Auth, teamId, *datadogV2.NewGetTeamMembershipsOptionalParameters().
			WithPageSize(pageSize).
			WithPageNumber(pageNumber))
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamMembership"))
			return
		}
		if err := utils.CheckForUnparsed(resp); err != nil {
			response.Diagnostics.AddError("response contains unparsedObject", err.Error())
			return
		}

		userTeams = append(userTeams, resp.GetData()...)
		if len(resp.GetData()) < 100 {
			break
		}

		pageNumber++
	}

	for _, userTeam := range userTeams {
		// we use team_id:user_id format for importing.
		// Hence, we need to check wether resource id or user id matches config.
		if userTeam.GetId() == state.ID.ValueString() || state.UserId.ValueString() == userTeam.Relationships.User.Data.GetId() {
			r.updateStateFromTeamResponse(ctx, &state, &userTeam)
			break
		}
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamMembershipResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	body, diags := r.buildTeamMembershipRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateTeamMembership(r.Auth, teamId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamMembership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateStateFromTeamResponse(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamMembershipResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()
	userId := state.UserId.ValueString()

	body, diags := r.buildTeamMembershipUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateTeamMembership(r.Auth, teamId, userId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamMembership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateStateFromTeamResponse(ctx, &state, resp.Data)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamMembershipResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()
	userId := state.UserId.ValueString()

	httpResp, err := r.Api.DeleteTeamMembership(r.Auth, teamId, userId)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting team_membership"))
		return
	}
}

func (r *teamMembershipResource) updateStateFromTeamResponse(ctx context.Context, state *TeamMembershipModel, resp *datadogV2.UserTeam) {
	state.ID = types.StringValue(resp.GetId())

	if role, ok := resp.Attributes.GetRoleOk(); ok {
		if role == nil {
			state.Role = types.StringNull()
		} else {
			state.Role = types.StringValue(string(*role))
		}
	}
}

func (r *teamMembershipResource) buildTeamMembershipRequestBody(ctx context.Context, state *TeamMembershipModel) (*datadogV2.UserTeamRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUserTeamAttributesWithDefaults()

	if !state.Role.IsNull() {
		role, _ := datadogV2.NewUserTeamRoleFromValue(state.Role.ValueString())
		attributes.SetRole(*role)
	}

	relationships := datadogV2.NewUserTeamRelationshipsWithDefaults()
	relationships.User = &datadogV2.RelationshipToUserTeamUser{
		Data: *datadogV2.NewRelationshipToUserTeamUserDataWithDefaults(),
	}
	relationships.User.Data.Id = state.UserId.ValueString()

	req := datadogV2.NewUserTeamRequestWithDefaults()
	req.Data = *datadogV2.NewUserTeamCreateWithDefaults()
	req.Data.SetAttributes(*attributes)
	req.Data.SetRelationships(*relationships)

	return req, diags
}

func (r *teamMembershipResource) buildTeamMembershipUpdateRequestBody(ctx context.Context, state *TeamMembershipModel) (*datadogV2.UserTeamUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUserTeamAttributesWithDefaults()

	if !state.Role.IsNull() {
		role, _ := datadogV2.NewUserTeamRoleFromValue(state.Role.ValueString())
		attributes.SetRole(*role)
	}

	req := datadogV2.NewUserTeamUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewUserTeamUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
