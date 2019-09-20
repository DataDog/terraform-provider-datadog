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

resource "datadog_logs_pipeline_order" "pipelines" {
	depends_on = [
		"datadog_logs_pipeline.pipeline_1",
		"datadog_logs_pipeline.pipeline_2"
	]
	name = "pipelines"
	pipelines = [
		"kGZarioHSEGNPGyy9gISkw",
        "EZXoa97wSHWnFNglBAB91Q",
        "xUjMTstsS0WPRNOFzxH5vg",
		"${datadog_logs_pipeline.pipeline_1.id}",
		"${datadog_logs_pipeline.pipeline_2.id}"
	]
}
`

const orderUpdateConfig = `
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

resource "datadog_logs_pipeline_order" "pipelines" {
	depends_on = [
		"datadog_logs_pipeline.pipeline_1",
		"datadog_logs_pipeline.pipeline_2"
	]
	name = "pipelines"
	pipelines = [
		"kGZarioHSEGNPGyy9gISkw",
        "EZXoa97wSHWnFNglBAB91Q",
        "xUjMTstsS0WPRNOFzxH5vg",
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
						"datadog_logs_pipeline_order.pipelines", "name", "pipelines"),
					resource.TestCheckResourceAttr(
						"datadog_logs_pipeline_order.pipelines", "pipelines.#", "5"),
				),
			},
			{
				Config: orderUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPipelineExists("datadog_logs_pipeline.pipeline_2"),
					testAccCheckPipelineExists("datadog_logs_pipeline.pipeline_1"),
					resource.TestCheckResourceAttr(
						"datadog_logs_pipeline_order.pipelines", "name", "pipelines"),
					resource.TestCheckResourceAttr(
						"datadog_logs_pipeline_order.pipelines", "pipelines.#", "5"),
				),
			},
		},
	})
}
