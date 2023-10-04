package datadog

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
		CustomizeDiff: resourceDatadogMonitorCustomizeDiff,
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
					Type:        schema.TypeInt,
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
					Description:   "A boolean indicating whether this monitor will notify when data stops reporting. Defaults to `false`.",
					Type:          schema.TypeBool,
					Optional:      true,
					Default:       false,
					ConflictsWith: []string{"on_missing_data"},
				},
				"on_missing_data": {
					Description:   "Controls how groups or monitors are treated if an evaluation does not return any data points. The default option results in different behavior depending on the monitor query type. For monitors using `Count` queries, an empty monitor evaluation is treated as 0 and is compared to the threshold conditions. For monitors using any query type other than `Count`, for example `Gauge`, `Measure`, or `Rate`, the monitor shows the last known status. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors. Valid values are: `show_no_data`, `show_and_notify_no_data`, `resolve`, and `default`.",
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
					Description: "**Deprecated**. See `new_group_delay`. Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non-negative integer. This value is ignored for simple monitors and monitors not grouped by host. Defaults to `300`. The only case when this should be used is to override the default and set `new_host_delay` to zero for monitors grouped by host.",
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
					Description: "The number of minutes before a monitor will notify when data stops reporting. Provider defaults to 10 minutes.\n\nWe recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.",
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
					Description: "A boolean indicating whether this monitor needs a full window of data before it's evaluated. Defaults to `true`. Datadog strongly recommends you set this to `false` for sparse metrics, otherwise some evaluations may be skipped.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"locked": {
					Description:   "A boolean indicating whether changes to this monitor should be restricted to the creator or admins. Defaults to `false`.",
					Type:          schema.TypeBool,
					Optional:      true,
					Deprecated:    "Use `restricted_roles`.",
					ConflictsWith: []string{"restricted_roles"},
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						// if restricted_roles is defined, ignore locked
						if _, ok := d.GetOk("restricted_roles"); ok {
							return true
						}
						return false
					},
				},
				"restricted_roles": {
					Description:   "A list of unique role identifiers to define which roles are allowed to edit the monitor. Editing a monitor includes any updates to the monitor configuration, monitor deletion, and muting of the monitor for any amount of time. Roles unique identifiers can be pulled from the [Roles API](https://docs.datadoghq.com/api/latest/roles/#list-roles) in the `data.id` field.",
					Type:          schema.TypeSet,
					Optional:      true,
					Elem:          &schema.Schema{Type: schema.TypeString},
					ConflictsWith: []string{"locked"},
				},
				"include_tags": {
					Description: "A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title. Defaults to `true`.",
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
					Elem:     &schema.Schema{Type: schema.TypeString},
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
					Computed:    true,
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
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"evaluation_window": {
								Description: "Configuration options for the evaluation window. If `hour_starts` is set, no other fields may be set. Otherwise, `day_starts` and `month_starts` must be set together.",
								Type:        schema.TypeList,
								Required:    true,
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
			}
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
						Schema: map[string]*schema.Schema{
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
											Type:         schema.TypeString,
											ValidateFunc: validation.StringIsNotEmpty,
											Required:     true,
											Description:  "The events search string.",
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

	var thresholdWindows datadogV1.MonitorThresholdWindowOptions

	if r, ok := d.GetOk("monitor_threshold_windows.0.recovery_window"); ok {
		thresholdWindows.SetRecoveryWindow(r.(string))
	}

	if r, ok := d.GetOk("monitor_threshold_windows.0.trigger_window"); ok {
		thresholdWindows.SetTriggerWindow(r.(string))
	}

	o := datadogV1.MonitorOptions{}
	o.SetThresholds(thresholds)
	o.SetNotifyNoData(d.Get("notify_no_data").(bool))
	o.SetRequireFullWindow(d.Get("require_full_window").(bool))
	o.SetIncludeTags(d.Get("include_tags").(bool))

	if thresholdWindows.HasRecoveryWindow() || thresholdWindows.HasTriggerWindow() {
		o.SetThresholdWindows(thresholdWindows)
	}

	if attr, ok := d.GetOk("notify_no_data"); ok {
		o.SetNotifyNoData(attr.(bool))
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
	if attr, ok := d.GetOk("no_data_timeframe"); ok && !onMissingDataOk {
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
	if attr, ok := d.GetOk("locked"); ok {
		o.SetLocked(attr.(bool))
	}
	if v, ok := d.GetOk("variables"); ok {
		variables := v.([]interface{})
		if len(variables) > 0 {
			// we always have either zero or one
			for _, v := range variables {
				m := v.(map[string]interface{})
				queries := m["event_query"].([]interface{})
				monitorVariables := make([]datadogV1.MonitorFormulaAndFunctionQueryDefinition, len(queries))
				for i, q := range queries {
					monitorVariables[i] = *buildMonitorFormulaAndFunctionEventQuery(q.(map[string]interface{}))
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

	if attr, ok := d.GetOk("scheduling_options"); ok {
		scheduling_options_list := attr.([]interface{})
		if scheduling_options_map, ok := scheduling_options_list[0].(map[string]interface{}); ok {
			if evaluation_window_map, ok := scheduling_options_map["evaluation_window"].([]interface{})[0].(map[string]interface{}); ok {
				scheduling_options := datadogV1.NewMonitorOptionsSchedulingOptions()
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
				scheduling_options.SetEvaluationWindow(*evaluation_window)
				o.SetSchedulingOptions(*scheduling_options)
			}
		}
	}

	if attr, ok := d.GetOk("notification_preset_name"); ok {
		o.SetNotificationPresetName(datadogV1.MonitorOptionsNotificationPresets(attr.(string)))
	}

	m := datadogV1.NewMonitor(d.Get("query").(string), monitorType)
	m.SetName(d.Get("name").(string))
	m.SetMessage(d.Get("message").(string))
	m.SetPriority(int64(d.Get("priority").(int)))
	m.SetOptions(o)

	u := datadogV1.NewMonitorUpdateRequest()
	u.SetType(monitorType)
	u.SetQuery(d.Get("query").(string))
	u.SetName(d.Get("name").(string))
	u.SetMessage(d.Get("message").(string))
	u.SetPriority(int64(d.Get("priority").(int)))
	u.SetOptions(o)

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

	return m, u
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

	// GroupBy
	if terraformGroupBys, ok := data["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy, len(terraformGroupBys))
		for i, g := range terraformGroupBys {
			groupBy := g.(map[string]interface{})

			// Facet
			datadogGroupBy := datadogV1.NewMonitorFormulaAndFunctionEventQueryGroupBy(groupBy["facet"].(string))

			// Limit
			if v, ok := groupBy["limit"].(int); ok && v != 0 {
				datadogGroupBy.SetLimit(int64(v))
			}

			// Sort
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
		eventQuery.SetGroupBy(datadogGroupBys)
	} else {
		emptyGroupBy := make([]datadogV1.MonitorFormulaAndFunctionEventQueryGroupBy, 0)
		eventQuery.SetGroupBy(emptyGroupBy)
	}

	definition := datadogV1.MonitorFormulaAndFunctionEventQueryDefinitionAsMonitorFormulaAndFunctionQueryDefinition(eventQuery)
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

	thresholdWindows := make(map[string]string)
	for k, v := range map[string]string{
		"recovery_window": m.Options.ThresholdWindows.GetRecoveryWindow(),
		"trigger_window":  m.Options.ThresholdWindows.GetTriggerWindow(),
	} {
		if v != "" {
			thresholdWindows[k] = v
		}
	}

	log.Printf("[DEBUG] monitor: %+v", m)
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
	if err := d.Set("priority", m.GetPriority()); err != nil {
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
	if err := d.Set("notify_no_data", m.Options.GetNotifyNoData()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("on_missing_data", m.Options.GetOnMissingData()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("group_retention_duration", m.Options.GetGroupRetentionDuration()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("no_data_timeframe", m.Options.NoDataTimeframe.Get()); err != nil {
		return diag.FromErr(err)
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
	if err := d.Set("locked", m.Options.GetLocked()); err != nil {
		return diag.FromErr(err)
	}

	if restrictedRoles, ok := m.GetRestrictedRolesOk(); ok && restrictedRoles != nil && len(*restrictedRoles) > 0 {
		// This helper function is defined in `resource_datadog_dashboard`
		restrictedRolesCopy := buildTerraformRestrictedRoles(restrictedRoles)
		if err := d.Set("restricted_roles", restrictedRolesCopy); err != nil {
			return diag.FromErr(err)
		}
	}

	if variables, ok := m.Options.GetVariablesOk(); ok && len(*variables) > 0 {
		log.Printf("[INFO] variables: %d, %+v", len(*variables), *variables)
		terraformVariables := buildTerraformMonitorVariables(*variables)
		if err := d.Set("variables", terraformVariables); err != nil {
			return diag.FromErr(err)
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
	}
	scheduling_options := make(map[string]interface{})
	if len(evaluation_window) > 0 {
		scheduling_options["evaluation_window"] = []interface{}{evaluation_window}
		if err := d.Set("scheduling_options", []interface{}{scheduling_options}); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("notification_preset_name", m.Options.GetNotificationPresetName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func buildTerraformMonitorVariables(datadogVariables []datadogV1.MonitorFormulaAndFunctionQueryDefinition) []map[string]interface{} {
	queries := make([]map[string]interface{}, len(datadogVariables))
	for i, query := range datadogVariables {
		terraformQuery := map[string]interface{}{}
		terraformEventQueryDefinition := query.MonitorFormulaAndFunctionEventQueryDefinition
		if terraformEventQueryDefinition != nil {
			if dataSource, ok := terraformEventQueryDefinition.GetDataSourceOk(); ok {
				terraformQuery["data_source"] = dataSource
			}
			if name, ok := terraformEventQueryDefinition.GetNameOk(); ok {
				terraformQuery["name"] = name
			}
			if indexes, ok := terraformEventQueryDefinition.GetIndexesOk(); ok {
				terraformQuery["indexes"] = indexes
			}
			if search, ok := terraformEventQueryDefinition.GetSearchOk(); ok {
				if len(search.GetQuery()) > 0 {
					terraformSearch := map[string]interface{}{}
					terraformSearch["query"] = search.GetQuery()
					terraformSearchList := []map[string]interface{}{terraformSearch}
					terraformQuery["search"] = terraformSearchList
				}
			}
			if compute, ok := terraformEventQueryDefinition.GetComputeOk(); ok {
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
				terraformComputeList := []map[string]interface{}{terraformCompute}
				terraformQuery["compute"] = terraformComputeList
			}
			if terraformEventQuery, ok := terraformEventQueryDefinition.GetGroupByOk(); ok {
				terraformGroupBys := make([]map[string]interface{}, len(*terraformEventQuery))
				for i, groupBy := range *terraformEventQuery {
					// Facet
					terraformGroupBy := map[string]interface{}{
						"facet": groupBy.GetFacet(),
					}
					// Limit
					if v, ok := groupBy.GetLimitOk(); ok {
						terraformGroupBy["limit"] = *v
					}
					// Sort
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
				terraformQuery["group_by"] = &terraformGroupBys
			}
			queries[i] = terraformQuery
		}
	}
	terraformVariables := make([]map[string]interface{}, 1) // only event_queries are supported for now
	terraformVariables[0] = map[string]interface{}{"event_query": queries}

	log.Printf("[INFO] queries: %+v", terraformVariables)
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
		m, httpresp, err = apiInstances.GetMonitorsApiV1().GetMonitor(auth, i)
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
