package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource = &datadogTeamDataSource{}
)

func NewDatadogApmRetentionFiltersOrderDataSource() datasource.DataSource {
	return &datadogApmRetentionFiltersOrderDataSource{}
}

type datadogApmRetentionFiltersOrderDataSource struct {
	Api  *datadogV2.APMRetentionFiltersApi
	Auth context.Context
}

func (r *datadogApmRetentionFiltersOrderDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetApmRetentionFiltersApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogApmRetentionFiltersOrderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apm_retention_filters_order"
}

func (d *datadogApmRetentionFiltersOrderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog [APM Retention Filters API](https://docs.datadoghq.com/api/v2/apm-retention-filters/) resource, which is used to manage Datadog APM retention filters order.",
		Attributes: map[string]schema.Attribute{
			"filter_ids": schema.ListAttribute{
				Description: "The filter IDs list. The order of filters IDs in this attribute defines the overall APM retention filters order.. If `filter_ids` is empty or not specified, it will import the actual order, and create the resource. Otherwise, it will try to update the order.",
				ElementType: types.StringType,
				Optional:    true,
			},
			"id": utils.ResourceIDAttribute(),
		}}
}

func (d *datadogApmRetentionFiltersOrderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ApmRetentionFiltersOrderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	ddResp, _, err := d.Api.ListApmRetentionFilters(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog team"))
		return
	}

	d.updateState(&state, &ddResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogApmRetentionFiltersOrderDataSource) updateState(state *ApmRetentionFiltersOrderModel, resp *datadogV2.RetentionFiltersResponse) {
	filterIds := GetApmFilterIds(*resp)
	state.ID = types.StringValue("apm-retention-filters")
	state.FilterIds = filterIds
}
