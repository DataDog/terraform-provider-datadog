package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricTagConfigurationDatasource_Exists(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metric := "foo.bar"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMetricTagConfigurationConfig(metric),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "metric_name", metric),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "exists", "true"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "metric_type", "gauge"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "tags.0", "environment"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "tags.1", "version"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "include_percentiles", "false"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "exclude_tags_mode", "false"),
				),
			},
		},
	})
}

func TestAccDatadogMetricTagConfigurationDatasource_DoesNotExist(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metric := "foo.missing"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMetricTagConfigurationConfig(metric),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "metric_name", metric),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "exists", "false"),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "metric_type", ""),
					resource.TestCheckResourceAttr("data.datadog_metric_tag_configuration.foo", "tags.#", "0"),
					resource.TestCheckNoResourceAttr("data.datadog_metric_tag_configuration.foo", "include_percentiles"),
					resource.TestCheckNoResourceAttr("data.datadog_metric_tag_configuration.foo", "exclude_tags_mode"),
				),
			},
		},
	})
}

func testAccDatasourceMetricTagConfigurationConfig(metric string) string {
	return fmt.Sprintf(`
data "datadog_metric_tag_configuration" "foo" {
	metric_name = "%s"
}
`, metric)
}
