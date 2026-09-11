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
	_ datasource.DataSource = &datadogMetricTagConfigurationDataSource{}
)

type datadogMetricTagConfigurationModel struct {
	ID                 types.String `tfsdk:"id"`
	MetricName         types.String `tfsdk:"metric_name"`
	Exists             types.Bool   `tfsdk:"exists"`
	MetricType         types.String `tfsdk:"metric_type"`
	Tags               types.List   `tfsdk:"tags"`
	IncludePercentiles types.Bool   `tfsdk:"include_percentiles"`
	ExcludeTagsMode    types.Bool   `tfsdk:"exclude_tags_mode"`
}

type datadogMetricTagConfigurationDataSource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

func NewDatadogMetricTagConfigurationDataSource() datasource.DataSource {
	return &datadogMetricTagConfigurationDataSource{}
}

func (d *datadogMetricTagConfigurationDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogMetricTagConfigurationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "metric_tag_configuration"
}

func (d *datadogMetricTagConfigurationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve the tag configuration for a metric, or to test whether one exists. " +
			"Unlike most Datadog data sources, this one returns successfully when the metric has no tag configuration; " +
			"in that case `exists` is `false` and the other computed attributes are unset. " +
			"This is intended for use with `for_each` and `import {}` blocks driven by dynamic discovery, where some " +
			"discovered metrics may not have a tag configuration yet.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric_name": schema.StringAttribute{
				Description: "The metric for which to fetch the tag configuration.",
				Required:    true,
			},
			// Computed values
			"exists": schema.BoolAttribute{
				Description: "Whether a tag configuration exists for this metric.",
				Computed:    true,
			},
			"metric_type": schema.StringAttribute{
				Description: "The metric type. One of `gauge`, `count`, `rate`, `distribution`. Empty when `exists` is `false`.",
				Computed:    true,
			},
			"tags": schema.ListAttribute{
				Description: "The list of queryable tag keys. Empty when `exists` is `false`.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"include_percentiles": schema.BoolAttribute{
				Description: "Whether percentile aggregations are configured. Only meaningful for distribution metrics. Null when `exists` is `false` or the field is not set.",
				Computed:    true,
			},
			"exclude_tags_mode": schema.BoolAttribute{
				Description: "If `true`, the `tags` list is treated as an exclude list rather than an include list. Null when `exists` is `false` or the field is not set.",
				Computed:    true,
			},
		},
	}
}

func (d *datadogMetricTagConfigurationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogMetricTagConfigurationModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	metricName := state.MetricName.ValueString()
	ddResp, httpResp, err := d.Api.ListTagConfigurationByName(d.Auth, metricName)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			d.setNotExists(ctx, &state, metricName)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching metric tag configuration"))
		return
	}

	data := ddResp.GetData()
	attrs := data.GetAttributes()
	state.ID = types.StringValue(metricName)
	state.Exists = types.BoolValue(true)
	state.MetricType = types.StringValue(string(attrs.GetMetricType()))
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, attrs.GetTags())
	if v, ok := attrs.GetIncludePercentilesOk(); ok && v != nil {
		state.IncludePercentiles = types.BoolValue(*v)
	} else {
		state.IncludePercentiles = types.BoolNull()
	}
	if v, ok := attrs.GetExcludeTagsModeOk(); ok && v != nil {
		state.ExcludeTagsMode = types.BoolValue(*v)
	} else {
		state.ExcludeTagsMode = types.BoolNull()
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogMetricTagConfigurationDataSource) setNotExists(ctx context.Context, state *datadogMetricTagConfigurationModel, metricName string) {
	state.ID = types.StringValue(metricName)
	state.Exists = types.BoolValue(false)
	state.MetricType = types.StringValue("")
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, []string{})
	state.IncludePercentiles = types.BoolNull()
	state.ExcludeTagsMode = types.BoolNull()
}
