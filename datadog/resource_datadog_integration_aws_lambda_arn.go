package datadog

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func accountAndLambdaArnFromID(id string) (string, string, error) {
	result := strings.Split(id, " ")
	if len(result) != 2 {
		return "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: %s", id)
	}
	return result[0], result[1], nil
}

func buildDatadogIntegrationAwsLambdaArnStruct(d *schema.ResourceData) *datadog.IntegrationAWSLambdaARNRequest {
	accountID := d.Get("account_id").(string)
	lambdaArn := d.Get("lambda_arn").(string)

	attachLambdaArnRequest := datadog.IntegrationAWSLambdaARNRequest{
		AccountID: &accountID,
		LambdaARN: &lambdaArn,
	}
	return &attachLambdaArnRequest
}

func resourceDatadogIntegrationAwsLambdaArn() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationAwsLambdaArnCreate,
		Read:   resourceDatadogIntegrationAwsLambdaArnRead,
		Delete: resourceDatadogIntegrationAwsLambdaArnDelete,
		Exists: resourceDatadogIntegrationAwsLambdaArnExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
	client := providerConf.CommunityClient

	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return false, translateClientError(err, "error getting aws log integrations for datadog account.")
	}

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return false, translateClientError(err, fmt.Sprintf("error getting aws account ID and lambda ARN from id: %s", d.Id()))
	}

	for _, logCollection := range *logCollections {
		if logCollection.GetAccountID() == accountID {
			for _, logCollectionLambdaArn := range logCollection.LambdaARNs {
				if lambdaArn == logCollectionLambdaArn.GetLambdaARN() {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

func resourceDatadogIntegrationAwsLambdaArnCreate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	attachLambdaArnRequest := buildDatadogIntegrationAwsLambdaArnStruct(d)
	err := client.AttachLambdaARNIntegrationAWS(attachLambdaArnRequest)

	if err != nil {
		return translateClientError(err, "error attaching Lambda ARN to AWS integration account")
	}

	d.SetId(fmt.Sprintf("%s %s", *attachLambdaArnRequest.AccountID, *attachLambdaArnRequest.LambdaARN))

	return resourceDatadogIntegrationAwsLambdaArnRead(d, meta)
}

func resourceDatadogIntegrationAwsLambdaArnRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return translateClientError(err, fmt.Sprintf("error getting aws account ID and lambda ARN from id: %s", d.Id()))
	}

	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return translateClientError(err, "error getting aws log integrations for datadog account.")
	}
	for _, logCollection := range *logCollections {
		if logCollection.GetAccountID() == accountID {
			for _, logCollectionLambdaArn := range logCollection.LambdaARNs {
				if lambdaArn == logCollectionLambdaArn.GetLambdaARN() {
					d.Set("account_id", logCollection.GetAccountID())
					d.Set("lambda_arn", logCollectionLambdaArn.GetLambdaARN())
					return nil
				}
			}
		}
	}
	return fmt.Errorf("error getting an AWS log Lambda: account_id=%s, lambda_arn=%s", accountID, lambdaArn)
}

func resourceDatadogIntegrationAwsLambdaArnDelete(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return translateClientError(err, fmt.Sprintf("error parsing account ID and lamdba ARN from ID: %s", d.Id()))
	}

	attachLambdaArnRequest := datadog.IntegrationAWSLambdaARNRequest{
		AccountID: &accountID,
		LambdaARN: &lambdaArn,
	}

	err = client.DeleteAWSLogCollection(&attachLambdaArnRequest)

	if err != nil {
		return translateClientError(err, "error deleting an AWS integration Lambda ARN")
	}

	return nil
}
