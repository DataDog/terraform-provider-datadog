package datadog

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogServiceLevelObjective() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog service level objective resource. This can be used to create and manage Datadog service level objectives.",
		CreateContext: resourceDatadogServiceLevelObjectiveCreate,
		ReadContext:   resourceDatadogServiceLevelObjectiveRead,
		UpdateContext: resourceDatadogServiceLevelObjectiveUpdate,
		DeleteContext: resourceDatadogServiceLevelObjectiveDelete,
		CustomizeDiff: resourceDatadogServiceLevelObjectiveCustomizeDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			// Common
			"name": {
				Description: "Name of Datadog service level objective",
				Type:        schema.TypeString,
				Required:    true,
			},
			"description": {
				Description:      "A description of this service level objective.",
				Type:             schema.TypeString,
				Optional:         true,
				StateFunc:        trimStateValue,
				DiffSuppressFunc: diffTrimmedValues,
			},
			"tags": {
				// we use TypeSet to represent tags, paradoxically to be able to maintain them ordered;
				// we order them explicitly in the read/create/update methods of this resource and using
				// TypeSet makes Terraform ignore differences in order when creating a plan
				Type:        schema.TypeSet,
				Description: "A list of tags to associate with your service level objective. This can help you categorize and filter service level objectives in the service level objectives page of the UI. Note: it's not currently possible to filter by these tags when querying via the API",
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"thresholds": {
				Description: "A list of thresholds and targets that define the service level objectives from the provided SLIs.",
				Type:        schema.TypeList,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"timeframe": {
							Description:      "The time frame for the objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API documentation page.",
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTimeframeFromValue),
						},
						"target": {
							Description:      "The objective's target in`[0,100]`.",
							Type:             schema.TypeFloat,
							Required:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"target_display": {
							Description:      "A string representation of the target that indicates its precision. It uses trailing zeros to show significant decimal places (e.g. `98.00`).",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogSLODisplayValueDiff,
						},
						"warning": {
							Description:      "The objective's warning value in `[0,100]`. This must be greater than the target value.",
							Type:             schema.TypeFloat,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogFloatIntDiff,
						},
						"warning_display": {
							Description:      "A string representation of the warning target (see the description of the target_display field for details).",
							Type:             schema.TypeString,
							Optional:         true,
							DiffSuppressFunc: suppressDataDogSLODisplayValueDiff,
						},
					},
				},
			},
			"type": {
				Description:      "The type of the service level objective. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/service-level-objectives/#create-a-slo-object).",
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewSLOTypeFromValue),
			},
			"force_delete": {
				Description: "A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. dashboards).",
				Type:        schema.TypeBool,
				Optional:    true,
			},

			// Metric-Based SLO
			"query": {
				// we use TypeList here because of https://github.com/hashicorp/terraform/issues/6215/
				Type:          schema.TypeList,
				MaxItems:      1,
				Optional:      true,
				ConflictsWith: []string{"monitor_ids", "groups"},
				Description:   "The metric query of good / total events",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"numerator": {
							Description:      "The sum of all the `good` events.",
							Type:             schema.TypeString,
							Required:         true,
							StateFunc:        trimStateValue,
							DiffSuppressFunc: diffTrimmedValues,
						},
						"denominator": {
							Description:      "The sum of the `total` events.",
							Type:             schema.TypeString,
							Required:         true,
							StateFunc:        trimStateValue,
							DiffSuppressFunc: diffTrimmedValues,
						},
					},
				},
			},

			// Monitor-Based SLO
			"monitor_ids": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"query"},
				Description:   "A static set of monitor IDs to use as part of the SLO",
				Elem:          &schema.Schema{Type: schema.TypeInt, MinItems: 1},
			},
			"groups": {
				Type:          schema.TypeSet,
				Optional:      true,
				Description:   "A static set of groups to filter monitor-based SLOs",
				ConflictsWith: []string{"query"},
				Elem:          &schema.Schema{Type: schema.TypeString, MinItems: 1},
			},
			"validate": {
				Description: "Whether or not to validate the SLO.",
				Type:        schema.TypeBool,
				Optional:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// This is never sent to the backend, so it should never generate a diff
					return true
				},
			},
		},
	}
}

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types

// ValidateServiceLevelObjectiveTypeString is a ValidateFunc that ensures the SLO is of one of the supported types

// Use CustomizeDiff to do monitor validation
func resourceDatadogServiceLevelObjectiveCustomizeDiff(ctx context.Context, diff *schema.ResourceDiff, meta interface{}) error {
	if validate, ok := diff.GetOkExists("validate"); ok && !validate.(bool) {
		// Explicitly skip validation
		log.Printf("[DEBUG] Validate is %v, skipping validation", validate.(bool))
		return nil
	}

	if val, ok := diff.GetOk("type"); ok && (val != string(datadogV1.SLOTYPE_MONITOR)) {
		// If the SLO is not a Monitor type, skip the validation
		log.Printf("[DEBUG] SLO type is: %v, skipping validation", val)
		return nil
	}

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	if attr, ok := diff.GetOk("monitor_ids"); ok {
		for _, v := range attr.(*schema.Set).List() {
			// Check that each monitor being added to the SLO exists
			if _, httpResponse, err := datadogClientV1.MonitorsApi.GetMonitor(authV1, int64(v.(int))); err != nil {
				return utils.TranslateClientError(err, httpResponse, "error finding monitor to add to SLO")
			}
		}
	}

	return nil
}

func buildServiceLevelObjectiveStructs(d *schema.ResourceData) (*datadogV1.ServiceLevelObjective, *datadogV1.ServiceLevelObjectiveRequest) {

	slo := datadogV1.NewServiceLevelObjectiveWithDefaults()
	slo.SetName(d.Get("name").(string))
	slo.SetType(datadogV1.SLOType(d.Get("type").(string)))
	slo.SetId(d.Id())

	slor := datadogV1.NewServiceLevelObjectiveRequestWithDefaults()
	slor.SetName(d.Get("name").(string))
	slor.SetType(datadogV1.SLOType(d.Get("type").(string)))

	if attr, ok := d.GetOk("description"); ok {
		slo.SetDescription(attr.(string))
		slor.SetDescription(attr.(string))
	}

	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if attr, ok := d.GetOk("monitor_ids"); ok {
			s := make([]int64, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, int64(v.(int)))
			}
			slo.SetMonitorIds(s)
			slor.SetMonitorIds(s)
		}
		if attr, ok := d.GetOk("groups"); ok {
			s := make([]string, 0)
			for _, v := range attr.(*schema.Set).List() {
				s = append(s, v.(string))
			}
			slo.SetGroups(s)
			slor.SetGroups(s)
		}
	default:
		// metric type
		if attr, ok := d.GetOk("query"); ok {
			queries := make([]map[string]interface{}, 0)
			raw := attr.([]interface{})
			for _, rawQuery := range raw {
				if query, ok := rawQuery.(map[string]interface{}); ok {
					queries = append(queries, query)
				}
			}
			if len(queries) >= 1 {
				// only use the first defined query
				slo.SetQuery(*datadogV1.NewServiceLevelObjectiveQuery(
					queries[0]["denominator"].(string),
					queries[0]["numerator"].(string)))
				slor.SetQuery(*datadogV1.NewServiceLevelObjectiveQuery(
					queries[0]["denominator"].(string),
					queries[0]["numerator"].(string)))
			}
		}
	}

	if attr, ok := d.GetOk("tags"); ok {
		s := make([]string, 0)
		for _, v := range attr.(*schema.Set).List() {
			s = append(s, v.(string))
		}
		slo.SetTags(s)
		slor.SetTags(s)
	}

	if _, ok := d.GetOk("thresholds"); ok {
		numThresholds := d.Get("thresholds.#").(int)
		sloThresholds := make([]datadogV1.SLOThreshold, 0)
		for i := 0; i < numThresholds; i++ {
			prefix := fmt.Sprintf("thresholds.%d.", i)
			t := datadogV1.NewSLOThresholdWithDefaults()

			if tf, ok := d.GetOk(prefix + "timeframe"); ok {
				t.SetTimeframe(datadogV1.SLOTimeframe(tf.(string)))
			}

			if targetValue, ok := d.GetOk(prefix + "target"); ok {
				if f, ok := floatOk(targetValue); ok {
					t.SetTarget(f)
				}
			}

			if warningValue, ok := d.GetOk(prefix + "warning"); ok {
				if f, ok := floatOk(warningValue); ok {
					t.SetWarning(f)
				}
			}

			if targetDisplayValue, ok := d.GetOk(prefix + "target_display"); ok {
				if s, ok := targetDisplayValue.(string); ok && strings.TrimSpace(s) != "" {
					t.SetTargetDisplay(strings.TrimSpace(targetDisplayValue.(string)))
				}
			}

			if warningDisplayValue, ok := d.GetOk(prefix + "warning_display"); ok {
				if s, ok := warningDisplayValue.(string); ok && strings.TrimSpace(s) != "" {
					t.SetWarningDisplay(strings.TrimSpace(warningDisplayValue.(string)))
				}
			}
			sloThresholds = append(sloThresholds, *t)
		}
		if len(sloThresholds) > 0 {
			slo.SetThresholds(sloThresholds)
			slor.SetThresholds(sloThresholds)
		}
	}

	return slo, slor
}

func floatOk(val interface{}) (float64, bool) {
	switch val := val.(type) {
	case float64:
		return val, true
	case *float64:
		return *val, true
	case string:
		f, err := strconv.ParseFloat(val, 64)
		if err == nil {
			return f, true
		}
	case *string:
		f, err := strconv.ParseFloat(*val, 64)
		if err == nil {
			return f, true
		}
	default:
		return 0, false
	}
	return 0, false
}

func resourceDatadogServiceLevelObjectiveCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	_, slor := buildServiceLevelObjectiveStructs(d)
	sloResp, httpResponse, err := datadogClientV1.ServiceLevelObjectivesApi.CreateSLO(authV1, *slor)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating service level objective")
	}
	if err := utils.CheckForUnparsed(sloResp); err != nil {
		return diag.FromErr(err)
	}

	slo := &sloResp.GetData()[0]
	d.SetId(slo.GetId())

	return updateSLOState(d, slo)
}

func resourceDatadogServiceLevelObjectiveRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	sloResp, httpresp, err := datadogClientV1.ServiceLevelObjectivesApi.GetSLO(authV1, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting service level objective")
	}
	if err := utils.CheckForUnparsed(sloResp); err != nil {
		return diag.FromErr(err)
	}

	return updateSLOStateFromRead(d, sloResp.Data)
}

func updateSLOState(d *schema.ResourceData, slo *datadogV1.ServiceLevelObjective) diag.Diagnostics {
	thresholds := make([]map[string]interface{}, 0)
	for _, threshold := range slo.GetThresholds() {
		t := map[string]interface{}{
			"timeframe": threshold.GetTimeframe(),
			"target":    threshold.GetTarget(),
		}
		if warning, ok := threshold.GetWarningOk(); ok {
			t["warning"] = warning
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
	tags = append(tags, slo.GetTags()...)

	if err := d.Set("name", slo.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", slo.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", slo.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("thresholds", thresholds); err != nil {
		return diag.FromErr(err)
	}
	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if len(slo.GetMonitorIds()) > 0 {
			if err := d.Set("monitor_ids", slo.GetMonitorIds()); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("groups", slo.GetGroups()); err != nil {
			return diag.FromErr(err)
		}
	default:
		// metric type
		query := make(map[string]interface{})
		q := slo.GetQuery()
		query["numerator"] = q.GetNumerator()
		query["denominator"] = q.GetDenominator()
		if err := d.Set("query", []map[string]interface{}{query}); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

// This duplicates updateSLOState for the SLOResponseData structure, which has mostly the same interface
func updateSLOStateFromRead(d *schema.ResourceData, slo *datadogV1.SLOResponseData) diag.Diagnostics {
	thresholds := make([]map[string]interface{}, 0)
	for _, threshold := range slo.GetThresholds() {
		t := map[string]interface{}{
			"timeframe": threshold.GetTimeframe(),
			"target":    threshold.GetTarget(),
		}
		if warning, ok := threshold.GetWarningOk(); ok {
			t["warning"] = warning
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
	tags = append(tags, slo.GetTags()...)

	if err := d.Set("name", slo.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", slo.GetDescription()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("type", slo.GetType()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("thresholds", thresholds); err != nil {
		return diag.FromErr(err)
	}
	switch slo.GetType() {
	case datadogV1.SLOTYPE_MONITOR:
		// monitor type
		if len(slo.GetMonitorIds()) > 0 {
			if err := d.Set("monitor_ids", slo.GetMonitorIds()); err != nil {
				return diag.FromErr(err)
			}
		}
		if err := d.Set("groups", slo.GetGroups()); err != nil {
			return diag.FromErr(err)
		}
	default:
		// metric type
		query := make(map[string]interface{})
		q := slo.GetQuery()
		query["numerator"] = q.GetNumerator()
		query["denominator"] = q.GetDenominator()
		if err := d.Set("query", []map[string]interface{}{query}); err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceDatadogServiceLevelObjectiveUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	slo, _ := buildServiceLevelObjectiveStructs(d)

	updatedSLO, httpResponse, err := datadogClientV1.ServiceLevelObjectivesApi.UpdateSLO(authV1, d.Id(), *slo)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating service level objective")
	}
	if err := utils.CheckForUnparsed(updatedSLO); err != nil {
		return diag.FromErr(err)
	}

	return updateSLOState(d, &updatedSLO.GetData()[0])
}

func resourceDatadogServiceLevelObjectiveDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	var err error

	var httpResponse *http.Response
	if d.Get("force_delete").(bool) {
		_, httpResponse, err = datadogClientV1.ServiceLevelObjectivesApi.DeleteSLO(authV1, d.Id(),
			*datadogV1.NewDeleteSLOOptionalParameters().WithForce("true"),
		)
	} else {
		_, httpResponse, err = datadogClientV1.ServiceLevelObjectivesApi.DeleteSLO(authV1, d.Id())
	}
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting service level objective")
	}
	return nil

}

// Ignore any diff that results from the mix of *_display string values from the
// DataDog API.
func suppressDataDogSLODisplayValueDiff(k, old, new string, d *schema.ResourceData) bool {
	sloType := d.Get("type")
	if sloType == datadogV1.SLOTYPE_MONITOR {
		// always suppress monitor type, this is controlled via API.
		return true
	}

	// metric type otherwise
	if old == "" || new == "" {
		// always suppress if not specified
		return true
	}

	return suppressDataDogFloatIntDiff(k, old, new, d)
}

func trimStateValue(val interface{}) string {
	return strings.TrimSpace(val.(string))
}

// Ignore any diff for trimmed state values.
func diffTrimmedValues(_, old, new string, _ *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}
