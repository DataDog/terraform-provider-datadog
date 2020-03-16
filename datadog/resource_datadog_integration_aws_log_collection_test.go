package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
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
	accProviders, cleanup := testAccProviders(t)
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
		client := accProvider.Meta().(*datadog.Client)
		logCollections, err := client.GetIntegrationAWSLogCollection()
		if err != nil {
			return err
		}
		for resourceType, r := range s.RootModule().Resources {
			if strings.Contains(resourceType, "datadog_integration_aws_log_collection") {
				accountId := r.Primary.Attributes["account_id"]
				for _, logCollection := range *logCollections {
					if *logCollection.AccountID == accountId && len(logCollection.Services) > 0 {
						return nil
					}
				}
				return fmt.Errorf("The AWS Log Collection is not configured for the account: accountId=%s", accountId)
			}
		}
		return fmt.Errorf("Unable to find AWS Log Collection in any account")
	}
}

func checkIntegrationAWSLogCollectionDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)
		logCollections, err := client.GetIntegrationAWSLogCollection()
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			accountId := r.Primary.Attributes["account_id"]
			for _, logCollection := range *logCollections {
				if *logCollection.AccountID == accountId && len(logCollection.Services) > 0 {
					return fmt.Errorf("The AWS Log Collection is still enabled for the account: accountId=%s", accountId)
				}
			}
		}
		return nil
	}
}
