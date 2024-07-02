package datadog

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func dataSourceDatadogPowerpack() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing powerpack for use in other resources.",
		ReadContext: dataSourceDatadogPowerpackRead,

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description: "A powerpack name, must match exactly one powerpack.",
					Type:        schema.TypeString,
					Optional:    true,
				},

				// Computed values
				"id": {
					Description: "ID of the powerpack",
					Type:        schema.TypeString,
					Computed:    true,
				},
			}
		},
	}
}

func dataSourceDatadogPowerpackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	err := retry.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *retry.RetryError {
		packsResponse, httpresp, err := apiInstances.GetPowerpackApiV2().ListPowerpacks(auth)
		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 504 || httpresp.StatusCode == 502) {
				return retry.RetryableError(utils.TranslateClientError(err, httpresp, "error querying dashboard, retrying"))
			}
			return retry.NonRetryableError(utils.TranslateClientError(err, httpresp, "error querying dashboard"))
		}

		searchedName := d.Get("name")
		var foundPowerpackIds = []string{}

		for _, pp := range packsResponse.GetData() {
			if pp.GetAttributes().Name == searchedName {
				foundPowerpackIds = append(foundPowerpackIds, pp.GetId())
			}
		}

		if len(foundPowerpackIds) == 0 {
			return retry.NonRetryableError(fmt.Errorf("couldn't find a powerpack named %s", searchedName))
		} else if len(foundPowerpackIds) > 1 {
			return retry.NonRetryableError(fmt.Errorf("%s returned more than one powerpack", searchedName))
		}

		d.SetId(foundPowerpackIds[0])
		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
