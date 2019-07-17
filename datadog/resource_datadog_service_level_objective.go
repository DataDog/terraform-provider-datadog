package datadog

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogServiceLevelObjective() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogServiceLevelObjectiveCreate,
		Read:   resourceDatadogServiceLevelObjectiveRead,
		Update: resourceDatadogServiceLevelObjectiveUpdate,
		Delete: resourceDatadogServiceLevelObjectiveDelete,
		Exists: resourceDatadogServiceLevelObjectiveExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogServiceLevelObjectiveImport,
		},

		Schema: map[string]*schema.Schema{
			// Common
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(val interface{}) string {
					return strings.TrimSpace(val.(string))
				},
			},
			"tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"thresholds": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeframe": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target": {
							Type:             schema.TypeFloat,
							Required:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"target_display": {
							Type:             schema.TypeString,
							Required:         false,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"warning": {
							Type:             schema.TypeFloat,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"warning_display": {
							Type:             schema.TypeString,
							Required:         false,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
					},
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: ValidateServiceLevelObjectiveTypeString,
			},

			// Metric-Based SLO
			"query": {
				Type:          schema.TypeMap,
				Optional:      true,
				ConflictsWith: []string{"monitor_ids", "monitor_search"},
				Description:   "The metric query of good / total events",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"numerator": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(val interface{}) string {
								return strings.TrimSpace(val.(string))
							},
						},
						"denominator": {
							Type:     schema.TypeString,
							Required: true,
							StateFunc: func(val interface{}) string {
								return strings.TrimSpace(val.(string))
							},
						},
					},
				},
			},

			// Monitor-Based SLO
			"monitor_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"query", "monitor_search"},
				Description:   "A static set of monitor IDs to use as part of the SLO",
				Elem:          &schema.Schema{Type: schema.TypeInt, MinItems: 1},
			},
			"monitor_search": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"query", "monitor_ids"},
				Description:   "A dynamic search on creation for the SLO",
			},
			"groups": {
				Type:          schema.TypeSet,
				Optional:      true,
				Description:   "A static set of groups to filter monitor-based SLOs",
				ConflictsWith: []string{"query"},
				Elem:          &schema.Schema{Type: schema.TypeString, MinItems: 1},
			},
		},
	}
}

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types
func ValidateServiceLevelObjectiveTypeString(v interface{}, k string) (ws []string, errors []error) {
	switch v.(string) {
	case datadog.ServiceLevelObjectiveTypeMonitor:
		break
	case datadog.ServiceLevelObjectiveTypeMetric:
		break
	default:
		errors = append(errors, fmt.Errorf("invalid type %s specified for SLO", v.(string)))
	}
	return
}

func buildServiceLevelObjectiveStruct(d *schema.ResourceData) *datadog.ServiceLevelObjective {

	slo := datadog.ServiceLevelObjective{
		Type:        datadog.String(d.Get("type").(string)),
		Name:        datadog.String(d.Get("name").(string)),
		Description: datadog.String(d.Get("description").(string)),
	}

	if attr, ok := d.GetOk("tags"); ok {
		tags := make([]string, 0)
		for _, s := range attr.(*schema.Set).List() {
			tags = append(tags, s.(string))
		}
		// sort to make them determinate
		if len(tags) > 0 {
			sort.Strings(tags)
			slo.Tags = tags
		}
	}

	if attr, ok := d.GetOk("thresholds"); ok {
		sloThresholds := make(datadog.ServiceLevelObjectiveThresholds, 0)
		for _, rawThreshold := range attr.([]interface{}) {
			threshold := rawThreshold.(map[string]interface{})
			t := datadog.ServiceLevelObjectiveThreshold{}
			if tf, ok := threshold["timeframe"]; ok {
				t.TimeFrame = datadog.String(tf.(string))
			}

			if targetValue, ok := threshold["target"]; ok {
				t.Target = datadog.Float64(targetValue.(float64))
			}

			if warningValue, ok := threshold["warning"]; ok {
				t.Warning = datadog.Float64(warningValue.(float64))
			}

			if targetDisplayValue, ok := threshold["target_display"]; ok {
				t.TargetDisplay = datadog.String(targetDisplayValue.(string))
			}

			if warningDisplayValue, ok := threshold["warning_display"]; ok {
				t.WarningDisplay = datadog.String(warningDisplayValue.(string))
			}

			sloThresholds = append(sloThresholds, &t)
		}
		sort.Sort(sloThresholds)
		slo.Thresholds = sloThresholds
	}

	switch d.Get("type").(string) {
	case datadog.ServiceLevelObjectiveTypeMonitor:
		// add monitor components
		if attr, ok := d.GetOk("monitor_ids"); ok {
			monitorIDs := make([]int, 0)
			for _, s := range attr.(*schema.Set).List() {
				monitorIDs = append(monitorIDs, s.(int))
			}
			if len(monitorIDs) > 0 {
				sort.Ints(monitorIDs)
				slo.MonitorIDs = monitorIDs
			}
		}
		if attr, ok := d.GetOk("monitor_search"); ok {
			if len(attr.(string)) > 0 {
				slo.MonitorSearch = datadog.String(attr.(string))
			}
		}
		if attr, ok := d.GetOk("groups"); ok {
			groups := make([]string, 0)
			for _, s := range attr.(*schema.Set).List() {
				groups = append(groups, s.(string))
			}
			if len(groups) > 0 {
				sort.Strings(groups)
				slo.Groups = groups
			}
		}
	default:
		// query type
		slo.Query = &datadog.ServiceLevelObjectiveMetricQuery{
			Numerator:   datadog.String(d.Get("query.numerator").(string)),
			Denominator: datadog.String(d.Get("query.denominator").(string)),
		}
	}

	return &slo
}

func resourceDatadogServiceLevelObjectiveCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	slo := buildServiceLevelObjectiveStruct(d)
	slo, err := client.CreateServiceLevelObjective(slo)
	if err != nil {
		return fmt.Errorf("error creating service level objective: %s", err.Error())
	}

	d.SetId(slo.GetID())

	return resourceDatadogServiceLevelObjectiveRead(d, meta)
}

func resourceDatadogServiceLevelObjectiveExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*datadog.Client)

	if _, err := client.GetServiceLevelObjective(d.Id()); err != nil {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "not found") || strings.Contains(errStr, "no slo specified") {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resourceDatadogServiceLevelObjectiveRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	slo, err := client.GetServiceLevelObjective(d.Id())
	if err != nil {
		return err
	}

	thresholds := make([]map[string]interface{}, 0)
	sort.Sort(slo.Thresholds)
	for _, threshold := range slo.Thresholds {
		t := map[string]interface{}{
			"timeframe": threshold.GetTimeFrame(),
			"target":    threshold.GetTarget(),
			"warning":   threshold.GetWarning(),
		}
		if targetDisplay, ok := threshold.GetTargetDisplayOk(); ok {
			t["target_display"] = targetDisplay
		}
		if warningDisplay, ok := threshold.GetWarningDisplayOk(); ok {
			t["warning_display"] = warningDisplay
		}
		thresholds = append(thresholds, t)
	}

	tags := make([]string, 0)
	for _, s := range slo.Tags {
		tags = append(tags, s)
	}
	sort.Strings(tags)

	log.Printf("[DEBUG] service level objective: %+v", slo)
	d.Set("name", slo.GetName())
	d.Set("description", slo.GetDescription())
	d.Set("type", slo.GetType())
	d.Set("tags", tags)
	d.Set("thresholds", thresholds)
	switch slo.GetType() {
	case datadog.ServiceLevelObjectiveTypeMonitor:
		// monitor type
		if len(slo.MonitorIDs) > 0 {
			sort.Ints(slo.MonitorIDs)
			d.Set("monitor_ids", slo.MonitorIDs)
		}
		if ms, ok := slo.GetMonitorSearchOk(); ok {
			d.Set("monitor_search", ms)
		}
		sort.Strings(slo.Groups)
		d.Set("groups", slo.Groups)
	default:
		// metric type
		query := make(map[string]interface{})
		q := slo.GetQuery()
		query["numerator"] = q.GetNumerator()
		query["denominator"] = q.GetDenominator()
		d.Set("query", query)
	}

	return nil
}

func resourceDatadogServiceLevelObjectiveUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)
	slo := &datadog.ServiceLevelObjective{
		ID: datadog.String(d.Id()),
	}

	if attr, ok := d.GetOk("name"); ok {
		slo.SetName(attr.(string))
	}

	if attr, ok := d.GetOk("description"); ok {
		slo.SetDescription(attr.(string))
	}

	if attr, ok := d.GetOk("type"); ok {
		slo.SetType(attr.(string))
	}

	switch slo.GetType() {
	case datadog.ServiceLevelObjectiveTypeMonitor:
		// monitor type
		if attr, ok := d.GetOk("monitor_ids"); ok {
			s := make([]int, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, v.(int))
			}
			sort.Ints(s)
			slo.MonitorIDs = s
		}
		if attr, ok := d.GetOk("monitor_search"); ok {
			slo.SetMonitorSearch(attr.(string))
		}
		if attr, ok := d.GetOk("groups"); ok {
			s := make([]string, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, v.(string))
			}
			sort.Strings(s)
			slo.Groups = s
		}
	default:
		// metric type
		if attr, ok := d.GetOk("query"); ok {
			query := attr.(map[string]interface{})
			slo.SetQuery(datadog.ServiceLevelObjectiveMetricQuery{
				Numerator:   datadog.String(query["numerator"].(string)),
				Denominator: datadog.String(query["denominator"].(string)),
			})
		}
	}

	if attr, ok := d.GetOk("tags"); ok {
		s := make([]string, 0)
		for _, v := range attr.(*schema.Set).List() {
			s = append(s, v.(string))
		}
		sort.Strings(s)
		slo.Tags = s
	}

	if attr, ok := d.GetOk("thresholds"); ok {
		sloThresholds := make(datadog.ServiceLevelObjectiveThresholds, 0)
		thresholds := attr.([]map[string]interface{})
		for _, threshold := range thresholds {
			t := datadog.ServiceLevelObjectiveThreshold{
				TimeFrame: datadog.String(threshold["timeframe"].(string)),
				Target:    datadog.Float64(threshold["target"].(float64)),
			}
			if warningValueRaw, ok := threshold["warning"]; ok {
				t.Warning = datadog.Float64(warningValueRaw.(float64))
			}
			// display settings
			if targetDisplay, ok := threshold["target_display"]; ok {
				t.TargetDisplay = datadog.String(targetDisplay.(string))
			}
			if warningDisplay, ok := threshold["warning_display"]; ok {
				t.WarningDisplay = datadog.String(warningDisplay.(string))
			}
			sloThresholds = append(sloThresholds, &t)
		}
		if len(sloThresholds) > 0 {
			sort.Sort(sloThresholds)
			slo.Thresholds = sloThresholds
		}
	}

	if _, err := client.UpdateServiceLevelObjective(slo); err != nil {
		return fmt.Errorf("error updating service level objective: %s", err.Error())
	}

	return resourceDatadogServiceLevelObjectiveRead(d, meta)
}

func resourceDatadogServiceLevelObjectiveDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	return client.DeleteServiceLevelObjective(d.Id())
}

func resourceDatadogServiceLevelObjectiveImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogServiceLevelObjectiveRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
