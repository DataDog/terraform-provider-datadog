package test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDatadogLogsArchivesOrderDatasource(t *testing.T) {
	_, accProviders := testAccProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: `data "datadog_logs_Archives_order" "order" {}`,
				Check:  resource.TestMatchResourceAttr("data.datadog_logs_archives_order.order", "archive_ids.#", regexp.MustCompile("[^0].*$")),
			},
		},
	})
}
