package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationConfluentResource() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog IntegrationConfluentResource resource. This can be used to create and manage Datadog integration_confluent_resource.",
		ReadContext:   resourceDatadogIntegrationConfluentResourceRead,
		CreateContext: resourceDatadogIntegrationConfluentResourceCreate,
		UpdateContext: resourceDatadogIntegrationConfluentResourceUpdate,
		DeleteContext: resourceDatadogIntegrationConfluentResourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "UPDATE ME",
			},

			"resource_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The resource type of the Resource. Can be `kafka`, `connector`, `ksql`, or `schema_registry`.",
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of strings representing tags. Can be a single key, or key-value pairs separated by a colon.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogIntegrationConfluentResourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	accountId := d.Get("account_id").(string)
	id := d.Id()

	resp, httpResp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentResource(auth, accountId, id)
	if err != nil {
		if httpResp != nil && httpResp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResp, "error calling GetConfluentResource")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return updateIntegrationConfluentResourceState(d, &resp)
}

func resourceDatadogIntegrationConfluentResourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountId := d.Get("account_id").(string)

	body := buildIntegrationConfluentResourceRequestBody(d)

	resp, httpResp, err := apiInstances.GetConfluentCloudApiV2().CreateConfluentResource(auth, accountId, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationConfluentResource")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationConfluentResourceState(d, &resp)
}

func buildIntegrationConfluentResourceRequestBody(d *schema.ResourceData) *datadogV2.ConfluentResourceRequest {
	attributes := datadogV2.NewConfluentResourceRequestAttributesWithDefaults()

	if resourceType, ok := d.GetOk("resource_type"); ok {
		attributes.SetResourceType(resourceType.(string))
	}
	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	req := datadogV2.NewConfluentResourceRequestWithDefaults()
	req.Data = *datadogV2.NewConfluentResourceRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogIntegrationConfluentResourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountId := d.Get("account_id").(string)

	id := d.Id()

	body := buildIntegrationConfluentResourceRequestBody(d)

	resp, httpResp, err := apiInstances.GetConfluentCloudApiV2().UpdateConfluentResource(auth, accountId, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResp, "error creating IntegrationConfluentResource")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateIntegrationConfluentResourceState(d, &resp)
}

func resourceDatadogIntegrationConfluentResourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	accountId := d.Get("account_id").(string)
	id := d.Id()

	httpResp, err := apiInstances.GetConfluentCloudApiV2().DeleteConfluentResource(auth, accountId, id)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpResp, "error deleting IntegrationConfluentResource")
	}

	return nil
}

func updateIntegrationConfluentResourceState(d *schema.ResourceData, resp *datadogV2.ConfluentResourceResponse) diag.Diagnostics {
	data := resp.GetData()
	attributes := data.GetAttributes()

	if err := d.Set("resource_type", attributes.GetResourceType()); err != nil {
		return diag.FromErr(err)
	}

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		if err := d.Set("tags", *tags); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
