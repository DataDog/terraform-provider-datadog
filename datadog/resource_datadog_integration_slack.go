package datadog

import (
	"fmt"
	"strings"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

// creating/modifying/deleting Slack integration and its service objects in parallel on one account
// is unsupported by the API right now; therefore we use the mutex to only operate on one at a time
var integrationSlackMutex = sync.Mutex{}

func resourceDatadogIntegrationSlack() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationSlackCreate,
		Read:   resourceDatadogIntegrationSlackRead,
		Update: resourceDatadogIntegrationSlackUpdate,
		Delete: resourceDatadogIntegrationSlackDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account": {Type: schema.TypeString, Required: true, ForceNew: true},
			"url":     {Type: schema.TypeString, Required: true, Sensitive: true},
			"channels": {
				Description: "Slack channels to use.",
				Type:        schema.TypeSet,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"transfer_all_user_comments": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func buildIntegrationSlack(d *schema.ResourceData, client *datadog.Client) (*datadog.IntegrationSlackRequest, bool, error) {
	slackActualIntegration, err := client.GetIntegrationSlack()
	slackRequest := &datadog.IntegrationSlackRequest{}

	needCreation := false
	if err != nil {
		// This is a real error
		if !strings.Contains(err.Error(), "slack not found") {
			return nil, false, err

		}
		// slack integration is not created right now
		needCreation = true
	}

	if slackActualIntegration != nil {
		// add existing channels except channels link to the wanted hook
		for _, channel := range slackActualIntegration.Channels {
			if channel.GetAccount() == d.Get("account").(string) {
				continue
			}
			slackRequest.Channels = append(slackRequest.Channels, channel)
		}
	}

	serviceHook := datadog.ServiceHookSlackRequest{}
	// add service hook
	serviceHook.SetUrl(d.Get("url").(string))
	serviceHook.SetAccount(d.Get("account").(string))
	serviceHooks := []datadog.ServiceHookSlackRequest{
		serviceHook,
	}

	slackRequest.ServiceHooks = serviceHooks

	// add channels
	for _, channelI := range d.Get("channels").(*schema.Set).List() {
		channel := channelI.(map[string]interface{})
		slackChannel := datadog.ChannelSlackRequest{}
		slackChannel.SetAccount(d.Get("account").(string))
		slackChannel.SetChannelName(channel["channel_name"].(string))
		slackChannel.SetTransferAllUserComments(channel["transfer_all_user_comments"].(bool))
		slackRequest.Channels = append(slackRequest.Channels, slackChannel)
	}

	return slackRequest, needCreation, nil
}

func resourceDatadogIntegrationSlackCreate(d *schema.ResourceData, meta interface{}) error {
	integrationSlackMutex.Lock()
	defer integrationSlackMutex.Unlock()

	client := meta.(*datadog.Client)
	slack, needCreation, err := buildIntegrationSlack(d, client)
	if err != nil {
		return err
	}

	if needCreation {
		if err := client.CreateIntegrationSlack(slack); err != nil {
			return fmt.Errorf("Failed to create integration slack using Datadog API: %s", err.Error())
		}
	} else {
		if err := client.UpdateIntegrationSlack(slack); err != nil {
			return fmt.Errorf("Failed to update integration slack using Datadog API: %s", err.Error())
		}
	}
	d.SetId(d.Get("account").(string))

	return nil
}

func resourceDatadogIntegrationSlackRead(d *schema.ResourceData, meta interface{}) error {
	integrationSlackMutex.Lock()
	defer integrationSlackMutex.Unlock()

	client := meta.(*datadog.Client)
	account := d.Id()

	slack, err := client.GetIntegrationSlack()
	if err != nil {
		if !strings.Contains(err.Error(), "slack not found") {
			return err

		}
		return nil
	}

	d.SetId("")
	for _, hook := range slack.ServiceHooks {
		if hook.GetAccount() == account {
			d.Set("account", account)
			d.SetId(account)
			break
		}
	}
	if d.Id() == "" {
		// Hook was remove from datadog
		return nil
	}

	var channels []map[string]interface{}
	for _, slackChannel := range slack.Channels {
		if slackChannel.GetAccount() != account {
			continue
		}
		channel := map[string]interface{}{
			"channel_name":               slackChannel.GetChannelName(),
			"transfer_all_user_comments": slackChannel.GetTransferAllUserComments(),
		}
		channels = append(channels, channel)
	}
	d.Set("channels", channels)

	return nil
}

func resourceDatadogIntegrationSlackUpdate(d *schema.ResourceData, meta interface{}) error {
	integrationSlackMutex.Lock()
	defer integrationSlackMutex.Unlock()

	client := meta.(*datadog.Client)
	slack, needCreation, err := buildIntegrationSlack(d, client)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if needCreation {
		if err := client.CreateIntegrationSlack(slack); err != nil {
			return fmt.Errorf("Failed to create integration slack using Datadog API: %s", err.Error())
		}
	} else {
		if err := client.UpdateIntegrationSlack(slack); err != nil {
			return fmt.Errorf("Failed to update integration slack using Datadog API: %s", err.Error())
		}
	}

	return nil
}

func resourceDatadogIntegrationSlackDelete(d *schema.ResourceData, meta interface{}) error {
	integrationSlackMutex.Lock()
	defer integrationSlackMutex.Unlock()

	client := meta.(*datadog.Client)

	// Remove channels from the hook. Removing hook from api is not actually supported.
	d.Set("channels", []interface{}{})
	slack, needCreation, err := buildIntegrationSlack(d, client)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	// slack integration is not enable
	if needCreation {
		return nil
	}

	if len(slack.Channels) == 0 {
		// If we will remove all channels just remove the integration.
		if err := client.DeleteIntegrationSlack(); err != nil {
			return fmt.Errorf("Failed to remove integration slack using Datadog API: %s", err.Error())
		}
	} else {
		if err := client.UpdateIntegrationSlack(slack); err != nil {
			return fmt.Errorf("Failed to update integration slack using Datadog API: %s", err.Error())
		}
	}

	return nil
}
