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
		Read: dataSourceDatadogMonitorsRead,

		Schema: map[string]*schema.Schema{
			"name_filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags_filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"monitor_tags_filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			// Computed values
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"escalation_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"query": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},

			// Options
			"thresholds": {
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
				Type:     schema.TypeBool,
				Computed: true,
			},
			"new_host_delay": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"evaluation_delay": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"no_data_timeframe": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"renotify_interval": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"notify_audit": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"timeout_h": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"require_full_window": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"include_tags": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"enable_logs_sample": {
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
