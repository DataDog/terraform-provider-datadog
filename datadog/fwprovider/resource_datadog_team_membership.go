package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &TeamMembershipResource{}
	_ resource.ResourceWithImportState = &TeamMembershipResource{}
)

type TeamMembershipResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type TeamMembershipModel struct {
	ID     types.String `tfsdk:"id"`
	TeamId types.String `tfsdk:"team_id"`
	Role   types.String `tfsdk:"role"`
}

func NewTeamMembershipResource() resource.Resource {
	return &TeamMembershipResource{}
}

func (r *TeamMembershipResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError("Unexpected Resource Configure Type", "")
		return
	}

	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *TeamMembershipResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "team_membership"
}

func (r *TeamMembershipResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog TeamMembership resource. This can be used to create and manage Datadog team_membership.",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the team the team membership is associated with.",
			},
			"user_id": schema.StringAttribute{
				Required:    true,
				Description: "UPDATE ME",
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The user's role within the team.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *TeamMembershipResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *TeamMembershipResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	teamId := state.TeamId.ValueString()
	pageSize := state.PageSize.ValueInt64()
	pageNumber := state.PageNumber.ValueInt64()
	sort := state.Sort.ValueString()
	filterKeyword := state.FilterKeyword.ValueString()
	resp, httpResp, err := r.Api.GetTeamMemberships(r.Auth, teamId, pageSize, pageNumber, sort, filterKeyword)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamMembership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *TeamMembershipResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
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
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *TeamMembershipResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()

	body, diags := r.buildTeamMembershipUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateTeamMembership(r.Auth, teamId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamMembership"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *TeamMembershipResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state TeamMembershipModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteTeamMembership(r.Auth, teamId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting team_membership"))
		return
	}
}

func (r *TeamMembershipResource) updateState(ctx context.Context, state *TeamMembershipModel, resp *datadogV2.UserTeamsResponse) {
	state.ID = types.StringValue(resp.GetUserId())

	if data, ok := resp.GetDataOk(); ok && len(*data) > 0 {
		state.Data = []*DataModel{}
		for _, dataDd := range *data {
			dataTfItem := DataModel{}
			if attributes, ok := dataDd.GetAttributesOk(); ok {
				attributesTf := AttributesModel{}
				if role, ok := attributes.GetRoleOk(); ok {
					attributesTf.Role = types.StringValue(*role)
				}

				dataTfItem.Attributes = &attributesTf
			}
			if id, ok := dataDd.GetIdOk(); ok {
				dataTfItem.Id = types.StringValue(*id)
			}
			if relationships, ok := dataDd.GetRelationshipsOk(); ok {
				relationshipsTf := RelationshipsModel{}
				if user, ok := relationships.GetUserOk(); ok {
					userTf := UserModel{}
					if data, ok := user.GetDataOk(); ok {
						dataTf := DataModel{}
						if id, ok := data.GetIdOk(); ok {
							dataTf.Id = types.StringValue(*id)
						}
						if typeVar, ok := data.GetTypeOk(); ok {
							dataTf.Type = types.StringValue(*typeVar)
						}

						userTf.Data = &dataTf
					}

					relationshipsTf.User = &userTf
				}

				dataTfItem.Relationships = &relationshipsTf
			}
			if typeVar, ok := dataDd.GetTypeOk(); ok {
				dataTfItem.Type = types.StringValue(*typeVar)
			}

			state.Data = append(state.Data, &dataTfItem)
		}
	}
}

func (r *TeamMembershipResource) buildTeamMembershipRequestBody(ctx context.Context, state *TeamMembershipModel) (*datadogV2.UserTeamRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUserTeamAttributesWithDefaults()

	if !state.Role.IsNull() {
		attributes.SetRole(state.Role.ValueString())
	}

	req := datadogV2.NewUserTeamRequestWithDefaults()
	req.Data = *datadogV2.NewUserTeamCreateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *TeamMembershipResource) buildTeamMembershipUpdateRequestBody(ctx context.Context, state *TeamMembershipModel) (*datadogV2.UserTeamUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewUserTeamAttributesWithDefaults()

	if !state.Role.IsNull() {
		attributes.SetRole(state.Role.ValueString())
	}

	req := datadogV2.NewUserTeamUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewUserTeamUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
