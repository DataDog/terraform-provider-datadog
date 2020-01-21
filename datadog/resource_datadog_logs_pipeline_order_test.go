package datadog

import (
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

const pipelinesConfig = `
resource "datadog_logs_custom_pipeline" "pipeline_1" {
	name = "my first pipeline"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_custom_pipeline" "pipeline_2" {
	name = "my second pipeline"
	is_enabled = true
	filter {
		query = "source:agent"
	}
}

resource "datadog_logs_pipeline_order" "pipelines" {
	depends_on = [
		"datadog_logs_custom_pipeline.pipeline_1",
		"datadog_logs_custom_pipeline.pipeline_2"
	]
	name = "pipelines"
	pipelines = [
		"${datadog_logs_custom_pipeline.pipeline_1.id}",
		"${datadog_logs_custom_pipeline.pipeline_2.id}"
	]
}
`

const orderUpdateConfig = `
resource "datadog_logs_custom_pipeline" "pipeline_1" {
	name = "my first pipeline"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_custom_pipeline" "pipeline_2" {
	name = "my second pipeline"
	is_enabled = true
	filter {
		query = "source:agent"
	}
}

resource "datadog_logs_pipeline_order" "pipelines" {
	depends_on = [
		"datadog_logs_custom_pipeline.pipeline_1",
		"datadog_logs_custom_pipeline.pipeline_2"
	]
	name = "pipelines"
	pipelines = [
		"${datadog_logs_custom_pipeline.pipeline_2.id}",
		"${datadog_logs_custom_pipeline.pipeline_1.id}"
	]
}
`
