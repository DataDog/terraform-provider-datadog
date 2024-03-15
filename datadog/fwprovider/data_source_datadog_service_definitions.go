package fwprovider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogServiceDefinitionsDataSource{}
)

type ServiceDefinitionSchema struct {
	Application string `json:"application"`
	DDService   string `json:"dd-service"`
	Description string `json:"description"`
	Team        string `json:"team"`
	Tier        string `json:"tier"`
}

type ServiceDefinitionModel struct {
	ID          types.String `tfsdk:"id" json:"-"`
	Application types.String `tfsdk:"application"`
	Service     types.String `tfsdk:"service"`
	Description types.String `tfsdk:"description"`
	Team        types.String `tfsdk:"team"`
}

type datadogServiceDefinitionsDataSourceModel struct {
	// Query Parameters
	RetrieveAll types.Bool `tfsdk:"retrieve_all"`

	// Results
	ID                 types.String              `tfsdk:"id"`
	ServiceDefinitions []*ServiceDefinitionModel `tfsdk:"service_definitions"`
}

func NewDatadogServiceDefinitionsDataSource() datasource.DataSource {
	return &datadogServiceDefinitionsDataSource{}
}

type datadogServiceDefinitionsDataSource struct {
	Api  *datadogV2.ServiceDefinitionApi
	Auth context.Context
}

func (r *datadogServiceDefinitionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetServiceDefinitionApiV2()
	r.Auth = providerData.Auth
}

func (d *datadogServiceDefinitionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "service_definitions"
}

func (d *datadogServiceDefinitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing Datadog service definitions.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Datasource Parameters
			"retrieve_all": schema.BoolAttribute{
				Description: "Retrieve all service definitions.",
				Required:    true,
			},

			// Computed values
			"service_definitions": schema.ListAttribute{
				Computed:    true,
				Description: "List of service definitions.",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"application": types.StringType,
						"service":     types.StringType,
						"description": types.StringType,
						"team":        types.StringType,
					},
				},
			},
		},
	}

}

func (d *datadogServiceDefinitionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state datadogServiceDefinitionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var optionalParams datadogV2.ListServiceDefinitionsOptionalParameters
	pageSize := 100
	pageNumber := int64(0)
	optionalParams.WithPageSize(int64(pageSize))

	var serviceDefinitons []datadogV2.ServiceDefinitionData
	for {
		optionalParams.WithPageNumber(pageNumber)

		ddResp, _, err := d.Api.ListServiceDefinitions(d.Auth, optionalParams)
		if err != nil {
			resp.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting service definitions"))
			return
		}

		serviceDefinitons = append(serviceDefinitons, ddResp.GetData()...)
		if len(ddResp.GetData()) < pageSize {
			break
		}
		pageNumber++
	}
	d.updateState(&state, &serviceDefinitons)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *datadogServiceDefinitionsDataSource) updateState(state *datadogServiceDefinitionsDataSourceModel, serviceDefinitionData *[]datadogV2.ServiceDefinitionData) {
	var serviceDefinitions []*ServiceDefinitionModel

	for _, serviceDefinition := range *serviceDefinitionData {
		var schema ServiceDefinitionSchema
		jsonData, _ := json.MarshalIndent(serviceDefinition.Attributes.Schema, "", "  ")
		_ = json.Unmarshal(jsonData, &schema)

		s := ServiceDefinitionModel{
			ID:          types.StringValue(serviceDefinition.GetId()),
			Application: types.StringValue(schema.Application),
			Service:     types.StringValue(schema.DDService),
			Description: types.StringValue(schema.Description),
			Team:        types.StringValue(schema.Team),
		}

		serviceDefinitions = append(serviceDefinitions, &s)
	}

	hashingData := fmt.Sprintf("datadog_service_definitions:%t", state.RetrieveAll.ValueBool())
	state.ID = types.StringValue(utils.ConvertToSha256(hashingData))
	state.ServiceDefinitions = serviceDefinitions
}
