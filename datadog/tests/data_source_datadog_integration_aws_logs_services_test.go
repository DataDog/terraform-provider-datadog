package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAWSLogsServicesDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSLogsServicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSLogsServicesCount(accProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSLogsServicesConfig() string {
	return `
data "datadog_integration_aws_logs_services" "foo" {
}`
}

func checkDatadogIntegrationAWSLogsServicesCount(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		awsLogsServices, _, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsServices(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_logs_services.foo"].Primary.Attributes
		awsLogsServicesCount, _ := strconv.Atoi(resourceAttributes["aws_logs_services.#"])

		if awsLogsServicesCount != len(awsLogsServices) {
			return fmt.Errorf("expected %d aws logs services, got %d aws logs services",
				awsLogsServicesCount, len(awsLogsServices))
		}

		return nil
	}
}
