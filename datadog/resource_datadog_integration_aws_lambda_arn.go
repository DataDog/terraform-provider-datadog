package datadog

import (
	"context"
	"fmt"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func buildDatadogIntegrationAwsLambdaArnStruct(d *schema.ResourceData) *datadogV1.AWSAccountAndLambdaRequest {
	accountID := d.Get("account_id").(string)
	lambdaArn := d.Get("lambda_arn").(string)

	attachLambdaArnRequest := datadogV1.NewAWSAccountAndLambdaRequest(accountID, lambdaArn)
	return attachLambdaArnRequest
}

func resourceDatadogIntegrationAwsLambdaArn() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a Datadog - Amazon Web Services integration Lambda ARN resource. This can be used to create and manage the log collection Lambdas for an account.\n\nUpdate operations are currently not supported with datadog API so any change forces a new resource.",
		CreateContext: resourceDatadogIntegrationAwsLambdaArnCreate,
		ReadContext:   resourceDatadogIntegrationAwsLambdaArnRead,
		DeleteContext: resourceDatadogIntegrationAwsLambdaArnDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Description: "Your AWS Account ID without dashes. If your account is a GovCloud or China account, specify the `access_key_id` here.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // waits for update API call support
			},
			"lambda_arn": {
				Description: "The ARN of the Datadog forwarder Lambda.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // waits for update API call support
			},
		},
	}
}

func resourceDatadogIntegrationAwsLambdaArnCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	attachLambdaArnRequest := buildDatadogIntegrationAwsLambdaArnStruct(d)
	response, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.CreateAWSLambdaARN(authV1, *attachLambdaArnRequest)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error attaching Lambda ARN to AWS integration account")
	}
	if err := utils.CheckForUnparsed(response); err != nil {
		return diag.FromErr(err)
	}

	res := response.(map[string]interface{})
	if status, ok := res["status"]; ok && status == "error" {
		return diag.FromErr(fmt.Errorf("error attaching Lambda ARN to AWS integration account: %s", httpresp.Body))
	}

	d.SetId(fmt.Sprintf("%s %s", attachLambdaArnRequest.GetAccountId(), attachLambdaArnRequest.GetLambdaArn()))

	return resourceDatadogIntegrationAwsLambdaArnRead(ctx, d, meta)
}

func resourceDatadogIntegrationAwsLambdaArnRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, lambdaArn, err := utils.AccountAndLambdaArnFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	logCollections, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting aws log integrations for datadog account.")
	}
	if err := utils.CheckForUnparsed(logCollections); err != nil {
		return diag.FromErr(err)
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

	d.SetId("")
	return nil
}

func resourceDatadogIntegrationAwsLambdaArnDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1

	accountID, lambdaArn, err := utils.AccountAndLambdaArnFromID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	attachLambdaArnRequest := datadogV1.NewAWSAccountAndLambdaRequest(accountID, lambdaArn)
	_, httpresp, err := datadogClientV1.AWSLogsIntegrationApi.DeleteAWSLambdaARN(authV1, *attachLambdaArnRequest)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting an AWS integration Lambda ARN")
	}

	return nil
}
