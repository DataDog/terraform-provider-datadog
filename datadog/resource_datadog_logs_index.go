package datadog

import (
	"context"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var indexSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the index.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"filter": {
		Description: "Logs filter",
		Type:        schema.TypeList,
		Required:    true,
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

func resourceDatadogLogsIndex() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Logs Index API resource. This can be used to create and manage Datadog logs indexes.",
		CreateContext: resourceDatadogLogsIndexCreate,
		UpdateContext: resourceDatadogLogsIndexUpdate,
		ReadContext:   resourceDatadogLogsIndexRead,
		DeleteContext: resourceDatadogLogsIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: indexSchema,
	}
}

func resourceDatadogLogsIndexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// This is a bit of a hack to ensure we fail fast if an index is about to be created, and
	// to ensure we provide a useful error message (and don't panic)
	// Indexes can only be updated, and the id is only set in the state if it was already imported
	if _, ok := d.GetOk("id"); !ok {
		return diag.Errorf("logs index creation is not allowed, please import the index first. index_name: %s", d.Get("name").(string))
	}
	return resourceDatadogLogsIndexUpdate(ctx, d, meta)
}

func updateLogsIndexState(d *schema.ResourceData, index *datadogV1.LogsIndex) diag.Diagnostics {
	if err := d.Set("name", index.GetName()); err != nil {
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	ddIndex, httpresp, err := datadogClientV1.LogsIndexesApi.GetLogsIndex(authV1, d.Id())
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, "error getting logs index")
	}
	return updateLogsIndexState(d, &ddIndex)
}

func resourceDatadogLogsIndexUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	ddIndex, err := buildDatadogIndex(d)
	if err != nil {
		return diag.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	tfName := d.Get("name").(string)
	updatedIndex, _, err := datadogClientV1.LogsIndexesApi.UpdateLogsIndex(authV1, tfName, *ddIndex)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return diag.Errorf("logs index creation is not allowed, index_name: %s", tfName)
		}
		return utils.TranslateClientErrorDiag(err, "error updating logs index")
	}
	d.SetId(tfName)
	return updateLogsIndexState(d, &updatedIndex)
}

func resourceDatadogLogsIndexDelete(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func buildDatadogIndex(d *schema.ResourceData) (*datadogV1.LogsIndexUpdateRequest, error) {
	var ddIndex datadogV1.LogsIndexUpdateRequest
	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		ddIndex.SetFilter(*buildDatadogIndexFilter(tfFilter[0].(map[string]interface{})))
	}

	ddIndex.ExclusionFilters = buildDatadogExclusionFilters(d.Get("exclusion_filter").([]interface{}))
	return &ddIndex, nil
}

func buildDatadogIndexFilter(tfFilter map[string]interface{}) *datadogV1.LogsFilter {
	ddFilter := datadogV1.NewLogsFilter()
	if tfQuery, exists := tfFilter["query"].(string); exists {
		ddFilter.SetQuery(tfQuery)
	}
	return ddFilter
}

func buildTerraformIndexFilter(ddFilter datadogV1.LogsFilter) *[]map[string]interface{} {
	tfFilter := map[string]interface{}{
		"query": ddFilter.GetQuery(),
	}
	return &[]map[string]interface{}{tfFilter}
}

func buildDatadogExclusionFilters(tfEFilters []interface{}) *[]datadogV1.LogsExclusion {
	ddEFilters := make([]datadogV1.LogsExclusion, len(tfEFilters))
	for i, tfEFilter := range tfEFilters {
		ddEFilters[i] = *buildDatadogExclusionFilter(tfEFilter.(map[string]interface{}))
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
