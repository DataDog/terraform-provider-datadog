package fwprovider

import (
	"context"
	"fmt"
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
	response.Diagnostics.AddError("not implemented", "Read is not implemented yet for the framework default rule resource")
}

func updateDefaultRuleResourceDataFromResponse(ctx context.Context, state *securityMonitoringDefaultRuleResourceModel, ruleResponse *datadogV2.SecurityMonitoringStandardRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Enabled = types.BoolValue(ruleResponse.GetIsEnabled())

	// SDKv2 default rule never wrote these to state, hiding API-side drift.
	// Surface it now; pair with an Update-side empty-string clear.
	// TODO: Remove this comment once the second half fix is done
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
		stateQuery := defaultRuleQueryModel{}

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

// When the API returns [] for an Optional list the user didn't author, copy
// the prior null back so post-apply state equals planned value.
func reconcileEmptyDefaultRuleQueryFields(apiState, prior []defaultRuleQueryModel) {
	for i := range apiState {
		hasPrior := i < len(prior)
		if isEmptyKnownList(apiState[i].DistinctFields) {
			if hasPrior {
				apiState[i].DistinctFields = prior[i].DistinctFields
			} else {
				apiState[i].DistinctFields = types.ListNull(types.StringType)
			}
		}
		if isEmptyKnownList(apiState[i].GroupByFields) {
			if hasPrior {
				apiState[i].GroupByFields = prior[i].GroupByFields
			} else {
				apiState[i].GroupByFields = types.ListNull(types.StringType)
			}
		}
		if isEmptyKnownList(apiState[i].Metrics) {
			if hasPrior {
				apiState[i].Metrics = prior[i].Metrics
			} else {
				apiState[i].Metrics = types.ListNull(types.StringType)
			}
		}
	}
}

func (r *securityMonitoringDefaultRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("not implemented", "Update is not implemented yet for the framework default rule resource")
}

func (r *securityMonitoringDefaultRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	// no-op
}

func (r *securityMonitoringDefaultRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}
