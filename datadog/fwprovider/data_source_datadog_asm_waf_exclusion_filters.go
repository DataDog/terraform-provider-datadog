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

	res, _, err := r.api.ListASMExclusionFilters(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching exclusion filters"))
		return
	}

	data := res.GetData()
	exclusionFiltersIds := make([]string, len(data))
	exclusionFilters := make([]asmWafExclusionFiltersModel, len(data))

	for idx, exclusionFilter := range res.GetData() {
		var exclusionFilterModel asmWafExclusionFiltersModel

		exclusionFilterModel.Id = types.StringValue(exclusionFilter.GetId())

		attributes := exclusionFilter.GetAttributes()
		exclusionFilterModel.Description = types.StringValue(attributes.GetDescription())
		exclusionFilterModel.Enabled = types.BoolValue(attributes.GetEnabled())
		exclusionFilterModel.PathGlob = types.StringValue(attributes.GetPathGlob())

		var scopes []attr.Value
		for _, scope := range attributes.GetScope() {
			scopeObject, diags := types.ObjectValue(map[string]attr.Type{
				"env":     types.StringType,
				"service": types.StringType,
			}, map[string]attr.Value{
				"env":     types.StringValue(scope.GetEnv()),
				"service": types.StringValue(scope.GetService()),
			})

			response.Diagnostics.Append(diags...)
			scopes = append(scopes, scopeObject)
		}

		tfScopes, diags := types.ListValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"env":     types.StringType,
				"service": types.StringType,
			},
		}, scopes)

		response.Diagnostics.Append(diags...)
		exclusionFilterModel.Scope = tfScopes

		exclusionFiltersIds[idx] = exclusionFilter.GetId()
		exclusionFilters[idx] = exclusionFilterModel
	}

	var exclusionFiltersIdsAttr []attr.Value
	for _, id := range exclusionFiltersIds {
		exclusionFiltersIdsAttr = append(exclusionFiltersIdsAttr, types.StringValue(id))
	}

	tfExclusionFiltersIds, diags := types.ListValue(types.StringType, exclusionFiltersIdsAttr)
	response.Diagnostics.Append(diags...)
	state.ExclusionFiltersIds = tfExclusionFiltersIds
	state.ExclusionFilters = exclusionFilters

	stateId := strings.Join(exclusionFiltersIds, "--")
	state.Id = types.StringValue(computeExclusionFiltersDataSourceID(&stateId))

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

func (r *asmWafExclusionFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Retrieves Datadog ASM WAF Exclusion Filters.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "A description for the exclusion filter.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates whether the exclusion filter is enabled.",
			},
			"path_glob": schema.StringAttribute{
				Computed:    true,
				Description: "The path glob for the exclusion filter.",
			},
			"scope": schema.ListAttribute{
				Description: "The scope of the exclusion filter. Each entry is a map with 'env' and 'service' keys.",
				Computed:    true,
				ElementType: types.MapType{
					ElemType: types.StringType,
				},
			},
		},
	}
}
