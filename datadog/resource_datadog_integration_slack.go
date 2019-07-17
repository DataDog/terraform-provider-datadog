package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

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
			"service_hook": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of service hook objects.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
						"url": {
							Type:      schema.TypeString,
							Required:  true,
							Sensitive: true,
						},
					},
				},
			},
			"channel": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "A list of slack channel objects to post to.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"channel_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"transfer_all_user_comments": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"account": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func buildIntegrationSlack(d *schema.ResourceData) (*datadog.IntegrationSlackRequest, error) {
	slack := &datadog.IntegrationSlackRequest{}

	serviceHooks := []datadog.ServiceHookSlackRequest{}
	configServiceHooks, ok := d.GetOk("service_hook")
	if ok {
		for _, sInterface := range configServiceHooks.([]interface{}) {
			s := sInterface.(map[string]interface{})

			serviceHook := datadog.ServiceHookSlackRequest{}
			serviceHook.SetAccount(s["account"].(string))
			serviceHook.SetUrl(s["url"].(string))

			serviceHooks = append(serviceHooks, serviceHook)
		}
	}
	slack.ServiceHooks = serviceHooks

	channels := []datadog.ChannelSlackRequest{}
	configChannels, ok := d.GetOk("channel")
	if ok {
		for _, sInterface := range configChannels.([]interface{}) {
			s := sInterface.(map[string]interface{})

			channel := datadog.ChannelSlackRequest{}
			channel.SetChannelName(s["channel_name"].(string))
			channel.SetTransferAllUserComments(s["transfer_all_user_comments"].(bool))
			channel.SetAccount(s["account"].(string))

			channels = append(channels, channel)
		}
	}
	slack.Channels = channels

	return slack, nil
}

func resourceDatadogIntegrationSlackCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	slack, err := buildIntegrationSlack(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if err := client.CreateIntegrationSlack(slack); err != nil {
		return fmt.Errorf("Failed to create integration slack using Datadog API: %s", err.Error())
	}

	_, err = client.GetIntegrationSlack()
	if err != nil {
		return fmt.Errorf("error retrieving integration slack: %s", err.Error())
	}

	d.SetId("0")

	return nil
}

func resourceDatadogIntegrationSlackRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	slack, err := client.GetIntegrationSlack()
	if err != nil {
		return err
	}

	serviceHooks := []map[string]string{}
	for _, serviceHook := range slack.ServiceHooks {
		serviceHooks = append(serviceHooks, map[string]string{
			"account": serviceHook.GetAccount(),
			"url":     serviceHook.GetUrl(),
		})
	}

	channels := []map[string]interface{}{}
	for _, channel := range slack.Channels {
		channels = append(channels, map[string]interface{}{
			"channel_name":               channel.GetChannelName(),
			"transfer_all_user_comments": channel.GetTransferAllUserComments(),
			"account":                    channel.GetAccount(),
		})
	}

	d.Set("service_hook", serviceHooks)
	d.Set("channel", channels)

	return nil
}

func resourceDatadogIntegrationSlackUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	slack, err := buildIntegrationSlack(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if err := client.UpdateIntegrationSlack(slack); err != nil {
		return fmt.Errorf("Failed to create integration slack using Datadog API: %s", err.Error())
	}

	return resourceDatadogIntegrationSlackRead(d, meta)
}

func resourceDatadogIntegrationSlackDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteIntegrationSlack(); err != nil {
		return fmt.Errorf("Error while deleting integration: %v", err)
	}

	return nil
}
