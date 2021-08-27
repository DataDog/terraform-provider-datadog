package datadog

import (
	"context"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
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

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "Your AWS Account ID without dashes. If your account is a GovCloud or China account, specify the `access_key_id` here.",
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID := d.Get("account_id").(string)

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	_, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.EnableAWSLogServices(authV1, *enableLogCollectionServices)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error enabling log collection services for Amazon Web Services integration account")
	}

	d.SetId(accountID)

	return resourceDatadogIntegrationAwsLogCollectionRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	_, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.EnableAWSLogServices(authV1, *enableLogCollectionServices)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating log collection services for Amazon Web Services integration account")
	}

	return resourceDatadogIntegrationAwsLogCollectionRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID := d.Id()

	logCollections, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1)
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
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID := d.Id()
	services := []string{}
	deleteLogCollectionServices := datadogV1.NewAWSLogsServicesRequest(accountID, services)
	_, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.EnableAWSLogServices(authV1, *deleteLogCollectionServices)

	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error disabling Amazon Web Services log collection")
	}

	return nil
}
