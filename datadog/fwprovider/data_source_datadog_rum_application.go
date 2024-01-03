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

var (
	_ datasource.DataSource = &RumApplicationDataSource{}
)

func NewRumApplicationDataSource() datasource.DataSource {
	return &RumApplicationDataSource{}
}

type RumApplicationDataSourceModel struct {
	// Query Parameters
	NameFilter types.String `tfsdk:"name_filter"`
	TypeFilter types.String `tfsdk:"type_filter"`
	// Results
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Type        types.String `tfsdk:"type"`
	ClientToken types.String `tfsdk:"client_token"`
}

type RumApplicationDataSource struct {
	Api  *datadogV2.RUMApi
	Auth context.Context
}

func (r *RumApplicationDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetRumApiV2()
	r.Auth = providerData.Auth
}

func (d *RumApplicationDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "rum_application"
}

func (d *RumApplicationDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
	return []datasource.ConfigValidator{
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("type_filter"),
		),
		datasourcevalidator.Conflicting(
			path.MatchRoot("id"),
			path.MatchRoot("name_filter"),
		),
	}
}

func (d *RumApplicationDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve a Datadog RUM Application.",
		Attributes: map[string]schema.Attribute{
			// Query Parameters
			"name_filter": schema.StringAttribute{
				Description: "The name used to search for a RUM application.",
				Optional:    true,
			},
			"type_filter": schema.StringAttribute{
				Description: "The type used to search for a RUM application.",
				Optional:    true,
			},
			// Datasource ID
			"id": schema.StringAttribute{
				Description: "ID of the RUM application. Cannot be used with name and type filters.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the RUM application.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the RUM application. Supported values are `browser`, `ios`, `android`, `react-native`, `flutter`.",
			},
			"client_token": schema.StringAttribute{
				Computed:    true,
				Description: "The client token.",
			},
		},
	}

}

func (d *RumApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state RumApplicationDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.ID.IsNull() {
		searchID := state.ID.ValueString()
		rumResponse, _, err := d.Api.GetRUMApplication(d.Auth, searchID)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, fmt.Sprintf("Couldn't find RUM application with id %s", searchID)))
			return
		}
		d.updateState(ctx, &state, &rumResponse)
	} else {
		rumResponse, _, err := d.Api.GetRUMApplications(d.Auth)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "Couldn't retrieve list of RUM Applications"))
			return
		}

		searchName, searchNameOk := state.NameFilter.ValueString(), !state.NameFilter.IsNull()
		searchType, searchTypeOk := state.TypeFilter.ValueString(), !state.TypeFilter.IsNull()
		bothSet := searchNameOk && searchTypeOk

		var foundRUMApplicationIDs []string
		for _, resp_data := range rumResponse.Data {
			if rum_app, ok := resp_data.GetAttributesOk(); ok {
				nameSetAndMatched := searchNameOk && rum_app.GetName() == searchName
				typeSetAndMatched := searchTypeOk && rum_app.GetType() == searchType
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
			resp.Diagnostics.AddError("RUM Application search failed", fmt.Sprintf("Couldn't find a RUM Application with name '%s' and type '%s'", searchName, searchType))
			return
		} else if len(foundRUMApplicationIDs) > 1 {
			resp.Diagnostics.AddError("RUM Application search failed", fmt.Sprintf("Searching for name '%s' and type '%s' returned more than one RUM application.", searchName, searchType))
			return
		}

		app_resp, _, app_err := d.Api.GetRUMApplication(d.Auth, foundRUMApplicationIDs[0])
		if app_err != nil {
			resp.Diagnostics.AddError("RUM Application search failed", fmt.Sprintf("Found RUM application with id %s, but couldn't retrieve details.", foundRUMApplicationIDs[0]))
			return
		}
		d.updateState(ctx, &state, &app_resp)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *RumApplicationDataSource) updateState(ctx context.Context, state *RumApplicationDataSourceModel, resp *datadogV2.RUMApplicationResponse) {
	state.ID = types.StringValue(resp.Data.GetId())

	data := resp.GetData()
	attributes := data.GetAttributes()

	if clientToken, ok := attributes.GetClientTokenOk(); ok {
		state.ClientToken = types.StringValue(*clientToken)
	}

	if name, ok := attributes.GetNameOk(); ok {
		state.Name = types.StringValue(*name)
	}

	if typeVar, ok := attributes.GetTypeOk(); ok {
		state.Type = types.StringValue(*typeVar)
	}
}
