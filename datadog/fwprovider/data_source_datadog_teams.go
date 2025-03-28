package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogTeamsDataSource{}
)

type TeamModel struct {
	Description types.String `tfsdk:"description"`
	Handle      types.String `tfsdk:"handle"`
	ID          types.String `tfsdk:"id"`
	LinkCount   types.Int64  `tfsdk:"link_count"`
	Name        types.String `tfsdk:"name"`
	Summary     types.String `tfsdk:"summary"`
	UserCount   types.Int64  `tfsdk:"user_count"`
}

type datadogTeamsDataSourceModel struct {
	// Query Parameters
	FilterKeyword types.String `tfsdk:"filter_keyword"`
	FilterMe      types.Bool   `tfsdk:"filter_me"`

	// Results
	ID    types.String `tfsdk:"id"`
	Teams []*TeamModel `tfsdk:"teams"`
}

type datadogTeamsDataSource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

func NewDatadogTeamsDataSource() datasource.DataSource {
	return &datadogTeamsDataSource{}
}

func (d *datadogTeamsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogTeamsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "teams"
}

func (d *datadogTeamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing teams for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"filter_keyword": schema.StringAttribute{
				Optional:    true,
				Description: "Search query. Can be team name, team handle, or email of team member.",
			},
			"filter_me": schema.BoolAttribute{
				Optional:    true,
				Description: "When true, only returns teams the current user belongs to.",
			},
		},
		Blocks: map[string]schema.Block{
			"teams": schema.ListNestedBlock{
				Description: "List of teams",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "Free-form markdown description/content for the team's homepage.",
						},
						"handle": schema.StringAttribute{
							Computed:    true,
							Description: "The team's handle.",
						},
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The team's identifier.",
						},
						"link_count": schema.Int64Attribute{
							Computed:    true,
							Description: "The number of links belonging to the team.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the team.",
						},
						"summary": schema.StringAttribute{
							Computed:    true,
							Description: "A brief summary of the team, derived from the `description`.",
						},
						"user_count": schema.Int64Attribute{
							Computed:    true,
							Description: "The number of users belonging to the team.",
						},
					},
				},
			},
		},
	}
}

func (d *datadogTeamsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogTeamsDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	var optionalParams datadogV2.ListTeamsOptionalParameters
	if !state.FilterKeyword.IsNull() {
		optionalParams.FilterKeyword = state.FilterKeyword.ValueStringPointer()
	}
	if !state.FilterMe.IsNull() {
		optionalParams.FilterMe = state.FilterMe.ValueBoolPointer()
	}

	optionalParams.WithSort(datadogV2.LISTTEAMSSORT_NAME)

	var teams []datadogV2.Team
	result, _ := d.Api.ListTeamsWithPagination(d.Auth, optionalParams)
	for paginationResult := range result {
		if paginationResult.Error != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(paginationResult.Error, "Error when calling `ListTeamsWithPagination`"))
			return
		}

		teams = append(teams, paginationResult.Item)
	}

	d.updateState(&state, &teams)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogTeamsDataSource) updateState(state *datadogTeamsDataSourceModel, teamsData *[]datadogV2.Team) {

	teams := make([]*TeamModel, 0, len(*teamsData))
	for _, team := range *teamsData {
		t := TeamModel{
			Description: types.StringValue(team.Attributes.GetDescription()),
			Handle:      types.StringValue(team.Attributes.GetHandle()),
			ID:          types.StringValue(team.GetId()),
			LinkCount:   types.Int64Value(int64(team.Attributes.GetLinkCount())),
			Name:        types.StringValue(team.Attributes.GetName()),
			Summary:     types.StringValue(team.Attributes.GetSummary()),
			UserCount:   types.Int64Value(int64(team.Attributes.GetUserCount())),
		}

		teams = append(teams, &t)
	}

	hashingData := fmt.Sprintf("%s:%t", state.FilterKeyword.ValueString(), state.FilterMe.ValueBool())

	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Teams = teams
}
