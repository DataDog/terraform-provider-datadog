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
	_ datasource.DataSourceWithConfigure = &applicationSecurityExclusionFiltersDataSource{}
)

type applicationSecurityExclusionFiltersDataSource struct {
	api  *datadogV2.ApplicationSecurityExclusionFiltersApi
	auth context.Context
}

type applicationSecurityExclusionFiltersDataSourceModel struct {
	Id                  types.String                               `tfsdk:"id"`
	ExclusionFiltersIds types.List                                 `tfsdk:"exclusion_filters_ids"`
	ExclusionFilters    []applicationSecurityExclusionFiltersModel `tfsdk:"exclusion_filters"`
}

func NewApplicationSecurityExclusionFiltersDataSource() datasource.DataSource {
	return &applicationSecurityExclusionFiltersDataSource{}
}

func (r *applicationSecurityExclusionFiltersDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetApplicationSecurityExclusionFiltersApiV2()
	r.auth = providerData.Auth
}

func (*applicationSecurityExclusionFiltersDataSource) Metadata(_ context.Context, _ datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "application_security_exclusion_filters"
}

func (r *applicationSecurityExclusionFiltersDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state applicationSecurityExclusionFiltersDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, _, err := r.api.ListApplicationSecurityExclusionFilters(r.auth)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error while fetching exclusion filters"))
		return
	}

	data := res.GetData()
	exclusionFilterIds := make([]string, len(data))
	exclusionFilters := make([]applicationSecurityExclusionFiltersModel, len(data))

	for idx, exclusionFilter := range res.GetData() {
		var exclusionFilterModel applicationSecurityExclusionFiltersModel
		exclusionFilterModel.Id = types.StringValue(exclusionFilter.GetId())

		attributes := exclusionFilter.GetAttributes()
		exclusionFilterModel.Description = types.StringValue(attributes.GetDescription())
		exclusionFilterModel.Enabled = types.BoolValue(attributes.GetEnabled())
		exclusionFilterModel.PathGlob = types.StringValue(attributes.GetPathGlob())

		var parameters []attr.Value
		for _, param := range attributes.GetParameters() {
			parameters = append(parameters, types.StringValue(param))
		}
		tfParameters, diags := types.ListValue(types.StringType, parameters)
		response.Diagnostics.Append(diags...)
		exclusionFilterModel.Parameters = tfParameters

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

		var rulesTargets []attr.Value
		for _, ruleTarget := range attributes.GetRulesTarget() {
			tags, tagsOk := ruleTarget.GetTagsOk()
			if tagsOk && tags != nil {
				ruleTargetObject, diags := types.ObjectValue(map[string]attr.Type{
					"category": types.StringType,
					"type":     types.StringType,
				}, map[string]attr.Value{
					"category": types.StringValue(tags.GetCategory()),
					"type":     types.StringValue(tags.GetType()),
				})
				response.Diagnostics.Append(diags...)
				rulesTargets = append(rulesTargets, ruleTargetObject)
			}
		}
		tfRulesTargets, diags := types.ListValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"category": types.StringType,
				"type":     types.StringType,
			},
		}, rulesTargets)
		response.Diagnostics.Append(diags...)
		exclusionFilterModel.RulesTarget = tfRulesTargets

		exclusionFilterIds[idx] = exclusionFilter.GetId()
		exclusionFilters[idx] = exclusionFilterModel
	}

	stateId := strings.Join(exclusionFilterIds, "--")
	state.Id = types.StringValue(computeExclusionFiltersDataSourceID(&stateId))
	tfExclusionFilterIds, diags := types.ListValueFrom(ctx, types.StringType, exclusionFilterIds)
	response.Diagnostics.Append(diags...)
	state.ExclusionFiltersIds = tfExclusionFilterIds
	state.ExclusionFilters = exclusionFilters

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func computeExclusionFiltersDataSourceID(exclusionFiltersIds *string) string {
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

func (r *applicationSecurityExclusionFiltersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Retrieves Datadog Application Security Exclusion Filters.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"exclusion_filters_ids": schema.ListAttribute{
				Computed:    true,
				Description: "List of IDs for the Application Security exclusion filters.",
				ElementType: types.StringType,
			},
			"exclusion_filters": schema.ListAttribute{
				Computed:    true,
				Description: "List of Application Security exclusion filters",
				ElementType: types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"id":          types.StringType,
						"description": types.StringType,
						"enabled":     types.BoolType,
						"path_glob":   types.StringType,
						"parameters": types.ListType{
							ElemType: types.StringType,
						},
						"scope": types.ListType{ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"env":     types.StringType,
								"service": types.StringType,
							},
						}},
						"rules_target": types.ListType{ElemType: types.ObjectType{
							AttrTypes: map[string]attr.Type{
								"category": types.StringType,
								"type":     types.StringType,
							},
						}},
					},
				},
			},
		},
	}
}
