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
	Name types.String `tfsdk:"name"`
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
			// Datasource Query Parameters
			"name": schema.StringAttribute{
				Description: "The name of the Powerpack to search for.",
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

	if !state.Name.IsNull() {
		var powerpacks []*datadogV2.PowerpackData
		response, _ := d.Api.ListPowerpacksWithPagination(d.Auth, *datadogV2.NewListPowerpacksOptionalParameters().WithPageLimit(100))
		for paginationResult := range response {
			if paginationResult.Error != nil {
				resp.Diagnostics.Append(utils.FrameworkErrorDiag(paginationResult.Error, "error getting powerpacks"))
				return
			}
			if paginationResult.Item.Attributes.GetName() == state.Name.ValueString() {
				powerpacks = append(powerpacks, &paginationResult.Item)
				break
			}
		}

		if len(powerpacks) == 0 {
			resp.Diagnostics.AddError(fmt.Sprintf("unable to find powerpack with name %s", state.Name.String()), "")
			return
		}

		if len(powerpacks) > 1 {
			resp.Diagnostics.AddError(fmt.Sprintf("multiple powerpacks found named %s, please provide a unique name", state.Name.String()), "")
			return
		}

		d.updateState(&state, powerpacks[0])
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogPowerpackDataSource) updateState(state *datadogPowerpackDataSourceModel, PowerpackData *datadogV2.PowerpackData) {
	state.ID = types.StringValue(PowerpackData.GetId())
}
