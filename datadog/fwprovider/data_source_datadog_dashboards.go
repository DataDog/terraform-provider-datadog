package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource              = &dashboardsDataSource{}
	_ datasource.DataSourceWithConfigure = &dashboardsDataSource{}
)

func NewDashboardsDataSource() datasource.DataSource {
	return &dashboardsDataSource{}
}

type dashboardModel struct {
	ID           types.String `tfsdk:"id"`
	Title        types.String `tfsdk:"title"`
	URL          types.String `tfsdk:"url"`
	Description  types.String `tfsdk:"description"`
	AuthorHandle types.String `tfsdk:"author_handle"`
}

type dashboardsDataSourceModel struct {
	ID         types.String      `tfsdk:"id"`
	Title      types.String      `tfsdk:"title"`
	ExactMatch types.Bool        `tfsdk:"exact_match"`
	Dashboards []*dashboardModel `tfsdk:"dashboards"`
}

type dashboardsDataSource struct {
	Api  *datadogV1.DashboardsApi
	Auth context.Context
}

func (d dashboardsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetDashboardsApiV1()
	d.Auth = providerData.Auth
}

func (d dashboardsDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "dashboards"
}

func (d dashboardsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing dashboards ",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"title": schema.StringAttribute{
				Required:    true,
				Description: "The dashboards name to search for.",
			},
			"exact_match": schema.BoolAttribute{
				Description: "Whether to use exact match when searching by name.",
				Optional:    true,
			},
			"dashboards": schema.ListAttribute{
				Description: "The list of dashboards.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":            types.StringType,
						"title":         types.StringType,
						"url":           types.StringType,
						"description":   types.StringType,
						"author_handle": types.StringType,
					},
				},
			},
		},
	}
}

func (d dashboardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dashboardsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, cancel := d.Api.ListDashboardsWithPagination(d.Auth)
	defer cancel()
	var dashboards []datadogV1.DashboardSummaryDefinition
	for dashboard := range ddResp {
		if dashboard.Error != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(dashboard.Error, fmt.Sprintf("error querying dashboard : %s", dashboard.Error)))
			return
		}
		if state.ExactMatch.ValueBool() {
			if dashboard.Item.GetTitle() == state.Title.ValueString() {
				dashboards = append(dashboards, dashboard.Item)
			}
		} else {
			if strings.Contains(dashboard.Item.GetTitle(), state.Title.ValueString()) {
				dashboards = append(dashboards, dashboard.Item)
			}
		}
	}
	state.ID = types.StringValue("DASHBOARD_LIST")
	updateState(&state, dashboards)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func updateState(state *dashboardsDataSourceModel, dashboards []datadogV1.DashboardSummaryDefinition) {
	for _, dashboard := range dashboards {
		state.Dashboards = append(state.Dashboards, &dashboardModel{
			ID:           types.StringValue(dashboard.GetId()),
			Title:        types.StringValue(dashboard.GetTitle()),
			URL:          types.StringValue(dashboard.GetUrl()),
			Description:  types.StringValue(dashboard.GetDescription()),
			AuthorHandle: types.StringValue(dashboard.GetAuthorHandle()),
		})
	}
}
