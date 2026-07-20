package test

import "testing"

const datadogDashboardV2HeatmapHistogramConfig = `
resource "datadog_dashboard_v2" "heatmap_histogram_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    heatmap_definition {
      title       = "Request duration distribution"
      show_legend = true

      xaxis {
        num_buckets = 60
      }

      request {
        request_type = "histogram"

        histogram_query {
          metric_query {
            data_source = "metrics"
            name        = "query1"
            query       = "histogram:trace.servlet.request{*}"
          }
        }

        style {
          palette = "dog_classic"
        }
      }
    }
  }
}
`

var datadogDashboardV2HeatmapHistogramAsserts = []string{
	"title = {{uniq}}",
	"widget.0.heatmap_definition.0.title = Request duration distribution",
	"widget.0.heatmap_definition.0.show_legend = true",
	"widget.0.heatmap_definition.0.xaxis.0.num_buckets = 60",
	"widget.0.heatmap_definition.0.request.0.request_type = histogram",
	"widget.0.heatmap_definition.0.request.0.histogram_query.0.metric_query.0.data_source = metrics",
	"widget.0.heatmap_definition.0.request.0.histogram_query.0.metric_query.0.name = query1",
	"widget.0.heatmap_definition.0.request.0.histogram_query.0.metric_query.0.query = histogram:trace.servlet.request{*}",
	"widget.0.heatmap_definition.0.request.0.style.0.palette = dog_classic",
}

func TestAccDatadogDashboardV2HeatmapHistogram(t *testing.T) {
	config, name := datadogDashboardV2HeatmapHistogramConfig, "datadog_dashboard_v2.heatmap_histogram_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2HeatmapHistogram", config, name, datadogDashboardV2HeatmapHistogramAsserts)
}

func TestAccDatadogDashboardV2HeatmapHistogram_import(t *testing.T) {
	config, name := datadogDashboardV2HeatmapHistogramConfig, "datadog_dashboard_v2.heatmap_histogram_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2HeatmapHistogram_import", config, name)
}
