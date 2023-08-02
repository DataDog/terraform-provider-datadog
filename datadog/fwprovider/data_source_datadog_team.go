package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogTeamDataSource{}
)

func NewDatadogTeamDataSource() datasource.DataSource {
	return &datadogTeamDataSource{}
}

type datadogTeamDataSourceModel struct {
	// Query Parameters
	TeamID        types.String `tfsdk:"team_id"`
	FilterKeyword types.String `tfsdk:"filter_keyword"`
	// Results
	ID          types.String `tfsdk:"id"`
	Description types.String `tfsdk:"description"`
	Handle      types.String `tfsdk:"handle"`
	LinkCount   types.Int64  `tfsdk:"link_count"`
	Summary     types.String `tfsdk:"summary"`
	UserCount   types.Int64  `tfsdk:"user_count"`
	Name        types.String `tfsdk:"name"`
}

type datadogTeamDataSource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

func (r *datadogTeamDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogTeamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "team"
}

func (d *datadogTeamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog team.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"team_id": schema.StringAttribute{
				Description: "The team's identifier.",
				Optional:    true,
				Computed:    true,
			},
			"filter_keyword": schema.StringAttribute{
				Description: "Search query. Can be team name, team handle, or email of team member.",
				Optional:    true,
			},
			// Computed values
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Free-form markdown description/content for the team's homepage.",
			},
			"handle": schema.StringAttribute{
				Computed:    true,
				Description: "The team's handle.",
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
				Computed:    true,
				Description: "The name of the team.",
			},
		},
	}

}

func (d *datadogTeamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogTeamDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.TeamID.IsNull() {
		teamID := state.TeamID.ValueString()
		ddResp, _, err := d.Api.GetTeam(d.Auth, teamID)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog team"))
			return
		}

		d.updateState(&state, ddResp.Data)
	} else if !state.FilterKeyword.IsNull() {
		filterKeyword := state.FilterKeyword.ValueString()
		optionalParams := datadogV2.ListTeamsOptionalParameters{
			FilterKeyword: &filterKeyword,
		}

		ddResp, _, err := d.Api.ListTeams(d.Auth, optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datadog teams"))
			return
		}

		if len(ddResp.Data) > 1 {
			resp.Diagnostics.AddError("filter keyword returned more than one result, use more specific search criteria", "")
			return
		}
		if len(ddResp.Data) == 0 {
			resp.Diagnostics.AddError("filter keyword returned no result", "")
			return
		}

		d.updateStateFromListResponse(&state, &ddResp.Data[0])
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogTeamDataSource) updateState(state *datadogTeamDataSourceModel, teamData *datadogV2.Team) {
	state.ID = types.StringValue(teamData.GetId())
	attributes := teamData.GetAttributes()

	state.Description = types.StringValue(attributes.GetDescription())
	state.Handle = types.StringValue(attributes.GetHandle())
	state.LinkCount = types.Int64Value(int64(attributes.GetLinkCount()))
	state.Name = types.StringValue(attributes.GetName())
	state.UserCount = types.Int64Value(int64(attributes.GetUserCount()))
	state.Summary = types.StringValue(attributes.GetSummary())
}

func (r *datadogTeamDataSource) updateStateFromListResponse(state *datadogTeamDataSourceModel, teamData *datadogV2.Team) {
	state.ID = types.StringValue(teamData.GetId())
	state.TeamID = types.StringValue(teamData.GetId())

	attributes := teamData.GetAttributes()
	state.Description = types.StringValue(attributes.GetDescription())
	state.Handle = types.StringValue(attributes.GetHandle())
	state.LinkCount = types.Int64Value(int64(attributes.GetLinkCount()))
	state.Name = types.StringValue(attributes.GetName())
	state.UserCount = types.Int64Value(int64(attributes.GetUserCount()))
	state.Summary = types.StringValue(attributes.GetSummary())
}
