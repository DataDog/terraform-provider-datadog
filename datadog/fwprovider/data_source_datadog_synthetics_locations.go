package fwprovider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogSyntheticsLocationsDataSource{}
)

func NewDatadogSyntheticsLocationsDataSource() datasource.DataSource {
	return &datadogSyntheticsLocationsDataSource{}
}

type datadogSyntheticsLocationsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Locations types.Map    `tfsdk:"locations"`
}

type datadogSyntheticsLocationsDataSource struct {
	Api  *datadogV1.SyntheticsApi
	Auth context.Context
}

func (d *datadogSyntheticsLocationsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetSyntheticsApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogSyntheticsLocationsDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "synthetics_locations"
}

func (d *datadogSyntheticsLocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"locations": schema.MapAttribute{
				Description: "A map of available Synthetics location IDs to names for Synthetics tests.",
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *datadogSyntheticsLocationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogSyntheticsLocationsDataSourceModel

	syntheticsLocations, _, err := d.Api.ListLocations(d.Auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting Synthetics Locations"))
		return
	}

	locationsMap := make(map[string]string)
	for _, location := range syntheticsLocations.GetLocations() {
		locationsMap[location.GetId()] = location.GetName()
	}

	locations, diags := types.MapValueFrom(ctx, types.StringType, locationsMap)
	response.Diagnostics.Append(diags...)
	state.Locations = locations

	state.ID = types.StringValue("datadog-synthetics-location")

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}
