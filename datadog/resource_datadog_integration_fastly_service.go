package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationFastlyService() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog IntegrationFastlyService resource. This can be used to create and manage Datadog integration_fastly_service.",
		ReadContext:   resourceDatadogIntegrationFastlyServiceRead,
		CreateContext: resourceDatadogIntegrationFastlyServiceCreate,
		UpdateContext: resourceDatadogIntegrationFastlyServiceUpdate,
		DeleteContext: resourceDatadogIntegrationFastlyServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UPDATE ME",
			},

			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of tags for the Fastly service.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogIntegrationFastlyServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	accountId := d.Get("account_id").(string)
	id := d.Id()

	resp, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().GetFastlyService(auth, accountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error calling GetFastlyService")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return updateIntegrationFastlyServiceState(d, &resp)
}

func resourceDatadogIntegrationFastlyServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountId := d.Get("account_id").(string)

	body := buildIntegrationFastlyServiceRequestBody(d)

	resp, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().CreateFastlyService(auth, accountId, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationFastlyService")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationFastlyServiceState(d, &resp)
}

func buildIntegrationFastlyServiceRequestBody(d *schema.ResourceData) *datadogV2.FastlyServiceRequest {
	attributes := datadogV2.NewFastlyServiceAttributesWithDefaults()
	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	req := datadogV2.NewFastlyServiceRequestWithDefaults()
	req.Data = *datadogV2.NewFastlyServiceDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogIntegrationFastlyServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountId := d.Get("account_id").(string)

	id := d.Id()

	body := buildIntegrationFastlyServiceRequestBody(d)

	resp, httpResp, err := apiInstances.GetFastlyIntegrationApiV2().UpdateFastlyService(auth, accountId, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationFastlyService")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationFastlyServiceState(d, &resp)
}

func resourceDatadogIntegrationFastlyServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	accountId := d.Get("account_id").(string)
	id := d.Id()

	httpResp, err := apiInstances.GetFastlyIntegrationApiV2().DeleteFastlyService(auth, accountId, id)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting IntegrationFastlyService")
	}

	return nil
}

func updateIntegrationFastlyServiceState(d *schema.ResourceData, resp *datadogV2.FastlyServiceResponse) diag.Diagnostics {
	data := resp.GetData()
	attributes := data.GetAttributes()

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		if err := d.Set("tags", *tags); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
