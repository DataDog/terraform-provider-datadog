package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccIntegrationGcpStsBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationGcpSts(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "automute", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "client_email", fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "is_cspm_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "is_security_command_center_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "resource_collection_enabled", "false"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "host_filters.*", "tag:one"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "host_filters.*", "tag:two"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "cloud_run_revision_filters.*", "tag:one"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "cloud_run_revision_filters.*", "tag:two"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "account_tags.*", "a:tag"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "account_tags.*", "another:one"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp_sts.foo", "account_tags.*", "and:another"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationGcpStsUpdated(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "automute", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "client_email", fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "is_cspm_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "is_security_command_center_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "resource_collection_enabled", "true"),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp_sts.foo", "host_filters"),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp_sts.foo", "cloud_run_revision_filters"),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp_sts.foo", "account_tags"),
				),
			},
		},
	})
}

func TestAccIntegrationGcpStsDefault(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogIntegrationGcpStsDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationGcpStsDefault(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogIntegrationGcpStsExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "automute", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "client_email", fmt.Sprintf("%s@test-project.iam.gserviceaccount.com", uniq)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp_sts.foo", "is_cspm_enabled", "false"),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp_sts.foo", "host_filters"),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp_sts.foo", "cloud_run_revision_filters"),
				),
			},
		},
	})
}

func testAccCheckDatadogIntegrationGcpSts(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp_sts" "foo" {
    automute = "false"
    client_email = "%s@test-project.iam.gserviceaccount.com"
    host_filters = ["tag:one", "tag:two"]
    cloud_run_revision_filters = ["tag:one", "tag:two"]
    is_cspm_enabled = "false"
    resource_collection_enabled = "false"
    is_security_command_center_enabled = "false"
    account_tags = ["a:tag", "another:one", "and:another"]
}`, uniq)
}

func testAccCheckDatadogIntegrationGcpStsUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp_sts" "foo" {
    automute = "true"
    client_email = "%s@test-project.iam.gserviceaccount.com"
    is_cspm_enabled = "true"
    resource_collection_enabled = "true"
    is_security_command_center_enabled = "true"
}`, uniq)
}

func testAccCheckDatadogIntegrationGcpStsDefault(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp_sts" "foo" {
    automute = "true"
    client_email = "%s@test-project.iam.gserviceaccount.com"
}`, uniq)
}

func testAccCheckDatadogIntegrationGcpStsDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
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
			if r.Type != "resource_datadog_integration_gcp_sts" {
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

func testAccCheckDatadogIntegrationGcpStsExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
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
		return utils.TranslateClientError(err, httpResp, "error retrieving Integration Gcp Sts")
	}

	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_integration_gcp_sts" {
			continue
		}

		for _, r := range s.RootModule().Resources {
			for _, account := range resp.GetData() {
				if account.GetId() == r.Primary.ID {
					return nil
				}
			}
			return utils.TranslateClientError(err, httpResp, "error retrieving Integration Gcp Sts")
		}
	}
	return nil
}
