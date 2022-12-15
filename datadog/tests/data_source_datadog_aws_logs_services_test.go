package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogAwsLogsServicesDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceAwsLogsServicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					dataAwsLogsServicesCountCheck(accProvider),
				),
			},
		},
	})
}

func testAccDatasourceAwsLogsServicesConfig() string {
	return `
data "datadog_aws_logs_services" "foo" {
}`
}

func dataAwsLogsServicesCountCheck(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		awsLogsServices, _, err := apiInstances.GetAWSLogsIntegrationApiV1().ListAWSLogsServices(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_aws_logs_services.foo"].Primary.Attributes
		awsLogsServicesCount, _ := strconv.Atoi(resourceAttributes["aws_logs_services_ids.#"])

		if awsLogsServicesCount != len(awsLogsServices) {
			return fmt.Errorf("expected %d aws logs services got %d aws logs services",
				awsLogsServicesCount, len(awsLogsServices))
		}

		return nil
	}
}
