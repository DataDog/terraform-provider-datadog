package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogIntegrationAWSAvailableLogsServicesDatasource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSAvailableLogsServicesConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSAvailableLogsServicesCount(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSAvailableLogsServicesConfig() string {
	return `data "datadog_integration_aws_available_logs_services" "foo" {}`
}

func checkDatadogIntegrationAWSAvailableLogsServicesCount(accProvider *fwprovider.FrameworkProvider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		awsLogsServices, _, err := apiInstances.GetAWSLogsIntegrationApiV2().ListAWSLogsServices(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_available_logs_services.foo"].Primary.Attributes
		awsLogsServicesCount, _ := strconv.Atoi(resourceAttributes["aws_logs_services.#"])

		servicesDd := awsLogsServices.Data.Attributes.LogsServices
		if awsLogsServicesCount != len(servicesDd) {
			return fmt.Errorf("expected %d aws logs services, got %d aws logs services",
				awsLogsServicesCount, len(servicesDd))
		}

		return nil
	}
}
