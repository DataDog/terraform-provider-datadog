package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultNoDataTimeframeMinutes = 10

var retryTimeout = time.Minute

func resourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.",
		CreateContext: resourceDatadogMonitorCreate,
		ReadContext:   resourceDatadogMonitorRead,
		UpdateContext: resourceDatadogMonitorUpdate,
		DeleteContext: resourceDatadogMonitorDelete,
		CustomizeDiff: customdiff.All(tagDiff, resourceDatadogMonitorCustomizeDiff),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description: "Name of Datadog monitor.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"message": {
					Description: "A message to include with notifications for this monitor.\n\nEmail notifications can be sent to specific users by using the same `@username` notation as events.",
					Type:        schema.TypeString,
					Required:    true,
					StateFunc: func(val interface{}) string {
						return strings.TrimSpace(val.(string))
					},
				},
				"escalation_message": {
					Description: "A message to include with a re-notification. Supports the `@username` notification allowed elsewhere.",
					Type:        schema.TypeString,
					Optional:    true,
					StateFunc: func(val interface{}) string {
						return strings.TrimSpace(val.(string))
					},
				},
				"query": {
					Description: "The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.\n\n**Note:** APM latency data is now available as Distribution Metrics. Existing monitors have been migrated automatically but all terraformed monitors can still use the existing metrics. We strongly recommend updating monitor definitions to query the new metrics. To learn more, or to see examples of how to update your terraform definitions to utilize the new distribution metrics, see the [detailed doc](https://docs.datadoghq.com/tracing/guide/ddsketch_trace_metrics/).",
					Type:        schema.TypeString,
					Required:    true,
					StateFunc: func(val interface{}) string {
						return strings.TrimSpace(val.(string))
					},
				},
				"type": {
					Description:      "The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.",
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorTypeFromValue),
					// Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
					DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
						if (oldVal == "query alert" && newVal == "metric alert") ||
							(oldVal == "metric alert" && newVal == "query alert") {
							log.Printf("[DEBUG] Monitor '%s' got a '%s' response for an expected '%s' type. Suppressing change.", d.Get("name"), newVal, oldVal)
							return true
						}
						return newVal == oldVal
					},
				},
				"priority": {
					Description: "Integer from 1 (high) to 5 (low) indicating alert severity.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				// Options
				"monitor_thresholds": {
					Description: "Alert thresholds of the monitor.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ok": {
								Description:  "The monitor `OK` threshold. Only supported in monitor type `service check`. Must be a number.",
								Type:         schema.TypeString,
								ValidateFunc: validators.ValidateFloatString,
								Optional:     true,
								DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
									monitorType := d.Get("type").(string)
									return monitorType != string(datadogV1.MONITORTYPE_SERVICE_CHECK)
								},
							},
							"warning": {
								Description:  "The monitor `WARNING` threshold. Must be a number.",
								Type:         schema.TypeString,
								ValidateFunc: validators.ValidateFloatString,
								Optional:     true,
							},
							"critical": {
								Description:  "The monitor `CRITICAL` threshold. Must be a number.",
								Type:         schema.TypeString,
								ValidateFunc: validators.ValidateFloatString,
								Optional:     true,
							},
							"unknown": {
								Description:  "The monitor `UNKNOWN` threshold. Only supported in monitor type `service check`. Must be a number.",
								Type:         schema.TypeString,
								ValidateFunc: validators.ValidateFloatString,
								Optional:     true,
								DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
									monitorType := d.Get("type").(string)
									return monitorType != string(datadogV1.MONITORTYPE_SERVICE_CHECK)
								},
							},
							"warning_recovery": {
								Description:  "The monitor `WARNING` recovery threshold. Must be a number.",
								Type:         schema.TypeString,
								ValidateFunc: validators.ValidateFloatString,
								Optional:     true,
							},
							"critical_recovery": {
								Description:  "The monitor `CRITICAL` recovery threshold. Must be a number.",
								Type:         schema.TypeString,
								ValidateFunc: validators.ValidateFloatString,
								Optional:     true,
							},
							"critical_query": {
								Description: "Query evaluated as a dynamic `CRITICAL` threshold. Only supported on metric monitors with a formula query and `options['variables']`. Cannot be combined with static thresholds. This field is in preview.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"critical_recovery_query": {
								Description: "Query evaluated as a dynamic `CRITICAL` recovery threshold. Only supported on metric monitors with a formula query and `options['variables']`. Cannot be combined with static thresholds. This field is in preview.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
					DiffSuppressFunc: suppressDataDogFloatIntDiff,
				},
				"monitor_threshold_windows": {
					Description: "A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"recovery_window": {
								Description: "Describes how long an anomalous metric must be normal before the alert recovers.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"trigger_window": {
								Description: "Describes how long a metric must be anomalous before an alert triggers.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
				"notify_no_data": {
					Description:   "A boolean indicating whether this monitor will notify when data stops reporting.",
					Type:          schema.TypeBool,
					Optional:      true,
					Default:       false,
					ConflictsWith: []string{"on_missing_data"},
				},
				"on_missing_data": {
					Description:   "Controls how groups or monitors are treated if an evaluation does not return any data points. The default option results in different behavior depending on the monitor query type. For monitors using `Count` queries, an empty monitor evaluation is treated as 0 and is compared to the threshold conditions. For monitors using any query type other than `Count`, for example `Gauge`, `Measure`, or `Rate`, the monitor shows the last known status. This option is not available for Service Check, Composite, or SLO monitors. Valid values are: `show_no_data`, `show_and_notify_no_data`, `resolve`, and `default`.",
					Type:          schema.TypeString,
					Optional:      true,
					ConflictsWith: []string{"notify_no_data", "no_data_timeframe"},
				},
				"group_retention_duration": {
					Description: "The time span after which groups with missing data are dropped from the monitor state. The minimum value is one hour, and the maximum value is 72 hours. Example values are: 60m, 1h, and 2d. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				// We only set new_group_delay in the monitor API payload if it is nonzero
				// because the SDKv2 terraform plugin API prevents unsetting new_group_delay
				// in updateMonitorState, so we can't reliably distinguish between new_group_delay
				// being unset (null) or set to zero.
				// Note that "new_group_delay overrides new_host_delay if it is set to a nonzero value"
				// refers to this terraform resource. In the API, setting new_group_delay
				// to any value, including zero, causes it to override new_host_delay.
				"new_group_delay": {
					Description: "The time (in seconds) to skip evaluations for new groups.\n\n`new_group_delay` overrides `new_host_delay` if it is set to a nonzero value.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"new_host_delay": {
					// Removing the default requires removing the default in the API as well (possibly only for
					// terraform user agents)
					Description: "**Deprecated**. See `new_group_delay`. Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non-negative integer. This value is ignored for simple monitors and monitors not grouped by host. The only case when this should be used is to override the default and set `new_host_delay` to zero for monitors grouped by host.",
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     300,
					Deprecated:  "Use `new_group_delay` except when setting `new_host_delay` to zero.",
				},
				"evaluation_delay": {
					Description: "(Only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.\n\nFor example, if the value is set to `300` (5min), the `timeframe` is set to `last_5m` and the time is 7:00, the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled metrics to ensure the monitor will always have data during evaluation.",
					Type:        schema.TypeInt,
					Computed:    true,
					Optional:    true,
				},
				"no_data_timeframe": {
					Description: "The number of minutes before a monitor will notify when data stops reporting.\n\nWe recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.",
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     defaultNoDataTimeframeMinutes,
					DiffSuppressFunc: func(k, oldVal, newVal string, d *schema.ResourceData) bool {
						if !d.Get("notify_no_data").(bool) {
							if newVal != oldVal {
								log.Printf("[DEBUG] Ignore the no_data_timeframe change of monitor '%s' because notify_no_data is false.", d.Get("name"))
							}
							return true
						}
						return newVal == oldVal
					},
					ConflictsWith: []string{"on_missing_data"},
				},
				"renotify_interval": {
					Description: "The number of minutes after the last notification before a monitor will re-notify on the current status. It will only re-notify if it's not resolved.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"renotify_occurrences": {
					Description: "The number of re-notification messages that should be sent on the current status.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"renotify_statuses": {
					Description: "The types of statuses for which re-notification messages should be sent.",
					Type:        schema.TypeSet,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorRenotifyStatusTypeFromValue),
					},
					Optional: true,
				},
				"notify_audit": {
					Description: "A boolean indicating whether tagged users will be notified on changes to this monitor. Defaults to `false`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"timeout_h": {
					Description: "The number of hours of the monitor not reporting data before it automatically resolves from a triggered state. The minimum allowed value is 0 hours. The maximum allowed value is 24 hours.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"require_full_window": {
					Description: "A boolean indicating whether this monitor needs a full window of data before it's evaluated. Datadog strongly recommends you set this to `false` for sparse metrics, otherwise some evaluations may be skipped. If there's a custom_schedule set, `require_full_window` must be false and will be ignored.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						if attr, ok := d.GetOk("scheduling_options"); ok {
							scheduling_options_list := attr.([]interface{})
							if scheduling_options_map, ok := scheduling_options_list[0].(map[string]interface{}); ok {
								custom_schedule_map, custom_schedule_found := scheduling_options_map["custom_schedule"].([]interface{})
								if custom_schedule_found && len(custom_schedule_map) > 0 {
									return true
								}
							}
						}
						return false
					},
				},
				"restricted_roles": {
					Description: "A list of unique role identifiers to define which roles are allowed to edit the monitor. Editing a monitor includes any updates to the monitor configuration, monitor deletion, and muting of the monitor for any amount of time. Roles unique identifiers can be pulled from the [Roles API](https://docs.datadoghq.com/api/latest/roles/#list-roles) in the `data.id` field.",
					Type:        schema.TypeSet,
					Optional:    true,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
					Deprecated:  "Use `datadog_restriction_policy` resource to manage permission.",
				},
				"include_tags": {
					Description: "A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"tags": {
					Description: "A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
					// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
					// we order them explicitly in the read/create/update methods of this resource and using
					// TypeSet makes Terraform ignore differences in order when creating a plan
					Type:     schema.TypeSet,
					Optional: true,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"ignore_tag_keys": {
					Type:        schema.TypeSet,
					Description: "Tag keys whose drift Terraform should ignore. Use this to keep specific tags managed outside Terraform (e.g. by the Datadog UI or a tagging service) without `terraform plan` reporting drift on every run. Other tags are still managed normally.",
					Optional:    true,
					Elem: &schema.Schema{
						Type: schema.TypeString,
						StateFunc: func(val any) string {
							return utils.NormalizeTag(val.(string))
						},
					},
				},
				"groupby_simple_monitor": {
					Description: "Whether or not to trigger one alert if any source breaches a threshold. This is only used by log monitors. Defaults to `false`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"notify_by": {
					Description: "Controls what granularity a monitor alerts on. Only available for monitors with groupings. For instance, a monitor grouped by `cluster`, `namespace`, and `pod` can be configured to only notify on each new `cluster` violating the alert conditions by setting `notify_by` to `['cluster']`. Tags mentioned in `notify_by` must be a subset of the grouping tags in the query. For example, a query grouped by `cluster` and `namespace` cannot notify on `region`. Setting `notify_by` to `[*]` configures the monitor to notify as a simple-alert.",
					Type:        schema.TypeSet,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				// since this is only useful for "log alert" type, we don't set a default value
				// if we did set it, it would be used for all types; we have to handle this manually
				// throughout the code
				"enable_logs_sample": {
					Description: "A boolean indicating whether or not to include a list of log values which triggered the alert. This is only used by log monitors. Defaults to `false`.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"enable_samples": {
					Description: "Whether or not a list of samples which triggered the alert is included. This is only used by CI Test and Pipeline monitors.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"force_delete": {
					Description: "A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. SLO, composite monitor).",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"validate": {
					Description: "If set to `false`, skip the validation call done during plan.",
					Type:        schema.TypeBool,
					Optional:    true,
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// This is never sent to the backend, so it should never generate a diff
						return true
					},
				},
				"variables": getMonitorFormulaQuerySchema(),
				"scheduling_options": {
					Description: "Configuration options for scheduling.",
					Type:        schema.TypeList,
					MaxItems:    1,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"evaluation_window": {
								Description: "Configuration options for the evaluation window. If `hour_starts` is set, no other fields may be set. Otherwise, `day_starts` and `month_starts` must be set together.",
								Type:        schema.TypeList,
								MaxItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"day_starts": {
											Description: "The time of the day at which a one day cumulative evaluation window starts. Must be defined in UTC time in `HH:mm` format.",
											Type:        schema.TypeString,
											Optional:    true,
										},
										"month_starts": {
											Description: "The day of the month at which a one month cumulative evaluation window starts. Must be a value of 1.",
											Type:        schema.TypeInt,
											Optional:    true,
										},
										"hour_starts": {
											Description: "The minute of the hour at which a one hour cumulative evaluation window starts. Must be between 0 and 59.",
											Type:        schema.TypeInt,
											Optional:    true,
										},
										"timezone": {
											Description: "The timezone for the cumulative evaluation window start time.",
											Type:        schema.TypeString,
											Optional:    true,
										},
									},
								},
							},
							"custom_schedule": {
								Description: "Configuration options for the custom schedules. If `start` is omitted, the monitor creation time will be used.",
								Type:        schema.TypeList,
								MaxItems:    1,
								Optional:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"recurrence": {
											Description: "A list of recurrence definitions. Length must be 1.",
											Type:        schema.TypeList,
											Required:    true,
											MaxItems:    1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"rrule": {
														Description: "Must be a valid `rrule`. See API docs for supported fields",
														Type:        schema.TypeString,
														Required:    true,
													},
													"start": {
														Description: "Time to start recurrence cycle. Similar to DTSTART. Expected format 'YYYY-MM-DDThh:mm:ss'",
														Type:        schema.TypeString,
														Optional:    true,
													},
													"timezone": {
														Description: "'tz database' format. Example: `America/New_York` or `UTC`",
														Type:        schema.TypeString,
														Required:    true,
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
				"notification_preset_name": {
					Description:      "Toggles the display of additional content sent in the monitor notification.",
					Type:             schema.TypeString,
					Optional:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorOptionsNotificationPresetsFromValue),
				},
				"draft_status": {
					Description:      "Indicates whether the monitor is in a draft or published state. When set to `draft`, the monitor appears as Draft and does not send notifications. When set to `published`, the monitor is active, and it evaluates conditions and sends notifications as configured.",
					Type:             schema.TypeString,
					Optional:         true,
					Default:          string(datadogV1.MONITORDRAFTSTATUS_PUBLISHED),
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorDraftStatusFromValue),
				},
				"assets": {
					Description: "List of monitor assets (for example, runbooks, dashboards, workflows) tied to this monitor.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Description: "Name for the monitor asset.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"url": {
								Description: "URL for the asset.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"category": {
								Description:      "Type of asset the entity represents on a monitor.",
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorAssetCategoryFromValue),
							},
							"resource_key": {
								Description: "Identifier of the internal Datadog resource that this asset represents.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"resource_type": {
								Description:      "Type of internal Datadog resource associated with a monitor asset.",
								Type:             schema.TypeString,
								Optional:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorAssetResourceTypeFromValue),
							},
						},
					},
				},
			}
		},
	}
}

// eventQueryVariableSchema returns the nested schema for a formula monitor variables event query (reused for aggregate_augmented augment/base branches).
func eventQueryVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"data_source": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventsDataSourceFromValue),
			Description:      "The data source for event platform-based queries.",
		},
		"search": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			MinItems:    1,
			Description: "The search options.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The events search string.",
					},
				},
			},
		},
		"indexes": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "An array of index names to query in the stream.",
		},
		"compute": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "The compute options.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventAggregationFromValue),
						Description:      "The aggregation methods for event platform queries.",
					},
					"interval": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "A time interval in milliseconds.",
					},
					"metric": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The measurable attribute to compute.",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The name assigned to this aggregation when multiple aggregations are defined for a query.",
					},
				},
			},
		},
		"group_by": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Group by options.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"facet": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The event facet.",
					},
					"source": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "For composite aggregate-augmented queries, identifies which sub-query this group-by facet refers to (for example `filter_query`).",
					},
					"limit": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "The number of groups to return.",
					},
					"sort": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "The options for sorting group by results.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"aggregation": {
									Type:             schema.TypeString,
									Required:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventAggregationFromValue),
									Description:      "The aggregation methods for the event platform queries.",
								},
								"metric": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The metric used for sorting group by results.",
								},
								"order": {
									Type:             schema.TypeString,
									Optional:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
									Description:      "Direction of sort.",
								},
							},
						},
					},
				},
			},
		},
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of query for use in formulas.",
		},
	}
}

func aggregateAugmentedQueryVariableSchema() map[string]*schema.Schema {
	refTableCol := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Reference table column name.",
		},
		"alias": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Optional alias for the column.",
		},
	}
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the query for use in formulas.",
		},
		"data_source": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionAggregateAugmentedDataSourceFromValue),
			Description:      "The data source for aggregate-augmented composite queries. Must be `aggregate_augmented_query`.",
		},
		"augment_reference_table": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Reference table augment query. Do not set `augment_event_query` in the same block.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the augment sub-query.",
					},
					"data_source": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionReferenceTableDataSourceFromValue),
						Description:      "Must be `reference_table`.",
					},
					"table_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the reference table.",
					},
					"query_filter": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional filter expression for the reference table query.",
					},
					"columns": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Columns to retrieve from the reference table.",
						Elem: &schema.Resource{
							Schema: refTableCol,
						},
					},
				},
			},
		},
		"augment_event_query": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Events augment query. Do not set `augment_reference_table` in the same block.",
			Elem: &schema.Resource{
				Schema: eventQueryVariableSchema(),
			},
		},
		"base_metrics_query": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Metrics base query. Do not set `base_event_query` in the same block.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"data_source": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionMetricsDataSourceFromValue),
						Description:      "The data source for metrics queries.",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The name of the query for use in formulas.",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The metrics query definition.",
					},
					"aggregator": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionMetricsAggregatorFromValue),
						Description:      "The aggregation method for metrics queries.",
					},
				},
			},
		},
		"base_event_query": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Events base query. Do not set `base_metrics_query` in the same block.",
			Elem: &schema.Resource{
				Schema: eventQueryVariableSchema(),
			},
		},
		"join_condition": {
			Type:        schema.TypeList,
			Required:    true,
			MaxItems:    1,
			MinItems:    1,
			Description: "Join condition between augment and base queries.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"augment_attribute": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Attribute from the augment query to join on.",
					},
					"base_attribute": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Attribute from the base query to join on.",
					},
					"join_type": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionAggregateQueryJoinTypeFromValue),
						Description:      "Join type (for example `inner`).",
					},
				},
			},
		},
		"compute": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "Compute aggregations for the aggregate-augmented query.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventAggregationFromValue),
						Description:      "The aggregation methods for compute steps.",
					},
					"interval": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "A time interval in milliseconds.",
					},
					"metric": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The measurable attribute to compute.",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The name assigned to this aggregation when multiple aggregations are defined.",
					},
				},
			},
		},
		"group_by": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "Group by options for the aggregate-augmented query. At least one block is required.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"facet": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The facet to group by.",
					},
					"source": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Identifies which sub-query this facet refers to (for example `filter_query`).",
					},
					"limit": {
						Type:        schema.TypeInt,
						Optional:    true,
						Description: "The number of groups to return.",
					},
					"sort": {
						Type:        schema.TypeList,
						Optional:    true,
						MaxItems:    1,
						Description: "Sort options for group by.",
						Elem: &schema.Resource{
							Schema: map[string]*schema.Schema{
								"aggregation": {
									Type:             schema.TypeString,
									Required:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventAggregationFromValue),
									Description:      "The aggregation methods for sorting.",
								},
								"metric": {
									Type:        schema.TypeString,
									Optional:    true,
									Description: "The metric used for sorting group by results.",
								},
								"order": {
									Type:             schema.TypeString,
									Optional:         true,
									ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
									Description:      "Direction of sort.",
								},
							},
						},
					},
				},
			},
		},
	}
}

// aggregateFilteredQueryVariableSchema defines variables for aggregate_filtered_query (formula monitors).
func aggregateFilteredQueryVariableSchema() map[string]*schema.Schema {
	refTableCol := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Reference table column name.",
		},
		"alias": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Optional alias for the column.",
		},
	}
	computeElem := map[string]*schema.Schema{
		"aggregation": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventAggregationFromValue),
			Description:      "The aggregation methods for compute steps.",
		},
		"interval": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "A time interval in milliseconds.",
		},
		"metric": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The measurable attribute to compute.",
		},
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name assigned to this aggregation when multiple aggregations are defined.",
		},
	}
	groupByElem := map[string]*schema.Schema{
		"facet": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The facet to group by.",
		},
		"source": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Identifies which sub-query this facet refers to (for example `filter_query`).",
		},
		"limit": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "The number of groups to return.",
		},
		"sort": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Sort options for group by.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"aggregation": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionEventAggregationFromValue),
						Description:      "The aggregation methods for sorting.",
					},
					"metric": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The metric used for sorting group by results.",
					},
					"order": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewQuerySortOrderFromValue),
						Description:      "Direction of sort.",
					},
				},
			},
		},
	}
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the query for use in formulas.",
		},
		"data_source": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionAggregateFilteredDataSourceFromValue),
			Description:      "The data source for aggregate-filtered composite queries. Must be `aggregate_filtered_query`.",
		},
		"filter_reference_table": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Reference table filter query. Do not set `filter_event_query` in the same block.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Name of the filter sub-query.",
					},
					"data_source": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionReferenceTableDataSourceFromValue),
						Description:      "Must be `reference_table`.",
					},
					"table_name": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Name of the reference table.",
					},
					"query_filter": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "Optional filter expression for the reference table query.",
					},
					"columns": {
						Type:        schema.TypeList,
						Optional:    true,
						Description: "Columns to retrieve from the reference table.",
						Elem: &schema.Resource{
							Schema: refTableCol,
						},
					},
				},
			},
		},
		"filter_event_query": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Events filter query. Do not set `filter_reference_table` in the same block.",
			Elem: &schema.Resource{
				Schema: eventQueryVariableSchema(),
			},
		},
		"base_metrics_query": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Metrics base query. Do not set `base_event_query` in the same block.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"data_source": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionMetricsDataSourceFromValue),
						Description:      "The data source for metrics queries.",
					},
					"name": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The name of the query for use in formulas.",
					},
					"query": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The metrics query definition.",
					},
					"aggregator": {
						Type:             schema.TypeString,
						Optional:         true,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionMetricsAggregatorFromValue),
						Description:      "The aggregation method for metrics queries.",
					},
				},
			},
		},
		"base_event_query": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Events base query. Do not set `base_metrics_query` in the same block.",
			Elem: &schema.Resource{
				Schema: eventQueryVariableSchema(),
			},
		},
		"filters": {
			Type:        schema.TypeList,
			Required:    true,
			MinItems:    1,
			Description: "Filter conditions mapping base query attributes to filter query attributes.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"base_attribute": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Attribute from the base query to filter on.",
					},
					"filter_attribute": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "Attribute from the filter query to match against.",
					},
					"exclude": {
						Type:        schema.TypeBool,
						Optional:    true,
						Description: "When true, exclude matching records instead of including them.",
					},
				},
			},
		},
		"compute": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Optional compute aggregations for the aggregate-filtered query.",
			Elem: &schema.Resource{
				Schema: computeElem,
			},
		},
		"group_by": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Optional group by options for the aggregate-filtered query.",
			Elem: &schema.Resource{
				Schema: groupByElem,
			},
		},
	}
}

// Monitor specific schema for formula and functions. Should be a strict
// subset of getFormulaQuerySchema with the appropriate types.
func getMonitorFormulaQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"event_query": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "A timeseries formula and functions events query.",
					Elem: &schema.Resource{
						Schema: eventQueryVariableSchema(),
					},
				},
				"cloud_cost_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    5,
					Description: "The Cloud Cost query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionCostDataSourceFromValue),
								Description:      "The data source for cloud cost queries.",
							},
							"query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The cloud cost query definition.",
							},
							"aggregator": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionCostAggregatorFromValue),
								Description:      "The aggregation methods available for cloud cost queries.",
							},
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the query for use in formulas.",
							},
						},
					},
				},
				"data_quality_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    5,
					Description: "The Data Quality query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The name of the query for use in formulas.",
							},
							"data_source": {
								Type:             schema.TypeString,
								Required:         true,
								ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorFormulaAndFunctionDataQualityDataSourceFromValue),
								Description:      "The data source for data quality queries. Valid value is `data_quality_metrics`.",
							},
							"schema_version": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Schema version for the data quality query.",
							},
							"measure": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The measure to query. Common values include `bytes`, `cardinality`, `custom`, `freshness`, `max`, `mean`, `min`, `nullness`, `percent_negative`, `percent_zero`, `row_count`, `stddev`, `sum`, `uniqueness`. Additional values may be supported.",
							},
							"filter": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Filter expression used to match on data entities. Uses AAstra query syntax.",
							},
							"scope": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Optional scoping expression to further filter metrics.",
							},
							"group_by": {
								Type:        schema.TypeList,
								Optional:    true,
								Elem:        &schema.Schema{Type: schema.TypeString},
								Description: "Optional grouping fields for aggregation.",
							},
							"monitor_options": {
								Type:        schema.TypeList,
								Optional:    true,
								MaxItems:    1,
								Description: "Monitor configuration options for data quality queries.",
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"custom_sql": {
											Type:        schema.TypeString,
											Optional:    true,
											Description: "Custom SQL query for the monitor.",
										},
										"custom_where": {
											Type:        schema.TypeString,
											Optional:    true,
											Description: "Custom WHERE clause for the query.",
										},
										"group_by_columns": {
											Type:        schema.TypeList,
											Optional:    true,
											Elem:        &schema.Schema{Type: schema.TypeString},
											Description: "Columns to group results by.",
										},
										"crontab_override": {
											Type:        schema.TypeString,
											Optional:    true,
											Description: "Crontab expression to override the default schedule.",
										},
										"model_type_override": {
											Type:        schema.TypeString,
											Optional:    true,
											Description: "Override for the model type. Valid values are `freshness`, `percentage`, `any`.",
										},
									},
								},
							},
						},
					},
				},
				"aggregate_augmented_query": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Aggregate-augmented composite query variables (reference table augment joined to a metrics or events base query).",
					Elem: &schema.Resource{
						Schema: aggregateAugmentedQueryVariableSchema(),
					},
				},
				"aggregate_filtered_query": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Aggregate-filtered composite query variables (filter base query results using a reference table or events filter query).",
					Elem: &schema.Resource{
						Schema: aggregateFilteredQueryVariableSchema(),
					},
				},
				"data_jobs_query": {
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    5,
					Description: "The Data Jobs query using formulas and functions.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Name of the query for use in formulas. Must be `run_query`.",
							},
							"jobs_query": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Filter expression used to select the jobs to monitor.",
							},
							"job_type": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "The type of job being monitored. Valid values include `databricks.job`, `spark.application`, `airflow.dag`, `dbt.job`, `dbt.model`, `dbt.test`, `glue.job`. Custom job types are supported with the `custom.ol.` prefix.",
							},
							"query_dialect": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Query dialect for data jobs queries. Currently only `metric` is supported.",
							},
						},
					},
				},
			},
		},
	}
}

func buildMonitorStruct(d utils.Resource) (*datadogV1.Monitor, *datadogV1.MonitorUpdateRequest) {
	var thresholds datadogV1.MonitorThresholds

	if r, ok := d.GetOk("monitor_thresholds.0.ok"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetOk(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.warning"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarning(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.unknown"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetUnknown(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCritical(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.warning_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetWarningRecovery(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical_recovery"); ok {
		v, _ := json.Number(r.(string)).Float64()
		thresholds.SetCriticalRecovery(v)
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical_query"); ok {
		thresholds.SetCriticalQuery(r.(string))
	}
	if r, ok := d.GetOk("monitor_thresholds.0.critical_recovery_query"); ok {
		thresholds.SetCriticalRecoveryQuery(r.(string))
	}

	var thresholdWindows datadogV1.MonitorThresholdWindowOptions

	if r, ok := d.GetOk("monitor_threshold_windows.0.recovery_window"); ok {
		thresholdWindows.SetRecoveryWindow(r.(string))
	}

	if r, ok := d.GetOk("monitor_threshold_windows.0.trigger_window"); ok {
		thresholdWindows.SetTriggerWindow(r.(string))
	}

	o := datadogV1.MonitorOptions{}
	hasCustomSchedule := false
	if attr, ok := d.GetOk("scheduling_options"); ok {
		scheduling_options_list := attr.([]interface{})

		if scheduling_options_map, ok := scheduling_options_list[0].(map[string]interface{}); ok {
			scheduling_options := datadogV1.NewMonitorOptionsSchedulingOptions()
			evaluation_window_list, evaluation_list_found := scheduling_options_map["evaluation_window"].([]interface{})
			if evaluation_list_found && len(evaluation_window_list) > 0 {
				if evaluation_window_map, evaluation_window_ok := evaluation_window_list[0].(map[string]interface{}); evaluation_window_ok {
					evaluation_window := datadogV1.NewMonitorOptionsSchedulingOptionsEvaluationWindow()
					day_month_scheduling := false
					if day_starts, ok := evaluation_window_map["day_starts"].(string); ok && day_starts != "" {
						evaluation_window.SetDayStarts(day_starts)
						day_month_scheduling = true
					}
					if month_starts, ok := evaluation_window_map["month_starts"].(int); ok && month_starts != 0 {
						evaluation_window.SetMonthStarts(int32(month_starts))
						day_month_scheduling = true
					}
					if hour_starts, ok := evaluation_window_map["hour_starts"].(int); ok && !day_month_scheduling {
						evaluation_window.SetHourStarts(int32(hour_starts))
					}
					if timezone, ok := evaluation_window_map["timezone"].(string); ok && timezone != "" {
						evaluation_window.SetTimezone(timezone)
					}
					scheduling_options.SetEvaluationWindow(*evaluation_window)
				}
			}
			custom_schedule_map, custom_schedule_found := scheduling_options_map["custom_schedule"].([]interface{})
			if custom_schedule_found && len(custom_schedule_map) > 0 {
				hasCustomSchedule = true
				if recurrences, ok := custom_schedule_map[0].(map[string]interface{})["recurrence"].([]interface{}); ok {
					recurrence := datadogV1.NewMonitorOptionsCustomScheduleRecurrence()
					firstRecurrence := recurrences[0].(map[string]interface{})
					if rrule, ok := firstRecurrence["rrule"].(string); ok {
						recurrence.SetRrule(rrule)
					}
					if start, ok := firstRecurrence["start"].(string); ok && start != "" {
						recurrence.SetStart(start)
					}
					if timezone, ok := firstRecurrence["timezone"].(string); ok {
						recurrence.SetTimezone(timezone)
					}
					newRecurrences := []datadogV1.MonitorOptionsCustomScheduleRecurrence{*recurrence}
					custom_schedule := datadogV1.NewMonitorOptionsCustomSchedule()
					custom_schedule.SetRecurrences(newRecurrences)
					scheduling_options.SetCustomSchedule(*custom_schedule)
				}
			}
			if scheduling_options.HasCustomSchedule() || scheduling_options.HasEvaluationWindow() {
				o.SetSchedulingOptions(*scheduling_options)
			}

		}
	}
	o.SetThresholds(thresholds)
	o.SetIncludeTags(d.Get("include_tags").(bool))
	if !hasCustomSchedule {
		o.SetNotifyNoData(d.Get("notify_no_data").(bool))
		o.SetRequireFullWindow(d.Get("require_full_window").(bool))
	} else {
		// this has to be done explicitly to override the default
		o.SetRequireFullWindow(false)
	}

	if thresholdWindows.HasRecoveryWindow() || thresholdWindows.HasTriggerWindow() {
		o.SetThresholdWindows(thresholdWindows)
	}

	if attr, ok := d.GetOk("group_retention_duration"); ok {
		o.SetGroupRetentionDuration(attr.(string))
	}
	if attr, ok := d.GetOk("new_group_delay"); ok {
		o.SetNewGroupDelay(int64(attr.(int)))
	}
	// Don't check with GetOk, doesn't work with 0 (we can't do the same for
	// new_group_delay because it would always override new_host_delay).
	o.SetNewHostDelay(int64(d.Get("new_host_delay").(int)))
	if attr, ok := d.GetOk("evaluation_delay"); ok {
		o.SetEvaluationDelay(int64(attr.(int)))
	}
	attr, onMissingDataOk := d.GetOk("on_missing_data")
	if onMissingDataOk {
		o.SetOnMissingData(datadogV1.OnMissingDataOption(attr.(string)))
	}
	// no_data_timeframe cannot be combined with on_missing_data. This provider
	// defaults no_data_timeframe to 10, so we need this extra logic to exclude
	// no_data_timeframe from the monitor definition when on_missing_data is set.
	if attr, ok := d.GetOk("no_data_timeframe"); ok && !onMissingDataOk && !hasCustomSchedule {
		o.SetNoDataTimeframe(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("renotify_interval"); ok {
		o.SetRenotifyInterval(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("renotify_occurrences"); ok {
		o.SetRenotifyOccurrences(int64(attr.(int)))
	}
	renotify_statuses := make([]datadogV1.MonitorRenotifyStatusType, 0)
	if attr, ok := d.GetOk("renotify_statuses"); ok {
		for _, s := range attr.(*schema.Set).List() {
			renotify_statuses = append(renotify_statuses, datadogV1.MonitorRenotifyStatusType(s.(string)))
		}
		o.SetRenotifyStatuses(renotify_statuses)
	} else {
		o.SetRenotifyStatuses(nil)
	}
	if attr, ok := d.GetOk("notify_audit"); ok {
		o.SetNotifyAudit(attr.(bool))
	}
	if attr, ok := d.GetOk("timeout_h"); ok {
		o.SetTimeoutH(int64(attr.(int)))
	}
	if attr, ok := d.GetOk("escalation_message"); ok {
		o.SetEscalationMessage(attr.(string))
	}

	if v, ok := d.GetOk("variables"); ok {
		variables := v.([]interface{})
		if len(variables) > 0 {
			// we always have either zero or one
			for _, v := range variables {
				if v == nil {
					// Empty `variables {}` block (e.g. produced by a dynamic block
					// whose inner content is itself empty) shows up as a nil
					// element. Skip it instead of panicking on the type assertion.
					continue
				}
				m, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				var monitorVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition
				if query, ok := m["event_query"]; ok {
					queries, _ := query.([]interface{})
					for _, q := range queries {
						if q == nil {
							continue // Skip nil query entries
						}
						queryMap, ok := q.(map[string]interface{})
						if !ok {
							continue
						}
						monitorVariables = append(monitorVariables, *buildMonitorFormulaAndFunctionEventQuery(queryMap))
					}
				}
				if query, ok := m["cloud_cost_query"]; ok {
					queries, _ := query.([]interface{})
					for _, q := range queries {
						if q == nil {
							continue // Skip nil query entries
						}
						queryMap, ok := q.(map[string]interface{})
						if !ok {
							continue
						}
						monitorVariables = append(monitorVariables, *buildMonitorFormulaAndFunctionCloudCostQuery(queryMap))
					}
				}
				if query, ok := m["data_quality_query"]; ok {
					queries, ok := query.([]interface{})
					if !ok {
						panic("variables.data_quality_query: expected a list but got invalid type")
					}
					for i, q := range queries {
						if q == nil {
							continue // Skip nil query entries
						}
						queryMap, ok := q.(map[string]interface{})
						if !ok {
							panic(fmt.Sprintf("variables.data_quality_query[%d]: expected a map/object but got invalid type", i))
						}
						monitorVariables = append(monitorVariables, *buildMonitorFormulaAndFunctionDataQualityQuery(queryMap))
					}
				}
				if query, ok := m["aggregate_augmented_query"]; ok {
					queries := query.([]interface{})
					for _, q := range queries {
						if q == nil {
							continue
						}
						queryMap, ok := q.(map[string]interface{})
						if !ok {
							panic("variables.aggregate_augmented_query: expected a map/object but got invalid type")
						}
						monitorVariables = append(monitorVariables, *buildMonitorFormulaAndFunctionAggregateAugmentedQuery(queryMap))
					}
				}
				if query, ok := m["aggregate_filtered_query"]; ok {
					queries := query.([]interface{})
					for _, q := range queries {
						if q == nil {
							continue
						}
						queryMap, ok := q.(map[string]interface{})
						if !ok {
							panic("variables.aggregate_filtered_query: expected a map/object but got invalid type")
						}
						monitorVariables = append(monitorVariables, *buildMonitorFormulaAndFunctionAggregateFilteredQuery(queryMap))
					}
				}
				if query, ok := m["data_jobs_query"]; ok {
					queries, ok := query.([]interface{})
					if !ok {
						panic("variables.data_jobs_query: expected a list but got invalid type")
					}
					for i, q := range queries {
						if q == nil {
							continue
						}
						queryMap, ok := q.(map[string]interface{})
						if !ok {
							panic(fmt.Sprintf("variables.data_jobs_query[%d]: expected a map/object but got invalid type", i))
						}
						monitorVariables = append(monitorVariables, *buildMonitorFormulaAndFunctionDataJobsQuery(queryMap))
					}
				}
				o.SetVariables(monitorVariables)
			}
		}
	}

	monitorType := datadogV1.MonitorType(d.Get("type").(string))
	if monitorType == datadogV1.MONITORTYPE_LOG_ALERT {
		if attr, ok := d.GetOk("enable_logs_sample"); ok {
			o.SetEnableLogsSample(attr.(bool))
		} else {
			o.SetEnableLogsSample(false)
		}

		if attr, ok := d.GetOk("groupby_simple_monitor"); ok {
			o.SetGroupbySimpleMonitor(attr.(bool))
		}
	}

	if monitorType == datadogV1.MONITORTYPE_CI_PIPELINES_ALERT || monitorType == datadogV1.MONITORTYPE_CI_TESTS_ALERT {
		if attr, ok := d.GetOk("enable_samples"); ok {
			o.SetEnableSamples(attr.(bool))
		} else {
			o.SetEnableSamples(false)
		}
	}

	if attr, ok := d.GetOk("notify_by"); ok {
		notifyBy := make([]string, 0)
		for _, s := range attr.(*schema.Set).List() {
			notifyBy = append(notifyBy, s.(string))
		}
		sort.Strings(notifyBy)
		o.SetNotifyBy(notifyBy)
	}

	if attr, ok := d.GetOk("notification_preset_name"); ok {
		o.SetNotificationPresetName(datadogV1.MonitorOptionsNotificationPresets(attr.(string)))
	}

	query := d.Get("query").(string)

	m := datadogV1.NewMonitor(query, monitorType)
	m.SetName(d.Get("name").(string))
	m.SetMessage(d.Get("message").(string))
	m.SetOptions(o)

	u := datadogV1.NewMonitorUpdateRequest()
	u.SetType(monitorType)
	u.SetQuery(query)
	u.SetName(d.Get("name").(string))
	u.SetMessage(d.Get("message").(string))
	u.SetOptions(o)

	if draftStatus, ok := d.GetOk("draft_status"); ok {
		ds := datadogV1.MonitorDraftStatus(draftStatus.(string))
		m.SetDraftStatus(ds)
		u.SetDraftStatus(ds)
	}

	if attr, ok := d.GetOk("priority"); ok {
		x, _ := strconv.ParseInt(attr.(string), 10, 64)
		m.SetPriority(x)
		u.SetPriority(x)
	} else {
		m.SetPriorityNil()
		u.SetPriorityNil()
	}

	var roles []string
	if attr, ok := d.GetOk("restricted_roles"); ok {
		for _, r := range attr.(*schema.Set).List() {
			roles = append(roles, r.(string))
		}
		sort.Strings(roles)
	}
	m.SetRestrictedRoles(roles)
	u.SetRestrictedRoles(roles)

	tags := make([]string, 0)
	if attr, ok := d.GetOk("tags"); ok {
		for _, s := range attr.(*schema.Set).List() {
			tags = append(tags, s.(string))
		}
		sort.Strings(tags)
	}
	m.SetTags(tags)
	u.SetTags(tags)

	// Assets
	if attr, ok := d.GetOk("assets"); ok {
		tfAssets := attr.([]interface{})
		assets := buildMonitorAssets(tfAssets)
		if len(assets) > 0 {
			m.SetAssets(assets)
			u.SetAssets(assets)
		}
	}

	return m, u
}

func buildEventQueryGroupBysFromTerraform(terraformGroupBys []interface{}) []datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy {
	if len(terraformGroupBys) == 0 {
		return make([]datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy, 0)
	}
	datadogGroupBys := make([]datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy, len(terraformGroupBys))
	for i, g := range terraformGroupBys {
		groupBy := g.(map[string]interface{})
		datadogGroupBy := datadogV1.NewMonitorFormulaAndFunctionEventQueryGroupBy(groupBy["facet"].(string))
		if src, ok := groupBy["source"].(string); ok && src != "" {
			datadogGroupBy.SetSource(src)
		}
		if v, ok := groupBy["limit"].(int); ok && v != 0 {
			datadogGroupBy.SetLimit(int64(v))
		}
		if v, ok := groupBy["sort"].([]interface{}); ok && len(v) > 0 {
			if v, ok := v[0].(map[string]interface{}); ok && len(v) > 0 {
				sortMap := &datadogV1.MonitorFormulaAndFunctionEventQueryGroupBySort{}
				if aggr, ok := v["aggregation"].(string); ok && len(aggr) > 0 {
					aggregation := datadogV1.MonitorFormulaAndFunctionEventAggregation(v["aggregation"].(string))
					sortMap.SetAggregation(aggregation)
				}
				if order, ok := v["order"].(string); ok && len(order) > 0 {
					eventSort := datadogV1.QuerySortOrder(order)
					sortMap.SetOrder(eventSort)
				}
				if metric, ok := v["metric"].(string); ok && len(metric) > 0 {
					sortMap.SetMetric(metric)
				}
				datadogGroupBy.SetSort(*sortMap)
			}
		}
		datadogGroupBys[i] = *datadogGroupBy
	}
	return datadogGroupBys
}

func buildEventQueryComputesFromTerraform(computeList []interface{}) []datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionCompute {
	out := make([]datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionCompute, 0, len(computeList))
	for _, c := range computeList {
		computeMap := c.(map[string]interface{})
		aggregation := datadogV1.MonitorFormulaAndFunctionEventAggregation(computeMap["aggregation"].(string))
		compute := datadogV1.NewMonitorFormulaAndFunctionEventQueryDefinitionCompute(aggregation)
		if interval, ok := computeMap["interval"].(int); ok && interval != 0 {
			compute.SetInterval(int64(interval))
		}
		if metric, ok := computeMap["metric"].(string); ok && len(metric) > 0 {
			compute.SetMetric(metric)
		}
		if name, ok := computeMap["name"].(string); ok && len(name) > 0 {
			compute.SetName(name)
		}
		out = append(out, *compute)
	}
	return out
}

func buildMonitorReferenceTableQueryDefinitionFromMap(m map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionReferenceTableQueryDefinition {
	ds := datadogV1.MonitorFormulaAndFunctionReferenceTableDataSource(m["data_source"].(string))
	ref := datadogV1.NewMonitorFormulaAndFunctionReferenceTableQueryDefinition(ds, m["table_name"].(string))
	if n, ok := m["name"].(string); ok && n != "" {
		ref.SetName(n)
	}
	if qf, ok := m["query_filter"].(string); ok && qf != "" {
		ref.SetQueryFilter(qf)
	}
	if cols, ok := m["columns"].([]interface{}); ok && len(cols) > 0 {
		outCols := make([]datadogV1.MonitorFormulaAndFunctionReferenceTableColumn, 0, len(cols))
		for _, c := range cols {
			if c == nil {
				continue
			}
			cm := c.(map[string]interface{})
			col := datadogV1.NewMonitorFormulaAndFunctionReferenceTableColumn(cm["name"].(string))
			if alias, ok := cm["alias"].(string); ok && alias != "" {
				col.SetAlias(alias)
			}
			outCols = append(outCols, *col)
		}
		if len(outCols) > 0 {
			ref.SetColumns(outCols)
		}
	}
	return ref
}

func buildAggregateBaseQueryFromTerraformMap(data map[string]interface{}, errCtx string) datadogV1.MonitorFormulaAndFunctionAggregateBaseQuery {
	if bmq, ok := data["base_metrics_query"].([]interface{}); ok && len(bmq) > 0 && bmq[0] != nil {
		m := bmq[0].(map[string]interface{})
		mq := datadogV1.NewMonitorFormulaAndFunctionMetricsQueryDefinition(
			datadogV1.MonitorFormulaAndFunctionMetricsDataSource(m["data_source"].(string)),
			m["query"].(string),
		)
		if n, ok := m["name"].(string); ok && n != "" {
			mq.SetName(n)
		}
		if ag, ok := m["aggregator"].(string); ok && ag != "" {
			mq.SetAggregator(datadogV1.MonitorFormulaAndFunctionMetricsAggregator(ag))
		}
		return datadogV1.MonitorFormulaAndFunctionMetricsQueryDefinitionAsMonitorFormulaAndFunctionAggregateBaseQuery(mq)
	}
	if beq, ok := data["base_event_query"].([]interface{}); ok && len(beq) > 0 && beq[0] != nil {
		ev := buildMonitorFormulaAndFunctionEventQuery(beq[0].(map[string]interface{}))
		return datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionAsMonitorFormulaAndFunctionAggregateBaseQuery(ev.MonitorFormulaAndFunctionEventQueryDefinition)
	}
	panic(errCtx + ": set either base_metrics_query or base_event_query")
}

func buildMonitorFormulaAndFunctionAggregateAugmentedQuery(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	var name string
	if v, ok := data["name"].(string); ok {
		name = v
	}
	dataSource := datadogV1.MonitorFormulaAndFunctionAggregateAugmentedDataSource(data["data_source"].(string))

	var augment datadogV1.MonitorFormulaAndFunctionAggregateAugmentQuery
	if art, ok := data["augment_reference_table"].([]interface{}); ok && len(art) > 0 && art[0] != nil {
		ref := buildMonitorReferenceTableQueryDefinitionFromMap(art[0].(map[string]interface{}))
		augment = datadogV1.MonitorFormulaAndFunctionReferenceTableQueryDefinitionAsMonitorFormulaAndFunctionAggregateAugmentQuery(ref)
	} else if aeq, ok := data["augment_event_query"].([]interface{}); ok && len(aeq) > 0 && aeq[0] != nil {
		ev := buildMonitorFormulaAndFunctionEventQuery(aeq[0].(map[string]interface{}))
		augment = datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionAsMonitorFormulaAndFunctionAggregateAugmentQuery(ev.MonitorFormulaAndFunctionEventQueryDefinition)
	} else {
		panic("aggregate_augmented_query: set either augment_reference_table or augment_event_query")
	}

	base := buildAggregateBaseQueryFromTerraformMap(data, "aggregate_augmented_query")

	jcList := data["join_condition"].([]interface{})
	jc := jcList[0].(map[string]interface{})
	join := datadogV1.NewMonitorFormulaAndFunctionAggregateQueryJoinCondition(
		jc["augment_attribute"].(string),
		jc["base_attribute"].(string),
		datadogV1.MonitorFormulaAndFunctionAggregateQueryJoinType(jc["join_type"].(string)),
	)

	computeList := data["compute"].([]interface{})
	computes := buildEventQueryComputesFromTerraform(computeList)

	var groupBy []datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy
	if terraformGroupBys, ok := data["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		groupBy = buildEventQueryGroupBysFromTerraform(terraformGroupBys)
	}

	def := datadogV1.NewMonitorFormulaAndFunctionAggregateAugmentedQueryDefinition(
		augment,
		base,
		computes,
		dataSource,
		groupBy,
		*join,
	)
	if name != "" {
		def.SetName(name)
	}
	out := datadogV1.MonitorFormulaAndFunctionAggregateAugmentedQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(def)
	return &out
}

func buildMonitorFormulaAndFunctionAggregateFilteredQuery(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	var name string
	if v, ok := data["name"].(string); ok {
		name = v
	}
	dataSource := datadogV1.MonitorFormulaAndFunctionAggregateFilteredDataSource(data["data_source"].(string))

	var filterQuery datadogV1.MonitorFormulaAndFunctionAggregateFilterQuery
	if frt, ok := data["filter_reference_table"].([]interface{}); ok && len(frt) > 0 && frt[0] != nil {
		ref := buildMonitorReferenceTableQueryDefinitionFromMap(frt[0].(map[string]interface{}))
		filterQuery = datadogV1.MonitorFormulaAndFunctionReferenceTableQueryDefinitionAsMonitorFormulaAndFunctionAggregateFilterQuery(ref)
	} else if feq, ok := data["filter_event_query"].([]interface{}); ok && len(feq) > 0 && feq[0] != nil {
		ev := buildMonitorFormulaAndFunctionEventQuery(feq[0].(map[string]interface{}))
		filterQuery = datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionAsMonitorFormulaAndFunctionAggregateFilterQuery(ev.MonitorFormulaAndFunctionEventQueryDefinition)
	} else {
		panic("aggregate_filtered_query: set either filter_reference_table or filter_event_query")
	}

	base := buildAggregateBaseQueryFromTerraformMap(data, "aggregate_filtered_query")

	filtersList := data["filters"].([]interface{})
	filters := make([]datadogV1.MonitorFormulaAndFunctionAggregateQueryFilter, 0, len(filtersList))
	for _, f := range filtersList {
		if f == nil {
			continue
		}
		fm := f.(map[string]interface{})
		fl := datadogV1.NewMonitorFormulaAndFunctionAggregateQueryFilter(
			fm["base_attribute"].(string),
			fm["filter_attribute"].(string),
		)
		if ex, ok := fm["exclude"].(bool); ok && ex {
			fl.SetExclude(ex)
		}
		filters = append(filters, *fl)
	}

	def := datadogV1.NewMonitorFormulaAndFunctionAggregateFilteredQueryDefinition(
		base,
		dataSource,
		filterQuery,
		filters,
	)
	if name != "" {
		def.SetName(name)
	}
	if computeList, ok := data["compute"].([]interface{}); ok && len(computeList) > 0 {
		def.SetCompute(buildEventQueryComputesFromTerraform(computeList))
	}
	if terraformGroupBys, ok := data["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		def.SetGroupBy(buildEventQueryGroupBysFromTerraform(terraformGroupBys))
	}
	out := datadogV1.MonitorFormulaAndFunctionAggregateFilteredQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(def)
	return &out
}

func buildMonitorFormulaAndFunctionEventQuery(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.MonitorFormulaAndFunctionEventsDataSource(data["data_source"].(string))
	computeList := data["compute"].([]interface{})
	computeMap := computeList[0].(map[string]interface{})
	aggregation := datadogV1.MonitorFormulaAndFunctionEventAggregation(computeMap["aggregation"].(string))
	compute := datadogV1.NewMonitorFormulaAndFunctionEventQueryDefinitionCompute(aggregation)
	if interval, ok := computeMap["interval"].(int); ok && interval != 0 {
		compute.SetInterval(int64(interval))
	}
	if metric, ok := computeMap["metric"].(string); ok && len(metric) > 0 {
		compute.SetMetric(metric)
	}
	if name, ok := computeMap["name"].(string); ok && len(name) > 0 {
		compute.SetName(name)
	}
	eventQuery := datadogV1.NewMonitorFormulaAndFunctionEventQueryDefinition(*compute, dataSource, data["name"].(string))
	eventQueryIndexes := data["indexes"].([]interface{})
	indexes := make([]string, len(eventQueryIndexes))
	for i, index := range eventQueryIndexes {
		indexes[i] = index.(string)
	}
	eventQuery.SetIndexes(indexes)

	if terraformSearches, ok := data["search"].([]interface{}); ok && len(terraformSearches) > 0 {
		terraformSearch := terraformSearches[0].(map[string]interface{})
		eventQuery.Search = datadogV1.NewMonitorFormulaAndFunctionEventQueryDefinitionSearch(terraformSearch["query"].(string))
	}

	if terraformGroupBys, ok := data["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		eventQuery.SetGroupBy(buildEventQueryGroupBysFromTerraform(terraformGroupBys))
	} else {
		eventQuery.SetGroupBy(make([]datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy, 0))
	}

	definition := datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(eventQuery)
	return &definition
}

func buildMonitorFormulaAndFunctionCloudCostQuery(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	dataSource := datadogV1.MonitorFormulaAndFunctionCostDataSource(data["data_source"].(string))

	cloudCostQuery := datadogV1.NewMonitorFormulaAndFunctionCostQueryDefinition(dataSource, data["name"].(string), data["query"].(string))

	if v, ok := data["aggregator"].(string); ok && len(v) != 0 {
		cloudCostQuery.SetAggregator(datadogV1.MonitorFormulaAndFunctionCostAggregator(v))
	}

	datadogV1.MonitorFormulaAndFunctionCostQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(cloudCostQuery)

	definition := datadogV1.MonitorFormulaAndFunctionCostQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(cloudCostQuery)
	return &definition
}

// getRequiredString safely extracts a required string field from a map with helpful error messages
func getRequiredString(data map[string]interface{}, fieldName, context string) string {
	val, ok := data[fieldName].(string)
	if !ok || val == "" {
		panic(fmt.Sprintf("%s: '%s' is required and must be a non-empty string", context, fieldName))
	}
	return val
}

// getOptionalString safely extracts an optional string field from a map
func getOptionalString(data map[string]interface{}, fieldName string) (string, bool) {
	val, ok := data[fieldName].(string)
	return val, ok && len(val) > 0
}

// getOptionalStringSlice safely extracts an optional string slice from a map, filtering out nil and empty values
func getOptionalStringSlice(data map[string]interface{}, fieldName string) []string {
	list, ok := data[fieldName].([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}
	result := make([]string, 0, len(list))
	for _, item := range list {
		if item == nil {
			continue
		}
		if strVal, ok := item.(string); ok && strVal != "" {
			result = append(result, strVal)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// buildDataQualityMonitorOptions builds monitor options from a map, returning nil if no options are set
func buildDataQualityMonitorOptions(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionDataQualityMonitorOptions {
	opts := datadogV1.NewMonitorFormulaAndFunctionDataQualityMonitorOptions()
	hasAnyOption := false

	if v, ok := getOptionalString(data, "custom_sql"); ok {
		opts.SetCustomSql(v)
		hasAnyOption = true
	}
	if v, ok := getOptionalString(data, "custom_where"); ok {
		opts.SetCustomWhere(v)
		hasAnyOption = true
	}
	if cols := getOptionalStringSlice(data, "group_by_columns"); cols != nil {
		opts.SetGroupByColumns(cols)
		hasAnyOption = true
	}
	if v, ok := getOptionalString(data, "crontab_override"); ok {
		opts.SetCrontabOverride(v)
		hasAnyOption = true
	}
	if v, ok := getOptionalString(data, "model_type_override"); ok {
		opts.SetModelTypeOverride(datadogV1.MonitorFormulaAndFunctionDataQualityModelTypeOverride(v))
		hasAnyOption = true
	}

	if !hasAnyOption {
		return nil
	}
	return opts
}

func buildMonitorFormulaAndFunctionDataQualityQuery(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	// Validate required fields with helpful error messages
	dataSourceVal := getRequiredString(data, "data_source", "data_quality_query")
	dataSource := datadogV1.MonitorFormulaAndFunctionDataQualityDataSource(dataSourceVal)

	measure := getRequiredString(data, "measure", "data_quality_query")
	name := getRequiredString(data, "name", "data_quality_query")
	filter := getRequiredString(data, "filter", "data_quality_query")

	dataQualityQuery := datadogV1.NewMonitorFormulaAndFunctionDataQualityQueryDefinition(dataSource, filter, measure, name)

	// Optional fields
	if v, ok := getOptionalString(data, "schema_version"); ok {
		dataQualityQuery.SetSchemaVersion(v)
	}

	if v, ok := getOptionalString(data, "scope"); ok {
		dataQualityQuery.SetScope(v)
	}

	// Group by - handle nil and empty arrays safely
	if groupBys := getOptionalStringSlice(data, "group_by"); groupBys != nil {
		dataQualityQuery.SetGroupBy(groupBys)
	}

	// Monitor options - handle nil, empty arrays, and empty maps safely
	if monitorOptionsList, ok := data["monitor_options"].([]interface{}); ok && len(monitorOptionsList) > 0 && monitorOptionsList[0] != nil {
		monitorOptionsData, ok := monitorOptionsList[0].(map[string]interface{})
		if !ok {
			panic("data_quality_query.monitor_options: expected a map/object but got invalid type. Ensure monitor_options is properly formatted as a block.")
		}

		if opts := buildDataQualityMonitorOptions(monitorOptionsData); opts != nil {
			dataQualityQuery.SetMonitorOptions(*opts)
		}
	}

	definition := datadogV1.MonitorFormulaAndFunctionDataQualityQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(dataQualityQuery)
	return &definition
}

func buildMonitorFormulaAndFunctionDataJobsQuery(data map[string]interface{}) *datadogV1.MonitorFormulaAndFunctionQueryDefinition {
	name := getRequiredString(data, "name", "data_jobs_query")
	jobsQuery := getRequiredString(data, "jobs_query", "data_jobs_query")
	jobType := getRequiredString(data, "job_type", "data_jobs_query")
	queryDialect := getRequiredString(data, "query_dialect", "data_jobs_query")

	dataJobsQuery := datadogV1.NewMonitorFormulaAndFunctionDataJobsQueryDefinition(jobType, jobsQuery, name, queryDialect)
	definition := datadogV1.MonitorFormulaAndFunctionDataJobsQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(dataJobsQuery)
	return &definition
}

// Use CustomizeDiff to do monitor validation
func resourceDatadogMonitorCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if _, ok := diff.GetOk("query"); !ok {
		// If "query" depends on other resources, we can't validate as the variables may not be interpolated yet.
		return nil
	}
	if _, ok := diff.GetOk("type"); !ok {
		// Same for type
		return nil
	}
	if validate, ok := diff.GetOkExists("validate"); ok && !validate.(bool) {
		// Explicitly skip validation
		return nil
	}
	m, _ := buildMonitorStruct(diff)

	hasID := false
	id, err := strconv.ParseInt(diff.Id(), 10, 64)
	if err == nil {
		hasID = true
	}

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	return retry.RetryContext(ctx, retryTimeout, func() *retry.RetryError {
		var httpresp *http.Response
		if hasID {
			_, httpresp, err = apiInstances.GetMonitorsApiV1().ValidateExistingMonitor(auth, id, *m)
		} else {
			_, httpresp, err = apiInstances.GetMonitorsApiV1().ValidateMonitor(auth, *m)
		}
		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 502 || httpresp.StatusCode == 504) {
				return retry.RetryableError(utils.TranslateClientError(err, httpresp, "error validating monitor, retrying"))
			}
			return retry.NonRetryableError(utils.TranslateClientError(err, httpresp, "error validating monitor"))
		}
		return nil
	})
}

func resourceDatadogMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	m, _ := buildMonitorStruct(d)
	mCreated, httpResponse, err := apiInstances.GetMonitorsApiV1().CreateMonitor(auth, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating monitor")
	}
	if err := utils.CheckForUnparsed(m); err != nil {
		return diag.FromErr(err)
	}
	mCreatedID := strconv.FormatInt(mCreated.GetId(), 10)
	d.SetId(mCreatedID)

	return updateMonitorState(d, meta, &mCreated)
}

func updateMonitorState(d *schema.ResourceData, meta interface{}, m *datadogV1.Monitor) diag.Diagnostics {
	thresholds := make(map[string]string)

	if v, ok := m.Options.Thresholds.GetOkOk(); ok {
		thresholds["ok"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetWarningOk(); ok {
		thresholds["warning"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetCriticalOk(); ok {
		thresholds["critical"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetUnknownOk(); ok {
		thresholds["unknown"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetWarningRecoveryOk(); ok {
		thresholds["warning_recovery"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetCriticalRecoveryOk(); ok {
		thresholds["critical_recovery"] = fmt.Sprintf("%v", *v)
	}
	if v, ok := m.Options.Thresholds.GetCriticalQueryOk(); ok {
		thresholds["critical_query"] = *v
	}
	if v, ok := m.Options.Thresholds.GetCriticalRecoveryQueryOk(); ok {
		thresholds["critical_recovery_query"] = *v
	}

	thresholdWindows := make(map[string]string)
	for k, v := range map[string]string{
		"recovery_window": m.Options.ThresholdWindows.GetRecoveryWindow(),
		"trigger_window":  m.Options.ThresholdWindows.GetTriggerWindow(),
	} {
		if v != "" {
			thresholdWindows[k] = v
		}
	}

	if err := d.Set("name", m.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("message", m.GetMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("query", m.GetQuery()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", m.GetType()); err != nil {
		return diag.FromErr(err)
	}

	if v, ok := m.GetDraftStatusOk(); ok && v != nil {
		if err := d.Set("draft_status", *v); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// Workaround to handle the api response missing the draft_status field when monitor-draft-status-api is not enabled
		if err := d.Set("draft_status", string(datadogV1.MONITORDRAFTSTATUS_PUBLISHED)); err != nil {
			return diag.FromErr(err)
		}
	}

	priorityStr := ""
	priority, _ := m.GetPriorityOk()
	if priority != nil {
		priorityStr = strconv.FormatInt(*priority, 10)
	}
	if err := d.Set("priority", priorityStr); err != nil {
		return diag.FromErr(err)
	}

	if len(thresholds) > 0 {
		if err := d.Set("monitor_thresholds", []interface{}{thresholds}); err != nil {
			return diag.FromErr(err)
		}
	}
	if len(thresholdWindows) > 0 {
		if err := d.Set("monitor_threshold_windows", []interface{}{thresholdWindows}); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("new_group_delay", m.Options.GetNewGroupDelay()); err != nil {
		return diag.FromErr(err)
	}
	if v, ok := m.Options.GetNewHostDelayOk(); ok && v != nil {
		if err := d.Set("new_host_delay", *v); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("evaluation_delay", m.Options.GetEvaluationDelay()); err != nil {
		return diag.FromErr(err)
	}
	if v, ok := m.Options.GetNotifyNoDataOk(); ok {
		if err := d.Set("notify_no_data", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("on_missing_data", m.Options.GetOnMissingData()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("group_retention_duration", m.Options.GetGroupRetentionDuration()); err != nil {
		return diag.FromErr(err)
	}
	if v, ok := m.Options.GetNoDataTimeframeOk(); ok {
		if err := d.Set("no_data_timeframe", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("renotify_interval", m.Options.GetRenotifyInterval()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("renotify_occurrences", m.Options.GetRenotifyOccurrences()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("notify_audit", m.Options.GetNotifyAudit()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("timeout_h", m.Options.GetTimeoutH()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("escalation_message", m.Options.GetEscalationMessage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("include_tags", m.Options.GetIncludeTags()); err != nil {
		return diag.FromErr(err)
	}

	if renotifyStatuses, ok := m.Options.GetRenotifyStatusesOk(); ok && len(*renotifyStatuses) > 0 {
		renotifyStatusesCopy := append([]datadogV1.MonitorRenotifyStatusType{}, *renotifyStatuses...)
		if err := d.Set("renotify_statuses", renotifyStatusesCopy); err != nil {
			return diag.FromErr(err)
		}
	}

	var tags []string
	tags = append(tags, m.GetTags()...)
	sort.Strings(tags)
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("require_full_window", m.Options.GetRequireFullWindow()); err != nil {
		return diag.FromErr(err)
	}

	if restrictedRoles, ok := m.GetRestrictedRolesOk(); ok {
		// This helper function is defined in `resource_datadog_dashboard`
		restrictedRolesCopy := buildTerraformRestrictedRoles(restrictedRoles)
		if err := d.Set("restricted_roles", restrictedRolesCopy); err != nil {
			return diag.FromErr(err)
		}
	}

	if variables, ok := m.Options.GetVariablesOk(); ok && len(*variables) > 0 {
		log.Printf("[INFO] variables: %d, %+v", len(*variables), *variables)
		if m.GetType() == datadogV1.MONITORTYPE_COST_ALERT {
			terraformVariables := buildTerraformCostMonitorVariables(*variables)
			if err := d.Set("variables", terraformVariables); err != nil {
				return diag.FromErr(err)
			}
		} else if m.GetType() == datadogV1.MonitorType("data-quality alert") {
			terraformVariables := buildTerraformDataQualityMonitorVariables(*variables)
			if err := d.Set("variables", terraformVariables); err != nil {
				return diag.FromErr(err)
			}
		} else if m.GetType() == datadogV1.MonitorType("data-jobs alert") {
			terraformVariables := buildTerraformDataJobsMonitorVariables(*variables)
			if err := d.Set("variables", terraformVariables); err != nil {
				return diag.FromErr(err)
			}
		} else {
			terraformVariables := buildTerraformMonitorVariables(*variables)
			if err := d.Set("variables", terraformVariables); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if m.GetType() == datadogV1.MONITORTYPE_LOG_ALERT {
		if err := d.Set("enable_logs_sample", m.Options.GetEnableLogsSample()); err != nil {
			return diag.FromErr(err)
		}
		if err := d.Set("groupby_simple_monitor", m.Options.GetGroupbySimpleMonitor()); err != nil {
			return diag.FromErr(err)
		}
	}

	if m.GetType() == datadogV1.MONITORTYPE_CI_PIPELINES_ALERT || m.GetType() == datadogV1.MONITORTYPE_CI_TESTS_ALERT {
		if err := d.Set("enable_samples", m.Options.GetEnableSamples()); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("notify_by", m.Options.GetNotifyBy()); err != nil {
		return diag.FromErr(err)
	}

	evaluation_window := make(map[string]interface{})
	if e, ok := m.Options.SchedulingOptions.GetEvaluationWindowOk(); ok {
		if d, ok := e.GetDayStartsOk(); ok {
			evaluation_window["day_starts"] = *d
		}
		if h, ok := e.GetHourStartsOk(); ok {
			evaluation_window["hour_starts"] = h
		}
		if m, ok := e.GetMonthStartsOk(); ok {
			evaluation_window["month_starts"] = m
		}
		if timezone, ok := e.GetTimezoneOk(); ok {
			evaluation_window["timezone"] = *timezone
		}
	}
	custom_schedule := make(map[string]interface{})
	if c, ok := m.Options.SchedulingOptions.GetCustomScheduleOk(); ok {
		if recurrences, ok := c.GetRecurrencesOk(); ok && len(*recurrences) > 0 {
			recurrence := make(map[string]interface{})
			r := (*recurrences)[0]
			if rrule, ok := r.GetRruleOk(); ok {
				recurrence["rrule"] = rrule
			}
			if start, ok := r.GetStartOk(); ok {
				recurrence["start"] = start
			}
			if timezone, ok := r.GetTimezoneOk(); ok {
				recurrence["timezone"] = timezone
			}
			value := [](interface{}){recurrence}
			custom_schedule["recurrence"] = value
		}
	}

	scheduling_options := make(map[string]interface{})
	if len(evaluation_window) > 0 {
		scheduling_options["evaluation_window"] = []interface{}{evaluation_window}
	}
	if len(custom_schedule) > 0 {
		scheduling_options["custom_schedule"] = []interface{}{custom_schedule}
	}

	if len(scheduling_options) > 0 {
		if err := d.Set("scheduling_options", []interface{}{scheduling_options}); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("notification_preset_name", m.Options.GetNotificationPresetName()); err != nil {
		return diag.FromErr(err)
	}

	// Assets -> state
	if assets, ok := m.GetAssetsOk(); ok && assets != nil {
		terraformAssets := buildTerraformMonitorAssets(*assets)
		if err := d.Set("assets", terraformAssets); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func terraformEventQueryDefinitionToMap(ev *datadogV1.MonitorFormulaAndFunctionEventQueryDefinition) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	if ev == nil {
		return terraformQuery
	}
	if dataSource, ok := ev.GetDataSourceOk(); ok {
		terraformQuery["data_source"] = dataSource
	}
	if name, ok := ev.GetNameOk(); ok {
		terraformQuery["name"] = name
	}
	if indexes, ok := ev.GetIndexesOk(); ok {
		terraformQuery["indexes"] = indexes
	}
	if search, ok := ev.GetSearchOk(); ok {
		terraformSearch := map[string]interface{}{"query": search.GetQuery()}
		terraformQuery["search"] = []map[string]interface{}{terraformSearch}
	}
	if compute, ok := ev.GetComputeOk(); ok {
		terraformCompute := map[string]interface{}{}
		if aggregation, ok := compute.GetAggregationOk(); ok {
			terraformCompute["aggregation"] = aggregation
		}
		if interval, ok := compute.GetIntervalOk(); ok {
			terraformCompute["interval"] = interval
		}
		if metric, ok := compute.GetMetricOk(); ok {
			terraformCompute["metric"] = metric
		}
		if name, ok := compute.GetNameOk(); ok {
			terraformCompute["name"] = name
		}
		terraformQuery["compute"] = []map[string]interface{}{terraformCompute}
	}
	if groups, ok := ev.GetGroupByOk(); ok {
		terraformGroupBys := make([]map[string]interface{}, len(*groups))
		for i, groupBy := range *groups {
			terraformGroupBy := map[string]interface{}{
				"facet": groupBy.GetFacet(),
			}
			if s, ok := groupBy.GetSourceOk(); ok && s != nil {
				terraformGroupBy["source"] = *s
			}
			if v, ok := groupBy.GetLimitOk(); ok {
				terraformGroupBy["limit"] = *v
			}
			if v, ok := groupBy.GetSortOk(); ok {
				terraformSort := map[string]interface{}{}
				if metric, ok := v.GetMetricOk(); ok {
					terraformSort["metric"] = metric
				}
				if order, ok := v.GetOrderOk(); ok {
					terraformSort["order"] = order
				}
				if aggregation, ok := v.GetAggregationOk(); ok {
					terraformSort["aggregation"] = aggregation
				}
				terraformGroupBy["sort"] = []map[string]interface{}{terraformSort}
			}
			terraformGroupBys[i] = terraformGroupBy
		}
		terraformQuery["group_by"] = terraformGroupBys
	}
	return terraformQuery
}

func terraformReferenceTableQueryDefinitionToMap(ref *datadogV1.MonitorFormulaAndFunctionReferenceTableQueryDefinition) map[string]interface{} {
	if ref == nil {
		return map[string]interface{}{}
	}
	m := map[string]interface{}{
		"data_source": ref.GetDataSource(),
		"table_name":  ref.GetTableName(),
	}
	if n, ok := ref.GetNameOk(); ok {
		m["name"] = n
	}
	if qf, ok := ref.GetQueryFilterOk(); ok {
		m["query_filter"] = qf
	}
	if cols, ok := ref.GetColumnsOk(); ok && len(*cols) > 0 {
		colList := make([]map[string]interface{}, 0, len(*cols))
		for _, c := range *cols {
			cm := map[string]interface{}{"name": c.GetName()}
			if alias, ok := c.GetAliasOk(); ok {
				cm["alias"] = alias
			}
			colList = append(colList, cm)
		}
		m["columns"] = colList
	}
	return m
}

func terraformAggregateAugmentedDefinitionToMap(def *datadogV1.MonitorFormulaAndFunctionAggregateAugmentedQueryDefinition) map[string]interface{} {
	out := map[string]interface{}{}
	if def == nil {
		return out
	}
	if n, ok := def.GetNameOk(); ok {
		out["name"] = *n
	}
	out["data_source"] = def.GetDataSource()

	aq := def.GetAugmentQuery()
	if aq.MonitorFormulaAndFunctionReferenceTableQueryDefinition != nil {
		out["augment_reference_table"] = []interface{}{terraformReferenceTableQueryDefinitionToMap(aq.MonitorFormulaAndFunctionReferenceTableQueryDefinition)}
	} else if aq.MonitorFormulaAndFunctionEventQueryDefinition != nil {
		out["augment_event_query"] = []interface{}{terraformEventQueryDefinitionToMap(aq.MonitorFormulaAndFunctionEventQueryDefinition)}
	}

	bq := def.GetBaseQuery()
	if bq.MonitorFormulaAndFunctionMetricsQueryDefinition != nil {
		mq := bq.MonitorFormulaAndFunctionMetricsQueryDefinition
		m := map[string]interface{}{
			"data_source": mq.GetDataSource(),
			"query":       mq.GetQuery(),
		}
		if n, ok := mq.GetNameOk(); ok {
			m["name"] = n
		}
		if ag, ok := mq.GetAggregatorOk(); ok {
			m["aggregator"] = ag
		}
		out["base_metrics_query"] = []interface{}{m}
	} else if bq.MonitorFormulaAndFunctionEventQueryDefinition != nil {
		out["base_event_query"] = []interface{}{terraformEventQueryDefinitionToMap(bq.MonitorFormulaAndFunctionEventQueryDefinition)}
	}

	jc := def.GetJoinCondition()
	out["join_condition"] = []map[string]interface{}{{
		"augment_attribute": jc.GetAugmentAttribute(),
		"base_attribute":    jc.GetBaseAttribute(),
		"join_type":         jc.GetJoinType(),
	}}

	if computes, ok := def.GetComputeOk(); ok && len(*computes) > 0 {
		tfComputes := make([]map[string]interface{}, 0, len(*computes))
		for _, c := range *computes {
			cm := map[string]interface{}{"aggregation": c.GetAggregation()}
			if interval, ok := c.GetIntervalOk(); ok {
				cm["interval"] = interval
			}
			if metric, ok := c.GetMetricOk(); ok {
				cm["metric"] = metric
			}
			if n, ok := c.GetNameOk(); ok {
				cm["name"] = n
			}
			tfComputes = append(tfComputes, cm)
		}
		out["compute"] = tfComputes
	}

	if groups, ok := def.GetGroupByOk(); ok && len(*groups) > 0 {
		terraformGroupBys := make([]map[string]interface{}, len(*groups))
		for i, groupBy := range *groups {
			tg := map[string]interface{}{"facet": groupBy.GetFacet()}
			if s, ok := groupBy.GetSourceOk(); ok {
				tg["source"] = *s
			}
			if v, ok := groupBy.GetLimitOk(); ok {
				tg["limit"] = *v
			}
			if v, ok := groupBy.GetSortOk(); ok {
				terraformSort := map[string]interface{}{}
				if metric, ok := v.GetMetricOk(); ok {
					terraformSort["metric"] = metric
				}
				if order, ok := v.GetOrderOk(); ok {
					terraformSort["order"] = order
				}
				if aggregation, ok := v.GetAggregationOk(); ok {
					terraformSort["aggregation"] = aggregation
				}
				tg["sort"] = []map[string]interface{}{terraformSort}
			}
			terraformGroupBys[i] = tg
		}
		out["group_by"] = terraformGroupBys
	}

	return out
}

func terraformAggregateFilteredDefinitionToMap(def *datadogV1.MonitorFormulaAndFunctionAggregateFilteredQueryDefinition) map[string]interface{} {
	out := map[string]interface{}{}
	if def == nil {
		return out
	}
	if n, ok := def.GetNameOk(); ok {
		out["name"] = *n
	}
	out["data_source"] = def.GetDataSource()

	fq := def.GetFilterQuery()
	if fq.MonitorFormulaAndFunctionReferenceTableQueryDefinition != nil {
		out["filter_reference_table"] = []interface{}{terraformReferenceTableQueryDefinitionToMap(fq.MonitorFormulaAndFunctionReferenceTableQueryDefinition)}
	} else if fq.MonitorFormulaAndFunctionEventQueryDefinition != nil {
		out["filter_event_query"] = []interface{}{terraformEventQueryDefinitionToMap(fq.MonitorFormulaAndFunctionEventQueryDefinition)}
	}

	bq := def.GetBaseQuery()
	if bq.MonitorFormulaAndFunctionMetricsQueryDefinition != nil {
		mq := bq.MonitorFormulaAndFunctionMetricsQueryDefinition
		m := map[string]interface{}{
			"data_source": mq.GetDataSource(),
			"query":       mq.GetQuery(),
		}
		if n, ok := mq.GetNameOk(); ok {
			m["name"] = n
		}
		if ag, ok := mq.GetAggregatorOk(); ok {
			m["aggregator"] = ag
		}
		out["base_metrics_query"] = []interface{}{m}
	} else if bq.MonitorFormulaAndFunctionEventQueryDefinition != nil {
		out["base_event_query"] = []interface{}{terraformEventQueryDefinitionToMap(bq.MonitorFormulaAndFunctionEventQueryDefinition)}
	}

	filters := def.GetFilters()
	tfFilters := make([]map[string]interface{}, 0, len(filters))
	for _, f := range filters {
		fm := map[string]interface{}{
			"base_attribute":   f.GetBaseAttribute(),
			"filter_attribute": f.GetFilterAttribute(),
		}
		// API may return exclude=false even when omitted on create; only persist true in state.
		if exclude, ok := f.GetExcludeOk(); ok && exclude != nil && *exclude {
			fm["exclude"] = true
		}
		tfFilters = append(tfFilters, fm)
	}
	out["filters"] = tfFilters

	if computes, ok := def.GetComputeOk(); ok && len(*computes) > 0 {
		tfComputes := make([]map[string]interface{}, 0, len(*computes))
		for _, c := range *computes {
			cm := map[string]interface{}{"aggregation": c.GetAggregation()}
			if interval, ok := c.GetIntervalOk(); ok {
				cm["interval"] = interval
			}
			if metric, ok := c.GetMetricOk(); ok {
				cm["metric"] = metric
			}
			if n, ok := c.GetNameOk(); ok {
				cm["name"] = n
			}
			tfComputes = append(tfComputes, cm)
		}
		out["compute"] = tfComputes
	}

	if groups, ok := def.GetGroupByOk(); ok && len(*groups) > 0 {
		terraformGroupBys := make([]map[string]interface{}, len(*groups))
		for i, groupBy := range *groups {
			tg := map[string]interface{}{"facet": groupBy.GetFacet()}
			if groupBy.AdditionalProperties != nil {
				if s, ok := groupBy.AdditionalProperties["source"].(string); ok && s != "" {
					tg["source"] = s
				}
			}
			if v, ok := groupBy.GetLimitOk(); ok {
				tg["limit"] = *v
			}
			if v, ok := groupBy.GetSortOk(); ok {
				terraformSort := map[string]interface{}{}
				if metric, ok := v.GetMetricOk(); ok {
					terraformSort["metric"] = metric
				}
				if order, ok := v.GetOrderOk(); ok {
					terraformSort["order"] = order
				}
				if aggregation, ok := v.GetAggregationOk(); ok {
					terraformSort["aggregation"] = aggregation
				}
				tg["sort"] = []map[string]interface{}{terraformSort}
			}
			terraformGroupBys[i] = tg
		}
		out["group_by"] = terraformGroupBys
	}

	return out
}

func buildTerraformCompositeMonitorVariables(datadogVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition) []map[string]interface{} {
	root := map[string]interface{}{}
	var eventQueries []map[string]interface{}
	var aggAug []map[string]interface{}
	var aggFiltered []map[string]interface{}
	for _, query := range datadogVariables {
		switch {
		case query.MonitorFormulaAndFunctionEventQueryDefinition != nil:
			eventQueries = append(eventQueries, terraformEventQueryDefinitionToMap(query.MonitorFormulaAndFunctionEventQueryDefinition))
		case query.MonitorFormulaAndFunctionAggregateAugmentedQueryDefinition != nil:
			aggAug = append(aggAug, terraformAggregateAugmentedDefinitionToMap(query.MonitorFormulaAndFunctionAggregateAugmentedQueryDefinition))
		case query.MonitorFormulaAndFunctionAggregateFilteredQueryDefinition != nil:
			aggFiltered = append(aggFiltered, terraformAggregateFilteredDefinitionToMap(query.MonitorFormulaAndFunctionAggregateFilteredQueryDefinition))
		}
	}
	if len(eventQueries) > 0 {
		root["event_query"] = eventQueries
	}
	if len(aggAug) > 0 {
		root["aggregate_augmented_query"] = aggAug
	}
	if len(aggFiltered) > 0 {
		root["aggregate_filtered_query"] = aggFiltered
	}
	if len(root) == 0 {
		return nil
	}
	terraformVariables := []map[string]interface{}{root}
	log.Printf("[INFO] composite monitor variables: %+v", terraformVariables)
	return terraformVariables
}

func buildTerraformMonitorVariables(datadogVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition) []map[string]interface{} {
	return buildTerraformCompositeMonitorVariables(datadogVariables)
}

func buildTerraformCostMonitorVariables(datadogVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition) []map[string]interface{} {
	queries := make([]map[string]interface{}, len(datadogVariables))
	for i, query := range datadogVariables {
		terraformQuery := map[string]interface{}{}
		terraformCostQueryDefinition := query.MonitorFormulaAndFunctionCostQueryDefinition
		if terraformCostQueryDefinition != nil {
			if dataSource, ok := terraformCostQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if name, ok := terraformCostQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if queryStr, ok := terraformCostQueryDefinition.GetQueryOk(); ok {
				terraformQuery["query"] = queryStr
			}
			if aggregator, ok := terraformCostQueryDefinition.GetAggregatorOk(); ok {
				terraformQuery["aggregator"] = aggregator
			}
			queries[i] = terraformQuery
		}
	}
	terraformVariables := make([]map[string]interface{}, 1)
	terraformVariables[0] = map[string]interface{}{"cloud_cost_query": queries}

	log.Printf("[INFO] queries: %+v", terraformVariables)
	return terraformVariables
}

func buildTerraformDataQualityMonitorVariables(datadogVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition) []map[string]interface{} {
	queries := make([]map[string]interface{}, len(datadogVariables))

	for i, query := range datadogVariables {
		terraformQuery := map[string]interface{}{}
		terraformDataQualityQueryDefinition := query.MonitorFormulaAndFunctionDataQualityQueryDefinition
		if terraformDataQualityQueryDefinition != nil {
			if dataSource, ok := terraformDataQualityQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if name, ok := terraformDataQualityQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if schemaVersion, ok := terraformDataQualityQueryDefinition.GetSchemaVersionOk(); ok {
				terraformQuery["schema_version"] = schemaVersion
			}
			if measure, ok := terraformDataQualityQueryDefinition.GetMeasureOk(); ok {
				terraformQuery["measure"] = measure
			}
			if filter, ok := terraformDataQualityQueryDefinition.GetFilterOk(); ok {
				terraformQuery["filter"] = filter
			}
			if scope, ok := terraformDataQualityQueryDefinition.GetScopeOk(); ok {
				terraformQuery["scope"] = scope
			}
			if groupBy, ok := terraformDataQualityQueryDefinition.GetGroupByOk(); ok {
				terraformQuery["group_by"] = groupBy
			}
			// Monitor options
			if monitorOptions, ok := terraformDataQualityQueryDefinition.GetMonitorOptionsOk(); ok {
				terraformMonitorOptions := map[string]interface{}{}
				if customSql, ok := monitorOptions.GetCustomSqlOk(); ok {
					terraformMonitorOptions["custom_sql"] = customSql
				}
				if customWhere, ok := monitorOptions.GetCustomWhereOk(); ok {
					terraformMonitorOptions["custom_where"] = customWhere
				}
				if groupByCols, ok := monitorOptions.GetGroupByColumnsOk(); ok {
					terraformMonitorOptions["group_by_columns"] = groupByCols
				}
				if crontabOverride, ok := monitorOptions.GetCrontabOverrideOk(); ok {
					terraformMonitorOptions["crontab_override"] = crontabOverride
				}
				if modelTypeOverride, ok := monitorOptions.GetModelTypeOverrideOk(); ok {
					terraformMonitorOptions["model_type_override"] = modelTypeOverride
				}
				terraformQuery["monitor_options"] = []map[string]interface{}{terraformMonitorOptions}
			}
			queries[i] = terraformQuery
		}
	}
	terraformVariables := make([]map[string]interface{}, 1)
	terraformVariables[0] = map[string]interface{}{"data_quality_query": queries}

	log.Printf("[INFO] data_quality_query variables: %+v", terraformVariables)
	return terraformVariables
}

func buildTerraformDataJobsMonitorVariables(datadogVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition) []map[string]interface{} {
	queries := make([]map[string]interface{}, len(datadogVariables))

	for i, query := range datadogVariables {
		terraformQuery := map[string]interface{}{}
		def := query.MonitorFormulaAndFunctionDataJobsQueryDefinition
		if def != nil {
			if name, ok := def.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if jobsQuery, ok := def.GetJobsQueryOk(); ok {
				terraformQuery["jobs_query"] = jobsQuery
			}
			if jobType, ok := def.GetJobTypeOk(); ok {
				terraformQuery["job_type"] = jobType
			}
			if queryDialect, ok := def.GetQueryDialectOk(); ok {
				terraformQuery["query_dialect"] = queryDialect
			}
			queries[i] = terraformQuery
		}
	}
	terraformVariables := make([]map[string]interface{}, 1)
	terraformVariables[0] = map[string]interface{}{"data_jobs_query": queries}

	log.Printf("[INFO] data_jobs_query variables: %+v", terraformVariables)
	return terraformVariables
}

func resourceDatadogMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}
	var (
		m        datadogV1.Monitor
		httpresp *http.Response
	)
	if err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *retry.RetryError {
		m, httpresp, err = apiInstances.GetMonitorsApiV1().GetMonitor(auth, i, *datadogV1.NewGetMonitorOptionalParameters().WithWithAssets(true))
		if err != nil {
			if httpresp != nil {
				if httpresp.StatusCode == 404 {
					d.SetId("")
					return nil
				} else if httpresp.StatusCode == 502 {
					return retry.RetryableError(utils.TranslateClientError(err, httpresp, "error getting monitor, retrying"))
				}
			}
			return retry.NonRetryableError(utils.TranslateClientError(err, httpresp, "error getting monitor"))
		}
		if err := utils.CheckForUnparsed(m); err != nil {
			return retry.NonRetryableError(err)
		}
		return nil
	}); err != nil {
		return diag.FromErr(err)
	}

	if d.Id() == "" {
		return nil
	}

	return updateMonitorState(d, meta, &m)
}

func resourceDatadogMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	_, m := buildMonitorStruct(d)
	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	m.Id = &i

	monitorResp, httpresp, err := apiInstances.GetMonitorsApiV1().UpdateMonitor(auth, i, *m)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating monitor")
	}
	if err := utils.CheckForUnparsed(monitorResp); err != nil {
		return diag.FromErr(err)
	}

	return updateMonitorState(d, meta, &monitorResp)
}

func resourceDatadogMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	i, err := strconv.ParseInt(d.Id(), 10, 64)
	if err != nil {
		return diag.FromErr(err)
	}

	var httpResponse *http.Response

	if d.Get("force_delete").(bool) {
		_, httpResponse, err = apiInstances.GetMonitorsApiV1().DeleteMonitor(auth, i,
			*datadogV1.NewDeleteMonitorOptionalParameters().WithForce("true"))
	} else {
		_, httpResponse, err = apiInstances.GetMonitorsApiV1().DeleteMonitor(auth, i)
	}

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting monitor")
	}

	return nil
}

// Ignore any diff that results from the mix of ints or floats returned from the
// DataDog API.
func suppressDataDogFloatIntDiff(_, old, new string, _ *schema.ResourceData) bool {
	oF, err := strconv.ParseFloat(old, 64)
	if err != nil {
		log.Printf("Error parsing float of old value (%s): %s", old, err)
		return false
	}

	nF, err := strconv.ParseFloat(new, 64)
	if err != nil {
		log.Printf("Error parsing float of new value (%s): %s", new, err)
		return false
	}

	// if the float values of these attributes are equivalent, ignore this
	// diff
	if oF == nF {
		return true
	}
	return false
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// buildMonitorAssets converts Terraform assets into API MonitorAsset slice.
func buildMonitorAssets(tfAssets []interface{}) []datadogV1.MonitorAsset {
	if len(tfAssets) == 0 {
		return nil
	}
	assets := make([]datadogV1.MonitorAsset, 0, len(tfAssets))
	for _, raw := range tfAssets {
		aMap, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		categoryStr, _ := aMap["category"].(string)
		nameStr, _ := aMap["name"].(string)
		urlStr, _ := aMap["url"].(string)
		category := datadogV1.MonitorAssetCategory(categoryStr)
		asset := datadogV1.NewMonitorAsset(category, nameStr, urlStr)
		if rk, ok := aMap["resource_key"].(string); ok && rk != "" {
			asset.SetResourceKey(rk)
		}
		if rt, ok := aMap["resource_type"].(string); ok && rt != "" {
			rtEnum := datadogV1.MonitorAssetResourceType(rt)
			asset.SetResourceType(rtEnum)
		}
		assets = append(assets, *asset)
	}
	return assets
}

// buildTerraformMonitorAssets flattens API assets into Terraform state shape.
func buildTerraformMonitorAssets(apiAssets []datadogV1.MonitorAsset) []map[string]interface{} {
	tfAssets := make([]map[string]interface{}, 0, len(apiAssets))
	for _, a := range apiAssets {
		tf := map[string]interface{}{
			"name":     a.GetName(),
			"url":      a.GetUrl(),
			"category": string(a.GetCategory()),
		}
		if rk, ok := a.GetResourceKeyOk(); ok && rk != nil {
			tf["resource_key"] = *rk
		}
		if rt, ok := a.GetResourceTypeOk(); ok && rt != nil {
			tf["resource_type"] = string(*rt)
		}
		tfAssets = append(tfAssets, tf)
	}
	return tfAssets
}
