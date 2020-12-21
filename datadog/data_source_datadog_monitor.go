package datadog

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing monitor for use in other resources.",
		Read:        dataSourceDatadogMonitorsRead,

		Schema: map[string]*schema.Schema{
			"name_filter": {
				Description: "A monitor name to limit the search.",
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags_filter": {
				Description: "A list of tags to limit the search. This filters on the monitor scope.",
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"monitor_tags_filter": {
				Description: "A list of monitor tags to limit the search. This filters on the tags set on the monitor itself.",
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Computed values
			"name": {
				Description: "Name of the monitor",
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Description: "Message included with notifications for this monitor.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"escalation_message": {
				Description: "Message included with a re-notification for this monitor.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"query": {
				Description: "Query of the monitor.",
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Description: "Type of the monitor.",
				Type:     schema.TypeString,
				Computed: true,
			},

			// Options
			"thresholds": {
				Description: "Alert thresholds of the monitor.",
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ok": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"warning": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"critical": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"unknown": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"warning_recovery": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"critical_recovery": {
							Type:     schema.TypeFloat,
							Computed: true,
						},
					},
				},
			},
			"threshold_windows": {
				Description: "Mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. This is only used by anomaly monitors.",
				Type:     schema.TypeMap,
				Computed: true,
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
				Type:     schema.TypeBool,
				Computed: true,
			},
			"new_host_delay": {
				Description: "Time (in seconds) allowing a host to boot and applications to fully start before starting the evaluation of monitor results.",
				Type:     schema.TypeInt,
				Computed: true,
			},
			"evaluation_delay": {
				Description: "Time (in seconds) for which evaluation is delayed. This is only used by metric monitors.",
				Type:     schema.TypeInt,
				Computed: true,
			},
			"no_data_timeframe": {
				Description: "The number of minutes before the monitor notifies when data stops reporting.",
				Type:     schema.TypeInt,
				Computed: true,
			},
			"renotify_interval": {
				Description: "The number of minutes after the last notification before the monitor re-notifies on the current status.",
				Type:     schema.TypeInt,
				Computed: true,
			},
			"notify_audit": {
				Description: "Whether or not tagged users are notified on changes to the monitor.",
				Type:     schema.TypeBool,
				Computed: true,
			},
			"timeout_h": {
				Description: "Number of hours of the monitor not reporting data before it automatically resolves from a triggered state.",
				Type:     schema.TypeInt,
				Computed: true,
			},
			"require_full_window": {
				Description: "Whether or not the monitor needs a full window of data before it is evaluated.",
				Type:     schema.TypeBool,
				Computed: true,
			},
			"locked": {
				Description: "Whether or not changes to the monitor are restricted to the creator or admins.",
				Type:     schema.TypeBool,
				Computed: true,
			},
			"include_tags": {
				Description: "Whether or not notifications from the monitor automatically inserts its triggering tags into the title.",
				Type:     schema.TypeBool,
				Computed: true,
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
				Type:     schema.TypeBool,
				Computed: true,
			},
		},
	}
}

func dataSourceDatadogMonitorsRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	req := datadogClientV1.MonitorsApi.ListMonitors(authV1)
	if v, ok := d.GetOk("name_filter"); ok {
		req = req.Name(v.(string))
	}
	if v, ok := d.GetOk("tags_filter"); ok {
		req = req.Tags(strings.Join(expandStringList(v.([]interface{})), ","))
	}
	if v, ok := d.GetOk("monitor_tags_filter"); ok {
		req = req.MonitorTags(strings.Join(expandStringList(v.([]interface{})), ","))
	}

	monitors, _, err := req.Execute()
	if err != nil {
		return translateClientError(err, "error querying monitors")
	}
	if len(monitors) > 1 {
		return fmt.Errorf("your query returned more than one result, please try a more specific search criteria")
	}
	if len(monitors) == 0 {
		return fmt.Errorf("your query returned no result, please try a less specific search criteria")
	}

	m := monitors[0]

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
	for _, s := range m.GetTags() {
		tags = append(tags, s)
	}
	sort.Strings(tags)

	d.SetId(strconv.FormatInt(m.GetId(), 10))
	d.Set("name", m.GetName())
	d.Set("message", m.GetMessage())
	d.Set("query", m.GetQuery())
	d.Set("type", m.GetType())

	d.Set("thresholds", thresholds)
	d.Set("threshold_windows", thresholdWindows)

	d.Set("new_host_delay", m.Options.GetNewHostDelay())
	d.Set("evaluation_delay", m.Options.GetEvaluationDelay())
	d.Set("notify_no_data", m.Options.GetNotifyNoData())
	d.Set("no_data_timeframe", m.Options.NoDataTimeframe)
	d.Set("renotify_interval", m.Options.GetRenotifyInterval())
	d.Set("notify_audit", m.Options.GetNotifyAudit())
	d.Set("timeout_h", m.Options.GetTimeoutH())
	d.Set("escalation_message", m.Options.GetEscalationMessage())
	d.Set("include_tags", m.Options.GetIncludeTags())
	d.Set("tags", tags)
	d.Set("require_full_window", m.Options.GetRequireFullWindow()) // TODO Is this one of those options that we neeed to check?
	d.Set("locked", m.Options.GetLocked())

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
