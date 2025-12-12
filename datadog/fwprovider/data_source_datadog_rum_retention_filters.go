package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &rumRetentionFiltersDataSource{}
)

func NewRumRetentionFiltersDataSource() datasource.DataSource {
	return &rumRetentionFiltersDataSource{}
}

type rumRetentionFiltersDataSource struct {
	Api  *datadogV2.RumRetentionFiltersApi
	Auth context.Context
}

type rumRetentionFiltersDataSourceModel struct {
	ID               types.String                        `tfsdk:"id"`
	ApplicationID    types.String                        `tfsdk:"application_id"`
	RetentionFilters []rumRetentionFilterDataSourceModel `tfsdk:"retention_filters"`
}

type rumRetentionFilterDataSourceModel struct {
	ID         types.String  `tfsdk:"id"`
	Name       types.String  `tfsdk:"name"`
	EventType  types.String  `tfsdk:"event_type"`
	SampleRate types.Float64 `tfsdk:"sample_rate"`
	Query      types.String  `tfsdk:"query"`
	Enabled    types.Bool    `tfsdk:"enabled"`
}

func (r *rumRetentionFiltersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRumRetentionFiltersApiV2()
	r.Auth = providerData.Auth
}

func (r *rumRetentionFiltersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "rum_retention_filters"
}

func (r *rumRetentionFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog RUM retention filters datasource. This can be used to retrieve all RUM retention filters for a given RUM application.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"application_id": schema.StringAttribute{
				Description: "RUM application ID.",
				Required:    true,
			},
			"retention_filters": schema.ListAttribute{
				Description: "The list of RUM retention filters.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"name":        types.StringType,
						"event_type":  types.StringType,
						"sample_rate": types.Float64Type,
						"query":       types.StringType,
						"enabled":     types.BoolType,
					},
				},
			},
		}}
}

func (r *rumRetentionFiltersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state rumRetentionFiltersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	state.ID = state.ApplicationID

	resp, _, err := r.Api.ListRetentionFilters(r.Auth, state.ApplicationID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing RUM retention filters"))
		return
	}

	r.updateState(&state, &resp)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *rumRetentionFiltersDataSource) updateState(state *rumRetentionFiltersDataSourceModel, resp *datadogV2.RumRetentionFiltersResponse) {
	retentionFilters := make([]rumRetentionFilterDataSourceModel, len(resp.GetData()))
	for i, retentionFilter := range resp.GetData() {
		retentionFilters[i] = rumRetentionFilterDataSourceModel{
			ID:         types.StringValue(*retentionFilter.Id),
			Name:       types.StringValue(*retentionFilter.Attributes.Name),
			EventType:  types.StringValue(string(*retentionFilter.Attributes.EventType)),
			SampleRate: types.Float64Value(*retentionFilter.Attributes.SampleRate),
			Query:      types.StringValue(*retentionFilter.Attributes.Query),
			Enabled:    types.BoolValue(*retentionFilter.Attributes.Enabled),
		}
	}
	state.RetentionFilters = retentionFilters
}
