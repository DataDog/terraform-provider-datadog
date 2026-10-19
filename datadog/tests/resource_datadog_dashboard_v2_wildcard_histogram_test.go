package test

import (
	"testing"
)

const datadogDashboardWildcardHistogramConfig = `
resource "datadog_dashboard_v2" "wildcard_histogram_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        wildcard_definition {
            title = "Latency distribution by endpoint"
            specification {
                type     = "vega-lite"
                contents = jsonencode({
                    "$schema" = "https://vega.github.io/schema/vega-lite/v5.json"
                    data      = { name = "query1" }
                    mark      = "bar"
                    encoding  = {
                        x = { field = "bucket", type = "ordinal", title = "Latency bucket" }
                        y = { field = "count",  type = "quantitative", title = "Sample count" }
                    }
                })
            }
            request {
                histogram_request {
                    histogram_query {
                        metric_query {
                            data_source = "metrics"
                            name        = "query1"
                            query       = "histogram:trace.Load{*}"
                        }
                    }
                    style {
                        palette = "dog_classic"
                    }
                }
            }
        }
    }
}
`

var datadogDashboardWildcardHistogramAsserts = []string{
	"title = {{uniq}}",
	"widget.0.wildcard_definition.0.title = Latency distribution by endpoint",
	"widget.0.wildcard_definition.0.specification.0.type = vega-lite",
	"widget.0.wildcard_definition.0.request.0.histogram_request.0.histogram_query.0.metric_query.0.data_source = metrics",
	"widget.0.wildcard_definition.0.request.0.histogram_request.0.histogram_query.0.metric_query.0.name = query1",
	"widget.0.wildcard_definition.0.request.0.histogram_request.0.histogram_query.0.metric_query.0.query = histogram:trace.Load{*}",
	"widget.0.wildcard_definition.0.request.0.histogram_request.0.style.0.palette = dog_classic",
}

func TestAccDatadogDashboardV2WildcardHistogram(t *testing.T) {
	config, name := datadogDashboardWildcardHistogramConfig, "datadog_dashboard_v2.wildcard_histogram_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2WildcardHistogram", config, name, datadogDashboardWildcardHistogramAsserts)
}

func TestAccDatadogDashboardV2WildcardHistogram_import(t *testing.T) {
	config, name := datadogDashboardWildcardHistogramConfig, "datadog_dashboard_v2.wildcard_histogram_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2WildcardHistogram_import", config, name)
}
