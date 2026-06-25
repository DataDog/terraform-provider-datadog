package test

import (
	"testing"
)

const datadogDashboardPointPlotConfig = `
resource "datadog_dashboard_v2" "point_plot_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        point_plot_definition {
            title = "Point Plot Widget"
            request {
                request_type = "data_projection"
                query {
                    query_string = "service:web-store"
                    data_source  = "logs"
                }
                projection {
                    type = "point_plot"
                    dimension {
                        column    = "time"
                        dimension = "time"
                    }
                    dimension {
                        column    = "duration"
                        dimension = "y"
                    }
                }
            }
            legend {
                type = "automatic"
            }
        }
    }
}
`

var datadogDashboardPointPlotAsserts = []string{
	"title = {{uniq}}",
	"widget.0.point_plot_definition.0.title = Point Plot Widget",
	"widget.0.point_plot_definition.0.request.0.request_type = data_projection",
	"widget.0.point_plot_definition.0.request.0.query.0.query_string = service:web-store",
	"widget.0.point_plot_definition.0.request.0.query.0.data_source = logs",
	"widget.0.point_plot_definition.0.request.0.projection.0.type = point_plot",
	"widget.0.point_plot_definition.0.request.0.projection.0.dimension.0.column = time",
	"widget.0.point_plot_definition.0.request.0.projection.0.dimension.0.dimension = time",
	"widget.0.point_plot_definition.0.request.0.projection.0.dimension.1.column = duration",
	"widget.0.point_plot_definition.0.request.0.projection.0.dimension.1.dimension = y",
	"widget.0.point_plot_definition.0.legend.0.type = automatic",
}

func TestAccDatadogDashboardV2PointPlot(t *testing.T) {
	config, name := datadogDashboardPointPlotConfig, "datadog_dashboard_v2.point_plot_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2PointPlot", config, name, datadogDashboardPointPlotAsserts)
}

func TestAccDatadogDashboardV2PointPlot_import(t *testing.T) {
	config, name := datadogDashboardPointPlotConfig, "datadog_dashboard_v2.point_plot_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2PointPlot", config, name)
}
