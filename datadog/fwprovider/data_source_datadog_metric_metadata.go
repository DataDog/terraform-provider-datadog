package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogMetricMetadataDataSource{}
)

type datadogMetricMetadataModel struct {
	ID             types.String `tfsdk:"id"`
	Type           types.String `tfsdk:"type"`
	MetricName     types.String `tfsdk:"metric_name"`
	ShortName      types.String `tfsdk:"short_name"`
	Description    types.String `tfsdk:"description"`
	Integration    types.String `tfsdk:"integration"`
	Unit           types.String `tfsdk:"unit"`
	PerUnit        types.String `tfsdk:"per_unit"`
	StatsdInterval types.Int64  `tfsdk:"statsd_interval"`
}

type datadogMetricMetadataDataSource struct {
	Api  *datadogV1.MetricsApi
	Auth context.Context
}

func NewDatadogMetricMetadataDataSource() datasource.DataSource {
	return &datadogMetricMetadataDataSource{}
}

func (d *datadogMetricMetadataDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetMetricsApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogMetricMetadataDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "metric_metadata"
}

func (d *datadogMetricMetadataDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve metadata associated with a metric to use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric_name": schema.StringAttribute{
				Description: "The metric for which to fetch metadata.",
				Required:    true,
			},
			// Computed values
			"type": schema.StringAttribute{
				Description: "The metric type.",
				Computed:    true,
			},
			"short_name": schema.StringAttribute{
				Description: "The metric short name.",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The metric description.",
				Computed:    true,
			},
			"integration": schema.StringAttribute{
				Description: "The metric integration.",
				Computed:    true,
			},
			"unit": schema.StringAttribute{
				Description: "The metric unit.",
				Computed:    true,
			},
			"per_unit": schema.StringAttribute{
				Description: "The per unit of the metric.",
				Computed:    true,
			},
			"statsd_interval": schema.Int64Attribute{
				Description: "The metric statsd interval.",
				Computed:    true,
			},
		},
	}
}

func (d *datadogMetricMetadataDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogMetricMetadataModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, _, err := d.Api.GetMetricMetadata(d.Auth, state.MetricName.ValueString())
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting metric metadata"))
		return
	}

	d.updateState(ctx, &state, ddResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogMetricMetadataDataSource) updateState(ctx context.Context, state *datadogMetricMetadataModel, metricMetadata datadogV1.MetricMetadata) {
	state.ID = types.StringValue(utils.ConvertToSha256(state.MetricName.String()))
	state.Type = types.StringValue(metricMetadata.GetType())
	state.ShortName = types.StringValue(metricMetadata.GetShortName())
	state.Description = types.StringValue(metricMetadata.GetDescription())
	state.Integration = types.StringValue(metricMetadata.GetIntegration())
	state.Unit = types.StringValue(metricMetadata.GetUnit())
	state.PerUnit = types.StringValue(metricMetadata.GetPerUnit())
	state.StatsdInterval = types.Int64Value(metricMetadata.GetStatsdInterval())
}
