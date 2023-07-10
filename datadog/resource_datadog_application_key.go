package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
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
			"owner": {
				Description: "Application Key owner ID",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// ignore existing resources without owner == assume it's current user to avoid breaking existing uses
					return new == ""
				},
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
	applicationKeyData := datadogV2.NewApplicationKeyCreateData(*applicationKeyAttributes, datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS)
	applicationKeyRequest := datadogV2.NewApplicationKeyCreateRequest(*applicationKeyData)

	return applicationKeyRequest
}

func buildDatadogApplicationKeyUpdateV2Struct(d *schema.ResourceData) *datadogV2.ApplicationKeyUpdateRequest {
	applicationKeyAttributes := datadogV2.NewApplicationKeyUpdateAttributes()
	applicationKeyAttributes.SetName(d.Get("name").(string))
	applicationKeyData := datadogV2.NewApplicationKeyUpdateData(*applicationKeyAttributes, d.Id(), datadogV2.APPLICATIONKEYSTYPE_APPLICATION_KEYS)
	applicationKeyRequest := datadogV2.NewApplicationKeyUpdateRequest(*applicationKeyData)

	return applicationKeyRequest
}

func updateApplicationKeyState(d *schema.ResourceData, applicationKeyData *datadogV2.FullApplicationKey) diag.Diagnostics {
	applicationKeyAttributes := applicationKeyData.GetAttributes()

	if err := d.Set("name", applicationKeyAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("key", applicationKeyAttributes.GetKey()); err != nil {
		return diag.FromErr(err)
	}

	applicationKeyRelationships := applicationKeyData.GetRelationships()
	applicationKeyOwner := applicationKeyRelationships.GetOwnedBy()
	applicationKeyOwnerData := applicationKeyOwner.GetData()

	if err := d.Set("owner", applicationKeyOwnerData.Id); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogApplicationKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if d.Get("owner") == "" {
		resp, httpResponse, err := apiInstances.GetKeyManagementApiV2().CreateCurrentUserApplicationKey(auth, *buildDatadogApplicationKeyCreateV2Struct(d))
		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error creating application key")
		}

		applicationKeyData := resp.GetData()
		d.SetId(applicationKeyData.GetId())

		return updateApplicationKeyState(d, &applicationKeyData)
	} else {
		// Managing key for another user requires using v1 API
		body := datadogV1.ApplicationKey{
			Name:  datadog.PtrString(d.Get("name").(string)),
			Owner: datadog.PtrString(d.Get("owner").(string)),
		}
		resp, httpResponse, err := apiInstances.GetKeyManagementApiV1().CreateApplicationKey(auth, body)

		if err != nil {
			return utils.TranslateClientErrorDiag(err, httpResponse, "error creating application key")
		}

		// Save the ID
		appkey := resp.GetApplicationKey()
		/* TODO: remove debugging
		responseContent, _ := json.MarshalIndent(resp, "", "  ")
		log.Printf("[INFO] %s", string(responseContent))
		*/
		d.SetId(appkey.GetHash())

		// Now call v2Update to set the v2 attributes
		return resourceDatadogApplicationKeyUpdate(ctx, d, meta)
	}
}

func resourceDatadogApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetKeyManagementApiV2().GetApplicationKey(auth, d.Id(), *datadogV2.NewGetApplicationKeyOptionalParameters())
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

	resp, httpResponse, err := apiInstances.GetKeyManagementApiV2().UpdateApplicationKey(auth, d.Id(), *buildDatadogApplicationKeyUpdateV2Struct(d))
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

	if httpResponse, err := apiInstances.GetKeyManagementApiV2().DeleteApplicationKey(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting application key")
	}

	return nil
}
