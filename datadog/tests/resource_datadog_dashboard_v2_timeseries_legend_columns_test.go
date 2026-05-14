package test

import (
	"testing"
)

// datadogDashboardV2TimeseriesLegendColumnsConfig tests timeseries widgets with
// legend_columns specified in non-alphabetical order. This exercises the SortNormalized
// plan modifier — without it, the API may return legend_columns in a different order
// than the HCL config, causing spurious diffs on terraform plan after import.
const datadogDashboardV2TimeseriesLegendColumnsConfig = `
resource "datadog_dashboard_v2" "timeseries_legend_columns_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"

    widget {
        timeseries_definition {
            title          = "CPU Utilization with legend columns"
            legend_columns = ["avg", "max", "min", "sum", "value"]
            legend_layout  = "auto"
            show_legend    = true

            request {
                display_type = "bars"

                formula {
                    formula_expression = "query1"
                }

                query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "avg:system.cpu.user{*}"
                    }
                }

                style {
                    line_type  = "solid"
                    line_width = "normal"
                    palette    = "dog_classic"
                }
            }

            yaxis {
                include_zero = true
                max          = "auto"
                min          = "auto"
                scale        = "linear"
            }
        }
    }
}
`

var datadogDashboardV2TimeseriesLegendColumnsAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",

	"widget.0.timeseries_definition.0.title = CPU Utilization with legend columns",
	"widget.0.timeseries_definition.0.legend_columns.# = 5",
	"widget.0.timeseries_definition.0.legend_columns.0 = avg",
	"widget.0.timeseries_definition.0.legend_columns.1 = max",
	"widget.0.timeseries_definition.0.legend_columns.2 = min",
	"widget.0.timeseries_definition.0.legend_columns.3 = sum",
	"widget.0.timeseries_definition.0.legend_columns.4 = value",
	"widget.0.timeseries_definition.0.legend_layout = auto",
	"widget.0.timeseries_definition.0.show_legend = true",

	"widget.0.timeseries_definition.0.request.0.display_type = bars",
	"widget.0.timeseries_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*}",
	"widget.0.timeseries_definition.0.request.0.style.0.line_type = solid",
	"widget.0.timeseries_definition.0.request.0.style.0.line_width = normal",
	"widget.0.timeseries_definition.0.request.0.style.0.palette = dog_classic",

	"widget.0.timeseries_definition.0.yaxis.0.include_zero = true",
	"widget.0.timeseries_definition.0.yaxis.0.max = auto",
	"widget.0.timeseries_definition.0.yaxis.0.min = auto",
	"widget.0.timeseries_definition.0.yaxis.0.scale = linear",
}

func TestAccDatadogDashboardV2TimeseriesLegendColumns(t *testing.T) {
	config, name := datadogDashboardV2TimeseriesLegendColumnsConfig, "datadog_dashboard_v2.timeseries_legend_columns_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2TimeseriesLegendColumns", config, name, datadogDashboardV2TimeseriesLegendColumnsAsserts)
}

func TestAccDatadogDashboardV2TimeseriesLegendColumns_import(t *testing.T) {
	config, name := datadogDashboardV2TimeseriesLegendColumnsConfig, "datadog_dashboard_v2.timeseries_legend_columns_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2TimeseriesLegendColumns_import", config, name)
}
