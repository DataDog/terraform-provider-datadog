package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

const testAccCheckDatadogIntegrationGCPConfig = `
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "super-awesome-project-id"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  host_filters   = "foo:bar,buzz:lightyear"
}
`
const testAccCheckDatadogIntegrationGCPEmptyHostFiltersConfig = `
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "super-awesome-project-id"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
}
`

func TestAccDatadogIntegrationGCP(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationGCPDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationGCPConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"project_id", "super-awesome-project-id"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key_id", "1234567890123456789012345678901234567890"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key", "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_email", "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_id", "123456789012345678901"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"host_filters", "foo:bar,buzz:lightyear"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationGCPEmptyHostFiltersConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"project_id", "super-awesome-project-id"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key_id", "1234567890123456789012345678901234567890"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"private_key", "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_email", "awesome-service-account@awesome-project-id.iam.gserviceaccount.com"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"client_id", "123456789012345678901"),
					resource.TestCheckResourceAttr(
						"datadog_integration_gcp.awesome_gcp_project_integration",
						"host_filters", ""),
				),
			},
		},
	})
}

func checkIntegrationGCPExists(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)
		integrations, err := client.ListIntegrationGCP()
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			projectID := r.Primary.ID
			for _, integration := range integrations {
				if integration.GetProjectID() == projectID {
					return nil
				}
			}
			return fmt.Errorf("The Google Cloud Platform integration doesn't exist: projectID=%s", projectID)
		}
		return nil
	}
}

func checkIntegrationGCPDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		client := accProvider.Meta().(*datadog.Client)
		integrations, err := client.ListIntegrationGCP()
		if err != nil {
			return err
		}
		for _, r := range s.RootModule().Resources {
			projectID := r.Primary.ID
			for _, integration := range integrations {
				if integration.GetProjectID() == projectID {
					return fmt.Errorf("The Google Cloud Platform integration still exist: projectID=%s", projectID)
				}
			}
		}
		return nil
	}
}
