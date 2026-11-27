package fwprovider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	frameworkPath "github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
	IfNotExists    types.Bool     `tfsdk:"if_not_exists"` // Deprecated: use IfTagExists
	IfTagExists    types.String   `tfsdk:"if_tag_exists"`
	SourceKeys     []types.String `tfsdk:"source_keys"`
}

type ruleQuery struct {
	Addition          *queryAddition `tfsdk:"addition"`
	CaseInsensitivity types.Bool     `tfsdk:"case_insensitivity"`
	IfNotExists       types.Bool     `tfsdk:"if_not_exists"` // Deprecated: use IfTagExists
	IfTagExists       types.String   `tfsdk:"if_tag_exists"`
	Query             types.String   `tfsdk:"query"`
}

type queryAddition struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}

type referenceTable struct {
	CaseInsensitivity types.Bool                `tfsdk:"case_insensitivity"`
	FieldPairs        []referenceTableFieldPair `tfsdk:"field_pairs"`
	IfNotExists       types.Bool                `tfsdk:"if_not_exists"` // Deprecated: use IfTagExists
	IfTagExists       types.String              `tfsdk:"if_tag_exists"`
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
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the ruleset.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Whether the ruleset is enabled.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			// position and version have no plan modifiers on purpose: version changes on
			// every update, and position can shift when another ruleset is deleted.
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
							PlanModifiers: []planmodifier.Map{
								preserveRuleMetadataModifier{},
							},
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
									Optional:           true,
									Computed:           true,
									Description:        "Whether to apply the mapping only if the destination key doesn't exist.",
									DeprecationMessage: "Use `if_tag_exists` instead. This field will be removed in a future release.",
									Validators: []validator.Bool{
										boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("if_tag_exists")),
									},
									PlanModifiers: []planmodifier.Bool{
										ifNotExistsFromIfTagExistsModifier{},
									},
								},
								"if_tag_exists": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Behavior when the tag already exists. Valid values: `append` (append to the existing tag value), `replace` (replace existing tag value), `do_not_apply` (never apply if tag already exists).",
									Validators: []validator.String{
										stringvalidator.OneOf("append", "replace", "do_not_apply"),
										stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("if_not_exists")),
									},
									PlanModifiers: []planmodifier.String{
										ifTagExistsFromIfNotExistsModifier{fallbackForFalse: ifTagExistsReplace},
									},
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
									PlanModifiers: []planmodifier.Bool{
										caseInsensitivityModifier{},
									},
								},
								"if_not_exists": schema.BoolAttribute{
									Optional:           true,
									Computed:           true,
									Description:        "Whether to apply the query only if the key doesn't exist.",
									DeprecationMessage: "Use `if_tag_exists` instead. This field will be removed in a future release.",
									Validators: []validator.Bool{
										boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("if_tag_exists")),
									},
									PlanModifiers: []planmodifier.Bool{
										ifNotExistsFromIfTagExistsModifier{},
									},
								},
								"if_tag_exists": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Behavior when the tag already exists. Valid values: `append` (append to the existing tag value), `replace` (replace existing tag value), `do_not_apply` (never apply if tag already exists).",
									Validators: []validator.String{
										stringvalidator.OneOf("append", "replace", "do_not_apply"),
										stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("if_not_exists")),
									},
									PlanModifiers: []planmodifier.String{
										ifTagExistsFromIfNotExistsModifier{fallbackForFalse: ifTagExistsAppend},
									},
								},
								"query": schema.StringAttribute{
									Optional:    true,
									Description: "The query string. Datadog normalizes queries to a canonical form (operator casing, spacing, redundant parentheses, quoting, and similar formatting differences). The provider validates the configured value against this canonical form during planning. If the value doesn't match, the plan fails with an error showing the canonical query to use in the configuration.",
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
									PlanModifiers: []planmodifier.Bool{
										caseInsensitivityModifier{},
									},
								},
								"if_not_exists": schema.BoolAttribute{
									Optional:           true,
									Computed:           true,
									Description:        "Whether to apply the reference table only if the key doesn't exist.",
									DeprecationMessage: "Use `if_tag_exists` instead. This field will be removed in a future release.",
									Validators: []validator.Bool{
										boolvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("if_tag_exists")),
									},
									PlanModifiers: []planmodifier.Bool{
										ifNotExistsFromIfTagExistsModifier{},
									},
								},
								"if_tag_exists": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Description: "Behavior when the tag already exists. Valid values: `append` (append to the existing tag value), `replace` (replace existing tag value), `do_not_apply` (never apply if tag already exists).",
									Validators: []validator.String{
										stringvalidator.OneOf("append", "replace", "do_not_apply"),
										stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("if_not_exists")),
									},
									PlanModifiers: []planmodifier.String{
										ifTagExistsFromIfNotExistsModifier{fallbackForFalse: ifTagExistsAppend},
									},
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
	if request.Plan.Raw.IsNull() {
		return
	}

	var config tagPipelineRulesetModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.validateCanonicalQueries(config.Rules, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if request.State.Raw.IsNull() {
		return
	}

	var plan tagPipelineRulesetModel
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

func (r *tagPipelineRulesetResource) validateCanonicalQueries(rules []ruleItem, diags *diag.Diagnostics) {
	for i, rule := range rules {
		if rule.Query == nil {
			continue
		}
		configured := rule.Query.Query
		if configured.IsNull() || configured.IsUnknown() || configured.ValueString() == "" {
			continue
		}

		raw := configured.ValueString()
		canonical, httpResp, err := r.canonicalizeQuery(raw)
		queryPath := path.Root("rules").AtListIndex(i).AtName("query").AtName("query")
		if err != nil {
			diags.AddAttributeError(
				queryPath,
				"Unable to validate tag pipeline query",
				fmt.Sprintf("Query %q could not be validated: %s", raw, utils.TranslateClientError(err, httpResp, "").Error()),
			)
			continue
		}

		if raw != canonical {
			diags.AddAttributeError(
				queryPath,
				"Tag pipeline query is not in canonical form",
				fmt.Sprintf(
					"Datadog normalizes tag pipeline queries to a canonical form. Update the configuration to the canonical value to avoid a persistent plan diff.\n\n  configured: %s\n  canonical:  %s",
					raw, canonical,
				),
			)
		}
	}
}

func (r *tagPipelineRulesetResource) canonicalizeQuery(raw string) (string, *http.Response, error) {
	body := datadogV2.RulesValidateQueryRequest{
		Data: &datadogV2.RulesValidateQueryRequestData{
			Type: datadogV2.RULESVALIDATEQUERYREQUESTDATATYPE_VALIDATE_QUERY,
			Attributes: &datadogV2.RulesValidateQueryRequestDataAttributes{
				Query: raw,
			},
		},
	}
	resp, httpResp, err := r.Api.ValidateQuery(r.Auth, body)
	if err != nil {
		return "", httpResp, err
	}
	data := resp.GetData()
	attributes := data.GetAttributes()
	return attributes.GetCanonical(), httpResp, nil
}

// --- CRUD ---

func (r *tagPipelineRulesetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagPipelineRulesetModel
	var config tagPipelineRulesetModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate rules configuration
	validateRules(plan.Rules, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	clearUnconfiguredTagExistsFields(&plan, config)
	apiReq := buildCreateRulesetRequestFromModel(plan)
	apiResp, response, err := r.Api.CreateTagPipelinesRuleset(r.Auth, apiReq)
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

	apiResp, response, err := r.Api.GetTagPipelinesRuleset(r.Auth, state.ID.ValueString())
	if err != nil {
		if response != nil && response.StatusCode == 404 {
			// Resource was deleted outside of Terraform - remove from state
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading ruleset", utils.TranslateClientError(err, response, "").Error())
		return
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

	clearUnconfiguredTagExistsFields(&plan, config)
	apiReq := buildUpdateRulesetRequestFromModel(plan)
	apiResp, response, err := r.Api.UpdateTagPipelinesRuleset(r.Auth, rulesetId, apiReq)
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

	httpResp, err := r.Api.DeleteTagPipelinesRuleset(r.Auth, state.ID.ValueString())
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Resource already deleted - this is fine
			return
		}
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
			mapping := datadogV2.DataAttributesRulesItemsMapping{
				DestinationKey: r.Mapping.DestinationKey.ValueString(),
				SourceKeys:     convertSourceKeys(r.Mapping.SourceKeys),
			}
			// Send whichever field the user specified - API handles conversion
			if !r.Mapping.IfTagExists.IsNull() && !r.Mapping.IfTagExists.IsUnknown() {
				mapping.IfTagExists = datadogV2.DataAttributesRulesItemsIfTagExists(r.Mapping.IfTagExists.ValueString()).Ptr()
			} else if !r.Mapping.IfNotExists.IsNull() && !r.Mapping.IfNotExists.IsUnknown() {
				val := r.Mapping.IfNotExists.ValueBool()
				mapping.IfNotExists = &val
			}
			rule.Mapping = *datadogV2.NewNullableDataAttributesRulesItemsMapping(&mapping)
		} else {
			rule.Mapping = *datadogV2.NewNullableDataAttributesRulesItemsMapping(nil)
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
				Query: r.Query.Query.ValueString(),
			}
			// Send whichever field the user specified - API handles conversion
			if !r.Query.IfTagExists.IsNull() && !r.Query.IfTagExists.IsUnknown() {
				query.IfTagExists = datadogV2.DataAttributesRulesItemsIfTagExists(r.Query.IfTagExists.ValueString()).Ptr()
			} else if !r.Query.IfNotExists.IsNull() && !r.Query.IfNotExists.IsUnknown() {
				val := r.Query.IfNotExists.ValueBool()
				query.IfNotExists = &val
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
				SourceKeys: convertSourceKeys(r.ReferenceTable.SourceKeys),
				TableName:  r.ReferenceTable.TableName.ValueString(),
			}
			// Send whichever field the user specified - API handles conversion
			if !r.ReferenceTable.IfTagExists.IsNull() && !r.ReferenceTable.IfTagExists.IsUnknown() {
				refTable.IfTagExists = datadogV2.DataAttributesRulesItemsIfTagExists(r.ReferenceTable.IfTagExists.ValueString()).Ptr()
			} else if !r.ReferenceTable.IfNotExists.IsNull() && !r.ReferenceTable.IfNotExists.IsUnknown() {
				val := r.ReferenceTable.IfNotExists.ValueBool()
				refTable.IfNotExists = &val
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
			mapping := datadogV2.DataAttributesRulesItemsMapping{
				DestinationKey: r.Mapping.DestinationKey.ValueString(),
				SourceKeys:     convertSourceKeys(r.Mapping.SourceKeys),
			}
			// Send whichever field the user specified - API handles conversion
			if !r.Mapping.IfTagExists.IsNull() && !r.Mapping.IfTagExists.IsUnknown() {
				mapping.IfTagExists = datadogV2.DataAttributesRulesItemsIfTagExists(r.Mapping.IfTagExists.ValueString()).Ptr()
			} else if !r.Mapping.IfNotExists.IsNull() && !r.Mapping.IfNotExists.IsUnknown() {
				val := r.Mapping.IfNotExists.ValueBool()
				mapping.IfNotExists = &val
			}
			rule.Mapping = *datadogV2.NewNullableDataAttributesRulesItemsMapping(&mapping)
		} else {
			rule.Mapping = *datadogV2.NewNullableDataAttributesRulesItemsMapping(nil)
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
				Query: r.Query.Query.ValueString(),
			}
			// Send whichever field the user specified - API handles conversion
			if !r.Query.IfTagExists.IsNull() && !r.Query.IfTagExists.IsUnknown() {
				query.IfTagExists = datadogV2.DataAttributesRulesItemsIfTagExists(r.Query.IfTagExists.ValueString()).Ptr()
			} else if !r.Query.IfNotExists.IsNull() && !r.Query.IfNotExists.IsUnknown() {
				val := r.Query.IfNotExists.ValueBool()
				query.IfNotExists = &val
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
				SourceKeys: convertSourceKeys(r.ReferenceTable.SourceKeys),
				TableName:  r.ReferenceTable.TableName.ValueString(),
			}
			// Send whichever field the user specified - API handles conversion
			if !r.ReferenceTable.IfTagExists.IsNull() && !r.ReferenceTable.IfTagExists.IsUnknown() {
				refTable.IfTagExists = datadogV2.DataAttributesRulesItemsIfTagExists(r.ReferenceTable.IfTagExists.ValueString()).Ptr()
			} else if !r.ReferenceTable.IfNotExists.IsNull() && !r.ReferenceTable.IfNotExists.IsUnknown() {
				val := r.ReferenceTable.IfNotExists.ValueBool()
				refTable.IfNotExists = &val
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

			model.Rules = []ruleItem{}
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
					IfNotExists: func() types.Bool {
						if mappingVal.IfNotExists != nil {
							return types.BoolValue(*mappingVal.IfNotExists)
						}
						return types.BoolNull()
					}(),
					IfTagExists: func() types.String {
						if mappingVal.IfTagExists != nil {
							return types.StringValue(string(*mappingVal.IfTagExists))
						}
						return types.StringNull()
					}(),
					SourceKeys: sourceKeys,
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
					IfNotExists: func() types.Bool {
						if queryVal.IfNotExists != nil {
							return types.BoolValue(*queryVal.IfNotExists)
						}
						return types.BoolNull()
					}(),
					IfTagExists: func() types.String {
						if queryVal.IfTagExists != nil {
							return types.StringValue(string(*queryVal.IfTagExists))
						}
						return types.StringNull()
					}(),
					Query: types.StringValue(queryVal.Query),
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
					IfTagExists: func() types.String {
						if refTableVal.IfTagExists != nil {
							return types.StringValue(string(*refTableVal.IfTagExists))
						}
						return types.StringNull()
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

// --- Plan modifiers ---
//
// Without these, every Computed attribute goes unknown whenever the ruleset
// changes, so plans show unchanged rules as "(known after apply)".

const (
	ifTagExistsAppend     = "append"
	ifTagExistsReplace    = "replace"
	ifTagExistsDoNotApply = "do_not_apply"
)

// ifTagExistsFromIfNotExists mirrors how the API derives if_tag_exists from the
// deprecated if_not_exists field. fallbackForFalse depends on the rule type:
// "replace" for mapping rules, "append" for query and reference table rules.
func ifTagExistsFromIfNotExists(ifNotExists bool, fallbackForFalse string) string {
	if ifNotExists {
		return ifTagExistsDoNotApply
	}
	return fallbackForFalse
}

// ifNotExistsFromIfTagExists mirrors how the API derives the deprecated
// if_not_exists field from if_tag_exists. ok is false for unrecognized values.
func ifNotExistsFromIfTagExists(ifTagExists string) (value bool, ok bool) {
	switch ifTagExists {
	case ifTagExistsDoNotApply:
		return true, true
	case ifTagExistsAppend, ifTagExistsReplace:
		return false, true
	default:
		return false, false
	}
}

// ifTagExistsFromIfNotExistsModifier plans if_tag_exists from if_not_exists
// when only if_not_exists is configured. It only acts when state already holds
// a value, which shows the API returns the field for this rule.
type ifTagExistsFromIfNotExistsModifier struct {
	fallbackForFalse string
}

func (m ifTagExistsFromIfNotExistsModifier) Description(_ context.Context) string {
	return "Derives if_tag_exists from if_not_exists when only if_not_exists is configured."
}

func (m ifTagExistsFromIfNotExistsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ifTagExistsFromIfNotExistsModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	if !resp.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	var ifNotExists types.Bool
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("if_not_exists"), &ifNotExists)
	if diags.HasError() || ifNotExists.IsNull() || ifNotExists.IsUnknown() {
		return
	}

	resp.PlanValue = types.StringValue(ifTagExistsFromIfNotExists(ifNotExists.ValueBool(), m.fallbackForFalse))
}

// ifNotExistsFromIfTagExistsModifier plans if_not_exists from if_tag_exists
// when only if_tag_exists is configured. It only acts when state already holds
// a value, which shows the API returns the field for this rule.
type ifNotExistsFromIfTagExistsModifier struct{}

func (m ifNotExistsFromIfTagExistsModifier) Description(_ context.Context) string {
	return "Derives if_not_exists from if_tag_exists when only if_tag_exists is configured."
}

func (m ifNotExistsFromIfTagExistsModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m ifNotExistsFromIfTagExistsModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !resp.PlanValue.IsUnknown() || req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	var ifTagExists types.String
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath().AtName("if_tag_exists"), &ifTagExists)
	if diags.HasError() || ifTagExists.IsNull() || ifTagExists.IsUnknown() {
		return
	}

	if value, ok := ifNotExistsFromIfTagExists(ifTagExists.ValueString()); ok {
		resp.PlanValue = types.BoolValue(value)
	}
}

// caseInsensitivityModifier plans false when case_insensitivity is not
// configured, which is what the API stores and what the request sends.
type caseInsensitivityModifier struct{}

func (m caseInsensitivityModifier) Description(_ context.Context) string {
	return "Defaults case_insensitivity to false when it is not configured."
}

func (m caseInsensitivityModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m caseInsensitivityModifier) PlanModifyBool(ctx context.Context, req planmodifier.BoolRequest, resp *planmodifier.BoolResponse) {
	if !resp.PlanValue.IsUnknown() {
		return
	}

	// Leave blocks that are absent from the configuration alone.
	var block types.Object
	diags := req.Config.GetAttribute(ctx, req.Path.ParentPath(), &block)
	if diags.HasError() || block.IsNull() || block.IsUnknown() {
		return
	}

	resp.PlanValue = types.BoolValue(false)
}

// preserveRuleMetadataModifier keeps a rule's metadata when the rule at the same
// position keeps its name. Metadata can't be set from Terraform, so the update
// request sends this value back to avoid wiping metadata set in the UI.
type preserveRuleMetadataModifier struct{}

func (m preserveRuleMetadataModifier) Description(_ context.Context) string {
	return "Keeps rule metadata from state when the rule name is unchanged."
}

func (m preserveRuleMetadataModifier) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (m preserveRuleMetadataModifier) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
	if !resp.PlanValue.IsUnknown() {
		return
	}

	resp.PlanValue = types.MapNull(types.StringType)
	if req.StateValue.IsNull() || req.StateValue.IsUnknown() {
		return
	}

	namePath := req.Path.ParentPath().AtName("name")
	var configName, stateName types.String
	if diags := req.Config.GetAttribute(ctx, namePath, &configName); diags.HasError() {
		return
	}
	if diags := req.State.GetAttribute(ctx, namePath, &stateName); diags.HasError() {
		return
	}

	if configName.Equal(stateName) {
		resp.PlanValue = req.StateValue
	}
}

// clearUnconfiguredTagExistsFields nulls if_tag_exists and if_not_exists in the
// plan wherever the configuration leaves them unset. The plan modifiers fill in
// the derived value so plans stay stable, but the API request should only carry
// the field the user configured.
func clearUnconfiguredTagExistsFields(plan *tagPipelineRulesetModel, config tagPipelineRulesetModel) {
	for i := range plan.Rules {
		if i >= len(config.Rules) {
			return
		}
		planRule, configRule := &plan.Rules[i], config.Rules[i]

		if planRule.Mapping != nil && configRule.Mapping != nil {
			if configRule.Mapping.IfTagExists.IsNull() {
				planRule.Mapping.IfTagExists = types.StringNull()
			}
			if configRule.Mapping.IfNotExists.IsNull() {
				planRule.Mapping.IfNotExists = types.BoolNull()
			}
		}
		if planRule.Query != nil && configRule.Query != nil {
			if configRule.Query.IfTagExists.IsNull() {
				planRule.Query.IfTagExists = types.StringNull()
			}
			if configRule.Query.IfNotExists.IsNull() {
				planRule.Query.IfNotExists = types.BoolNull()
			}
		}
		if planRule.ReferenceTable != nil && configRule.ReferenceTable != nil {
			if configRule.ReferenceTable.IfTagExists.IsNull() {
				planRule.ReferenceTable.IfTagExists = types.StringNull()
			}
			if configRule.ReferenceTable.IfNotExists.IsNull() {
				planRule.ReferenceTable.IfNotExists = types.BoolNull()
			}
		}
	}
}
