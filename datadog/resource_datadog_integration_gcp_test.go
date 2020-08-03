package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccCheckDatadogIntegrationGCPConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_gcp" "awesome_gcp_project_integration" {
  project_id     = "%s"
  private_key_id = "1234567890123456789012345678901234567890"
  private_key    = "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n"
  client_email   = "%s@awesome-project-id.iam.gserviceaccount.com"
  client_id      = "123456789012345678901"
  host_filters   = "foo:bar,buzz:lightyear"
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
}`, uniq, uniq)
}

func TestAccDatadogIntegrationGCP(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	client := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationGCPDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationGCPConfig(client),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(accProvider),
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
				),
			},
			{
				Config: testAccCheckDatadogIntegrationGCPEmptyHostFiltersConfig(client),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationGCPExists(accProvider),
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
				),
			},
		},
	})
}

func checkIntegrationGCPExists(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		integrations, _, err := datadogClientV1.GCPIntegrationApi.ListGCPIntegration(authV1).Execute()
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

func checkIntegrationGCPDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		integrations, _, err := datadogClientV1.GCPIntegrationApi.ListGCPIntegration(authV1).Execute()
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
