package datadog

import (
	"context"
	"fmt"
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

const testAccDatadogIntegrationAWSLogCollectionConfig = `
resource "datadog_integration_aws" "account" {
  account_id                       = "1234567890"
  role_name                        = "testacc-datadog-integration-role"
}

resource "datadog_integration_aws_log_collection" "main" {
  account_id = "1234567890"
  services = ["lambda"]
  depends_on = [datadog_integration_aws.account]
}
`

func TestAccDatadogIntegrationAWSLogCollection(t *testing.T) {
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationAWSLogCollectionDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSLogCollectionConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSLogCollectionExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws_log_collection.main",
						"account_id", "1234567890"),
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
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSLogCollectionExistsHelper(s, authV1, datadogClientV1)
	}
}

func checkIntegrationAWSLogCollectionExistsHelper(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return err
	}
	for resourceType, r := range s.RootModule().Resources {
		if strings.Contains(resourceType, "datadog_integration_aws_log_collection") {
			accountId := r.Primary.Attributes["account_id"]
			for _, logCollection := range logCollections {
				if *logCollection.AccountId == accountId && len(logCollection.GetServices()) > 0 {
					return nil
				}
			}
			return fmt.Errorf("The AWS Log Collection is not configured for the account: accountId=%s", accountId)
		}
	}
	return fmt.Errorf("Unable to find AWS Log Collection in any account")
}

func checkIntegrationAWSLogCollectionDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSLogCollectionDestroyHelper(s, authV1, datadogClientV1)
	}
}

func checkIntegrationAWSLogCollectionDestroyHelper(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV1.APIClient) error {
	logCollections, _, err := datadogClientV1.AWSLogsIntegrationApi.ListAWSLogsIntegrations(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		for _, logCollection := range logCollections {
			if *logCollection.AccountId == accountId && len(logCollection.GetServices()) > 0 {
				return fmt.Errorf("The AWS Log Collection is still enabled for the account: accountId=%s", accountId)
			}
		}
	}
	return nil
}
