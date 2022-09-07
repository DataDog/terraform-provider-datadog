package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogLogsPipelines() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to list all existing logs pipelines for use in other resources.",
		ReadContext: dataSourceDatadogLogsPipelinesRead,
		Schema: map[string]*schema.Schema{
			// Computed values
			"logs_pipelines": {
				Description: "List of logs pipelines",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Description: "ID of the pipeline",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"filter": {
							Description: "Pipelines filter",
							Type:        schema.TypeList,
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"query": {
										Description: "Pipeline filter criteria.",
										Type:        schema.TypeString,
										Required:    true,
									},
								},
							},
						},
						"name": {
							Description: "The name of the pipeline.",
							Type:        schema.TypeString,
							Computed:    true,
						},
						"is_enabled": {
							Description: "Whether or not the pipeline is enabled.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"is_read_only": {
							Description: "Whether or not the pipeline can be edited.",
							Type:        schema.TypeBool,
							Computed:    true,
						},
						"type": {
							Description: "Whether or not the pipeline can be edited.",
							Type:        schema.TypeString,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func dataSourceDatadogLogsPipelinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	logsPipelines, httpresp, err := apiInstances.GetLogsPipelinesApiV1().ListLogsPipelines(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying log pipelines")
	}
	if err := utils.CheckForUnparsed(logsPipelines); err != nil {
		return diag.FromErr(err)
	}

	tflogsPipelines := make([]map[string]interface{}, len(logsPipelines))
	for i, pipeline := range logsPipelines {
		tflogsPipelines[i] = map[string]interface{}{
			"name":         pipeline.Name,
			"id":           pipeline.Id,
			"filter":       buildTerraformIndexFilter(*pipeline.Filter),
			"is_enabled":   pipeline.IsEnabled,
			"is_read_only": pipeline.IsReadOnly,
			"type":         pipeline.Type,
		}
	}
	if err := d.Set("logs_pipelines", tflogsPipelines); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("log-pipelines")
	return nil
}
