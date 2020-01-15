package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	datadog "github.com/zorkian/go-datadog-api"
)

const testAccCheckDatadogIntegrationAzureConfig = `
resource "datadog_integration_azure" "an_azure_integration" {
  tenant_name   = "testc44-1234-5678-9101-cc00736ftest"
  client_id     = "testc7f6-1234-5678-9101-3fcbf464test"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
  host_filters  = "foo:bar,buzz:lightyear"
}
`

func TestAccDatadogIntegrationAzure(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkIntegrationAzureDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAzureConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAzureExists,
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"tenant_name", "testc44-1234-5678-9101-cc00736ftest"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"client_id", "testc7f6-1234-5678-9101-3fcbf464test"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"client_secret", "testingx./Sw*g/Y33t..R1cH+hScMDt"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"host_filters", "foo:bar,buzz:lightyear"),
				),
			},
		},
	},
	)
}

func checkIntegrationAzureExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	integrations, err := client.ListIntegrationAzure()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		tenantName := r.Primary.ID
		for _, integration := range integrations {
			if integration.GetTenantName() == tenantName {
				return nil
			}
		}
		return fmt.Errorf("The Azure integration doesn't exist: tenantName=%s", tenantName)
	}
	return nil
}

func checkIntegrationAzureDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	integrations, err := client.ListIntegrationAzure()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		tenantName := r.Primary.ID
		for _, integration := range integrations {
			if integration.GetTenantName() == tenantName {
				return fmt.Errorf("The Azure integration still exist: tenantName=%s", tenantName)
			}
		}
	}
	return nil
}
