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

func resourceDatadogIntegrationAwsLambdaArn() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogIntegrationAwsLambdaArnCreate,
		Read:   resourceDatadogIntegrationAwsLambdaArnRead,
		Delete: resourceDatadogIntegrationAwsLambdaArnDelete,
		Exists: resourceDatadogIntegrationAwsLambdaArnExists,
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
	client := meta.(*datadog.Client)

	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return false, err
	}

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return false, err
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
	client := meta.(*datadog.Client)

	accountID := d.Get("account_id").(string)
	lambdaArn := d.Get("lambda_arn").(string)

	attachLambdaArnRequest := datadog.IntegrationAWSLambdaARNRequest{
		AccountID: &accountID,
		LambdaARN: &lambdaArn,
	}
	err := client.AttachLambdaARNIntegrationAWS(&attachLambdaArnRequest)

	if err != nil {
		return fmt.Errorf("error attaching Lambda ARN to AWS integration account: %s", err.Error())
	}

	d.SetId(fmt.Sprintf("%s %s", accountID, lambdaArn))

	return resourceDatadogIntegrationAwsLambdaArnRead(d, meta)
}

func resourceDatadogIntegrationAwsLambdaArnRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*datadog.Client)

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return err
	}

	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return err
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
	client := meta.(*datadog.Client)

	accountID, lambdaArn, err := accountAndLambdaArnFromID(d.Id())
	if err != nil {
		return err
	}

	attachLambdaArnRequest := datadog.IntegrationAWSLambdaARNRequest{
		AccountID: &accountID,
		LambdaARN: &lambdaArn,
	}

	err = client.DeleteAWSLogCollection(&attachLambdaArnRequest)

	if err != nil {
		return fmt.Errorf("error deleting an AWS integration Lambda ARN: %s", err.Error())
	}

	return nil
}

func resourceDatadogIntegrationAwsLambdaArnImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogIntegrationAwsLambdaArnRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
