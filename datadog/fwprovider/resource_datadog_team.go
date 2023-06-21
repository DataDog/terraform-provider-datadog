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
	_ resource.ResourceWithConfigure   = &teamResource{}
	_ resource.ResourceWithImportState = &teamResource{}
)

type teamResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamModel struct {
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Handle      types.String `tfsdk:"handle"`
	LinkCount   types.Int64  `tfsdk:"link_count"`
	Summary     types.String `tfsdk:"summary"`
	UserCount   types.Int64  `tfsdk:"user_count"`
	Name        types.String `tfsdk:"name"`
}

func NewTeamResource() resource.Resource {
	return &teamResource{}
}

func (r *teamResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team"
}

func (r *teamResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Team resource. This can be used to create and manage Datadog team.",
		Attributes: map[string]schema.Attribute{
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Free-form markdown description/content for the team's homepage.",
			},
			"handle": schema.StringAttribute{
				Required:    true,
				Description: "The team's identifier",
			},
			"link_count": schema.Int64Attribute{
				Description: "The number of links belonging to the team.",
				Computed:    true,
			},
			"summary": schema.StringAttribute{
				Description: "A brief summary of the team, derived from the `description`.",
				Computed:    true,
			},
			"user_count": schema.Int64Attribute{
				Description: "The number of users belonging to the team.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the team.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *teamResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *teamResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetTeam(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Team"))
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

func (r *teamResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildTeamRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateTeam(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Team"))
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

func (r *teamResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state teamModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	body, diags := r.buildTeamUpdateRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateTeam(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving Team"))
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

func (r *teamResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteTeam(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting team"))
		return
	}
}

func (r *teamResource) updateState(ctx context.Context, state *teamModel, resp *datadogV2.TeamResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if description, ok := attributes.GetDescriptionOk(); ok && description != nil {
		state.Description = types.StringValue(*description)
	}

	if handle, ok := attributes.GetHandleOk(); ok {
		state.Handle = types.StringValue(*handle)
	}

	if linkCount, ok := attributes.GetLinkCountOk(); ok {
		state.LinkCount = types.Int64Value(int64(*linkCount))
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if summary, ok := attributes.GetSummaryOk(); ok && summary != nil {
		state.Summary = types.StringValue(*summary)
	} else {
		state.Summary = types.StringNull()
	}

	if userCount, ok := attributes.GetUserCountOk(); ok {
		state.UserCount = types.Int64Value(int64(*userCount))
	}
}

func (r *teamResource) buildTeamRequestBody(ctx context.Context, state *teamModel) (*datadogV2.TeamCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewTeamCreateAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}
	attributes.SetHandle(state.Handle.ValueString())

	attributes.SetName(state.Name.ValueString())

	req := datadogV2.NewTeamCreateRequestWithDefaults()
	req.Data = *datadogV2.NewTeamCreateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}

func (r *teamResource) buildTeamUpdateRequestBody(ctx context.Context, state *teamModel) (*datadogV2.TeamUpdateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewTeamUpdateAttributesWithDefaults()

	if !state.Description.IsNull() {
		attributes.SetDescription(state.Description.ValueString())
	}

	attributes.SetHandle(state.Handle.ValueString())

	attributes.SetName(state.Name.ValueString())

	req := datadogV2.NewTeamUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewTeamUpdateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
