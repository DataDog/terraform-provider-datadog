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
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &teamLinkResource{}
	_ resource.ResourceWithImportState = &teamLinkResource{}
)

type teamLinkResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamLinkModel struct {
	ID       types.String `tfsdk:"id"`
	TeamId   types.String `tfsdk:"team_id"`
	Label    types.String `tfsdk:"label"`
	Position types.Int64  `tfsdk:"position"`
	Url      types.String `tfsdk:"url"`
}

func NewTeamLinkResource() resource.Resource {
	return &teamLinkResource{}
}

func (r *teamLinkResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamLinkResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_link"
}

func (r *teamLinkResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog TeamLink resource. This can be used to create and manage Datadog team_link.",
		Attributes: map[string]schema.Attribute{
			"team_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the team the link is associated with.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"label": schema.StringAttribute{
				Required:    true,
				Description: "The link's label.",
			},
			"position": schema.Int64Attribute{
				Computed:    true,
				Optional:    true,
				Description: "The link's position, used to sort links for the team.",
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL for the link.",
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (r *teamLinkResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	result := strings.SplitN(request.ID, ":", 2)
	if len(result) != 2 {
		response.Diagnostics.AddError("error retrieving team_id or resource id from given ID", "")
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("team_id"), result[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("id"), result[1])...)
}

func (r *teamLinkResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()
	resp, httpResp, err := r.Api.GetTeamLink(r.Auth, teamId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamLink"))
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

func (r *teamLinkResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamLinkModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	body, diags := r.buildTeamLinkRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateTeamLink(r.Auth, teamId, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamLink"))
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

func (r *teamLinkResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state teamLinkModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()

	body, diags := r.buildTeamLinkRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateTeamLink(r.Auth, teamId, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamLink"))
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

func (r *teamLinkResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamLinkModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	teamId := state.TeamId.ValueString()

	id := state.ID.ValueString()

	httpResp, err := r.Api.DeleteTeamLink(r.Auth, teamId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting team_link"))
		return
	}
}

func (r *teamLinkResource) updateState(ctx context.Context, state *teamLinkModel, resp *datadogV2.TeamLinkResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if label, ok := attributes.GetLabelOk(); ok {
		state.Label = types.StringValue(*label)
	}

	if position, ok := attributes.GetPositionOk(); ok {
		state.Position = types.Int64Value(int64(*position))
	}

	if teamId, ok := attributes.GetTeamIdOk(); ok {
		state.TeamId = types.StringValue(*teamId)
	}

	if url, ok := attributes.GetUrlOk(); ok {
		state.Url = types.StringValue(*url)
	}
}

func (r *teamLinkResource) buildTeamLinkRequestBody(ctx context.Context, state *teamLinkModel) (*datadogV2.TeamLinkCreateRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	attributes := datadogV2.NewTeamLinkAttributesWithDefaults()

	attributes.SetLabel(state.Label.ValueString())
	if !state.Position.IsNull() {
		attributes.SetPosition(int32(state.Position.ValueInt64()))
	}
	if !state.TeamId.IsNull() {
		attributes.SetTeamId(state.TeamId.ValueString())
	}
	attributes.SetUrl(state.Url.ValueString())

	req := datadogV2.NewTeamLinkCreateRequestWithDefaults()
	req.Data = *datadogV2.NewTeamLinkCreateWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
