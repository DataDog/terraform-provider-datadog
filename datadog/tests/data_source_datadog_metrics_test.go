package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricsDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	query := "foo."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMetricsConfig(query),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metrics.foo", "query", query),
					resource.TestCheckResourceAttr("data.datadog_metrics.foo", "metrics.0", "foo.bar"),
					resource.TestCheckResourceAttr("data.datadog_metrics.foo", "metrics.1", "foo.baz"),
				),
			},
		},
	})
}

func testAccDatasourceMetricsConfig(query string) string {
	return fmt.Sprintf(`
data "datadog_metrics" "foo" {
	query = "%s"
}
`, query)
}
