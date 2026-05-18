package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure    = &securityMonitoringDefaultRuleResource{}
	_ resource.ResourceWithImportState  = &securityMonitoringDefaultRuleResource{}
	_ resource.ResourceWithUpgradeState = &securityMonitoringDefaultRuleResource{}
)

type securityMonitoringDefaultRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

type securityMonitoringDefaultRuleResourceModel struct {
	ID            types.String `tfsdk:"id"`
	CustomMessage types.String `tfsdk:"custom_message"`
	CustomName    types.String `tfsdk:"custom_name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Type          types.String `tfsdk:"type"`
	CustomTags    types.Set    `tfsdk:"custom_tags"`

	Cases   []defaultRuleCaseModel    `tfsdk:"case"`
	Queries []defaultRuleQueryModel   `tfsdk:"query"`
	Filters []ruleFilterModel         `tfsdk:"filter"`
	Options []defaultRuleOptionsModel `tfsdk:"options"`
}

type defaultRuleCaseModel struct {
	Status        types.String `tfsdk:"status"`
	CustomStatus  types.String `tfsdk:"custom_status"`
	Notifications types.List   `tfsdk:"notifications"`
}

type defaultRuleQueryModel struct {
	AgentRules               []defaultRuleQueryAgentRuleModel `tfsdk:"agent_rule"`
	Aggregation              types.String                     `tfsdk:"aggregation"`
	DistinctFields           types.List                       `tfsdk:"distinct_fields"`
	GroupByFields            types.List                       `tfsdk:"group_by_fields"`
	HasOptionalGroupByFields types.Bool                       `tfsdk:"has_optional_group_by_fields"`
	DataSource               types.String                     `tfsdk:"data_source"`
	Metric                   types.String                     `tfsdk:"metric"`
	Metrics                  types.List                       `tfsdk:"metrics"`
	Name                     types.String                     `tfsdk:"name"`
	Query                    types.String                     `tfsdk:"query"`
	CustomQueryExtension     types.String                     `tfsdk:"custom_query_extension"`
}

type defaultRuleQueryAgentRuleModel struct {
	AgentRuleID types.String `tfsdk:"agent_rule_id"`
	Expression  types.String `tfsdk:"expression"`
}

type defaultRuleOptionsModel struct {
	DecreaseCriticalityBasedOnEnv types.Bool `tfsdk:"decrease_criticality_based_on_env"`
}

func NewSecurityMonitoringDefaultRuleResource() resource.Resource {
	return &securityMonitoringDefaultRuleResource{}
}

func (r *securityMonitoringDefaultRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringDefaultRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_default_rule"
}

func (r *securityMonitoringDefaultRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = securityMonitoringDefaultRuleSchema(1)
}

// securityMonitoringDefaultRuleSchema returns the versioned schema. Version 0 is
// the SDKv2-era schema (same shape); version 1 is the current FW schema after
// migration. Keeping them in one function prevents attribute drift between versions.
func securityMonitoringDefaultRuleSchema(version int64) schema.Schema {
	return schema.Schema{
		Version:     version,
		Description: "Provides a Datadog Security Monitoring Rule API resource for default rules. It can only be imported, you can't create a default rule.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"custom_message": schema.StringAttribute{
				Optional:    true,
				Description: "Custom Message (will override default message) for generated signals.",
			},
			"custom_name": schema.StringAttribute{
				Optional:    true,
				Description: "The name (will override default name) of the rule.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable the rule.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "The rule type.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_tags": schema.SetAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Custom tags for generated signals.",
			},
		},
		Blocks: map[string]schema.Block{
			"case": schema.ListNestedBlock{
				Description: "Cases of the rule, this is used to update notifications.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"status": schema.StringAttribute{
							Required:    true,
							Description: "Status of the rule case to match.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							},
						},
						"custom_status": schema.StringAttribute{
							Optional:    true,
							Description: "Status of the rule case to override.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							},
						},
						"notifications": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Notification targets for each rule case.",
						},
					},
				},
			},
			"query": schema.ListNestedBlock{
				Description: "Queries for selecting logs which are part of the rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"aggregation": schema.StringAttribute{
							Computed:    true,
							Description: "The aggregation type.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"distinct_fields": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Field for which the cardinality is measured. Sent as an array.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"group_by_fields": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Fields to group by.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"has_optional_group_by_fields": schema.BoolAttribute{
							Computed:    true,
							Description: "When false, events without a group-by value are ignored by the rule. When true, events with missing group-by fields are processed with `N/A`, replacing the missing values.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"data_source": schema.StringAttribute{
							Computed:    true,
							Description: "Source of events.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"metric": schema.StringAttribute{
							Computed:           true,
							Description:        "The target field to aggregate over when using the `sum`, `max`, or `geo_data` aggregations.",
							DeprecationMessage: "Configure `metrics` instead. This attribute will be removed in the next major version of the provider.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"metrics": schema.ListAttribute{
							Computed:    true,
							ElementType: types.StringType,
							Description: "Group of target fields to aggregate over when using the `sum`, `max`, `geo_data`, or `new_value` aggregations. The `sum`, `max`, and `geo_data` aggregations only accept one value in this list, whereas the `new_value` aggregation accepts up to five values.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "Name of the query. Not compatible with `new_value` aggregations.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"query": schema.StringAttribute{
							Computed:    true,
							Description: "Query to run on logs.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"custom_query_extension": schema.StringAttribute{
							Optional:    true,
							Description: "Query extension to append to the logs query.",
						},
					},
					Blocks: map[string]schema.Block{
						"agent_rule": schema.ListNestedBlock{
							Description:        "**Deprecated**. It won't be applied anymore.",
							DeprecationMessage: "`agent_rule` has been deprecated in favor of new Agent Rule resource.",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"agent_rule_id": schema.StringAttribute{
										Required:    true,
										Description: "**Deprecated**. It won't be applied anymore.",
									},
									"expression": schema.StringAttribute{
										Required:    true,
										Description: "**Deprecated**. It won't be applied anymore.",
									},
								},
							},
						},
					},
				},
			},
			"filter": schema.ListNestedBlock{
				Description: "Additional queries to filter matched events before they are processed.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The type of filtering action. Allowed enum values: require, suppress",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringFilterActionFromValue),
							},
						},
						"query": schema.StringAttribute{
							Required:    true,
							Description: "Query for selecting logs to apply the filtering action.",
						},
					},
				},
			},
			"options": schema.ListNestedBlock{
				Description: "Options on default rules. Note that only a subset of fields can be updated on default rule options.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"decrease_criticality_based_on_env": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "If true, signals in non-production environments have a lower severity than what is defined by the rule case, which can reduce noise. The decrement is applied when the environment tag of the signal starts with `staging`, `test`, or `dev`. Only available when the rule type is `log_detection`.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

// UpgradeState migrates state written by SDKv2 (schema version 0) to version 1.
// SDKv2 declared case/query/options as Optional+Computed and auto-populated all
// API-returned blocks into state even when the user didn't declare them. FW blocks
// cannot be Computed, so those extra rows must be stripped here. After the upgrade,
// Read will repopulate only the blocks whose statuses/indices are present in the
// user's actual config (via the prior-state filter in Read).
func (r *securityMonitoringDefaultRuleResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	v0 := securityMonitoringDefaultRuleSchema(0)
	return map[int64]resource.StateUpgrader{
		0: {
			PriorSchema: &v0,
			StateUpgrader: func(ctx context.Context, req resource.UpgradeStateRequest, resp *resource.UpgradeStateResponse) {
				var state securityMonitoringDefaultRuleResourceModel
				resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
				if resp.Diagnostics.HasError() {
					return
				}
				state.Cases = nil
				state.Options = nil
				state.Queries = nil
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			},
		},
	}
}

func securityMonitoringDefaultRuleDeprecationWarning(rule *datadogV2.SecurityMonitoringStandardRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	if deprecationTimestampMs, ok := rule.GetDeprecationDateOk(); ok {
		deprecation := time.UnixMilli(*deprecationTimestampMs)
		diags.AddWarning(
			fmt.Sprintf("Rule will be deprecated on %s.", deprecation.Format("Jan _2 2006")),
			"Please consider deleting the associated resource. "+
				"After the depreciation date, the rule will stop triggering signals. "+
				" Moreover, the API will reject any call to update the rule, which might break your Terraform pipeline. "+
				"The Datadog team performs regular audit of all detection rules to maintain high fidelity signal quality. "+
				"We will be replacing this rule with an improved third party detection rule after the depreciation date.",
		)
	}
	return diags
}

func (r *securityMonitoringDefaultRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError(
		"Default rule cannot be created",
		"cannot create a default rule, please import it first before making changes",
	)
}

func (r *securityMonitoringDefaultRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state securityMonitoringDefaultRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	ruleResponse, _, err := r.api.GetSecurityMonitoringRule(r.auth, state.ID.ValueString())
	if err != nil {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error reading security monitoring default rule"))
		return
	}
	rule := ruleResponse.SecurityMonitoringStandardRuleResponse
	if rule == nil {
		response.Diagnostics.AddError("unsupported rule type", "signal rule type is not currently supported")
		return
	}

	if err := utils.CheckForUnparsed(ruleResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	priorCases := state.Cases
	priorOptions := state.Options
	priorCustomMessage := state.CustomMessage
	priorCustomName := state.CustomName
	priorQueries := state.Queries

	response.Diagnostics.Append(updateDefaultRuleResourceDataFromResponse(ctx, &state, rule, priorCustomMessage, priorCustomName, priorQueries, priorCases)...)

	// Status-based filter for case: keep only API cases whose status appeared
	// in the prior state. If prior was empty (fresh import or user never declared
	// cases), set to nil so state matches a no-block config.
	state.Cases = filterCasesByPriorStatuses(state.Cases, priorCases)

	// Index-based truncation for query: keep only as many API queries as the
	// prior state had. Queries are positional; the user must declare all of
	// them to touch any, so prior count equals what the user declared.
	if len(priorQueries) == 0 {
		state.Queries = nil
	} else if len(state.Queries) > len(priorQueries) {
		state.Queries = state.Queries[:len(priorQueries)]
	}

	// Options: keep only when prior state had an options block.
	if len(priorOptions) == 0 {
		state.Options = nil
	}

	response.Diagnostics.Append(securityMonitoringDefaultRuleDeprecationWarning(rule)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// filterCasesByPriorStatuses returns only the API cases whose status appears in
// priorCases. When priorCases is empty, returns nil (no cases in state), which
// matches a config with no case blocks declared.
func filterCasesByPriorStatuses(apiCases []defaultRuleCaseModel, priorCases []defaultRuleCaseModel) []defaultRuleCaseModel {
	if len(priorCases) == 0 {
		return nil
	}
	priorStatuses := make(map[string]bool, len(priorCases))
	for _, c := range priorCases {
		priorStatuses[c.Status.ValueString()] = true
	}
	var filtered []defaultRuleCaseModel
	for _, c := range apiCases {
		if priorStatuses[c.Status.ValueString()] {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// updateDefaultRuleResourceDataFromResponse populates state from the API response.
// "" sentinel carry-forwards for custom_message, custom_name, custom_query_extension,
// and case notifications prevent perpetual diffs when the API omits a field after
// it was explicitly cleared by setting it to "".
func updateDefaultRuleResourceDataFromResponse(
	ctx context.Context,
	state *securityMonitoringDefaultRuleResourceModel,
	ruleResponse *datadogV2.SecurityMonitoringStandardRuleResponse,
	referenceCustomMessage, referenceCustomName types.String,
	referenceQueries []defaultRuleQueryModel,
	referenceCases []defaultRuleCaseModel,
) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Enabled = types.BoolValue(ruleResponse.GetIsEnabled())

	if customMessage, ok := ruleResponse.GetCustomMessageOk(); ok {
		state.CustomMessage = types.StringValue(*customMessage)
	} else if referenceCustomMessage == types.StringValue("") {
		state.CustomMessage = types.StringValue("")
	} else {
		state.CustomMessage = types.StringNull()
	}
	if customName, ok := ruleResponse.GetCustomNameOk(); ok {
		state.CustomName = types.StringValue(*customName)
	} else if referenceCustomName == types.StringValue("") {
		state.CustomName = types.StringValue("")
	} else {
		state.CustomName = types.StringNull()
	}

	var caseDiags diag.Diagnostics
	state.Cases, caseDiags = extractDefaultRuleCases(ctx, ruleResponse.GetCases(), referenceCases)
	diags.Append(caseDiags...)

	if filters, ok := ruleResponse.GetFiltersOk(); ok {
		state.Filters = extractDefaultRuleFilters(*filters)
	}

	if ruleType, ok := ruleResponse.GetTypeOk(); ok {
		state.Type = types.StringValue(string(*ruleType))
	}

	// options are only meaningful for log_detection rules.
	responseOptions := ruleResponse.GetOptions()
	var stateOptions []defaultRuleOptionsModel
	if ruleResponse.GetType() == datadogV2.SECURITYMONITORINGRULETYPEREAD_LOG_DETECTION {
		stateOptions = append(stateOptions, defaultRuleOptionsModel{
			DecreaseCriticalityBasedOnEnv: types.BoolValue(responseOptions.GetDecreaseCriticalityBasedOnEnv()),
		})
	}
	state.Options = stateOptions

	var queryDiags diag.Diagnostics
	state.Queries, queryDiags = extractDefaultRuleQueries(ctx, ruleResponse.GetQueries(), referenceQueries)
	diags.Append(queryDiags...)

	var tagsDiags diag.Diagnostics
	state.CustomTags, tagsDiags = extractDefaultRuleCustomTags(ctx, ruleResponse.GetTags(), ruleResponse.GetDefaultTags())
	diags.Append(tagsDiags...)

	return diags
}

// extractDefaultRuleCases builds case state from the API response.
//
// notifications carry-forward: when the API returns empty notifications ([]) and
// the matching prior case had null Notifications (user never declared the field),
// we restore null so the plan value (null) equals state (null) and no perpetual
// diff occurs. If the prior held an explicit empty list, we keep the empty list so
// the user's explicit "= []" is honoured.
func extractDefaultRuleCases(ctx context.Context, responseRuleCases []datadogV2.SecurityMonitoringRuleCase, referenceCases []defaultRuleCaseModel) ([]defaultRuleCaseModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	refByStatus := make(map[string]defaultRuleCaseModel, len(referenceCases))
	for _, rc := range referenceCases {
		refByStatus[rc.Status.ValueString()] = rc
	}

	stateCases := make([]defaultRuleCaseModel, len(responseRuleCases))
	for idx, apiCase := range responseRuleCases {
		stateCase := defaultRuleCaseModel{
			Status:        types.StringValue(string(apiCase.GetStatus())),
			Notifications: types.ListNull(types.StringType),
		}

		if rawCS, ok := apiCase.GetCustomStatusOk(); ok && rawCS != nil {
			if v := string(*rawCS); v != "" {
				stateCase.CustomStatus = types.StringValue(v)
			}
		}

		if notifications, ok := apiCase.GetNotificationsOk(); ok {
			if len(*notifications) == 0 {
				// API returned empty. Carry forward null when prior was null to
				// prevent a null ↔ [] perpetual diff.
				ref, hasRef := refByStatus[string(apiCase.GetStatus())]
				if hasRef && ref.Notifications.IsNull() {
					stateCase.Notifications = types.ListNull(types.StringType)
				} else {
					var listDiags diag.Diagnostics
					stateCase.Notifications, listDiags = types.ListValueFrom(ctx, types.StringType, *notifications)
					diags.Append(listDiags...)
				}
			} else {
				var listDiags diag.Diagnostics
				stateCase.Notifications, listDiags = types.ListValueFrom(ctx, types.StringType, *notifications)
				diags.Append(listDiags...)
			}
		}

		stateCases[idx] = stateCase
	}
	return stateCases, diags
}

func extractDefaultRuleFilters(responseFilters []datadogV2.SecurityMonitoringFilter) []ruleFilterModel {
	filters := make([]ruleFilterModel, len(responseFilters))
	for idx, responseFilter := range responseFilters {
		filters[idx] = ruleFilterModel{
			Action: types.StringValue(string(responseFilter.GetAction())),
			Query:  types.StringValue(responseFilter.GetQuery()),
		}
	}
	return filters
}

// extractDefaultRuleQueries builds query state from the API response.
// custom_query_extension carries forward only when the matching reference holds ""
// (see updateDefaultRuleResourceDataFromResponse for the rationale).
func extractDefaultRuleQueries(ctx context.Context, responseRuleQueries []datadogV2.SecurityMonitoringStandardRuleQuery, referenceQueries []defaultRuleQueryModel) ([]defaultRuleQueryModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	stateQueries := make([]defaultRuleQueryModel, len(responseRuleQueries))
	for idx, responseQuery := range responseRuleQueries {
		stateQuery := defaultRuleQueryModel{
			DistinctFields: types.ListNull(types.StringType),
			GroupByFields:  types.ListNull(types.StringType),
			Metrics:        types.ListNull(types.StringType),
		}

		if agg, ok := responseQuery.GetAggregationOk(); ok {
			stateQuery.Aggregation = types.StringValue(string(*agg))
		}
		if gbf, ok := responseQuery.GetGroupByFieldsOk(); ok {
			var listDiags diag.Diagnostics
			stateQuery.GroupByFields, listDiags = types.ListValueFrom(ctx, types.StringType, *gbf)
			diags.Append(listDiags...)
		}
		if hasGbf, ok := responseQuery.GetHasOptionalGroupByFieldsOk(); ok {
			stateQuery.HasOptionalGroupByFields = types.BoolValue(*hasGbf)
		}
		if df, ok := responseQuery.GetDistinctFieldsOk(); ok {
			var listDiags diag.Diagnostics
			stateQuery.DistinctFields, listDiags = types.ListValueFrom(ctx, types.StringType, *df)
			diags.Append(listDiags...)
		}
		if ds, ok := responseQuery.GetDataSourceOk(); ok {
			stateQuery.DataSource = types.StringValue(string(*ds))
		}
		if metric, ok := responseQuery.GetMetricOk(); ok {
			stateQuery.Metric = types.StringValue(*metric)
		}
		if m, ok := responseQuery.GetMetricsOk(); ok {
			var listDiags diag.Diagnostics
			stateQuery.Metrics, listDiags = types.ListValueFrom(ctx, types.StringType, *m)
			diags.Append(listDiags...)
		}
		if n, ok := responseQuery.GetNameOk(); ok {
			stateQuery.Name = types.StringValue(*n)
		}
		if q, ok := responseQuery.GetQueryOk(); ok {
			stateQuery.Query = types.StringValue(*q)
		}
		if cqe, ok := responseQuery.GetCustomQueryExtensionOk(); ok {
			stateQuery.CustomQueryExtension = types.StringValue(*cqe)
		} else if idx < len(referenceQueries) && referenceQueries[idx].CustomQueryExtension == types.StringValue("") {
			stateQuery.CustomQueryExtension = types.StringValue("")
		}

		stateQueries[idx] = stateQuery
	}
	return stateQueries, diags
}

// custom_tags = api.tags - api.default_tags. Returns null when empty so
// "user never set custom_tags" doesn't render as an empty set.
func extractDefaultRuleCustomTags(ctx context.Context, apiTags []string, apiDefaultTags []string) (types.Set, diag.Diagnostics) {
	defaultTags := make(map[string]bool, len(apiDefaultTags))
	for _, t := range apiDefaultTags {
		defaultTags[t] = true
	}
	customTags := make([]string, 0, len(apiTags))
	for _, tag := range apiTags {
		if _, ok := defaultTags[tag]; !ok {
			customTags = append(customTags, tag)
		}
	}
	if len(customTags) == 0 {
		return types.SetNull(types.StringType), nil
	}
	return types.SetValueFrom(ctx, types.StringType, customTags)
}

func (r *securityMonitoringDefaultRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan securityMonitoringDefaultRuleResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	var priorState securityMonitoringDefaultRuleResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &priorState)...)
	if response.Diagnostics.HasError() {
		return
	}

	ruleID := priorState.ID.ValueString()
	currentResponse, httpResponse, err := r.api.GetSecurityMonitoringRule(r.auth, ruleID)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			response.Diagnostics.AddError("default rule not found", "default rule does not exist")
			return
		}
		response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error fetching default rule"))
		return
	}
	rule := currentResponse.SecurityMonitoringStandardRuleResponse
	if rule == nil {
		response.Diagnostics.AddError("unsupported rule type", "signal rule type is not currently supported")
		return
	}
	if err := utils.CheckForUnparsed(currentResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}
	if !rule.GetIsDefault() {
		response.Diagnostics.Append(utils.FrameworkErrorDiag(errors.New("rule is not a default rule"), "cannot update non-default rule"))
		return
	}

	payload, shouldUpdate, payloadDiags := buildSecMonDefaultRuleUpdatePayload(ctx, rule, &plan)
	response.Diagnostics.Append(payloadDiags...)
	if response.Diagnostics.HasError() {
		return
	}

	updatedRule := rule
	if shouldUpdate {
		updateResponse, _, err := r.api.UpdateSecurityMonitoringRule(r.auth, ruleID, *payload)
		if err != nil {
			response.Diagnostics.Append(utils.FrameworkErrorDiag(err, "error updating security monitoring default rule"))
			return
		}
		if err := utils.CheckForUnparsed(updateResponse); err != nil {
			response.Diagnostics.AddError("response contains unparsed object", err.Error())
			return
		}
		updatedRule = updateResponse.SecurityMonitoringStandardRuleResponse
	}

	// Use plan as the source of truth for blocks so that post-apply state equals
	// plan exactly — the FW consistency requirement. Only scalar fields that the
	// API owns are overridden from the API response.
	state := plan
	state.ID = types.StringValue(ruleID)

	state.Enabled = types.BoolValue(updatedRule.GetIsEnabled())

	if ruleType, ok := updatedRule.GetTypeOk(); ok {
		state.Type = types.StringValue(string(*ruleType))
	}

	if customMessage, ok := updatedRule.GetCustomMessageOk(); ok {
		state.CustomMessage = types.StringValue(*customMessage)
	} else if plan.CustomMessage == types.StringValue("") {
		state.CustomMessage = types.StringValue("")
	} else {
		state.CustomMessage = types.StringNull()
	}
	if customName, ok := updatedRule.GetCustomNameOk(); ok {
		state.CustomName = types.StringValue(*customName)
	} else if plan.CustomName == types.StringValue("") {
		state.CustomName = types.StringValue("")
	} else {
		state.CustomName = types.StringNull()
	}

	var tagsDiags diag.Diagnostics
	state.CustomTags, tagsDiags = extractDefaultRuleCustomTags(ctx, updatedRule.GetTags(), updatedRule.GetDefaultTags())
	response.Diagnostics.Append(tagsDiags...)

	if filters, ok := updatedRule.GetFiltersOk(); ok {
		state.Filters = extractDefaultRuleFilters(*filters)
	}

	// Blocks (Cases, Queries, Options) stay verbatim from plan (set via state := plan above).

	response.Diagnostics.Append(securityMonitoringDefaultRuleDeprecationWarning(updatedRule)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func buildSecMonDefaultRuleUpdatePayload(ctx context.Context, currentState *datadogV2.SecurityMonitoringStandardRuleResponse, plan *securityMonitoringDefaultRuleResourceModel) (*datadogV2.SecurityMonitoringRuleUpdatePayload, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}
	isSignalCorrelation := isSignalCorrelationSchema(plan.Type)

	isEnabled := plan.Enabled.ValueBool()
	payload.IsEnabled = &isEnabled

	shouldUpdate := currentState.GetIsEnabled() != isEnabled

	matchedCases := 0
	modifiedCases := 0

	updatedRuleCase := make([]datadogV2.SecurityMonitoringRuleCase, len(currentState.GetCases()))
	for i, ruleCase := range currentState.GetCases() {
		updatedRuleCase[i] = ruleCase

		if planCase, ok := findRuleCaseForStatus(plan.Cases, ruleCase.GetStatus()); ok {

			matchedCases++

			var planNotifications []string
			if !planCase.Notifications.IsNull() && !planCase.Notifications.IsUnknown() {
				notifDiags := planCase.Notifications.ElementsAs(ctx, &planNotifications, false)
				diags.Append(notifDiags...)
			}
			if planNotifications == nil {
				planNotifications = []string{}
			}
			if !stringSliceEquals(planNotifications, ruleCase.GetNotifications()) {
				modifiedCases++
				shouldUpdate = true
				updatedRuleCase[i].Notifications = planNotifications
			}

			if !planCase.CustomStatus.IsNull() && !planCase.CustomStatus.IsUnknown() && planCase.CustomStatus.ValueString() != "" {
				planCustomStatus := datadogV2.SecurityMonitoringRuleSeverity(planCase.CustomStatus.ValueString())
				if planCustomStatus != ruleCase.GetCustomStatus() {
					modifiedCases++
					shouldUpdate = true
					updatedRuleCase[i].CustomStatus = &planCustomStatus
				}
			} else if cs := ruleCase.GetCustomStatus(); string(cs) != "" {
				updatedRuleCase[i].CustomStatus = nil
				modifiedCases++
				shouldUpdate = true
			}

		}
		// Un-declared cases are left identity (no notification changes sent to API) —
		// the block is the "I manage this case" toggle. Within a declared case, an
		// absent or empty notifications field clears them via the API (handled above).
	}

	if !isSignalCorrelation && len(plan.Queries) > 0 {
		if len(plan.Queries) != len(currentState.GetQueries()) {
			diags.AddError(
				"query block count mismatch",
				fmt.Sprintf("rule %s has %d query blocks in the API; declare all of them (at minimum as empty) to manage any",
					currentState.GetId(), len(currentState.GetQueries())),
			)
			return nil, false, diags
		}
		payloadQueries := make([]datadogV2.SecurityMonitoringRuleQuery, len(plan.Queries))
		for idx, planQuery := range plan.Queries {
			var existingQuery *datadogV2.SecurityMonitoringStandardRuleQuery
			if idx < len(currentState.GetQueries()) {
				existingQuery = &currentState.GetQueries()[idx]
			}
			built := buildUpdateDefaultRuleQuery(&planQuery, existingQuery)
			if built != nil {
				payloadQueries[idx] = *built
			}
		}
		payload.SetQueries(payloadQueries)

		if !compareQueries(currentState.GetQueries(), payloadQueries) {
			shouldUpdate = true
		}
	}

	if !plan.CustomMessage.IsNull() && !plan.CustomMessage.IsUnknown() {
		customMessage := plan.CustomMessage.ValueString()
		payload.SetCustomMessage(customMessage)
		if currentCustomMessage, ok := currentState.GetCustomMessageOk(); ok {
			if *currentCustomMessage != customMessage {
				shouldUpdate = true
			}
		} else if customMessage != "" {
			shouldUpdate = true
		}
	} else if _, ok := currentState.GetCustomMessageOk(); ok {
		payload.SetCustomMessage("")
		shouldUpdate = true
	}

	if !plan.CustomName.IsNull() && !plan.CustomName.IsUnknown() {
		customName := plan.CustomName.ValueString()
		payload.SetCustomName(customName)
		if currentCustomName, ok := currentState.GetCustomNameOk(); ok {
			if *currentCustomName != customName {
				shouldUpdate = true
			}
		} else if customName != "" {
			shouldUpdate = true
		}
	} else if _, ok := currentState.GetCustomNameOk(); ok {
		payload.SetCustomName("")
		shouldUpdate = true
	}

	if matchedCases < len(plan.Cases) {
		diags.AddError(
			"invalid case",
			"attempted to update notifications for non-existing case for rule "+currentState.GetId(),
		)
		return nil, false, diags
	}

	if modifiedCases > 0 {
		payload.Cases = updatedRuleCase
	}

	payloadFilters := buildPayloadFilters(plan.Filters)
	if !compareFilters(currentState.GetFilters(), payloadFilters) {
		payload.Filters = payloadFilters
		shouldUpdate = true
	}

	if len(plan.Options) > 0 {
		payloadOptions := buildPayloadDefaultRuleOptions(plan.Options, string(currentState.GetType()))
		payload.SetOptions(*payloadOptions)
		currentOptions := currentState.GetOptions()
		if !compareOptions(&currentOptions, payloadOptions) {
			shouldUpdate = true
		}
	}

	defaultTags := currentState.GetDefaultTags()
	tagSet := make(map[string]bool, len(defaultTags))
	for _, tag := range defaultTags {
		tagSet[tag] = true
	}
	if !plan.CustomTags.IsNull() && !plan.CustomTags.IsUnknown() {
		var customTags []string
		tagsDiags := plan.CustomTags.ElementsAs(ctx, &customTags, false)
		diags.Append(tagsDiags...)
		for _, t := range customTags {
			tagSet[t] = true
		}
	}
	payloadTags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		payloadTags = append(payloadTags, tag)
	}
	sort.Strings(payloadTags)
	payload.SetTags(payloadTags)
	if !compareTags(currentState.GetTags(), payloadTags) {
		shouldUpdate = true
	}

	return &payload, shouldUpdate, diags
}

// buildUpdateDefaultRuleQuery builds a query payload for a default rule update.
// For default rules, only custom_query_extension is writable; all other fields
// are owned by Datadog and must always come from the existing API state.
func buildUpdateDefaultRuleQuery(planQuery *defaultRuleQueryModel, existingQuery *datadogV2.SecurityMonitoringStandardRuleQuery) *datadogV2.SecurityMonitoringRuleQuery {
	payloadQuery := datadogV2.SecurityMonitoringStandardRuleQuery{}

	if existingQuery != nil {
		if v, ok := existingQuery.GetAggregationOk(); ok {
			payloadQuery.SetAggregation(*v)
		}
		if v, ok := existingQuery.GetGroupByFieldsOk(); ok {
			payloadQuery.SetGroupByFields(*v)
		}
		if v, ok := existingQuery.GetHasOptionalGroupByFieldsOk(); ok {
			payloadQuery.SetHasOptionalGroupByFields(*v)
		}
		if v, ok := existingQuery.GetDistinctFieldsOk(); ok {
			payloadQuery.SetDistinctFields(*v)
		}
		if v, ok := existingQuery.GetDataSourceOk(); ok {
			payloadQuery.SetDataSource(*v)
		}
		if v, ok := existingQuery.GetMetricOk(); ok {
			payloadQuery.SetMetric(*v)
		}
		if v, ok := existingQuery.GetMetricsOk(); ok {
			payloadQuery.SetMetrics(*v)
		}
		if v, ok := existingQuery.GetNameOk(); ok {
			payloadQuery.SetName(*v)
		}
		if v, ok := existingQuery.GetQueryOk(); ok {
			payloadQuery.SetQuery(*v)
		}
	}

	if !planQuery.CustomQueryExtension.IsNull() && !planQuery.CustomQueryExtension.IsUnknown() {
		payloadQuery.SetCustomQueryExtension(planQuery.CustomQueryExtension.ValueString())
	} else {
		payloadQuery.SetCustomQueryExtension("")
	}

	standardRuleQuery := datadogV2.SecurityMonitoringStandardRuleQueryAsSecurityMonitoringRuleQuery(&payloadQuery)
	return &standardRuleQuery
}

// compareQueries reports whether the query payloads represent no change.
// Only custom_query_extension is writable on default rules.
func compareQueries(currentQueries []datadogV2.SecurityMonitoringStandardRuleQuery, payloadQueries []datadogV2.SecurityMonitoringRuleQuery) bool {
	if len(currentQueries) != len(payloadQueries) {
		return false
	}
	for i, current := range currentQueries {
		payload := payloadQueries[i].SecurityMonitoringStandardRuleQuery
		if payload == nil {
			return false
		}
		if current.GetCustomQueryExtension() != payload.GetCustomQueryExtension() {
			return false
		}
	}
	return true
}

func compareFilters(currentFilters, payloadFilters []datadogV2.SecurityMonitoringFilter) bool {
	if len(currentFilters) != len(payloadFilters) {
		return false
	}
	for i, currentFilter := range currentFilters {
		if currentFilter.GetAction() != payloadFilters[i].GetAction() {
			return false
		}
		if currentFilter.GetQuery() != payloadFilters[i].GetQuery() {
			return false
		}
	}
	return true
}

func compareTags(currentTags, payloadTags []string) bool {
	if len(currentTags) != len(payloadTags) {
		return false
	}
	sorted := make([]string, len(currentTags))
	copy(sorted, currentTags)
	sort.Strings(sorted)
	return stringSliceEquals(sorted, payloadTags)
}

func stringSliceEquals(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func compareOptions(currentOptions, payloadOptions *datadogV2.SecurityMonitoringRuleOptions) bool {
	if currentOptions == nil && payloadOptions == nil {
		return true
	}
	if currentOptions == nil || payloadOptions == nil {
		return false
	}
	return currentOptions.GetDecreaseCriticalityBasedOnEnv() == payloadOptions.GetDecreaseCriticalityBasedOnEnv()
}

func findRuleCaseForStatus(planCases []defaultRuleCaseModel, status datadogV2.SecurityMonitoringRuleSeverity) (*defaultRuleCaseModel, bool) {
	for i := range planCases {
		if datadogV2.SecurityMonitoringRuleSeverity(planCases[i].Status.ValueString()) == status {
			return &planCases[i], true
		}
	}
	return nil, false
}

// decrease_criticality_based_on_env is only valid for log_detection rules.
func buildPayloadDefaultRuleOptions(planOptions []defaultRuleOptionsModel, ruleType string) *datadogV2.SecurityMonitoringRuleOptions {
	payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
	opt := planOptions[0]
	if ruleType == string(datadogV2.SECURITYMONITORINGRULETYPEREAD_LOG_DETECTION) &&
		!opt.DecreaseCriticalityBasedOnEnv.IsNull() && !opt.DecreaseCriticalityBasedOnEnv.IsUnknown() {
		payloadOptions.SetDecreaseCriticalityBasedOnEnv(opt.DecreaseCriticalityBasedOnEnv.ValueBool())
	}
	return payloadOptions
}

func (r *securityMonitoringDefaultRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// no-op
}

func (r *securityMonitoringDefaultRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
