package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func dataSourceDatadogDashboard() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing dashboard, for use in other resources. In particular, it can be used in a monitor message to link to a specific dashboard.",
		ReadContext: dataSourceDatadogDashboardRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description:  "The dashboard name to search for. Must only match one dashboard.",
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: validation.StringIsNotEmpty,
				},
				// Computed values
				"title": {
					Description: "The name of the dashboard.",
					Type:        schema.TypeString,
					Computed:    true,
				},
				"url": {
					Description: "The URL to a specific dashboard.",
					Type:        schema.TypeString,
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *retry.RetryError {
		dashResponse, httpresp, err := apiInstances.GetDashboardsApiV1().ListDashboards(auth)
		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 504 || httpresp.StatusCode == 502) {
				return retry.RetryableError(utils.TranslateClientError(err, httpresp, "error querying dashboard, retrying"))
			}
			return retry.NonRetryableError(utils.TranslateClientError(err, httpresp, "error querying dashboard"))
		}

		searchedName := d.Get("name")
		var foundDashes []datadogV1.DashboardSummaryDefinition

		for _, dash := range dashResponse.GetDashboards() {
			if dash.GetTitle() == searchedName {
				foundDashes = append(foundDashes, dash)
			}
		}

		if len(foundDashes) == 0 {
			return retry.NonRetryableError(fmt.Errorf("couldn't find a dashboard named %s", searchedName))
		} else if len(foundDashes) > 1 {
			return retry.NonRetryableError(fmt.Errorf("%s returned more than one dashboard", searchedName))
		}

		if err := utils.CheckForUnparsed(foundDashes[0]); err != nil {
			return retry.NonRetryableError(err)
		}

		d.SetId(foundDashes[0].GetId())
		d.Set("url", foundDashes[0].GetUrl())
		d.Set("title", foundDashes[0].GetTitle())

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
