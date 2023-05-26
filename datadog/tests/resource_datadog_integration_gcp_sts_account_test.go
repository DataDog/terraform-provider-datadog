package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccIntegrationGcpStsAccountBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationGcpStsAccountDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationGcpStsAccount(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationGcpStsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "automute", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "client_email", fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "is_cspm_enabled", "false"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts_account.foo", "host_filters.*", "tag:one"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts_account.foo", "host_filters.*", "tag:two"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationGcpStsAccountUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationGcpStsAccountExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "automute", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "client_email", fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "is_cspm_enabled", "true"),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp_sts_account.foo", "host_filters"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationGcpStsAccount(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_gcp_sts_account" "foo" {
    automute = "false"
    client_email = "%s@test-project.iam.gserviceaccount.com"
    host_filters = ["tag:one", "tag:two"]
    is_cspm_enabled = "false"
}`, uniq)
}

func testAccCheckDatadogIntegrationGcpStsAccountUpdated(uniq string) string {
	// Update me to make use of the unique value
	return fmt.Sprintf(`
resource "datadog_integration_gcp_sts_account" "foo" {
    automute = "true"
    client_email = "%s@test-project.iam.gserviceaccount.com"
    is_cspm_enabled = "true"
}`, uniq)
}

func testAccCheckDatadogIntegrationGcpStsAccountDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := IntegrationGcpStsAccountDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func IntegrationGcpStsAccountDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_integration_gcp_sts_account" {
				continue
			}

			resp, httpResp, err := apiInstances.GetGCPIntegrationApiV2().ListGCPSTSAccounts(auth)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving IntegrationGcpStsAccount %s", err)}
			}

			for _, account := range resp.GetData() {
				if account.Id == &r.Primary.ID {
					return &utils.RetryableError{Prob: "IntegrationGcpStsAccount still exists"}
				}
			}

			return nil
		}
		return nil
	})
	return err
}

func testAccCheckDatadogIntegrationGcpStsAccountExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := integrationGcpStsAccountExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func integrationGcpStsAccountExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	resp, httpResp, err := apiInstances.GetGCPIntegrationApiV2().ListGCPSTSAccounts(auth)
	if err != nil {
		return utils.TranslateClientError(err, httpResp, "error retrieving IntegrationGcpStsAccount")
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_gcp_sts_account" {
			continue
		}

		for _, r := range s.RootModule().Resources {
			for _, account := range resp.GetData() {
				if account.GetId() == r.Primary.ID {
					return nil
				}
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving IntegrationGcpStsAccount")
		}
	}
	return nil
}
