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
	_ datasource.DataSource = &datadogTeamHierarchyLinksDataSource{}
)

type datadogTeamHierarchyLinksDataSource struct {
	Api  *datadogV2.TeamsApi
	Auth context.Context
}

type datadogTeamHierarchyLinksDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	LinkId           types.String `tfsdk:"link_id"`
	FilterParentTeam types.String `tfsdk:"filter_parent_team"`
	FilterSubTeam    types.String `tfsdk:"filter_sub_team"`

	// Computed values
	CreatedAt     types.String `tfsdk:"created_at"`
	ProvisionedBy types.String `tfsdk:"provisioned_by"`
}

func NewDatadogTeamHierarchyLinksDataSource() datasource.DataSource {
	return &datadogTeamHierarchyLinksDataSource{}
}

func (d *datadogTeamHierarchyLinksDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetTeamsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogTeamHierarchyLinksDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "team_hierarchy_links"
}

func (d *datadogTeamHierarchyLinksDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Datadog Team Hierarchy Links.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"link_id": schema.StringAttribute{
				Optional:    true,
				Description: "The team hierarchy link’s identifier",
			},
			"filter_parent_team": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by parent team ID",
			},
			"filter_sub_team": schema.StringAttribute{
				Optional:    true,
				Description: "Filter by sub team ID",
			},
			// Computed values
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the team hierarchy link was created",
			},
			"provisioned_by": schema.StringAttribute{
				Computed:    true,
				Description: "The provisioner of the team hierarchy link",
			},
		},
	}
}

func (d *datadogTeamHierarchyLinksDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogTeamHierarchyLinksDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.LinkId.IsNull() {
		teamHierarchyLinkId := state.LinkId.ValueString()
		ddResp, _, err := d.Api.GetTeamHierarchyLink(d.Auth, teamHierarchyLinkId)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog teamHierarchyLinks"))
			return
		}

		d.updateState(ctx, &state, ddResp.Data)
	} else {
		filterParentTeam := state.FilterParentTeam.ValueString()
		filterSubTeam := state.FilterSubTeam.ValueString()

		optionalParams := datadogV2.ListTeamHierarchyLinksOptionalParameters{
			FilterParentTeam: &filterParentTeam,
			FilterSubTeam:    &filterSubTeam,
		}

		ddResp, _, err := d.Api.ListTeamHierarchyLinks(d.Auth, optionalParams)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing datadog teamHierarchyLinks"))
			return
		}

		if len(ddResp.Data) > 1 {
			response.Diagnostics.AddError("filters returned more than one result, use more specific search criteria", "")
			return
		}
		if len(ddResp.Data) == 0 {
			response.Diagnostics.AddError("filters returned no results", "")
			return
		}

		d.updateStateFromListResponse(ctx, &state, &ddResp.Data[0])
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogTeamHierarchyLinksDataSource) updateState(ctx context.Context, state *datadogTeamHierarchyLinksDataSourceModel, teamHierarchyLinksData *datadogV2.TeamHierarchyLink) {
	state.ID = types.StringValue(teamHierarchyLinksData.GetId())

	attributes := teamHierarchyLinksData.GetAttributes()
	state.CreatedAt = types.StringValue(attributes.GetCreatedAt().String())
	state.ProvisionedBy = types.StringValue(attributes.GetProvisionedBy())
}

func (d *datadogTeamHierarchyLinksDataSource) updateStateFromListResponse(ctx context.Context, state *datadogTeamHierarchyLinksDataSourceModel, teamHierarchyLinksData *datadogV2.TeamHierarchyLink) {
	state.ID = types.StringValue(teamHierarchyLinksData.GetId())
	state.LinkId = types.StringValue(teamHierarchyLinksData.GetId())

	attributes := teamHierarchyLinksData.GetAttributes()
	state.CreatedAt = types.StringValue(attributes.GetCreatedAt().String())
	state.ProvisionedBy = types.StringValue(attributes.GetProvisionedBy())
}
