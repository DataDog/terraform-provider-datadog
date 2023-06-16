package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
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

		resourceAttributes := state.RootModule().Resources["data.datadog_logs_pipelines.foo"].Primary.Attributes
		logPipelinesCount, _ := strconv.Atoi(resourceAttributes["logs_pipelines.#"])

		if logPipelinesCount != len(logsPipelines) {
			return fmt.Errorf("expected %d pipelines got %d pipelines",
				logPipelinesCount, len(logsPipelines))
		}

		return nil
	}
}

func TestAccDatadogLogsPipelinesDatasourceReadonly(t *testing.T) {
	t.Parallel()
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceLogsPipelinesReadonlyConfig(),
				Check: resource.ComposeTestCheckFunc(
					dataLogsPipelinesReadonlyCountCheck(accProvider),
				),
			},
		},
	})
}

func testAccDatasourceLogsPipelinesReadonlyConfig() string {
	return `
data "datadog_logs_pipelines" "foo" {
	is_read_only = true
}`
}

func dataLogsPipelinesReadonlyCountCheck(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		logsPipelines, _, err := apiInstances.GetLogsPipelinesApiV1().ListLogsPipelines(auth)
		var readOnlyPipelines []datadogV1.LogsPipeline
		for _, pipeline := range logsPipelines {
			if pipeline.GetIsReadOnly() {
				readOnlyPipelines = append(readOnlyPipelines, pipeline)
			}
		}
		if err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_logs_pipelines.foo"].Primary.Attributes
		logPipelinesCount, _ := strconv.Atoi(resourceAttributes["logs_pipelines.#"])

		if logPipelinesCount != len(readOnlyPipelines) {
			return fmt.Errorf("expected %d pipelines got %d pipelines",
				logPipelinesCount, len(readOnlyPipelines))
		}

		return nil
	}
}
