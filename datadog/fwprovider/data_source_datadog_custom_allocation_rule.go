package fwprovider

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ datasource.DataSource = &datadogCustomAllocationRuleDataSource{}
)

type datadogCustomAllocationRuleDataSource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type datadogCustomAllocationRuleDataSourceModel struct {
	// Datasource ID
	ID types.String `tfsdk:"id"`

	// Query Parameters
	RuleId types.Int64 `tfsdk:"rule_id"`

	// Computed values
	Created              types.String            `tfsdk:"created"`
	Enabled              types.Bool              `tfsdk:"enabled"`
	LastModifiedUserUuid types.String            `tfsdk:"last_modified_user_uuid"`
	OrderId              types.Int64             `tfsdk:"order_id"`
	Rejected             types.Bool              `tfsdk:"rejected"`
	RuleName             types.String            `tfsdk:"rule_name"`
	Type                 types.String            `tfsdk:"type"`
	Updated              types.String            `tfsdk:"updated"`
	Version              types.Int64             `tfsdk:"version"`
	Provider             types.List              `tfsdk:"providernames"`
	CostsToAllocate      []*costsToAllocateModel `tfsdk:"costs_to_allocate"`
	Strategy             *strategyModel          `tfsdk:"strategy"`
}

func NewDatadogCustomAllocationRuleDataSource() datasource.DataSource {
	return &datadogCustomAllocationRuleDataSource{}
}

func (d *datadogCustomAllocationRuleDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	d.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	d.Auth = providerData.Auth
}

func (d *datadogCustomAllocationRuleDataSource) Metadata(_ context.Context, request datasource.MetadataRequest, response *datasource.MetadataResponse) {
	response.TypeName = "custom_allocation_rule"
}

func (d *datadogCustomAllocationRuleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, response *datasource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Use this data source to retrieve information about an existing Datadog datadog_custom_allocation_rule.",
		Attributes: map[string]schema.Attribute{
			// Datasource ID
			"id": utils.ResourceIDAttribute(),
			// Query Parameters
			"rule_id": schema.Int64Attribute{
				Optional:    true,
				Description: "UPDATE ME",
			},
			// Computed values
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `created`.",
			},
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "The `attributes` `enabled`.",
			},
			"last_modified_user_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `last_modified_user_uuid`.",
			},
			"order_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The `attributes` `order_id`.",
			},
			"rejected": schema.BoolAttribute{
				Computed:    true,
				Description: "The `attributes` `rejected`.",
			},
			"rule_name": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `rule_name`.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `type`.",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `updated`.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The `attributes` `version`.",
			},
			"providernames": schema.ListAttribute{
				Computed:    true,
				Description: "The `attributes` `provider`.",
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			// Computed values
			"costs_to_allocate": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"condition": schema.StringAttribute{
							Computed:    true,
							Description: "The `items` `condition`.",
						},
						"tag": schema.StringAttribute{
							Computed:    true,
							Description: "The `items` `tag`.",
						},
						"value": schema.StringAttribute{
							Computed:    true,
							Description: "The `items` `value`.",
						},
						"values": schema.ListAttribute{
							Computed:    true,
							Description: "The `items` `values`.",
							ElementType: types.StringType,
						},
					},
				},
			},
			"strategy": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"granularity": schema.StringAttribute{
						Computed:    true,
						Description: "The `strategy` `granularity`.",
					},
					"method": schema.StringAttribute{
						Computed:    true,
						Description: "The `strategy` `method`.",
					},
					"allocated_by_tag_keys": schema.ListAttribute{
						Computed:    true,
						Description: "The `strategy` `allocated_by_tag_keys`.",
						ElementType: types.StringType,
					},
					"evaluate_grouped_by_tag_keys": schema.ListAttribute{
						Computed:    true,
						Description: "The `strategy` `evaluate_grouped_by_tag_keys`.",
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					"allocated_by": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"percentage": schema.Int64Attribute{
									Computed:    true,
									Description: "The `items` `percentage`. The numeric value format should be a 32bit float value.",
								},
							},
							Blocks: map[string]schema.Block{
								"allocated_tags": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												Computed:    true,
												Description: "The `items` `key`.",
											},
											"value": schema.StringAttribute{
												Computed:    true,
												Description: "The `items` `value`.",
											},
										},
									},
								},
							},
						},
					},
					"allocated_by_filters": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"condition": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `condition`.",
								},
								"tag": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `tag`.",
								},
								"value": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `value`.",
								},
								"values": schema.ListAttribute{
									Computed:    true,
									Description: "The `items` `values`.",
									ElementType: types.StringType,
								},
							},
						},
					},
					"based_on_costs": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"condition": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `condition`.",
								},
								"tag": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `tag`.",
								},
								"value": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `value`.",
								},
								"values": schema.ListAttribute{
									Computed:    true,
									Description: "The `items` `values`.",
									ElementType: types.StringType,
								},
							},
						},
					},
					"evaluate_grouped_by_filters": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"condition": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `condition`.",
								},
								"tag": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `tag`.",
								},
								"value": schema.StringAttribute{
									Computed:    true,
									Description: "The `items` `value`.",
								},
								"values": schema.ListAttribute{
									Computed:    true,
									Description: "The `items` `values`.",
									ElementType: types.StringType,
								},
							},
						},
					},
					"based_on_timeseries": schema.SingleNestedBlock{
						Attributes: map[string]schema.Attribute{},
					},
				},
			},
		},
	}
}

func (d *datadogCustomAllocationRuleDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state datadogCustomAllocationRuleDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	ruleId := state.RuleId.ValueInt64()
	ddResp, _, err := d.Api.GetArbitraryCostRule(d.Auth, ruleId)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error getting datadog custom allocation rule"))
		return
	}

	if data, ok := ddResp.GetDataOk(); ok {
		d.updateState(ctx, &state, data)
	}

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (d *datadogCustomAllocationRuleDataSource) updateState(ctx context.Context, state *datadogCustomAllocationRuleDataSourceModel, data *datadogV2.ArbitraryRuleResponseData) {
	if id, ok := data.GetIdOk(); ok {
		state.ID = types.StringValue(*id)
	}

	if attributes, ok := data.GetAttributesOk(); ok {
		if created, ok := attributes.GetCreatedOk(); ok {
			state.Created = types.StringValue(created.String())
		}

		if enabled, ok := attributes.GetEnabledOk(); ok {
			state.Enabled = types.BoolValue(*enabled)
		}

		if lastModifiedUserUuid, ok := attributes.GetLastModifiedUserUuidOk(); ok {
			state.LastModifiedUserUuid = types.StringValue(*lastModifiedUserUuid)
		}

		if orderId, ok := attributes.GetOrderIdOk(); ok {
			state.OrderId = types.Int64Value(*orderId)
		}

		if rejected, ok := attributes.GetRejectedOk(); ok {
			state.Rejected = types.BoolValue(*rejected)
		}

		if ruleName, ok := attributes.GetRuleNameOk(); ok {
			state.RuleName = types.StringValue(*ruleName)
		}

		if typeVar, ok := attributes.GetTypeOk(); ok {
			state.Type = types.StringValue(*typeVar)
		}

		if updated, ok := attributes.GetUpdatedOk(); ok {
			state.Updated = types.StringValue(updated.String())
		}

		if version, ok := attributes.GetVersionOk(); ok {
			state.Version = types.Int64Value(int64(*version))
		}

		if provider, ok := attributes.GetProviderOk(); ok && len(*provider) > 0 {
			state.Provider, _ = types.ListValueFrom(ctx, types.StringType, *provider)
		}

		if costsToAllocate, ok := attributes.GetCostsToAllocateOk(); ok && len(*costsToAllocate) > 0 {
			state.CostsToAllocate = []*costsToAllocateModel{}
			for _, costsToAllocateDd := range *costsToAllocate {
				costsToAllocateTfItem := costsToAllocateModel{}

				costsToAllocateTf := costsToAllocateModel{}
				if condition, ok := costsToAllocateDd.GetConditionOk(); ok {
					costsToAllocateTf.Condition = types.StringValue(*condition)
				}
				if tag, ok := costsToAllocateDd.GetTagOk(); ok {
					costsToAllocateTf.Tag = types.StringValue(*tag)
				}
				if value, ok := costsToAllocateDd.GetValueOk(); ok {
					costsToAllocateTf.Value = types.StringValue(*value)
				}
				if values, ok := costsToAllocateDd.GetValuesOk(); ok && len(*values) > 0 {
					costsToAllocateTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
				}
				costsToAllocateTfItem = costsToAllocateTf

				state.CostsToAllocate = append(state.CostsToAllocate, &costsToAllocateTfItem)
			}
		}

		if strategy, ok := attributes.GetStrategyOk(); ok {
			strategyTf := strategyModel{}
			if allocatedBy, ok := strategy.GetAllocatedByOk(); ok && len(*allocatedBy) > 0 {
				strategyTf.AllocatedBy = []*allocatedByModel{}
				for _, allocatedByDd := range *allocatedBy {
					allocatedByTfItem := allocatedByModel{}

					allocatedByTf := allocatedByModel{}
					if allocatedTags, ok := allocatedByDd.GetAllocatedTagsOk(); ok && len(*allocatedTags) > 0 {
						allocatedByTf.AllocatedTags = []*allocatedTagsModel{}
						for _, allocatedTagsDd := range *allocatedTags {
							allocatedTagsTfItem := allocatedTagsModel{}

							allocatedTagsTf := allocatedTagsModel{}
							if key, ok := allocatedTagsDd.GetKeyOk(); ok {
								allocatedTagsTf.Key = types.StringValue(*key)
							}
							if value, ok := allocatedTagsDd.GetValueOk(); ok {
								allocatedTagsTf.Value = types.StringValue(*value)
							}
							allocatedTagsTfItem = allocatedTagsTf

							allocatedByTf.AllocatedTags = append(allocatedByTf.AllocatedTags, &allocatedTagsTfItem)
						}
					}
					if percentage, ok := allocatedByDd.GetPercentageOk(); ok {
						allocatedByTf.Percentage = types.Int64Value(int64(*percentage))
					}
					allocatedByTfItem = allocatedByTf

					strategyTf.AllocatedBy = append(strategyTf.AllocatedBy, &allocatedByTfItem)
				}
			}
			if allocatedByFilters, ok := strategy.GetAllocatedByFiltersOk(); ok && len(*allocatedByFilters) > 0 {
				strategyTf.AllocatedByFilters = []*allocatedByFiltersModel{}
				for _, allocatedByFiltersDd := range *allocatedByFilters {
					allocatedByFiltersTfItem := allocatedByFiltersModel{}

					allocatedByFiltersTf := allocatedByFiltersModel{}
					if condition, ok := allocatedByFiltersDd.GetConditionOk(); ok {
						allocatedByFiltersTf.Condition = types.StringValue(*condition)
					}
					if tag, ok := allocatedByFiltersDd.GetTagOk(); ok {
						allocatedByFiltersTf.Tag = types.StringValue(*tag)
					}
					if value, ok := allocatedByFiltersDd.GetValueOk(); ok {
						allocatedByFiltersTf.Value = types.StringValue(*value)
					}
					if values, ok := allocatedByFiltersDd.GetValuesOk(); ok && len(*values) > 0 {
						allocatedByFiltersTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
					}
					allocatedByFiltersTfItem = allocatedByFiltersTf

					strategyTf.AllocatedByFilters = append(strategyTf.AllocatedByFilters, &allocatedByFiltersTfItem)
				}
			}
			if allocatedByTagKeys, ok := strategy.GetAllocatedByTagKeysOk(); ok && len(*allocatedByTagKeys) > 0 {
				strategyTf.AllocatedByTagKeys, _ = types.ListValueFrom(ctx, types.StringType, *allocatedByTagKeys)
			}
			if basedOnCosts, ok := strategy.GetBasedOnCostsOk(); ok && len(*basedOnCosts) > 0 {
				strategyTf.BasedOnCosts = []*basedOnCostsModel{}
				for _, basedOnCostsDd := range *basedOnCosts {
					basedOnCostsTfItem := basedOnCostsModel{}

					basedOnCostsTf := basedOnCostsModel{}
					if condition, ok := basedOnCostsDd.GetConditionOk(); ok {
						basedOnCostsTf.Condition = types.StringValue(*condition)
					}
					if tag, ok := basedOnCostsDd.GetTagOk(); ok {
						basedOnCostsTf.Tag = types.StringValue(*tag)
					}
					if value, ok := basedOnCostsDd.GetValueOk(); ok {
						basedOnCostsTf.Value = types.StringValue(*value)
					}
					if values, ok := basedOnCostsDd.GetValuesOk(); ok && len(*values) > 0 {
						basedOnCostsTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
					}
					basedOnCostsTfItem = basedOnCostsTf

					strategyTf.BasedOnCosts = append(strategyTf.BasedOnCosts, &basedOnCostsTfItem)
				}
			}
			if basedOnTimeseries, ok := strategy.GetBasedOnTimeseriesOk(); ok {
				_ = basedOnTimeseries
				basedOnTimeseriesTf := basedOnTimeseriesModel{}
				strategyTf.BasedOnTimeseries = &basedOnTimeseriesTf
			}
			if evaluateGroupedByFilters, ok := strategy.GetEvaluateGroupedByFiltersOk(); ok && len(*evaluateGroupedByFilters) > 0 {
				strategyTf.EvaluateGroupedByFilters = []*evaluateGroupedByFiltersModel{}
				for _, evaluateGroupedByFiltersDd := range *evaluateGroupedByFilters {
					evaluateGroupedByFiltersTfItem := evaluateGroupedByFiltersModel{}

					evaluateGroupedByFiltersTf := evaluateGroupedByFiltersModel{}
					if condition, ok := evaluateGroupedByFiltersDd.GetConditionOk(); ok {
						evaluateGroupedByFiltersTf.Condition = types.StringValue(*condition)
					}
					if tag, ok := evaluateGroupedByFiltersDd.GetTagOk(); ok {
						evaluateGroupedByFiltersTf.Tag = types.StringValue(*tag)
					}
					if value, ok := evaluateGroupedByFiltersDd.GetValueOk(); ok {
						evaluateGroupedByFiltersTf.Value = types.StringValue(*value)
					}
					if values, ok := evaluateGroupedByFiltersDd.GetValuesOk(); ok && len(*values) > 0 {
						evaluateGroupedByFiltersTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
					}
					evaluateGroupedByFiltersTfItem = evaluateGroupedByFiltersTf

					strategyTf.EvaluateGroupedByFilters = append(strategyTf.EvaluateGroupedByFilters, &evaluateGroupedByFiltersTfItem)
				}
			}
			if evaluateGroupedByTagKeys, ok := strategy.GetEvaluateGroupedByTagKeysOk(); ok && len(*evaluateGroupedByTagKeys) > 0 {
				strategyTf.EvaluateGroupedByTagKeys, _ = types.ListValueFrom(ctx, types.StringType, *evaluateGroupedByTagKeys)
			}
			if granularity, ok := strategy.GetGranularityOk(); ok {
				strategyTf.Granularity = types.StringValue(*granularity)
			}
			if method, ok := strategy.GetMethodOk(); ok {
				strategyTf.Method = types.StringValue(*method)
			}

			state.Strategy = &strategyTf
		}
	}
}
