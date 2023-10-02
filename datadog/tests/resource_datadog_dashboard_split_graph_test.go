package test

import (
	"testing"
)

const datadogDashboardSplitGraphConfig = `
resource "datadog_dashboard" "sunburst_dashboard" {
  title        = "{{uniq}}"
  description  = "Created using the Datadog provider in Terraform"
  layout_type  = "ordered"
  widget {
    split_graph_definition {
		title = "hello",
		source_widget_definition{
			title_size = "16"
			title_align = "left"
			title = "system.cpu.user"
			request {
				query {
					metric_query {
					  data_source = "metrics"
					  query       = "avg:system.cpu.user{foo:bar} by {env}"
					  name        = "my_query_1"
					  aggregator  = "sum"
					}
				}
				style {
					line_width = "thin"
					palette = "dog_classic"
					line_type = "solid"
				}
				display_type = "line"
			}
		},
		split_config{
			split_dimensions{
				one_graph_per = "service"
			}
			limit = 24
			sort{
				compute{
					aggregation = "sum"
					metric = "system.cpu.user"
				}
				order = "desc"
			}
			static_splits{
				split_vector{
					tag_key = "service"
					tag_values = ["cassandra"]
				}
				split_vector{
					tag_key = "datacenter"
					tag_values = []
				}
			}
			static_splits{
				split_vector{
					tag_key = "demo"
					tag_values = ["env"]
				}
			}
		},
		size = "md",
		has_uniform_y_axes = "true",
		live_span = "5m"
	}
  }
}
`

var datadogDashboardSplitGraphAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"title = {{uniq}}",
	"widget.0.split_graph_definition.0.title = hello",
	"widget.0.split_graph_definition.0.source_widget_definition.title_size = 16",
	"widget.0.split_graph_definition.0.source_widget_definition.title_align = left",
	"widget.0.split_graph_definition.0.source_widget_definition.title = system.cpu.user",
	"widget.0.split_graph_definition.0.source_widget_definition.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.split_graph_definition.0.source_widget_definition.request.0.query.0.metric_query.0.query = avg:system.cpu.user{foo:bar} by {env}",
	"widget.0.split_graph_definition.0.source_widget_definition.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.split_graph_definition.0.source_widget_definition.request.0.query.0.metric_query.0.aggregator = sum",
	"widget.0.split_graph_definition.0.source_widget_definition.request.0.style.0.palette = dog_classic",
	"widget.0.split_graph_definition.0.source_widget_definition.request.0.display_type = line",
}

func TestAccDatadogDashboardSplitGraph(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSunburstConfig, "datadog_dashboard.split_graph_dashboard", datadogDashboardSplitGraphAsserts)
}

func TestAccDatadogDashboardSplitGraph_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSplitGraphConfig, "datadog_dashboard.split_graph_dashboard")
}
