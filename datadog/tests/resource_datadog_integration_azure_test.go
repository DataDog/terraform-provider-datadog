package test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccCheckDatadogIntegrationAzureConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_azure" "an_azure_integration" {
  tenant_name   = "%s"
  client_id     = "testc7f6-1234-5678-9101-3fcbf464test"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
  host_filters  = "foo:bar,buzz:lightyear"
}`, uniq)
}

func TestAccDatadogIntegrationAzure(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	tenantName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkIntegrationAzureDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogIntegrationAzureConfig(tenantName),
				Check: resource.ComposeTestCheckFunc(
					checkIntegrationAzureExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"tenant_name", tenantName),
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

func checkIntegrationAzureExistsHelper(authV1 context.Context, s *terraform.State, client *datadogV1.APIClient) error {
	integrations, _, err := client.AzureIntegrationApi.ListAzureIntegration(authV1).Execute()
	if err != nil {
		return err
	}
	for _, r := range s.RootModule().Resources {
		tenantName, _, err := utils.TenantAndClientFromID(r.Primary.ID)
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
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := checkIntegrationAzureExistsHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func checkIntegrationAzureDestroyHelper(authV1 context.Context, s *terraform.State, client *datadogV1.APIClient) error {
	integrations, _, err := client.AzureIntegrationApi.ListAzureIntegration(authV1).Execute()
	if err != nil && !strings.Contains(string(err.(datadogV1.GenericOpenAPIError).Body()), "Azure Integration not yet installed.") {
		return fmt.Errorf("Error listing Azure Accounts: Response %s: %v", err.(datadogV1.GenericOpenAPIError).Body(), err)
	}
	for _, r := range s.RootModule().Resources {
		if r.Type == "datadog_integration_azure" {
			tenantName, _, err := utils.TenantAndClientFromID(r.Primary.ID)
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
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := checkIntegrationAzureDestroyHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}
