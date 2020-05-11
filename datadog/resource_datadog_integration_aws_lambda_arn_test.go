package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccountAndLambdaArnFromID(t *testing.T) {
	cases := map[string]struct {
		id        string
		accountId string
		lambdaArn string
		err       error
	}{
		"basic":               {"123456789 /aws/lambda/my-fun-funct", "123456789", "/aws/lambda/my-fun-funct", nil},
		"no delimiter":        {"123456789", "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: 123456789")},
		"multiple delimiters": {"123456789 extra bits", "", "", fmt.Errorf("error extracting account ID and Lambda ARN from an AWS integration id: 123456789 extra bits")},
	}
	for name, tc := range cases {
		accountId, lambdaArn, err := accountAndLambdaArnFromID(tc.id)

		if err != nil && tc.err != nil && err.Error() != tc.err.Error() {
			t.Errorf("%s: errors should be '%s', not `%s`", name, tc.err.Error(), err.Error())
		} else if err != nil && tc.err == nil {
			t.Errorf("%s: errors should be nil, not `%s`", name, err.Error())
		} else if err == nil && tc.err != nil {
			t.Errorf("%s: errors should be '%s', not nil", name, tc.err.Error())
		}

		if accountId != tc.accountId {
			t.Errorf("%s: account ID '%s' didn't match `%s`", name, accountId, tc.accountId)
		}
		if lambdaArn != tc.lambdaArn {
			t.Errorf("%s: lambda arn '%s' didn't match `%s`", name, lambdaArn, tc.lambdaArn)
		}
	}
}

const testAccDatadogIntegrationAWSLambdaArnConfig = `
resource "datadog_integration_aws" "account" {
  account_id                       = "1234567890"
  role_name                        = "testacc-datadog-integration-role"
}

resource "datadog_integration_aws_lambda_arn" "main_collector" {
  account_id = "1234567890"
  lambda_arn = "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
  depends_on = [datadog_integration_aws.account]
}
`

func TestAccDatadogIntegrationAWSLambdaArn(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationAWSLambdaArnDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLambdaArnConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLambdaArnExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_lambda_arn.main_collector",
						"account_id", "1234567890"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_lambda_arn.main_collector",
						"lambda_arn", "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"),
				),
			},
		},
	})
}

func checkIntegrationAWSLambdaArnExists(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAwsLambdaArnExistsHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAwsLambdaArnExistsHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return err
	}

	for resourceType, r := range s.RootModule().Resources {
		if strings.Contains(resourceType, "datadog_integration_aws_lambda_arn") {
			accountId := r.Primary.Attributes["account_id"]
			lambdaArn := r.Primary.Attributes["lambda_arn"]
			for _, logCollection := range logCollections {
				for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
					if logCollection.GetAccountId() == accountId && logCollectionLambdaArn.GetArn() == lambdaArn {
						return nil
					}
				}
			}
			return fmt.Errorf("The AWS Lambda ARN is not attached to the account: accountId=%s, lambdaArn=%s", accountId, lambdaArn)
		}
	}
	return fmt.Errorf("Unable to find AWS Lambda ARN in any account")
}

func checkIntegrationAWSLambdaArnDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSLambdaArnDestroyHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAWSLambdaArnDestroyHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		lambdaArn := r.Primary.Attributes["lambda_arn"]
		for _, logCollection := range logCollections {
			for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
				if logCollection.GetAccountId() == accountId && logCollectionLambdaArn.GetArn() == lambdaArn {
					return fmt.Errorf("The AWS Lambda ARN is still attached to the account: accountId=%s, lambdaArn=%s", accountId, lambdaArn)
				}
			}
		}
	}
	return nil
}
