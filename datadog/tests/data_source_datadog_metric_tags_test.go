package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogMetricTagsDatasource(t *testing.T) {
	t.Parallel()
	_, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	metric := "datadog.agent.running"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceMetricTagsConfig(metric),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.datadog_metric_tags.agent", "metric", metric),
					resource.TestCheckResourceAttr("data.datadog_metric_tags.agent", "tags.0", "foo:bar"),
				),
			},
		},
	})
}

func testAccDatasourceMetricTagsConfig(metric string) string {
	return fmt.Sprintf(`
data "datadog_metric_tags" "agent" {
	metric = "%s"
}
`, metric)
}
