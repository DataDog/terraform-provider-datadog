package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogLogsIndexes() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list several existing logs indexes for use in other resources.",
		ReadContext: dataSourceDatadogLogsIndexesRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				// Computed values
				"logs_indexes": {
					Description: "List of logs indexes",
					Type:        schema.TypeList,
					Computed:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"name": {
								Description: "The name of the index.",
								Type:        schema.TypeString,
								Computed:    true,
							},
							"daily_limit": {
								Description: "The number of log events you can send in this index per day before you are rate-limited.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"retention_days": {
								Description: "The number of days before logs are deleted from this index.",
								Type:        schema.TypeInt,
								Computed:    true,
							},
							"filter": {
								Description: "Logs filter",
								Type:        schema.TypeList,
								Computed:    true,
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
								Computed:    true,
								Elem: &schema.Resource{
									Schema: dataSourceLogsIndexesExclusionFilterSchema,
								},
							},
						},
					},
				},
			}
		},
	}
}

var dataSourceLogsIndexesExclusionFilterSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the exclusion filter.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"is_enabled": {
		Description: "A boolean stating if the exclusion is active or not.",
		Type:        schema.TypeBool,
		Computed:    true,
	},
	"filter": {
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"query": {
					Description: "Only logs matching the filter criteria and the query of the parent index will be considered for this exclusion filter.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"sample_rate": {
					Description: "The fraction of logs excluded by the exclusion filter, when active.",
					Type:        schema.TypeFloat,
					Computed:    true,
				},
			},
		},
	},
}

func dataSourceDatadogLogsIndexesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	logsIndexes, httpresp, err := apiInstances.GetLogsIndexesApiV1().ListLogIndexes(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying log indexes")
	}

	diags := diag.Diagnostics{}
	tfLogsIndexes := make([]map[string]interface{}, len(logsIndexes.GetIndexes()))
	for i, l := range logsIndexes.GetIndexes() {
		if err := utils.CheckForUnparsed(l); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  fmt.Sprintf("skipping logs index with name: %s", l.GetName()),
				Detail:   fmt.Sprintf("logs index contains unparsed object: %v", err),
			})
			continue
		}

		tfLogsIndexes[i] = map[string]interface{}{
			"name":             l.GetName(),
			"daily_limit":      l.GetDailyLimit(),
			"retention_days":   l.GetNumRetentionDays(),
			"filter":           buildTerraformIndexFilter(l.GetFilter()),
			"exclusion_filter": buildTerraformExclusionFilters(l.GetExclusionFilters()),
		}
	}
	if err := d.Set("logs_indexes", tfLogsIndexes); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("log-indexes")

	return diags
}
