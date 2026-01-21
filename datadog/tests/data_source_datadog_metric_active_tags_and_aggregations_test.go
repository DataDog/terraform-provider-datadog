package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricActiveTagsAndAggregationsDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metric := "foo.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMetricActiveTagsAndAggregationsConfig(metric),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metric_active_tags_and_aggregations.foo", "metric", metric),
					resource.TestCheckResourceAttr("data.datadog_metric_active_tags_and_aggregations.foo", "active_tags.0", "environment"),
					resource.TestCheckResourceAttr("data.datadog_metric_active_tags_and_aggregations.foo", "active_tags.1", "version"),
					resource.TestCheckResourceAttr("data.datadog_metric_active_tags_and_aggregations.foo", "active_aggregations.0.space", "avg"),
					resource.TestCheckResourceAttr("data.datadog_metric_active_tags_and_aggregations.foo", "active_aggregations.0.time", "sum"),
				),
			},
		},
	})
}

func testAccDatasourceMetricActiveTagsAndAggregationsConfig(metric string) string {
	return fmt.Sprintf(`
data "datadog_metric_active_tags_and_aggregations" "foo" {
	metric = "%s"
}
`, metric)
}
