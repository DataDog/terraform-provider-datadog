package datadog

import (
	"fmt"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func accountAndLambdaArnFromID(id string) (string, string, error) {
	result := strings.Split(id, " ")
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: %s", id)
	}
	return result[0], result[1], nil
}

func buildDatadogIntegrationAwsLambdaArnStruct(d *schema.ResourceData) *datadogV1.AWSAccountAndLambdaRequest {
	accountID := d.Get("account_id").(string)
	lambdaArn := d.Get("lambda_arn").(string)

	attachLambdaArnRequest := datadogV1.NewAWSAccountAndLambdaRequest(accountID, lambdaArn)
	return attachLambdaArnRequest
}

func resourceDatadogIntegrationAwsLambdaArn() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the log collection Lambdas for an account.\n\nUpdate operations are currently not supported with datadog API so any change forces a new resource.",
		Create:      resourceDatadogIntegrationAwsLambdaArnCreate,
		Read:        resourceDatadogIntegrationAwsLambdaArnRead,
		Delete:      resourceDatadogIntegrationAwsLambdaArnDelete,
		Exists:      resourceDatadogIntegrationAwsLambdaArnExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogIntegrationAwsLambdaArnImport,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // waits for update API call support
			},
			"lambda_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true, // waits for update API call support
			},
		},
	}
}

func resourceDatadogIntegrationAwsLambdaArnExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return false, translateClientError(err, "error getting aws log integrations for datadog account.")
	}

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return false, translateClientError(err, fmt.Sprintf("error getting aws account ID and lambda ARN from id: %s", d.Id()))
	}

	for _, logCollection := range logCollections {
		if logCollection.GetAccountId() == accountID {
			for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
				if lambdaArn == logCollectionLambdaArn.GetArn() {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func resourceDatadogIntegrationAwsLambdaArnCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	attachLambdaArnRequest := buildDatadogIntegrationAwsLambdaArnStruct(d)
	_, _, err := datadogClientV1.AWSLogsIntegrationApi.CreateAWSLambdaARN(authV1).Body(*attachLambdaArnRequest).Execute()
	if err != nil {
		return translateClientError(err, "error attaching Lambda ARN to AWS integration account")
	}

	d.SetId(fmt.Sprintf("%s %s", attachLambdaArnRequest.GetAccountId(), attachLambdaArnRequest.GetLambdaArn()))

	return resourceDatadogIntegrationAwsLambdaArnRead(d, meta)
}

func resourceDatadogIntegrationAwsLambdaArnRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return translateClientError(err, fmt.Sprintf("error getting aws account ID and lambda ARN from id: %s", d.Id()))
	}

	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return translateClientError(err, "error getting aws log integrations for datadog account.")
	}
	for _, logCollection := range logCollections {
		if logCollection.GetAccountId() == accountID {
			for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
				if lambdaArn == logCollectionLambdaArn.GetArn() {
					d.Set("account_id", logCollection.GetAccountId())
					d.Set("lambda_arn", logCollectionLambdaArn.GetArn())
					return nil
				}
			}
		}
	}
	return fmt.Errorf("error getting an AWS log Lambda: account_id=%s, lambda_arn=%s", accountID, lambdaArn)
}

func resourceDatadogIntegrationAwsLambdaArnDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return translateClientError(err, fmt.Sprintf("error parsing account ID and lamdba ARN from ID: %s", d.Id()))
	}

	attachLambdaArnRequest := datadogV1.NewAWSAccountAndLambdaRequest(accountID, lambdaArn)
	_, _, err = datadogClientV1.AWSLogsIntegrationApi.DeleteAWSLambdaARN(authV1).Body(*attachLambdaArnRequest).Execute()
	if err != nil {
		return translateClientError(err, "error deleting an AWS integration Lambda ARN")
	}

	return nil
}

func resourceDatadogIntegrationAwsLambdaArnImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationAwsLambdaArnRead(d, meta); err != nil {
		return nil, translateClientError(err, "error importing lambda arn resource.")
	}
	return []*schema.ResourceData{d}, nil
}
