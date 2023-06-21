package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogLogsIndexesDatasource(t *testing.T) {
	t.Parallel()
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
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		logsIndexes, _, err := apiInstances.GetLogsIndexesApiV1().ListLogIndexes(auth)
		if err != nil {
			return err
		}
		if err := utils.CheckForUnparsed(logsIndexes); err != nil {
			return err
		}

		resourceAttributes := state.RootModule().Resources["data.datadog_logs_indexes.foo"].Primary.Attributes
		logIndexesCount, _ := strconv.Atoi(resourceAttributes["logs_indexes.#"])

		if logIndexesCount != len(logsIndexes.GetIndexes()) {
			return fmt.Errorf("expected %d indexes got %d indexes",
				logIndexesCount, len(logsIndexes.GetIndexes()))
		}

		return nil
	}
}
