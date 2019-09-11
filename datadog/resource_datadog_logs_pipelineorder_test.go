package datadog

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

const pipelinesConfig = `
resource "datadog_logs_pipeline" "pipeline_1" {
	name = "my first pipeline"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_pipeline" "pipeline_2" {
	name = "my second pipeline"
	is_enabled = true
	filter {
		query = "source:agent"
	}
}

resource "datadog_logs_pipelineorder" "pipelines" {
	depends_on = [
		"datadog_logs_pipeline.pipeline_1",
		"datadog_logs_pipeline.pipeline_2"
	]
	name = "pipelines"
	pipelines = [
		"-fhApkC3S0uzZmztKPaGWA",
        "TPOHRa2PS2WNbfQOu0Oaxw",
        "DltQ5IkGQOOXPNc_ut6X2w",
		"${datadog_logs_pipeline.pipeline_1.id}",
		"${datadog_logs_pipeline.pipeline_2.id}"
	]
}
`

const orderUpdateConfig = `
resource "datadog_logs_pipelineorder" "pipelines" {
	depends_on = [
		"datadog_logs_pipeline.pipeline_1",
		"datadog_logs_pipeline.pipeline_2"
	]
	name = "pipelines"
	pipelines = [
		"-fhApkC3S0uzZmztKPaGWA",
		"TPOHRa2PS2WNbfQOu0Oaxw",
		"DltQ5IkGQOOXPNc_ut6X2w",
		"${datadog_logs_pipeline.pipeline_2.id}",
		"${datadog_logs_pipeline.pipeline_1.id}"
	]
}
`

func TestAccDatadogLogsPipelineOrder_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPipelineDestroy,
		Steps: []resource.TestStep{
			{
				Config: pipelinesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists("datadog_logs_pipeline.pipeline_1"),
					testAccCheckPipelineExists("datadog_logs_pipeline.pipeline_2"),
					resource.TestCheckResourceAttr(
						"datadog_logs_pipelineorder.pipelines", "name", "pipelines"),
					resource.TestCheckResourceAttr(
						"datadog_logs_pipelineorder.pipelines", "pipelines.#", "5"),
				),
			},
		},
	})
}
