package datadog

import (
	"context"
	"sync"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var webhookMutex = sync.Mutex{}

func resourceDatadogWebhook() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog webhook resource. This can be used to create and manage Datadog webhooks.",
		CreateContext: resourceDatadogWebhookCreate,
		ReadContext:   resourceDatadogWebhookRead,
		UpdateContext: resourceDatadogWebhookUpdate,
		DeleteContext: resourceDatadogWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the webhook. It corresponds with `<WEBHOOK_NAME>`.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"url": {
				Description: "The URL of the webhook.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"payload": {
				Description: "The payload of the webhook.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
			"custom_headers": {
				Description: "The headers attached to the webhook.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"encode_as": {
				Description:      "Encoding type.",
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: validators.ValidateEnumValue(datadogV1.NewWebhooksIntegrationEncodingFromValue),
				Computed:         true,
			},
		},
	}
}

func updateWebhookState(d *schema.ResourceData, webhook *datadogV1.WebhooksIntegration) diag.Diagnostics {
	// Required attributes
	if err := d.Set("name", webhook.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("url", webhook.GetUrl()); err != nil {
		return diag.FromErr(err)
	}

	// Optional attributes
	if v, ok := webhook.GetPayloadOk(); ok {
		if err := d.Set("payload", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := webhook.GetCustomHeadersOk(); ok {
		if err := d.Set("custom_headers", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := webhook.GetEncodeAsOk(); ok {
		if err := d.Set("encode_as", v); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDatadogWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookMutex.Lock()
	defer webhookMutex.Unlock()

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	webhook, err := buildWebhookCreatePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, httpResponse, err := datadogClientV1.WebhooksIntegrationApi.CreateWebhooksIntegration(authV1, *webhook)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating webhooks custom variable")
	}

	d.SetId(resp.GetName())

	return updateWebhookState(d, &resp)
}

func resourceDatadogWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	resp, httpResponse, err := datadogClientV1.WebhooksIntegrationApi.GetWebhooksIntegration(authV1, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting webhook")
	}
	return updateWebhookState(d, &resp)
}

func resourceDatadogWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookMutex.Lock()
	defer webhookMutex.Unlock()

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	webhook, err := buildWebhookUpdatePayload(d)
	if err != nil {
		return diag.FromErr(err)
	}

	resp, httpResponse, err := datadogClientV1.WebhooksIntegrationApi.UpdateWebhooksIntegration(authV1, d.Id(), *webhook)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating webhook")
	}

	d.SetId(resp.GetName())

	return updateWebhookState(d, &resp)
}

func resourceDatadogWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookMutex.Lock()
	defer webhookMutex.Unlock()

	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	if httpResponse, err := datadogClientV1.WebhooksIntegrationApi.DeleteWebhooksIntegration(authV1, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting webhook")
	}

	return nil
}

func buildWebhookCreatePayload(d *schema.ResourceData) (*datadogV1.WebhooksIntegration, error) {
	payload := datadogV1.WebhooksIntegration{}

	payload.SetName(d.Get("name").(string))
	payload.SetUrl(d.Get("url").(string))
	if v, ok := d.GetOk("payload"); ok {
		payload.SetPayload(v.(string))
	}
	if v, ok := d.GetOk("custom_headers"); ok {
		payload.SetCustomHeaders(v.(string))
	}
	if v, ok := d.GetOk("encode_as"); ok {
		encoding, err := datadogV1.NewWebhooksIntegrationEncodingFromValue(v.(string))
		if err != nil {
			return nil, err
		}
		payload.SetEncodeAs(*encoding)
	}

	return &payload, nil
}

func buildWebhookUpdatePayload(d *schema.ResourceData) (*datadogV1.WebhooksIntegrationUpdateRequest, error) {
	payload := datadogV1.WebhooksIntegrationUpdateRequest{}

	payload.SetName(d.Get("name").(string))
	payload.SetUrl(d.Get("url").(string))
	if v, ok := d.GetOk("payload"); ok {
		payload.SetPayload(v.(string))
	}
	if v, ok := d.GetOk("custom_headers"); ok {
		payload.SetCustomHeaders(v.(string))
	}
	if v, ok := d.GetOk("encode_as"); ok {
		encoding, err := datadogV1.NewWebhooksIntegrationEncodingFromValue(v.(string))
		if err != nil {
			return nil, err
		}
		payload.SetEncodeAs(*encoding)
	}

	return &payload, nil
}
