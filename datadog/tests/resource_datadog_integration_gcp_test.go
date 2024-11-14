package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

func testAccCheckDatadogIntegrationGCPConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id                   = "%s"
  private_key_id               = "1234567890123456789012345678901234567890"
  private_key                  = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email                 = "%s@awesome-project-id.iam.gserviceaccount.com"
  client_id                    = "123456789012345678901"
  host_filters                 = "foo:bar,buzz:lightyear"
  cloud_run_revision_filters   = ["tag:one", "tag:two"]
}`, uniq, uniq)
}

func testAccCheckDatadogIntegrationGCPEmptyHostFiltersConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "%s"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "%s@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  cloud_run_revision_filters   = ["tag:one", "tag:two"]
}`, uniq, uniq)
}

func testAccCheckDatadogIntegrationGCPUpdatePrivateKeyConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "%s"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n key updated \n-----END PRIVATE KEY-----\n"
  client_email   = "%s@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  automute       = true
}`, uniq, uniq)
}

func TestAccDatadogIntegrationGCP(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	client := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             checkIntegrationGCPDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationGCPConfig(client),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"project_id", client),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key_id", "1234567890123456789012345678901234567890"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key", "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_email", fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", client)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_id", "123456789012345678901"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"host_filters", "foo:bar,buzz:lightyear"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration", "cloud_run_revision_filters.*", "tag:two"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration", "cloud_run_revision_filters.*", "tag:one"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"automute", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"is_security_command_center_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"is_resource_change_collection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"cspm_resource_collection_enabled", "false"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationGCPEmptyHostFiltersConfig(client),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"project_id", client),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key_id", "1234567890123456789012345678901234567890"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key", "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_email", fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", client)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_id", "123456789012345678901"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"host_filters", ""),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration", "cloud_run_revision_filters.*", "tag:two"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration", "cloud_run_revision_filters.*", "tag:one"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"automute", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"is_security_command_center_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"is_resource_change_collection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"cspm_resource_collection_enabled", "false"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationGCPUpdatePrivateKeyConfig(client),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"project_id", client),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key_id", "1234567890123456789012345678901234567890"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key", "-----BEGIN PRIVATE KEY-----\n key updated \n-----END PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_email", fmt.Sprintf("%s@awesome-project-id.iam.gserviceaccount.com", client)),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_id", "123456789012345678901"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"host_filters", ""),
					resource.TestCheckNoResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration", "cloud_run_revision_filters"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"automute", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"is_security_command_center_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"is_resource_change_collection_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"cspm_resource_collection_enabled", "false"),
				),
			},
		},
	})
}

func checkIntegrationGCPExists(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		integrations, _, err := apiInstances.GetGCPIntegrationApiV1().ListGCPIntegration(auth)
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			projectID := r.Primary.ID
			for _, integration := range integrations {
				if integration.GetProjectId() == projectID {
					return nil
				}
			}
			return fmt.Errorf("the Google Cloud Platform integration doesn't exist: projectID=%s", projectID)
		}
		return nil
	}
}

func checkIntegrationGCPDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		integrations, _, err := apiInstances.GetGCPIntegrationApiV1().ListGCPIntegration(auth)
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			projectID := r.Primary.ID
			for _, integration := range integrations {
				if integration.GetProjectId() == projectID {
					return fmt.Errorf("the Google Cloud Platform integration still exist: projectID=%s", projectID)
				}
			}
		}
		return nil
	}
}
