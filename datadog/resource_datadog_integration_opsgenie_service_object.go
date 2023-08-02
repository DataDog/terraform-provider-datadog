package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationOpsgenieService() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource for interacting with Datadog Opsgenie Service API.",
		CreateContext: resourceDatadogIntegrationOpsgenieServiceCreate,
		ReadContext:   resourceDatadogIntegrationOpsgenieServiceRead,
		UpdateContext: resourceDatadogIntegrationOpsgenieServiceUpdate,
		DeleteContext: resourceDatadogIntegrationOpsgenieServiceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"name": {
					Description: "The name for the Opsgenie service.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"opsgenie_api_key": {
					Description: "The Opsgenie API key for the Opsgenie service. Note: Since the Datadog API never returns Opsgenie API keys, it is impossible to detect [drifts](https://www.hashicorp.com/blog/detecting-and-managing-drift-with-terraform). The best way to solve a drift is to manually mark the Service Object resource with [terraform taint](https://www.terraform.io/docs/commands/taint.html) to have it destroyed and recreated.",
					Type:        schema.TypeString,
					Required:    true,
					Sensitive:   true,
				},
				"region": {
					Description:      "The region for the Opsgenie service.",
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: validators.ValidateEnumValue(datadogV2.NewOpsgenieServiceRegionTypeFromValue),
				},
				"custom_url": {
					Description: "The custom url for a custom region.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			}
		},
	}
}

func buildOpsgenieServiceCreateRequest(d *schema.ResourceData) *datadogV2.OpsgenieServiceCreateRequest {
	region := datadogV2.OpsgenieServiceRegionType(d.Get("region").(string))
	serviceAttributes := datadogV2.NewOpsgenieServiceCreateAttributes(d.Get("name").(string), d.Get("opsgenie_api_key").(string), region)
	if customUrl, ok := d.GetOk("custom_url"); ok {
		serviceAttributes.SetCustomUrl(customUrl.(string))
	}
	serviceData := datadogV2.NewOpsgenieServiceCreateData(*serviceAttributes, datadogV2.OPSGENIESERVICETYPE_OPSGENIE_SERVICE)
	serviceRequest := datadogV2.NewOpsgenieServiceCreateRequest(*serviceData)

	return serviceRequest
}

func buildOpsgenieServiceUpdateRequest(d *schema.ResourceData) *datadogV2.OpsgenieServiceUpdateRequest {
	region := datadogV2.OpsgenieServiceRegionType(d.Get("region").(string))
	serviceAttributes := datadogV2.NewOpsgenieServiceUpdateAttributesWithDefaults()
	serviceAttributes.SetName(d.Get("name").(string))
	serviceAttributes.SetOpsgenieApiKey(d.Get("opsgenie_api_key").(string))
	serviceAttributes.SetRegion(region)
	if customUrl, ok := d.GetOk("custom_url"); ok {
		serviceAttributes.SetCustomUrl(customUrl.(string))
	}
	serviceData := datadogV2.NewOpsgenieServiceUpdateData(*serviceAttributes, d.Id(), datadogV2.OPSGENIESERVICETYPE_OPSGENIE_SERVICE)
	serviceRequest := datadogV2.NewOpsgenieServiceUpdateRequest(*serviceData)

	return serviceRequest
}

func updateOpsgenieServiceState(d *schema.ResourceData, serviceData *datadogV2.OpsgenieServiceResponseData) diag.Diagnostics {
	serviceAttributes := serviceData.GetAttributes()

	if err := d.Set("name", serviceAttributes.GetName()); err != nil {
		return diag.FromErr(err)
	}
	// Only update opsgenie_api_key if not set on d - the API endpoints never return
	// the keys, so this is how we recognize new values.
	if _, ok := d.GetOk("opsgenie_api_key"); !ok {
		if err := d.Set("opsgenie_api_key", maskedSecret); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("region", serviceAttributes.GetRegion()); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("custom_url", serviceAttributes.GetCustomUrl()); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceDatadogIntegrationOpsgenieServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetOpsgenieIntegrationApiV2().CreateOpsgenieService(auth, *buildOpsgenieServiceCreateRequest(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error creating Opsgenie service")
	}

	serviceData := resp.GetData()
	d.SetId(serviceData.GetId())

	return updateOpsgenieServiceState(d, &serviceData)
}

func resourceDatadogIntegrationOpsgenieServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetOpsgenieIntegrationApiV2().GetOpsgenieService(auth, d.Id())
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpResponse, "error getting Opsgenie service")
	}
	serviceData := resp.GetData()
	return updateOpsgenieServiceState(d, &serviceData)
}

func resourceDatadogIntegrationOpsgenieServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	resp, httpResponse, err := apiInstances.GetOpsgenieIntegrationApiV2().UpdateOpsgenieService(auth, d.Id(), *buildOpsgenieServiceUpdateRequest(d))
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error updating Opsgenie service")
	}
	serviceData := resp.GetData()
	return updateOpsgenieServiceState(d, &serviceData)
}

func resourceDatadogIntegrationOpsgenieServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	if httpResponse, err := apiInstances.GetOpsgenieIntegrationApiV2().DeleteOpsgenieService(auth, d.Id()); err != nil {
		return utils.TranslateClientErrorDiag(err, httpResponse, "error deleting Opsgenie service")
	}

	return nil
}
