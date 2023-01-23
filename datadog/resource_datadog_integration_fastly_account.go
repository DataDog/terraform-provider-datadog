package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationFastlyAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog IntegrationFastlyAccount resource. This can be used to create and manage Datadog integration_fastly_account.",
		ReadContext:   resourceDatadogIntegrationFastlyAccountRead,
		CreateContext: resourceDatadogIntegrationFastlyAccountCreate,
		UpdateContext: resourceDatadogIntegrationFastlyAccountUpdate,
		DeleteContext: resourceDatadogIntegrationFastlyAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API key for the Fastly account.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Fastly account.",
			},
			"services": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of services belonging to the parent account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The id of the Fastly service",
						},
						"tags": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "A list of tags for the Fastly service.",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
		},
	}
}

func resourceDatadogIntegrationFastlyAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()

	resp, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().GetFastlyAccount(auth, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error calling GetFastlyAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return updateIntegrationFastlyAccountState(d, &resp)
}

func resourceDatadogIntegrationFastlyAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	body := buildIntegrationFastlyAccountRequestBody(d)

	resp, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().CreateFastlyAccount(auth, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationFastlyAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationFastlyAccountState(d, &resp)
}

func buildIntegrationFastlyAccountRequestBody(d *schema.ResourceData) *datadogV2.FastlyAccountCreateRequest {
	attributes := datadogV2.NewFastlyAccountCreateRequestAttributesWithDefaults()

	attributes.SetApiKey(d.Get("api_key").(string))

	attributes.SetName(d.Get("name").(string))
	services := []datadogV2.FastlyService{}
	for _, s := range d.Get("services").([]interface{}) {
		sMap := s.(map[string]interface{})
		servicesItem := datadogV2.NewFastlyServiceWithDefaults()
		servicesItem.SetId(sMap["id"].(string))

		tags := []string{}
		for _, tagsItem := range sMap["tags"].([]interface{}) {
			tags = append(tags, tagsItem.(string))
		}
		servicesItem.SetTags(tags)

		services = append(services, *servicesItem)
	}
	attributes.SetServices(services)

	req := datadogV2.NewFastlyAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogIntegrationFastlyAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	body := buildIntegrationFastlyAccountUpdateRequestBody(d)

	resp, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().UpdateFastlyAccount(auth, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationFastlyAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationFastlyAccountState(d, &resp)
}

func buildIntegrationFastlyAccountUpdateRequestBody(d *schema.ResourceData) *datadogV2.FastlyAccountUpdateRequest {
	attributes := datadogV2.NewFastlyAccountUpdateRequestAttributesWithDefaults()

	if apiKey, ok := d.GetOk("api_key"); ok {
		attributes.SetApiKey(apiKey.(string))
	}

	req := datadogV2.NewFastlyAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogIntegrationFastlyAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()

	httpResp, err := apiInstances.GetFastlyIntegrationApiV2().DeleteFastlyAccount(auth, id)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting IntegrationFastlyAccount")
	}

	return nil
}

func updateIntegrationFastlyAccountState(d *schema.ResourceData, resp *datadogV2.FastlyAccountResponse) diag.Diagnostics {
	data := resp.GetData()
	attributes := data.GetAttributes()

	if err := d.Set("name", attributes.GetName()); err != nil {
		return diag.FromErr(err)
	}

	servicesTf := make([]map[string]interface{}, 0)
	for _, servicesDd := range attributes.GetServices() {
		servicesTfItem := map[string]interface{}{}
		servicesTfItem["id"] = servicesDd.GetId()
		servicesTfItem["tags"] = servicesDd.GetTags()

		servicesTf = append(servicesTf, servicesTfItem)

	}
	if err := d.Set("services", servicesTf); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
