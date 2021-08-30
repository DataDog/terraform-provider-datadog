package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogApiKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog API Key resource. This can be used to create and manage Datadog API Keys.",
		CreateContext: resourceDatadogApiKeyCreate,
		ReadContext:   resourceDatadogApiKeyRead,
		UpdateContext: resourceDatadogApiKeyUpdate,
		DeleteContext: resourceDatadogApiKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name for API Key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"key": {
				Description: "The value of the API Key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func buildDatadogApiKeyCreateV2Struct(d *schema.ResourceData) *datadogV2.APIKeyCreateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyCreateAttributes(d.Get("name").(string))
	apiKeyData := datadogV2.NewAPIKeyCreateData(*apiKeyAttributes, datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyCreateRequest(*apiKeyData)

	return apiKeyRequest
}

func buildDatadogApiKeyUpdateV2Struct(d *schema.ResourceData) *datadogV2.APIKeyUpdateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyUpdateAttributes(d.Get("name").(string))
	apiKeyData := datadogV2.NewAPIKeyUpdateData(*apiKeyAttributes, d.Id(), datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyUpdateRequest(*apiKeyData)

	return apiKeyRequest
}

func updateApiKeyState(d *schema.ResourceData, apiKeyData *datadogV2.FullAPIKey) diag.Diagnostics {
	apiKeyAttributes := apiKeyData.GetAttributes()

	if err := d.Set("name", apiKeyAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key", apiKeyAttributes.GetKey()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogApiKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	resp, httpResponse, err := datadogClientV2.KeyManagementApi.CreateAPIKey(authV2, *buildDatadogApiKeyCreateV2Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating api key")
	}

	apiKeyData := resp.GetData()
	d.SetId(apiKeyData.GetId())

	return updateApiKeyState(d, &apiKeyData)
}

func resourceDatadogApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	resp, httpResponse, err := datadogClientV2.KeyManagementApi.GetAPIKey(authV2, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting api key")
	}
	apiKeyData := resp.GetData()
	return updateApiKeyState(d, &apiKeyData)
}

func resourceDatadogApiKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	resp, httpResponse, err := datadogClientV2.KeyManagementApi.UpdateAPIKey(authV2, d.Id(), *buildDatadogApiKeyUpdateV2Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating api key")
	}
	apiKeyData := resp.GetData()
	return updateApiKeyState(d, &apiKeyData)
}

func resourceDatadogApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if httpResponse, err := datadogClientV2.KeyManagementApi.DeleteAPIKey(authV2, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting api key")
	}

	return nil
}
