package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricsDatasource(t *testing.T) {
	if !isReplaying() {
		t.Skip("This test only supports replaying - requires specific metrics in test account")
	}
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	query := "foo."

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
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
