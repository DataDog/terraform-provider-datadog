package test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
)

func logsArchiveOrderCheckCount(accProvider func() (*schema.Provider, error)) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		auth := providerConf.Auth
		apiInstances := providerConf.DatadogApiInstances

		logsArchiveOrder, _, err := apiInstances.GetLogsArchivesApiV2().GetLogsArchiveOrder(auth)
		if err != nil {
			return err
		}
		return logsArchiveOrderCount(state, len(logsArchiveOrder.Data.Attributes.ArchiveIds))
	}
}

func logsArchiveOrderCount(state *terraform.State, responseCount int) error {
	resourceAttributes := state.RootModule().Resources["data.datadog_logs_archives_order.order"].Primary.Attributes
	logsArchiveCount, _ := strconv.Atoi(resourceAttributes["archive_ids.#"])

	if logsArchiveCount != responseCount {
		return fmt.Errorf("expected %d logs archives in order, got %d logs archives",
			responseCount, logsArchiveCount)
	}
	return nil
}

func TestAccDatadogLogsArchivesOrderDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	uniq := uniqueAWSAccountID(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "datadog_logs_archives_order" "order" {}`,
				Check:  resource.TestMatchResourceAttr("data.datadog_logs_archives_order.order", "archive_ids.#", regexp.MustCompile("[^0].*$")),
			},
			{
				Config: testAccDatasourceDatadogLogsArchiveOrderWithArchive(uniq),
				Check: resource.ComposeTestCheckFunc(
					logsArchiveOrderCheckCount(accProvider),
				),
			},
		},
	})
}

func testAccDatasourceDatadogLogsArchiveOrderWithArchive(uniq string) string {
	return fmt.Sprintf(`
data "datadog_logs_archives_order" "order" {
	depends_on = ["datadog_logs_archive.sample_archive"]
}

resource "datadog_integration_aws" "account" {
	account_id         = "%[1]s"
	role_name          = "testacc-datadog-integration-role"
}

resource "datadog_logs_archive" "sample_archive" {
	name = "my first s3 archive"
	query = "service:tutu"
	s3_archive {
		bucket 		 = "my-bucket"
		path 		 = "/path/foo"
		account_id   = datadog_integration_aws.account.account_id
		role_name    = "testacc-datadog-integration-role"
	}
	rehydration_tags = ["team:intake", "team:app"]
	include_tags = true
	rehydration_max_scan_size_in_gb = 123
}
`, uniq)
}
