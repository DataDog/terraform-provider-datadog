package datadog

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogIntegrationAwsLogCollection() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationAwsLogCollectionCreate,
		Read:   resourceDatadogIntegrationAwsLogCollectionRead,
		Update: resourceDatadogIntegrationAwsLogCollectionUpdate,
		Delete: resourceDatadogIntegrationAwsLogCollectionDelete,
		Exists: resourceDatadogIntegrationAwsLogCollectionExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"services": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceDatadogIntegrationAwsLogCollectionExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return false, translateClientError(err, "error getting aws integration log collection.")
	}

	accountID := d.Id()

	for _, logCollection := range *logCollections {
		if logCollection.GetAccountID() == accountID {
			return true, nil
		}
	}
	return false, nil
}

func buildDatadogIntegrationAwsLogCollectionStruct(d *schema.ResourceData) datadog.IntegrationAWSServicesLogCollection {
	accountID := d.Get("account_id").(string)
	services := []string{}
	if attr, ok := d.GetOk("services"); ok {
		for _, s := range attr.([]interface{}) {
			services = append(services, s.(string))
		}
	}

	enableLogCollectionServices := datadog.IntegrationAWSServicesLogCollection{
		AccountID: &accountID,
		Services:  services,
	}

	return enableLogCollectionServices
}

func resourceDatadogIntegrationAwsLogCollectionCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	accountID := d.Get("account_id").(string)

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	err := client.EnableLogCollectionAWSServices(&enableLogCollectionServices)

	if err != nil {
		return translateClientError(err, "error enabling log collection services for Amazon Web Services integration account")
	}

	d.SetId(accountID)

	return resourceDatadogIntegrationAwsLogCollectionRead(d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	enableLogCollectionServices := buildDatadogIntegrationAwsLogCollectionStruct(d)
	err := client.EnableLogCollectionAWSServices(&enableLogCollectionServices)

	if err != nil {
		return translateClientError(err, "error updating log collection services for Amazon Web Services integration account")
	}

	return resourceDatadogIntegrationAwsLogCollectionRead(d, meta)
}

func resourceDatadogIntegrationAwsLogCollectionRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	accountID := d.Id()

	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return translateClientError(err, "error getting log collection for aws integration.")
	}
	for _, logCollection := range *logCollections {
		if logCollection.GetAccountID() == accountID {
			d.Set("account_id", logCollection.GetAccountID())
			d.Set("services", logCollection.Services)
			return nil
		}
	}
	return fmt.Errorf("error getting Amazon Web Services log collection: account_id=%s", accountID)
}

func resourceDatadogIntegrationAwsLogCollectionDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	accountID := d.Id()
	services := []string{}

	deleteLogCollectionServices := datadog.IntegrationAWSServicesLogCollection{
		AccountID: &accountID,
		Services:  services,
	}
	err := client.EnableLogCollectionAWSServices(&deleteLogCollectionServices)

	if err != nil {
		return translateClientError(err, "error disabling Amazon Web Services log collection")
	}

	return nil
}
