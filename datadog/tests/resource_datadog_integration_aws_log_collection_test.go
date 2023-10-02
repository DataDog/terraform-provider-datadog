package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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

func testAccDatadogIntegrationAWSLogCollectionConfigAccessKey(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  access_key_id     = "%s"
  secret_access_key = "testacc-datadog-integration-secret"
}

resource "datadog_integration_aws_lambda_arn" "lambda" {
  account_id = datadog_integration_aws.account.access_key_id
  lambda_arn = "arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
  depends_on = [datadog_integration_aws.account]
}

resource "datadog_integration_aws_log_collection" "main" {
  account_id = datadog_integration_aws.account.access_key_id
  services = ["lambda"]
  depends_on = [datadog_integration_aws_lambda_arn.lambda]
}`, uniq)
}

func TestAccDatadogIntegrationAWSLogCollection(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAWSLogCollectionDestroy(accProvider),
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

func TestAccDatadogIntegrationAWSLogCollectionAccessKey(t *testing.T) {
	t.Parallel()
	if !isReplaying() {
		t.Skip("Account ID is not returned with invalid AWS accounts using access_key_id")
		return
	}
	ctx, accProviders := testAccProviders(context.Background(), t)
	accessKeyID := uniqueAWSAccessKeyID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAWSLogCollectionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLogCollectionConfigAccessKey(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLogCollectionExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_log_collection.main",
						"account_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_log_collection.main",
						"services.0", "lambda"),
				),
			},
		},
	})
}

func checkIntegrationAWSLogCollectionExists(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		return checkIntegrationAWSLogCollectionExistsHelper(auth, s, apiInstances)
	}
}

func checkIntegrationAWSLogCollectionExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	logCollections, _, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsIntegrations(ctx)
	if err != nil {
		return err
	}
	for resourceType, r := range s.RootModule().Resources {
		if strings.Contains(resourceType, "datadog_integration_aws_log_collection") {
			accountID := r.Primary.Attributes["account_id"]
			for _, logCollection := range logCollections {
				if v, ok := logCollection.GetAccountIdOk(); ok && *v == accountID && len(logCollection.GetServices()) > 0 {
					return nil
				}
			}
			return fmt.Errorf("The AWS Log Collection is not configured for the account: accountID=%s", accountID)
		}
	}
	return fmt.Errorf("Unable to find AWS Log Collection in any account")
}

func checkIntegrationAWSLogCollectionDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		return checkIntegrationAWSLogCollectionDestroyHelper(auth, s, apiInstances)
	}
}

func checkIntegrationAWSLogCollectionDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	logCollections, _, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsIntegrations(ctx)
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountID := r.Primary.Attributes["account_id"]
		err = utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Primary.ID != "" {
					for _, logCollection := range logCollections {
						if v, ok := logCollection.GetAccountIdOk(); ok && *v == accountID && len(logCollection.GetServices()) > 0 {
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
