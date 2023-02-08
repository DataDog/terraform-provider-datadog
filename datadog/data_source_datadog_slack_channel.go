package datadog

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogIntegrationSlack() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve a Datadog Slack Integration.",
		ReadContext: dataSourceDatadogIntegrationSlackRead,

		Schema: map[string]*schema.Schema{
			"channel_name": {
				Description:  "Your Service name in Slack.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceDatadogIntegrationSlackRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		searchedName := d.Get("channel_name")

		resp, httpresp, err := apiInstances.GetSlackIntegrationApiV1().GetSlackIntegrationChannels(auth, searchedName.(string))
		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 504 || httpresp.StatusCode == 502) {
				return resource.RetryableError(utils.TranslateClientError(err, httpresp, "error querying pagerduty integrations, retrying"))
			}
			return resource.NonRetryableError(utils.TranslateClientError(err, httpresp, "error querying pagerduty integrations"))
		}

		if serviceName, ok := resp.GetServiceNameOk(); !ok {
			return resource.NonRetryableError(fmt.Errorf("couldn't find a pagerduty integration named %s", serviceName))
		} else {
			d.SetId(*serviceName)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil

}
