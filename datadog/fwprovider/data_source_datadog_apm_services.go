package fwprovider

import (
	"context"
	"sort"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogApmServicesDataSource{}
)

// ApmServiceModel represents a single APM service entry
type ApmServiceModel struct {
	Name     types.String `tfsdk:"name"`
	IsTraced types.Bool   `tfsdk:"is_traced"`
	IsUsm    types.Bool   `tfsdk:"is_usm"`
}

type datadogApmServicesDataSourceModel struct {
	// Query Parameters
	FilterEnv types.String `tfsdk:"filter_env"`

	// Results
	ID       types.String       `tfsdk:"id"`
	Services []*ApmServiceModel `tfsdk:"services"`
}

// NewDatadogApmServicesDataSource creates a new APM services data source
func NewDatadogApmServicesDataSource() datasource.DataSource {
	return &datadogApmServicesDataSource{}
}

type datadogApmServicesDataSource struct {
	Api  *datadogV2.APMApi
	Auth context.Context
}

func (d *datadogApmServicesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetAPMApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogApmServicesDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "apm_services"
}

func (d *datadogApmServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to list the APM services reporting traces or universal service monitoring data to Datadog.",
		Attributes: map[string]schema.Attribute{
			// Datasource Parameters
			"filter_env": schema.StringAttribute{
				Description: "Filter services by environment. Can be set to `*` to return all services across all environments.",
				Required:    true,
			},

			// Computed values
			"id": utils.ResourceIDAttribute(),
			"services": schema.ListAttribute{
				Description: "List of APM services.",
				Computed:    true,
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":      types.StringType,
						"is_traced": types.BoolType,
						"is_usm":    types.BoolType,
					},
				},
			},
		},
	}
}

func (d *datadogApmServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogApmServicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filterEnv := state.FilterEnv.ValueString()

	ddResp, httpResp, err := d.Api.GetServiceList(d.Auth, filterEnv)
	if err != nil {
		resp.Diagnostics.Append(utils.FrameworkErrorDiag(utils.TranslateClientError(err, httpResp, "error getting apm service list"), ""))
		return
	}

	d.updateState(&state, &ddResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *datadogApmServicesDataSource) updateState(state *datadogApmServicesDataSourceModel, resp *datadogV2.ServiceList) {
	data := resp.GetData()
	attributes := data.GetAttributes()
	names := attributes.GetServices()
	metadata := attributes.GetMetadata()

	services := make([]*ApmServiceModel, 0, len(names))
	for idx, name := range names {
		service := &ApmServiceModel{Name: types.StringValue(name)}
		if idx < len(metadata) {
			service.IsTraced = types.BoolValue(metadata[idx].GetIsTraced())
			service.IsUsm = types.BoolValue(metadata[idx].GetIsUsm())
		} else {
			service.IsTraced = types.BoolValue(false)
			service.IsUsm = types.BoolValue(false)
		}
		services = append(services, service)
	}

	// Making sure that the ordering is stable
	sort.Slice(services, func(i, j int) bool {
		return services[i].Name.ValueString() < services[j].Name.ValueString()
	})

	state.ID = types.StringValue(state.FilterEnv.ValueString())
	state.Services = services
}
