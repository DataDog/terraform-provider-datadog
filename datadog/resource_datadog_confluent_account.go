package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogConfluentAccount() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog ConfluentAccount resource. This can be used to create and manage Datadog confluent_account.",
		ReadContext:   resourceDatadogConfluentAccountRead,
		CreateContext: resourceDatadogConfluentAccountCreate,
		UpdateContext: resourceDatadogConfluentAccountUpdate,
		DeleteContext: resourceDatadogConfluentAccountDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"api_key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API key associated with your Confluent account.",
			},
			"api_secret": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The API secret associated with your Confluent account.",
			},
			"resources": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of Confluent resources associated with the Confluent account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID associated with a Confluent resource.",
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
				},
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

func resourceDatadogConfluentAccountRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()

	resp, httpresp, err := apiInstances.GetConfluentCloudApiV2().GetConfluentAccount(auth, id)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			// Delete the resource from the local state since it doesn't exist anymore in the actual state
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error calling GetConfluentAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}

	return updateConfluentAccountState(d, &resp)
}

func resourceDatadogConfluentAccountCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	body := buildConfluentAccountRequestBody(d)

	resp, httpresp, err := apiInstances.GetConfluentCloudApiV2().CreateConfluentAccount(auth, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating ConfluentAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateConfluentAccountState(d, &resp)
}

func buildConfluentAccountRequestBody(d *schema.ResourceData) *datadogV2.ConfluentAccountCreateRequest {
	attributes := datadogV2.NewConfluentAccountCreateRequestAttributesWithDefaults()

	attributes.SetApiKey(d.Get("api_key").(string))

	attributes.SetApiSecret(d.Get("api_secret").(string))
	resources := []datadogV2.ConfluentAccountResourceAttributes{}
	for _, s := range d.Get("resources").([]interface{}) {
		sMap := s.(map[string]interface{})
		resourcesItem := datadogV2.NewConfluentAccountResourceAttributesWithDefaults()
		resourcesItem.SetId(sMap["id"].(string))
		resourcesItem.SetResourceType(sMap["resource_type"].(string))

		tags := []string{}
		for _, tagsItem := range sMap["tags"].([]interface{}) {
			tags = append(tags, tagsItem.(string))
		}
		resourcesItem.SetTags(tags)

		resources = append(resources, *resourcesItem)
	}
	attributes.SetResources(resources)

	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	req := datadogV2.NewConfluentAccountCreateRequestWithDefaults()
	req.Data = *datadogV2.NewConfluentAccountCreateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogConfluentAccountUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	id := d.Id()

	body := buildConfluentAccountUpdateRequestBody(d)

	resp, httpresp, err := apiInstances.GetConfluentCloudApiV2().UpdateConfluentAccount(auth, id, *body)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating ConfluentAccount")
	}
	if err := utils.CheckForUnparsed(resp); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(resp.Data.GetId())

	return updateConfluentAccountState(d, &resp)
}

func buildConfluentAccountUpdateRequestBody(d *schema.ResourceData) *datadogV2.ConfluentAccountUpdateRequest {
	attributes := datadogV2.NewConfluentAccountUpdateRequestAttributesWithDefaults()

	attributes.SetApiKey(d.Get("api_key").(string))

	attributes.SetApiSecret(d.Get("api_secret").(string))
	tags := []string{}
	for _, s := range d.Get("tags").([]interface{}) {
		tags = append(tags, s.(string))
	}
	attributes.SetTags(tags)

	req := datadogV2.NewConfluentAccountUpdateRequestWithDefaults()
	req.Data = *datadogV2.NewConfluentAccountUpdateRequestDataWithDefaults()
	req.Data.SetAttributes(*attributes)

	return req
}

func resourceDatadogConfluentAccountDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth
	id := d.Id()

	httpresp, err := apiInstances.GetConfluentCloudApiV2().DeleteConfluentAccount(auth, id)
	if err != nil {
		// The resource is assumed to still exist, and all prior state is preserved.
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting ConfluentAccount")
	}

	return nil
}

func updateConfluentAccountState(d *schema.ResourceData, resp *datadogV2.ConfluentAccountResponse) diag.Diagnostics {
	data := resp.GetData()
	attributes := data.GetAttributes()

	if err := d.Set("api_key", attributes.GetApiKey()); err != nil {
		return diag.FromErr(err)
	}

	resourcesTf := make([]map[string]interface{}, 0)
	for _, resourcesDd := range attributes.GetResources() {
		resourcesTfItem := map[string]interface{}{}
		resourcesTfItem["resource_type"] = resourcesDd.GetResourceType()
		resourcesTfItem["tags"] = resourcesDd.GetTags()

		resourcesTf = append(resourcesTf, resourcesTfItem)

	}
	if err := d.Set("resources", resourcesTf); err != nil {
		return diag.FromErr(err)
	}

	if tags, ok := attributes.GetTagsOk(); ok && len(*tags) > 0 {
		if err := d.Set("tags", *tags); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
