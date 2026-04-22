package fwprovider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	fwutils "github.com/terraform-providers/terraform-provider-datadog/datadog/internal/fwutils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"
)

var (
	_ resource.ResourceWithConfigure   = &securityMonitoringRuleResource{}
	_ resource.ResourceWithImportState = &securityMonitoringRuleResource{}
)

type securityMonitoringRuleResource struct {
	api  *datadogV2.SecurityMonitoringApi
	auth context.Context
}

type securityMonitoringRuleResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Message          types.String `tfsdk:"message"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	HasExtendedTitle types.Bool   `tfsdk:"has_extended_title"`
	Type             types.String `tfsdk:"type"`
	Tags             types.Set    `tfsdk:"tags"`
	GroupSignalsBy   types.List   `tfsdk:"group_signals_by"`
	Validate         types.Bool   `tfsdk:"validate"`

	Cases             []ruleCaseModel           `tfsdk:"case"`
	ThirdPartyCases   []thirdPartyCaseModel     `tfsdk:"third_party_case"`
	Queries           []ruleQueryModel          `tfsdk:"query"`
	SignalQueries     []signalQueryModel        `tfsdk:"signal_query"`
	Filters           []ruleFilterModel         `tfsdk:"filter"`
	ReferenceTables   []ruleReferenceTableModel `tfsdk:"reference_tables"`
	CalculatedFields  []calculatedFieldModel    `tfsdk:"calculated_field"`
	SchedulingOptions []schedulingOptionsModel  `tfsdk:"scheduling_options"`
	Options           []ruleOptionsModel        `tfsdk:"options"`
}

type ruleCaseModel struct {
	Name          types.String          `tfsdk:"name"`
	Condition     types.String          `tfsdk:"condition"`
	Notifications types.List            `tfsdk:"notifications"`
	Status        types.String          `tfsdk:"status"`
	Actions       []ruleCaseActionModel `tfsdk:"action"`
}

type ruleCaseActionModel struct {
	Type    types.String                 `tfsdk:"type"`
	Options []ruleCaseActionOptionsModel `tfsdk:"options"`
}

type ruleCaseActionOptionsModel struct {
	Duration types.Int64 `tfsdk:"duration"`
}

type thirdPartyCaseModel struct {
	Name          types.String `tfsdk:"name"`
	Query         types.String `tfsdk:"query"`
	Notifications types.List   `tfsdk:"notifications"`
	Status        types.String `tfsdk:"status"`
}

type ruleQueryModel struct {
	AgentRules               []ruleQueryAgentRuleModel `tfsdk:"agent_rule"`
	Aggregation              types.String              `tfsdk:"aggregation"`
	DistinctFields           types.List                `tfsdk:"distinct_fields"`
	GroupByFields            types.List                `tfsdk:"group_by_fields"`
	HasOptionalGroupByFields types.Bool                `tfsdk:"has_optional_group_by_fields"`
	DataSource               types.String              `tfsdk:"data_source"`
	Metric                   types.String              `tfsdk:"metric"`
	Metrics                  types.List                `tfsdk:"metrics"`
	Name                     types.String              `tfsdk:"name"`
	Query                    types.String              `tfsdk:"query"`
	Indexes                  types.List                `tfsdk:"indexes"`
}

type ruleQueryAgentRuleModel struct {
	AgentRuleID types.String `tfsdk:"agent_rule_id"`
	Expression  types.String `tfsdk:"expression"`
}

type signalQueryModel struct {
	Aggregation          types.String `tfsdk:"aggregation"`
	Name                 types.String `tfsdk:"name"`
	CorrelatedByFields   types.List   `tfsdk:"correlated_by_fields"`
	CorrelatedQueryIndex types.String `tfsdk:"correlated_query_index"`
	RuleID               types.String `tfsdk:"rule_id"`
	DefaultRuleID        types.String `tfsdk:"default_rule_id"`
}

type ruleFilterModel struct {
	Query  types.String `tfsdk:"query"`
	Action types.String `tfsdk:"action"`
}

type ruleReferenceTableModel struct {
	TableName     types.String `tfsdk:"table_name"`
	ColumnName    types.String `tfsdk:"column_name"`
	LogFieldPath  types.String `tfsdk:"log_field_path"`
	RuleQueryName types.String `tfsdk:"rule_query_name"`
	CheckPresence types.Bool   `tfsdk:"check_presence"`
}

type calculatedFieldModel struct {
	Name       types.String `tfsdk:"name"`
	Expression types.String `tfsdk:"expression"`
}

type schedulingOptionsModel struct {
	Rrule    types.String `tfsdk:"rrule"`
	Start    types.String `tfsdk:"start"`
	Timezone types.String `tfsdk:"timezone"`
}

type ruleOptionsModel struct {
	DetectionMethod               types.String                    `tfsdk:"detection_method"`
	EvaluationWindow              types.Int64                     `tfsdk:"evaluation_window"`
	KeepAlive                     types.Int64                     `tfsdk:"keep_alive"`
	MaxSignalDuration             types.Int64                     `tfsdk:"max_signal_duration"`
	DecreaseCriticalityBasedOnEnv types.Bool                      `tfsdk:"decrease_criticality_based_on_env"`
	NewValueOptions               []newValueOptionsModel          `tfsdk:"new_value_options"`
	ImpossibleTravelOptions       []impossibleTravelOptionsModel  `tfsdk:"impossible_travel_options"`
	AnomalyDetectionOptions       []anomalyDetectionOptionsModel  `tfsdk:"anomaly_detection_options"`
	ThirdPartyRuleOptions         []thirdPartyRuleOptionsModel    `tfsdk:"third_party_rule_options"`
	SequenceDetectionOptions      []sequenceDetectionOptionsModel `tfsdk:"sequence_detection_options"`
}

type newValueOptionsModel struct {
	LearningMethod        types.String `tfsdk:"learning_method"`
	LearningDuration      types.Int64  `tfsdk:"learning_duration"`
	LearningThreshold     types.Int64  `tfsdk:"learning_threshold"`
	ForgetAfter           types.Int64  `tfsdk:"forget_after"`
	InstantaneousBaseline types.Bool   `tfsdk:"instantaneous_baseline"`
}

type impossibleTravelOptionsModel struct {
	BaselineUserLocations types.Bool `tfsdk:"baseline_user_locations"`
}

type anomalyDetectionOptionsModel struct {
	BucketDuration         types.Int64 `tfsdk:"bucket_duration"`
	LearningDuration       types.Int64 `tfsdk:"learning_duration"`
	DetectionTolerance     types.Int64 `tfsdk:"detection_tolerance"`
	LearningPeriodBaseline types.Int64 `tfsdk:"learning_period_baseline"`
	InstantaneousBaseline  types.Bool  `tfsdk:"instantaneous_baseline"`
}

type thirdPartyRuleOptionsModel struct {
	DefaultNotifications types.List                 `tfsdk:"default_notifications"`
	DefaultStatus        types.String               `tfsdk:"default_status"`
	SignalTitleTemplate  types.String               `tfsdk:"signal_title_template"`
	RootQueries          []thirdPartyRootQueryModel `tfsdk:"root_query"`
}

type thirdPartyRootQueryModel struct {
	Query         types.String `tfsdk:"query"`
	GroupByFields types.List   `tfsdk:"group_by_fields"`
}

type sequenceDetectionOptionsModel struct {
	Steps           []sequenceStepModel           `tfsdk:"steps"`
	StepTransitions []sequenceStepTransitionModel `tfsdk:"step_transitions"`
}

type sequenceStepModel struct {
	Name             types.String `tfsdk:"name"`
	Condition        types.String `tfsdk:"condition"`
	EvaluationWindow types.Int64  `tfsdk:"evaluation_window"`
}

type sequenceStepTransitionModel struct {
	Parent           types.String `tfsdk:"parent"`
	Child            types.String `tfsdk:"child"`
	EvaluationWindow types.Int64  `tfsdk:"evaluation_window"`
}

func NewSecurityMonitoringRuleResource() resource.Resource {
	return &securityMonitoringRuleResource{}
}

func (r *securityMonitoringRuleResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	providerData := request.ProviderData.(*FrameworkProvider)
	r.api = providerData.DatadogApiInstances.GetSecurityMonitoringApiV2()
	r.auth = providerData.Auth
}

func (r *securityMonitoringRuleResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "security_monitoring_rule"
}

func (r *securityMonitoringRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Description: "Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule, use `datadog_security_monitoring_default_rule` instead.",
		Attributes: map[string]schema.Attribute{
			"id": utils.ResourceIDAttribute(),
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the rule.",
			},
			"message": schema.StringAttribute{
				Required:    true,
				Description: "Message for generated signals.",
			},
			"enabled": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the rule is enabled.",
			},
			"has_extended_title": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the notifications include the triggering group-by values in their title.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("log_detection"),
				Description: "The rule type.",
				Validators: []validator.String{
					stringvalidator.OneOf(
						string(datadogV2.SECURITYMONITORINGRULETYPEREAD_APPLICATION_SECURITY),
						string(datadogV2.SECURITYMONITORINGRULETYPEREAD_LOG_DETECTION),
						string(datadogV2.SECURITYMONITORINGRULETYPEREAD_WORKLOAD_SECURITY),
						string(datadogV2.SECURITYMONITORINGSIGNALRULETYPE_SIGNAL_CORRELATION),
					),
				},
			},
			"tags": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Tags for generated signals. Note: if default tags are present at provider level, they will be added to this resource.",
			},
			"group_signals_by": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Additional grouping to perform on top of the query grouping.",
			},
			"validate": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether or not to validate the Rule.",
			},
		},
		Blocks: map[string]schema.Block{
			"case": schema.ListNestedBlock{
				Description: "Cases for generating signals.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the case.",
						},
						"condition": schema.StringAttribute{
							Optional:    true,
							Description: "A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.",
						},
						"notifications": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Notification targets for each rule case.",
						},
						"status": schema.StringAttribute{
							Required:    true,
							Description: "Severity of the Security Signal.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							},
						},
					},
					Blocks: map[string]schema.Block{
						"action": schema.ListNestedBlock{
							Description: "Action to perform when the case trigger",
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"type": schema.StringAttribute{
										Required:    true,
										Description: "Type of action to perform when the case triggers.",
										Validators: []validator.String{
											validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleCaseActionTypeFromValue),
										},
									},
								},
								Blocks: map[string]schema.Block{
									"options": schema.ListNestedBlock{
										Description: "Options for the action.",
										Validators: []validator.List{
											listvalidator.SizeAtMost(1),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"duration": schema.Int64Attribute{
													Optional:    true,
													Description: "Duration of the action in seconds.",
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
			"third_party_case": schema.ListNestedBlock{
				Description: "Cases for generating signals for third-party rules. Only required and accepted for third-party rules",
				Validators: []validator.List{
					listvalidator.SizeAtMost(10),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the case.",
						},
						"query": schema.StringAttribute{
							Optional:    true,
							Description: "A query to associate a third-party event to this case.",
						},
						"notifications": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Notification targets for each rule case.",
						},
						"status": schema.StringAttribute{
							Required:    true,
							Description: "Severity of the Security Signal.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
							},
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
							Default:     stringdefault.StaticString(string(datadogV2.SECURITYMONITORINGRULEQUERYAGGREGATION_COUNT)),
							Description: "The aggregation type. For Signal Correlation rules, it must be event_count.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
							},
						},
						"distinct_fields": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Field for which the cardinality is measured. Sent as an array.",
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
							},
						},
						"group_by_fields": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Fields to group by.",
							Validators: []validator.List{
								listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
							},
						},
						"has_optional_group_by_fields": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "When false, events without a group-by value are ignored by the rule. When true, events with missing group-by fields are processed with `N/A`, replacing the missing values.",
						},
						"data_source": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(string(datadogV2.SECURITYMONITORINGSTANDARDDATASOURCE_LOGS)),
							Description: "Source of events.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringStandardDataSourceFromValue),
							},
						},
						"metric": schema.StringAttribute{
							Optional:           true,
							Description:        "The target field to aggregate over when using the `sum`, `max`, or `geo_data` aggregations.",
							DeprecationMessage: "Configure `metrics` instead. This attribute will be removed in the next major version of the provider.",
						},
						"metrics": schema.ListAttribute{
							Optional:    true,
							Computed:    true,
							ElementType: types.StringType,
							Description: "Group of target fields to aggregate over when using the `sum`, `max`, `geo_data`, or `new_value` aggregations. The `sum`, `max`, and `geo_data` aggregations only accept one value in this list, whereas the `new_value` aggregation accepts up to five values.",
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the query. Not compatible with `new_value` aggregations.",
						},
						"query": schema.StringAttribute{
							Required:    true,
							Description: "Query to run on logs.",
						},
						"indexes": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of indexes to run the query on when the data source is `logs`. Supports only one element. Used only for scheduled rules (in other words, when `scheduling_options` is defined).",
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
			"signal_query": schema.ListNestedBlock{
				Description: "Queries for selecting logs which are part of the rule.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"aggregation": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(string(datadogV2.SECURITYMONITORINGRULEQUERYAGGREGATION_EVENT_COUNT)),
							Description: "The aggregation type. For Signal Correlation rules, it must be event_count.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
							},
						},
						"name": schema.StringAttribute{
							Optional:    true,
							Description: "Name of the query. Not compatible with `new_value` aggregations.",
						},
						"correlated_by_fields": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "Fields to correlate by.",
						},
						"correlated_query_index": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
							Description: "Index of the rule query used to retrieve the correlated field. An empty string applies correlation on the non-projected per query attributes of the rule.",
						},
						"rule_id": schema.StringAttribute{
							Required:    true,
							Description: "Rule ID of the signal to correlate.",
						},
						"default_rule_id": schema.StringAttribute{
							Optional:    true,
							Description: "Default Rule ID of the signal to correlate. This value is READ-ONLY.",
						},
					},
				},
			},
			"filter": schema.ListNestedBlock{
				Description: "Additional queries to filter matched events before they are processed. **Note**: This field is deprecated for log detection, signal correlation, and workload security rules.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"query": schema.StringAttribute{
							Required:    true,
							Description: "Query for selecting logs to apply the filtering action.",
						},
						"action": schema.StringAttribute{
							Required:    true,
							Description: "The type of filtering action.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringFilterActionFromValue),
							},
						},
					},
				},
			},
			"reference_tables": schema.ListNestedBlock{
				Description: "Reference tables for filtering query results.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"table_name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the reference table.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"column_name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the column in the reference table.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"log_field_path": schema.StringAttribute{
							Required:    true,
							Description: "The field in the log that should be matched against the reference table.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"rule_query_name": schema.StringAttribute{
							Required:    true,
							Description: "The name of the query to filter.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"check_presence": schema.BoolAttribute{
							Required:    true,
							Description: "Whether to include or exclude logs that match the reference table.",
						},
					},
				},
			},
			"calculated_field": schema.ListNestedBlock{
				Description: "One or more calculated fields. Available only for scheduled rules (in other words, when `scheduling_options` is defined).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Field name.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"expression": schema.StringAttribute{
							Required:    true,
							Description: "Expression.",
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
			},
			"scheduling_options": schema.ListNestedBlock{
				Description: "Options for scheduled rules. When this field is present, the rule runs based on the schedule. When absent, it runs in real time on ingested logs.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"rrule": schema.StringAttribute{
							Required:    true,
							Description: "Schedule for the rule queries, written in RRULE syntax. See [RFC](https://icalendar.org/iCalendar-RFC-5545/3-8-5-3-recurrence-rule.html) for syntax reference.",
						},
						"start": schema.StringAttribute{
							Required:    true,
							Description: "Start date for the schedule, in ISO 8601 format without timezone.",
						},
						"timezone": schema.StringAttribute{
							Required:    true,
							Description: "Time zone of the start date, in the [tz database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) format.",
						},
					},
				},
			},
			"options": schema.ListNestedBlock{
				Description: "Options on rules.",
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"detection_method": schema.StringAttribute{
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString("threshold"),
							Description: "The detection method.",
							Validators: []validator.String{
								validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleDetectionMethodFromValue),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"evaluation_window": schema.Int64Attribute{
							Optional:    true,
							Description: "A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.",
							Validators: []validator.Int64{
								validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
							},
						},
						"keep_alive": schema.Int64Attribute{
							Optional:    true,
							Description: "Once a signal is generated, the signal will remain \"open\" if a case is matched at least once within this keep alive window (in seconds).",
							Validators: []validator.Int64{
								validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleKeepAliveFromValue),
							},
						},
						"max_signal_duration": schema.Int64Attribute{
							Optional:    true,
							Description: "A signal will \"close\" regardless of the query being matched once the time exceeds the maximum duration (in seconds). This time is calculated from the first seen timestamp.",
							Validators: []validator.Int64{
								validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleMaxSignalDurationFromValue),
							},
						},
						"decrease_criticality_based_on_env": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "If true, signals in non-production environments have a lower severity than what is defined by the rule case, which can reduce noise. The decrement is applied when the environment tag of the signal starts with `staging`, `test`, or `dev`. Only available when the rule type is `log_detection`.",
						},
					},
					Blocks: map[string]schema.Block{
						"new_value_options": schema.ListNestedBlock{
							Description: "New value rules specific options.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"learning_method": schema.StringAttribute{
										Optional:    true,
										Computed:    true,
										Default:     stringdefault.StaticString("duration"),
										Description: "The learning method used to determine when signals should be generated for values that weren't learned.",
										Validators: []validator.String{
											validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningMethodFromValue),
										},
									},
									"learning_duration": schema.Int64Attribute{
										Optional:    true,
										Computed:    true,
										Default:     int64default.StaticInt64(1),
										Description: "The duration in days during which values are learned, and after which signals will be generated for values that weren't learned. If set to 0, a signal will be generated for all new values after the first value is learned.",
										Validators: []validator.Int64{
											validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningDurationFromValue),
										},
									},
									"learning_threshold": schema.Int64Attribute{
										Optional:    true,
										Computed:    true,
										Default:     int64default.StaticInt64(0),
										Description: "A number of occurrences after which signals are generated for values that weren't learned.",
										Validators: []validator.Int64{
											validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningThresholdFromValue),
										},
									},
									"forget_after": schema.Int64Attribute{
										Required:    true,
										Description: "The duration in days after which a learned value is forgotten.",
										Validators: []validator.Int64{
											validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleNewValueOptionsForgetAfterFromValue),
										},
									},
									"instantaneous_baseline": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "When set to true, Datadog uses previous values that fall within the defined learning window to construct the baseline, enabling the system to establish an accurate baseline more rapidly rather than relying solely on gradual learning over time.",
									},
								},
							},
						},
						"impossible_travel_options": schema.ListNestedBlock{
							Description: "Options for rules using the impossible travel detection method.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"baseline_user_locations": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "If true, signals are suppressed for the first 24 hours. During that time, Datadog learns the user's regular access locations. This can be helpful to reduce noise and infer VPN usage or credentialed API access.",
									},
								},
							},
						},
						"anomaly_detection_options": schema.ListNestedBlock{
							Description: "Options for rules using the anomaly detection method.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"bucket_duration": schema.Int64Attribute{
										Optional:    true,
										Description: "Duration in seconds of the time buckets used to aggregate events matched by the rule. Valid values are 300, 600, 900, 1800, 3600, 10800.",
										Validators: []validator.Int64{
											validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptionsBucketDurationFromValue),
										},
									},
									"learning_duration": schema.Int64Attribute{
										Optional:    true,
										Description: "Learning duration in hours. Anomaly detection waits for at least this amount of historical data before it starts evaluating. Valid values are 1, 6, 12, 24, 48, 168, 336.",
										Validators: []validator.Int64{
											validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptionsLearningDurationFromValue),
										},
									},
									"detection_tolerance": schema.Int64Attribute{
										Optional:    true,
										Description: "An optional parameter that sets how permissive anomaly detection is. Higher values require higher deviations before triggering a signal. Valid values are 1, 2, 3, 4, 5.",
										Validators: []validator.Int64{
											validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptionsDetectionToleranceFromValue),
										},
									},
									"learning_period_baseline": schema.Int64Attribute{
										Optional:    true,
										Description: "An optional override baseline to apply while the rule is in the learning period. Must be greater than or equal to 0.",
										Validators: []validator.Int64{
											int64validator.AtLeast(0),
										},
									},
									"instantaneous_baseline": schema.BoolAttribute{
										Optional:    true,
										Computed:    true,
										Default:     booldefault.StaticBool(false),
										Description: "When set to true, Datadog uses previous values that fall within the defined learning window to construct the baseline, enabling the system to establish an accurate baseline more rapidly rather than relying solely on gradual learning over time.",
									},
								},
							},
						},
						"third_party_rule_options": schema.ListNestedBlock{
							Description: "Options for rules using the third-party detection method.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									"default_notifications": schema.ListAttribute{
										Optional:    true,
										ElementType: types.StringType,
										Description: "Notification targets for the default rule case, when none of the third-party cases match.",
									},
									"default_status": schema.StringAttribute{
										Required:    true,
										Description: "Severity of the default rule case, when none of the third-party cases match.",
										Validators: []validator.String{
											validators.NewEnumValidator[validator.String](datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
										},
									},
									"signal_title_template": schema.StringAttribute{
										Optional:    true,
										Description: "A template for the signal title; if omitted, the title is generated based on the case name.",
									},
								},
								Blocks: map[string]schema.Block{
									"root_query": schema.ListNestedBlock{
										Description: "Queries to be combined with third-party case queries. Each of them can have different group by fields, to aggregate differently based on the type of alert.",
										Validators: []validator.List{
											listvalidator.IsRequired(),
											listvalidator.SizeAtLeast(1),
											listvalidator.SizeAtMost(10),
										},
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"query": schema.StringAttribute{
													Required:    true,
													Description: "Query to filter logs.",
												},
												"group_by_fields": schema.ListAttribute{
													Optional:    true,
													ElementType: types.StringType,
													Description: "Fields to group by. If empty, each log triggers a signal.",
													Validators: []validator.List{
														listvalidator.ValueStringsAre(stringvalidator.LengthAtLeast(1)),
													},
												},
											},
										},
									},
								},
							},
						},
						"sequence_detection_options": schema.ListNestedBlock{
							Description: "Options for rules using the sequence detection method.",
							Validators: []validator.List{
								listvalidator.SizeAtMost(1),
							},
							NestedObject: schema.NestedBlockObject{
								Blocks: map[string]schema.Block{
									"steps": schema.ListNestedBlock{
										Description: "Sequence steps.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"name": schema.StringAttribute{
													Required:    true,
													Description: "Unique name of the step.",
												},
												"condition": schema.StringAttribute{
													Required:    true,
													Description: "Condition for the step to match.",
												},
												"evaluation_window": schema.Int64Attribute{
													Optional:    true,
													Description: "Evaluation window for the step.",
													Validators: []validator.Int64{
														validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
													},
												},
											},
										},
									},
									"step_transitions": schema.ListNestedBlock{
										Description: "Edges of the step graph.",
										NestedObject: schema.NestedBlockObject{
											Attributes: map[string]schema.Attribute{
												"parent": schema.StringAttribute{
													Required:    true,
													Description: "Parent step name.",
												},
												"child": schema.StringAttribute{
													Required:    true,
													Description: "Child step name.",
												},
												"evaluation_window": schema.Int64Attribute{
													Optional:    true,
													Description: "Maximum time allowed to transition from parent to child.",
													Validators: []validator.Int64{
														validators.NewEnumValidator[validator.Int64](datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
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
			},
		},
	}
}

// securityMonitoringRuleInterface Common Interface to securityMonitoringRuleCreateInterface and SecurityMonitoringRuleReadInterface
type securityMonitoringRuleInterface interface {
	GetFilters() []datadogV2.SecurityMonitoringFilter
	GetFiltersOk() (*[]datadogV2.SecurityMonitoringFilter, bool)
	SetFilters(v []datadogV2.SecurityMonitoringFilter)
	GetHasExtendedTitle() bool
	GetHasExtendedTitleOk() (*bool, bool)
	SetHasExtendedTitle(v bool)
	GetIsEnabled() bool
	GetIsEnabledOk() (*bool, bool)
	SetIsEnabled(v bool)
	GetMessage() string
	GetMessageOk() (*string, bool)
	SetMessage(v string)
	GetName() string
	GetNameOk() (*string, bool)
	SetName(v string)
	GetOptions() datadogV2.SecurityMonitoringRuleOptions
	GetOptionsOk() (*datadogV2.SecurityMonitoringRuleOptions, bool)
	SetOptions(v datadogV2.SecurityMonitoringRuleOptions)
	GetTags() []string
	GetTagsOk() (*[]string, bool)
	SetTags(v []string)
}

// securityMonitoringRuleCreateInterface Common interface to SecurityMonitoringStandardRuleCreatePayload and SecurityMonitoringSignalRuleCreatePayload
type securityMonitoringRuleCreateInterface interface {
	securityMonitoringRuleInterface
	SetCases(v []datadogV2.SecurityMonitoringRuleCaseCreate)
	GetCases() []datadogV2.SecurityMonitoringRuleCaseCreate
}

// securityMonitoringRuleResponseInterface Common interface to SecurityMonitoringStandardRuleResponse and SecurityMonitoringSignalRuleResponse
type securityMonitoringRuleResponseInterface interface {
	securityMonitoringRuleInterface
	SetCases(v []datadogV2.SecurityMonitoringRuleCase)
	GetCases() []datadogV2.SecurityMonitoringRuleCase
	GetDeprecationDateOk() (*int64, bool)
}

// Null-preservation rules applied throughout the readers below:
//   - Required fields: always set from API.
//   - Optional+Computed fields: always set from API; plan modifiers reconcile drift.
//   - Optional-only string: set only when API returns a non-empty value.
//   - Optional-only int64: set only when API returns a non-zero value.
//   - Optional-only list: set only when API returned a non-empty list (ok && len>0).
func updateCommonResourceDataFromResponse(ctx context.Context, state *securityMonitoringRuleResourceModel, ruleResponse securityMonitoringRuleResponseInterface) diag.Diagnostics {
	var diags diag.Diagnostics

	state.Message = types.StringValue(ruleResponse.GetMessage())
	state.Name = types.StringValue(ruleResponse.GetName())
	state.HasExtendedTitle = types.BoolValue(ruleResponse.GetHasExtendedTitle())
	state.Enabled = types.BoolValue(ruleResponse.GetIsEnabled())
	state.Tags = fwutils.ToTerraformSetString(ctx, ruleResponse.GetTagsOk)

	if filters, ok := ruleResponse.GetFiltersOk(); ok {
		state.Filters = extractFiltersFromRuleResponse(*filters)
	}

	var optsDiags diag.Diagnostics
	state.Options, optsDiags = extractTfOptions(ctx, ruleResponse.GetOptions())
	diags.Append(optsDiags...)

	return diags
}

func extractThirdPartyCases(ctx context.Context, responseThirdPartyCases []datadogV2.SecurityMonitoringThirdPartyRuleCase) ([]thirdPartyCaseModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	tfThirdPartyCases := make([]thirdPartyCaseModel, len(responseThirdPartyCases))
	for idx, thirdPartyCase := range responseThirdPartyCases {
		tfThirdPartyCase := thirdPartyCaseModel{
			Status: types.StringValue(string(thirdPartyCase.GetStatus())),
		}
		if v, ok := thirdPartyCase.GetNameOk(); ok && *v != "" {
			tfThirdPartyCase.Name = types.StringValue(*v)
		}
		if v, ok := thirdPartyCase.GetQueryOk(); ok && *v != "" {
			tfThirdPartyCase.Query = types.StringValue(*v)
		}
		if notifications, ok := thirdPartyCase.GetNotificationsOk(); ok && len(*notifications) > 0 {
			var listDiags diag.Diagnostics
			tfThirdPartyCase.Notifications, listDiags = types.ListValueFrom(ctx, types.StringType, *notifications)
			diags.Append(listDiags...)
		}
		tfThirdPartyCases[idx] = tfThirdPartyCase
	}
	return tfThirdPartyCases, diags
}

func updateStandardResourceDataFromResponse(ctx context.Context, state *securityMonitoringRuleResourceModel, ruleResponse *datadogV2.SecurityMonitoringStandardRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(updateCommonResourceDataFromResponse(ctx, state, ruleResponse)...)

	opts := ruleResponse.GetOptions()
	if opts.GetDetectionMethod() == datadogV2.SECURITYMONITORINGRULEDETECTIONMETHOD_THIRD_PARTY {
		var tpDiags diag.Diagnostics
		state.ThirdPartyCases, tpDiags = extractThirdPartyCases(ctx, ruleResponse.GetThirdPartyCases())
		diags.Append(tpDiags...)
	} else {
		var caseDiags diag.Diagnostics
		state.Cases, caseDiags = extractRuleCases(ctx, ruleResponse.GetCases())
		diags.Append(caseDiags...)

		var queryDiags diag.Diagnostics
		state.Queries, queryDiags = extractStandardRuleQueries(ctx, ruleResponse.GetQueries())
		diags.Append(queryDiags...)
	}

	if ruleType, ok := ruleResponse.GetTypeOk(); ok {
		state.Type = types.StringValue(string(*ruleType))
	}

	if referenceTables, ok := ruleResponse.GetReferenceTablesOk(); ok && len(*referenceTables) > 0 {
		state.ReferenceTables = extractReferenceTables(*referenceTables)
	}

	if groupSignalsBy, ok := ruleResponse.GetGroupSignalsByOk(); ok && len(*groupSignalsBy) > 0 {
		var listDiags diag.Diagnostics
		state.GroupSignalsBy, listDiags = types.ListValueFrom(ctx, types.StringType, *groupSignalsBy)
		diags.Append(listDiags...)
	}

	if calculatedFields, ok := ruleResponse.GetCalculatedFieldsOk(); ok && len(*calculatedFields) > 0 {
		state.CalculatedFields = extractCalculatedFields(*calculatedFields)
	}

	if schedulingOptions, ok := ruleResponse.GetSchedulingOptionsOk(); ok {
		state.SchedulingOptions = extractSchedulingOptions(schedulingOptions)
	}

	return diags
}

func extractStandardRuleQueries(ctx context.Context, responseRuleQueries []datadogV2.SecurityMonitoringStandardRuleQuery) ([]ruleQueryModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	ruleQueries := make([]ruleQueryModel, len(responseRuleQueries))
	for idx, responseRuleQuery := range responseRuleQueries {
		ruleQuery := ruleQueryModel{}

		// Required
		if query, ok := responseRuleQuery.GetQueryOk(); ok {
			ruleQuery.Query = types.StringValue(*query)
		}

		// Optional+Computed with defaults — always set
		if aggregation, ok := responseRuleQuery.GetAggregationOk(); ok {
			ruleQuery.Aggregation = types.StringValue(string(*aggregation))
		}
		if hasGbf, ok := responseRuleQuery.GetHasOptionalGroupByFieldsOk(); ok {
			ruleQuery.HasOptionalGroupByFields = types.BoolValue(*hasGbf)
		}
		if dataSource, ok := responseRuleQuery.GetDataSourceOk(); ok {
			ruleQuery.DataSource = types.StringValue(string(*dataSource))
		}

		// Optional+Computed without default — set even when empty
		if metrics, ok := responseRuleQuery.GetMetricsOk(); ok {
			var listDiags diag.Diagnostics
			ruleQuery.Metrics, listDiags = types.ListValueFrom(ctx, types.StringType, *metrics)
			diags.Append(listDiags...)
		}

		// Optional-only — only set when API returns a meaningful value
		if name, ok := responseRuleQuery.GetNameOk(); ok && *name != "" {
			ruleQuery.Name = types.StringValue(*name)
		}
		if metric, ok := responseRuleQuery.GetMetricOk(); ok && *metric != "" {
			ruleQuery.Metric = types.StringValue(*metric)
		}
		if distinctFields, ok := responseRuleQuery.GetDistinctFieldsOk(); ok && len(*distinctFields) > 0 {
			var listDiags diag.Diagnostics
			ruleQuery.DistinctFields, listDiags = types.ListValueFrom(ctx, types.StringType, *distinctFields)
			diags.Append(listDiags...)
		}
		if groupByFields, ok := responseRuleQuery.GetGroupByFieldsOk(); ok && len(*groupByFields) > 0 {
			var listDiags diag.Diagnostics
			ruleQuery.GroupByFields, listDiags = types.ListValueFrom(ctx, types.StringType, *groupByFields)
			diags.Append(listDiags...)
		}
		// The API returns a single "index" string; our schema stores it as a list.
		if index, ok := responseRuleQuery.GetIndexOk(); ok && *index != "" {
			var listDiags diag.Diagnostics
			ruleQuery.Indexes, listDiags = types.ListValueFrom(ctx, types.StringType, []string{*index})
			diags.Append(listDiags...)
		}

		ruleQueries[idx] = ruleQuery
	}
	return ruleQueries, diags
}

func updateSignalResourceDataFromResponse(ctx context.Context, state *securityMonitoringRuleResourceModel, resp *datadogV2.SecurityMonitoringSignalRuleResponse) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(updateCommonResourceDataFromResponse(ctx, state, resp)...)

	var caseDiags diag.Diagnostics
	state.Cases, caseDiags = extractRuleCases(ctx, resp.GetCases())
	diags.Append(caseDiags...)

	var queryDiags diag.Diagnostics
	state.SignalQueries, queryDiags = extractSignalRuleQueries(ctx, resp.GetQueries())
	diags.Append(queryDiags...)

	if ruleType, ok := resp.GetTypeOk(); ok {
		state.Type = types.StringValue(string(*ruleType))
	}

	return diags
}

func extractSignalRuleQueries(ctx context.Context, responseRuleQueries []datadogV2.SecurityMonitoringSignalRuleResponseQuery) ([]signalQueryModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	ruleQueries := make([]signalQueryModel, len(responseRuleQueries))
	for idx, responseRuleQuery := range responseRuleQueries {
		ruleQuery := signalQueryModel{}

		// Required
		if ruleId, ok := responseRuleQuery.GetRuleIdOk(); ok {
			ruleQuery.RuleID = types.StringValue(*ruleId)
		}

		// Optional+Computed with defaults — always set
		if aggregation, ok := responseRuleQuery.GetAggregationOk(); ok {
			ruleQuery.Aggregation = types.StringValue(string(*aggregation))
		}
		// correlated_query_index is Optional+Computed with Default(""); the API returns int32.
		if correlatedQueryIndex, ok := responseRuleQuery.GetCorrelatedQueryIndexOk(); ok {
			ruleQuery.CorrelatedQueryIndex = types.StringValue(fmt.Sprintf("%d", *correlatedQueryIndex))
		} else {
			ruleQuery.CorrelatedQueryIndex = types.StringValue("")
		}

		// Optional-only — only set when API returns a meaningful value
		if name, ok := responseRuleQuery.GetNameOk(); ok && *name != "" {
			ruleQuery.Name = types.StringValue(*name)
		}
		if defaultRuleId, ok := responseRuleQuery.GetDefaultRuleIdOk(); ok && *defaultRuleId != "" {
			ruleQuery.DefaultRuleID = types.StringValue(*defaultRuleId)
		}
		if correlatedByFields, ok := responseRuleQuery.GetCorrelatedByFieldsOk(); ok && len(*correlatedByFields) > 0 {
			var listDiags diag.Diagnostics
			ruleQuery.CorrelatedByFields, listDiags = types.ListValueFrom(ctx, types.StringType, *correlatedByFields)
			diags.Append(listDiags...)
		}

		ruleQueries[idx] = ruleQuery
	}
	return ruleQueries, diags
}

func extractFiltersFromRuleResponse(ruleResponseFilter []datadogV2.SecurityMonitoringFilter) []ruleFilterModel {
	filters := make([]ruleFilterModel, len(ruleResponseFilter))
	for idx, responseFilter := range ruleResponseFilter {
		filters[idx] = ruleFilterModel{
			Query:  types.StringValue(responseFilter.GetQuery()),
			Action: types.StringValue(string(responseFilter.GetAction())),
		}
	}
	return filters
}

func extractRuleCases(ctx context.Context, responseRulesCases []datadogV2.SecurityMonitoringRuleCase) ([]ruleCaseModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	ruleCases := make([]ruleCaseModel, len(responseRulesCases))
	for idx, responseRuleCase := range responseRulesCases {
		ruleCase := ruleCaseModel{
			Status: types.StringValue(string(responseRuleCase.GetStatus())),
		}
		if name, ok := responseRuleCase.GetNameOk(); ok && *name != "" {
			ruleCase.Name = types.StringValue(*name)
		}
		if condition, ok := responseRuleCase.GetConditionOk(); ok && *condition != "" {
			ruleCase.Condition = types.StringValue(*condition)
		}
		if notification, ok := responseRuleCase.GetNotificationsOk(); ok && len(*notification) > 0 {
			var listDiags diag.Diagnostics
			ruleCase.Notifications, listDiags = types.ListValueFrom(ctx, types.StringType, *notification)
			diags.Append(listDiags...)
		}
		if actions, ok := responseRuleCase.GetActionsOk(); ok && len(*actions) > 0 {
			ruleCase.Actions = extractRuleCaseActions(*actions)
		}
		ruleCases[idx] = ruleCase
	}
	return ruleCases, diags
}

func extractRuleCaseActions(apiActions []datadogV2.SecurityMonitoringRuleCaseAction) []ruleCaseActionModel {
	tfActions := make([]ruleCaseActionModel, len(apiActions))
	for idx, action := range apiActions {
		tfAction := ruleCaseActionModel{
			Type: types.StringValue(string(action.GetType())),
		}
		if options, ok := action.GetOptionsOk(); ok {
			tfOptions := ruleCaseActionOptionsModel{}
			if duration, ok := options.GetDurationOk(); ok {
				tfOptions.Duration = types.Int64Value(*duration)
			}
			if !tfOptions.Duration.IsNull() {
				tfAction.Options = []ruleCaseActionOptionsModel{tfOptions}
			}
		}
		tfActions[idx] = tfAction
	}
	return tfActions
}

func extractTfOptions(ctx context.Context, options datadogV2.SecurityMonitoringRuleOptions) ([]ruleOptionsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	tfOptions := ruleOptionsModel{}

	// Optional+Computed with defaults — always set
	if detectionMethod, ok := options.GetDetectionMethodOk(); ok {
		tfOptions.DetectionMethod = types.StringValue(string(*detectionMethod))
	}
	if decreaseCriticalityBasedOnEnv, ok := options.GetDecreaseCriticalityBasedOnEnvOk(); ok {
		tfOptions.DecreaseCriticalityBasedOnEnv = types.BoolValue(*decreaseCriticalityBasedOnEnv)
	}

	// Optional-only int64 — only set when API returns a non-zero value
	if evaluationWindow, ok := options.GetEvaluationWindowOk(); ok && *evaluationWindow != 0 {
		tfOptions.EvaluationWindow = types.Int64Value(int64(*evaluationWindow))
	}
	if keepAlive, ok := options.GetKeepAliveOk(); ok && *keepAlive != 0 {
		tfOptions.KeepAlive = types.Int64Value(int64(*keepAlive))
	}
	if maxSignalDuration, ok := options.GetMaxSignalDurationOk(); ok && *maxSignalDuration != 0 {
		tfOptions.MaxSignalDuration = types.Int64Value(int64(*maxSignalDuration))
	}

	// Sub-options blocks — only set when API returned them
	if newValueOptions, ok := options.GetNewValueOptionsOk(); ok {
		tfOptions.NewValueOptions = []newValueOptionsModel{extractNewValueOptions(newValueOptions)}
	}
	if impossibleTravelOptions, ok := options.GetImpossibleTravelOptionsOk(); ok {
		tfOptions.ImpossibleTravelOptions = []impossibleTravelOptionsModel{extractImpossibleTravelOptions(impossibleTravelOptions)}
	}
	if anomalyDetectionOptions, ok := options.GetAnomalyDetectionOptionsOk(); ok {
		tfOptions.AnomalyDetectionOptions = []anomalyDetectionOptionsModel{extractAnomalyDetectionOptions(anomalyDetectionOptions)}
	}
	if thirdPartyOptions, ok := options.GetThirdPartyRuleOptionsOk(); ok {
		var tpDiags diag.Diagnostics
		tfOptions.ThirdPartyRuleOptions, tpDiags = extractThirdPartyRuleOptions(ctx, thirdPartyOptions)
		diags.Append(tpDiags...)
	}
	if seqOptions, ok := options.GetSequenceDetectionOptionsOk(); ok {
		tfOptions.SequenceDetectionOptions = []sequenceDetectionOptionsModel{extractSequenceDetectionOptions(seqOptions)}
	}

	return []ruleOptionsModel{tfOptions}, diags
}

func extractReferenceTables(referenceTables []datadogV2.SecurityMonitoringReferenceTable) []ruleReferenceTableModel {
	tfReferenceTables := make([]ruleReferenceTableModel, len(referenceTables))
	for idx, referenceTable := range referenceTables {
		tfReferenceTables[idx] = ruleReferenceTableModel{
			TableName:     types.StringValue(referenceTable.GetTableName()),
			ColumnName:    types.StringValue(referenceTable.GetColumnName()),
			LogFieldPath:  types.StringValue(referenceTable.GetLogFieldPath()),
			RuleQueryName: types.StringValue(referenceTable.GetRuleQueryName()),
			CheckPresence: types.BoolValue(referenceTable.GetCheckPresence()),
		}
	}
	return tfReferenceTables
}

func extractSchedulingOptions(schedulingOptions *datadogV2.SecurityMonitoringSchedulingOptions) []schedulingOptionsModel {
	if schedulingOptions == nil {
		return nil
	}
	tfSchedulingOptions := schedulingOptionsModel{
		Rrule: types.StringValue(schedulingOptions.GetRrule()),
	}
	if start, ok := schedulingOptions.GetStartOk(); ok && *start != "" {
		tfSchedulingOptions.Start = types.StringValue(*start)
	}
	if timezone, ok := schedulingOptions.GetTimezoneOk(); ok && *timezone != "" {
		tfSchedulingOptions.Timezone = types.StringValue(*timezone)
	}
	return []schedulingOptionsModel{tfSchedulingOptions}
}

func extractCalculatedFields(calculatedFields []datadogV2.CalculatedField) []calculatedFieldModel {
	tfCalculatedFields := make([]calculatedFieldModel, len(calculatedFields))
	for idx, calculatedField := range calculatedFields {
		tfCalculatedFields[idx] = calculatedFieldModel{
			Name:       types.StringValue(calculatedField.Name),
			Expression: types.StringValue(calculatedField.Expression),
		}
	}
	return tfCalculatedFields
}

func extractNewValueOptions(newValueOptions *datadogV2.SecurityMonitoringRuleNewValueOptions) newValueOptionsModel {
	return newValueOptionsModel{
		// Required
		ForgetAfter: types.Int64Value(int64(newValueOptions.GetForgetAfter())),
		// Optional+Computed with defaults — always set
		LearningMethod:        types.StringValue(string(newValueOptions.GetLearningMethod())),
		LearningDuration:      types.Int64Value(int64(newValueOptions.GetLearningDuration())),
		LearningThreshold:     types.Int64Value(int64(newValueOptions.GetLearningThreshold())),
		InstantaneousBaseline: types.BoolValue(bool(newValueOptions.GetInstantaneousBaseline())),
	}
}

func extractImpossibleTravelOptions(impossibleTravelOptions *datadogV2.SecurityMonitoringRuleImpossibleTravelOptions) impossibleTravelOptionsModel {
	return impossibleTravelOptionsModel{
		// Optional+Computed with default false — always set
		BaselineUserLocations: types.BoolValue(impossibleTravelOptions.GetBaselineUserLocations()),
	}
}

func extractAnomalyDetectionOptions(anomalyDetectionOptions *datadogV2.SecurityMonitoringRuleAnomalyDetectionOptions) anomalyDetectionOptionsModel {
	tfAnomalyDetectionOptions := anomalyDetectionOptionsModel{
		// Optional+Computed with default false — always set
		InstantaneousBaseline: types.BoolValue(bool(anomalyDetectionOptions.GetInstantaneousBaseline())),
	}
	// Optional-only int64 — only set when non-zero (all valid enum values are > 0)
	if v := anomalyDetectionOptions.GetBucketDuration(); v != 0 {
		tfAnomalyDetectionOptions.BucketDuration = types.Int64Value(int64(v))
	}
	if v := anomalyDetectionOptions.GetLearningDuration(); v != 0 {
		tfAnomalyDetectionOptions.LearningDuration = types.Int64Value(int64(v))
	}
	if v := anomalyDetectionOptions.GetDetectionTolerance(); v != 0 {
		tfAnomalyDetectionOptions.DetectionTolerance = types.Int64Value(int64(v))
	}
	// learning_period_baseline: Optional-only, 0 is a valid override value — use ok form.
	if v, ok := anomalyDetectionOptions.GetLearningPeriodBaselineOk(); ok {
		tfAnomalyDetectionOptions.LearningPeriodBaseline = types.Int64Value(int64(*v))
	}
	return tfAnomalyDetectionOptions
}

func extractThirdPartyRuleOptions(ctx context.Context, thirdPartyOptions *datadogV2.SecurityMonitoringRuleThirdPartyOptions) ([]thirdPartyRuleOptionsModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	tfThirdPartyOptions := thirdPartyRuleOptionsModel{
		DefaultStatus: types.StringValue(string(thirdPartyOptions.GetDefaultStatus())),
	}

	// Optional-only list — keep null when API returns empty
	if v, ok := thirdPartyOptions.GetDefaultNotificationsOk(); ok && len(*v) > 0 {
		var listDiags diag.Diagnostics
		tfThirdPartyOptions.DefaultNotifications, listDiags = types.ListValueFrom(ctx, types.StringType, *v)
		diags.Append(listDiags...)
	}

	// Optional-only string
	if v, ok := thirdPartyOptions.GetSignalTitleTemplateOk(); ok && *v != "" {
		tfThirdPartyOptions.SignalTitleTemplate = types.StringValue(*v)
	}

	tfRootQueries := thirdPartyOptions.GetRootQueries()
	tfThirdPartyOptions.RootQueries = make([]thirdPartyRootQueryModel, len(tfRootQueries))
	for idx, rootQuery := range tfRootQueries {
		tfRootQuery := thirdPartyRootQueryModel{
			Query: types.StringValue(rootQuery.GetQuery()),
		}
		if v, ok := rootQuery.GetGroupByFieldsOk(); ok && len(*v) > 0 {
			var listDiags diag.Diagnostics
			tfRootQuery.GroupByFields, listDiags = types.ListValueFrom(ctx, types.StringType, *v)
			diags.Append(listDiags...)
		}
		tfThirdPartyOptions.RootQueries[idx] = tfRootQuery
	}

	return []thirdPartyRuleOptionsModel{tfThirdPartyOptions}, diags
}

func extractSequenceDetectionOptions(seqOptions *datadogV2.SecurityMonitoringRuleSequenceDetectionOptions) sequenceDetectionOptionsModel {
	tfSeqOptions := sequenceDetectionOptionsModel{}

	steps := seqOptions.GetSteps()
	if len(steps) > 0 {
		tfSeqOptions.Steps = make([]sequenceStepModel, len(steps))
		for idx, step := range steps {
			stepMap := sequenceStepModel{
				Name:      types.StringValue(step.GetName()),
				Condition: types.StringValue(step.GetCondition()),
			}
			// Optional-only int64 — only set when non-zero
			if v, ok := step.GetEvaluationWindowOk(); ok && *v != 0 {
				stepMap.EvaluationWindow = types.Int64Value(int64(*v))
			}
			tfSeqOptions.Steps[idx] = stepMap
		}
	}

	transitions := seqOptions.GetStepTransitions()
	if len(transitions) > 0 {
		tfSeqOptions.StepTransitions = make([]sequenceStepTransitionModel, len(transitions))
		for idx, tr := range transitions {
			trMap := sequenceStepTransitionModel{
				Parent: types.StringValue(tr.GetParent()),
				Child:  types.StringValue(tr.GetChild()),
			}
			// Optional-only int64 — only set when non-zero
			if v, ok := tr.GetEvaluationWindowOk(); ok && *v != 0 {
				trMap.EvaluationWindow = types.Int64Value(int64(*v))
			}
			tfSeqOptions.StepTransitions[idx] = trMap
		}
	}

	return tfSeqOptions
}

func isSignalCorrelationSchema(model *securityMonitoringRuleResourceModel) bool {
	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		_, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(model.Type.ValueString())
		return err == nil
	}
	return false
}

func checkQueryConsistency(model *securityMonitoringRuleResourceModel) error {
	if len(model.Queries) > 0 && len(model.SignalQueries) > 0 {
		return fmt.Errorf("query list and signal query list cannot be both populated")
	}
	isSignalCorrelation := isSignalCorrelationSchema(model)
	if !isSignalCorrelation && len(model.SignalQueries) > 0 {
		return fmt.Errorf("signal query list should not be populated for this rule type")
	}
	if isSignalCorrelation && len(model.Queries) > 0 {
		return fmt.Errorf("query list should not be populated for this rule type")
	}
	return nil
}

func buildCreatePayloadFromModel(ctx context.Context, model *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringRuleCreatePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err := checkQueryConsistency(model); err != nil {
		diags.AddError("invalid query configuration", err.Error())
		return &datadogV2.SecurityMonitoringRuleCreatePayload{}, diags
	}
	if isSignalCorrelationSchema(model) {
		payload, d := buildCreateSignalPayload(ctx, model)
		diags.Append(d...)
		createPayload := datadogV2.SecurityMonitoringSignalRuleCreatePayloadAsSecurityMonitoringRuleCreatePayload(payload)
		return &createPayload, diags
	}
	payload, d := buildCreateStandardPayload(ctx, model)
	diags.Append(d...)
	createPayload := datadogV2.SecurityMonitoringStandardRuleCreatePayloadAsSecurityMonitoringRuleCreatePayload(payload)
	return &createPayload, diags
}

func buildValidatePayloadFromModel(ctx context.Context, model *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringRuleValidatePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	if err := checkQueryConsistency(model); err != nil {
		diags.AddError("invalid query configuration", err.Error())
		return &datadogV2.SecurityMonitoringRuleValidatePayload{}, diags
	}
	if isSignalCorrelationSchema(model) {
		payload, d := buildSignalPayload(ctx, model)
		diags.Append(d...)
		createPayload := datadogV2.SecurityMonitoringSignalRulePayloadAsSecurityMonitoringRuleValidatePayload(payload)
		return &createPayload, diags
	}
	payload, d := buildStandardPayload(ctx, model)
	diags.Append(d...)
	createPayload := datadogV2.SecurityMonitoringStandardRulePayloadAsSecurityMonitoringRuleValidatePayload(payload)
	return &createPayload, diags
}

func buildCreateCommonPayload(ctx context.Context, model *securityMonitoringRuleResourceModel, payload securityMonitoringRuleCreateInterface) {
	payload.SetIsEnabled(model.Enabled.ValueBool())
	payload.SetMessage(model.Message.ValueString())
	payload.SetName(model.Name.ValueString())
	payload.SetHasExtendedTitle(model.HasExtendedTitle.ValueBool())

	if len(model.Options) > 0 {
		ruleType := model.Type.ValueString()
		payloadOptions := buildPayloadOptions(ctx, model.Options, ruleType)
		payload.SetOptions(*payloadOptions)
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		model.Tags.ElementsAs(ctx, &tags, false)
		payload.SetTags(tags)
	}

	if len(model.Filters) > 0 {
		payload.SetFilters(buildPayloadFilters(model.Filters))
	}
}

func isThirdPartyRule(model *securityMonitoringRuleResourceModel) bool {
	if len(model.Options) == 0 {
		return false
	}
	if !model.Options[0].DetectionMethod.IsNull() && !model.Options[0].DetectionMethod.IsUnknown() {
		return datadogV2.SecurityMonitoringRuleDetectionMethod(model.Options[0].DetectionMethod.ValueString()) == datadogV2.SECURITYMONITORINGRULEDETECTIONMETHOD_THIRD_PARTY
	}
	return false
}

func buildCreateStandardPayload(ctx context.Context, model *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringStandardRuleCreatePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := datadogV2.SecurityMonitoringStandardRuleCreatePayload{}
	buildCreateCommonPayload(ctx, model, &payload)

	if isThirdPartyRule(model) {
		payload.SetThirdPartyCases(buildPayloadThirdPartyCases(ctx, model.ThirdPartyCases))
	} else {
		payload.SetCases(buildCreatePayloadCases(ctx, model.Cases))
		payload.SetQueries(buildCreateStandardPayloadQueries(ctx, model.Queries))
	}

	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		if ruleType, err := datadogV2.NewSecurityMonitoringRuleTypeCreateFromValue(model.Type.ValueString()); err == nil {
			payload.SetType(*ruleType)
		} else {
			diags.AddError("invalid rule type", err.Error())
			return &payload, diags
		}
	}

	if len(model.ReferenceTables) > 0 {
		payload.SetReferenceTables(buildPayloadReferenceTables(model.ReferenceTables))
	}

	if !model.GroupSignalsBy.IsNull() && !model.GroupSignalsBy.IsUnknown() && len(model.GroupSignalsBy.Elements()) > 0 {
		var groupSignalsBy []string
		model.GroupSignalsBy.ElementsAs(ctx, &groupSignalsBy, false)
		payload.SetGroupSignalsBy(groupSignalsBy)
	}

	if len(model.SchedulingOptions) > 0 {
		payload.SetSchedulingOptions(*buildPayloadSchedulingOptions(model.SchedulingOptions))
	}

	if len(model.CalculatedFields) > 0 {
		payload.SetCalculatedFields(buildPayloadCalculatedFields(model.CalculatedFields))
	}

	return &payload, diags
}

func buildStandardPayload(ctx context.Context, model *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringStandardRulePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := datadogV2.SecurityMonitoringStandardRulePayload{}
	buildCreateCommonPayload(ctx, model, &payload)

	if isThirdPartyRule(model) {
		payload.SetThirdPartyCases(buildPayloadThirdPartyCases(ctx, model.ThirdPartyCases))
	} else {
		payload.SetCases(buildCreatePayloadCases(ctx, model.Cases))
		payload.SetQueries(buildCreateStandardPayloadQueries(ctx, model.Queries))
	}

	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		if ruleType, err := datadogV2.NewSecurityMonitoringRuleTypeCreateFromValue(model.Type.ValueString()); err == nil {
			payload.SetType(*ruleType)
		} else {
			diags.AddError("invalid rule type", err.Error())
			return &payload, diags
		}
	}

	if len(model.ReferenceTables) > 0 {
		payload.SetReferenceTables(buildPayloadReferenceTables(model.ReferenceTables))
	}

	if !model.GroupSignalsBy.IsNull() && !model.GroupSignalsBy.IsUnknown() && len(model.GroupSignalsBy.Elements()) > 0 {
		var groupSignalsBy []string
		model.GroupSignalsBy.ElementsAs(ctx, &groupSignalsBy, false)
		payload.SetGroupSignalsBy(groupSignalsBy)
	}

	return &payload, diags
}

func buildCreateSignalPayload(ctx context.Context, model *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringSignalRuleCreatePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := datadogV2.SecurityMonitoringSignalRuleCreatePayload{}
	buildCreateCommonPayload(ctx, model, &payload)
	payload.SetCases(buildCreatePayloadCases(ctx, model.Cases))
	queries, d := buildCreateSignalPayloadQueries(ctx, model.SignalQueries)
	diags.Append(d...)
	if !d.HasError() {
		payload.SetQueries(queries)
	} else {
		diags.Append(d...)
		return &payload, diags
	}

	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		if ruleType, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(model.Type.ValueString()); err == nil {
			payload.SetType(*ruleType)
		} else {
			diags.AddError("invalid signal rule type", err.Error())
			return &payload, diags
		}
	}

	return &payload, diags
}

func buildSignalPayload(ctx context.Context, model *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringSignalRulePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := datadogV2.SecurityMonitoringSignalRulePayload{}
	buildCreateCommonPayload(ctx, model, &payload)
	payload.SetCases(buildCreatePayloadCases(ctx, model.Cases))
	queries, d := buildCreateSignalPayloadQueries(ctx, model.SignalQueries)
	if !d.HasError() {
		payload.SetQueries(queries)
	} else {
		diags.Append(d...)
		return &payload, diags
	}

	if !model.Type.IsNull() && !model.Type.IsUnknown() {
		if ruleType, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(model.Type.ValueString()); err == nil {
			payload.SetType(*ruleType)
		} else {
			diags.AddError("invalid signal rule type", err.Error())
			return &payload, diags
		}
	}

	return &payload, diags
}

func buildCreatePayloadCases(ctx context.Context, cases []ruleCaseModel) []datadogV2.SecurityMonitoringRuleCaseCreate {
	payloadCases := make([]datadogV2.SecurityMonitoringRuleCaseCreate, len(cases))
	for idx, c := range cases {
		status := datadogV2.SecurityMonitoringRuleSeverity(c.Status.ValueString())
		structRuleCase := datadogV2.NewSecurityMonitoringRuleCaseCreate(status)
		fwutils.SetOptString(c.Name, structRuleCase.SetName)
		fwutils.SetOptString(c.Condition, structRuleCase.SetCondition)
		fwutils.SetOptStringList(c.Notifications, structRuleCase.SetNotifications, ctx)
		if len(c.Actions) > 0 {
			structRuleCase.SetActions(buildPayloadCaseActions(c.Actions))
		}
		payloadCases[idx] = *structRuleCase
	}
	return payloadCases
}

func buildPayloadThirdPartyCase(ctx context.Context, tc thirdPartyCaseModel) *datadogV2.SecurityMonitoringThirdPartyRuleCaseCreate {
	status := datadogV2.SecurityMonitoringRuleSeverity(tc.Status.ValueString())
	thirdPartyCase := datadogV2.NewSecurityMonitoringThirdPartyRuleCaseCreate(status)
	fwutils.SetOptString(tc.Query, thirdPartyCase.SetQuery)
	fwutils.SetOptString(tc.Name, thirdPartyCase.SetName)
	fwutils.SetOptStringList(tc.Notifications, thirdPartyCase.SetNotifications, ctx)
	return thirdPartyCase
}

func buildPayloadThirdPartyCases(ctx context.Context, cases []thirdPartyCaseModel) []datadogV2.SecurityMonitoringThirdPartyRuleCaseCreate {
	payloadCases := make([]datadogV2.SecurityMonitoringThirdPartyRuleCaseCreate, len(cases))
	for idx, tc := range cases {
		payloadCases[idx] = *buildPayloadThirdPartyCase(ctx, tc)
	}
	return payloadCases
}

func buildPayloadCalculatedFields(fields []calculatedFieldModel) []datadogV2.CalculatedField {
	calculatedFields := make([]datadogV2.CalculatedField, len(fields))
	for idx, f := range fields {
		calculatedFields[idx] = datadogV2.CalculatedField{
			Name:       f.Name.ValueString(),
			Expression: f.Expression.ValueString(),
		}
	}
	return calculatedFields
}

func buildPayloadSchedulingOptions(opts []schedulingOptionsModel) *datadogV2.SecurityMonitoringSchedulingOptions {
	o := opts[0]
	schedulingOptions := datadogV2.NewSecurityMonitoringSchedulingOptions()
	schedulingOptions.SetRrule(o.Rrule.ValueString())
	schedulingOptions.SetStart(o.Start.ValueString())
	schedulingOptions.SetTimezone(o.Timezone.ValueString())
	return schedulingOptions
}

func buildPayloadOptions(ctx context.Context, opts []ruleOptionsModel, ruleType string) *datadogV2.SecurityMonitoringRuleOptions {
	payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
	if len(opts) == 0 {
		return payloadOptions
	}
	tfOptions := opts[0]

	if !tfOptions.DetectionMethod.IsNull() && !tfOptions.DetectionMethod.IsUnknown() {
		detectionMethod := datadogV2.SecurityMonitoringRuleDetectionMethod(tfOptions.DetectionMethod.ValueString())
		payloadOptions.DetectionMethod = &detectionMethod
	}
	if !tfOptions.EvaluationWindow.IsNull() && !tfOptions.EvaluationWindow.IsUnknown() {
		evaluationWindow := datadogV2.SecurityMonitoringRuleEvaluationWindow(tfOptions.EvaluationWindow.ValueInt64())
		payloadOptions.EvaluationWindow = &evaluationWindow
	}
	if !tfOptions.KeepAlive.IsNull() && !tfOptions.KeepAlive.IsUnknown() {
		keepAlive := datadogV2.SecurityMonitoringRuleKeepAlive(tfOptions.KeepAlive.ValueInt64())
		payloadOptions.KeepAlive = &keepAlive
	}
	if !tfOptions.MaxSignalDuration.IsNull() && !tfOptions.MaxSignalDuration.IsUnknown() {
		maxSignalDuration := datadogV2.SecurityMonitoringRuleMaxSignalDuration(tfOptions.MaxSignalDuration.ValueInt64())
		payloadOptions.MaxSignalDuration = &maxSignalDuration
	}
	if !tfOptions.DecreaseCriticalityBasedOnEnv.IsNull() && !tfOptions.DecreaseCriticalityBasedOnEnv.IsUnknown() &&
		ruleType == string(datadogV2.SECURITYMONITORINGRULETYPECREATE_LOG_DETECTION) {
		payloadOptions.SetDecreaseCriticalityBasedOnEnv(tfOptions.DecreaseCriticalityBasedOnEnv.ValueBool())
	}

	if len(tfOptions.NewValueOptions) > 0 {
		if p, ok := buildPayloadNewValueOptions(tfOptions.NewValueOptions); ok {
			payloadOptions.NewValueOptions = p
		}
	}

	if len(tfOptions.ImpossibleTravelOptions) > 0 {
		if p, ok := buildPayloadImpossibleTravelOptions(tfOptions.ImpossibleTravelOptions); ok {
			payloadOptions.ImpossibleTravelOptions = p
		}
	}

	if len(tfOptions.AnomalyDetectionOptions) > 0 {
		if p, ok := buildPayloadAnomalyDetectionOptions(tfOptions.AnomalyDetectionOptions); ok {
			payloadOptions.AnomalyDetectionOptions = p
		}
	}

	if len(tfOptions.ThirdPartyRuleOptions) > 0 {
		if p, ok := buildPayloadThirdPartyRuleOptions(ctx, tfOptions.ThirdPartyRuleOptions); ok {
			payloadOptions.ThirdPartyRuleOptions = p
		}
	}

	if len(tfOptions.SequenceDetectionOptions) > 0 {
		if p, ok := buildPayloadSequenceDetectionOptions(tfOptions.SequenceDetectionOptions); ok {
			payloadOptions.SequenceDetectionOptions = p
		}
	}

	return payloadOptions
}

func buildPayloadImpossibleTravelOptions(opts []impossibleTravelOptionsModel) (*datadogV2.SecurityMonitoringRuleImpossibleTravelOptions, bool) {
	options := datadogV2.NewSecurityMonitoringRuleImpossibleTravelOptions()
	o := opts[0]
	hasPayload := false
	if !o.BaselineUserLocations.IsNull() && !o.BaselineUserLocations.IsUnknown() {
		hasPayload = true
		v := o.BaselineUserLocations.ValueBool()
		options.BaselineUserLocations = &v
	}
	return options, hasPayload
}

func buildPayloadAnomalyDetectionOptions(opts []anomalyDetectionOptionsModel) (*datadogV2.SecurityMonitoringRuleAnomalyDetectionOptions, bool) {
	options := datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptions()
	o := opts[0]
	hasPayload := false

	if !o.BucketDuration.IsNull() && !o.BucketDuration.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleAnomalyDetectionOptionsBucketDuration(o.BucketDuration.ValueInt64())
		options.BucketDuration = &v
	}
	if !o.LearningDuration.IsNull() && !o.LearningDuration.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleAnomalyDetectionOptionsLearningDuration(o.LearningDuration.ValueInt64())
		options.LearningDuration = &v
	}
	if !o.DetectionTolerance.IsNull() && !o.DetectionTolerance.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleAnomalyDetectionOptionsDetectionTolerance(o.DetectionTolerance.ValueInt64())
		options.DetectionTolerance = &v
	}
	if !o.InstantaneousBaseline.IsNull() && !o.InstantaneousBaseline.IsUnknown() {
		hasPayload = true
		options.SetInstantaneousBaseline(o.InstantaneousBaseline.ValueBool())
	}
	// Optional-only: 0 is a valid value ("immediately generate signals"), null means not set.
	if !o.LearningPeriodBaseline.IsNull() && !o.LearningPeriodBaseline.IsUnknown() {
		hasPayload = true
		v := o.LearningPeriodBaseline.ValueInt64()
		options.LearningPeriodBaseline = &v
	}

	return options, hasPayload
}

func buildPayloadNewValueOptions(opts []newValueOptionsModel) (*datadogV2.SecurityMonitoringRuleNewValueOptions, bool) {
	o := opts[0]
	payload := datadogV2.NewSecurityMonitoringRuleNewValueOptions()
	hasPayload := false

	if !o.LearningMethod.IsNull() && !o.LearningMethod.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningMethod(o.LearningMethod.ValueString())
		payload.LearningMethod = &v
	}
	if !o.LearningDuration.IsNull() && !o.LearningDuration.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningDuration(o.LearningDuration.ValueInt64())
		payload.LearningDuration = &v
	}
	if !o.LearningThreshold.IsNull() && !o.LearningThreshold.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningThreshold(o.LearningThreshold.ValueInt64())
		payload.LearningThreshold = &v
	}
	if !o.ForgetAfter.IsNull() && !o.ForgetAfter.IsUnknown() {
		hasPayload = true
		v := datadogV2.SecurityMonitoringRuleNewValueOptionsForgetAfter(o.ForgetAfter.ValueInt64())
		payload.ForgetAfter = &v
	}
	if !o.InstantaneousBaseline.IsNull() && !o.InstantaneousBaseline.IsUnknown() {
		hasPayload = true
		payload.SetInstantaneousBaseline(o.InstantaneousBaseline.ValueBool())
	}

	return payload, hasPayload
}

func buildPayloadThirdPartyRuleOptions(ctx context.Context, opts []thirdPartyRuleOptionsModel) (*datadogV2.SecurityMonitoringRuleThirdPartyOptions, bool) {
	o := opts[0]
	payload := datadogV2.NewSecurityMonitoringRuleThirdPartyOptions()
	hasPayload := false

	if !o.DefaultStatus.IsNull() && !o.DefaultStatus.IsUnknown() {
		hasPayload = true
		payload.SetDefaultStatus(datadogV2.SecurityMonitoringRuleSeverity(o.DefaultStatus.ValueString()))
	}

	if !o.DefaultNotifications.IsNull() && !o.DefaultNotifications.IsUnknown() {
		var notifications []string
		o.DefaultNotifications.ElementsAs(ctx, &notifications, false)
		if len(notifications) > 0 {
			hasPayload = true
		}
		payload.SetDefaultNotifications(notifications)
	}

	if !o.SignalTitleTemplate.IsNull() && !o.SignalTitleTemplate.IsUnknown() {
		hasPayload = true
		payload.SetSignalTitleTemplate(o.SignalTitleTemplate.ValueString())
	}

	if len(o.RootQueries) > 0 {
		hasPayload = true
		payloadRootQueries := make([]datadogV2.SecurityMonitoringThirdPartyRootQuery, len(o.RootQueries))
		for idx, rq := range o.RootQueries {
			payloadRootQueries[idx] = *buildRootQueryPayload(ctx, rq)
		}
		payload.SetRootQueries(payloadRootQueries)
	}

	return payload, hasPayload
}

func buildPayloadSequenceDetectionOptions(opts []sequenceDetectionOptionsModel) (*datadogV2.SecurityMonitoringRuleSequenceDetectionOptions, bool) {
	o := opts[0]
	options := datadogV2.NewSecurityMonitoringRuleSequenceDetectionOptions()
	hasPayload := false

	if len(o.Steps) > 0 {
		hasPayload = true
		payloadSteps := make([]datadogV2.SecurityMonitoringRuleSequenceDetectionStep, len(o.Steps))
		for idx, s := range o.Steps {
			step := datadogV2.SecurityMonitoringRuleSequenceDetectionStep{}
			fwutils.SetOptString(s.Name, step.SetName)
			fwutils.SetOptString(s.Condition, step.SetCondition)
			if !s.EvaluationWindow.IsNull() && !s.EvaluationWindow.IsUnknown() {
				ew := datadogV2.SecurityMonitoringRuleEvaluationWindow(s.EvaluationWindow.ValueInt64())
				step.SetEvaluationWindow(ew)
			}
			payloadSteps[idx] = step
		}
		options.SetSteps(payloadSteps)
	}

	if len(o.StepTransitions) > 0 {
		hasPayload = true
		payloadTransitions := make([]datadogV2.SecurityMonitoringRuleSequenceDetectionStepTransition, len(o.StepTransitions))
		for idx, tr := range o.StepTransitions {
			transition := datadogV2.SecurityMonitoringRuleSequenceDetectionStepTransition{}
			fwutils.SetOptString(tr.Parent, transition.SetParent)
			fwutils.SetOptString(tr.Child, transition.SetChild)
			if !tr.EvaluationWindow.IsNull() && !tr.EvaluationWindow.IsUnknown() {
				ew := datadogV2.SecurityMonitoringRuleEvaluationWindow(tr.EvaluationWindow.ValueInt64())
				transition.SetEvaluationWindow(ew)
			}
			payloadTransitions[idx] = transition
		}
		options.SetStepTransitions(payloadTransitions)
	}

	return options, hasPayload
}

func buildRootQueryPayload(ctx context.Context, rq thirdPartyRootQueryModel) *datadogV2.SecurityMonitoringThirdPartyRootQuery {
	payloadRootQuery := datadogV2.NewSecurityMonitoringThirdPartyRootQuery()
	fwutils.SetOptString(rq.Query, payloadRootQuery.SetQuery)
	fwutils.SetOptStringList(rq.GroupByFields, payloadRootQuery.SetGroupByFields, ctx)
	return payloadRootQuery
}

func buildCreateStandardPayloadQueries(ctx context.Context, queries []ruleQueryModel) []datadogV2.SecurityMonitoringStandardRuleQuery {
	payloadQueries := make([]datadogV2.SecurityMonitoringStandardRuleQuery, len(queries))
	for idx, q := range queries {
		payloadQuery := datadogV2.SecurityMonitoringStandardRuleQuery{}

		fwutils.SetOptString(q.Aggregation, func(v string) {
			agg := datadogV2.SecurityMonitoringRuleQueryAggregation(v)
			payloadQuery.SetAggregation(agg)
		})
		fwutils.SetOptStringList(q.GroupByFields, payloadQuery.SetGroupByFields, ctx)
		if !q.HasOptionalGroupByFields.IsNull() && !q.HasOptionalGroupByFields.IsUnknown() {
			payloadQuery.SetHasOptionalGroupByFields(q.HasOptionalGroupByFields.ValueBool())
		}
		fwutils.SetOptStringList(q.DistinctFields, payloadQuery.SetDistinctFields, ctx)
		fwutils.SetOptString(q.DataSource, func(v string) {
			ds := datadogV2.SecurityMonitoringStandardDataSource(v)
			payloadQuery.SetDataSource(ds)
		})
		fwutils.SetOptString(q.Metric, payloadQuery.SetMetric)
		fwutils.SetOptStringList(q.Metrics, payloadQuery.SetMetrics, ctx)
		fwutils.SetOptString(q.Name, payloadQuery.SetName)

		if !q.Indexes.IsNull() && !q.Indexes.IsUnknown() {
			var indexes []string
			q.Indexes.ElementsAs(ctx, &indexes, false)
			if len(indexes) > 0 {
				payloadQuery.SetIndex(indexes[0])
			}
		}

		payloadQuery.SetQuery(q.Query.ValueString())
		payloadQueries[idx] = payloadQuery
	}
	return payloadQueries
}

func buildCreateSignalPayloadQueries(ctx context.Context, queries []signalQueryModel) ([]datadogV2.SecurityMonitoringSignalRuleQuery, diag.Diagnostics) {
	var diags diag.Diagnostics
	payloadQueries := make([]datadogV2.SecurityMonitoringSignalRuleQuery, len(queries))
	for idx, q := range queries {
		payloadQuery := datadogV2.SecurityMonitoringSignalRuleQuery{}

		fwutils.SetOptString(q.Aggregation, func(v string) {
			agg := datadogV2.SecurityMonitoringRuleQueryAggregation(v)
			payloadQuery.SetAggregation(agg)
		})
		fwutils.SetOptStringList(q.CorrelatedByFields, payloadQuery.SetCorrelatedByFields, ctx)

		if !q.CorrelatedQueryIndex.IsNull() && !q.CorrelatedQueryIndex.IsUnknown() && q.CorrelatedQueryIndex.ValueString() != "" {
			if vInt, err := strconv.Atoi(q.CorrelatedQueryIndex.ValueString()); err == nil {
				payloadQuery.SetCorrelatedQueryIndex(int32(vInt))
			}
		}

		fwutils.SetOptString(q.Name, payloadQuery.SetName)
		payloadQuery.SetRuleId(q.RuleID.ValueString())

		if !q.DefaultRuleID.IsNull() && !q.DefaultRuleID.IsUnknown() && q.DefaultRuleID.ValueString() != "" {
			diags.AddError("invalid field", "defaultRuleId cannot be set")
			return payloadQueries, diags
		}

		payloadQueries[idx] = payloadQuery
	}
	return payloadQueries, diags
}

func buildPayloadFilters(filters []ruleFilterModel) []datadogV2.SecurityMonitoringFilter {
	payloadFilters := make([]datadogV2.SecurityMonitoringFilter, len(filters))
	for idx, f := range filters {
		payloadFilter := datadogV2.SecurityMonitoringFilter{}
		action := datadogV2.SecurityMonitoringFilterAction(f.Action.ValueString())
		payloadFilter.SetAction(action)
		payloadFilter.SetQuery(f.Query.ValueString())
		payloadFilters[idx] = payloadFilter
	}
	return payloadFilters
}

func buildPayloadReferenceTables(tables []ruleReferenceTableModel) []datadogV2.SecurityMonitoringReferenceTable {
	payloadTables := make([]datadogV2.SecurityMonitoringReferenceTable, len(tables))
	for idx, t := range tables {
		rt := datadogV2.SecurityMonitoringReferenceTable{}
		rt.SetTableName(t.TableName.ValueString())
		rt.SetColumnName(t.ColumnName.ValueString())
		rt.SetLogFieldPath(t.LogFieldPath.ValueString())
		rt.SetRuleQueryName(t.RuleQueryName.ValueString())
		rt.SetCheckPresence(t.CheckPresence.ValueBool())
		payloadTables[idx] = rt
	}
	return payloadTables
}

func buildPayloadCaseActions(actions []ruleCaseActionModel) []datadogV2.SecurityMonitoringRuleCaseAction {
	payloadActions := make([]datadogV2.SecurityMonitoringRuleCaseAction, len(actions))
	for idx, a := range actions {
		actionType := datadogV2.SecurityMonitoringRuleCaseActionType(a.Type.ValueString())
		payloadOptions := datadogV2.NewSecurityMonitoringRuleCaseActionOptions()
		if len(a.Options) > 0 {
			fwutils.SetOptInt64(a.Options[0].Duration, payloadOptions.SetDuration)
		}
		payloadActions[idx] = datadogV2.SecurityMonitoringRuleCaseAction{
			Type:    &actionType,
			Options: payloadOptions,
		}
	}
	return payloadActions
}

func buildUpdatePayloadFromModel(ctx context.Context, model, prior *securityMonitoringRuleResourceModel) (*datadogV2.SecurityMonitoringRuleUpdatePayload, diag.Diagnostics) {
	var diags diag.Diagnostics
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}

	if err := checkQueryConsistency(model); err != nil {
		diags.AddError("invalid query configuration", err.Error())
		return &payload, diags
	}
	isSignalCorrelation := isSignalCorrelationSchema(model)

	if isThirdPartyRule(model) {
		payloadThirdPartyCases := make([]datadogV2.SecurityMonitoringThirdPartyRuleCase, len(model.ThirdPartyCases))
		for idx, tc := range model.ThirdPartyCases {
			payloadCase := datadogV2.SecurityMonitoringThirdPartyRuleCase{}
			payloadCase.SetStatus(datadogV2.SecurityMonitoringRuleSeverity(tc.Status.ValueString()))
			fwutils.SetOptStringList(tc.Notifications, payloadCase.SetNotifications, ctx)
			fwutils.SetOptString(tc.Query, payloadCase.SetQuery)
			fwutils.SetOptString(tc.Name, payloadCase.SetName)
			payloadThirdPartyCases[idx] = payloadCase
		}
		payload.SetThirdPartyCases(payloadThirdPartyCases)
	} else {
		payloadCases := make([]datadogV2.SecurityMonitoringRuleCase, len(model.Cases))
		for idx, c := range model.Cases {
			structRuleCase := datadogV2.SecurityMonitoringRuleCase{}
			structRuleCase.SetStatus(datadogV2.SecurityMonitoringRuleSeverity(c.Status.ValueString()))
			fwutils.SetOptString(c.Name, structRuleCase.SetName)
			fwutils.SetOptString(c.Condition, structRuleCase.SetCondition)
			fwutils.SetOptStringList(c.Notifications, structRuleCase.SetNotifications, ctx)
			if len(c.Actions) > 0 {
				structRuleCase.SetActions(buildPayloadCaseActions(c.Actions))
			}
			payloadCases[idx] = structRuleCase
		}
		payload.SetCases(payloadCases)

		var payloadQueries []datadogV2.SecurityMonitoringRuleQuery
		if isSignalCorrelation {
			payloadQueries = make([]datadogV2.SecurityMonitoringRuleQuery, len(model.SignalQueries))
			for idx, q := range model.SignalQueries {
				pq, d := buildUpdateSignalRuleQuery(ctx, q)
				diags.Append(d...)
				if diags.HasError() {
					return &payload, diags
				}
				payloadQueries[idx] = pq
			}
		} else {
			payloadQueries = make([]datadogV2.SecurityMonitoringRuleQuery, len(model.Queries))
			for idx, q := range model.Queries {
				pq := buildUpdateStandardRuleQuery(ctx, q)
				payloadQueries[idx] = *pq
			}
		}
		if len(payloadQueries) > 0 {
			payload.SetQueries(payloadQueries)
		}
	}

	payload.SetIsEnabled(model.Enabled.ValueBool())
	payload.SetHasExtendedTitle(model.HasExtendedTitle.ValueBool())
	fwutils.SetOptString(model.Message, payload.SetMessage)
	fwutils.SetOptString(model.Name, payload.SetName)

	if len(model.Options) > 0 {
		payload.Options = buildPayloadOptions(ctx, model.Options, model.Type.ValueString())
	}

	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		var tags []string
		model.Tags.ElementsAs(ctx, &tags, false)
		payload.SetTags(tags)
	} else {
		payload.SetTags([]string{})
	}

	payload.SetFilters(buildPayloadFilters(model.Filters))

	if !isSignalCorrelation {
		// Mirror SDK behavior: only send reference_tables when present, or empty if removed
		// from config. Leave untouched when it was never configured.
		if len(model.ReferenceTables) > 0 {
			payload.SetReferenceTables(buildPayloadReferenceTables(model.ReferenceTables))
		} else if prior != nil && len(prior.ReferenceTables) > 0 {
			payload.SetReferenceTables([]datadogV2.SecurityMonitoringReferenceTable{})
		}

		// Same pattern for group_signals_by.
		planHasGroupSignalsBy := !model.GroupSignalsBy.IsNull() && !model.GroupSignalsBy.IsUnknown() && len(model.GroupSignalsBy.Elements()) > 0
		priorHadGroupSignalsBy := prior != nil && !prior.GroupSignalsBy.IsNull() && !prior.GroupSignalsBy.IsUnknown() && len(prior.GroupSignalsBy.Elements()) > 0
		if planHasGroupSignalsBy {
			var groupSignalsBy []string
			model.GroupSignalsBy.ElementsAs(ctx, &groupSignalsBy, false)
			payload.SetGroupSignalsBy(groupSignalsBy)
		} else if priorHadGroupSignalsBy {
			payload.SetGroupSignalsBy([]string{})
		}

		if len(model.SchedulingOptions) > 0 {
			payload.SetSchedulingOptions(*buildPayloadSchedulingOptions(model.SchedulingOptions))
		} else {
			payload.SetSchedulingOptionsNil()
		}

		payload.SetCalculatedFields(buildPayloadCalculatedFields(model.CalculatedFields))
	}

	return &payload, diags
}

func buildUpdateStandardRuleQuery(ctx context.Context, query ruleQueryModel) *datadogV2.SecurityMonitoringRuleQuery {
	payloadQuery := datadogV2.SecurityMonitoringStandardRuleQuery{}

	fwutils.SetOptString(query.Aggregation, func(v string) {
		agg := datadogV2.SecurityMonitoringRuleQueryAggregation(v)
		payloadQuery.SetAggregation(agg)
	})
	fwutils.SetOptStringList(query.GroupByFields, payloadQuery.SetGroupByFields, ctx)
	if !query.HasOptionalGroupByFields.IsNull() && !query.HasOptionalGroupByFields.IsUnknown() {
		payloadQuery.SetHasOptionalGroupByFields(query.HasOptionalGroupByFields.ValueBool())
	}
	fwutils.SetOptStringList(query.DistinctFields, payloadQuery.SetDistinctFields, ctx)
	fwutils.SetOptString(query.DataSource, func(v string) {
		ds := datadogV2.SecurityMonitoringStandardDataSource(v)
		payloadQuery.SetDataSource(ds)
	})
	fwutils.SetOptString(query.Metric, payloadQuery.SetMetric)
	fwutils.SetOptStringList(query.Metrics, payloadQuery.SetMetrics, ctx)
	fwutils.SetOptString(query.Name, payloadQuery.SetName)
	fwutils.SetOptString(query.Query, payloadQuery.SetQuery)

	if !query.Indexes.IsNull() && !query.Indexes.IsUnknown() {
		var indexes []string
		query.Indexes.ElementsAs(ctx, &indexes, false)
		if len(indexes) > 0 {
			payloadQuery.SetIndex(indexes[0])
		}
	}

	standardRuleQuery := datadogV2.SecurityMonitoringStandardRuleQueryAsSecurityMonitoringRuleQuery(&payloadQuery)
	return &standardRuleQuery
}

func buildUpdateSignalRuleQuery(ctx context.Context, query signalQueryModel) (datadogV2.SecurityMonitoringRuleQuery, diag.Diagnostics) {
	var diags diag.Diagnostics
	payloadQuery := datadogV2.SecurityMonitoringSignalRuleQuery{}

	fwutils.SetOptString(query.Aggregation, func(v string) {
		agg := datadogV2.SecurityMonitoringRuleQueryAggregation(v)
		payloadQuery.SetAggregation(agg)
	})
	fwutils.SetOptStringList(query.CorrelatedByFields, payloadQuery.SetCorrelatedByFields, ctx)

	if !query.CorrelatedQueryIndex.IsNull() && !query.CorrelatedQueryIndex.IsUnknown() && query.CorrelatedQueryIndex.ValueString() != "" {
		if vInt, err := strconv.Atoi(query.CorrelatedQueryIndex.ValueString()); err == nil {
			payloadQuery.SetCorrelatedQueryIndex(int32(vInt))
		}
	}

	fwutils.SetOptString(query.Name, payloadQuery.SetName)
	payloadQuery.SetRuleId(query.RuleID.ValueString())

	if !query.DefaultRuleID.IsNull() && !query.DefaultRuleID.IsUnknown() && query.DefaultRuleID.ValueString() != "" {
		diags.AddError("invalid field", "defaultRuleId cannot be set")
	}

	return datadogV2.SecurityMonitoringSignalRuleQueryAsSecurityMonitoringRuleQuery(&payloadQuery), diags
}

func (r *securityMonitoringRuleResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Create is not yet implemented")
}

func (r *securityMonitoringRuleResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Read is not yet implemented")
}

func (r *securityMonitoringRuleResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Update is not yet implemented")
}

func (r *securityMonitoringRuleResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule Delete is not yet implemented")
}

func (r *securityMonitoringRuleResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	response.Diagnostics.AddError("not implemented", "security_monitoring_rule ImportState is not yet implemented")
}
