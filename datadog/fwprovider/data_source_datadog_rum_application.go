package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var _ datasource.DataSourceWithConfigure = &rumApplicationDataSource{}
var _ datasource.DataSourceWithConfigValidators = &rumApplicationDataSource{}

func NewRumApplicationDataSource() datasource.DataSource {
	return &rumApplicationDataSource{}
}

type rumApplicationDataSource struct {
	Api  *datadogV2.RUMApi
	Auth context.Context
}

type rumApplicationDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	NameFilter  types.String `tfsdk:"name_filter"`
	TypeFilter  types.String `tfsdk:"type_filter"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	ClientToken types.String `tfsdk:"client_token"`
}

func (d *rumApplicationDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "rum_application"
}

func (d *rumApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve a Datadog RUM Application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "ID of the RUM application. Cannot be used with name and type filters.",
			},
			"name_filter": schema.StringAttribute{
				Optional:    true,
				Description: "The name used to search for a RUM application",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the RUM application",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The RUM application type. Supported values are `browser`, `ios`, `android`, `react-native`, `flutter`",
			},
			"type_filter": schema.StringAttribute{
				Optional:    true,
				Description: "The type used to search for a RUM application",
			},
			"client_token": schema.StringAttribute{
				Computed:    true,
				Description: "The client token",
			},
		},
	}
}

func (d *rumApplicationDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetRumApiV2()
	d.Auth = providerData.Auth
}

func (d *rumApplicationDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("name_filter"),
			path.MatchRoot("id"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("type_filter"),
			path.MatchRoot("id"),
		),
	}
}

func (d *rumApplicationDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state rumApplicationDataSourceModel

	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		resp, _, err := d.Api.GetRUMApplication(d.Auth, state.ID.ValueString())
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("Couldn't find RUM application with id %s", state.ID.ValueString())))
		}
		d.updateState(&state, resp.Data.GetAttributes())
	} else {
		resp, _, err := d.Api.GetRUMApplications(d.Auth)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Couldn't retrieve list of RUM Applications"))
		}
		searchName := state.NameFilter
		searchType := state.TypeFilter
		bothSet := !searchName.IsNull() && !searchType.IsNull()
		var foundRUMApplicationIDs []string
		for _, resp_data := range resp.Data {
			if rum_app, ok := resp_data.GetAttributesOk(); ok {
				nameSetAndMatched := !state.NameFilter.IsNull() && types.StringValue(rum_app.GetName()) == searchName
				typeSetAndMatched := !searchType.IsNull() && types.StringValue(rum_app.GetType()) == searchType
				if bothSet {
					if nameSetAndMatched && typeSetAndMatched {
						foundRUMApplicationIDs = append(foundRUMApplicationIDs, rum_app.GetApplicationId())
					}
				} else if nameSetAndMatched || typeSetAndMatched {
					foundRUMApplicationIDs = append(foundRUMApplicationIDs, rum_app.GetApplicationId())
				}
			}
		}

		if len(foundRUMApplicationIDs) == 0 {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("Couldn't find a RUM Application with name '%s' and type '%s'", searchName, searchType)))
		} else if len(foundRUMApplicationIDs) > 1 {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("Searching for name '%s' and type '%s' returned more than one RUM application.", searchName, searchType)))
		}

		app_resp, _, app_err := d.Api.GetRUMApplication(d.Auth, foundRUMApplicationIDs[0])
		if app_err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("Found RUM application with id %s, but couldn't retrieve details.", foundRUMApplicationIDs[0])))
		}
		d.updateState(&state, app_resp.Data.GetAttributes())
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *rumApplicationDataSource) updateState(state *rumApplicationDataSourceModel, rumApplication datadogV2.RUMApplicationAttributes) {
	state.ID = types.StringValue(rumApplication.GetApplicationId())
	state.Name = types.StringValue(rumApplication.GetName())
	state.Type = types.StringValue(rumApplication.GetType())
	state.ClientToken = types.StringValue(rumApplication.GetClientToken())

}
