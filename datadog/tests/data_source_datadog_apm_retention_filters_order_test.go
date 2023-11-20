package test

import (
	"context"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogApmRetentionFiltersOrderDatasource(t *testing.T) {
	t.Parallel()

	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatadogApmRetentionFiltersOrderConfig(),
				// by default there is two default APM retention filters
				Check: resource.TestMatchResourceAttr("data.datadog_apm_retention_filters_order.order", "filter_ids.#", regexp.MustCompile("[^0].*$")),
			},
		},
	})
}

func testAccDatasourceDatadogApmRetentionFiltersOrderConfig() string {
	return `
data "datadog_apm_retention_filters_order" "order" {}
`
}
