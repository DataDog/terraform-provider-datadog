package test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccDatadogIntegrationAWSLogCollectionConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  account_id                       = "%s"
  role_name                        = "testacc-datadog-integration-role"
}

resource "datadog_integration_aws_lambda_arn" "lambda" {
  account_id = datadog_integration_aws.account.account_id
  lambda_arn = "arn:aws:lambda:us-east-1:%s:function:datadog-forwarder-Forwarder"
  depends_on = [datadog_integration_aws.account]
}

resource "datadog_integration_aws_log_collection" "main" {
  account_id = datadog_integration_aws.account.account_id
  services = ["lambda"]
  depends_on = [datadog_integration_aws_lambda_arn.lambda]
}`, uniq, uniq)
}

func TestAccDatadogIntegrationAWSLogCollection(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationAWSLogCollectionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLogCollectionConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLogCollectionExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_log_collection.main",
						"account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_log_collection.main",
						"services.0", "lambda"),
				),
			},
		},
	})
}

func checkIntegrationAWSLogCollectionExists(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSLogCollectionExistsHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAWSLogCollectionExistsHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return err
	}
	for resourceType, r := range s.RootModule().Resources {
		if strings.Contains(resourceType, "datadog_integration_aws_log_collection") {
			accountID := r.Primary.Attributes["account_id"]
			for _, logCollection := range logCollections {
				if *logCollection.AccountId == accountID && len(logCollection.GetServices()) > 0 {
					return nil
				}
			}
			return fmt.Errorf("The AWS Log Collection is not configured for the account: accountID=%s", accountID)
		}
	}
	return fmt.Errorf("Unable to find AWS Log Collection in any account")
}

func checkIntegrationAWSLogCollectionDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSLogCollectionDestroyHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAWSLogCollectionDestroyHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountID := r.Primary.Attributes["account_id"]
		err = utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Primary.ID != "" {
					for _, logCollection := range logCollections {
						if *logCollection.AccountId == accountID && len(logCollection.GetServices()) > 0 {
							return &utils.RetryableError{Prob: fmt.Sprintf("The AWS Log Collection is still enabled for the account: accountID=%s", accountID)}
						}
					}
				}
			}
			return nil
		})
		return err
	}
	return nil
}
