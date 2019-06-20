package datadog

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

const logAlertMonitorType = "log alert"

func resourceDatadogMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogMonitorCreate,
		Read:   resourceDatadogMonitorRead,
		Update: resourceDatadogMonitorUpdate,
		Delete: resourceDatadogMonitorDelete,
		Exists: resourceDatadogMonitorExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogMonitorImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"message": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"escalation_message": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"query": {
				Type:     schema.TypeString,
				Required: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			// Options
			"thresholds": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ok": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"warning": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"critical": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"unknown": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"warning_recovery": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"critical_recovery": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
					},
				},
				DiffSuppressFunc: suppressDataDogFloatIntDiff,
			},
			"threshold_windows": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recovery_window": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"trigger_window": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"notify_no_data": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"new_host_delay": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"evaluation_delay": {
				Type:     schema.TypeInt,
				Computed: true,
				Optional: true,
			},
			"no_data_timeframe": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
			},
			"renotify_interval": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"notify_audit": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"timeout_h": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"require_full_window": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"locked": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"silenced": {
				Type:       schema.TypeMap,
				Optional:   true,
				Elem:       schema.TypeInt,
				Deprecated: "use Downtime Resource instead",
			},
			"include_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			// since this is only useful for "log alert" type, we don't set a default value
			// if we did set it, it would be used for all types; we have to handle this manually
			// throughout the code
			"enable_logs_sample": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func buildMonitorStruct(d *schema.ResourceData) *datadog.Monitor {

	var thresholds datadog.ThresholdCount

	if r, ok := d.GetOk("thresholds.ok"); ok {
		thresholds.SetOk(json.Number(r.(string)))
	}
	if r, ok := d.GetOk("thresholds.warning"); ok {
		thresholds.SetWarning(json.Number(r.(string)))
	}
	if r, ok := d.GetOk("thresholds.unknown"); ok {
		thresholds.SetUnknown(json.Number(r.(string)))
	}
	if r, ok := d.GetOk("thresholds.critical"); ok {
		thresholds.SetCritical(json.Number(r.(string)))
	}
	if r, ok := d.GetOk("thresholds.warning_recovery"); ok {
		thresholds.SetWarningRecovery(json.Number(r.(string)))
	}
	if r, ok := d.GetOk("thresholds.critical_recovery"); ok {
		thresholds.SetCriticalRecovery(json.Number(r.(string)))
	}

	var threshold_windows datadog.ThresholdWindows

	if r, ok := d.GetOk("threshold_windows.recovery_window"); ok {
		threshold_windows.SetRecoveryWindow(r.(string))
	}

	if r, ok := d.GetOk("threshold_windows.trigger_window"); ok {
		threshold_windows.SetTriggerWindow(r.(string))
	}

	o := datadog.Options{
		Thresholds:        &thresholds,
		ThresholdWindows:  &threshold_windows,
		NotifyNoData:      datadog.Bool(d.Get("notify_no_data").(bool)),
		RequireFullWindow: datadog.Bool(d.Get("require_full_window").(bool)),
		IncludeTags:       datadog.Bool(d.Get("include_tags").(bool)),
	}
	if attr, ok := d.GetOk("silenced"); ok {
		s := make(map[string]int)
		// TODO: this is not very defensive, test if we can fail on non int input
		for k, v := range attr.(map[string]interface{}) {
			s[k] = v.(int)
		}
		o.Silenced = s
	}
	if attr, ok := d.GetOk("notify_no_data"); ok {
		o.SetNotifyNoData(attr.(bool))
	}
	if attr, ok := d.GetOk("new_host_delay"); ok {
		o.SetNewHostDelay(attr.(int))
	}
	if attr, ok := d.GetOk("evaluation_delay"); ok {
		o.SetEvaluationDelay(attr.(int))
	}
	if attr, ok := d.GetOk("no_data_timeframe"); ok {
		o.NoDataTimeframe = datadog.NoDataTimeframe(attr.(int))
	}
	if attr, ok := d.GetOk("renotify_interval"); ok {
		o.SetRenotifyInterval(attr.(int))
	}
	if attr, ok := d.GetOk("notify_audit"); ok {
		o.SetNotifyAudit(attr.(bool))
	}
	if attr, ok := d.GetOk("timeout_h"); ok {
		o.SetTimeoutH(attr.(int))
	}
	if attr, ok := d.GetOk("escalation_message"); ok {
		o.SetEscalationMessage(attr.(string))
	}
	if attr, ok := d.GetOk("locked"); ok {
		o.SetLocked(attr.(bool))
	}

	m := datadog.Monitor{
		Type:    datadog.String(d.Get("type").(string)),
		Query:   datadog.String(d.Get("query").(string)),
		Name:    datadog.String(d.Get("name").(string)),
		Message: datadog.String(d.Get("message").(string)),
		Options: &o,
	}

	if m.GetType() == logAlertMonitorType {
		if attr, ok := d.GetOk("enable_logs_sample"); ok {
			o.SetEnableLogsSample(attr.(bool))
		} else {
			o.SetEnableLogsSample(false)
		}
	}

	if attr, ok := d.GetOk("tags"); ok {
		tags := []string{}
		for _, s := range attr.(*schema.Set).List() {
			tags = append(tags, s.(string))
		}
		sort.Strings(tags)
		m.Tags = tags
	}

	return &m
}

func resourceDatadogMonitorExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*datadog.Client)

	i, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, err
	}

	if _, err = client.GetMonitor(i); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func getUnmutedScopes(d *schema.ResourceData) []string {
	var unmuteScopes []string
	if attr, ok := d.GetOk("silenced"); ok {
		for k, v := range attr.(map[string]interface{}) {
			if v.(int) == -1 {
				unmuteScopes = append(unmuteScopes, k)
			}
		}
		log.Printf("[DEBUG] Unmute Scopes are: %v", unmuteScopes)
	}
	return unmuteScopes
}

func resourceDatadogMonitorCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*datadog.Client)

	m := buildMonitorStruct(d)
	m, err := client.CreateMonitor(m)
	if err != nil {
		return fmt.Errorf("error updating monitor: %s", err.Error())
	}

	d.SetId(strconv.Itoa(m.GetId()))

	return resourceDatadogMonitorRead(d, meta)
}

func resourceDatadogMonitorRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	i, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	m, err := client.GetMonitor(i)
	if err != nil {
		return err
	}

	thresholds := make(map[string]string)
	for k, v := range map[string]json.Number{
		"ok":                m.Options.Thresholds.GetOk(),
		"warning":           m.Options.Thresholds.GetWarning(),
		"critical":          m.Options.Thresholds.GetCritical(),
		"unknown":           m.Options.Thresholds.GetUnknown(),
		"warning_recovery":  m.Options.Thresholds.GetWarningRecovery(),
		"critical_recovery": m.Options.Thresholds.GetCriticalRecovery(),
	} {
		s := v.String()
		if s != "" {
			thresholds[k] = s
		}
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

	tags := []string{}
	for _, s := range m.Tags {
		tags = append(tags, s)
	}
	sort.Strings(tags)

	log.Printf("[DEBUG] monitor: %v", m)
	d.Set("name", m.GetName())
	d.Set("message", m.GetMessage())
	d.Set("query", m.GetQuery())
	if typ, ok := m.GetTypeOk(); ok {
		if d.Get("type").(string) == "metric alert" && typ == "query alert" ||
			d.Get("type").(string) == "query alert" && typ == "metric alert" {
			/* Datadog API quirk, see https://github.com/hashicorp/terraform/issues/13784
			*                     and https://github.com/terraform-providers/terraform-provider-datadog/issues/241
			* If current type of monitor is "metric alert" and the API is returning "query alert",
			* we want to keep "metric alert". We previously had this as DiffSuppressFunc on "type".
			* After adding a call to "resourceDatadogMonitorRead" in create/update methods, this
			* started creating the monitor as "query alert". The same applies for the reverse, when the
			* current type of monitor is "query alert" and the API is returning "metric alert"
			* To make sure that the behaviour stays
			* the same, we added this code (which made DiffSuppressFunc useless, so we removed it).
			 */
		} else {
			d.Set("type", typ)
		}
	}
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

	if m.GetType() == logAlertMonitorType {
		d.Set("enable_logs_sample", m.Options.GetEnableLogsSample())
	}

	// The Datadog API doesn't return old timestamps or support a special value for unmuting scopes
	// So we provide this functionality by saving values to the state
	apiSilenced := m.Options.Silenced
	configSilenced := d.Get("silenced").(map[string]interface{})

	for _, scope := range getUnmutedScopes(d) {
		if _, ok := apiSilenced[scope]; !ok {
			apiSilenced[scope] = -1
		}
	}

	// Ignore any timestamps in the past that aren't -1 or 0
	for k, v := range configSilenced {
		if v.(int) < int(time.Now().Unix()) && v.(int) != 0 && v.(int) != -1 {
			// sync the state with whats in the config so its ignored
			apiSilenced[k] = v.(int)
		}
	}
	d.Set("silenced", apiSilenced)

	return nil
}

func resourceDatadogMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	m := &datadog.Monitor{}

	i, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	m.Id = datadog.Int(i)
	if attr, ok := d.GetOk("name"); ok {
		m.SetName(attr.(string))
	}
	if attr, ok := d.GetOk("message"); ok {
		m.SetMessage(attr.(string))
	}
	if attr, ok := d.GetOk("query"); ok {
		m.SetQuery(attr.(string))
	}

	if attr, ok := d.GetOk("tags"); ok {
		s := make([]string, 0)
		for _, v := range attr.(*schema.Set).List() {
			s = append(s, v.(string))
		}
		sort.Strings(s)
		m.Tags = s
	}

	o := datadog.Options{
		NotifyNoData:      datadog.Bool(d.Get("notify_no_data").(bool)),
		RequireFullWindow: datadog.Bool(d.Get("require_full_window").(bool)),
		IncludeTags:       datadog.Bool(d.Get("include_tags").(bool)),
	}
	if attr, ok := d.GetOk("thresholds"); ok {
		thresholds := attr.(map[string]interface{})
		o.Thresholds = &datadog.ThresholdCount{} // TODO: This is a little annoying..
		if thresholds["ok"] != nil {
			o.Thresholds.SetOk(json.Number(thresholds["ok"].(string)))
		}
		if thresholds["warning"] != nil {
			o.Thresholds.SetWarning(json.Number(thresholds["warning"].(string)))
		}
		if thresholds["critical"] != nil {
			o.Thresholds.SetCritical(json.Number(thresholds["critical"].(string)))
		}
		if thresholds["unknown"] != nil {
			o.Thresholds.SetUnknown(json.Number(thresholds["unknown"].(string)))
		}
		if thresholds["warning_recovery"] != nil {
			o.Thresholds.SetWarningRecovery(json.Number(thresholds["warning_recovery"].(string)))
		}
		if thresholds["critical_recovery"] != nil {
			o.Thresholds.SetCriticalRecovery(json.Number(thresholds["critical_recovery"].(string)))
		}
	}

	if attr, ok := d.GetOk("threshold_windows"); ok {
		thresholdWindows := attr.(map[string]interface{})
		o.ThresholdWindows = &datadog.ThresholdWindows{}
		if thresholdWindows["recovery_window"] != nil {
			o.ThresholdWindows.SetRecoveryWindow(thresholdWindows["recovery_window"].(string))
		}
		if thresholdWindows["trigger_window"] != nil {
			o.ThresholdWindows.SetTriggerWindow(thresholdWindows["trigger_window"].(string))
		}
	}

	newHostDelay := d.Get("new_host_delay")
	o.SetNewHostDelay(newHostDelay.(int))

	if attr, ok := d.GetOk("evaluation_delay"); ok {
		o.SetEvaluationDelay(attr.(int))
	}
	if attr, ok := d.GetOk("no_data_timeframe"); ok {
		o.NoDataTimeframe = datadog.NoDataTimeframe(attr.(int))
	}
	if attr, ok := d.GetOk("renotify_interval"); ok {
		o.SetRenotifyInterval(attr.(int))
	}
	if attr, ok := d.GetOk("notify_audit"); ok {
		o.SetNotifyAudit(attr.(bool))
	}
	if attr, ok := d.GetOk("timeout_h"); ok {
		o.SetTimeoutH(attr.(int))
	}
	if attr, ok := d.GetOk("escalation_message"); ok {
		o.SetEscalationMessage(attr.(string))
	}

	silenced := false
	configuredSilenced := map[string]int{}
	if attr, ok := d.GetOk("silenced"); ok {
		// TODO: this is not very defensive, test if we can fail non int input
		s := make(map[string]int)
		for k, v := range attr.(map[string]interface{}) {
			s[k] = v.(int)
			configuredSilenced[k] = v.(int)
		}
		o.Silenced = s
		silenced = true
	}
	if attr, ok := d.GetOk("locked"); ok {
		o.SetLocked(attr.(bool))
	}
	// can't use m.GetType here, since it's not filled for purposes of updating
	if d.Get("type") == logAlertMonitorType {
		if attr, ok := d.GetOk("enable_logs_sample"); ok {
			o.SetEnableLogsSample(attr.(bool))
		} else {
			o.SetEnableLogsSample(false)
		}
	}

	m.Options = &o

	if err = client.UpdateMonitor(m); err != nil {
		return fmt.Errorf("error updating monitor: %s", err.Error())
	}

	var retval error
	if retval = resourceDatadogMonitorRead(d, meta); retval != nil {
		return retval
	}

	// if the silenced section was removed from the config, we unmute it via the API
	// The API wouldn't automatically unmute the monitor if the config is just missing
	// else we check what other silenced scopes were added from API response in the
	// "read" above and add them to "unmutedScopes" to be explicitly unmuted (because
	// they're "drift")
	unmutedScopes := getUnmutedScopes(d)
	if newSilenced, ok := d.GetOk("silenced"); ok && !silenced {
		retval = client.UnmuteMonitorScopes(*m.Id, &datadog.UnmuteMonitorScopes{AllScopes: datadog.Bool(true)})
		d.Set("silenced", map[string]int{})
	} else {
		for scope := range newSilenced.(map[string]interface{}) {
			if _, ok := configuredSilenced[scope]; !ok {
				unmutedScopes = append(unmutedScopes, scope)
			}
		}
	}

	// Similarly, if the silenced attribute is -1, lets unmute those scopes
	if len(unmutedScopes) != 0 {
		for _, scope := range unmutedScopes {
			client.UnmuteMonitorScopes(*m.Id, &datadog.UnmuteMonitorScopes{Scope: &scope})
		}
	}

	return resourceDatadogMonitorRead(d, meta)
}

func resourceDatadogMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	i, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}

	if err = client.DeleteMonitor(i); err != nil {
		return err
	}

	return nil
}

func resourceDatadogMonitorImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogMonitorRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

// Ignore any diff that results from the mix of ints or floats returned from the
// DataDog API.
func suppressDataDogFloatIntDiff(k, old, new string, d *schema.ResourceData) bool {
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
