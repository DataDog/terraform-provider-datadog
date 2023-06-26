package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogTeamMembershipsDataSource{}
)

type datadogTeamMembershipsDataSourceModel struct {
	// Query Parameters
	TeamID        types.String `tfsdk:"team_id"`
	FilterKeyword types.String `tfsdk:"filter_keyword"`
	// Results
	ID              types.String           `tfsdk:"id"`
	TeamMemberships []*TeamMembershipModel `tfsdk:"team_memberships"`
}

func NewDatadogTeamMembershipsDataSource() datasource.DataSource {
	return &datadogTeamMembershipsDataSource{}
}

type datadogTeamMembershipsDataSource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

func (r *datadogTeamMembershipsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogTeamMembershipsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "team_memberships"
}

func (d *datadogTeamMembershipsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Datadog team memberships.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"team_id": schema.StringAttribute{
				Description: "The team's identifier.",
				Required:    true,
			},
			"filter_keyword": schema.StringAttribute{
				Description: "Search query, can be user email or name.",
				Optional:    true,
			},
			// Computed values
			"team_memberships": schema.ListAttribute{
				Computed:    true,
				Description: "List of team memberships.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"team_id": types.StringType,
						"user_id": types.StringType,
						"role":    types.StringType,
						"id":      types.StringType,
					},
				},
			},
		},
	}

}

func (d *datadogTeamMembershipsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogTeamMembershipsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var optionalParams datadogV2.GetTeamMembershipsOptionalParameters
	teamID := state.TeamID.ValueString()

	if !state.FilterKeyword.IsNull() {
		optionalParams.FilterKeyword = state.FilterKeyword.ValueStringPointer()
	}

	pageSize := int64(100)
	pageNumber := int64(0)

	var userTeams []datadogV2.UserTeam
	for {
		optionalParams.PageNumber = &pageNumber
		optionalParams.PageSize = &pageSize

		ddResp, _, err := d.Api.GetTeamMemberships(d.Auth, teamID, optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting team memberships"))
			return
		}

		userTeams = append(userTeams, ddResp.GetData()...)
		if len(ddResp.GetData()) < 100 {
			break
		}
		pageNumber++
	}

	d.updateState(&state, &userTeams)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogTeamMembershipsDataSource) updateState(state *datadogTeamMembershipsDataSourceModel, teamData *[]datadogV2.UserTeam) {
	state.ID = types.StringValue(fmt.Sprintf("%s:%s", state.TeamID.ValueString(), state.FilterKeyword.ValueString()))

	var teamMemberships []*TeamMembershipModel
	for _, user := range *teamData {
		membership := TeamMembershipModel{
			ID:     types.StringValue(user.GetId()),
			TeamId: types.StringValue(state.TeamID.ValueString()),
			UserId: types.StringValue(user.Relationships.User.Data.GetId()),
			Role:   types.StringValue(string(user.Attributes.GetRole())),
		}

		teamMemberships = append(teamMemberships, &membership)
	}

	state.TeamMemberships = teamMemberships
}
