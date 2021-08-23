package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV2 "github.com/DataDog/datadog-api-client-go/api/v2/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceDatadogApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about an existing application key.",
		ReadContext: dataSourceDatadogApplicationKeyRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Id for Application Key.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"name": {
				Description: "Name for Application Key.",
				Type:        schema.TypeString,
				Optional:    true,
			},

			// Computed values
			"key": {
				Description: "The value of the Application Key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func dataSourceDatadogApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV2 := providerConf.DatadogClientV2
	authV2 := providerConf.AuthV2

	if id := d.Get("id").(string); id != "" {
		resp, httpResponse, err := datadogClientV2.KeyManagementApi.GetCurrentUserApplicationKey(authV2, id)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error getting application key")
		}
		applicationKeyData := resp.GetData()
		d.SetId(applicationKeyData.GetId())
		return updateApplicationKeyState(d, &applicationKeyData)
	} else if name := d.Get("name").(string); name != "" {
		optionalParams := datadogV2.NewListCurrentUserApplicationKeysOptionalParameters()
		optionalParams.WithFilter(name)

		applicationKeysResponse, httpResponse, err := datadogClientV2.KeyManagementApi.ListCurrentUserApplicationKeys(authV2, *optionalParams)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error getting application keys")
		}

		applicationKeysData := applicationKeysResponse.GetData()

		if len(applicationKeysData) > 1 {
			return diag.Errorf("your query returned more than one result, please try a more specific search criteria")
		}
		if len(applicationKeysData) == 0 {
			return diag.Errorf("your query returned no result, please try a less specific search criteria")
		}

		applicationKeyPartialData := applicationKeysData[0]

		id := applicationKeyPartialData.GetId()
		applicationKeyResponse, httpResponse, err := datadogClientV2.KeyManagementApi.GetCurrentUserApplicationKey(authV2, id)
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error getting application key")
		}
		applicationKeyFullData := applicationKeyResponse.GetData()
		d.SetId(applicationKeyFullData.GetId())
		return updateApplicationKeyState(d, &applicationKeyFullData)
	}

	return diag.Errorf("missing id or name parameter")
}
