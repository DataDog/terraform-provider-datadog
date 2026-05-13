package fwprovider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.ResourceWithConfigure   = &securityMonitoringDefaultRuleResource{}
	_ resource.ResourceWithImportState = &securityMonitoringDefaultRuleResource{}
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
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource for default rules. It can only be imported, you can't create a default rule.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"custom_message": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Custom Message (will override default message) for generated signals.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"custom_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name (will override default name) of the rule.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
							Optional:    true,
							Computed:    true,
							Description: "The aggregation type. For Signal Correlation rules, it must be event_count.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
							},
						},
						"distinct_fields": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Field for which the cardinality is measured. Sent as an array.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
							},
						},
						"group_by_fields": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Fields to group by.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
							},
						},
						"has_optional_group_by_fields": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Description: "When false, events without a group-by value are ignored by the rule. When true, events with missing group-by fields are processed with `N/A`, replacing the missing values.",
							PlanModifiers: []planmodifier.Bool{
								boolplanmodifier.UseStateForUnknown(),
							},
						},
						"data_source": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Source of events.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								validators.SecurityMonitoringDataSourceWarningValidator(),
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringStandardDataSourceFromValue),
							},
						},
						"metric": schema.StringAttribute{
							Optional:           true,
							Computed:           true,
							Description:        "The target field to aggregate over when using the `sum`, `max`, or `geo_data` aggregations.",
							DeprecationMessage: "Configure `metrics` instead. This attribute will be removed in the next major version of the provider.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"metrics": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Group of target fields to aggregate over when using the `sum`, `max`, `geo_data`, or `new_value` aggregations. The `sum`, `max`, and `geo_data` aggregations only accept one value in this list, whereas the `new_value` aggregation accepts up to five values.",
							PlanModifiers: []planmodifier.List{
								listplanmodifier.UseStateForUnknown(),
							},
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Name of the query. Not compatible with `new_value` aggregations.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"query": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Query to run on logs.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"custom_query_extension": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Description: "Query extension to append to the logs query.",
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
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
	if err := utils.CheckForUnparsed(ruleResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	rule := ruleResponse.SecurityMonitoringStandardRuleResponse
	if rule == nil {
		response.Diagnostics.AddError("unsupported rule type", "signal rule type is not currently supported")
		return
	}

	response.Diagnostics.Append(updateDefaultRuleResourceDataFromResponse(ctx, &state, rule)...)

	response.Diagnostics.Append(securityMonitoringDefaultRuleDeprecationWarning(rule)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func updateDefaultRuleResourceDataFromResponse(ctx context.Context, state *securityMonitoringDefaultRuleResourceModel, ruleResponse *datadogV2.SecurityMonitoringStandardRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Enabled = types.BoolValue(ruleResponse.GetIsEnabled())

	if customMessage, ok := ruleResponse.GetCustomMessageOk(); ok {
		state.CustomMessage = types.StringValue(*customMessage)
	} else {
		state.CustomMessage = types.StringNull()
	}
	if customName, ok := ruleResponse.GetCustomNameOk(); ok {
		state.CustomName = types.StringValue(*customName)
	} else {
		state.CustomName = types.StringNull()
	}

	var caseDiags diag.Diagnostics
	state.Cases, caseDiags = extractDefaultRuleCases(ctx, ruleResponse.GetCases())
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
	state.Queries, queryDiags = extractDefaultRuleQueries(ctx, ruleResponse.GetQueries())
	diags.Append(queryDiags...)

	var tagsDiags diag.Diagnostics
	state.CustomTags, tagsDiags = extractDefaultRuleCustomTags(ctx, ruleResponse.GetTags(), ruleResponse.GetDefaultTags())
	diags.Append(tagsDiags...)

	return diags
}

func extractDefaultRuleCases(ctx context.Context, responseRuleCases []datadogV2.SecurityMonitoringRuleCase) ([]defaultRuleCaseModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	stateCases := make([]defaultRuleCaseModel, len(responseRuleCases))
	for idx, apiCase := range responseRuleCases {
		stateCase := defaultRuleCaseModel{
			Status:        types.StringValue(string(apiCase.GetStatus())),
			Notifications: types.ListNull(types.StringType),
		}
		if customStatus, ok := apiCase.GetCustomStatusOk(); ok && customStatus != nil {
			stateCase.CustomStatus = types.StringValue(string(*customStatus))
		}
		if notifications, ok := apiCase.GetNotificationsOk(); ok {
			var listDiags diag.Diagnostics
			stateCase.Notifications, listDiags = types.ListValueFrom(ctx, types.StringType, *notifications)
			diags.Append(listDiags...)
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

func extractDefaultRuleQueries(ctx context.Context, responseRuleQueries []datadogV2.SecurityMonitoringStandardRuleQuery) ([]defaultRuleQueryModel, diag.Diagnostics) {
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
	if err := utils.CheckForUnparsed(currentResponse); err != nil {
		response.Diagnostics.AddError("response contains unparsed object", err.Error())
		return
	}

	rule := currentResponse.SecurityMonitoringStandardRuleResponse
	if rule == nil {
		response.Diagnostics.AddError("unsupported rule type", "signal rule type is not currently supported")
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

	state := plan
	state.ID = types.StringValue(ruleID)

	response.Diagnostics.Append(updateDefaultRuleResourceDataFromResponse(ctx, &state, updatedRule)...)

	response.Diagnostics.Append(securityMonitoringDefaultRuleDeprecationWarning(updatedRule)...)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func buildSecMonDefaultRuleUpdatePayload(ctx context.Context, currentState *datadogV2.SecurityMonitoringStandardRuleResponse, plan *securityMonitoringDefaultRuleResourceModel) (*datadogV2.SecurityMonitoringRuleUpdatePayload, bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}
	isSignalCorrelation := isSignalCorrelationSchema(plan.Type)

	isEnabled := plan.Enabled.ValueBool()
	payload.IsEnabled = &isEnabled

	shouldUpdate := false
	if currentState.GetIsEnabled() != isEnabled {
		shouldUpdate = true
	}

	matchedCases := 0
	modifiedCases := 0

	updatedRuleCase := make([]datadogV2.SecurityMonitoringRuleCase, len(currentState.GetCases()))
	for i, ruleCase := range currentState.GetCases() {
		updatedRuleCase[i] = ruleCase

		if planCase, ok := findRuleCaseForStatus(plan.Cases, ruleCase.GetStatus()); ok {

			// Update rule case notifications when rule added to terraform configuration

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
			}
		} else {

			// Clear rule case notifications when rule case removed from terraform configuration

			emptyNotifications := []string{}
			if !stringSliceEquals(emptyNotifications, ruleCase.GetNotifications()) {
				modifiedCases++
				shouldUpdate = true
				updatedRuleCase[i].Notifications = emptyNotifications
			}
		}
	}

	if !isSignalCorrelation && len(plan.Queries) > 0 {
		payloadQueries := make([]datadogV2.SecurityMonitoringRuleQuery, len(plan.Queries))
		for idx, planQuery := range plan.Queries {
			// For default rules, merge with existing query to preserve unspecified fields
			var existingQuery *datadogV2.SecurityMonitoringStandardRuleQuery
			if idx < len(currentState.GetQueries()) {
				existingQuery = &currentState.GetQueries()[idx]
			}
			built, qDiags := buildUpdateDefaultRuleQuery(ctx, &planQuery, existingQuery)
			diags.Append(qDiags...)
			if built != nil {
				payloadQueries[idx] = *built
			}
		}
		payload.SetQueries(payloadQueries)

		// Compare queries including custom_query_extension
		if !compareQueries(currentState.GetQueries(), payloadQueries) {
			shouldUpdate = true
		}
	}

	// custom_message: send whenever plan has a known value (including ""),
	// so removing the user override propagates to the API.
	if !plan.CustomMessage.IsNull() && !plan.CustomMessage.IsUnknown() {
		customMessage := plan.CustomMessage.ValueString()
		payload.SetCustomMessage(customMessage)

		// Check if custom_message exists in current state and compare
		if currentCustomMessage, ok := currentState.GetCustomMessageOk(); ok {
			if *currentCustomMessage != customMessage {
				shouldUpdate = true
			}
		} else if customMessage != "" {
			// Custom message doesn't exist in the current state, so this is a change
			shouldUpdate = true
		}
	}

	if !plan.CustomName.IsNull() && !plan.CustomName.IsUnknown() {
		customName := plan.CustomName.ValueString()
		payload.SetCustomName(customName)
		if currentCustomName, ok := currentState.GetCustomNameOk(); ok {
			if *currentCustomName != customName {
				shouldUpdate = true
			}
		} else if customName != "" {
			// Custom name doesn't exist in the current state, so this is a change
			shouldUpdate = true
		}
	}

	if matchedCases < len(plan.Cases) {
		// Enable partial state so that we don't persist the changes
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
	payload.SetTags(payloadTags)
	if !compareTags(currentState.GetTags(), payloadTags) {
		shouldUpdate = true
	}

	return &payload, shouldUpdate, diags
}

// buildUpdateDefaultRuleQuery merges the plan with the existing API query so
// fields the user didn't author keep their server-side value.
func buildUpdateDefaultRuleQuery(ctx context.Context, planQuery *defaultRuleQueryModel, existingQuery *datadogV2.SecurityMonitoringStandardRuleQuery) (*datadogV2.SecurityMonitoringRuleQuery, diag.Diagnostics) {
	var diags diag.Diagnostics
	payloadQuery := datadogV2.SecurityMonitoringStandardRuleQuery{}

	if existingQuery != nil {
		if planQuery.Aggregation.IsNull() || planQuery.Aggregation.IsUnknown() {
			if v, ok := existingQuery.GetAggregationOk(); ok {
				payloadQuery.SetAggregation(*v)
			}
		}
		if planQuery.GroupByFields.IsNull() || planQuery.GroupByFields.IsUnknown() {
			if v, ok := existingQuery.GetGroupByFieldsOk(); ok {
				payloadQuery.SetGroupByFields(*v)
			}
		}
		if planQuery.HasOptionalGroupByFields.IsNull() || planQuery.HasOptionalGroupByFields.IsUnknown() {
			if v, ok := existingQuery.GetHasOptionalGroupByFieldsOk(); ok {
				payloadQuery.SetHasOptionalGroupByFields(*v)
			}
		}
		if planQuery.DistinctFields.IsNull() || planQuery.DistinctFields.IsUnknown() {
			if v, ok := existingQuery.GetDistinctFieldsOk(); ok {
				payloadQuery.SetDistinctFields(*v)
			}
		}
		if planQuery.DataSource.IsNull() || planQuery.DataSource.IsUnknown() {
			if v, ok := existingQuery.GetDataSourceOk(); ok {
				payloadQuery.SetDataSource(*v)
			}
		}
		if planQuery.Metric.IsNull() || planQuery.Metric.IsUnknown() {
			if v, ok := existingQuery.GetMetricOk(); ok {
				payloadQuery.SetMetric(*v)
			}
		}
		if planQuery.Metrics.IsNull() || planQuery.Metrics.IsUnknown() {
			if v, ok := existingQuery.GetMetricsOk(); ok {
				payloadQuery.SetMetrics(*v)
			}
		}
		if planQuery.Name.IsNull() || planQuery.Name.IsUnknown() {
			if v, ok := existingQuery.GetNameOk(); ok {
				payloadQuery.SetName(*v)
			}
		}
		if planQuery.Query.IsNull() || planQuery.Query.IsUnknown() {
			if v, ok := existingQuery.GetQueryOk(); ok {
				payloadQuery.SetQuery(*v)
			}
		}
		if planQuery.CustomQueryExtension.IsNull() || planQuery.CustomQueryExtension.IsUnknown() {
			if v, ok := existingQuery.GetCustomQueryExtensionOk(); ok {
				payloadQuery.SetCustomQueryExtension(*v)
			}
		}
	}

	if !planQuery.Aggregation.IsNull() && !planQuery.Aggregation.IsUnknown() {
		payloadQuery.SetAggregation(datadogV2.SecurityMonitoringRuleQueryAggregation(planQuery.Aggregation.ValueString()))
	}
	if !planQuery.GroupByFields.IsNull() && !planQuery.GroupByFields.IsUnknown() {
		var groupByFields []string
		listDiags := planQuery.GroupByFields.ElementsAs(ctx, &groupByFields, false)
		diags.Append(listDiags...)
		payloadQuery.SetGroupByFields(groupByFields)
	}
	if !planQuery.HasOptionalGroupByFields.IsNull() && !planQuery.HasOptionalGroupByFields.IsUnknown() {
		payloadQuery.SetHasOptionalGroupByFields(planQuery.HasOptionalGroupByFields.ValueBool())
	}
	if !planQuery.DistinctFields.IsNull() && !planQuery.DistinctFields.IsUnknown() {
		var distinctFields []string
		listDiags := planQuery.DistinctFields.ElementsAs(ctx, &distinctFields, false)
		diags.Append(listDiags...)
		payloadQuery.SetDistinctFields(distinctFields)
	}
	if !planQuery.DataSource.IsNull() && !planQuery.DataSource.IsUnknown() {
		payloadQuery.SetDataSource(datadogV2.SecurityMonitoringStandardDataSource(planQuery.DataSource.ValueString()))
	}
	if !planQuery.Metric.IsNull() && !planQuery.Metric.IsUnknown() {
		payloadQuery.SetMetric(planQuery.Metric.ValueString())
	}
	if !planQuery.Metrics.IsNull() && !planQuery.Metrics.IsUnknown() {
		var metrics []string
		listDiags := planQuery.Metrics.ElementsAs(ctx, &metrics, false)
		diags.Append(listDiags...)
		payloadQuery.SetMetrics(metrics)
	}
	if !planQuery.Name.IsNull() && !planQuery.Name.IsUnknown() {
		payloadQuery.SetName(planQuery.Name.ValueString())
	}
	if !planQuery.Query.IsNull() && !planQuery.Query.IsUnknown() {
		payloadQuery.SetQuery(planQuery.Query.ValueString())
	}
	if !planQuery.CustomQueryExtension.IsNull() && !planQuery.CustomQueryExtension.IsUnknown() {
		payloadQuery.SetCustomQueryExtension(planQuery.CustomQueryExtension.ValueString())
	}

	standardRuleQuery := datadogV2.SecurityMonitoringStandardRuleQueryAsSecurityMonitoringRuleQuery(&payloadQuery)
	return &standardRuleQuery, diags
}

func compareQueries(currentQueries []datadogV2.SecurityMonitoringStandardRuleQuery, payloadQueries []datadogV2.SecurityMonitoringRuleQuery) bool {
	if len(currentQueries) != len(payloadQueries) {
		return false
	}

	// For now, we'll assume queries are different if they exist in the payload
	// This is a simplified approach - in a more complete implementation,
	// we would need to extract the standard query from the payload query
	// and compare each field individually

	// Since we're building the payload from Terraform config and comparing with current state,
	// if there are any queries in the payload, we should check if they differ from current state
	// For simplicity, we'll return false (indicating a change) if there are queries in the payload
	// This ensures that any query changes are detected

	return len(payloadQueries) == 0
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
	return stringSliceEquals(currentTags, payloadTags)
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
	// Compare decrease_criticality_based_on_env
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
