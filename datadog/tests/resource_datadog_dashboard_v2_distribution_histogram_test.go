package test

import (
	"testing"
)

const datadogDashboardDistributionHistogramConfig = `
resource "datadog_dashboard_v2" "distribution_histogram_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        distribution_definition {
            title = "CPU Usage Distribution"
            request {
                request_type = "histogram"
                histogram_query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "avg:container.cpu.usage.dist{*}"
                    }
                }
                style {
                    palette = "cool"
                }
            }
        }
    }
}
`

var datadogDashboardDistributionHistogramAsserts = []string{
	"title = {{uniq}}",
	"widget.0.distribution_definition.0.title = CPU Usage Distribution",
	"widget.0.distribution_definition.0.request.0.request_type = histogram",
	"widget.0.distribution_definition.0.request.0.histogram_query.0.metric_query.0.data_source = metrics",
	"widget.0.distribution_definition.0.request.0.histogram_query.0.metric_query.0.name = query1",
	"widget.0.distribution_definition.0.request.0.histogram_query.0.metric_query.0.query = avg:container.cpu.usage.dist{*}",
	"widget.0.distribution_definition.0.request.0.style.0.palette = cool",
}

func TestAccDatadogDashboardV2DistributionHistogram(t *testing.T) {
	config, name := datadogDashboardDistributionHistogramConfig, "datadog_dashboard_v2.distribution_histogram_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2DistributionHistogram", config, name, datadogDashboardDistributionHistogramAsserts)
}

func TestAccDatadogDashboardV2DistributionHistogram_import(t *testing.T) {
	config, name := datadogDashboardDistributionHistogramConfig, "datadog_dashboard_v2.distribution_histogram_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2DistributionHistogram_import", config, name)
}
