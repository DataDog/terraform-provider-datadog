package datadog

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDatadogSecurityMonitoringRule() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Security Monitoring Rule API resource. This can be used to create and manage Datadog security monitoring rules. To change settings for a default rule, use `datadog_security_monitoring_default_rule` instead.",
		CreateContext: resourceDatadogSecurityMonitoringRuleCreate,
		ReadContext:   resourceDatadogSecurityMonitoringRuleRead,
		UpdateContext: resourceDatadogSecurityMonitoringRuleUpdate,
		DeleteContext: resourceDatadogSecurityMonitoringRuleDelete,
		CustomizeDiff: customdiff.All(resourceDatadogSecurityMonitoringRuleCustomizeDiff, tagDiff),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return datadogSecurityMonitoringRuleSchema( /* includeValidate= */ true)
		},
	}
}

func datadogSecurityMonitoringRuleSchema(includeValidate bool) map[string]*schema.Schema {
	basicSchema := map[string]*schema.Schema{
		"case": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Cases for generating signals.",
			MaxItems:    10,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the case.",
					},
					"condition": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "A rule case contains logical operations (`>`,`>=`, `&&`, `||`) to determine if a signal should be generated based on the event counts in the previously defined queries.",
					},
					"notifications": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Notification targets for each rule case.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"status": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
						Required:         true,
						Description:      "Severity of the Security Signal.",
					},
					"action": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Action to perform when the case trigger",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"type": {
									Type:             schema.TypeString,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleCaseActionTypeFromValue),
									Required:         true,
									Description:      "Type of action to perform when the case triggers.",
								},
								"options": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Options for the action.",
									MaxItems:    1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"duration": {
												Type:        schema.TypeInt,
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

		"third_party_case": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Cases for generating signals for third-party rules. Only required and accepted for third-party rules",
			MaxItems:    10,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the case.",
					},
					"query": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "A query to associate a third-party event to this case.",
					},
					"notifications": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Notification targets for each rule case.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"status": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
						Required:         true,
						Description:      "Severity of the Security Signal.",
					},
				},
			},
		},

		"enabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "Whether the rule is enabled.",
		},

		"message": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Message for generated signals.",
		},

		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the rule.",
		},

		"has_extended_title": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "Whether the notifications include the triggering group-by values in their title.",
		},

		"options": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Options on rules.",

			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"detection_method": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleDetectionMethodFromValue),
						Optional:         true,
						Description:      "The detection method.",
						Default:          "threshold",
						ForceNew:         true,
					},

					"evaluation_window": {
						Type:             schema.TypeInt,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
						Optional:         true,
						Description:      "A time window is specified to match when at least one of the cases matches true. This is a sliding window and evaluates in real time.",
					},

					"keep_alive": {
						Type:             schema.TypeInt,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleKeepAliveFromValue),
						Optional:         true,
						Description:      "Once a signal is generated, the signal will remain “open” if a case is matched at least once within this keep alive window (in seconds).",
					},

					"max_signal_duration": {
						Type:             schema.TypeInt,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleMaxSignalDurationFromValue),
						Optional:         true,
						Description:      "A signal will “close” regardless of the query being matched once the time exceeds the maximum duration (in seconds). This time is calculated from the first seen timestamp.",
					},

					"new_value_options": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "New value rules specific options.",

						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"learning_method": {
									Type:             schema.TypeString,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningMethodFromValue),
									Optional:         true,
									Default:          "duration",
									Description:      "The learning method used to determine when signals should be generated for values that weren't learned.",
								},
								"learning_duration": {
									Type:             schema.TypeInt,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningDurationFromValue),
									Optional:         true,
									Default:          1,
									Description:      "The duration in days during which values are learned, and after which signals will be generated for values that weren't learned. If set to 0, a signal will be generated for all new values after the first value is learned.",
								},
								"learning_threshold": {
									Type:             schema.TypeInt,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleNewValueOptionsLearningThresholdFromValue),
									Optional:         true,
									Default:          0,
									Description:      "A number of occurrences after which signals are generated for values that weren't learned.",
								},
								"forget_after": {
									Type:             schema.TypeInt,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleNewValueOptionsForgetAfterFromValue),
									Required:         true,
									Description:      "The duration in days after which a learned value is forgotten.",
								},
							},
						},
					},

					"impossible_travel_options": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Options for rules using the impossible travel detection method.",

						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"baseline_user_locations": {
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
									Description: "If true, signals are suppressed for the first 24 hours. During that time, Datadog learns the user's regular access locations. This can be helpful to reduce noise and infer VPN usage or credentialed API access.",
								},
							},
						},
					},

					"anomaly_detection_options": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Options for rules using the anomaly detection method.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"bucket_duration": {
									Type:             schema.TypeInt,
									Optional:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptionsBucketDurationFromValue),
									Description:      "Duration in seconds of the time buckets used to aggregate events matched by the rule. Valid values are 300, 600, 900, 1800, 3600, 10800.",
								},
								"learning_duration": {
									Type:             schema.TypeInt,
									Optional:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptionsLearningDurationFromValue),
									Description:      "Learning duration in hours. Anomaly detection waits for at least this amount of historical data before it starts evaluating. Valid values are 1, 6, 12, 24, 48, 168, 336.",
								},
								"detection_tolerance": {
									Type:             schema.TypeInt,
									Optional:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptionsDetectionToleranceFromValue),
									Description:      "An optional parameter that sets how permissive anomaly detection is. Higher values require higher deviations before triggering a signal. Valid values are 1, 2, 3, 4, 5.",
								},
								"learning_period_baseline": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IntAtLeast(0),
									Description:  "An optional override baseline to apply while the rule is in the learning period. Must be greater than or equal to 0.",
								},
							},
						},
					},

					"third_party_rule_options": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Options for rules using the third-party detection method.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"default_notifications": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Notification targets for the default rule case, when none of the third-party cases match.",
									Elem:        &schema.Schema{Type: schema.TypeString},
								},
								"default_status": {
									Type:             schema.TypeString,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleSeverityFromValue),
									Required:         true,
									Description:      "Severity of the default rule case, when none of the third-party cases match.",
								},
								"signal_title_template": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "A template for the signal title; if omitted, the title is generated based on the case name.",
								},
								"root_query": {
									Type:        schema.TypeList,
									Required:    true,
									MaxItems:    10,
									Description: "Queries to be combined with third-party case queries. Each of them can have different group by fields, to aggregate differently based on the type of alert.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"query": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Query to filter logs.",
											},
											"group_by_fields": {
												Type:        schema.TypeList,
												Optional:    true,
												Description: "Fields to group by. If empty, each log triggers a signal.",
												Elem: &schema.Schema{
													Type:             schema.TypeString,
													ValidateDiagFunc: validators.ValidateNonEmptyStrings,
												},
											},
										},
									},
								},
							},
						},
					},

					"sequence_detection_options": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Options for rules using the sequence detection method.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"steps": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Sequence steps.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"name": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Unique name of the step.",
											},
											"condition": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Condition for the step to match.",
											},
											"evaluation_window": {
												Type:             schema.TypeInt,
												Optional:         true,
												ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
												Description:      "Evaluation window for the step.",
											},
										},
									},
								},
								"step_transitions": {
									Type:        schema.TypeList,
									Optional:    true,
									Description: "Edges of the step graph.",
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"parent": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Parent step name.",
											},
											"child": {
												Type:        schema.TypeString,
												Required:    true,
												Description: "Child step name.",
											},
											"evaluation_window": {
												Type:             schema.TypeInt,
												Optional:         true,
												ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleEvaluationWindowFromValue),
												Description:      "Maximum time allowed to transition from parent to child.",
											},
										},
									},
								},
							},
						},
					},

					"decrease_criticality_based_on_env": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "If true, signals in non-production environments have a lower severity than what is defined by the rule case, which can reduce noise. The decrement is applied when the environment tag of the signal starts with `staging`, `test`, or `dev`. Only available when the rule type is `log_detection`.",
					},
				},
			},
		},

		"query": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Queries for selecting logs which are part of the rule.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"agent_rule": {
						Type:        schema.TypeList,
						Deprecated:  "`agent_rule` has been deprecated in favor of new Agent Rule resource.",
						Optional:    true,
						Description: "**Deprecated**. It won't be applied anymore.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"agent_rule_id": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "**Deprecated**. It won't be applied anymore.",
								},
								"expression": {
									Type:        schema.TypeString,
									Required:    true,
									Description: "**Deprecated**. It won't be applied anymore.",
								},
							},
						},
					},
					"aggregation": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
						Optional:         true,
						Description:      "The aggregation type. For Signal Correlation rules, it must be event_count.",
						Default:          datadogV2.SECURITYMONITORINGRULEQUERYAGGREGATION_COUNT,
					},
					"distinct_fields": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Field for which the cardinality is measured. Sent as an array.",
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						},
					},
					"group_by_fields": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Fields to group by.",
						Elem: &schema.Schema{
							Type:             schema.TypeString,
							ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						},
					},
					"has_optional_group_by_fields": {
						Type:        schema.TypeBool,
						Optional:    true,
						Default:     false,
						Description: "When false, events without a group-by value are ignored by the rule. When true, events with missing group-by fields are processed with `N/A`, replacing the missing values.",
					},
					"data_source": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringStandardDataSourceFromValue),
						Optional:         true,
						Description:      "Source of events.",
						Default:          datadogV2.SECURITYMONITORINGSTANDARDDATASOURCE_LOGS,
					},
					"metric": {
						Type:        schema.TypeString,
						Deprecated:  "Configure `metrics` instead. This attribute will be removed in the next major version of the provider.",
						Optional:    true,
						Description: "The target field to aggregate over when using the `sum`, `max`, or `geo_data` aggregations.",
					},
					"metrics": {
						Type:        schema.TypeList,
						Computed:    true,
						Optional:    true,
						Description: "Group of target fields to aggregate over when using the `sum`, `max`, `geo_data`, or `new_value` aggregations. The `sum`, `max`, and `geo_data` aggregations only accept one value in this list, whereas the `new_value` aggregation accepts up to five values.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the query. Not compatible with `new_value` aggregations.",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query to run on logs.",
					},
					"indexes": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "List of indexes to run the query on when the data source is `logs`. Supports only one element. Used only for scheduled rules (in other words, when `scheduling_options` is defined).",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
				},
			},
		},

		"signal_query": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Queries for selecting logs which are part of the rule.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringRuleQueryAggregationFromValue),
						Optional:         true,
						Description:      "The aggregation type. For Signal Correlation rules, it must be event_count.",
						Default:          datadogV2.SECURITYMONITORINGRULEQUERYAGGREGATION_EVENT_COUNT,
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the query. Not compatible with `new_value` aggregations.",
					},
					"correlated_by_fields": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Fields to correlate by.",
						Elem:        &schema.Schema{Type: schema.TypeString},
					},
					"correlated_query_index": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Index of the rule query used to retrieve the correlated field. An empty string applies correlation on the non-projected per query attributes of the rule.",
						Default:     "",
					},
					"rule_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Rule ID of the signal to correlate.",
					},
					"default_rule_id": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Default Rule ID of the signal to correlate. This value is READ-ONLY.",
					},
				},
			},
		},

		"tags": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Tags for generated signals. Note: if default tags are present at provider level, they will be added to this resource.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},

		"filter": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Additional queries to filter matched events before they are processed. **Note**: This field is deprecated for log detection, signal correlation, and workload security rules.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Query for selecting logs to apply the filtering action.",
					},
					"action": {
						Type:             schema.TypeString,
						Required:         true,
						Description:      "The type of filtering action.",
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringFilterActionFromValue),
					},
				},
			},
		},

		"type": {
			Type: schema.TypeString,
			ValidateDiagFunc: validators.ValidateStringEnumValue(
				datadogV2.SECURITYMONITORINGRULETYPEREAD_APPLICATION_SECURITY, datadogV2.SECURITYMONITORINGRULETYPEREAD_LOG_DETECTION,
				datadogV2.SECURITYMONITORINGRULETYPEREAD_WORKLOAD_SECURITY, datadogV2.SECURITYMONITORINGSIGNALRULETYPE_SIGNAL_CORRELATION),
			Optional:    true,
			Description: "The rule type.",
			Default:     "log_detection",
		},

		"reference_tables": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Reference tables for filtering query results.",

			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"table_name": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						Required:         true,
						Description:      "The name of the reference table.",
					},
					"column_name": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						Required:         true,
						Description:      "The name of the column in the reference table.",
					},
					"log_field_path": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						Required:         true,
						Description:      "The field in the log that should be matched against the reference table.",
					},
					"rule_query_name": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						Required:         true,
						Description:      "The name of the query to filter.",
					},
					"check_presence": {
						Type:        schema.TypeBool,
						Required:    true,
						Description: "Whether to include or exclude logs that match the reference table.",
					},
				},
			},
		},
		"group_signals_by": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Additional grouping to perform on top of the query grouping.",
			Elem:        &schema.Schema{Type: schema.TypeString},
		},

		"calculated_field": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "One or more calculated fields. Available only for scheduled rules (in other words, when `scheduling_options` is defined).",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						Required:         true,
						Description:      "Field name.",
					},
					"expression": {
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateNonEmptyStrings,
						Required:         true,
						Description:      "Expression.",
					},
				},
			},
		},

		"scheduling_options": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Options for scheduled rules. When this field is present, the rule runs based on the schedule. When absent, it runs in real time on ingested logs.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"rrule": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Schedule for the rule queries, written in RRULE syntax. See [RFC](https://icalendar.org/iCalendar-RFC-5545/3-8-5-3-recurrence-rule.html) for syntax reference.",
					},
					"start": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Start date for the schedule, in ISO 8601 format without timezone.",
					},
					"timezone": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Time zone of the start date, in the [tz database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) format.",
					},
				},
			},
		},
	}
	if includeValidate {
		basicSchema["validate"] = &schema.Schema{
			Description: "Whether or not to validate the Rule.",
			Type:        schema.TypeBool,
			Optional:    true,
			DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
				// This is never sent to the backend, so it should never generate a diff
				return true
			},
		}
	}
	return basicSchema
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

func resourceDatadogSecurityMonitoringRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ruleCreate, err := buildCreatePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}
	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().CreateSecurityMonitoringRule(auth, *ruleCreate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if response.SecurityMonitoringStandardRuleResponse != nil {
		d.SetId(response.SecurityMonitoringStandardRuleResponse.GetId())
	} else if response.SecurityMonitoringSignalRuleResponse != nil {
		d.SetId(response.SecurityMonitoringSignalRuleResponse.GetId())
	} else {
		return diag.FromErr(fmt.Errorf("SecurityMonitoringStandardRuleResponse and SecurityMonitoringSignalRuleResponse are both empty"))
	}

	return nil
}

func isSignalCorrelationSchema(d utils.Resource) bool {
	if v, ok := d.GetOk("type"); ok {
		_, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(v.(string))
		return err == nil
	}
	return false
}

func checkQueryConsistency(d utils.Resource) error {
	query := d.Get("query").([]interface{})
	signalQuery := d.Get("signal_query").([]interface{})
	if len(query) > 0 && len(signalQuery) > 0 {
		return fmt.Errorf("query list and signal query list cannot be both populated")
	}
	isSignalCorrelation := isSignalCorrelationSchema(d)
	if !isSignalCorrelation && len(signalQuery) > 0 {
		return fmt.Errorf("signal query list should not be populated for this rule type")
	}
	if isSignalCorrelation && len(query) > 0 {
		return fmt.Errorf("query list should not be populated for this rule type")
	}
	return nil
}

func buildCreatePayload(d utils.Resource) (*datadogV2.SecurityMonitoringRuleCreatePayload, error) {

	if err := checkQueryConsistency(d); err != nil {
		return &datadogV2.SecurityMonitoringRuleCreatePayload{}, err
	}
	if isSignalCorrelationSchema(d) {
		payload, err := buildCreateSignalPayload(d)
		createPayload := datadogV2.SecurityMonitoringSignalRuleCreatePayloadAsSecurityMonitoringRuleCreatePayload(payload)
		return &createPayload, err
	}
	payload, err := buildCreateStandardPayload(d)
	createPayload := datadogV2.SecurityMonitoringStandardRuleCreatePayloadAsSecurityMonitoringRuleCreatePayload(payload)
	return &createPayload, err
}

func buildValidatePayload(d utils.Resource) (*datadogV2.SecurityMonitoringRuleValidatePayload, error) {

	if err := checkQueryConsistency(d); err != nil {
		return &datadogV2.SecurityMonitoringRuleValidatePayload{}, err
	}
	if isSignalCorrelationSchema(d) {
		payload, err := buildSignalPayload(d)
		createPayload := datadogV2.SecurityMonitoringSignalRulePayloadAsSecurityMonitoringRuleValidatePayload(payload)
		return &createPayload, err
	}
	payload, err := buildStandardPayload(d)
	createPayload := datadogV2.SecurityMonitoringStandardRulePayloadAsSecurityMonitoringRuleValidatePayload(payload)
	return &createPayload, err
}

func buildCreateCommonPayload(d utils.Resource, payload securityMonitoringRuleCreateInterface) {
	payload.SetIsEnabled(d.Get("enabled").(bool))
	payload.SetMessage(d.Get("message").(string))
	payload.SetName(d.Get("name").(string))
	payload.SetHasExtendedTitle(d.Get("has_extended_title").(bool))

	if v, ok := d.GetOk("options"); ok {
		tfOptionsList := v.([]interface{})
		payloadOptions := buildPayloadOptions(tfOptionsList, d.Get("type").(string))
		payload.SetOptions(*payloadOptions)
	}

	if v, ok := d.GetOk("tags"); ok {
		tfTags := v.(*schema.Set)
		tags := make([]string, tfTags.Len())
		for i, value := range tfTags.List() {
			tags[i] = value.(string)
		}
		payload.SetTags(tags)
	}

	if v, ok := d.GetOk("filter"); ok {
		tfFilterList := v.([]interface{})
		payload.SetFilters(buildPayloadFilters(tfFilterList))
	}
}

func isThirdPartyRule(d utils.Resource) bool {
	tfOptionsList, ok := d.GetOk("options")

	if !ok {
		return false
	}

	options := extractMapFromInterface(tfOptionsList.([]interface{}))

	if detectionMethod, ok := options["detection_method"]; ok {
		return datadogV2.SecurityMonitoringRuleDetectionMethod(detectionMethod.(string)) == datadogV2.SECURITYMONITORINGRULEDETECTIONMETHOD_THIRD_PARTY
	} else {
		return false
	}
}

func buildCreateStandardPayload(d utils.Resource) (*datadogV2.SecurityMonitoringStandardRuleCreatePayload, error) {
	payload := datadogV2.SecurityMonitoringStandardRuleCreatePayload{}
	buildCreateCommonPayload(d, &payload)

	if isThirdPartyRule(d) {
		payload.SetThirdPartyCases(buildPayloadThirdPartyCases(d))
	} else {
		payload.SetCases(buildCreatePayloadCases(d))
		payload.SetQueries(buildCreateStandardPayloadQueries(d))
	}

	if v, ok := d.GetOk("type"); ok {
		if ruleType, err := datadogV2.NewSecurityMonitoringRuleTypeCreateFromValue(v.(string)); err == nil {
			payload.SetType(*ruleType)
		} else {
			return &payload, err
		}
	}

	if v, ok := d.GetOk("reference_tables"); ok {
		tfReferenceTables := v.([]interface{})
		payload.SetReferenceTables(buildPayloadReferenceTables(tfReferenceTables))
	}

	if v, ok := d.GetOk("group_signals_by"); ok {
		payload.SetGroupSignalsBy(parseStringArray(v.([]interface{})))
	}

	if v, ok := d.GetOk("scheduling_options"); ok {
		tfSchedulingOptionsList := v.([]any)
		schedulingOptions := buildPayloadSchedulingOptions(tfSchedulingOptionsList)
		payload.SetSchedulingOptions(*schedulingOptions)
	}

	if v, ok := d.GetOk("calculated_field"); ok {
		tfCalculatedFields := v.([]any)
		payload.SetCalculatedFields(buildPayloadCalculatedFields(tfCalculatedFields))
	}

	return &payload, nil
}

func buildStandardPayload(d utils.Resource) (*datadogV2.SecurityMonitoringStandardRulePayload, error) {
	payload := datadogV2.SecurityMonitoringStandardRulePayload{}
	buildCreateCommonPayload(d, &payload)

	if isThirdPartyRule(d) {
		payload.SetThirdPartyCases(buildPayloadThirdPartyCases(d))
	} else {
		payload.SetCases(buildCreatePayloadCases(d))
		payload.SetQueries(buildCreateStandardPayloadQueries(d))
	}

	if v, ok := d.GetOk("type"); ok {
		if ruleType, err := datadogV2.NewSecurityMonitoringRuleTypeCreateFromValue(v.(string)); err == nil {
			payload.SetType(*ruleType)
		} else {
			return &payload, err
		}
	}

	if v, ok := d.GetOk("reference_tables"); ok {
		tfReferenceTables := v.([]interface{})
		payload.SetReferenceTables(buildPayloadReferenceTables(tfReferenceTables))
	}

	if v, ok := d.GetOk("group_signals_by"); ok {
		payload.SetGroupSignalsBy(parseStringArray(v.([]interface{})))
	}

	return &payload, nil
}

func buildCreateSignalPayload(d utils.Resource) (*datadogV2.SecurityMonitoringSignalRuleCreatePayload, error) {
	payload := datadogV2.SecurityMonitoringSignalRuleCreatePayload{}
	buildCreateCommonPayload(d, &payload)
	payload.SetCases(buildCreatePayloadCases(d))
	if queries, err := buildCreateSignalPayloadQueries(d); err == nil {
		payload.SetQueries(queries)
	} else {
		return &payload, err
	}

	if v, ok := d.GetOk("type"); ok {
		if ruleType, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(v.(string)); err == nil {
			payload.SetType(*ruleType)
		} else {
			return &payload, err
		}
	}

	return &payload, nil
}

func buildSignalPayload(d utils.Resource) (*datadogV2.SecurityMonitoringSignalRulePayload, error) {
	payload := datadogV2.SecurityMonitoringSignalRulePayload{}
	buildCreateCommonPayload(d, &payload)
	payload.SetCases(buildCreatePayloadCases(d))
	if queries, err := buildCreateSignalPayloadQueries(d); err == nil {
		payload.SetQueries(queries)
	} else {
		return &payload, err
	}

	if v, ok := d.GetOk("type"); ok {
		if ruleType, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(v.(string)); err == nil {
			payload.SetType(*ruleType)
		} else {
			return &payload, err
		}
	}

	return &payload, nil
}

func buildCreatePayloadCases(d utils.Resource) []datadogV2.SecurityMonitoringRuleCaseCreate {
	tfCases := d.Get("case").([]interface{})
	payloadCases := make([]datadogV2.SecurityMonitoringRuleCaseCreate, len(tfCases))

	for idx, ruleCaseIf := range tfCases {
		ruleCase := ruleCaseIf.(map[string]interface{})
		status := datadogV2.SecurityMonitoringRuleSeverity(ruleCase["status"].(string))
		structRuleCase := datadogV2.NewSecurityMonitoringRuleCaseCreate(status)
		if v, ok := ruleCase["name"]; ok {
			name := v.(string)
			structRuleCase.SetName(name)
		}
		if v, ok := ruleCase["condition"]; ok {
			condition := v.(string)
			structRuleCase.SetCondition(condition)
		}
		if v, ok := ruleCase["notifications"]; ok {
			structRuleCase.SetNotifications(parseStringArray(v.([]interface{})))
		}
		if action, ok := ruleCase["action"]; ok && len(action.([]interface{})) > 0 {
			structRuleCase.SetActions(buildPayloadCaseActions(action.([]interface{})))
		}
		payloadCases[idx] = *structRuleCase
	}
	return payloadCases
}

func buildPayloadThirdPartyCase(tfThirdPartyCase map[string]interface{}) *datadogV2.SecurityMonitoringThirdPartyRuleCaseCreate {
	status := datadogV2.SecurityMonitoringRuleSeverity(tfThirdPartyCase["status"].(string))
	thirdPartyCase := datadogV2.NewSecurityMonitoringThirdPartyRuleCaseCreate(status)

	if v, ok := tfThirdPartyCase["query"]; ok {
		thirdPartyCase.SetQuery(v.(string))
	}

	if v, ok := tfThirdPartyCase["name"]; ok {
		thirdPartyCase.SetName(v.(string))
	}

	if v, ok := tfThirdPartyCase["notifications"]; ok {
		thirdPartyCase.SetNotifications(parseStringArray(v.([]interface{})))
	}

	return thirdPartyCase
}

func buildPayloadThirdPartyCases(d utils.Resource) []datadogV2.SecurityMonitoringThirdPartyRuleCaseCreate {
	tfThirdPartyCases := d.Get("third_party_case").([]interface{})
	payloadThirdPartyCases := make([]datadogV2.SecurityMonitoringThirdPartyRuleCaseCreate, len(tfThirdPartyCases))

	for idx, tfThirdPartyCase := range tfThirdPartyCases {
		payloadThirdPartyCases[idx] = *buildPayloadThirdPartyCase(tfThirdPartyCase.(map[string]interface{}))
	}

	return payloadThirdPartyCases
}

func buildPayloadCalculatedFields(tfCalculatedFields []any) []datadogV2.CalculatedField {
	calculatedFields := make([]datadogV2.CalculatedField, len(tfCalculatedFields))

	for idx, tfCalculatedFieldUntyped := range tfCalculatedFields {
		tfCalculatedField := tfCalculatedFieldUntyped.(map[string]any)

		calculatedFields[idx] = datadogV2.CalculatedField{
			Name:       tfCalculatedField["name"].(string),
			Expression: tfCalculatedField["expression"].(string),
		}
	}

	return calculatedFields
}

func buildPayloadSchedulingOptions(tfSchedulingOptionsList []any) *datadogV2.SecurityMonitoringSchedulingOptions {
	tfSchedulingOptions := extractMapFromInterface(tfSchedulingOptionsList)
	schedulingOptions := datadogV2.NewSecurityMonitoringSchedulingOptions()

	schedulingOptions.SetRrule(tfSchedulingOptions["rrule"].(string))
	schedulingOptions.SetStart(tfSchedulingOptions["start"].(string))
	schedulingOptions.SetTimezone(tfSchedulingOptions["timezone"].(string))

	return schedulingOptions
}

func buildPayloadOptions(tfOptionsList []interface{}, ruleType string) *datadogV2.SecurityMonitoringRuleOptions {
	payloadOptions := datadogV2.NewSecurityMonitoringRuleOptions()
	tfOptions := extractMapFromInterface(tfOptionsList)

	if v, ok := tfOptions["detection_method"]; ok {
		detectionMethod := datadogV2.SecurityMonitoringRuleDetectionMethod(v.(string))
		payloadOptions.DetectionMethod = &detectionMethod
	}
	if v, ok := tfOptions["evaluation_window"]; ok {
		evaluationWindow := datadogV2.SecurityMonitoringRuleEvaluationWindow(v.(int))
		payloadOptions.EvaluationWindow = &evaluationWindow
	}
	if v, ok := tfOptions["keep_alive"]; ok {
		keepAlive := datadogV2.SecurityMonitoringRuleKeepAlive(v.(int))
		payloadOptions.KeepAlive = &keepAlive
	}
	if v, ok := tfOptions["max_signal_duration"]; ok {
		maxSignalDuration := datadogV2.SecurityMonitoringRuleMaxSignalDuration(v.(int))
		payloadOptions.MaxSignalDuration = &maxSignalDuration
	}
	if v, ok := tfOptions["decrease_criticality_based_on_env"]; ok && ruleType == string(datadogV2.SECURITYMONITORINGRULETYPECREATE_LOG_DETECTION) {
		payloadOptions.SetDecreaseCriticalityBasedOnEnv(v.(bool))
	}

	if v, ok := tfOptions["new_value_options"]; ok {
		tfNewValueOptionsList := v.([]interface{})
		if payloadNewValueOptions, ok := buildPayloadNewValueOptions(tfNewValueOptionsList); ok {
			payloadOptions.NewValueOptions = payloadNewValueOptions
		}
	}

	if v, ok := tfOptions["impossible_travel_options"]; ok {
		tfImpossibleTravelOptionsList := v.([]interface{})
		if payloadImpossibleTravelOptions, ok := buildPayloadImpossibleTravelOptions(tfImpossibleTravelOptionsList); ok {
			payloadOptions.ImpossibleTravelOptions = payloadImpossibleTravelOptions
		}
	}

	if v, ok := tfOptions["anomaly_detection_options"]; ok {
		tfAnomalyDetectionOptionsList := v.([]interface{})
		if payloadAnomalyDetectionOptions, ok := buildPayloadAnomalyDetectionOptions(tfAnomalyDetectionOptionsList); ok {
			payloadOptions.AnomalyDetectionOptions = payloadAnomalyDetectionOptions
		}
	}

	if v, ok := tfOptions["third_party_rule_options"]; ok {
		tfThirdPartyOptionsList := v.([]interface{})
		if payloadThirdPartyRuleOptions, ok := buildPayloadThirdPartyRuleOptions(tfThirdPartyOptionsList); ok {
			payloadOptions.ThirdPartyRuleOptions = payloadThirdPartyRuleOptions
		}
	}

	if v, ok := tfOptions["sequence_detection_options"]; ok {
		tfSequenceDetectionOptionsList := v.([]interface{})
		if payloadSequenceDetectionOptions, ok := buildPayloadSequenceDetectionOptions(tfSequenceDetectionOptionsList); ok {
			payloadOptions.SequenceDetectionOptions = payloadSequenceDetectionOptions
		}
	}

	return payloadOptions
}

func buildPayloadImpossibleTravelOptions(tfOptionsList []interface{}) (*datadogV2.SecurityMonitoringRuleImpossibleTravelOptions, bool) {
	options := datadogV2.NewSecurityMonitoringRuleImpossibleTravelOptions()
	tfOptions := extractMapFromInterface(tfOptionsList)

	hasPayload := false

	if v, ok := tfOptions["baseline_user_locations"]; ok {
		hasPayload = true
		shouldBaselineUserLocations := v.(bool)
		options.BaselineUserLocations = &shouldBaselineUserLocations
	}

	return options, hasPayload
}

func buildPayloadAnomalyDetectionOptions(tfOptionsList []interface{}) (*datadogV2.SecurityMonitoringRuleAnomalyDetectionOptions, bool) {
	options := datadogV2.NewSecurityMonitoringRuleAnomalyDetectionOptions()
	tfOptions := extractMapFromInterface(tfOptionsList)

	hasPayload := false

	if v, ok := tfOptions["bucket_duration"]; ok {
		hasPayload = true
		bucketDuration := datadogV2.SecurityMonitoringRuleAnomalyDetectionOptionsBucketDuration(v.(int))
		options.BucketDuration = &bucketDuration
	}

	if v, ok := tfOptions["learning_duration"]; ok {
		hasPayload = true
		learningDuration := datadogV2.SecurityMonitoringRuleAnomalyDetectionOptionsLearningDuration(v.(int))
		options.LearningDuration = &learningDuration
	}

	if v, ok := tfOptions["detection_tolerance"]; ok {
		hasPayload = true
		detectionTolerance := datadogV2.SecurityMonitoringRuleAnomalyDetectionOptionsDetectionTolerance(v.(int))
		options.DetectionTolerance = &detectionTolerance
	}

	if v, ok := tfOptions["learning_period_baseline"]; ok {
		hasPayload = true
		learningPeriodBaseline := int64(v.(int))
		options.LearningPeriodBaseline = &learningPeriodBaseline
	}

	return options, hasPayload
}

func buildPayloadNewValueOptions(tfOptionsList []interface{}) (*datadogV2.SecurityMonitoringRuleNewValueOptions, bool) {
	payloadNewValueRulesOptions := datadogV2.NewSecurityMonitoringRuleNewValueOptions()
	tfOptions := extractMapFromInterface(tfOptionsList)
	hasPayload := false
	if v, ok := tfOptions["learning_method"]; ok {
		hasPayload = true
		learningMethod := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningMethod(v.(string))
		payloadNewValueRulesOptions.LearningMethod = &learningMethod
	}
	if v, ok := tfOptions["learning_duration"]; ok {
		hasPayload = true
		learningDuration := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningDuration(v.(int))
		payloadNewValueRulesOptions.LearningDuration = &learningDuration
	}
	if v, ok := tfOptions["learning_threshold"]; ok {
		hasPayload = true
		learningThreshold := datadogV2.SecurityMonitoringRuleNewValueOptionsLearningThreshold(v.(int))
		payloadNewValueRulesOptions.LearningThreshold = &learningThreshold
	}
	if v, ok := tfOptions["forget_after"]; ok {
		hasPayload = true
		forgetAfter := datadogV2.SecurityMonitoringRuleNewValueOptionsForgetAfter(v.(int))
		payloadNewValueRulesOptions.ForgetAfter = &forgetAfter
	}
	return payloadNewValueRulesOptions, hasPayload
}

func buildPayloadThirdPartyRuleOptions(tfOptionsList []interface{}) (*datadogV2.SecurityMonitoringRuleThirdPartyOptions, bool) {
	payload := datadogV2.NewSecurityMonitoringRuleThirdPartyOptions()
	hasPayload := false

	tfOptions := extractMapFromInterface(tfOptionsList)

	if v, ok := tfOptions["default_status"]; ok {
		hasPayload = true
		payload.SetDefaultStatus(datadogV2.SecurityMonitoringRuleSeverity(v.(string)))
	}

	if v, ok := tfOptions["default_notifications"]; ok {
		tfNotifications := v.([]interface{})

		if len(tfNotifications) > 0 {
			hasPayload = true
		}

		payload.SetDefaultNotifications(parseStringArray(tfNotifications))
	}

	if v, ok := tfOptions["signal_title_template"]; ok {
		hasPayload = true
		payload.SetSignalTitleTemplate(v.(string))
	}

	if v, ok := tfOptions["root_query"]; ok {
		tfRootQueries := v.([]interface{})

		if len(tfRootQueries) > 0 {
			hasPayload = true
		}

		payloadRootQueries := make([]datadogV2.SecurityMonitoringThirdPartyRootQuery, len(tfRootQueries))

		for idx, tfQuery := range tfRootQueries {
			payloadRootQueries[idx] = *buildRootQueryPayload(tfQuery.(map[string]interface{}))
		}

		payload.SetRootQueries(payloadRootQueries)
	}

	return payload, hasPayload
}

func buildPayloadSequenceDetectionOptions(tfOptionsList []interface{}) (*datadogV2.SecurityMonitoringRuleSequenceDetectionOptions, bool) {
	options := datadogV2.NewSecurityMonitoringRuleSequenceDetectionOptions()
	hasPayload := false

	tfOptions := extractMapFromInterface(tfOptionsList)

	if v, ok := tfOptions["steps"]; ok {
		tfSteps := v.([]interface{})
		if len(tfSteps) > 0 {
			hasPayload = true
		}
		payloadSteps := make([]datadogV2.SecurityMonitoringRuleSequenceDetectionStep, len(tfSteps))
		for idx, stepIf := range tfSteps {
			stepMap := stepIf.(map[string]interface{})
			step := datadogV2.SecurityMonitoringRuleSequenceDetectionStep{}
			if v, ok := stepMap["name"]; ok {
				step.SetName(v.(string))
			}
			if v, ok := stepMap["condition"]; ok {
				step.SetCondition(v.(string))
			}
			if v, ok := stepMap["evaluation_window"]; ok {
				ew := datadogV2.SecurityMonitoringRuleEvaluationWindow(v.(int))
				step.SetEvaluationWindow(ew)
			}
			payloadSteps[idx] = step
		}
		options.SetSteps(payloadSteps)
	}

	if v, ok := tfOptions["step_transitions"]; ok {
		tfTransitions := v.([]interface{})
		if len(tfTransitions) > 0 {
			hasPayload = true
		}
		payloadTransitions := make([]datadogV2.SecurityMonitoringRuleSequenceDetectionStepTransition, len(tfTransitions))
		for idx, trIf := range tfTransitions {
			trMap := trIf.(map[string]interface{})
			transition := datadogV2.SecurityMonitoringRuleSequenceDetectionStepTransition{}
			if v, ok := trMap["parent"]; ok {
				transition.SetParent(v.(string))
			}
			if v, ok := trMap["child"]; ok {
				transition.SetChild(v.(string))
			}
			if v, ok := trMap["evaluation_window"]; ok {
				ew := datadogV2.SecurityMonitoringRuleEvaluationWindow(v.(int))
				transition.SetEvaluationWindow(ew)
			}
			payloadTransitions[idx] = transition
		}
		options.SetStepTransitions(payloadTransitions)
	}

	return options, hasPayload
}

func buildRootQueryPayload(rootQuery map[string]interface{}) *datadogV2.SecurityMonitoringThirdPartyRootQuery {
	payloadRootQuery := datadogV2.NewSecurityMonitoringThirdPartyRootQuery()

	if v, ok := rootQuery["query"]; ok {
		payloadRootQuery.SetQuery(v.(string))
	}

	if v, ok := rootQuery["group_by_fields"]; ok {
		payloadRootQuery.SetGroupByFields(parseStringArray(v.([]interface{})))
	}

	return payloadRootQuery
}

func parseStringArray(array []interface{}) []string {
	parsed := make([]string, len(array))

	for idx, value := range array {
		parsed[idx] = value.(string)
	}

	return parsed
}

func extractMapFromInterface(tfOptionsList []interface{}) map[string]interface{} {
	var tfOptions map[string]interface{}
	if len(tfOptionsList) == 0 || tfOptionsList[0] == nil {
		tfOptions = make(map[string]interface{})
	} else {
		tfOptions = tfOptionsList[0].(map[string]interface{})
	}
	return tfOptions
}

func buildCreateStandardPayloadQueries(d utils.Resource) []datadogV2.SecurityMonitoringStandardRuleQuery {
	tfQueries := d.Get("query").([]interface{})
	payloadQueries := make([]datadogV2.SecurityMonitoringStandardRuleQuery, len(tfQueries))
	for idx, tfQuery := range tfQueries {
		query := tfQuery.(map[string]interface{})
		payloadQuery := datadogV2.SecurityMonitoringStandardRuleQuery{}

		if v, ok := query["aggregation"]; ok {
			aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
			payloadQuery.SetAggregation(aggregation)
		}

		if v, ok := query["group_by_fields"]; ok {
			payloadQuery.SetGroupByFields(parseStringArray(v.([]interface{})))
		}

		if v, ok := query["has_optional_group_by_fields"]; ok {
			payloadQuery.SetHasOptionalGroupByFields(v.(bool))
		}

		if v, ok := query["distinct_fields"]; ok {
			payloadQuery.SetDistinctFields(parseStringArray(v.([]interface{})))
		}

		if v, ok := query["data_source"]; ok {
			dataSource := datadogV2.SecurityMonitoringStandardDataSource(v.(string))
			payloadQuery.SetDataSource(dataSource)
		}

		if v, ok := query["metric"]; ok {
			payloadQuery.SetMetric(v.(string))
		}

		if v, ok := query["metrics"]; ok && v != nil {
			payloadQuery.SetMetrics(parseStringArray(v.([]interface{})))
		}

		if v, ok := query["name"]; ok {
			name := v.(string)
			payloadQuery.SetName(name)
		}

		if v, ok := query["indexes"]; ok {
			if indexes := parseStringArray(v.([]any)); len(indexes) > 0 {
				payloadQuery.SetIndex(indexes[0])
			}
		}

		payloadQuery.SetQuery(query["query"].(string))

		payloadQueries[idx] = payloadQuery
	}
	return payloadQueries
}

func buildCreateSignalPayloadQueries(d utils.Resource) ([]datadogV2.SecurityMonitoringSignalRuleQuery, error) {
	tfQueries := d.Get("signal_query").([]interface{})
	payloadQueries := make([]datadogV2.SecurityMonitoringSignalRuleQuery, len(tfQueries))
	for idx, tfQuery := range tfQueries {
		query := tfQuery.(map[string]interface{})
		payloadQuery := datadogV2.SecurityMonitoringSignalRuleQuery{}

		if v, ok := query["aggregation"]; ok {
			aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
			payloadQuery.SetAggregation(aggregation)
		}

		if v, ok := query["correlated_by_fields"]; ok {
			payloadQuery.SetCorrelatedByFields(parseStringArray(v.([]interface{})))
		}

		if v, ok := query["correlated_query_index"]; ok && len(v.(string)) > 0 {
			if vInt, err := strconv.Atoi(v.(string)); err == nil {
				payloadQuery.SetCorrelatedQueryIndex(int32(vInt))
			}
		}

		if v, ok := query["name"]; ok {
			name := v.(string)
			payloadQuery.SetName(name)
		}

		payloadQuery.SetRuleId(query["rule_id"].(string))

		if v, ok := query["default_rule_id"].(string); ok && v != "" {
			return payloadQueries, fmt.Errorf("default_rule_id cannot be set")
		}

		payloadQueries[idx] = payloadQuery
	}
	return payloadQueries, nil
}

func buildPayloadFilters(tfFilters []interface{}) []datadogV2.SecurityMonitoringFilter {
	payloadFilters := make([]datadogV2.SecurityMonitoringFilter, len(tfFilters))
	for idx, tfFilter := range tfFilters {
		filter := tfFilter.(map[string]interface{})
		payloadFilter := datadogV2.SecurityMonitoringFilter{}

		action := datadogV2.SecurityMonitoringFilterAction(filter["action"].(string))
		payloadFilter.SetAction(action)

		payloadFilter.SetQuery(filter["query"].(string))

		payloadFilters[idx] = payloadFilter
	}
	return payloadFilters
}

func buildPayloadReferenceTables(tfReferenceTables []interface{}) []datadogV2.SecurityMonitoringReferenceTable {
	payloadReferenceTables := make([]datadogV2.SecurityMonitoringReferenceTable, len(tfReferenceTables))
	for idx, tfReferenceTable := range tfReferenceTables {
		referenceTable := tfReferenceTable.(map[string]interface{})
		payloadReferenceTable := datadogV2.SecurityMonitoringReferenceTable{}

		payloadReferenceTable.SetTableName(referenceTable["table_name"].(string))
		payloadReferenceTable.SetColumnName(referenceTable["column_name"].(string))
		payloadReferenceTable.SetLogFieldPath(referenceTable["log_field_path"].(string))
		payloadReferenceTable.SetRuleQueryName(referenceTable["rule_query_name"].(string))
		payloadReferenceTable.SetCheckPresence(referenceTable["check_presence"].(bool))

		payloadReferenceTables[idx] = payloadReferenceTable
	}
	return payloadReferenceTables
}

func buildPayloadCaseActions(tfActions []any) []datadogV2.SecurityMonitoringRuleCaseAction {
	payloadActions := make([]datadogV2.SecurityMonitoringRuleCaseAction, len(tfActions))
	for actionIdx, actionIf := range tfActions {
		action := actionIf.(map[string]any)
		actionType := datadogV2.SecurityMonitoringRuleCaseActionType(action["type"].(string))
		payloadOptions := datadogV2.NewSecurityMonitoringRuleCaseActionOptions()
		if tfOptionsList, ok := action["options"]; ok {
			tfOptions := extractMapFromInterface(tfOptionsList.([]any))
			for k, v := range tfOptions {
				if k == "duration" {
					payloadOptions.SetDuration(int64(v.(int)))
				}
			}
		}
		payloadActions[actionIdx] = datadogV2.SecurityMonitoringRuleCaseAction{
			Type:    &actionType,
			Options: payloadOptions,
		}
	}
	return payloadActions
}

func resourceDatadogSecurityMonitoringRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()
	ruleResponse, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().GetSecurityMonitoringRule(auth, id)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}
	if err := utils.CheckForUnparsed(ruleResponse); err != nil {
		return diag.FromErr(err)
	}

	if ruleResponse.SecurityMonitoringStandardRuleResponse != nil {
		updateStandardResourceDataFromResponse(d, ruleResponse.SecurityMonitoringStandardRuleResponse)
	} else if ruleResponse.SecurityMonitoringSignalRuleResponse != nil {
		updateSignalResourceDataFromResponse(d, ruleResponse.SecurityMonitoringSignalRuleResponse)
	}
	return nil
}

func updateCommonResourceDataFromResponse(d *schema.ResourceData, ruleResponse securityMonitoringRuleResponseInterface) {
	d.Set("message", ruleResponse.GetMessage())
	d.Set("name", ruleResponse.GetName())
	d.Set("has_extended_title", ruleResponse.GetHasExtendedTitle())
	d.Set("enabled", ruleResponse.GetIsEnabled())

	options := extractTfOptions(ruleResponse.GetOptions())

	d.Set("options", []map[string]interface{}{options})

	if _, ok := ruleResponse.GetFiltersOk(); ok {
		filters := extractFiltersFromRuleResponse(ruleResponse.GetFilters())
		d.Set("filter", filters)
	}

	if tags, ok := ruleResponse.GetTagsOk(); ok {
		d.Set("tags", *tags)
	}
}

func extractThirdPartyCases(responseThirdPartyCases []datadogV2.SecurityMonitoringThirdPartyRuleCase) []map[string]interface{} {
	tfThirdPartyCases := make([]map[string]interface{}, len(responseThirdPartyCases))

	for idx, thirdPartyCase := range responseThirdPartyCases {
		tfThirdPartyCase := make(map[string]interface{})

		if status, ok := thirdPartyCase.GetStatusOk(); ok {
			tfThirdPartyCase["status"] = *status
		}

		if name, ok := thirdPartyCase.GetNameOk(); ok {
			tfThirdPartyCase["name"] = *name
		}

		if notifications, ok := thirdPartyCase.GetNotificationsOk(); ok {
			tfThirdPartyCase["notifications"] = *notifications
		}

		if query, ok := thirdPartyCase.GetQueryOk(); ok {
			tfThirdPartyCase["query"] = *query
		}

		tfThirdPartyCases[idx] = tfThirdPartyCase
	}

	return tfThirdPartyCases
}

func updateStandardResourceDataFromResponse(d *schema.ResourceData, ruleResponse *datadogV2.SecurityMonitoringStandardRuleResponse) {
	updateCommonResourceDataFromResponse(d, ruleResponse)
	if options, ok := ruleResponse.GetOptionsOk(); ok && options.GetDetectionMethod() == datadogV2.SECURITYMONITORINGRULEDETECTIONMETHOD_THIRD_PARTY {
		d.Set("third_party_case", extractThirdPartyCases(ruleResponse.GetThirdPartyCases()))
	} else {
		d.Set("case", extractRuleCases(ruleResponse.GetCases()))
		d.Set("query", extractStandardRuleQueries(ruleResponse.GetQueries()))
	}

	if ruleType, ok := ruleResponse.GetTypeOk(); ok {
		d.Set("type", *ruleType)
	}

	if referenceTables, ok := ruleResponse.GetReferenceTablesOk(); ok {
		refTables := extractReferenceTables(*referenceTables)
		d.Set("reference_tables", refTables)
	}
	if groupSignalsBy, ok := ruleResponse.GetGroupSignalsByOk(); ok {
		d.Set("group_signals_by", groupSignalsBy)
	}

	if calculatedFields, ok := ruleResponse.GetCalculatedFieldsOk(); ok {
		d.Set("calculated_field", extractCalculatedFields(*calculatedFields))
	}

	if schedulingOptions, ok := ruleResponse.GetSchedulingOptionsOk(); ok {
		d.Set("scheduling_options", []any{extractSchedulingOptions(schedulingOptions)})
	}
}

func extractStandardRuleQueries(responseRuleQueries []datadogV2.SecurityMonitoringStandardRuleQuery) []map[string]interface{} {
	ruleQueries := make([]map[string]interface{}, len(responseRuleQueries))

	for idx, responseRuleQuery := range responseRuleQueries {
		ruleQuery := make(map[string]interface{})

		if aggregation, ok := responseRuleQuery.GetAggregationOk(); ok {
			ruleQuery["aggregation"] = *aggregation
		}
		if distinctFields, ok := responseRuleQuery.GetDistinctFieldsOk(); ok {
			ruleQuery["distinct_fields"] = *distinctFields
		}
		if groupByFields, ok := responseRuleQuery.GetGroupByFieldsOk(); ok {
			ruleQuery["group_by_fields"] = *groupByFields
		}
		if hasGbf, ok := responseRuleQuery.GetHasOptionalGroupByFieldsOk(); ok {
			ruleQuery["has_optional_group_by_fields"] = *hasGbf
		}
		if dataSource, ok := responseRuleQuery.GetDataSourceOk(); ok {
			ruleQuery["data_source"] = *dataSource
		}
		if metric, ok := responseRuleQuery.GetMetricOk(); ok {
			ruleQuery["metric"] = *metric
		}
		if metrics, ok := responseRuleQuery.GetMetricsOk(); ok {
			ruleQuery["metrics"] = *metrics
		}
		if name, ok := responseRuleQuery.GetNameOk(); ok {
			ruleQuery["name"] = *name
		}
		if query, ok := responseRuleQuery.GetQueryOk(); ok {
			ruleQuery["query"] = *query
		}
		if index, ok := responseRuleQuery.GetIndexOk(); ok {
			ruleQuery["indexes"] = []string{*index}
		}

		ruleQueries[idx] = ruleQuery
	}

	return ruleQueries
}

func updateSignalResourceDataFromResponse(d *schema.ResourceData, ruleResponse *datadogV2.SecurityMonitoringSignalRuleResponse) {
	updateCommonResourceDataFromResponse(d, ruleResponse)

	d.Set("case", extractRuleCases(ruleResponse.GetCases()))

	ruleQueries := make([]map[string]interface{}, len(ruleResponse.GetQueries()))
	for idx := range ruleResponse.GetQueries() {
		ruleQuery := make(map[string]interface{})
		responseRuleQuery := ruleResponse.GetQueries()[idx]

		if aggregation, ok := responseRuleQuery.GetAggregationOk(); ok {
			ruleQuery["aggregation"] = *aggregation
		}
		if correlatedByFields, ok := responseRuleQuery.GetCorrelatedByFieldsOk(); ok {
			ruleQuery["correlated_by_fields"] = *correlatedByFields
		}
		if correlatedQueryIndex, ok := responseRuleQuery.GetCorrelatedQueryIndexOk(); ok {
			ruleQuery["correlated_query_index"] = fmt.Sprintf("%d", *correlatedQueryIndex)
		}
		if name, ok := responseRuleQuery.GetNameOk(); ok {
			ruleQuery["name"] = *name
		}
		if ruleId, ok := responseRuleQuery.GetRuleIdOk(); ok {
			ruleQuery["rule_id"] = *ruleId
		}
		if defaultRuleId, ok := responseRuleQuery.GetDefaultRuleIdOk(); ok {
			ruleQuery["default_rule_id"] = *defaultRuleId
		}

		ruleQueries[idx] = ruleQuery
	}
	d.Set("signal_query", ruleQueries)

	if ruleType, ok := ruleResponse.GetTypeOk(); ok {
		d.Set("type", *ruleType)
	}

	if tags, ok := ruleResponse.GetTagsOk(); ok {
		d.Set("tags", *tags)
	}
}

func extractFiltersFromRuleResponse(ruleResponseFilter []datadogV2.SecurityMonitoringFilter) []interface{} {

	filters := make([]interface{}, len(ruleResponseFilter))
	for idx, responseFilter := range ruleResponseFilter {
		filter := make(map[string]interface{})
		if query, ok := responseFilter.GetQueryOk(); ok {
			filter["query"] = *query
		}
		if action, ok := responseFilter.GetActionOk(); ok {
			filter["action"] = *action
		}

		filters[idx] = filter
	}
	return filters
}

func extractRuleCases(responseRulesCases []datadogV2.SecurityMonitoringRuleCase) []map[string]interface{} {
	ruleCases := make([]map[string]interface{}, len(responseRulesCases))
	for idx, responseRuleCase := range responseRulesCases {
		ruleCase := make(map[string]interface{})

		if name, ok := responseRuleCase.GetNameOk(); ok {
			ruleCase["name"] = *name
		}
		if condition, ok := responseRuleCase.GetConditionOk(); ok {
			ruleCase["condition"] = *condition
		}
		if notifications, ok := responseRuleCase.GetNotificationsOk(); ok {
			ruleCase["notifications"] = *notifications
		}
		ruleCase["status"] = responseRuleCase.GetStatus()
		if actions, ok := responseRuleCase.GetActionsOk(); ok {
			tfActions := make([]map[string]interface{}, len(*actions))
			for idx, action := range *actions {
				tfAction := make(map[string]interface{})
				tfAction["type"] = action.GetType()
				if options, ok := action.GetOptionsOk(); ok {
					tfOptions := make(map[string]interface{})
					if duration, ok := options.GetDurationOk(); ok {
						tfOptions["duration"] = duration
					}
					if len(tfOptions) > 0 {
						tfAction["options"] = []any{tfOptions}
					}
				}
				tfActions[idx] = tfAction
			}
			ruleCase["action"] = tfActions
		}

		ruleCases[idx] = ruleCase
	}
	return ruleCases
}

func extractTfOptions(options datadogV2.SecurityMonitoringRuleOptions) map[string]interface{} {
	tfOptions := make(map[string]interface{})
	if evaluationWindow, ok := options.GetEvaluationWindowOk(); ok {
		tfOptions["evaluation_window"] = *evaluationWindow
	}
	if keepAlive, ok := options.GetKeepAliveOk(); ok {
		tfOptions["keep_alive"] = *keepAlive
	}
	if maxSignalDuration, ok := options.GetMaxSignalDurationOk(); ok {
		tfOptions["max_signal_duration"] = *maxSignalDuration
	}
	if decreaseCriticalityBasedOnEnv, ok := options.GetDecreaseCriticalityBasedOnEnvOk(); ok {
		tfOptions["decrease_criticality_based_on_env"] = *decreaseCriticalityBasedOnEnv
	}
	if detectionMethod, ok := options.GetDetectionMethodOk(); ok {
		tfOptions["detection_method"] = *detectionMethod
	}
	if newValueOptions, ok := options.GetNewValueOptionsOk(); ok {
		tfNewValueOptions := make(map[string]interface{})
		tfNewValueOptions["forget_after"] = int(newValueOptions.GetForgetAfter())
		tfNewValueOptions["learning_method"] = string(newValueOptions.GetLearningMethod())
		tfNewValueOptions["learning_duration"] = int(newValueOptions.GetLearningDuration())
		tfNewValueOptions["learning_threshold"] = int(newValueOptions.GetLearningThreshold())
		tfOptions["new_value_options"] = []map[string]interface{}{tfNewValueOptions}
	}
	if impossibleTravelOptions, ok := options.GetImpossibleTravelOptionsOk(); ok {
		tfImpossibleTravelOptions := make(map[string]interface{})
		tfImpossibleTravelOptions["baseline_user_locations"] = impossibleTravelOptions.GetBaselineUserLocations()
		tfOptions["impossible_travel_options"] = []map[string]interface{}{tfImpossibleTravelOptions}
	}
	if anomalyDetectionOptions, ok := options.GetAnomalyDetectionOptionsOk(); ok {
		tfAnomalyDetectionOptions := make(map[string]interface{})
		tfAnomalyDetectionOptions["bucket_duration"] = int(anomalyDetectionOptions.GetBucketDuration())
		tfAnomalyDetectionOptions["learning_duration"] = int(anomalyDetectionOptions.GetLearningDuration())
		tfAnomalyDetectionOptions["detection_tolerance"] = int(anomalyDetectionOptions.GetDetectionTolerance())
		if learningPeriodBaseline, ok := anomalyDetectionOptions.GetLearningPeriodBaselineOk(); ok {
			tfAnomalyDetectionOptions["learning_period_baseline"] = *learningPeriodBaseline
		}
		tfOptions["anomaly_detection_options"] = []map[string]interface{}{tfAnomalyDetectionOptions}
	}
	if thirdPartyOptions, ok := options.GetThirdPartyRuleOptionsOk(); ok {
		tfThirdPartyOptions := make(map[string]interface{})
		tfThirdPartyOptions["default_status"] = thirdPartyOptions.GetDefaultStatus()

		if defaultNotifications, ok := thirdPartyOptions.GetDefaultNotificationsOk(); ok {
			tfThirdPartyOptions["default_notifications"] = *defaultNotifications
		}

		if signalTitleTemplate, ok := thirdPartyOptions.GetSignalTitleTemplateOk(); ok {
			tfThirdPartyOptions["signal_title_template"] = signalTitleTemplate
		}

		tfRootQueries := make([]map[string]interface{}, len(thirdPartyOptions.GetRootQueries()))

		for idx, rootQuery := range thirdPartyOptions.GetRootQueries() {
			tfRootQueries[idx] = map[string]interface{}{
				"query":           rootQuery.GetQuery(),
				"group_by_fields": rootQuery.GetGroupByFields(),
			}
		}

		tfThirdPartyOptions["root_query"] = tfRootQueries

		tfOptions["third_party_rule_options"] = []map[string]interface{}{tfThirdPartyOptions}
	}
	if seqOptions, ok := options.GetSequenceDetectionOptionsOk(); ok {
		tfSeqOptions := make(map[string]interface{})
		steps := seqOptions.GetSteps()
		tfSteps := make([]map[string]interface{}, len(steps))
		for idx, step := range steps {
			stepMap := make(map[string]interface{})
			if name, ok := step.GetNameOk(); ok {
				stepMap["name"] = *name
			}
			if cond, ok := step.GetConditionOk(); ok {
				stepMap["condition"] = *cond
			}
			if ew, ok := step.GetEvaluationWindowOk(); ok {
				stepMap["evaluation_window"] = *ew
			}
			tfSteps[idx] = stepMap
		}
		if len(tfSteps) > 0 {
			tfSeqOptions["steps"] = tfSteps
		}
		transitions := seqOptions.GetStepTransitions()
		tfTransitions := make([]map[string]interface{}, len(transitions))
		for idx, tr := range transitions {
			trMap := make(map[string]interface{})
			if parent, ok := tr.GetParentOk(); ok {
				trMap["parent"] = *parent
			}
			if child, ok := tr.GetChildOk(); ok {
				trMap["child"] = *child
			}
			if ew, ok := tr.GetEvaluationWindowOk(); ok {
				trMap["evaluation_window"] = *ew
			}
			tfTransitions[idx] = trMap
		}
		if len(tfTransitions) > 0 {
			tfSeqOptions["step_transitions"] = tfTransitions
		}
		if len(tfSeqOptions) > 0 {
			tfOptions["sequence_detection_options"] = []map[string]interface{}{tfSeqOptions}
		}
	}
	return tfOptions
}

func extractReferenceTables(referenceTables []datadogV2.SecurityMonitoringReferenceTable) []interface{} {
	tfReferenceTables := make([]interface{}, len(referenceTables))
	for idx, referenceTable := range referenceTables {
		tfReferenceTable := make(map[string]interface{})
		tfReferenceTable["table_name"] = referenceTable.GetTableName()
		tfReferenceTable["column_name"] = referenceTable.GetColumnName()
		tfReferenceTable["log_field_path"] = referenceTable.GetLogFieldPath()
		tfReferenceTable["rule_query_name"] = referenceTable.GetRuleQueryName()
		tfReferenceTable["check_presence"] = referenceTable.GetCheckPresence()
		tfReferenceTables[idx] = tfReferenceTable
	}
	return tfReferenceTables
}

func extractSchedulingOptions(schedulingOptions *datadogV2.SecurityMonitoringSchedulingOptions) map[string]any {
	tfSchedulingOptions := make(map[string]any)
	tfSchedulingOptions["rrule"] = schedulingOptions.GetRrule()
	if start, ok := schedulingOptions.GetStartOk(); ok {
		tfSchedulingOptions["start"] = *start
	}
	if timezone, ok := schedulingOptions.GetTimezoneOk(); ok {
		tfSchedulingOptions["timezone"] = *timezone
	}
	return tfSchedulingOptions
}

func extractCalculatedFields(calculatedFields []datadogV2.CalculatedField) []any {
	tfCalculatedFields := make([]any, len(calculatedFields))

	for idx, calculatedField := range calculatedFields {
		tfCalculatedField := make(map[string]any)
		tfCalculatedField["name"] = calculatedField.Name
		tfCalculatedField["expression"] = calculatedField.Expression
		tfCalculatedFields[idx] = tfCalculatedField
	}

	return tfCalculatedFields
}

func resourceDatadogSecurityMonitoringRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ruleUpdate, err := buildUpdatePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}
	response, httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().UpdateSecurityMonitoringRule(auth, d.Id(), *ruleUpdate)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating security monitoring rule")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	if response.SecurityMonitoringStandardRuleResponse != nil {
		updateStandardResourceDataFromResponse(d, response.SecurityMonitoringStandardRuleResponse)
	} else if response.SecurityMonitoringSignalRuleResponse != nil {
		updateSignalResourceDataFromResponse(d, response.SecurityMonitoringSignalRuleResponse)
	}

	return nil
}

func buildUpdatePayload(d *schema.ResourceData) (*datadogV2.SecurityMonitoringRuleUpdatePayload, error) {
	payload := datadogV2.SecurityMonitoringRuleUpdatePayload{}
	if err := checkQueryConsistency(d); err != nil {
		return &datadogV2.SecurityMonitoringRuleUpdatePayload{}, err
	}

	isSignalCorrelation := isSignalCorrelationSchema(d)

	if isThirdPartyRule(d) {
		tfThirdPartyCases := d.Get("third_party_case").([]interface{})
		payloadThirdPartyCases := make([]datadogV2.SecurityMonitoringThirdPartyRuleCase, len(tfThirdPartyCases))

		for idx, tfThirdPartyCase := range tfThirdPartyCases {
			parsedCase := tfThirdPartyCase.(map[string]interface{})
			payloadCase := datadogV2.SecurityMonitoringThirdPartyRuleCase{}

			if v, ok := parsedCase["status"]; ok {
				payloadCase.SetStatus(datadogV2.SecurityMonitoringRuleSeverity(v.(string)))
			}
			if v, ok := parsedCase["notifications"]; ok {
				payloadCase.SetNotifications(parseStringArray(v.([]interface{})))
			}
			if v, ok := parsedCase["query"]; ok {
				payloadCase.SetQuery(v.(string))
			}
			if v, ok := parsedCase["name"]; ok {
				payloadCase.SetName(v.(string))
			}

			payloadThirdPartyCases[idx] = payloadCase
		}

		payload.SetThirdPartyCases(payloadThirdPartyCases)
	} else {
		tfCases := d.Get("case").([]interface{})
		payloadCases := make([]datadogV2.SecurityMonitoringRuleCase, len(tfCases))

		for idx, tfRuleCase := range tfCases {
			structRuleCase := datadogV2.SecurityMonitoringRuleCase{}

			ruleCase := tfRuleCase.(map[string]interface{})
			status := datadogV2.SecurityMonitoringRuleSeverity(ruleCase["status"].(string))
			structRuleCase.SetStatus(status)

			if name, ok := ruleCase["name"]; ok {
				structRuleCase.SetName(name.(string))
			}
			if condition, ok := ruleCase["condition"]; ok {
				structRuleCase.SetCondition(condition.(string))
			}
			if v, ok := ruleCase["notifications"]; ok {
				structRuleCase.SetNotifications(parseStringArray(v.([]interface{})))
			}
			if action, ok := ruleCase["action"]; ok && len(action.([]interface{})) > 0 {
				structRuleCase.SetActions(buildPayloadCaseActions(action.([]interface{})))
			}
			payloadCases[idx] = structRuleCase
		}
		payload.SetCases(payloadCases)

		var v interface{}
		var ok bool
		if isSignalCorrelation {
			v, ok = d.GetOk("signal_query")
		} else {
			v, ok = d.GetOk("query")
		}
		var err error
		if ok {
			tfQueries := v.([]interface{})
			payloadQueries := make([]datadogV2.SecurityMonitoringRuleQuery, len(tfQueries))
			for idx, tfQuery := range tfQueries {
				if isSignalCorrelation {
					if payloadQueries[idx], err = buildUpdateSignalRuleQuery(tfQuery); err != nil {
						return &payload, err
					}
				} else {
					payloadQueries[idx] = *buildUpdateStandardRuleQuery(tfQuery)
				}
			}

			payload.SetQueries(payloadQueries)
		}
	}

	payload.SetIsEnabled(d.Get("enabled").(bool))
	payload.SetHasExtendedTitle(d.Get("has_extended_title").(bool))

	if v, ok := d.GetOk("message"); ok {
		payload.SetMessage(v.(string))
	}

	if v, ok := d.GetOk("name"); ok {
		payload.SetName(v.(string))
	}

	if v, ok := d.GetOk("options"); ok {
		payload.Options = buildPayloadOptions(v.([]interface{}), d.Get("type").(string))
	}

	if v, ok := d.GetOk("tags"); ok {
		tfTags := v.(*schema.Set)
		tags := make([]string, tfTags.Len())
		for i, value := range tfTags.List() {
			tags[i] = value.(string)
		}
		payload.SetTags(tags)
	} else {
		payload.SetTags([]string{})
	}

	tfFilters := d.Get("filter")
	payload.SetFilters(buildPayloadFilters(tfFilters.([]interface{})))

	if !isSignalCorrelation {
		if v, ok := d.GetOk("reference_tables"); ok {
			tfReferenceTables := v.([]interface{})
			payload.SetReferenceTables(buildPayloadReferenceTables(tfReferenceTables))
		} else if d.HasChange("reference_tables") {
			// Only send empty list if reference_tables was removed in config
			payload.SetReferenceTables(make([]datadogV2.SecurityMonitoringReferenceTable, 0))
		}

		if v, ok := d.GetOk("group_signals_by"); ok {
			payload.SetGroupSignalsBy(parseStringArray(v.([]interface{})))
		} else if d.HasChange("group_signals_by") {
			// Only send empty list if group_signals_by was removed in config
			payload.SetGroupSignalsBy([]string{})
		}

		if v, ok := d.GetOk("scheduling_options"); ok {
			tfSchedulingOptions := v.([]any)
			payload.SetSchedulingOptions(*buildPayloadSchedulingOptions(tfSchedulingOptions))
		} else {
			payload.SetSchedulingOptionsNil()
		}

		if v, ok := d.GetOk("calculated_field"); ok {
			payload.SetCalculatedFields(buildPayloadCalculatedFields(v.([]any)))
		} else {
			payload.SetCalculatedFields([]datadogV2.CalculatedField{})
		}
	}

	return &payload, nil
}

func buildUpdateStandardRuleQuery(tfQuery interface{}) *datadogV2.SecurityMonitoringRuleQuery {
	query := tfQuery.(map[string]interface{})
	payloadQuery := datadogV2.SecurityMonitoringStandardRuleQuery{}

	if v, ok := query["aggregation"]; ok {
		aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
		payloadQuery.SetAggregation(aggregation)
	}

	if v, ok := query["group_by_fields"]; ok {
		payloadQuery.SetGroupByFields(parseStringArray(v.([]interface{})))
	}

	if v, ok := query["has_optional_group_by_fields"]; ok {
		payloadQuery.SetHasOptionalGroupByFields(v.(bool))
	}

	if v, ok := query["distinct_fields"]; ok {
		payloadQuery.SetDistinctFields(parseStringArray(v.([]interface{})))
	}

	if v, ok := query["data_source"]; ok {
		dataSource := datadogV2.SecurityMonitoringStandardDataSource(v.(string))
		payloadQuery.SetDataSource(dataSource)
	}

	if v, ok := query["metric"]; ok {
		metric := v.(string)
		payloadQuery.SetMetric(metric)
	}

	if v, ok := query["metrics"]; ok {
		payloadQuery.SetMetrics(parseStringArray(v.([]interface{})))
	}

	if v, ok := query["name"]; ok {
		name := v.(string)
		payloadQuery.SetName(name)
	}

	if v, ok := query["query"]; ok {
		queryQuery := v.(string)
		payloadQuery.SetQuery(queryQuery)
	}

	if v, ok := query["custom_query_extension"]; ok {
		queryExtension := v.(string)
		payloadQuery.SetCustomQueryExtension(queryExtension)
	}

	if v, ok := query["indexes"]; ok {
		if indexes := parseStringArray(v.([]any)); len(indexes) > 0 {
			payloadQuery.SetIndex(indexes[0])
		}
	}

	standardRuleQuery := datadogV2.SecurityMonitoringStandardRuleQueryAsSecurityMonitoringRuleQuery(&payloadQuery)
	return &standardRuleQuery
}

func buildUpdateSignalRuleQuery(tfQuery interface{}) (datadogV2.SecurityMonitoringRuleQuery, error) {
	query := tfQuery.(map[string]interface{})
	payloadQuery := datadogV2.SecurityMonitoringSignalRuleQuery{}

	if v, ok := query["aggregation"]; ok {
		aggregation := datadogV2.SecurityMonitoringRuleQueryAggregation(v.(string))
		payloadQuery.SetAggregation(aggregation)
	}

	if v, ok := query["correlated_by_fields"]; ok {
		payloadQuery.SetCorrelatedByFields(parseStringArray(v.([]interface{})))
	}

	if v, ok := query["correlated_query_index"]; ok && len(v.(string)) > 0 {
		if vInt, err := strconv.Atoi(v.(string)); err == nil {
			payloadQuery.SetCorrelatedQueryIndex(int32(vInt))
		}
	}

	if v, ok := query["name"]; ok {
		name := v.(string)
		payloadQuery.SetName(name)
	}

	ruleIdQuery := query["rule_id"].(string)
	payloadQuery.SetRuleId(ruleIdQuery)

	var err error
	if v, ok := query["default_rule_id"].(string); ok && v != "" {
		err = fmt.Errorf("default_rule_id cannot be set")
	}

	return datadogV2.SecurityMonitoringSignalRuleQueryAsSecurityMonitoringRuleQuery(&payloadQuery), err
}

func resourceDatadogSecurityMonitoringRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().DeleteSecurityMonitoringRule(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting security monitoring rule")
	}

	return nil
}

func resourceDatadogSecurityMonitoringRuleCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if validate, ok := diff.GetOkExists("validate"); !ok || !validate.(bool) {
		// Explicitly skip validation
		log.Printf("[DEBUG] Validate is %v, skipping validation", validate.(bool))
		return nil
	}

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if payload, err := buildValidatePayload(diff); err == nil {
		if httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().ValidateSecurityMonitoringRule(auth, *payload); err != nil || httpResponse == nil {
			return utils.TranslateClientError(err, httpResponse, "error validating security monitoring rule")
		}
	} else {
		log.Printf("[DEBUG] Skipping validation due to an error: %v", err)
	}
	return nil
}
