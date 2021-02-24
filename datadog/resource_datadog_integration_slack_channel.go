package datadog

import (
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogIntegrationSlackChannel() *schema.Resource {
	return &schema.Resource{
		Description: "Resource for interacting with the Datadog Slack channel API",
		Create:      resourceDatadogIntegrationSlackChannelCreate,
		Read:        resourceDatadogIntegrationSlackChannelRead,
		Update:      resourceDatadogIntegrationSlackChannelUpdate,
		Delete:      resourceDatadogIntegrationSlackChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
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
		},
	}
}

func buildDatadogSlackChannel(d *schema.ResourceData) (*datadogV1.SlackIntegrationChannel, error) {
	k := utils.NewResourceDataKey(d, "")
	datadogSlackChannel := datadogV1.NewSlackIntegrationChannelWithDefaults()

	if v, ok := k.GetOkWith("channel_name"); ok {
		datadogSlackChannel.SetName(v.(string))
	}

	k.Add("display.0")
	resultDisplay := datadogV1.NewSlackIntegrationChannelDisplayWithDefaults()
	resultDisplay.SetMessage(k.GetWith("message").(bool))
	resultDisplay.SetNotified(k.GetWith("notified").(bool))
	resultDisplay.SetSnapshot(k.GetWith("snapshot").(bool))
	resultDisplay.SetTags(k.GetWith("tags").(bool))
	k.Remove("display.0")

	datadogSlackChannel.SetDisplay(*resultDisplay)

	return datadogSlackChannel, nil
}

func resourceDatadogIntegrationSlackChannelCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddSlackChannel, err := buildDatadogSlackChannel(d)
	accountName := d.Get("account_name").(string)

	createdChannel, _, err := datadogClient.SlackIntegrationApi.CreateSlackIntegrationChannel(auth, accountName).Body(*ddSlackChannel).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error creating slack channel")
	}

	d.SetId(fmt.Sprintf("%s:%s", accountName, ddSlackChannel.GetName()))
	return updateSlackChannelState(d, &createdChannel)
}

func resourceDatadogIntegrationSlackChannelRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	accountName, channelName, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return err
	}

	slackChannel, httpResp, err := datadogClient.SlackIntegrationApi.GetSlackIntegrationChannel(auth, accountName, channelName).Execute()
	if err != nil {
		if httpResp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientError(err, "error getting slack channel")
	}

	return updateSlackChannelState(d, &slackChannel)
}

func resourceDatadogIntegrationSlackChannelUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	ddObject, err := buildDatadogSlackChannel(d)
	accountName, channelName, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return err
	}

	slackChannel, _, err := datadogClient.SlackIntegrationApi.UpdateSlackIntegrationChannel(auth, accountName, channelName).Body(*ddObject).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error updating slack channel")
	}

	// Handle case where channel name is updated
	d.SetId(fmt.Sprintf("%s:%s", accountName, slackChannel.GetName()))

	return updateSlackChannelState(d, &slackChannel)
}

func resourceDatadogIntegrationSlackChannelDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClient := providerConf.DatadogClientV1
	auth := providerConf.AuthV1

	accountName, channelName, err := utils.AccountNameAndChannelNameFromID(d.Id())
	if err != nil {
		return err
	}

	_, err = datadogClient.SlackIntegrationApi.RemoveSlackIntegrationChannel(auth, accountName, channelName).Execute()
	if err != nil {
		return utils.TranslateClientError(err, "error deleting slack channel")
	}

	return nil
}

func updateSlackChannelState(d *schema.ResourceData, slackChannel *datadogV1.SlackIntegrationChannel) error {
	if err := d.Set("channel_name", slackChannel.GetName()); err != nil {
		return err
	}

	tfChannelDisplay := buildTerraformSlackChannelDisplay(slackChannel.GetDisplay())
	if err := d.Set("display", []map[string]interface{}{tfChannelDisplay}); err != nil {
		return err
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
