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

func dataSourceDatadogIntegrationPagerdutySO() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve individual Service Objects of Datadog - PagerDuty integrations. Note that the Datadog - PagerDuty integration must be activated in the Datadog UI in order for this resource to be usable.",
		ReadContext: dataSourceDatadogIntegrationPagerdutySORead,

		Schema: map[string]*schema.Schema{
			"service_name": {
				Description:  "Your Service name in PagerDuty.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceDatadogIntegrationPagerdutySORead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		searchedName := d.Get("service_name")

		resp, httpresp, err := apiInstances.GetPagerDutyIntegrationApiV1().GetPagerDutyIntegrationService(auth, searchedName.(string))

		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 504 || httpresp.StatusCode == 502) {
				return resource.RetryableError(utils.TranslateClientError(err, httpresp, "error querying pagerduty integrations, retrying"))
			}
			if httpresp != nil && httpresp.StatusCode == 404 {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(utils.TranslateClientError(err, httpresp, "error querying pagerduty integrations"))
		}

		if serviceName, ok := resp.GetServiceNameOk(); !ok {
			d.SetId("")
			return resource.NonRetryableError(fmt.Errorf("couldn't find a pagerduty integration service named %s", *serviceName))
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
