package fwprovider

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

var (
	_ resource.ResourceWithConfigure   = &tagPipelineRulesetResource{}
	_ resource.ResourceWithImportState = &tagPipelineRulesetResource{}
	_ resource.ResourceWithModifyPlan  = &tagPipelineRulesetResource{}
)

type tagPipelineRulesetResource struct {
	Api  *datadogV2.CloudCostManagementApi
	Auth context.Context
}

func NewTagPipelineRulesetResource() resource.Resource {
	return &tagPipelineRulesetResource{}
}

type tagPipelineRulesetModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Enabled  types.Bool   `tfsdk:"enabled"`
	Position types.Int64  `tfsdk:"position"`
	Version  types.Int64  `tfsdk:"version"`
	Rules    []ruleItem   `tfsdk:"rules"`
}

type ruleItem struct {
	Enabled        types.Bool      `tfsdk:"enabled"`
	Name           types.String    `tfsdk:"name"`
	Metadata       types.Map       `tfsdk:"metadata"`
	Mapping        *ruleMapping    `tfsdk:"mapping"`
	Query          *ruleQuery      `tfsdk:"query"`
	ReferenceTable *referenceTable `tfsdk:"reference_table"`
}

type ruleMapping struct {
	DestinationKey types.String   `tfsdk:"destination_key"`
	IfNotExists    types.Bool     `tfsdk:"if_not_exists"`
	SourceKeys     []types.String `tfsdk:"source_keys"`
}

type ruleQuery struct {
	Addition          *queryAddition `tfsdk:"addition"`
	CaseInsensitivity types.Bool     `tfsdk:"case_insensitivity"`
	IfNotExists       types.Bool     `tfsdk:"if_not_exists"`
	Query             types.String   `tfsdk:"query"`
}

type queryAddition struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type referenceTable struct {
	CaseInsensitivity types.Bool                `tfsdk:"case_insensitivity"`
	FieldPairs        []referenceTableFieldPair `tfsdk:"field_pairs"`
	IfNotExists       types.Bool                `tfsdk:"if_not_exists"`
	SourceKeys        []types.String            `tfsdk:"source_keys"`
	TableName         types.String              `tfsdk:"table_name"`
}

type referenceTableFieldPair struct {
	InputColumn types.String `tfsdk:"input_column"`
	OutputKey   types.String `tfsdk:"output_key"`
}

func (r *tagPipelineRulesetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "tag_pipeline_ruleset"
}

func (r *tagPipelineRulesetResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provides a Datadog Tag Pipeline Ruleset resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the ruleset.",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the ruleset.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the ruleset is enabled.",
			},
			"position": schema.Int64Attribute{
				Computed:    true,
				Description: "The position of the ruleset in the pipeline.",
			},
			"version": schema.Int64Attribute{
				Computed:    true,
				Description: "The version of the ruleset.",
			},
		},
		Blocks: map[string]schema.Block{
			"rules": schema.ListNestedBlock{
				Description: "The rules in the ruleset.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"enabled": schema.BoolAttribute{
							Required:    true,
							Description: "Whether the rule is enabled.",
						},
						"name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the rule.",
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
									Optional:    true,
									Description: "The destination key for the mapping.",
								},
								"if_not_exists": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Whether to apply the mapping only if the destination key doesn't exist.",
								},
								"source_keys": schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Description: "The source keys for the mapping.",
								},
							},
						},
						"query": schema.SingleNestedBlock{
							Description: "The query configuration for the rule.",
							Attributes: map[string]schema.Attribute{
								"case_insensitivity": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Whether the query matching is case insensitive.",
								},
								"if_not_exists": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Whether to apply the query only if the key doesn't exist.",
								},
								"query": schema.StringAttribute{
									Optional:    true,
									Description: "The query string.",
								},
							},
							Blocks: map[string]schema.Block{
								"addition": schema.SingleNestedBlock{
									Description: "The addition configuration for the query.",
									Attributes: map[string]schema.Attribute{
										"key": schema.StringAttribute{
											Optional:    true,
											Description: "The key to add.",
										},
										"value": schema.StringAttribute{
											Optional:    true,
											Description: "The value to add.",
										},
									},
								},
							},
						},
						"reference_table": schema.SingleNestedBlock{
							Description: "The reference table configuration for the rule.",
							Attributes: map[string]schema.Attribute{
								"case_insensitivity": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Whether the reference table lookup is case insensitive.",
								},
								"if_not_exists": schema.BoolAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Whether to apply the reference table only if the key doesn't exist.",
								},
								"source_keys": schema.ListAttribute{
									ElementType: types.StringType,
									Optional:    true,
									Description: "The source keys for the reference table lookup.",
								},
								"table_name": schema.StringAttribute{
									Optional:    true,
									Description: "The name of the reference table.",
								},
							},
							Blocks: map[string]schema.Block{
								"field_pairs": schema.ListNestedBlock{
									Description: "The field pairs for the reference table.",
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"input_column": schema.StringAttribute{
												Optional:    true,
												Description: "The input column name.",
											},
											"output_key": schema.StringAttribute{
												Optional:    true,
												Description: "The output key name.",
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

func (r *tagPipelineRulesetResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	providerData := req.ProviderData.(*FrameworkProvider)
	r.Api = providerData.DatadogApiInstances.GetCloudCostManagementApiV2()
	r.Auth = providerData.Auth
}

func (r *tagPipelineRulesetResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	if request.State.Raw.IsNull() {
		return
	}
	if request.Plan.Raw.IsNull() {
		return
	}

	var config, plan tagPipelineRulesetModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if response.Diagnostics.HasError() {
		return
	}

	for index := range plan.Rules {
		if config.Rules[index].Mapping == nil {
			plan.Rules[index].Mapping = nil
		}
		if config.Rules[index].Query == nil {
			plan.Rules[index].Query = nil
		}
		if config.Rules[index].ReferenceTable == nil {
			plan.Rules[index].ReferenceTable = nil
		}
	}
	response.Diagnostics.Append(response.Plan.Set(ctx, &plan)...)
}

// --- CRUD ---

func (r *tagPipelineRulesetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagPipelineRulesetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate rules configuration
	validateRules(plan.Rules, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildCreateRulesetRequestFromModel(plan)
	apiResp, response, err := r.Api.CreateRuleset(r.Auth, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating ruleset", utils.TranslateClientError(err, response, "").Error())
		return
	}

	// Create a fresh model from the API response to ensure clean state
	var newState tagPipelineRulesetModel
	setModelFromRulesetResp(&newState, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *tagPipelineRulesetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagPipelineRulesetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiResp, response, err := r.Api.GetRuleset(r.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading ruleset", utils.TranslateClientError(err, response, "").Error())
		return
	}
	if apiResp.Data == nil {
		tflog.Debug(ctx, "GetRuleset response with empty data")
	}
	if apiResp.Data != nil && apiResp.Data.Attributes == nil {
		tflog.Debug(ctx, "GetRuleset response with empty Attributes", map[string]interface{}{
			"has_unparsed_object": apiResp.Data.UnparsedObject != nil,
		})
	}
	if apiResp.Data != nil && apiResp.Data.Attributes != nil {
		attr := apiResp.Data.Attributes
		tflog.Debug(ctx, "GetRuleset response",
			map[string]interface{}{"name": attr.Name, "enabled": attr.Enabled, "position": attr.Position, "version": attr.Version, "rules_count": len(attr.Rules)})
	}

	setModelFromRulesetResp(&state, apiResp)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *tagPipelineRulesetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagPipelineRulesetModel
	var config tagPipelineRulesetModel
	var state tagPipelineRulesetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use the ID and version from the current state, not the plan (needed for the update API)
	plan.ID = state.ID
	plan.Version = state.Version

	// Validate rules configuration
	validateRules(plan.Rules, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	rulesetId := plan.ID.ValueString()
	if rulesetId == "" {
		resp.Diagnostics.AddError("Error updating ruleset", "Ruleset ID is empty")
		return
	}

	apiReq := buildUpdateRulesetRequestFromModel(plan)
	apiResp, response, err := r.Api.UpdateRuleset(r.Auth, rulesetId, apiReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ruleset", fmt.Sprintf("RulesetID: %s, Error: %s", rulesetId, utils.TranslateClientError(err, response, "").Error()))
		return
	}

	var newState tagPipelineRulesetModel
	setModelFromRulesetResp(&newState, apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *tagPipelineRulesetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagPipelineRulesetModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.Api.DeleteRuleset(r.Auth, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error deleting ruleset", err.Error())
		return
	}
}

func (r *tagPipelineRulesetResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, frameworkPath.Root("id"), req, resp)
}

// --- Helper functions to map between model and API types ---

// convertMetadataToMap converts types.Map metadata to map[string]string
func convertMetadataToMap(metadata types.Map) map[string]string {
	if metadata.IsNull() || metadata.IsUnknown() {
		return nil
	}
	result := make(map[string]string)
	for k, v := range metadata.Elements() {
		if strVal, ok := v.(types.String); ok {
			result[k] = strVal.ValueString()
		}
	}
	return result
}

// convertSourceKeys converts []types.String to []string
func convertSourceKeys(sourceKeys []types.String) []string {
	result := make([]string, len(sourceKeys))
	for i, sk := range sourceKeys {
		result[i] = sk.ValueString()
	}
	return result
}

// validateRules validates the rules configuration and adds diagnostics errors if validation fails.
func validateRules(rules []ruleItem, diagnostics *diag.Diagnostics) {
	for i, rule := range rules {
		// Count how many rule types are defined
		ruleTypeCount := 0
		if rule.Mapping != nil {
			ruleTypeCount++
		}
		if rule.Query != nil {
			ruleTypeCount++
		}
		if rule.ReferenceTable != nil {
			ruleTypeCount++
		}

		// Exactly one rule type must be defined
		if ruleTypeCount == 0 {
			diagnostics.AddError(
				"Missing rule configuration",
				fmt.Sprintf("rules[%d] must define exactly one of: mapping, query, or reference_table", i),
			)
			continue
		}
		if ruleTypeCount > 1 {
			diagnostics.AddError(
				"Multiple rule configurations",
				fmt.Sprintf("rules[%d] can only define one of: mapping, query, or reference_table", i),
			)
			continue
		}

		// Validate mapping block
		if rule.Mapping != nil {
			if rule.Mapping.DestinationKey.IsNull() || rule.Mapping.DestinationKey.ValueString() == "" {
				diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("rules[%d].mapping.destination_key is required when mapping block is used", i),
				)
			}
			if len(rule.Mapping.SourceKeys) == 0 {
				diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("rules[%d].mapping.source_keys is required when mapping block is used", i),
				)
			}
		}

		// Validate query block
		if rule.Query != nil {
			if rule.Query.Query.IsNull() || rule.Query.Query.ValueString() == "" {
				diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("rules[%d].query.query is required when query block is used", i),
				)
			}
			// Addition block is required for query rules
			if rule.Query.Addition == nil {
				diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("rules[%d].query.addition block is required when query block is used", i),
				)
			} else {
				if rule.Query.Addition.Key.IsNull() || rule.Query.Addition.Key.ValueString() == "" {
					diagnostics.AddError(
						"Missing required attribute",
						fmt.Sprintf("rules[%d].query.addition.key is required when addition block is used", i),
					)
				}
				if rule.Query.Addition.Value.IsNull() || rule.Query.Addition.Value.ValueString() == "" {
					diagnostics.AddError(
						"Missing required attribute",
						fmt.Sprintf("rules[%d].query.addition.value is required when addition block is used", i),
					)
				}
			}
		}

		// Validate reference_table block
		if rule.ReferenceTable != nil {
			if rule.ReferenceTable.TableName.IsNull() || rule.ReferenceTable.TableName.ValueString() == "" {
				diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("rules[%d].reference_table.table_name is required when reference_table block is used", i),
				)
			}
			if len(rule.ReferenceTable.SourceKeys) == 0 {
				diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("rules[%d].reference_table.source_keys is required when reference_table block is used", i),
				)
			}
		}
	}
}

func buildCreateRulesetRequestFromModel(plan tagPipelineRulesetModel) datadogV2.CreateRulesetRequest {
	// Convert rules
	var rules []datadogV2.CreateRulesetRequestDataAttributesRulesItems
	for _, r := range plan.Rules {
		rule := datadogV2.CreateRulesetRequestDataAttributesRulesItems{
			Enabled: r.Enabled.ValueBool(),
			Name:    r.Name.ValueString(),
		}

		// Set metadata if provided
		if metadata := convertMetadataToMap(r.Metadata); metadata != nil {
			rule.Metadata = metadata
		}

		// Set mapping if provided
		if r.Mapping != nil {
			mapping := datadogV2.CreateRulesetRequestDataAttributesRulesItemsMapping{
				DestinationKey: r.Mapping.DestinationKey.ValueString(),
				IfNotExists:    !r.Mapping.IfNotExists.IsNull() && r.Mapping.IfNotExists.ValueBool(),
				SourceKeys:     convertSourceKeys(r.Mapping.SourceKeys),
			}
			rule.Mapping = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsMapping(&mapping)
		} else {
			rule.Mapping = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsMapping(nil)
		}

		// Set query if provided
		if r.Query != nil {
			query := datadogV2.CreateRulesetRequestDataAttributesRulesItemsQuery{
				CaseInsensitivity: func() *bool {
					if !r.Query.CaseInsensitivity.IsNull() {
						val := r.Query.CaseInsensitivity.ValueBool()
						return &val
					}
					return nil
				}(),
				IfNotExists: !r.Query.IfNotExists.IsNull() && r.Query.IfNotExists.ValueBool(),
				Query:       r.Query.Query.ValueString(),
			}
			// Addition is required for query rules
			if r.Query.Addition != nil {
				addition := datadogV2.CreateRulesetRequestDataAttributesRulesItemsQueryAddition{
					Key:   r.Query.Addition.Key.ValueString(),
					Value: r.Query.Addition.Value.ValueString(),
				}
				query.Addition = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsQueryAddition(&addition)
			}
			rule.Query = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsQuery(&query)
		} else {
			rule.Query = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsQuery(nil)
		}

		// Set reference table if provided
		if r.ReferenceTable != nil {
			var fieldPairs []datadogV2.CreateRulesetRequestDataAttributesRulesItemsReferenceTableFieldPairsItems
			for _, fp := range r.ReferenceTable.FieldPairs {
				fieldPairs = append(fieldPairs, datadogV2.CreateRulesetRequestDataAttributesRulesItemsReferenceTableFieldPairsItems{
					InputColumn: fp.InputColumn.ValueString(),
					OutputKey:   fp.OutputKey.ValueString(),
				})
			}
			refTable := datadogV2.CreateRulesetRequestDataAttributesRulesItemsReferenceTable{
				CaseInsensitivity: func() *bool {
					if !r.ReferenceTable.CaseInsensitivity.IsNull() {
						val := r.ReferenceTable.CaseInsensitivity.ValueBool()
						return &val
					}
					return nil
				}(),
				FieldPairs: fieldPairs,
				IfNotExists: func() *bool {
					if !r.ReferenceTable.IfNotExists.IsNull() {
						val := r.ReferenceTable.IfNotExists.ValueBool()
						return &val
					}
					return nil
				}(),
				SourceKeys: convertSourceKeys(r.ReferenceTable.SourceKeys),
				TableName:  r.ReferenceTable.TableName.ValueString(),
			}
			rule.ReferenceTable = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsReferenceTable(&refTable)
		} else {
			rule.ReferenceTable = *datadogV2.NewNullableCreateRulesetRequestDataAttributesRulesItemsReferenceTable(nil)
		}

		rules = append(rules, rule)
	}

	// Build attributes
	attributes := datadogV2.CreateRulesetRequestDataAttributes{}

	// Always set rules, but use nil for empty rulesets to properly represent absence
	// This allows rulesets with no rules to be created without causing API errors
	if len(rules) > 0 {
		attributes.Rules = rules
	} else {
		// For empty rulesets, set an empty array explicitly using make
		// The API requires the rules field to be present, even if empty
		attributes.Rules = make([]datadogV2.CreateRulesetRequestDataAttributesRulesItems, 0)
	}

	if !plan.Enabled.IsNull() {
		attributes.Enabled = plan.Enabled.ValueBoolPointer()
	}

	// Build data - set Id to the user-provided name, API will generate UUID and return name in attributes
	nameValue := plan.Name.ValueString()
	data := datadogV2.CreateRulesetRequestData{
		Id:         &nameValue,
		Attributes: &attributes,
		Type:       datadogV2.CREATERULESETREQUESTDATATYPE_CREATE_RULESET,
	}

	// Build and return the top-level object
	return datadogV2.CreateRulesetRequest{
		Data: &data,
	}
}

func buildUpdateRulesetRequestFromModel(plan tagPipelineRulesetModel) datadogV2.UpdateRulesetRequest {
	// Convert rules
	var rules []datadogV2.UpdateRulesetRequestDataAttributesRulesItems
	for _, r := range plan.Rules {
		rule := datadogV2.UpdateRulesetRequestDataAttributesRulesItems{
			Enabled: r.Enabled.ValueBool(),
			Name:    r.Name.ValueString(),
		}

		// Set metadata if provided
		if metadata := convertMetadataToMap(r.Metadata); metadata != nil {
			rule.Metadata = metadata
		}

		// Set mapping if provided
		if r.Mapping != nil {
			mapping := datadogV2.UpdateRulesetRequestDataAttributesRulesItemsMapping{
				DestinationKey: r.Mapping.DestinationKey.ValueString(),
				IfNotExists:    !r.Mapping.IfNotExists.IsNull() && r.Mapping.IfNotExists.ValueBool(),
				SourceKeys:     convertSourceKeys(r.Mapping.SourceKeys),
			}
			rule.Mapping = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsMapping(&mapping)
		} else {
			rule.Mapping = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsMapping(nil)
		}

		// Set query if provided
		if r.Query != nil {
			query := datadogV2.UpdateRulesetRequestDataAttributesRulesItemsQuery{
				CaseInsensitivity: func() *bool {
					if !r.Query.CaseInsensitivity.IsNull() {
						val := r.Query.CaseInsensitivity.ValueBool()
						return &val
					}
					return nil
				}(),
				IfNotExists: !r.Query.IfNotExists.IsNull() && r.Query.IfNotExists.ValueBool(),
				Query:       r.Query.Query.ValueString(),
			}
			// Addition is required for query rules
			if r.Query.Addition != nil {
				addition := datadogV2.UpdateRulesetRequestDataAttributesRulesItemsQueryAddition{
					Key:   r.Query.Addition.Key.ValueString(),
					Value: r.Query.Addition.Value.ValueString(),
				}
				query.Addition = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsQueryAddition(&addition)
			}
			rule.Query = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsQuery(&query)
		} else {
			rule.Query = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsQuery(nil)
		}

		// Set reference table if provided
		if r.ReferenceTable != nil {
			var fieldPairs []datadogV2.UpdateRulesetRequestDataAttributesRulesItemsReferenceTableFieldPairsItems
			for _, fp := range r.ReferenceTable.FieldPairs {
				fieldPairs = append(fieldPairs, datadogV2.UpdateRulesetRequestDataAttributesRulesItemsReferenceTableFieldPairsItems{
					InputColumn: fp.InputColumn.ValueString(),
					OutputKey:   fp.OutputKey.ValueString(),
				})
			}
			refTable := datadogV2.UpdateRulesetRequestDataAttributesRulesItemsReferenceTable{
				CaseInsensitivity: func() *bool {
					if !r.ReferenceTable.CaseInsensitivity.IsNull() {
						val := r.ReferenceTable.CaseInsensitivity.ValueBool()
						return &val
					}
					return nil
				}(),
				FieldPairs: fieldPairs,
				IfNotExists: func() *bool {
					if !r.ReferenceTable.IfNotExists.IsNull() {
						val := r.ReferenceTable.IfNotExists.ValueBool()
						return &val
					}
					return nil
				}(),
				SourceKeys: convertSourceKeys(r.ReferenceTable.SourceKeys),
				TableName:  r.ReferenceTable.TableName.ValueString(),
			}
			rule.ReferenceTable = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsReferenceTable(&refTable)
		} else {
			rule.ReferenceTable = *datadogV2.NewNullableUpdateRulesetRequestDataAttributesRulesItemsReferenceTable(nil)
		}

		rules = append(rules, rule)
	}

	// Build attributes
	attributes := datadogV2.UpdateRulesetRequestDataAttributes{
		Enabled:     plan.Enabled.ValueBool(),
		LastVersion: plan.Version.ValueInt64Pointer(),
	}

	// Always include the rules field - use empty array for rulesets with no rules
	// The API accepts empty arrays: []
	if len(rules) > 0 {
		attributes.Rules = rules
	} else {
		// For empty rulesets, set an empty array explicitly using make
		// The API requires the rules field to be present, even if empty
		attributes.Rules = make([]datadogV2.UpdateRulesetRequestDataAttributesRulesItems, 0)
	}

	// Add name via AdditionalProperties since it's not in the explicit struct fields
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		attributes.AdditionalProperties = map[string]interface{}{
			"name": plan.Name.ValueString(),
		}
	}

	// Build data
	data := datadogV2.UpdateRulesetRequestData{
		Attributes: &attributes,
		Type:       datadogV2.UPDATERULESETREQUESTDATATYPE_UPDATE_RULESET,
	}

	// Build and return the top-level object
	return datadogV2.UpdateRulesetRequest{
		Data: &data,
	}
}

func setModelFromRulesetResp(model *tagPipelineRulesetModel, apiResp datadogV2.RulesetResp) {
	if apiResp.Data == nil {
		return
	}
	data := apiResp.Data

	// Handle case where Attributes is nil but UnparsedObject has the data
	// This happens when the API returns fields the generated client doesn't know about (like rules:null)
	// In this case, the entire Data object fails to unmarshal and everything goes into UnparsedObject
	if data.Attributes == nil && data.UnparsedObject != nil {
		// Extract ID from top-level UnparsedObject
		if id, ok := data.UnparsedObject["id"].(string); ok && id != "" {
			model.ID = types.StringValue(id)
		} else {
			model.ID = types.StringValue("")
		}

		// Try to extract attributes from UnparsedObject
		if attributesRaw, ok := data.UnparsedObject["attributes"].(map[string]interface{}); ok {
			// Extract name
			if name, ok := attributesRaw["name"].(string); ok && name != "" {
				model.Name = types.StringValue(name)
			} else {
				model.Name = types.StringValue("")
			}

			// Extract enabled
			if enabled, ok := attributesRaw["enabled"].(bool); ok {
				model.Enabled = types.BoolValue(enabled)
			} else {
				model.Enabled = types.BoolValue(true) // default
			}

			// Extract position
			if position, ok := attributesRaw["position"].(float64); ok {
				model.Position = types.Int64Value(int64(position))
			} else {
				model.Position = types.Int64Value(0)
			}

			// Extract version
			if version, ok := attributesRaw["version"].(float64); ok {
				model.Version = types.Int64Value(int64(version))
			} else {
				model.Version = types.Int64Value(1)
			}

			// Handle rules - could be null or an empty array
			if rulesRaw, ok := attributesRaw["rules"]; ok && rulesRaw != nil {
				// Rules will be handled below
				model.Rules = []ruleItem{}
			} else {
				model.Rules = []ruleItem{}
			}
		}
		return
	}

	if data.Attributes == nil {
		return
	}
	attr := data.Attributes

	// Set ID from the proper field when Attributes is available
	if data.Id != nil && *data.Id != "" {
		model.ID = types.StringValue(*data.Id)
	} else {
		model.ID = types.StringValue("")
	}

	if attr.Name != "" {
		model.Name = types.StringValue(attr.Name)
	} else {
		model.Name = types.StringValue("")
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
			// Initialize all rule type fields to nil to ensure clean state
			Mapping:        nil,
			Query:          nil,
			ReferenceTable: nil,
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

		// Set mapping if present (and ensure others are nil)
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

		// Set query if present (and ensure others are nil)
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

		// Set reference table if present (and ensure others are nil)
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
