package datadog

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

var indexSchema = map[string]*schema.Schema{
	"name": {Type: schema.TypeString, Required: true},
	"filter": {
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"query": {Type: schema.TypeString, Required: true},
			},
		},
	},
	"exclusion_filter": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: exclusionFilterSchema,
		},
	},
}

var exclusionFilterSchema = map[string]*schema.Schema{
	"name":       {Type: schema.TypeString, Optional: true},
	"is_enabled": {Type: schema.TypeBool, Optional: true},
	"filter": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"query":       {Type: schema.TypeString, Optional: true},
				"sample_rate": {Type: schema.TypeFloat, Optional: true},
			},
		},
	},
}

func resourceDatadogLogsIndex() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogLogsIndexCreate,
		Update: resourceDatadogLogsIndexUpdate,
		Read:   resourceDatadogLogsIndexRead,
		Delete: resourceDatadogLogsIndexDelete,
		Exists: resourceDatadogLogsIndexExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: indexSchema,
	}
}

func resourceDatadogLogsIndexCreate(d *schema.ResourceData, meta interface{}) error {
	// This is a bit of a hack to ensure we fail fast if an index is about to be created, and
	// to ensure we provide a useful error message (and don't panic)
	// Indexes can only be updated, and the id is only set in the state if it was already imported
	if _, ok := d.GetOk("id"); !ok {
		return fmt.Errorf("logs index creation is not allowed, please import the index first. index_name: %s", d.Get("name").(string))
	}
	return resourceDatadogLogsIndexUpdate(d, meta)
}

func resourceDatadogLogsIndexRead(d *schema.ResourceData, meta interface{}) error {
	ddIndex, err := meta.(*datadog.Client).GetLogsIndex(d.Id())
	if err != nil {
		return err
	}
	if err = d.Set("name", ddIndex.GetName()); err != nil {
		return err
	}
	if err = d.Set("filter", buildTerraformFilter(ddIndex.Filter)); err != nil {
		return err
	}
	if err = d.Set("exclusion_filter", buildTerraformExclusionFilters(ddIndex.ExclusionFilters)); err != nil {
		return err
	}
	return nil
}

func resourceDatadogLogsIndexUpdate(d *schema.ResourceData, meta interface{}) error {
	ddIndex, err := buildDatadogIndex(d)
	if err != nil {
		return err
	}
	client := meta.(*datadog.Client)
	tfName := d.Get("name").(string)
	if _, err := client.UpdateLogsIndex(tfName, ddIndex); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return fmt.Errorf("logs index creation is not allowed, index_name: %s", tfName)
		}
		return fmt.Errorf("error updating logs index: (%s)", err.Error())
	}
	d.SetId(tfName)
	return resourceDatadogLogsIndexRead(d, meta)
}

func resourceDatadogLogsIndexDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceDatadogLogsIndexExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	client := meta.(*datadog.Client)
	if _, err := client.GetLogsIndex(d.Id()); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func buildDatadogIndex(d *schema.ResourceData) (*datadog.LogsIndex, error) {
	var ddIndex datadog.LogsIndex
	if tfFilter := d.Get("filter").([]interface{}); len(tfFilter) > 0 {
		ddIndex.SetFilter(buildDatadogFilter(tfFilter[0].(map[string]interface{})))
	}

	ddIndex.ExclusionFilters = buildDatadogExclusionFilters(d.Get("exclusion_filter").([]interface{}))
	return &ddIndex, nil
}

func buildDatadogExclusionFilters(tfEFilters []interface{}) []datadog.ExclusionFilter {
	ddEFilters := make([]datadog.ExclusionFilter, len(tfEFilters))
	for i, tfEFilter := range tfEFilters {
		ddEFilters[i] = buildDatadogExclusionFilter(tfEFilter.(map[string]interface{}))
	}
	return ddEFilters
}

func buildDatadogExclusionFilter(tfEFilter map[string]interface{}) datadog.ExclusionFilter {
	ddEFilter := datadog.ExclusionFilter{}
	if tfName, exists := tfEFilter["name"].(string); exists {
		ddEFilter.SetName(tfName)
	}
	if tfIsEnabled, exists := tfEFilter["is_enabled"].(bool); exists {
		ddEFilter.SetIsEnabled(tfIsEnabled)
	}
	if tfFs, exists := tfEFilter["filter"].([]interface{}); exists && len(tfFs) > 0 {
		tfFilter := tfFs[0].(map[string]interface{})
		ddFilter := datadog.Filter{}
		if tfQuery, exist := tfFilter["query"].(string); exist {
			ddFilter.SetQuery(tfQuery)
		}
		if tfSampleRate, exist := tfFilter["sample_rate"].(float64); exist {
			ddFilter.SetSampleRate(tfSampleRate)
		}
		ddEFilter.SetFilter(ddFilter)
	}
	return ddEFilter
}

func buildTerraformExclusionFilters(ddEFilters []datadog.ExclusionFilter) []map[string]interface{} {
	tfEFilters := make([]map[string]interface{}, len(ddEFilters))
	for i, ddEFilter := range ddEFilters {
		tfEFilters[i] = buildTerraformExclusionFilter(ddEFilter)
	}
	return tfEFilters
}

func buildTerraformExclusionFilter(ddEFilter datadog.ExclusionFilter) map[string]interface{} {
	tfEFilter := make(map[string]interface{})
	ddFilter := ddEFilter.GetFilter()
	tfFilter := map[string]interface{}{
		"query":       ddFilter.GetQuery(),
		"sample_rate": ddFilter.GetSampleRate(),
	}
	tfEFilter["filter"] = []map[string]interface{}{tfFilter}
	tfEFilter["name"] = ddEFilter.GetName()
	tfEFilter["is_enabled"] = ddEFilter.GetIsEnabled()
	return tfEFilter
}
