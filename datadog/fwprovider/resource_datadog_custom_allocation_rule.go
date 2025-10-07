package fwprovider

import (
	"context"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &datadogCustomAllocationRuleResource{}
	_ resource.ResourceWithImportState = &datadogCustomAllocationRuleResource{}
)

type datadogCustomAllocationRuleResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

type datadogCustomAllocationRuleModel struct {
	ID              types.String            `tfsdk:"id"`
	Enabled         types.Bool              `tfsdk:"enabled"`
	OrderId         types.Int64             `tfsdk:"order_id"`
	RuleName        types.String            `tfsdk:"rule_name"`
	Type            types.String            `tfsdk:"type"`
	Provider        types.List              `tfsdk:"providernames"`
	CostsToAllocate []*costsToAllocateModel `tfsdk:"costs_to_allocate"`
	Strategy        *strategyModel          `tfsdk:"strategy"`
	// Computed fields
	Rejected             types.Bool   `tfsdk:"rejected"`
	Created              types.String `tfsdk:"created"`
	LastModifiedUserUuid types.String `tfsdk:"last_modified_user_uuid"`
	Updated              types.String `tfsdk:"updated"`
	Version              types.Int64  `tfsdk:"version"`
}

type costsToAllocateModel struct {
	Condition types.String `tfsdk:"condition"`
	Tag       types.String `tfsdk:"tag"`
	Value     types.String `tfsdk:"value"`
	Values    types.List   `tfsdk:"values"`
}

type strategyModel struct {
	Granularity              types.String                     `tfsdk:"granularity"`
	Method                   types.String                     `tfsdk:"method"`
	AllocatedByTagKeys       types.List                       `tfsdk:"allocated_by_tag_keys"`
	EvaluateGroupedByTagKeys types.List                       `tfsdk:"evaluate_grouped_by_tag_keys"`
	AllocatedBy              []*allocatedByModel              `tfsdk:"allocated_by"`
	AllocatedByFilters       []*allocatedByFiltersModel       `tfsdk:"allocated_by_filters"`
	BasedOnCosts             []*basedOnCostsModel             `tfsdk:"based_on_costs"`
	EvaluateGroupedByFilters []*evaluateGroupedByFiltersModel `tfsdk:"evaluate_grouped_by_filters"`
	BasedOnTimeseries        *basedOnTimeseriesModel          `tfsdk:"based_on_timeseries"`
}
type allocatedByModel struct {
	Percentage    types.Int64           `tfsdk:"percentage"`
	AllocatedTags []*allocatedTagsModel `tfsdk:"allocated_tags"`
}
type allocatedTagsModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type allocatedByFiltersModel struct {
	Condition types.String `tfsdk:"condition"`
	Tag       types.String `tfsdk:"tag"`
	Value     types.String `tfsdk:"value"`
	Values    types.List   `tfsdk:"values"`
}

type basedOnCostsModel struct {
	Condition types.String `tfsdk:"condition"`
	Tag       types.String `tfsdk:"tag"`
	Value     types.String `tfsdk:"value"`
	Values    types.List   `tfsdk:"values"`
}

type evaluateGroupedByFiltersModel struct {
	Condition types.String `tfsdk:"condition"`
	Tag       types.String `tfsdk:"tag"`
	Value     types.String `tfsdk:"value"`
	Values    types.List   `tfsdk:"values"`
}

type basedOnTimeseriesModel struct {
}

// filterValueListValidator validates each filter in a list
type filterValueListValidator struct{}

func (v filterValueListValidator) Description(ctx context.Context) string {
	return "Ensures that 'values' is used with 'in'/'not in' operators and 'value' is used with all other operators"
}

func (v filterValueListValidator) MarkdownDescription(ctx context.Context) string {
	return "Ensures that `values` is used with `in`/`not in` operators and `value` is used with all other operators"
}

func (v filterValueListValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	// Get list elements as a slice of Objects
	elements := req.ConfigValue.Elements()

	// Validate each element
	for idx, element := range elements {
		objVal, ok := element.(types.Object)
		if !ok || objVal.IsNull() || objVal.IsUnknown() {
			continue
		}

		attrs := objVal.Attributes()

		// Extract condition, value, and values
		conditionAttr, hasCondition := attrs["condition"]
		valueAttr, hasValue := attrs["value"]
		valuesAttr, hasValues := attrs["values"]

		if !hasCondition {
			continue
		}

		condition, ok := conditionAttr.(types.String)
		if !ok || condition.IsNull() || condition.IsUnknown() {
			continue
		}

		conditionStr := condition.ValueString()

		// Check if value is set
		valueSet := false
		if hasValue {
			if v, ok := valueAttr.(types.String); ok && !v.IsNull() && !v.IsUnknown() {
				valueSet = true
			}
		}

		// Check if values is set
		valuesSet := false
		if hasValues {
			if v, ok := valuesAttr.(types.List); ok && !v.IsNull() && !v.IsUnknown() {
				valuesSet = true
			}
		}

		// Validate based on condition
		multiValueOperators := []string{"in", "not in"}
		isMultiValueOp := false
		for _, op := range multiValueOperators {
			if conditionStr == op {
				isMultiValueOp = true
				break
			}
		}

		if isMultiValueOp {
			// For 'in' and 'not in', values must be set and value must not be set
			if !valuesSet {
				resp.Diagnostics.AddAttributeError(
					req.Path.AtListIndex(idx),
					"Invalid Filter Configuration",
					"When condition is 'in' or 'not in', the 'values' field must be set (not 'value')",
				)
			}
			if valueSet {
				resp.Diagnostics.AddAttributeError(
					req.Path.AtListIndex(idx),
					"Invalid Filter Configuration",
					"When condition is 'in' or 'not in', only 'values' should be set (not both 'value' and 'values')",
				)
			}
		} else {
			// For all other operators, value must be set and values must not be set
			if !valueSet {
				resp.Diagnostics.AddAttributeError(
					req.Path.AtListIndex(idx),
					"Invalid Filter Configuration",
					"When condition is not 'in' or 'not in', the 'value' field must be set (not 'values')",
				)
			}
			if valuesSet {
				resp.Diagnostics.AddAttributeError(
					req.Path.AtListIndex(idx),
					"Invalid Filter Configuration",
					"When condition is not 'in' or 'not in', only 'value' should be set (not both 'value' and 'values')",
				)
			}
		}
	}
}

func NewDatadogCustomAllocationRuleResource() resource.Resource {
	return &datadogCustomAllocationRuleResource{}
}

func (r *datadogCustomAllocationRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData, _ := request.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *datadogCustomAllocationRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "custom_allocation_rule"
}

func (r *datadogCustomAllocationRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog DatadogCustomAllocationRule resource. This can be used to create and manage Datadog datadog_custom_allocation_rule.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "The `attributes` `enabled`. Whether the rule is enabled.",
			},
			"order_id": schema.Int64Attribute{
				Computed:    true,
				Description: "The `attributes` `order_id`. This field is read-only and returned by the API. Use the `datadog_custom_allocation_rule_order` resource to manage the order of rules.",
			},
			"rejected": schema.BoolAttribute{
				Computed:    true,
				Description: "The `attributes` `rejected`. This field is read-only and returned by the API after a rule was created, if it failed to apply.",
			},
			"created": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `created`. The timestamp when the rule was created.",
			},
			"last_modified_user_uuid": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `last_modified_user_uuid`. The UUID of the user who last modified the rule.",
			},
			"updated": schema.StringAttribute{
				Computed:    true,
				Description: "The `attributes` `updated`. The timestamp of the last update.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The `attributes` `version`. The rule version number of the rule. Can be used in the `datadog_custom_allocation_rule_order` resource to manage the order of rules.",
			},
			"rule_name": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `rule_name`. This field is immutable - changing it will force replacement of the resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				Required:    true,
				Description: "The `attributes` `type`. The type of the rule.",
			},
			"providernames": schema.ListAttribute{
				Required:    true,
				Description: "The `attributes` `provider`. The cloud providers the rule should apply to.",
				ElementType: types.StringType,
			},
			"id": utils.ResourceIDAttribute(),
		},
		Blocks: map[string]schema.Block{
			"costs_to_allocate": schema.ListNestedBlock{
				Validators: []validator.List{
					filterValueListValidator{},
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"condition": schema.StringAttribute{
							Optional:    true,
							Description: "The `items` `condition`.",
						},
						"tag": schema.StringAttribute{
							Optional:    true,
							Description: "The `items` `tag`.",
						},
						"value": schema.StringAttribute{
							Optional:    true,
							Description: "The `items` `value`. Use this for single-value conditions (not 'in'/'not in').",
						},
						"values": schema.ListAttribute{
							Optional:    true,
							Description: "The `items` `values`. Use this for multi-value conditions ('in'/'not in').",
							ElementType: types.StringType,
						},
					},
				},
			},
			"strategy": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"granularity": schema.StringAttribute{
						Optional:    true,
						Description: "The `strategy` `granularity`.",
					},
					"method": schema.StringAttribute{
						Optional:    true,
						Description: "The `strategy` `method`.",
					},
					"allocated_by_tag_keys": schema.ListAttribute{
						Optional:    true,
						Description: "The `strategy` `allocated_by_tag_keys`.",
						ElementType: types.StringType,
					},
					"evaluate_grouped_by_tag_keys": schema.ListAttribute{
						Optional:    true,
						Description: "The `strategy` `evaluate_grouped_by_tag_keys`.",
						ElementType: types.StringType,
					},
				},
				Blocks: map[string]schema.Block{
					"allocated_by": schema.ListNestedBlock{
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"percentage": schema.Int64Attribute{
									Optional:    true,
									Description: "The `items` `percentage`. The numeric value format should be a 32bit float value.",
								},
							},
							Blocks: map[string]schema.Block{
								"allocated_tags": schema.ListNestedBlock{
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"key": schema.StringAttribute{
												Optional:    true,
												Description: "The `items` `key`.",
											},
											"value": schema.StringAttribute{
												Optional:    true,
												Description: "The `items` `value`.",
											},
										},
									},
								},
							},
						},
					},
					"allocated_by_filters": schema.ListNestedBlock{
						Validators: []validator.List{
							filterValueListValidator{},
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"condition": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `condition`.",
								},
								"tag": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `tag`.",
								},
								"value": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `value`.",
								},
								"values": schema.ListAttribute{
									Optional:    true,
									Description: "The `items` `values`.",
									ElementType: types.StringType,
								},
							},
						},
					},
					"based_on_costs": schema.ListNestedBlock{
						Validators: []validator.List{
							filterValueListValidator{},
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"condition": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `condition`.",
								},
								"tag": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `tag`.",
								},
								"value": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `value`.",
								},
								"values": schema.ListAttribute{
									Optional:    true,
									Description: "The `items` `values`.",
									ElementType: types.StringType,
								},
							},
						},
					},
					"evaluate_grouped_by_filters": schema.ListNestedBlock{
						Validators: []validator.List{
							filterValueListValidator{},
						},
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"condition": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `condition`.",
								},
								"tag": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `tag`.",
								},
								"value": schema.StringAttribute{
									Optional:    true,
									Description: "The `items` `value`.",
								},
								"values": schema.ListAttribute{
									Optional:    true,
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

func (r *datadogCustomAllocationRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), request, response)
}

func (r *datadogCustomAllocationRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state datadogCustomAllocationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		response.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	resp, httpResp, err := r.Api.GetArbitraryCostRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DatadogCustomAllocationRule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}

	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datadogCustomAllocationRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var state datadogCustomAllocationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	body, diags := r.buildDatadogCustomAllocationRuleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.CreateArbitraryCostRule(r.Auth, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DatadogCustomAllocationRule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datadogCustomAllocationRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state datadogCustomAllocationRuleModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Get the current state to preserve order_id
	var currentState datadogCustomAllocationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &currentState)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		response.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	// Preserve the order_id from current state
	state.OrderId = currentState.OrderId

	body, diags := r.buildDatadogCustomAllocationRuleRequestBody(ctx, &state)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	resp, _, err := r.Api.UpdateArbitraryCostRule(r.Auth, id, *body)
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error retrieving DatadogCustomAllocationRule"))
		return
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		response.Diagnostics.AddError("response contains unparsedObject", err.Error())
		return
	}
	r.updateState(ctx, &state, &resp)

	// Save data into Terraform state
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *datadogCustomAllocationRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state datadogCustomAllocationRuleModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	id, err := strconv.ParseInt(state.ID.ValueString(), 10, 64)
	if err != nil {
		response.Diagnostics.AddError("Invalid ID", err.Error())
		return
	}

	httpResp, err := r.Api.DeleteArbitraryCostRule(r.Auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error deleting datadog_custom_allocation_rule"))
		return
	}
}

func (r *datadogCustomAllocationRuleResource) updateState(ctx context.Context, state *datadogCustomAllocationRuleModel, resp *datadogV2.ArbitraryRuleResponse) {
	if data, ok := resp.GetDataOk(); ok {
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
			} else {
				state.Rejected = types.BoolNull()
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
					if values, ok := costsToAllocateDd.GetValuesOk(); ok && values != nil && len(*values) > 0 {
						costsToAllocateTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
					} else {
						costsToAllocateTf.Values = types.ListNull(types.StringType)
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
						if values, ok := allocatedByFiltersDd.GetValuesOk(); ok && values != nil && len(*values) > 0 {

							allocatedByFiltersTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
						} else {
							allocatedByFiltersTf.Values = types.ListNull(types.StringType)
						}
						allocatedByFiltersTfItem = allocatedByFiltersTf

						strategyTf.AllocatedByFilters = append(strategyTf.AllocatedByFilters, &allocatedByFiltersTfItem)
					}
				}
				if allocatedByTagKeys, ok := strategy.GetAllocatedByTagKeysOk(); ok && len(*allocatedByTagKeys) > 0 {

					strategyTf.AllocatedByTagKeys, _ = types.ListValueFrom(ctx, types.StringType, *allocatedByTagKeys)
				} else {
					strategyTf.AllocatedByTagKeys = types.ListNull(types.StringType)
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
						if values, ok := basedOnCostsDd.GetValuesOk(); ok && values != nil && len(*values) > 0 {

							basedOnCostsTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
						} else {
							basedOnCostsTf.Values = types.ListNull(types.StringType)
						}
						basedOnCostsTfItem = basedOnCostsTf

						strategyTf.BasedOnCosts = append(strategyTf.BasedOnCosts, &basedOnCostsTfItem)
					}
				}
				if _, ok := strategy.GetBasedOnTimeseriesOk(); ok {
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
						if values, ok := evaluateGroupedByFiltersDd.GetValuesOk(); ok && values != nil && len(*values) > 0 {

							evaluateGroupedByFiltersTf.Values, _ = types.ListValueFrom(ctx, types.StringType, *values)
						} else {
							evaluateGroupedByFiltersTf.Values = types.ListNull(types.StringType)
						}
						evaluateGroupedByFiltersTfItem = evaluateGroupedByFiltersTf

						strategyTf.EvaluateGroupedByFilters = append(strategyTf.EvaluateGroupedByFilters, &evaluateGroupedByFiltersTfItem)
					}
				}
				if evaluateGroupedByTagKeys, ok := strategy.GetEvaluateGroupedByTagKeysOk(); ok && len(*evaluateGroupedByTagKeys) > 0 {

					strategyTf.EvaluateGroupedByTagKeys, _ = types.ListValueFrom(ctx, types.StringType, *evaluateGroupedByTagKeys)
				} else {
					strategyTf.EvaluateGroupedByTagKeys = types.ListNull(types.StringType)
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
}

func (r *datadogCustomAllocationRuleResource) buildDatadogCustomAllocationRuleRequestBody(ctx context.Context, state *datadogCustomAllocationRuleModel) (*datadogV2.ArbitraryCostUpsertRequest, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	req := &datadogV2.ArbitraryCostUpsertRequest{}
	attributes := datadogV2.NewArbitraryCostUpsertRequestDataAttributesWithDefaults()

	if !state.Enabled.IsNull() {
		attributes.SetEnabled(state.Enabled.ValueBool())
	}
	// Include OrderId to preserve rule order position during updates (but not on create)
	if !state.OrderId.IsNull() && state.OrderId.ValueInt64() > 0 {
		attributes.SetOrderId(state.OrderId.ValueInt64())
	}
	// Note: Rejected is a computed field and should not be sent in requests
	if !state.RuleName.IsNull() {
		attributes.SetRuleName(state.RuleName.ValueString())
	}
	if !state.Type.IsNull() {
		attributes.SetType(state.Type.ValueString())
	}

	if !state.Provider.IsNull() {
		var provider []string
		diags.Append(state.Provider.ElementsAs(ctx, &provider, false)...)
		attributes.SetProvider(provider)
	}

	var costsToAllocate []datadogV2.ArbitraryCostUpsertRequestDataAttributesCostsToAllocateItems
	for _, costsToAllocateTFItem := range state.CostsToAllocate {
		condition := costsToAllocateTFItem.Condition.ValueString()
		tag := costsToAllocateTFItem.Tag.ValueString()
		costsToAllocateDDItem := datadogV2.NewArbitraryCostUpsertRequestDataAttributesCostsToAllocateItems(condition, tag)
		if !costsToAllocateTFItem.Value.IsNull() {
			costsToAllocateDDItem.SetValue(costsToAllocateTFItem.Value.ValueString())
		}

		if !costsToAllocateTFItem.Values.IsNull() {
			var values []string
			diags.Append(costsToAllocateTFItem.Values.ElementsAs(ctx, &values, false)...)
			costsToAllocateDDItem.SetValues(values)
		}
		costsToAllocate = append(costsToAllocate, *costsToAllocateDDItem)
	}
	attributes.SetCostsToAllocate(costsToAllocate)

	if state.Strategy != nil {
		var strategy datadogV2.ArbitraryCostUpsertRequestDataAttributesStrategy

		strategy.SetGranularity(state.Strategy.Granularity.ValueString())
		strategy.SetMethod(state.Strategy.Method.ValueString())

		var allocatedByTagKeys []string
		diags.Append(state.Strategy.AllocatedByTagKeys.ElementsAs(ctx, &allocatedByTagKeys, false)...)
		strategy.SetAllocatedByTagKeys(allocatedByTagKeys)

		var evaluateGroupedByTagKeys []string
		diags.Append(state.Strategy.EvaluateGroupedByTagKeys.ElementsAs(ctx, &evaluateGroupedByTagKeys, false)...)
		strategy.SetEvaluateGroupedByTagKeys(evaluateGroupedByTagKeys)

		if state.Strategy.AllocatedBy != nil {
			var allocatedBy []datadogV2.ArbitraryCostUpsertRequestDataAttributesStrategyAllocatedByItems
			for _, allocatedByTFItem := range state.Strategy.AllocatedBy {
				var allocatedTags []datadogV2.ArbitraryCostUpsertRequestDataAttributesStrategyAllocatedByItemsAllocatedTagsItems
				if allocatedByTFItem.AllocatedTags != nil {
					for _, allocatedTagsTFItem := range allocatedByTFItem.AllocatedTags {
						key := allocatedTagsTFItem.Key.ValueString()
						value := allocatedTagsTFItem.Value.ValueString()
						allocatedTagsDDItem := datadogV2.NewArbitraryCostUpsertRequestDataAttributesStrategyAllocatedByItemsAllocatedTagsItems(key, value)
						allocatedTags = append(allocatedTags, *allocatedTagsDDItem)
					}
				}

				percentage := float64(allocatedByTFItem.Percentage.ValueInt64())
				allocatedByDDItem := datadogV2.NewArbitraryCostUpsertRequestDataAttributesStrategyAllocatedByItems(allocatedTags, percentage)
				allocatedBy = append(allocatedBy, *allocatedByDDItem)
			}
			strategy.SetAllocatedBy(allocatedBy)
		}

		if state.Strategy.AllocatedByFilters != nil {
			var allocatedByFilters []datadogV2.ArbitraryCostUpsertRequestDataAttributesStrategyAllocatedByFiltersItems
			for _, allocatedByFiltersTFItem := range state.Strategy.AllocatedByFilters {
				condition := allocatedByFiltersTFItem.Condition.ValueString()
				tag := allocatedByFiltersTFItem.Tag.ValueString()
				allocatedByFiltersDDItem := datadogV2.NewArbitraryCostUpsertRequestDataAttributesStrategyAllocatedByFiltersItems(condition, tag)

				if !allocatedByFiltersTFItem.Value.IsNull() {
					allocatedByFiltersDDItem.SetValue(allocatedByFiltersTFItem.Value.ValueString())
				}

				if !allocatedByFiltersTFItem.Values.IsNull() {
					var values []string
					diags.Append(allocatedByFiltersTFItem.Values.ElementsAs(ctx, &values, false)...)
					allocatedByFiltersDDItem.SetValues(values)
				}
				allocatedByFilters = append(allocatedByFilters, *allocatedByFiltersDDItem)
			}
			strategy.SetAllocatedByFilters(allocatedByFilters)
		}

		if state.Strategy.BasedOnCosts != nil {
			var basedOnCosts []datadogV2.ArbitraryCostUpsertRequestDataAttributesStrategyBasedOnCostsItems
			for _, basedOnCostsTFItem := range state.Strategy.BasedOnCosts {
				condition := basedOnCostsTFItem.Condition.ValueString()
				tag := basedOnCostsTFItem.Tag.ValueString()
				basedOnCostsDDItem := datadogV2.NewArbitraryCostUpsertRequestDataAttributesStrategyBasedOnCostsItems(condition, tag)

				if !basedOnCostsTFItem.Value.IsNull() {
					basedOnCostsDDItem.SetValue(basedOnCostsTFItem.Value.ValueString())
				}

				if !basedOnCostsTFItem.Values.IsNull() {
					var values []string
					diags.Append(basedOnCostsTFItem.Values.ElementsAs(ctx, &values, false)...)
					basedOnCostsDDItem.SetValues(values)
				}
				basedOnCosts = append(basedOnCosts, *basedOnCostsDDItem)
			}
			strategy.SetBasedOnCosts(basedOnCosts)
		}

		if state.Strategy.EvaluateGroupedByFilters != nil {
			var evaluateGroupedByFilters []datadogV2.ArbitraryCostUpsertRequestDataAttributesStrategyEvaluateGroupedByFiltersItems
			for _, evaluateGroupedByFiltersTFItem := range state.Strategy.EvaluateGroupedByFilters {
				condition := evaluateGroupedByFiltersTFItem.Condition.ValueString()
				tag := evaluateGroupedByFiltersTFItem.Tag.ValueString()
				evaluateGroupedByFiltersDDItem := datadogV2.NewArbitraryCostUpsertRequestDataAttributesStrategyEvaluateGroupedByFiltersItems(condition, tag)

				if !evaluateGroupedByFiltersTFItem.Value.IsNull() {
					evaluateGroupedByFiltersDDItem.SetValue(evaluateGroupedByFiltersTFItem.Value.ValueString())
				}

				if !evaluateGroupedByFiltersTFItem.Values.IsNull() {
					var values []string
					diags.Append(evaluateGroupedByFiltersTFItem.Values.ElementsAs(ctx, &values, false)...)
					evaluateGroupedByFiltersDDItem.SetValues(values)
				}
				evaluateGroupedByFilters = append(evaluateGroupedByFilters, *evaluateGroupedByFiltersDDItem)
			}
			strategy.SetEvaluateGroupedByFilters(evaluateGroupedByFilters)
		}

		var basedOnTimeseries map[string]interface{}
		strategy.BasedOnTimeseries = basedOnTimeseries
		attributes.SetStrategy(strategy)
	}

	req.Data = datadogV2.NewArbitraryCostUpsertRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req, diags
}
