package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

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
								"instantaneous_baseline": {
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
									Description: "When set to true, Datadog uses previous values that fall within the defined learning window to construct the baseline, enabling the system to establish an accurate baseline more rapidly rather than relying solely on gradual learning over time.",
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
								"instantaneous_baseline": {
									Type:        schema.TypeBool,
									Optional:    true,
									Default:     false,
									Description: "When set to true, Datadog uses previous values that fall within the defined learning window to construct the baseline, enabling the system to establish an accurate baseline more rapidly rather than relying solely on gradual learning over time.",
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
						ValidateDiagFunc: validators.ValidateSecurityMonitoringDataSource(validators.ValidateEnumValue(datadogV2.NewSecurityMonitoringStandardDataSourceFromValue)),
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

func isSignalCorrelationSchema(d utils.Resource) bool {
	if v, ok := d.GetOk("type"); ok {
		_, err := datadogV2.NewSecurityMonitoringSignalRuleTypeFromValue(v.(string))
		return err == nil
	}
	return false
}

// isLearningPeriodBaselineConfigured checks the raw HCL config to determine whether the user
// explicitly set learning_period_baseline. SDKv2 always populates optional int attrs with their
// zero value, so map key presence alone cannot distinguish "set to 0" from "not set".
func isLearningPeriodBaselineConfigured(d utils.Resource) bool {
	val, diags := d.GetRawConfigAt(
		cty.GetAttrPath("options").IndexInt(0).
			GetAttr("anomaly_detection_options").IndexInt(0).
			GetAttr("learning_period_baseline"),
	)
	return !diags.HasError() && !val.IsNull()
}

func buildPayloadOptions(d utils.Resource, tfOptionsList []interface{}, ruleType string) *datadogV2.SecurityMonitoringRuleOptions {
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
		learningPeriodBaselineConfigured := isLearningPeriodBaselineConfigured(d)
		if payloadAnomalyDetectionOptions, ok := buildPayloadAnomalyDetectionOptions(tfAnomalyDetectionOptionsList, learningPeriodBaselineConfigured); ok {
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

func buildPayloadAnomalyDetectionOptions(tfOptionsList []interface{}, learningPeriodBaselineConfigured bool) (*datadogV2.SecurityMonitoringRuleAnomalyDetectionOptions, bool) {
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

	if v, ok := tfOptions["instantaneous_baseline"]; ok {
		hasPayload = true
		options.SetInstantaneousBaseline(v.(bool))
	}

	// Only include learning_period_baseline when the user explicitly set it in HCL.
	// SDKv2 always populates optional int attrs with 0, but 0 is a valid API value
	// (means "immediately generate signals"), so we use the raw config to distinguish.
	if learningPeriodBaselineConfigured {
		hasPayload = true
		learningPeriodBaseline := int64(tfOptions["learning_period_baseline"].(int))
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
	if v, ok := tfOptions["instantaneous_baseline"]; ok {
		hasPayload = true
		payloadNewValueRulesOptions.SetInstantaneousBaseline(v.(bool))
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
		tfNewValueOptions["instantaneous_baseline"] = bool(newValueOptions.GetInstantaneousBaseline())
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
		tfAnomalyDetectionOptions["instantaneous_baseline"] = bool(anomalyDetectionOptions.GetInstantaneousBaseline())
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

func resourceDatadogSecurityMonitoringRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetSecurityMonitoringApiV2().DeleteSecurityMonitoringRule(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting security monitoring rule")
	}

	return nil
}
