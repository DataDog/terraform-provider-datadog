package datadog

import "fmt"

func pipelinesConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "pipeline_1" {
	name = "%s-first"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_custom_pipeline" "pipeline_2" {
	name = "%s-second"
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
	name = "%s"
	pipelines = [
		"${datadog_logs_custom_pipeline.pipeline_1.id}",
		"${datadog_logs_custom_pipeline.pipeline_2.id}"
	]
}`, uniq, uniq, uniq)
}

func orderUpdateConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_logs_custom_pipeline" "pipeline_1" {
	name = "%s-first"
	is_enabled = true
	filter {
		query = "source:redis"
	}
}
resource "datadog_logs_custom_pipeline" "pipeline_2" {
	name = "%s-second"
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
	name = "%s"
	pipelines = [
		"${datadog_logs_custom_pipeline.pipeline_2.id}",
		"${datadog_logs_custom_pipeline.pipeline_1.id}"
	]
}`, uniq, uniq, uniq)
}
