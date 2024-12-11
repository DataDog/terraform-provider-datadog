package datadog

import (
	"context"
	"log"
	"regexp"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var logsIndexMutex = sync.Mutex{}

var indexSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the index. Index names cannot be modified after creation. If this value is changed, a new index will be created.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"disable_daily_limit": {
		Description: "If true, sets the daily_limit value to null and the index is not limited on a daily basis (any specified daily_limit value in the request is ignored). If false or omitted, the index's current daily_limit is maintained.",
		Type:        schema.TypeBool,
		Optional:    true,
		Computed:    true,
	},
	"daily_limit": {
		Description:      "The number of log events you can send in this index per day before you are rate-limited.",
		Type:             schema.TypeInt,
		Optional:         true,
		ValidateFunc:     validation.IntAtLeast(1),
		DiffSuppressFunc: suppressDiffWhenDisabledDailyLimit,
	},
	"daily_limit_reset": {
		Description:      "Object containing options to override the default daily limit reset time.",
		Type:             schema.TypeList,
		Optional:         true,
		Computed:         true,
		MaxItems:         1,
		DiffSuppressFunc: suppressDiffWhenDisabledDailyLimit,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"reset_time": {
					Description:  "String in `HH:00` format representing the time of day the daily limit should be reset. The hours must be between 00 and 23 (inclusive).",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([0-1][0-9]|2[0-3]):00$`), "must be in HH:00 format with the hours (HH) between 00 and 23 (inclusive)"),
				},
				"reset_utc_offset": {
					Description:  "String in `(-|+)HH:00` format representing the UTC offset to apply to the given reset time. The hours must be between -12 and +14 (inclusive).",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringMatch(regexp.MustCompile(`^(-(0[0-9]|1[0-2])|\+(0[0-9]|1[0-4])):00$`), "must be in +HH:00 or -HH:00 format with the hours (HH) between -12 and +14 (inclusive)"),
				},
			},
		},
	},
	"daily_limit_warning_threshold_percentage": {
		Description:      "A percentage threshold of the daily quota at which a Datadog warning event is generated.",
		Type:             schema.TypeFloat,
		Optional:         true,
		Computed:         true,
		ValidateFunc:     validation.FloatBetween(50, 99.99),
		DiffSuppressFunc: suppressDiffWhenDisabledDailyLimit,
	},
	"retention_days": {
		Description: "The number of days logs are stored in Standard Tier before aging into the Flex Tier or being deleted from the index.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"flex_retention_days": {
		Description: "The total number of days logs are stored in Standard and Flex Tier before being deleted from the index.",
		Type:        schema.TypeInt,
		Optional:    true,
		Computed:    true,
	},
	"filter": {
		Description: "Logs filter",
		Type:        schema.TypeList,
		Required:    true,
		MaxItems:    1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"query": {
					Description: "Logs filter criteria. Only logs matching this filter criteria are considered for this index.",
					Type:        schema.TypeString,
					Required:    true,
				},
			},
		},
	},
	"exclusion_filter": {
		Description: "List of exclusion filters.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: exclusionFilterSchema,
		},
	},
}

var exclusionFilterSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the exclusion filter.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"is_enabled": {
		Description: "A boolean stating if the exclusion is active or not.",
		Type:        schema.TypeBool,
		Optional:    true,
	},
	"filter": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"query": {
					Description: "Only logs matching the filter criteria and the query of the parent index will be considered for this exclusion filter.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"sample_rate": {
					Description: "The fraction of logs excluded by the exclusion filter, when active.",
					Type:        schema.TypeFloat,
					Optional:    true,
				},
			},
		},
	},
}

func suppressDiffWhenDisabledDailyLimit(_, old, new string, d *schema.ResourceData) bool {
	// Ignore diff if disable_daily_limit is set to true
	if v, ok := d.GetOk("disable_daily_limit"); ok && v.(bool) {
		log.Printf("[DEBUG] Ignoring change because disable_daily_limit is set to true on index %s.", d.Get("name"))
		return true
	}
	return false
}

func resourceDatadogLogsIndex() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Index API resource. This can be used to create and manage Datadog logs indexes.  \n**Note:** It is not possible to delete logs indexes through Terraform, so an index remains in your account after the resource is removed from your terraform config. Reach out to support to delete a logs index.",
		CreateContext: resourceDatadogLogsIndexCreate,
		UpdateContext: resourceDatadogLogsIndexUpdate,
		ReadContext:   resourceDatadogLogsIndexRead,
		DeleteContext: resourceDatadogLogsIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return indexSchema
		},
	}
}

func resourceDatadogLogsIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logsIndexMutex.Lock()
	defer logsIndexMutex.Unlock()

	ddIndex := buildDatadogIndexCreateRequest(d)
	createdIndex, httpResponse, err := apiInstances.GetLogsIndexesApiV1().CreateLogsIndex(auth, *ddIndex)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating logs index")
	}
	if err := utils.CheckForUnparsed(createdIndex); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(createdIndex.GetName())

	return updateLogsIndexState(d, &createdIndex)
}

func updateLogsIndexState(d *schema.ResourceData, index *datadogV1.LogsIndex) diag.Diagnostics {
	if err := d.Set("name", index.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("disable_daily_limit", !index.HasDailyLimit()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("daily_limit", index.GetDailyLimit()); err != nil {
		return diag.FromErr(err)
	}
	if daily_limit_reset, ok := index.GetDailyLimitResetOk(); ok {
		if err := d.Set("daily_limit_reset", buildTerraformIndexDailyLimitReset(*daily_limit_reset)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("daily_limit_reset", nil); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("daily_limit_warning_threshold_percentage", index.GetDailyLimitWarningThresholdPercentage()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("retention_days", index.GetNumRetentionDays()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("flex_retention_days", index.GetNumFlexLogsRetentionDays()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("filter", buildTerraformIndexFilter(index.GetFilter())); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("exclusion_filter", buildTerraformExclusionFilters(index.GetExclusionFilters())); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogLogsIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	ddIndex, httpresp, err := apiInstances.GetLogsIndexesApiV1().GetLogsIndex(auth, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting logs index")
	}
	if err := utils.CheckForUnparsed(ddIndex); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsIndexState(d, &ddIndex)
}

func resourceDatadogLogsIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logsIndexMutex.Lock()
	defer logsIndexMutex.Unlock()

	ddIndex := buildDatadogIndexUpdateRequest(d)
	tfName := d.Get("name").(string)
	updatedIndex, httpResponse, err := apiInstances.GetLogsIndexesApiV1().UpdateLogsIndex(auth, tfName, *ddIndex)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating logs index")
	}
	if err := utils.CheckForUnparsed(updatedIndex); err != nil {
		return diag.FromErr(err)
	}
	return updateLogsIndexState(d, &updatedIndex)
}

func resourceDatadogLogsIndexDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func buildDatadogIndexUpdateRequest(d *schema.ResourceData) *datadogV1.LogsIndexUpdateRequest {
	var ddIndex datadogV1.LogsIndexUpdateRequest
	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		ddIndex.SetFilter(*buildDatadogIndexFilter(tfFilter[0].(map[string]interface{})))
	}

	if v, ok := d.GetOk("daily_limit"); ok {
		ddIndex.SetDailyLimit(int64(v.(int)))
	}
	if tfDailyLimitReset := d.Get("daily_limit_reset").([]interface{}); len(tfDailyLimitReset) > 0 {
		ddIndex.SetDailyLimitReset(*buildDatadogIndexDailyLimitReset(tfDailyLimitReset[0].(map[string]interface{})))
	}
	if v, ok := d.GetOk("daily_limit_warning_threshold_percentage"); ok {
		ddIndex.SetDailyLimitWarningThresholdPercentage(float64(v.(float64)))
	}
	if v, ok := d.GetOk("disable_daily_limit"); ok {
		ddIndex.SetDisableDailyLimit(v.(bool))
	}
	if !d.GetRawConfig().GetAttr("retention_days").IsNull() {
		ddIndex.SetNumRetentionDays(int64(d.Get("retention_days").(int)))
	}
	if v, ok := d.GetOk("flex_retention_days"); ok {
		ddIndex.SetNumFlexLogsRetentionDays(int64(v.(int)))
	}

	ddIndex.ExclusionFilters = *buildDatadogExclusionFilters(d.Get("exclusion_filter").([]interface{}))
	return &ddIndex
}

func buildDatadogIndexCreateRequest(d *schema.ResourceData) *datadogV1.LogsIndex {
	var ddIndex datadogV1.LogsIndex

	ddIndex.SetName(d.Get("name").(string))

	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		ddIndex.SetFilter(*buildDatadogIndexFilter(tfFilter[0].(map[string]interface{})))
	}
	if v, ok := d.GetOk("daily_limit"); ok {
		ddIndex.SetDailyLimit(int64(v.(int)))
	}
	if tfDailyLimitReset := d.Get("daily_limit_reset").([]interface{}); len(tfDailyLimitReset) > 0 {
		ddIndex.SetDailyLimitReset(*buildDatadogIndexDailyLimitReset(tfDailyLimitReset[0].(map[string]interface{})))
	}
	if v, ok := d.GetOk("daily_limit_warning_threshold_percentage"); ok {
		ddIndex.SetDailyLimitWarningThresholdPercentage(float64(v.(float64)))
	}
	if v, ok := d.GetOk("retention_days"); ok {
		ddIndex.SetNumRetentionDays(int64(v.(int)))
	}
	if v, ok := d.GetOk("flex_retention_days"); ok {
		ddIndex.SetNumFlexLogsRetentionDays(int64(v.(int)))
		if _, isRetentionSet := ddIndex.GetNumRetentionDaysOk(); !isRetentionSet {
			// NOTE: Null retention is not an acceptable value on creation with flex. Must be explicitly 0.
			ddIndex.SetNumRetentionDays(0)
		}
	}
	ddIndex.ExclusionFilters = *buildDatadogExclusionFilters(d.Get("exclusion_filter").([]interface{}))
	return &ddIndex
}

func buildDatadogIndexFilter(tfFilter map[string]interface{}) *datadogV1.LogsFilter {
	ddFilter := datadogV1.NewLogsFilter()
	if tfQuery, exists := tfFilter["query"].(string); exists {
		ddFilter.SetQuery(tfQuery)
	}
	return ddFilter
}

func buildDatadogIndexDailyLimitReset(tfDailyLimitReset map[string]interface{}) *datadogV1.LogsDailyLimitReset {
	ddDailyLimitReset := datadogV1.NewLogsDailyLimitReset()
	if tfResetTime, exists := tfDailyLimitReset["reset_time"].(string); exists {
		ddDailyLimitReset.SetResetTime(tfResetTime)
	}
	if tfResetUtcOffset, exists := tfDailyLimitReset["reset_utc_offset"].(string); exists {
		ddDailyLimitReset.SetResetUtcOffset(tfResetUtcOffset)
	}
	return ddDailyLimitReset
}

func buildTerraformIndexFilter(ddFilter datadogV1.LogsFilter) *[]map[string]interface{} {
	tfFilter := map[string]interface{}{
		"query": ddFilter.GetQuery(),
	}
	return &[]map[string]interface{}{tfFilter}
}

func buildTerraformIndexDailyLimitReset(ddDailyLimitReset datadogV1.LogsDailyLimitReset) *[]map[string]interface{} {
	tfDailyLimitReset := map[string]interface{}{
		"reset_time":       ddDailyLimitReset.GetResetTime(),
		"reset_utc_offset": ddDailyLimitReset.GetResetUtcOffset(),
	}
	return &[]map[string]interface{}{tfDailyLimitReset}
}

func buildDatadogExclusionFilters(tfEFilters []interface{}) *[]datadogV1.LogsExclusion {
	ddEFilters := make([]datadogV1.LogsExclusion, 0)
	for _, tfEFilter := range tfEFilters {
		if v, ok := tfEFilter.(map[string]interface{}); ok {
			ddEFilters = append(ddEFilters, *buildDatadogExclusionFilter(v))
		}
	}
	return &ddEFilters
}

func buildDatadogExclusionFilter(tfEFilter map[string]interface{}) *datadogV1.LogsExclusion {
	ddEFilter := datadogV1.NewLogsExclusionWithDefaults()
	if tfName, exists := tfEFilter["name"].(string); exists {
		ddEFilter.SetName(tfName)
	}
	if tfIsEnabled, exists := tfEFilter["is_enabled"].(bool); exists {
		ddEFilter.SetIsEnabled(tfIsEnabled)
	}
	if tfFs, exists := tfEFilter["filter"].([]interface{}); exists && len(tfFs) > 0 {
		tfFilter := tfFs[0].(map[string]interface{})
		ddFilter := datadogV1.NewLogsExclusionFilterWithDefaults()
		if tfQuery, exist := tfFilter["query"].(string); exist {
			ddFilter.SetQuery(tfQuery)
		}
		if tfSampleRate, exist := tfFilter["sample_rate"].(float64); exist {
			ddFilter.SetSampleRate(tfSampleRate)
		}
		ddEFilter.SetFilter(*ddFilter)
	}
	return ddEFilter
}

func buildTerraformExclusionFilters(ddEFilters []datadogV1.LogsExclusion) *[]map[string]interface{} {
	tfEFilters := make([]map[string]interface{}, len(ddEFilters))
	for i, ddEFilter := range ddEFilters {
		tfEFilters[i] = *buildTerraformExclusionFilter(ddEFilter)
	}
	return &tfEFilters
}

func buildTerraformExclusionFilter(ddEFilter datadogV1.LogsExclusion) *map[string]interface{} {
	tfEFilter := make(map[string]interface{})
	ddFilter := ddEFilter.GetFilter()
	tfFilter := map[string]interface{}{
		"query":       ddFilter.GetQuery(),
		"sample_rate": ddFilter.GetSampleRate(),
	}
	tfEFilter["filter"] = []map[string]interface{}{tfFilter}
	tfEFilter["name"] = ddEFilter.GetName()
	tfEFilter["is_enabled"] = ddEFilter.GetIsEnabled()
	return &tfEFilter
}
