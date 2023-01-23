package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationCloudflareAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog IntegrationCloudflareAccount resource. This can be used to create and manage Datadog integration_cloudflare_account.",
		ReadContext:   resourceDatadogIntegrationCloudflareAccountRead,
		CreateContext: resourceDatadogIntegrationCloudflareAccountCreate,
		UpdateContext: resourceDatadogIntegrationCloudflareAccountUpdate,
		DeleteContext: resourceDatadogIntegrationCloudflareAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API key (or token) for the Cloudflare account.",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The email associated with the Cloudflare account. If an API key is provided (and not a token), this field is also required.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Cloudflare account.",
			},
		},
	}
}

func resourceDatadogIntegrationCloudflareAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()

	resp, httpResp, err := apiInstances.GetCloudflareIntegrationApiV2().GetCloudflareAccount(auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error calling GetCloudflareAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return updateIntegrationCloudflareAccountState(d, &resp)
}

func resourceDatadogIntegrationCloudflareAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	body := buildIntegrationCloudflareAccountRequestBody(d)

	resp, httpResp, err := apiInstances.GetCloudflareIntegrationApiV2().CreateCloudflareAccount(auth, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationCloudflareAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationCloudflareAccountState(d, &resp)
}

func buildIntegrationCloudflareAccountRequestBody(d *schema.ResourceData) *datadogV2.CloudflareAccountCreateRequest {
	attributes := datadogV2.NewCloudflareAccountCreateRequestAttributesWithDefaults()

	attributes.SetApiKey(d.Get("api_key").(string))

	if email, ok := d.GetOk("email"); ok {
		attributes.SetEmail(email.(string))
	}

	attributes.SetName(d.Get("name").(string))

	req := datadogV2.NewCloudflareAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewCloudflareAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogIntegrationCloudflareAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	body := buildIntegrationCloudflareAccountUpdateRequestBody(d)

	resp, httpResp, err := apiInstances.GetCloudflareIntegrationApiV2().UpdateCloudflareAccount(auth, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationCloudflareAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationCloudflareAccountState(d, &resp)
}

func buildIntegrationCloudflareAccountUpdateRequestBody(d *schema.ResourceData) *datadogV2.CloudflareAccountUpdateRequest {
	attributes := datadogV2.NewCloudflareAccountUpdateRequestAttributesWithDefaults()

	attributes.SetApiKey(d.Get("api_key").(string))

	if email, ok := d.GetOk("email"); ok {
		attributes.SetEmail(email.(string))
	}

	req := datadogV2.NewCloudflareAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewCloudflareAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogIntegrationCloudflareAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()

	httpResp, err := apiInstances.GetCloudflareIntegrationApiV2().DeleteCloudflareAccount(auth, id)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting IntegrationCloudflareAccount")
	}

	return nil
}

func updateIntegrationCloudflareAccountState(d *schema.ResourceData, resp *datadogV2.CloudflareAccountResponse) diag.Diagnostics {
	data := resp.GetData()
	attributes := data.GetAttributes()

	if email, ok := attributes.GetEmailOk(); ok {
		if err := d.Set("email", email); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := d.Set("name", attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
