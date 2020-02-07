package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
)

func resourceDatadogIntegrationWebhook() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationWebhookCreate,
		Read:   resourceDatadogIntegrationWebhookRead,
		Update: resourceDatadogIntegrationWebhookUpdate,
		Delete: resourceDatadogIntegrationWebhookDelete,
		Exists: resourceDatadogIntegrationWebhookExists,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"payload": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"custom_headers": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"encode_as_form": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func buildIntegrationWebhook(d *schema.ResourceData) (datadog.Webhook, error) {
	w := datadog.Webhook{}
	w.SetName(d.Get("name").(string))
	w.SetURL(d.Get("url").(string))

	if attr, ok := d.GetOk("payload"); ok {
		w.SetCustomPayload(attr.(string))
		w.SetUseCustomPayload("true")
	}
	if attr, ok := d.GetOk("custom_headers"); ok {
		w.SetHeaders(attr.(string))
	}
	if attr, ok := d.GetOk("encode_as_form"); ok {
		w.SetEncodeAsForm(attr.(string))
	}
	return w, nil
}

func resourceDatadogIntegrationWebhookCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	w, err := buildIntegrationWebhook(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}

	if err := client.CreateIntegrationWebhook(&datadog.IntegrationWebhookRequest{Webhooks: []datadog.Webhook{w}}); err != nil {
		return fmt.Errorf("Failed to create integration webhook: %s", err.Error())
	}

	d.SetId(w.GetName())

	return nil
}

func resourceDatadogIntegrationWebhookRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)
	webhookName := d.Id()

	WebhookIntegration, err := client.GetIntegrationWebhook()
	if err != nil {
		return fmt.Errorf("error retrieving integration webhook: %s", err.Error())
	}
	for _, wk := range WebhookIntegration.Webhooks {
		if wk.GetName() == webhookName {
			d.Set("url", wk.GetURL)
			d.Set("payload", wk.GetCustomPayload)
			d.Set("custom_headers", wk.GetHeaders)
			d.Set("encode_as_form", wk.GetEncodeAsForm)
			return nil
		}
	}
	return fmt.Errorf("error getting a webhook integration: name=%s", webhookName)
}

func resourceDatadogIntegrationWebhookUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	w, err := buildIntegrationWebhook(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	webhookIntegration, _ := client.GetIntegrationWebhook()
	for i, wk := range webhookIntegration.Webhooks {
		if wk.GetName() == w.GetName() {
			webhookIntegration.Webhooks[i] = w
		}
	}

	if err := client.UpdateIntegrationWebhook(webhookIntegration); err != nil {
		return fmt.Errorf("Failed to update integration webhook: %s", err.Error())
	}

	return resourceDatadogIntegrationWebhookRead(d, meta)
}

func resourceDatadogIntegrationWebhookDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	if err := client.DeleteIntegrationWebhook(); err != nil {
		return fmt.Errorf("Error while deleting webhook integration: %v", err)
	}

	return nil
}

func resourceDatadogIntegrationWebhookExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*datadog.Client)

	webhookIntegration, err := client.GetIntegrationWebhook()
	if err != nil {
		return false, err
	}
	webhookName := d.Id()
	for _, wk := range webhookIntegration.Webhooks {
		if wk.GetName() == webhookName {
			return true, nil
		}
	}

	return false, nil
}
