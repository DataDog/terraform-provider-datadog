package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogMetricTagsDataSource{}
)

type datadogMetricTagsModel struct {
	ID     types.String `tfsdk:"id"`
	Metric types.String `tfsdk:"metric"`
	Tags   types.List   `tfsdk:"tags"`
}

type datadogMetricTagsDataSource struct {
	Api  *datadogV2.MetricsApi
	Auth context.Context
}

func NewDatadogMetricTagsDataSource() datasource.DataSource {
	return &datadogMetricTagsDataSource{}
}

func (d *datadogMetricTagsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetMetricsApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogMetricTagsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "metric_tags"
}

func (d *datadogMetricTagsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve tags associated with a metric to use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"metric": schema.StringAttribute{
				Description: "The metric for which to fetch tags.",
				Required:    true,
			},
			// Computed values
			"tags": schema.ListAttribute{
				Description: "The tags associated with the metric.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *datadogMetricTagsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogMetricTagsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, _, err := d.Api.ListTagsByMetricName(d.Auth, state.Metric.ValueString())
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing metric tags"))
		return
	}

	tagsData := ddResp.GetData()
	tags := tagsData.Attributes.GetTags()

	d.updateState(ctx, &state, tags)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogMetricTagsDataSource) updateState(ctx context.Context, state *datadogMetricTagsModel, tags []string) {
	hashingData := fmt.Sprintf("%s:%s", state.Metric.ValueString(), strings.Join(tags, ":"))

	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Tags, _ = types.ListValueFrom(ctx, types.StringType, tags)
}
