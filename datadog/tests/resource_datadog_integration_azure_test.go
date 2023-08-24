package test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	dd "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func testAccCheckDatadogIntegrationAzureConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_azure" "an_azure_integration" {
  tenant_name   = "%[1]s"
  client_id     = "testc7f6-1234-5678-9101-3fcbf464test"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
  host_filters  = "foo:bar,buzz:lightyear"
}

resource "datadog_integration_azure" "an_azure_integration_two" {
  depends_on    = [datadog_integration_azure.an_azure_integration]
  tenant_name   = "%[1]s"
  client_id     = "testc7f6-1234-5678-9101-3fcbf123test"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
  host_filters  = "foo:bar,buzz:lightyear"
}`, uniq)
}

func testAccCheckDatadogIntegrationAzureConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_integration_azure" "an_azure_integration" {
  tenant_name   = "%s"
  client_id     = "testc7f6-1234-5678-9101-3fcbf464test"
  client_secret = "testingx./Sw*g/Y33t..R1cH+hScMDt"
  app_service_plan_filters = "bar:baz,stinky:pete"
  automute      = true
  cspm_enabled  = true
  custom_metrics_enabled = true
}`, uniq)
}

func TestAccDatadogIntegrationAzure(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	tenantName := fmt.Sprintf("aaaaaaaa-bbbb-cccc-dddd-%dee", clockFromContext(ctx).Now().Unix())
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      checkIntegrationAzureDestroy(accProvider),
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
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"app_service_plan_filters", ""),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"automute", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"cspm_enabled", "false"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"custom_metrics_enabled", "false"),
					resource.TestCheckResourceAttr("datadog_integration_azure.an_azure_integration_two",
						"tenant_name", tenantName),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration_two",
						"client_id", "testc7f6-1234-5678-9101-3fcbf123test"),
				),
			},
			{
				Config: testAccCheckDatadogIntegrationAzureConfigUpdated(tenantName),
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
						"host_filters", ""),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"app_service_plan_filters", "bar:baz,stinky:pete"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"automute", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"cspm_enabled", "true"),
					resource.TestCheckResourceAttr(
						"datadog_integration_azure.an_azure_integration",
						"custom_metrics_enabled", "true"),
				),
			},
		},
	},
	)
}

func checkIntegrationAzureExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	integrations, _, err := apiInstances.GetAzureIntegrationApiV1().ListAzureIntegration(ctx)
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

func checkIntegrationAzureExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := checkIntegrationAzureExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func checkIntegrationAzureDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	integrations, _, err := apiInstances.GetAzureIntegrationApiV1().ListAzureIntegration(ctx)
	if err != nil && !strings.Contains(string(err.(dd.GenericOpenAPIError).Body()), "Azure Integration not yet installed.") {
		return fmt.Errorf("Error listing Azure Accounts: Response %s: %v", err.(dd.GenericOpenAPIError).Body(), err)
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

func checkIntegrationAzureDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := checkIntegrationAzureDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}
