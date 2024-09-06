package fwprovider

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSourceWithConfigure = &asmWafExclusionFiltersDataSource{}
)

type asmWafExclusionFiltersDataSource struct {
	api  *datadogV2.ASMExclusionFiltersApi
	auth context.Context
}

type asmWafExclusionFiltersDataSourceModel struct {
	Id                  types.String                  `tfsdk:"id"`
	ExclusionFiltersIds types.List                    `tfsdk:"exclusion_filters_ids"`
	ExclusionFilters    []asmWafExclusionFiltersModel `tfsdk:"exclusion_filters"`
}

func NewAsmWafExclusionFiltersDataSource() datasource.DataSource {
	return &asmWafExclusionFiltersDataSource{}
}

func (r *asmWafExclusionFiltersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetASMExclusionFiltersApiV2()
	r.auth = providerData.Auth
}

func (*asmWafExclusionFiltersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "asm_waf_exclusion_filters"
}

func (r *asmWafExclusionFiltersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state asmWafExclusionFiltersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Fetch the exclusion filters using the API
	res, _, err := r.api.ListASMExclusionFilters(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching exclusion filters"))
		return
	}

	data := res.GetData()
	exclusionFiltersIds := make([]string, len(data))
	exclusionFilters := make([]asmWafExclusionFiltersModel, len(data))

	// Iterate through the exclusion filters data received
	for idx, exclusionFilter := range res.GetData() {
		var exclusionFilterModel asmWafExclusionFiltersModel

		// Direct mapping of fields from exclusionFilter struct
		exclusionFilterModel.Id = types.StringValue(exclusionFilter.GetId())

		// Corrected access to Attributes (with uppercase A)
		attributes := exclusionFilter.GetAttributes()

		exclusionFilterModel.Description = types.StringValue(attributes.GetDescription())
		exclusionFilterModel.Enabled = types.BoolValue(attributes.GetEnabled())
		exclusionFilterModel.Search_Query = types.StringValue(attributes.GetSearchQuery())

		// Collect the exclusion filter IDs and model
		exclusionFiltersIds[idx] = exclusionFilter.GetId()
		exclusionFilters[idx] = exclusionFilterModel
	}

	// Set the state ID based on the exclusion filters IDs
	stateId := strings.Join(exclusionFiltersIds, "--")
	state.Id = types.StringValue(computeExclusionFiltersDataSourceID(&stateId))

	// Convert the exclusion filter IDs to a Terraform list
	tfExclusionFiltersIds, diags := types.ListValueFrom(ctx, types.StringType, exclusionFiltersIds)
	response.Diagnostics.Append(diags...)
	state.ExclusionFiltersIds = tfExclusionFiltersIds
	state.ExclusionFilters = exclusionFilters

	// Save the state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func computeExclusionFiltersDataSourceID(exclusionFiltersIds *string) string { // return to state.Id
	// Key for hashing
	var b strings.Builder
	if exclusionFiltersIds != nil {
		b.WriteString(*exclusionFiltersIds)
	}
	keyStr := b.String()
	h := sha256.New()
	h.Write([]byte(keyStr))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (*asmWafExclusionFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about existing WAF exclusion filters.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"exclusion_filters_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of IDs for the exclusion filters.",
				ElementType: types.StringType,
			},
			"exclusion_filters": schema.ListAttribute{
				Computed:    true,
				Description: "List of exclusion filters",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id": types.StringType,
						// "type":        types.StringType,
						"description": types.StringType,
						"enabled":     types.BoolType,
						// "path_glob":    types.StringType,
						"search_query": types.StringType,
					},
				},
			},
		},
	}
}
