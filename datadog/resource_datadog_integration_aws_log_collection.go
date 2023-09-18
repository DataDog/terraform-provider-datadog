package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceDatadogIntegrationAwsLogCollection() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.",
		CreateContext: resourceDatadogIntegrationAwsLogCollectionCreate,
		ReadContext:   resourceDatadogIntegrationAwsLogCollectionRead,
		UpdateContext: resourceDatadogIntegrationAwsLogCollectionUpdate,
		DeleteContext: resourceDatadogIntegrationAwsLogCollectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"account_id": {
					Description: "Your AWS Account ID without dashes.",
					Type:        schema.TypeString,
					Required:    true,
					ForceNew:    true,
				},
				"services": {
					Description: "A list of services to collect logs from. See the [api docs](https://docs.datadoghq.com/api/v1/aws-logs-integration/#get-list-of-aws-log-ready-services) for more details on which services are supported.",
					Type:        schema.TypeList,
					Required:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			}
		},
	}
}

func buildDatadogIntegrationAwsLogCollectionStruct(d *schema.ResourceData) *datadogV1.AWSLogsServicesRequest {
	accountID := d.Get("account_id").(string)
	services := []string{}
	if attr, ok := d.GetOk("services"); ok {
		for _, s := range attr.([]interface{}) {
			services = append(services, s.(string))
		}
	}

	enableLogCollectionServices := datadogV1.NewAWSLogsServicesRequest(accountID, services)

	return enableLogCollectionServices
}

func resourceDatadogIntegrationAwsLogCollectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	// shared with datadog_integration_aws resource
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID := d.Get("account_id").(string)

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	response, httpresp, err := apiInstances.GetAWSLogsIntegrationApiV1().EnableAWSLogServices(auth, *enableLogCollectionServices)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error enabling log collection services for Amazon Web Services integration account")
	}
	res := response.(map[string]interface{})
	if status, ok := res["status"]; ok && status == "error" {
		return diag.FromErr(fmt.Errorf("error creating aws log collection: %s", httpresp.Body))
	}

	d.SetId(accountID)

	return resourceDatadogIntegrationAwsLogCollectionRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	// shared with datadog_integration_aws resource
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	_, httpresp, err := apiInstances.GetAWSLogsIntegrationApiV1().EnableAWSLogServices(auth, *enableLogCollectionServices)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating log collection services for Amazon Web Services integration account")
	}

	return resourceDatadogIntegrationAwsLogCollectionRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	accountID := d.Id()

	logCollections, httpresp, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsIntegrations(auth)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting log collection for aws integration.")
	}
	if err := utils.CheckForUnparsed(logCollections); err != nil {
		return diag.FromErr(err)
	}
	for _, logCollection := range logCollections {
		if logCollection.GetAccountId() == accountID {
			d.Set("account_id", logCollection.GetAccountId())
			d.Set("services", logCollection.GetServices())
			return nil
		}
	}

	d.SetId("")
	return nil
}

func resourceDatadogIntegrationAwsLogCollectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	// shared with datadog_integration_aws resource
	integrationAwsMutex.Lock()
	defer integrationAwsMutex.Unlock()

	accountID := d.Id()
	services := []string{}
	deleteLogCollectionServices := datadogV1.NewAWSLogsServicesRequest(accountID, services)
	_, httpresp, err := apiInstances.GetAWSLogsIntegrationApiV1().EnableAWSLogServices(auth, *deleteLogCollectionServices)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error disabling Amazon Web Services log collection")
	}

	return nil
}
