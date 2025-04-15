package datadog

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogMetricTags() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve tags associated with a metric to use in other resources.",
		ReadContext: dataSourceDatadogUserRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"metric": {
					Description: "The metric for which to fetch tags.",
					Type:        schema.TypeString,
					Required:    true,
				},
				// Computed values
				"tags": {
					Description: "The tags associated with the metric.",
					Type:        schema.TypeList,
					Computed:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func dataSourceDatadogMetricTagsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	metric := d.Get("metric").(string)
	res, httpresp, err := apiInstances.GetMetricsApiV2().ListTagsByMetricName(auth, metric)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error querying metric tags")
	}

	tagsData := res.GetData()
	tags := tagsData.Attributes.GetTags()

	d.SetId(tagsData.GetId())

	if err := d.Set("metric", metric); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", tags); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
