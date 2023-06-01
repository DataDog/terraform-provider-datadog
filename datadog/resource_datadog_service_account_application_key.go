package datadog

import (
	"context"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func resourceDatadogServiceAccountApplicationKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog application key resource for the specified service account. This can be used to create and manage Datadog service account application keys.",
		CreateContext: resourceDatadogServiceAccountApplicationKeyCreate,
		ReadContext:   resourceDatadogServiceAccountApplicationKeyRead,
		// Only the owner of the application key can do an update on it
		// Since the scope of this is management of keys through CI/CD
		// It doesn't make sense to create an update handler
		UpdateContext: nil,
		DeleteContext: resourceDatadogServiceAccountApplicationKeyDelete,
		// The ServiceAccount API and ApplicationKey API do not allow fetching
		// of the "key" (password) unless using the /current_user/application_keys api.
		// This provides complications in a CI/CD environment. We can still use this
		// resource since everything will be cached in the terraform state.
		// This also means a datasource is not possible or necessary
		Importer: nil,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "Name for the service account application key.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"service_account_id": {
				Description: "ID of the service account that owns this key.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},

			// Computed values
			"key": {
				Description: "The value of the service account application key.",
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				ForceNew:    true,
			},
			"id": {
				Description: "ID for the service account application key association.",
				Type:        schema.TypeString,
				Computed:    true,
			},
		},
	}
}

func buildDatadogServiceAccountAppKeyCreateV2Struct(d *schema.ResourceData) *datadogV2.ApplicationKeyCreateRequest {
	return buildDatadogApplicationKeyCreateV2Struct(d)
}

func updateServiceAccountApplicationKeyStateOnCreate(d *schema.ResourceData, applicationKeyData *datadogV2.FullApplicationKey) diag.Diagnostics {
	applicationKeyAttributes := applicationKeyData.GetAttributes()

	if err := d.Set("key", applicationKeyAttributes.GetKey()); err != nil {
		return diag.FromErr(err)
	}

	if diagErr := updateServiceAccountApplicationKeyStateFull(d, applicationKeyData); diagErr != nil {
		return diagErr
	}

	return nil
}

func updateServiceAccountApplicationKeyStateFull(d *schema.ResourceData, applicationKeyData *datadogV2.FullApplicationKey) diag.Diagnostics {
	applicationKeyAttributes := applicationKeyData.GetAttributes()
	relationships := applicationKeyData.GetRelationships()

	if err := d.Set("name", applicationKeyAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("service_account_id", extractOwnerIdFromApplicationKeyRelationshipsHelper(&relationships)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateServiceAccountApplicationKeyStatePartial(d *schema.ResourceData, applicationKeyData *datadogV2.PartialApplicationKey) diag.Diagnostics {
	applicationKeyAttributes := applicationKeyData.GetAttributes()
	relationships := applicationKeyData.GetRelationships()

	if err := d.Set("name", applicationKeyAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("service_account_id", extractOwnerIdFromApplicationKeyRelationshipsHelper(&relationships)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDatadogServiceAccountApplicationKeyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetServiceAccountsApiV2().CreateServiceAccountApplicationKey(auth, d.Get("service_account_id").(string), *buildDatadogServiceAccountAppKeyCreateV2Struct(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating api key")
	}

	appKeyData := resp.GetData()
	d.SetId(appKeyData.GetId())

	return updateServiceAccountApplicationKeyStateOnCreate(d, &appKeyData)
}

func resourceDatadogServiceAccountApplicationKeyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	keyId := d.Id()

	_, httpResponse, err := apiInstances.GetServiceAccountsApiV2().GetServiceAccountApplicationKey(
		auth, d.Get("service_account_id").(string), keyId)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting application key")
	}

	appKey, httpResponse, err := apiInstances.GetKeyManagementApiV2().GetApplicationKey(auth, d.Id())

	applicationKeyData := appKey.GetData()

	return updateServiceAccountApplicationKeyStateFull(d, &applicationKeyData)
}

func resourceDatadogServiceAccountApplicationKeyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	httpResponse, err := apiInstances.GetServiceAccountsApiV2().DeleteServiceAccountApplicationKey(auth, d.Get("service_account_id").(string), d.Id())
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting api key")
	}

	return nil
}

func extractOwnerIdFromApplicationKeyRelationshipsHelper(relationships *datadogV2.ApplicationKeyRelationships) string {
	ownedBy := relationships.GetOwnedBy()
	ownerData := ownedBy.GetData()

	return ownerData.GetId()
}
