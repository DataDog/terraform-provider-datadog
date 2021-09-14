package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testAccDatadogIntegrationAWSLambdaArnConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  account_id                       = "%s"
  role_name                        = "testacc-datadog-integration-role"
}

resource "datadog_integration_aws_lambda_arn" "main_collector" {
  account_id = "%s"
  lambda_arn = "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
  depends_on = [datadog_integration_aws.account]
}`, uniq, uniq)
}

func testAccDatadogIntegrationAWSLambdaArnConfigAccessKey(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  access_key_id     = "%s"
  secret_access_key = "testacc-datadog-integration-secret"
}

resource "datadog_integration_aws_lambda_arn" "main_collector" {
  account_id = datadog_integration_aws.account.access_key_id
  lambda_arn = "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
  depends_on = [datadog_integration_aws.account]
}`, uniq)
}

func TestAccDatadogIntegrationAWSLambdaArnAccessKey(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	accessKeyID := uniqueAWSAccessKeyID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAWSLambdaArnDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLambdaArnConfigAccessKey(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLambdaArnExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_lambda_arn.main_collector",
						"account_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_lambda_arn.main_collector",
						"lambda_arn", "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"),
				),
			},
		},
	})
}

func TestAccDatadogIntegrationAWSLambdaArn(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAWSLambdaArnDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLambdaArnConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLambdaArnExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_lambda_arn.main_collector",
						"account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_lambda_arn.main_collector",
						"lambda_arn", "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"),
				),
			},
		},
	})
}

func checkIntegrationAWSLambdaArnExists(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAwsLambdaArnExistsHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAwsLambdaArnExistsHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(ctx)
	if err != nil {
		return err
	}

	for resourceType, r := range s.RootModule().Resources {
		if strings.Contains(resourceType, "datadog_integration_aws_lambda_arn") {
			accountID := r.Primary.Attributes["account_id"]
			lambdaArn := r.Primary.Attributes["lambda_arn"]
			for _, logCollection := range logCollections {
				for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
					if logCollection.GetAccountId() == accountID && logCollectionLambdaArn.GetArn() == lambdaArn {
						return nil
					}
				}
			}
			return fmt.Errorf("The AWS Lambda ARN is not attached to the account: accountID=%s, lambdaArn=%s", accountID, lambdaArn)
		}
	}
	return fmt.Errorf("Unable to find AWS Lambda ARN in any account")
}

func checkIntegrationAWSLambdaArnDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSLambdaArnDestroyHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAWSLambdaArnDestroyHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(ctx)
	if err != nil {
		return err
	}

	err = utils.Retry(2, 5, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Primary.ID != "" {
				accountID := r.Primary.Attributes["account_id"]
				lambdaArn := r.Primary.Attributes["lambda_arn"]
				for _, logCollection := range logCollections {
					for _, logCollectionLambdaArn := range logCollection.GetLambdas() {
						if logCollection.GetAccountId() == accountID && logCollectionLambdaArn.GetArn() == lambdaArn {
							return &utils.RetryableError{Prob: fmt.Sprintf("The AWS Lambda ARN is still attached to the account: accountID=%s, lambdaArn=%s", accountID, lambdaArn)}
						}
					}
				}
			}
		}
		return nil
	})
	return err
}
