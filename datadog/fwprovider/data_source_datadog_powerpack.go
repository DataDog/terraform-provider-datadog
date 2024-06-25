package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogPowerpackDataSource{}
)

func NewDatadogPowerpackDataSource() datasource.DataSource {
	return &datadogPowerpackDataSource{}
}

type datadogPowerpackDataSourceModel struct {
	// Query Parameters
	PowerpackName types.String `tfsdk:"powerpack_name"`

	// Results
	ID types.String `tfsdk:"id"`
}

type datadogPowerpackDataSource struct {
	Api  *datadogV2.PowerpackApi
	Auth context.Context
}

func (r *datadogPowerpackDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetPowerpackApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogPowerpackDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "powerpack"
}

func (d *datadogPowerpackDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog Powerpack.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"powerpack_name": schema.StringAttribute{
				Description: "The name of the powerpack to find",
				Computed:    false,
				Required:    true,
			},
		},
	}

}

func (d *datadogPowerpackDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogPowerpackDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.PowerpackName.IsNull() {

		var params datadogV2.ListPowerpacksOptionalParameters
		var pageSize = 100
		var pageOffset = 0

		params.WithPageLimit(int64(pageSize))

		var powerPacks []*datadogV2.PowerpackData

		for {
			params.WithPageOffset(int64(pageOffset))

			ddResp, _, err := d.Api.ListPowerpacks(d.Auth, params)
			if err != nil {
				resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting powerpacks"))
				return
			}

			for _, r := range ddResp.GetData() {
				if r.Attributes.GetName() == state.PowerpackName.ValueString() {
					powerPacks = append(powerPacks, &r)
					break
				}
			}

			if len(ddResp.GetData()) < pageSize {
				break
			}

			pageOffset++
		}

		if len(powerPacks) == 0 {
			resp.Diagnostics.AddError(fmt.Sprintf("unable to find powerpack with name %s", state.PowerpackName.String()), "")
			return
		}

		if len(powerPacks) > 1 {
			resp.Diagnostics.AddError(fmt.Sprintf("multiple powerpacks found named %s, please provide a unique name", state.PowerpackName.String()), "")
			return
		}

		d.updateStateFromListResponse(&state, powerPacks[0])
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogPowerpackDataSource) updateState(state *datadogPowerpackDataSourceModel, PowerpackData *datadogV2.PowerpackData) {
	state.ID = types.StringValue(PowerpackData.GetId())
}

func (r *datadogPowerpackDataSource) updateStateFromListResponse(state *datadogPowerpackDataSourceModel, PowerpackData *datadogV2.PowerpackData) {
	state.ID = types.StringValue(PowerpackData.GetId())
}
