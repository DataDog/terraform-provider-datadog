package fwprovider

import (
	"context"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogMetricsDataSource{}
)

type datadogMetricsModel struct {
	ID      types.String `tfsdk:"id"`
	Query   types.String `tfsdk:"query"`
	Metrics types.List   `tfsdk:"metrics"`
}

type datadogMetricsDataSource struct {
	Api  *datadogV1.MetricsApi
	Auth context.Context
}

func NewDatadogMetricsDataSource() datasource.DataSource {
	return &datadogMetricsDataSource{}
}

func (d *datadogMetricsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetMetricsApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogMetricsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "metrics"
}

func (d *datadogMetricsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list metrics for use in other resources.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"query": schema.StringAttribute{
				Description: "The search query to use when listing metrics.",
				Required:    true,
			},
			// Computed values
			"metrics": schema.ListAttribute{
				Description: "The metrics returned by the search query.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *datadogMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogMetricsModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	ddResp, _, err := d.Api.ListMetrics(d.Auth, state.Query.ValueString())
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error listing metrics"))
		return
	}

	results := ddResp.GetResults()
	metrics := results.GetMetrics()

	d.updateState(ctx, &state, metrics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogMetricsDataSource) updateState(ctx context.Context, state *datadogMetricsModel, metrics []string) {
	hashingData := fmt.Sprintf("%s:%s", state.Query.ValueString(), strings.Join(metrics, ":"))

	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.Metrics, _ = types.ListValueFrom(ctx, types.StringType, metrics)
}
