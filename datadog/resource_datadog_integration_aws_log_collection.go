package datadog

import (
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceDatadogIntegrationAwsLogCollection() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog - Amazon Web Services integration log collection resource. This can be used to manage which AWS services logs are collected from for an account.",
		Create:      resourceDatadogIntegrationAwsLogCollectionCreate,
		Read:        resourceDatadogIntegrationAwsLogCollectionRead,
		Update:      resourceDatadogIntegrationAwsLogCollectionUpdate,
		Delete:      resourceDatadogIntegrationAwsLogCollectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationAwsLogCollectionImport,
		},

		Schema: map[string]*schema.Schema{
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

func resourceDatadogIntegrationAwsLogCollectionCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID := d.Get("account_id").(string)

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	_, _, err := datadogClientV1.AWSLogsIntegrationApi.EnableAWSLogServices(authV1).Body(*enableLogCollectionServices).Execute()
	if err != nil {
		return TranslateClientError(err, "error enabling log collection services for Amazon Web Services integration account")
	}

	d.SetId(accountID)

	return resourceDatadogIntegrationAwsLogCollectionRead(d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	_, _, err := datadogClientV1.AWSLogsIntegrationApi.EnableAWSLogServices(authV1).Body(*enableLogCollectionServices).Execute()
	if err != nil {
		return TranslateClientError(err, "error updating log collection services for Amazon Web Services integration account")
	}

	return resourceDatadogIntegrationAwsLogCollectionRead(d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID := d.Id()

	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return TranslateClientError(err, "error getting log collection for aws integration.")
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

func resourceDatadogIntegrationAwsLogCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID := d.Id()
	services := []string{}
	deleteLogCollectionServices := datadogV1.NewAWSLogsServicesRequest(accountID, services)
	_, _, err := datadogClientV1.AWSLogsIntegrationApi.EnableAWSLogServices(authV1).Body(*deleteLogCollectionServices).Execute()

	if err != nil {
		return TranslateClientError(err, "error disabling Amazon Web Services log collection")
	}

	return nil
}

func resourceDatadogIntegrationAwsLogCollectionImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationAwsLogCollectionRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
