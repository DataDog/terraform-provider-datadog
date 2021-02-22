package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccDatadogIntegrationAWSConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  	account_id                       = "%s"
  	role_name                        = "testacc-datadog-integration-role"
}`, uniq)
}

func testAccDatadogIntegrationAWSUpdateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  	account_id                       = "%s"
  	role_name                        = "testacc-datadog-integration-role"
	filter_tags                      = ["key:value"]
  	host_tags                        = ["key:value", "key2:value2"]
  	account_specific_namespace_rules = {
    	    auto_scaling = false
    	    opsworks = true
  	}
  	excluded_regions                 = ["us-east-1", "us-west-2"]
}`, uniq)
}

func TestAccDatadogIntegrationAWS(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"role_name", "testacc-datadog-integration-role"),
				),
			}, {
				Config: testAccDatadogIntegrationAWSUpdateConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"role_name", "testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"filter_tags.0", "key:value"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"host_tags.0", "key:value"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"host_tags.1", "key2:value2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_specific_namespace_rules.auto_scaling", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_specific_namespace_rules.opsworks", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"excluded_regions.0", "us-east-1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"excluded_regions.1", "us-west-2"),
				),
			},
		},
	})
}

func checkIntegrationAWSExists(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSExistsHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAWSExistsHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	integrations, _, err := datadogClientV1.AWSIntegrationApi.ListAWSAccounts(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		for _, account := range integrations.GetAccounts() {
			if account.GetAccountId() == accountId {
				return nil
			}
		}
		return fmt.Errorf("The AWS integration does not exists for account: accountId=%s", accountId)
	}
	return nil
}

func checkIntegrationAWSDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return checkIntegrationAWSDestroyHelper(authV1, s, datadogClientV1)
	}
}

func checkIntegrationAWSDestroyHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	integrations, _, err := datadogClientV1.AWSIntegrationApi.ListAWSAccounts(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountId := r.Primary.Attributes["account_id"]
		for _, account := range integrations.GetAccounts() {
			if account.GetAccountId() == accountId {
				return fmt.Errorf("The AWS integration still exists for account: accountId=%s", accountId)
			}
		}
	}
	return nil
}
