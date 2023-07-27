package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogWebhookCustomVariable() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog webhooks custom variable resource. This can be used to create and manage Datadog webhooks custom variables.",
		CreateContext: resourceDatadogWebhookCustomVariableCreate,
		ReadContext:   resourceDatadogWebhookCustomVariableRead,
		UpdateContext: resourceDatadogWebhookCustomVariableUpdate,
		DeleteContext: resourceDatadogWebhookCustomVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description: "The name of the variable. It corresponds with `<CUSTOM_VARIABLE_NAME>`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"value": {
					Description: "The value of the custom variable.",
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
				},
				"is_secret": {
					Description: "Whether the custom variable is secret or not.",
					Type:        schema.TypeBool,
					Required:    true,
				},
			}
		},
	}
}

func updateWebhookCustomVariableState(d *schema.ResourceData, customVariable *datadogV1.WebhooksIntegrationCustomVariableResponse) diag.Diagnostics {
	if err := d.Set("name", customVariable.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if v, ok := customVariable.GetValueOk(); ok {
		if err := d.Set("value", v); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("is_secret", customVariable.GetIsSecret()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogWebhookCustomVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookMutex.Lock()
	defer webhookMutex.Unlock()

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetWebhooksIntegrationApiV1().CreateWebhooksIntegrationCustomVariable(auth, datadogV1.WebhooksIntegrationCustomVariable{
		Name:     d.Get("name").(string),
		Value:    d.Get("value").(string),
		IsSecret: d.Get("is_secret").(bool),
	})
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating webhooks custom variable")
	}

	d.SetId(resp.GetName())

	return updateWebhookCustomVariableState(d, &resp)
}

func resourceDatadogWebhookCustomVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetWebhooksIntegrationApiV1().GetWebhooksIntegrationCustomVariable(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting webhooks custom variable")
	}
	return updateWebhookCustomVariableState(d, &resp)
}

func resourceDatadogWebhookCustomVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookMutex.Lock()
	defer webhookMutex.Unlock()

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetWebhooksIntegrationApiV1().UpdateWebhooksIntegrationCustomVariable(auth, d.Id(), datadogV1.WebhooksIntegrationCustomVariableUpdateRequest{
		Name:     datadog.PtrString(d.Get("name").(string)),
		Value:    datadog.PtrString(d.Get("value").(string)),
		IsSecret: datadog.PtrBool(d.Get("is_secret").(bool)),
	})
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating webhooks custom variable key")
	}

	d.SetId(resp.GetName())

	return updateWebhookCustomVariableState(d, &resp)
}

func resourceDatadogWebhookCustomVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	webhookMutex.Lock()
	defer webhookMutex.Unlock()

	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetWebhooksIntegrationApiV1().DeleteWebhooksIntegrationCustomVariable(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting webhooks custom variable")
	}

	return nil
}
