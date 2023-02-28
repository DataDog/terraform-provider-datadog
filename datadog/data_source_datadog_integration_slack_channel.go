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

func dataSourceDatadogIntegrationSlackChannel() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve individual Slack channel objects of Datadog Slack integrations. " +
			"Note that the Datadog Slack Integration must be activated in the Datadog UI in order for this resource to be usable.",
		ReadContext: dataSourceDatadogIntegrationSlackChannelRead,
		Schema: map[string]*schema.Schema{
			"account_name": {
				Description:  "Your Account name in Slack.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"channel_name": {
				Description:  "Your Channel name in Slack.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
		},
	}
}

func dataSourceDatadogIntegrationSlackChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	err := resource.RetryContext(ctx, d.Timeout(schema.TimeoutRead), func() *resource.RetryError {
		searchedAccountName := d.Get("account_name")
		searchedChannelName := d.Get("channel_name")

		resp, httpresp, err := apiInstances.GetSlackIntegrationApiV1().GetSlackIntegrationChannel(auth, searchedAccountName.(string), searchedChannelName.(string))

		if err != nil {
			if httpresp != nil && (httpresp.StatusCode == 504 || httpresp.StatusCode == 502) {
				return resource.RetryableError(utils.TranslateClientError(err, httpresp, "error querying slack integrations, retrying"))
			}
			if httpresp != nil && httpresp.StatusCode == 404 {
				d.SetId("")
				return nil
			}
			return resource.NonRetryableError(utils.TranslateClientError(err, httpresp, "error querying slack integrations"))
		}

		if channelName, ok := resp.GetNameOk(); !ok {
			d.SetId("")
			return resource.NonRetryableError(fmt.Errorf("couldn't find a slack integration channel named %s", *channelName))
		} else {
			d.SetId(*channelName)
		}

		return nil
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}
