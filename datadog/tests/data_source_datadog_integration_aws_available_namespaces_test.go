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

func TestAccDatadogIntegrationAWSAvailableNamespacesDatasource(t *testing.T) {
	_, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		PreCheck:                 func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceIntegrationAWSAvailableNamespacesConfig(),
				Check: resource.ComposeTestCheckFunc(
					checkDatadogIntegrationAWSAvailableNamespacesCount(providers.frameworkProvider),
				),
			},
		},
	})
}

func testAccDatasourceIntegrationAWSAvailableNamespacesConfig() string {
	return `data "datadog_integration_aws_available_namespaces" "foo" {}`
}

func checkDatadogIntegrationAWSAvailableNamespacesCount(accProvider *fwprovider.FrameworkProvider) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		awsNamespaces, _, err := apiInstances.GetAWSIntegrationApiV2().ListAWSNamespaces(auth)
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_integration_aws_available_namespaces.foo"].Primary.Attributes
		awsNamespacesCount, _ := strconv.Atoi(resourceAttributes["aws_namespaces.#"])

		namespacesDd := awsNamespaces.Data.Attributes.Namespaces
		if awsNamespacesCount != len(namespacesDd) {
			return fmt.Errorf("expected %d aws namespaces, got %d aws namespaces",
				awsNamespacesCount, len(namespacesDd))
		}

		return nil
	}
}
