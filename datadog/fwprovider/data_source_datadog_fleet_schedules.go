package fwprovider

import (
	"context"
	"fmt"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &fleetSchedulesDataSource{}

type fleetSchedulesDataSourceModel struct {
	Schedules  []fleetScheduleDataSourceModel `tfsdk:"schedules"`
	TotalCount types.Int64                    `tfsdk:"total_count"`
}

type fleetSchedulesDataSource struct {
	Api  fleetSchedulesAPI
	Auth context.Context
}

func NewFleetSchedulesDataSource() datasource.DataSource {
	return &fleetSchedulesDataSource{}
}

func (d *fleetSchedulesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}
	providerData, ok := request.ProviderData.(*FrameworkProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *FrameworkProvider, got %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}
	d.Api = providerData.DatadogApiInstances.GetFleetAutomationApiV2()
	d.Auth = providerData.Auth
}

func (d *fleetSchedulesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "fleet_schedules"
}

func (d *fleetSchedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Lists Fleet Automation Agent upgrade schedules in ID order. An empty result is successful. Reading schedules requires an application key with the `hosts_read` permission.",
		Attributes: map[string]schema.Attribute{
			"schedules": schema.ListNestedAttribute{
				Description: "Fleet Automation schedules, sorted by ID.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: fleetScheduleDataSourceAttributes(false),
				},
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of schedules returned by the API.",
				Computed:    true,
			},
		},
	}
}

func (d *fleetSchedulesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	apiResponse, httpResponse, err := d.Api.ListFleetSchedulesV2(d.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResponse, "error listing Fleet Automation schedules"), ""))
		return
	}

	state, diags := fleetSchedulesDataSourceModelFromAPI(ctx, apiResponse)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func fleetSchedulesDataSourceModelFromAPI(ctx context.Context, apiResponse datadogV2.FleetSchedulesV2Response) (fleetSchedulesDataSourceModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	schedules := append([]datadogV2.FleetScheduleV2(nil), apiResponse.GetData()...)
	sort.Slice(schedules, func(i, j int) bool { return schedules[i].GetId() < schedules[j].GetId() })
	state := fleetSchedulesDataSourceModel{
		Schedules:  make([]fleetScheduleDataSourceModel, 0, len(schedules)),
		TotalCount: types.Int64Value(int64(len(schedules))),
	}
	if meta, ok := apiResponse.GetMetaOk(); ok && meta != nil {
		if page, ok := meta.GetPageOk(); ok && page != nil {
			if totalCount, ok := page.GetTotalCountOk(); ok && totalCount != nil {
				state.TotalCount = types.Int64Value(*totalCount)
			}
		}
	}
	for _, schedule := range schedules {
		model, modelDiags := fleetScheduleDataSourceModelFromAPI(ctx, schedule)
		diags.Append(modelDiags...)
		state.Schedules = append(state.Schedules, model)
	}
	return state, diags
}
