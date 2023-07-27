package datadog

import (
	"context"
	"fmt"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// creating/modifying/deleting Slack Channel integration in parallel on one account
// is unsupported by the API right now; therefore we use the mutex to only operate on one at a time
var integrationSlackChannelMutex = sync.Mutex{}

func resourceDatadogIntegrationSlackChannel() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with the Datadog Slack channel API",
		CreateContext: resourceDatadogIntegrationSlackChannelCreate,
		ReadContext:   resourceDatadogIntegrationSlackChannelRead,
		UpdateContext: resourceDatadogIntegrationSlackChannelUpdate,
		DeleteContext: resourceDatadogIntegrationSlackChannelDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"channel_name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Slack channel name.",
				},
				"account_name": {
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
					Description: "Slack account name.",
				},
				"display": {
					Type:        schema.TypeList,
					Required:    true,
					Description: "Configuration options for what is shown in an alert event message.",
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"message": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Show the main body of the alert event.",
								Default:     true,
							},
							"notified": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Show the list of @-handles in the alert event.",
								Default:     true,
							},
							"snapshot": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Show the alert event's snapshot image.",
								Default:     true,
							},
							"tags": {
								Type:        schema.TypeBool,
								Optional:    true,
								Description: "Show the scopes on which the monitor alerted.",
								Default:     true,
							},
						},
					},
				},
			}
		},
	}
}

func buildDatadogSlackChannel(d *schema.ResourceData) *datadogV1.SlackIntegrationChannel {
	datadogSlackChannel := datadogV1.NewSlackIntegrationChannelWithDefaults()

	if v, ok := d.GetOk("channel_name"); ok {
		datadogSlackChannel.SetName(v.(string))
	}

	resultDisplay := datadogV1.NewSlackIntegrationChannelDisplayWithDefaults()
	resultDisplay.SetMessage(d.Get("display.0.message").(bool))
	resultDisplay.SetNotified(d.Get("display.0.notified").(bool))
	resultDisplay.SetSnapshot(d.Get("display.0.snapshot").(bool))
	resultDisplay.SetTags(d.Get("display.0.tags").(bool))

	datadogSlackChannel.SetDisplay(*resultDisplay)

	return datadogSlackChannel
}

func resourceDatadogIntegrationSlackChannelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationSlackChannelMutex.Lock()
	defer integrationSlackChannelMutex.Unlock()

	ddSlackChannel := buildDatadogSlackChannel(d)
	accountName := d.Get("account_name").(string)

	createdChannel, httpresp, err := apiInstances.GetSlackIntegrationApiV1().CreateSlackIntegrationChannel(auth, accountName, *ddSlackChannel)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating slack channel")
	}
	if err := utils.CheckForUnparsed(createdChannel); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s:%s", accountName, ddSlackChannel.GetName()))
	return updateSlackChannelState(d, &createdChannel)
}

func resourceDatadogIntegrationSlackChannelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountName, channelName, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	slackChannel, httpresp, err := apiInstances.GetSlackIntegrationApiV1().GetSlackIntegrationChannel(auth, accountName, channelName)
	if err != nil {
		if httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting slack channel")
	}
	if err := utils.CheckForUnparsed(slackChannel); err != nil {
		return diag.FromErr(err)
	}

	return updateSlackChannelState(d, &slackChannel)
}

func resourceDatadogIntegrationSlackChannelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationSlackChannelMutex.Lock()
	defer integrationSlackChannelMutex.Unlock()

	ddObject := buildDatadogSlackChannel(d)
	accountName, channelName, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	slackChannel, httpresp, err := apiInstances.GetSlackIntegrationApiV1().UpdateSlackIntegrationChannel(auth, accountName, channelName, *ddObject)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating slack channel")
	}
	if err := utils.CheckForUnparsed(slackChannel); err != nil {
		return diag.FromErr(err)
	}

	// Handle case where channel name is updated
	d.SetId(fmt.Sprintf("%s:%s", accountName, slackChannel.GetName()))

	return updateSlackChannelState(d, &slackChannel)
}

func resourceDatadogIntegrationSlackChannelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	integrationSlackChannelMutex.Lock()
	defer integrationSlackChannelMutex.Unlock()

	accountName, channelName, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	httpresp, err := apiInstances.GetSlackIntegrationApiV1().RemoveSlackIntegrationChannel(auth, accountName, channelName)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting slack channel")
	}

	return nil
}

func updateSlackChannelState(d *schema.ResourceData, slackChannel *datadogV1.SlackIntegrationChannel) diag.Diagnostics {
	if err := d.Set("channel_name", slackChannel.GetName()); err != nil {
		return diag.FromErr(err)
	}

	accountName, _, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("account_name", accountName); err != nil {
		return diag.FromErr(err)
	}

	tfChannelDisplay := buildTerraformSlackChannelDisplay(slackChannel.GetDisplay())
	if err := d.Set("display", []map[string]interface{}{tfChannelDisplay}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func buildTerraformSlackChannelDisplay(ddChannelDisplay datadogV1.SlackIntegrationChannelDisplay) map[string]interface{} {
	tfChannelDisplay := map[string]interface{}{}
	tfChannelDisplay["message"] = ddChannelDisplay.GetMessage()
	tfChannelDisplay["notified"] = ddChannelDisplay.GetNotified()
	tfChannelDisplay["snapshot"] = ddChannelDisplay.GetSnapshot()
	tfChannelDisplay["tags"] = ddChannelDisplay.GetTags()

	return tfChannelDisplay
}
