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
	_ datasource.DataSource = &datadogMetricActiveTagsAndAggregationsDataSource{}
)

type datadogMetricActiveTagsAndAggregationsModel struct {
	ID                 types.String                             `tfsdk:"id"`
	Metric             types.String                             `tfsdk:"metric"`
	ActiveTags         types.List                               `tfsdk:"active_tags"`
	Window             types.Int64                              `tfsdk:"window"`
	ActiveAggregations []metricActiveAggregationDataSourceModel `tfsdk:"active_aggregations"`
}

type metricActiveAggregationDataSourceModel struct {
	Space types.String `tfsdk:"space"`
	Time  types.String `tfsdk:"time"`
}

type datadogMetricActiveTagsAndAggregationsDataSource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

func NewDatadogMetricActiveTagsAndAggregationsDataSource() datasource.DataSource {
	return &datadogMetricActiveTagsAndAggregationsDataSource{}
}

func (d *datadogMetricActiveTagsAndAggregationsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogMetricActiveTagsAndAggregationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "metric_active_tags_and_aggregations"
}

func (d *datadogMetricActiveTagsAndAggregationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve active tags and aggregations associated with a metric to use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric": schema.StringAttribute{
				Description: "The metric for which to fetch tags.",
				Required:    true,
			},
			"window": schema.Int64Attribute{
				Description: "The number of seconds to look back from now.",
				Optional:    true,
			},
			// Computed values
			"active_tags": schema.ListAttribute{
				Description: "The active tags associated with the metric.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"active_aggregations": schema.ListAttribute{
				Description: "The active aggregations associated with the metric.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"space": types.StringType,
						"time":  types.StringType,
					},
				},
			},
		},
	}
}

func (d *datadogMetricActiveTagsAndAggregationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogMetricActiveTagsAndAggregationsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := *datadogV2.NewListActiveMetricConfigurationsOptionalParameters()
	if !state.Window.IsNull() {
		params = *params.WithWindowSeconds(state.Window.ValueInt64())
	}
	ddResp, _, err := d.Api.ListActiveMetricConfigurations(d.Auth, state.Metric.String(), params)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing active metric tags and aggregations"))
		return
	}

	data := ddResp.GetData()
	id := data.Id
	tags := data.Attributes.GetActiveTags()
	aggregations := data.Attributes.GetActiveAggregations()

	d.updateState(ctx, &state, *id, tags, aggregations)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogMetricActiveTagsAndAggregationsDataSource) updateState(ctx context.Context, state *datadogMetricActiveTagsAndAggregationsModel, id string, tags []string, aggregations []datadogV2.MetricCustomAggregation) {
	state.ID = types.StringValue(id)
	state.ActiveTags, _ = types.ListValueFrom(ctx, types.StringType, tags)
	activeAggs := make([]metricActiveAggregationDataSourceModel, len(aggregations))
	for i, activeAgg := range aggregations {
		activeAggs[i] = metricActiveAggregationDataSourceModel{
			Time:  types.StringValue(string(activeAgg.Time)),
			Space: types.StringValue(string(activeAgg.Space)),
		}
	}
	state.ActiveAggregations = activeAggs
}
