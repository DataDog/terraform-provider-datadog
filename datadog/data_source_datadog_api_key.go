package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogApiKey() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing api key.",
		ReadContext: dataSourceDatadogApiKeyRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Id for API Key.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Name for API Key.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			// Computed values
			"key": {
				Description: "The value of the API Key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func dataSourceDatadogApiKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if id := d.Get("id").(string); id != "" {
		resp, httpResponse, err := datadogClientV2.KeyManagementApi.GetAPIKey(authV2, id)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error getting api key")
		}
		apiKeyData := resp.GetData()
		d.SetId(apiKeyData.GetId())
		return updateApiKeyState(d, &apiKeyData)
	} else if name := d.Get("name").(string); name != "" {
		optionalParams := datadogV2.NewListAPIKeysOptionalParameters()
		optionalParams.WithFilter(name)

		apiKeysResponse, httpResponse, err := datadogClientV2.KeyManagementApi.ListAPIKeys(authV2, *optionalParams)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error getting api keys")
		}

		apiKeysData := apiKeysResponse.GetData()

		if len(apiKeysData) > 1 {
			return diag.Errorf("your query returned more than one result, please try a more specific search criteria")
		}
		if len(apiKeysData) == 0 {
			return diag.Errorf("your query returned no result, please try a less specific search criteria")
		}

		apiKeyPartialData := apiKeysData[0]

		id := apiKeyPartialData.GetId()
		apiKeyResponse, httpResponse, err := datadogClientV2.KeyManagementApi.GetAPIKey(authV2, id)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error getting api key")
		}
		apiKeyFullData := apiKeyResponse.GetData()
		d.SetId(apiKeyFullData.GetId())
		return updateApiKeyState(d, &apiKeyFullData)
	}

	return diag.Errorf("missing id or name parameter")
}
