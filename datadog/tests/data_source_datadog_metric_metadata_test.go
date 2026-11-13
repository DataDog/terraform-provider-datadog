package test

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricMetadataDatasource(t *testing.T) {
	t.Parallel()
	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metric := strings.ReplaceAll(uniqueEntityName(ctx, t), "-", "_")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: submitTestMetricPreConfig(t, metric, "gauge"),
				Config:    testAccDatasourceMetricMetadataConfig(metric),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "metric_name", metric),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "short_name", "short name for metric_metadata foo"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "description", "some description"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "unit", "byte"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "per_unit", "second"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "statsd_interval", "1"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "type", "gauge"),
					resource.TestCheckResourceAttr("data.datadog_metric_metadata.foo", "integration", ""),
				),
			},
		},
	})
}

func testAccDatasourceMetricMetadataConfig(metric string) string {
	return testAccCheckDatadogMetricMetadataConfig(metric, "some description") + `
data "datadog_metric_metadata" "foo" {
	metric_name = datadog_metric_metadata.foo.metric
}
`
}
