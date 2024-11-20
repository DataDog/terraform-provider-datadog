package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"

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

func TestAccDatadogIntegrationAWSLogsServicesDatasourceV2(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSLogsServicesV2Config(),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSLogsServicesV2Count(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSLogsServicesV2Config() string {
	return `data "datadog_integration_aws_logs_services" "foo" {}`
}

func checkDatadogIntegrationAWSLogsServicesV2Count(accProvider *fwprovider.FrameworkProvider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		awsLogsServices, _, err := apiInstances.GetAWSLogsIntegrationApiV2().ListAWSLogsServices(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_logs_services.foo"].Primary.Attributes
		awsLogsServicesCount, _ := strconv.Atoi(resourceAttributes["aws_logs_services.#"])

		servicesDd := awsLogsServices.Data.Attributes.LogsServices
		if awsLogsServicesCount != len(servicesDd) {
			return fmt.Errorf("expected %d aws logs services, got %d aws logs services",
				awsLogsServicesCount, len(servicesDd))
		}

		return nil
	}
}
