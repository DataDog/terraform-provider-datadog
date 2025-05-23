package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogLogsIndexesOrderDatasource(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniq := strings.ToLower(strings.ReplaceAll(uniqueEntityName(ctx, t), "_", "-"))

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatadogLogsIndexesOrderWithIndexConfig(uniq),
				// given the persistent nature of log indexes we can't check for exact value
				Check: resource.TestMatchResourceAttr("data.datadog_logs_indexes_order.order", "index_names.#", regexp.MustCompile("[^0].*$")),
			},
			{
				Config: testAccDatasourceDatadogLogsIndexesOrderConfig(),
				// given the persistent nature of log indexes we can't check for exact value
				// we can assume that at least one index already exists given the previous test
				Check: resource.TestMatchResourceAttr("data.datadog_logs_indexes_order.order", "index_names.#", regexp.MustCompile("[^0].*$")),
			},
		},
	})
}

func testAccDatasourceDatadogLogsIndexesOrderConfig() string {
	return `
data "datadog_logs_indexes_order" "order" {}
`
}

func testAccDatasourceDatadogLogsIndexesOrderWithIndexConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_logs_indexes_order" "order" {
	depends_on = [
		datadog_logs_index.sample_index,
	]
}

resource "datadog_logs_index" "sample_index" {
	name           = "%s-sample"
	filter {
		query = "project:sample"
	}
}
`, uniq)
}
