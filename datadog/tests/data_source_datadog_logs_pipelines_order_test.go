package test

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccDatadogLogsPipelinesOrderDatasource(t *testing.T) {
	t.Parallel()

	ctx, _, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := strings.ToLower(uniqueEntityName(ctx, t))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatadogLogsPipelinesOrderWithPipelineConfig(uniq),
				// given the persistent nature of log pipelines we can't check for exact value
				Check: resource.TestMatchResourceAttr("data.datadog_logs_pipelines_order.order", "pipeline_ids.#", regexp.MustCompile("[^0].*$")),
			},
			{
				Config: testAccDatasourceDatadogLogsPipelinesOrderConfig(),
				// given the persistent nature of log pipelines we can't check for exact value
				// we can assume that at least one pipeline already exists given the previous test
				Check: resource.TestMatchResourceAttr("data.datadog_logs_pipelines_order.order", "pipeline_ids.#", regexp.MustCompile("[^0].*$")),
			},
		},
	})
}

func testAccDatasourceDatadogLogsPipelinesOrderWithPipelineConfig(uniq string) string {
	return fmt.Sprintf(`
data "datadog_logs_pipelines_order" "order" {
	depends_on = [
		datadog_logs_custom_pipeline.sample_pipeline,
	]
}

resource "datadog_logs_custom_pipeline" "sample_pipeline" {
	name           = "%s-sample"
	filter {
		query = "project:sample"
	}
}
`, uniq)
}

func testAccDatasourceDatadogLogsPipelinesOrderConfig() string {
	return `
data "datadog_logs_pipelines_order" "order" {}
`
}
