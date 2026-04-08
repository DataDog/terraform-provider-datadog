package test

import (
	"testing"
)

// datadogDashboardV2QueryValueConditionalFormatsConfig tests query_value widgets
// with formula-style requests that include conditional_formats at the request level.
// This exercises the scalarWithConditionalFormatsConfig fix — without it, conditional_formats
// are silently dropped during import/read for formula-capable widgets.
const datadogDashboardV2QueryValueConditionalFormatsConfig = `
resource "datadog_dashboard_v2" "query_value_conditional_formats_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"

    widget {
        query_value_definition {
            autoscale = true
            precision = 0
            title     = "Error count with conditional formats"

            request {
                conditional_formats {
                    comparator = "<"
                    hide_value = false
                    palette    = "white_on_green"
                    value      = 1
                }
                conditional_formats {
                    comparator = "<"
                    hide_value = false
                    palette    = "black_on_light_yellow"
                    value      = 5
                }
                conditional_formats {
                    comparator = ">="
                    hide_value = false
                    palette    = "black_on_light_red"
                    value      = 5
                }

                formula {
                    formula_expression = "default_zero(query1)"
                }

                query {
                    metric_query {
                        aggregator  = "sum"
                        data_source = "metrics"
                        name        = "query1"
                        query       = "sum:system.cpu.user{*}.as_count()"
                    }
                }
            }

            timeseries_background {
                type = "area"
                yaxis {
                    include_zero = true
                }
            }
        }
    }
}
`

var datadogDashboardV2QueryValueConditionalFormatsAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",

	"widget.0.query_value_definition.0.autoscale = true",
	"widget.0.query_value_definition.0.precision = 0",
	"widget.0.query_value_definition.0.title = Error count with conditional formats",

	"widget.0.query_value_definition.0.request.0.conditional_formats.# = 3",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.value = 1",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.comparator = <",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.palette = black_on_light_yellow",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.value = 5",
	"widget.0.query_value_definition.0.request.0.conditional_formats.2.comparator = >=",
	"widget.0.query_value_definition.0.request.0.conditional_formats.2.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.2.palette = black_on_light_red",
	"widget.0.query_value_definition.0.request.0.conditional_formats.2.value = 5",

	"widget.0.query_value_definition.0.request.0.formula.0.formula_expression = default_zero(query1)",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.query_value_definition.0.request.0.query.0.metric_query.0.query = sum:system.cpu.user{*}.as_count()",

	"widget.0.query_value_definition.0.timeseries_background.0.type = area",
	"widget.0.query_value_definition.0.timeseries_background.0.yaxis.0.include_zero = true",
}

func TestAccDatadogDashboardV2QueryValueConditionalFormats(t *testing.T) {
	config, name := datadogDashboardV2QueryValueConditionalFormatsConfig, "datadog_dashboard_v2.query_value_conditional_formats_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2QueryValueConditionalFormats", config, name, datadogDashboardV2QueryValueConditionalFormatsAsserts)
}

func TestAccDatadogDashboardV2QueryValueConditionalFormats_import(t *testing.T) {
	config, name := datadogDashboardV2QueryValueConditionalFormatsConfig, "datadog_dashboard_v2.query_value_conditional_formats_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2QueryValueConditionalFormats_import", config, name)
}
