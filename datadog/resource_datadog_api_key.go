package datadog

import (
	"context"
	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogApiKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog API Key resource. This can be used to create and manage Datadog API Keys.",
		CreateContext: resourceDatadogApiKeyCreate,
		ReadContext:   resourceDatadogApiKeyRead,
		DeleteContext: resourceDatadogApiKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name for API Key.",
				Type:        schema.TypeString,
				ForceNew:    true,
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

func buildDatadogApiKeyV2Struct(d *schema.ResourceData) *datadogV2.APIKeyCreateRequest {
	apiKeyAttributes := datadogV2.NewAPIKeyCreateAttributes(d.Get("name").(string))
	apiKeyData := datadogV2.NewAPIKeyCreateData(*apiKeyAttributes, datadogV2.APIKEYSTYPE_API_KEYS)
	apiKeyRequest := datadogV2.NewAPIKeyCreateRequest(*apiKeyData)

	return apiKeyRequest
}

func updateApiKeyState(d *schema.ResourceData, apiKey *datadogV2.APIKeyResponse) diag.Diagnostics {
	apiKeyData := apiKey.GetData()
	apiKeyAttributes := apiKeyData.GetAttributes()

	d.SetId(apiKeyData.GetId())
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

	apiKeyResponse, _, err := datadogClientV2.KeyManagementApi.CreateAPIKey(authV2, *buildDatadogApiKeyV2Struct(d))
	if err != nil {
		return diag.FromErr(err)
	}
	return updateApiKeyState(d, &apiKeyResponse)
}

func resourceDatadogApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	apiKeyResponse, _, err := datadogClientV2.KeyManagementApi.GetAPIKey(authV2, d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	return updateApiKeyState(d, &apiKeyResponse)
}

func resourceDatadogApiKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if _, err := datadogClientV2.KeyManagementApi.DeleteAPIKey(authV2, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
