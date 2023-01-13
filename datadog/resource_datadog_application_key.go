package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog Application Key resource. This can be used to create and manage Datadog Application Keys.",
		CreateContext: resourceDatadogApplicationKeyCreate,
		ReadContext:   resourceDatadogApplicationKeyRead,
		UpdateContext: resourceDatadogApplicationKeyUpdate,
		DeleteContext: resourceDatadogApplicationKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name for Application Key.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"scopes": {
				Description: "Authorization scopes for the Application Key.",
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"key": {
				Description: "The value of the Application Key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func buildDatadogApplicationKeyCreateV2Struct(d *schema.ResourceData) *datadogV2.ApplicationKeyCreateRequest {
	applicationKeyAttributes := datadogV2.NewApplicationKeyCreateAttributes(d.Get("name").(string))
	tfScopes := d.Get("scopes").([]interface{})
	ddScope := make([]string, len(tfScopes))
	for i, tfScope := range tfScopes {
		ddScope[i] = tfScope.(string)
	}
	applicationKeyAttributes.SetScopes(ddScope)
	applicationKeyData := datadogV2.NewApplicationKeyCreateData(*applicationKeyAttributes, datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS)
	applicationKeyRequest := datadogV2.NewApplicationKeyCreateRequest(*applicationKeyData)

	return applicationKeyRequest
}

func buildDatadogApplicationKeyUpdateV2Struct(d *schema.ResourceData) *datadogV2.ApplicationKeyUpdateRequest {
	applicationKeyAttributes := datadogV2.NewApplicationKeyUpdateAttributes()
	applicationKeyAttributes.SetName(d.Get("name").(string))
	tfScopes := d.Get("scopes").([]interface{})
	ddScope := make([]string, len(tfScopes))
	for i, tfScope := range tfScopes {
		ddScope[i] = tfScope.(string)
	}
	applicationKeyAttributes.SetScopes(ddScope)
	applicationKeyData := datadogV2.NewApplicationKeyUpdateData(*applicationKeyAttributes, d.Id(), datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS)
	applicationKeyRequest := datadogV2.NewApplicationKeyUpdateRequest(*applicationKeyData)

	return applicationKeyRequest
}

func updateApplicationKeyState(d *schema.ResourceData, applicationKeyData *datadogV2.FullApplicationKey) diag.Diagnostics {
	applicationKeyAttributes := applicationKeyData.GetAttributes()

	if err := d.Set("name", applicationKeyAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("scopes", applicationKeyAttributes.GetScopes()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key", applicationKeyAttributes.GetKey()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogApplicationKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetKeyManagementApiV2().CreateCurrentUserApplicationKey(auth, *buildDatadogApplicationKeyCreateV2Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating application key")
	}

	applicationKeyData := resp.GetData()
	d.SetId(applicationKeyData.GetId())

	return updateApplicationKeyState(d, &applicationKeyData)
}

func resourceDatadogApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetKeyManagementApiV2().GetCurrentUserApplicationKey(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting application key")
	}
	applicationKeyData := resp.GetData()
	return updateApplicationKeyState(d, &applicationKeyData)
}

func resourceDatadogApplicationKeyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetKeyManagementApiV2().UpdateCurrentUserApplicationKey(auth, d.Id(), *buildDatadogApplicationKeyUpdateV2Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating application key")
	}
	applicationKeyData := resp.GetData()
	return updateApplicationKeyState(d, &applicationKeyData)
}

func resourceDatadogApplicationKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetKeyManagementApiV2().DeleteCurrentUserApplicationKey(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting application key")
	}

	return nil
}
