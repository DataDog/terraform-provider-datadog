package datadog

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationAzureDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAzureConfig,
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAzureExists(accProvider),
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

func checkIntegrationAzureExistsHelper(s *terraform.State, authV1 context.Context, client *datadogV1.APIClient) error {
	integrations, _, err := client.AzureIntegrationApi.ListAzureIntegration(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		tenantName, _, err := tenantAndClientFromID(r.Primary.ID)
		if err != nil {
			return err
		}
		for _, integration := range integrations {
			if integration.GetTenantName() == tenantName {
				return nil
			}
		}
		return fmt.Errorf("The Azure integration doesn't exist: tenantName=%s", tenantName)
	}
	return nil
}

func checkIntegrationAzureExists(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := checkIntegrationAzureExistsHelper(s, authV1, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func checkIntegrationAzureDestroyHelper(s *terraform.State, authV1 context.Context, client *datadogV1.APIClient) error {
	integrations, _, err := client.AzureIntegrationApi.ListAzureIntegration(authV1).Execute()
	if err != nil && !strings.Contains(string(err.(datadog.GenericOpenAPIError).Body()), "Azure Integration not yet installed.") {
		return fmt.Errorf("Error listing Azure Accounts: Response %s: %v", err.(datadog.GenericOpenAPIError).Body(), err)
	}
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_integration_azure" {
			tenantName, _, err := tenantAndClientFromID(r.Primary.ID)
			if err != nil {
				return err
			}
			for _, integration := range integrations {
				if integration.GetTenantName() == tenantName {
					return fmt.Errorf("The Azure integration still exist: tenantName=%s", tenantName)
				}
			}
		}
	}
	return nil
}

func checkIntegrationAzureDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := checkIntegrationAzureDestroyHelper(s, authV1, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}
