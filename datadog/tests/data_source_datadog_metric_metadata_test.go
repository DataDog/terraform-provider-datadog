package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricMetadataDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metric := "foo"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMetricMetadataConfig(metric),
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
	return fmt.Sprintf(`
data "datadog_metric_metadata" "foo" {
	metric_name = "%s"
}
`, metric)
}
