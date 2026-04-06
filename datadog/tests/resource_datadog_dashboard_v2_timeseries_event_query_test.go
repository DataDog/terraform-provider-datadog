package test

import (
	"testing"
)

const datadogDashboardV2TimeseriesEventQueryConfig = `
resource "datadog_dashboard_v2" "timeseries_event_query_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"
    widget {
        timeseries_definition {
            request {
                event_query {
                    index = "*"
                    search_query = "status:error"
                    compute_query {
                        aggregation = "count"
                    }
                    group_by {
                        facet = "service"
                        limit = 10
                        sort_query {
                            aggregation = "count"
                            order       = "desc"
                        }
                    }
                }
                display_type = "bars"
            }
            title = "Timeseries with event_query"
        }
    }
}
`

var datadogDashboardV2TimeseriesEventQueryAsserts = []string{
	"title = {{uniq}}",
	"widget.0.timeseries_definition.0.request.0.event_query.0.index = *",
	"widget.0.timeseries_definition.0.request.0.event_query.0.search_query = status:error",
	"widget.0.timeseries_definition.0.request.0.event_query.0.compute_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.0.event_query.0.group_by.0.facet = service",
	"widget.0.timeseries_definition.0.request.0.event_query.0.group_by.0.limit = 10",
	"widget.0.timeseries_definition.0.request.0.event_query.0.group_by.0.sort_query.0.aggregation = count",
	"widget.0.timeseries_definition.0.request.0.event_query.0.group_by.0.sort_query.0.order = desc",
	"widget.0.timeseries_definition.0.request.0.display_type = bars",
	"widget.0.timeseries_definition.0.title = Timeseries with event_query",
}

func TestAccDatadogDashboardV2TimeseriesEventQuery(t *testing.T) {
	config, name := datadogDashboardV2TimeseriesEventQueryConfig, "datadog_dashboard_v2.timeseries_event_query_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2TimeseriesEventQuery", config, name, datadogDashboardV2TimeseriesEventQueryAsserts)
}

func TestAccDatadogDashboardV2TimeseriesEventQuery_import(t *testing.T) {
	config, name := datadogDashboardV2TimeseriesEventQueryConfig, "datadog_dashboard_v2.timeseries_event_query_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2TimeseriesEventQuery_import", config, name)
}
