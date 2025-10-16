package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type tagPipelineRulesetDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func NewTagPipelineRulesetDataSource() datasource.DataSource {
	return &tagPipelineRulesetDataSource{}
}

type tagPipelineRulesetDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Position types.Int64  `tfsdk:"position"`
	Version  types.Int64  `tfsdk:"version"`
	Rules    []ruleItem   `tfsdk:"rules"`
}

func (d *tagPipelineRulesetDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "tag_pipeline_ruleset"
}

func (d *tagPipelineRulesetDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog tag pipeline ruleset.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The ID of the ruleset.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the ruleset.",
				Computed:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the ruleset is enabled.",
				Computed:    true,
			},
			"position": schema.Int64Attribute{
				Description: "The position of the ruleset in the pipeline.",
				Computed:    true,
			},
			"version": schema.Int64Attribute{
				Description: "The version of the ruleset.",
				Computed:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"rules": schema.ListNestedBlock{
				Description: "The rules in the ruleset.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Description: "Whether the rule is enabled.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the rule.",
							Computed:    true,
						},
						"metadata": schema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Rule metadata key-value pairs.",
						},
					},
					Blocks: map[string]schema.Block{
						"mapping": schema.SingleNestedBlock{
							Description: "The mapping configuration for the rule.",
							Attributes: map[string]schema.Attribute{
								"destination_key": schema.StringAttribute{
									Description: "The destination key for the mapping.",
									Computed:    true,
								},
								"if_not_exists": schema.BoolAttribute{
									Description: "Whether to apply the mapping only if the destination key doesn't exist.",
									Computed:    true,
								},
								"source_keys": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "The source keys for the mapping.",
									Computed:    true,
								},
							},
						},
						"query": schema.SingleNestedBlock{
							Description: "The query configuration for the rule.",
							Attributes: map[string]schema.Attribute{
								"case_insensitivity": schema.BoolAttribute{
									Description: "Whether the query matching is case insensitive.",
									Computed:    true,
								},
								"if_not_exists": schema.BoolAttribute{
									Description: "Whether to apply the query only if the key doesn't exist.",
									Computed:    true,
								},
								"query": schema.StringAttribute{
									Description: "The query string.",
									Computed:    true,
								},
							},
							Blocks: map[string]schema.Block{
								"addition": schema.SingleNestedBlock{
									Description: "The addition configuration for the query.",
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Description: "The key to add.",
											Computed:    true,
										},
										"value": schema.StringAttribute{
											Description: "The value to add.",
											Computed:    true,
										},
									},
								},
							},
						},
						"reference_table": schema.SingleNestedBlock{
							Description: "The reference table configuration for the rule.",
							Attributes: map[string]schema.Attribute{
								"case_insensitivity": schema.BoolAttribute{
									Description: "Whether the reference table lookup is case insensitive.",
									Computed:    true,
								},
								"if_not_exists": schema.BoolAttribute{
									Description: "Whether to apply the reference table only if the key doesn't exist.",
									Computed:    true,
								},
								"source_keys": schema.ListAttribute{
									ElementType: types.StringType,
									Description: "The source keys for the reference table lookup.",
									Computed:    true,
								},
								"table_name": schema.StringAttribute{
									Description: "The name of the reference table.",
									Computed:    true,
								},
							},
							Blocks: map[string]schema.Block{
								"field_pairs": schema.ListNestedBlock{
									Description: "The field pairs for the reference table.",
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"input_column": schema.StringAttribute{
												Description: "The input column name.",
												Computed:    true,
											},
											"output_key": schema.StringAttribute{
												Description: "The output key name.",
												Computed:    true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *tagPipelineRulesetDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *tagPipelineRulesetDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tagPipelineRulesetDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, _, err := d.Api.GetTagPipelinesRuleset(d.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading ruleset", err.Error())
		return
	}

	setDataSourceModelFromRulesetResp(&state, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setDataSourceModelFromRulesetResp(model *tagPipelineRulesetDataSourceModel, apiResp datadogV2.RulesetResp) {
	if apiResp.Data == nil || apiResp.Data.Attributes == nil {
		return
	}
	data := apiResp.Data
	attr := data.Attributes

	// Set top-level fields
	if data.Id != nil && *data.Id != "" {
		model.ID = types.StringValue(*data.Id)
	} else {
		model.ID = types.StringValue("")
	}
	// The ruleset name comes from attributes.name, not from data.Id
	if attr.Name != "" {
		model.Name = types.StringValue(attr.Name)
	} else {
		// Fallback: if attributes.name is empty, try using the data.Id (original name)
		if data.Id != nil {
			model.Name = types.StringValue(*data.Id)
		}
	}
	model.Enabled = types.BoolValue(attr.Enabled)
	model.Position = types.Int64Value(int64(attr.Position))
	model.Version = types.Int64Value(attr.Version)

	// Set rules
	var rules []ruleItem
	for _, apiRule := range attr.Rules {
		rule := ruleItem{
			Enabled: types.BoolValue(apiRule.Enabled),
			Name:    types.StringValue(apiRule.Name),
		}

		// Set metadata if present
		if len(apiRule.Metadata) > 0 {
			metadata := make(map[string]types.String)
			for k, v := range apiRule.Metadata {
				metadata[k] = types.StringValue(v)
			}
			mapValue, diags := types.MapValueFrom(context.Background(), types.StringType, metadata)
			if diags.HasError() {
				// Handle error - for now just set null
				rule.Metadata = types.MapNull(types.StringType)
			} else {
				rule.Metadata = mapValue
			}
		} else {
			// Set empty map
			rule.Metadata = types.MapNull(types.StringType)
		}

		// Set mapping if present
		if apiRule.Mapping.IsSet() {
			mappingVal := apiRule.Mapping.Get()
			if mappingVal != nil {
				sourceKeys := make([]types.String, len(mappingVal.SourceKeys))
				for i, sk := range mappingVal.SourceKeys {
					sourceKeys[i] = types.StringValue(sk)
				}
				rule.Mapping = &ruleMapping{
					DestinationKey: types.StringValue(mappingVal.DestinationKey),
					IfNotExists:    types.BoolValue(mappingVal.IfNotExists),
					SourceKeys:     sourceKeys,
				}
			}
		}

		// Set query if present
		if apiRule.Query.IsSet() {
			queryVal := apiRule.Query.Get()
			if queryVal != nil {
				query := &ruleQuery{
					CaseInsensitivity: func() types.Bool {
						if queryVal.CaseInsensitivity != nil {
							return types.BoolValue(*queryVal.CaseInsensitivity)
						}
						return types.BoolNull()
					}(),
					IfNotExists: types.BoolValue(queryVal.IfNotExists),
					Query:       types.StringValue(queryVal.Query),
				}
				if queryVal.Addition.IsSet() {
					additionVal := queryVal.Addition.Get()
					if additionVal != nil {
						query.Addition = &queryAddition{
							Key:   types.StringValue(additionVal.Key),
							Value: types.StringValue(additionVal.Value),
						}
					}
				}
				rule.Query = query
			}
		}

		// Set reference table if present
		if apiRule.ReferenceTable.IsSet() {
			refTableVal := apiRule.ReferenceTable.Get()
			if refTableVal != nil {
				var fieldPairs []referenceTableFieldPair
				for _, fp := range refTableVal.FieldPairs {
					fieldPairs = append(fieldPairs, referenceTableFieldPair{
						InputColumn: types.StringValue(fp.InputColumn),
						OutputKey:   types.StringValue(fp.OutputKey),
					})
				}
				sourceKeys := make([]types.String, len(refTableVal.SourceKeys))
				for i, sk := range refTableVal.SourceKeys {
					sourceKeys[i] = types.StringValue(sk)
				}
				rule.ReferenceTable = &referenceTable{
					CaseInsensitivity: func() types.Bool {
						if refTableVal.CaseInsensitivity != nil {
							return types.BoolValue(*refTableVal.CaseInsensitivity)
						}
						return types.BoolNull()
					}(),
					FieldPairs: fieldPairs,
					IfNotExists: func() types.Bool {
						if refTableVal.IfNotExists != nil {
							return types.BoolValue(*refTableVal.IfNotExists)
						}
						return types.BoolNull()
					}(),
					SourceKeys: sourceKeys,
					TableName:  types.StringValue(refTableVal.TableName),
				}
			}
		}

		rules = append(rules, rule)
	}
	model.Rules = rules
}
