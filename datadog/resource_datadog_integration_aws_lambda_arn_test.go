package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
	"testing"
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
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkIntegrationAWSLambdaArnDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLambdaArnConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLambdaArnExists,
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

func checkIntegrationAWSLambdaArnExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		lambdaArn := r.Primary.Attributes["lambda_arn"]
		for _, logCollection := range *logCollections {
			for _, logCollectionLambdaArn := range logCollection.LambdaARNs {
				if *logCollection.AccountID == accountId && *logCollectionLambdaArn.LambdaARN == lambdaArn {
					return nil
				}
			}
		}
		return fmt.Errorf("The AWS Lambda ARN is not attached to the account: accountId=%s, lambdaArn=%s", accountId, lambdaArn)
	}
	return nil
}

func checkIntegrationAWSLambdaArnDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	logCollections, err := client.GetIntegrationAWSLogCollection()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		lambdaArn := r.Primary.Attributes["lambda_arn"]
		for _, logCollection := range *logCollections {
			for _, logCollectionLambdaArn := range logCollection.LambdaARNs {
				if *logCollection.AccountID == accountId && *logCollectionLambdaArn.LambdaARN == lambdaArn {
					return fmt.Errorf("The AWS Lambda ARN is still attached to the account: accountId=%s, lambdaArn=%s", accountId, lambdaArn)
				}
			}
		}
	}
	return nil
}
