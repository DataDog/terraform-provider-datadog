package datadog

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing monitor for use in other resources.",
		ReadContext: dataSourceDatadogMonitorRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name_filter": {
					Description: "A monitor name to limit the search.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tags_filter": {
					Description: "A list of tags to limit the search. This filters on the monitor scope.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"monitor_tags_filter": {
					Description: "A list of monitor tags to limit the search. This filters on the tags set on the monitor itself.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},

				// Computed values
				"name": {
					Description: "Name of the monitor",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"message": {
					Description: "Message included with notifications for this monitor",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"escalation_message": {
					Description: "Message included with a re-notification for this monitor.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"query": {
					Description: "Query of the monitor.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"type": {
					Description: "Type of the monitor.",
					Type:        schema.TypeString,
					Computed:    true,
				},

				// Options
				"monitor_thresholds": {
					Description: "Alert thresholds of the monitor.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"ok": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"warning": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"critical": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"unknown": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"warning_recovery": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"critical_recovery": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"monitor_threshold_windows": {
					Description: "Mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. This is only used by anomaly monitors.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"recovery_window": {
								Type:     schema.TypeString,
								Computed: true,
							},
							"trigger_window": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},
				"notify_no_data": {
					Description: "Whether or not this monitor notifies when data stops reporting.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"on_missing_data": {
					Description: "Controls how groups or monitors are treated if an evaluation does not return any data points. The default option results in different behavior depending on the monitor query type. For monitors using `Count` queries, an empty monitor evaluation is treated as 0 and is compared to the threshold conditions. For monitors using any query type other than `Count`, for example `Gauge`, `Measure`, or `Rate`, the monitor shows the last known status. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors. Valid values are: `show_no_data`, `show_and_notify_no_data`, `resolve`, and `default`.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"group_retention_duration": {
					Description: "The time span after which groups with missing data are dropped from the monitor state. The minimum value is one hour, and the maximum value is 72 hours. Example values are: 60m, 1h, and 2d. This option is only available for APM Trace Analytics, Audit Trail, CI, Error Tracking, Event, Logs, and RUM monitors.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"new_group_delay": {
					Description: "Time (in seconds) to skip evaluations for new groups.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"new_host_delay": {
					Description: "Time (in seconds) allowing a host to boot and applications to fully start before starting the evaluation of monitor results.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"evaluation_delay": {
					Description: "Time (in seconds) for which evaluation is delayed. This is only used by metric monitors.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"no_data_timeframe": {
					Description: "The number of minutes before the monitor notifies when data stops reporting.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"renotify_interval": {
					Description: "The number of minutes after the last notification before the monitor re-notifies on the current status.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"renotify_occurrences": {
					Description: "The number of re-notification messages that should be sent on the current status.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"renotify_statuses": {
					Description: "The types of statuses for which re-notification messages should be sent.",
					Type:        schema.TypeSet,
					Elem: &schema.Schema{
						Type:             schema.TypeString,
						ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewMonitorRenotifyStatusTypeFromValue),
					},
					Computed: true,
				},
				"notify_audit": {
					Description: "Whether or not tagged users are notified on changes to the monitor.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"timeout_h": {
					Description: "Number of hours of the monitor not reporting data before it automatically resolves from a triggered state.",
					Type:        schema.TypeInt,
					Computed:    true,
				},
				"require_full_window": {
					Description: "Whether or not the monitor needs a full window of data before it is evaluated.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"locked": {
					Description: "Whether or not changes to the monitor are restricted to the creator or admins.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"restricted_roles": {
					// Uncomment when generally available
					// Description: "A list of role identifiers to associate with the monitor. Cannot be used with `locked`.",
					Type:     schema.TypeSet,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"include_tags": {
					Description: "Whether or not notifications from the monitor automatically inserts its triggering tags into the title.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"tags": {
					Description: "List of tags associated with the monitor.",
					// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
					// we order them explicitly in the read/create/update methods of this resource and using
					// TypeSet makes Terraform ignore differences in order when creating a plan
					Type:     schema.TypeSet,
					Computed: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"enable_logs_sample": {
					Description: "Whether or not a list of log values which triggered the alert is included. This is only used by log monitors.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"enable_samples": {
					Description: "Whether or not a list of samples which triggered the alert is included. This is only used by CI Test and Pipeline monitors.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"groupby_simple_monitor": {
					Description: "Whether or not to trigger one alert if any source breaches a threshold.",
					Type:        schema.TypeBool,
					Computed:    true,
				},
				"notify_by": {
					Description: "Controls what granularity a monitor alerts on. Only available for monitors with groupings. For instance, a monitor grouped by `cluster`, `namespace`, and `pod` can be configured to only notify on each new `cluster` violating the alert conditions by setting `notify_by` to `['cluster']`. Tags mentioned in `notify_by` must be a subset of the grouping tags in the query. For example, a query grouped by `cluster` and `namespace` cannot notify on `region`. Setting `notify_by` to `[*]` configures the monitor to notify as a simple-alert.",
					Type:        schema.TypeSet,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"scheduling_options": {
					Description: "Configuration options for scheduling.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"evaluation_window": {
								Description: "Configuration options for the evaluation window. If `hour_starts` is set, no other fields may be set. Otherwise, `day_starts` and `month_starts` must be set together.",
								Type:        schema.TypeList,
								Computed:    true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"day_starts": {
											Description: "The time of the day at which a one day cumulative evaluation window starts. Must be defined in UTC time in `HH:mm` format.",
											Type:        schema.TypeString,
											Computed:    true,
										},
										"month_starts": {
											Description: "The day of the month at which a one month cumulative evaluation window starts. Must be a value of 1.",
											Type:        schema.TypeInt,
											Computed:    true,
										},
										"hour_starts": {
											Description: "The minute of the hour at which a one hour cumulative evaluation window starts. Must be between 0 and 59.",
											Type:        schema.TypeInt,
											Computed:    true,
										},
									},
								},
							},
						},
					},
				},
				"notification_preset_name": {
					Description: "Toggles the display of additional content sent in the monitor notification. Valid values are: `show_all`, `hide_query`, `hide_handles`, and `hide_all`.",
					Type:        schema.TypeString,
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	optionalParams := datadogV1.NewListMonitorsOptionalParameters()
	if v, ok := d.GetOk("name_filter"); ok {
		optionalParams = optionalParams.WithName(v.(string))
	}
	if v, ok := d.GetOk("tags_filter"); ok {
		optionalParams = optionalParams.WithTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}
	if v, ok := d.GetOk("monitor_tags_filter"); ok {
		optionalParams = optionalParams.WithMonitorTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}

	monitors, httpresp, err := apiInstances.GetMonitorsApiV1().ListMonitors(auth, *optionalParams)
	if len(monitors) > 1 {
		return diag.Errorf("your query returned more than one result, please try a more specific search criteria")
	}
	if len(monitors) == 0 {
		return diag.Errorf("your query returned no result, please try a less specific search criteria")
	}
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying monitors")
	}

	m := monitors[0]
	if err := utils.CheckForUnparsed(m); err != nil {
		return diag.FromErr(err)
	}

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

	var tags []string
	tags = append(tags, m.GetTags()...)
	sort.Strings(tags)

	var restricted_roles []string
	restricted_roles = append(restricted_roles, m.GetRestrictedRoles()...)
	sort.Strings(restricted_roles)

	d.SetId(strconv.FormatInt(m.GetId(), 10))
	d.Set("name", m.GetName())
	d.Set("message", m.GetMessage())
	d.Set("query", m.GetQuery())
	d.Set("type", m.GetType())

	if err := d.Set("monitor_thresholds", []interface{}{thresholds}); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("monitor_threshold_windows", []interface{}{thresholdWindows}); err != nil {
		return diag.FromErr(err)
	}

	d.Set("new_group_delay", m.Options.GetNewGroupDelay())
	d.Set("new_host_delay", m.Options.GetNewHostDelay())
	d.Set("evaluation_delay", m.Options.GetEvaluationDelay())
	d.Set("notify_no_data", m.Options.GetNotifyNoData())
	d.Set("on_missing_data", m.Options.GetOnMissingData())
	d.Set("group_retention_duration", m.Options.GetGroupRetentionDuration())
	d.Set("no_data_timeframe", m.Options.GetNoDataTimeframe())
	d.Set("renotify_interval", m.Options.GetRenotifyInterval())
	d.Set("renotify_occurrences", m.Options.GetRenotifyOccurrences())
	d.Set("renotify_statuses", m.Options.GetRenotifyStatuses())
	d.Set("notify_audit", m.Options.GetNotifyAudit())
	d.Set("timeout_h", m.Options.GetTimeoutH())
	d.Set("escalation_message", m.Options.GetEscalationMessage())
	d.Set("include_tags", m.Options.GetIncludeTags())
	d.Set("tags", tags)
	d.Set("require_full_window", m.Options.GetRequireFullWindow()) // TODO Is this one of those options that we neeed to check?
	d.Set("locked", m.Options.GetLocked())
	d.Set("restricted_roles", restricted_roles)
	d.Set("groupby_simple_monitor", m.Options.GetGroupbySimpleMonitor())
	d.Set("notify_by", m.Options.GetNotifyBy())
	d.Set("notification_preset_name", m.Options.GetNotificationPresetName())

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

	if m.GetType() == datadogV1.MONITORTYPE_LOG_ALERT {
		d.Set("enable_logs_sample", m.Options.GetEnableLogsSample())
	}

	return nil
}

func expandStringList(configured []interface{}) []string {
	vs := make([]string, 0, len(configured))
	for _, v := range configured {
		val, ok := v.(string)
		if ok && val != "" {
			vs = append(vs, v.(string))
		}
	}
	return vs
}
