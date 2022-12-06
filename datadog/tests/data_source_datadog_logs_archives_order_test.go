package test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogLogsArchivesOrderDatasource(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
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
				Check:  resource.TestCheckTypeSetElemAttrPair("data.datadog_logs_archives_order.order", "archive_ids.*", "datadog_logs_archive.sample_archive", "id"),
			},
		},
	})
}

func testAccDatasourceDatadogLogsArchiveOrderWithArchive(uniq string) string {
	return fmt.Sprintf(`
data "datadog_logs_archives_order" "order" {
	depends_on = [
		datadog_logs_archive.sample_archive,
	]
}

resource "datadog_integration_aws" "account" {
	account_id         = "%[1]s"
	role_name          = "testacc-datadog-integration-role"
}

resource "datadog_logs_archive" "sample_archive" {
	depends_on = ["datadog_integration_aws.account"]
	name = "my first s3 archive"
	query = "service:tutu"
	s3_archive {
		bucket 		 = "my-bucket"
		path 		 = "/path/foo"
		account_id   = "%[1]s"
		role_name    = "testacc-datadog-integration-role"
	}
	rehydration_tags = ["team:intake", "team:app"]
	include_tags = true
	rehydration_max_scan_size_in_gb = 123
}
`, uniq)
}
