package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func testAccDatadogIntegrationAWSConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  	account_id                       = "%s"
	role_name                        = "1234@testacc-datadog-integration-role"
}`, uniq)
}

func testAccDatadogIntegrationAWSUpdateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account" {
  	account_id                       = "%s"
	role_name                        = "1234@testacc-datadog-integration-role"
	filter_tags                      = ["key:value"]
  	host_tags                        = ["key:value", "key2:value2"]
  	account_specific_namespace_rules = {
    	    auto_scaling = false
    	    opsworks = true
  	}
  	excluded_regions                 = ["us-east-1", "us-west-2"]
	metrics_collection_enabled       = false
	resource_collection_enabled      = true
	cspm_resource_collection_enabled = true
}`, uniq)
}

func TestAccDatadogIntegrationAWS(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accountID := uniqueAWSAccountID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAWSDestroy(accProvider),
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
						"role_name", "1234@testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"metrics_collection_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"resource_collection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"cspm_resource_collection_enabled", "false"),
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
						"role_name", "1234@testacc-datadog-integration-role"),
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
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"metrics_collection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"resource_collection_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"cspm_resource_collection_enabled", "true"),
				),
			},
			{
				Config: testAccDatadogIntegrationAWSConfig(accountID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_id", accountID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"role_name", "1234@testacc-datadog-integration-role"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"excluded_regions.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"filter_tags.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"host_tags.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"account_specific_namespace_rules.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"metrics_collection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"resource_collection_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account",
						"cspm_resource_collection_enabled", "true"),
				),
			},
		},
	})
}

func testAccDatadogIntegrationAWSAccessKeyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account_access_key" {
  	access_key_id                       = "%s"
  	secret_access_key                   = "testacc-datadog-integration-secret"
}`, uniq)
}

func testAccDatadogIntegrationAWSAccessKeyUpdateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_aws" "account_access_key" {
  	access_key_id                    = "%s"
  	secret_access_key                = "testacc-datadog-integration-secret"
	filter_tags                      = ["key:value"]
  	host_tags                        = ["key:value", "key2:value2"]
  	account_specific_namespace_rules = {
    	    auto_scaling = false
    	    opsworks = true
  	}
  	excluded_regions                 = ["us-east-1", "us-west-2"]
}`, uniq)
}

func TestAccDatadogIntegrationAWSAccessKey(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accessKeyID := uniqueAWSAccessKeyID(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAWSDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatadogIntegrationAWSAccessKeyConfig(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"access_key_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"secret_access_key", "testacc-datadog-integration-secret"),
				),
			}, {
				Config: testAccDatadogIntegrationAWSAccessKeyUpdateConfig(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"access_key_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"secret_access_key", "testacc-datadog-integration-secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"filter_tags.0", "key:value"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"host_tags.0", "key:value"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"host_tags.1", "key2:value2"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"account_specific_namespace_rules.auto_scaling", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"account_specific_namespace_rules.opsworks", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"excluded_regions.0", "us-east-1"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"excluded_regions.1", "us-west-2"),
				),
			},
			{
				Config: testAccDatadogIntegrationAWSAccessKeyConfig(accessKeyID),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAWSExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"access_key_id", accessKeyID),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"secret_access_key", "testacc-datadog-integration-secret"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"excluded_regions.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"filter_tags.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"host_tags.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_integration_aws.account_access_key",
						"account_specific_namespace_rules.#", "0"),
				),
			},
		},
	})
}

func checkIntegrationAWSExists(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		return checkIntegrationAWSExistsHelper(auth, s, apiInstances)
	}
}

func checkIntegrationAWSExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	integrations, _, err := apiInstances.GetAWSIntegrationApiV1().ListAWSAccounts(ctx)
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountID := r.Primary.Attributes["account_id"]
		accessKeyID := r.Primary.Attributes["access_key_id"]
		for _, account := range integrations.GetAccounts() {
			if (accountID != "" && account.GetAccountId() == accountID) ||
				(accessKeyID != "" && account.GetAccessKeyId() == accessKeyID) {
				return nil
			}
		}
		return fmt.Errorf("The AWS integration does not exists for account: accountID=%s", accountID)
	}
	return nil
}

func checkIntegrationAWSDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		return checkIntegrationAWSDestroyHelper(auth, s, apiInstances)
	}
}

func checkIntegrationAWSDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	integrations, _, err := apiInstances.GetAWSIntegrationApiV1().ListAWSAccounts(ctx)
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		accountID := r.Primary.Attributes["account_id"]
		accessKeyID := r.Primary.Attributes["access_key_id"]
		for _, account := range integrations.GetAccounts() {
			if (accountID != "" && account.GetAccountId() == accountID) ||
				(accessKeyID != "" && account.GetAccessKeyId() == accessKeyID) {
				return fmt.Errorf("The AWS integration still exists for account: accountID=%s", accountID)
			}
		}
	}
	return nil
}
