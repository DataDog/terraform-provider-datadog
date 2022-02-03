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

func TestAccDatadogLogsIndexesDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceLogsIndexesConfig(),
				Check: resource.ComposeTestCheckFunc(
					dataLogsIndexesCountCheck(accProvider),
				),
			},
		},
	})
}

func testAccDatasourceLogsIndexesConfig() string {
	return fmt.Sprintf(`
data "datadog_logs_indexes" "foo" {
}`)
}

func dataLogsIndexesCountCheck(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		authV1 := providerConf.AuthV1
		client := providerConf.DatadogClientV1

		logsIndexes, _, err := client.LogsIndexesApi.ListLogIndexes(authV1)
		if err != nil {
			return err
		}
		if err := utils.CheckForUnparsed(logsIndexes); err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_logs_indexes.foo"].Primary.Attributes
		logIndexesCount, _ := strconv.Atoi(resourceAttributes["logs_indexes.#"])

		if logIndexesCount != len(logsIndexes.GetIndexes()) {
			return fmt.Errorf("expected %d agent rules got %d agent rules",
				logIndexesCount, len(logsIndexes.GetIndexes()))
		}

		return nil
	}
}
