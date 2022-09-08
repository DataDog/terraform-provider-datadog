package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogLogsPipelinesDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceLogsPipelinesConfig(),
				Check: resource.ComposeTestCheckFunc(
					dataLogsPipelinesCountCheck(accProvider),
				),
			},
		},
	})
}

func testAccDatasourceLogsPipelinesConfig() string {
	return `
data "datadog_logs_pipelines" "foo" {
}`
}

func dataLogsPipelinesCountCheck(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		logsPipelines, _, err := apiInstances.GetLogsPipelinesApiV1().ListLogsPipelines(auth)
		if err != nil {
			return err
		}
		if err := utils.CheckForUnparsed(logsPipelines); err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_logs_pipelines.foo"].Primary.Attributes
		logPipelinesCount, _ := strconv.Atoi(resourceAttributes["logs_pipelines.#"])

		if logPipelinesCount != len(logsPipelines) {
			return fmt.Errorf("expected %d pipelines got %d pipelines",
				logPipelinesCount, len(logsPipelines))
		}

		return nil
	}
}
