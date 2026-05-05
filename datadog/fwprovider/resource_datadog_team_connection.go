package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &teamConnectionResource{}
	_ resource.ResourceWithImportState = &teamConnectionResource{}
)

type teamConnectionResource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type teamConnectionModel struct {
	ID            types.String       `tfsdk:"id"`
	Team          *teamConnectionRef `tfsdk:"team"`
	ConnectedTeam *teamConnectionRef `tfsdk:"connected_team"`
	Source        types.String       `tfsdk:"source"`
}

type teamConnectionRef struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func NewTeamConnectionResource() resource.Resource {
	return &teamConnectionResource{}
}

func (r *teamConnectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (r *teamConnectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "team_connection"
}

func (r *teamConnectionResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Team Connection resource. This can be used to create and manage connections between a Datadog team and an external team (e.g. GitHub).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The source of the connection (e.g. github).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"team": schema.SingleNestedBlock{
				Description: "The Datadog team reference.",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:    true,
						Description: "The ID of the Datadog team.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"type": schema.StringAttribute{
						Required:    true,
						Description: "The resource type of the Datadog team.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("team"),
						},
					},
				},
			},
			"connected_team": schema.SingleNestedBlock{
				Description: "The external connected team reference (e.g. a GitHub team).",
				Validators: []validator.Object{
					objectvalidator.IsRequired(),
				},
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Required:    true,
						Description: "The ID of the external connected team.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"type": schema.StringAttribute{
						Required:    true,
						Description: "The resource type of the external connected team.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("github_team"),
						},
					},
				},
			},
		},
	}
}

func (r *teamConnectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *teamConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state teamConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	opts := datadogV2.NewListTeamConnectionsOptionalParameters().WithFilterConnectionIds([]string{id})

	resp, httpResp, err := r.Api.ListTeamConnections(r.Auth, *opts)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving TeamConnection"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := resp.GetData()
	if len(data) == 0 {
		response.State.RemoveResource(ctx)
		return
	}

	r.updateState(&state, &data[0])
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state teamConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body := r.buildCreateRequestBody(&state)

	resp, _, err := r.Api.CreateTeamConnections(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error creating TeamConnection"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	data := resp.GetData()
	if len(data) == 0 {
		response.Diagnostics.AddError("empty response", "no team connection returned in create response")
		return
	}

	r.updateState(&state, &data[0])
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *teamConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("Update not supported for this resource", "Team connections are immutable. All fields require replacement.")
}

func (r *teamConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state teamConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := state.ID.ValueString()
	deleteItem := datadogV2.NewTeamConnectionDeleteRequestDataItem(id, datadogV2.TEAMCONNECTIONTYPE_TEAM_CONNECTION)
	body := datadogV2.NewTeamConnectionDeleteRequest([]datadogV2.TeamConnectionDeleteRequestDataItem{*deleteItem})

	httpResp, err := r.Api.DeleteTeamConnections(r.Auth, *body)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting TeamConnection"))
		return
	}
}

func (r *teamConnectionResource) updateState(state *teamConnectionModel, conn *datadogV2.TeamConnection) {
	state.ID = types.StringValue(conn.GetId())

	if attrs, ok := conn.GetAttributesOk(); ok {
		if source, ok := attrs.GetSourceOk(); ok {
			state.Source = types.StringValue(*source)
		}
	}

	if rels, ok := conn.GetRelationshipsOk(); ok {
		if team, ok := rels.GetTeamOk(); ok {
			if data, ok := team.GetDataOk(); ok {
				if state.Team == nil {
					state.Team = &teamConnectionRef{}
				}
				state.Team.ID = types.StringValue(data.GetId())
				state.Team.Type = types.StringValue(string(data.GetType()))
			}
		}
		if connTeam, ok := rels.GetConnectedTeamOk(); ok {
			if data, ok := connTeam.GetDataOk(); ok {
				if state.ConnectedTeam == nil {
					state.ConnectedTeam = &teamConnectionRef{}
				}
				state.ConnectedTeam.ID = types.StringValue(data.GetId())
				state.ConnectedTeam.Type = types.StringValue(string(data.GetType()))
			}
		}
	}
}

func (r *teamConnectionResource) buildCreateRequestBody(state *teamConnectionModel) *datadogV2.TeamConnectionCreateRequest {
	createData := datadogV2.NewTeamConnectionCreateData(datadogV2.TEAMCONNECTIONTYPE_TEAM_CONNECTION)

	attrs := datadogV2.NewTeamConnectionAttributes()
	if !state.Source.IsNull() && !state.Source.IsUnknown() {
		attrs.SetSource(state.Source.ValueString())
	}
	createData.SetAttributes(*attrs)

	teamRefData := datadogV2.NewTeamRefData(state.Team.ID.ValueString(), datadogV2.TeamRefDataType(state.Team.Type.ValueString()))
	teamRef := datadogV2.NewTeamRef()
	teamRef.SetData(*teamRefData)

	connTeamRefData := datadogV2.NewConnectedTeamRefData(state.ConnectedTeam.ID.ValueString(), datadogV2.ConnectedTeamRefDataType(state.ConnectedTeam.Type.ValueString()))
	connTeamRef := datadogV2.NewConnectedTeamRef()
	connTeamRef.SetData(*connTeamRefData)

	rels := datadogV2.NewTeamConnectionRelationships()
	rels.SetTeam(*teamRef)
	rels.SetConnectedTeam(*connTeamRef)
	createData.SetRelationships(*rels)

	return datadogV2.NewTeamConnectionCreateRequest([]datadogV2.TeamConnectionCreateData{*createData})
}
