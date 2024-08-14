package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogDashboardListDatasource{}
)

func NewDatadogDashboardListDataSource() datasource.DataSource {
	return &datadogDashboardListDatasource{}
}

type datadogDashboardListDatasourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type datadogDashboardListDatasource struct {
	Api  *datadogV1.DashboardListsApi
	Auth context.Context
}

func (d *datadogDashboardListDatasource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetDashboardListsApiV1()
	d.Auth = providerData.Auth
}

func (d *datadogDashboardListDatasource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "dashboard_list"
}

func (d *datadogDashboardListDatasource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing dashboard list, for use in other resources. In particular, it can be used in a dashboard to register it in the list.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Description: "A dashboard list name to limit the search.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"id": utils.ResourceIDAttribute(),
		},
	}
}

func (d *datadogDashboardListDatasource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogDashboardListDatasourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	listResponse, httpresp, err := d.Api.ListDashboardLists(d.Auth)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpresp, ""), "error querying dashboard lists"))
		return
	}

	searchedName := state.Name.ValueString()
	var foundList *datadogV1.DashboardList
	for _, dashList := range listResponse.GetDashboardLists() {
		if dashList.GetName() == searchedName {
			foundList = &dashList
			break
		}
	}

	if foundList == nil {
		errString := "Couldn't find a dashboard list named" + searchedName
		resp.Diagnostics.AddError(errString, "")
		return
	}

	if err := utils.CheckForUnparsed(foundList); err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, ""))
	}

	id := foundList.GetId()
	state.ID = types.StringValue(strconv.Itoa(int(id)))
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
