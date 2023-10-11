package test

import (
	"testing"
)

const datadogDashboardSplitGraphConfigWithStaticSplits = `
resource "datadog_dashboard" "split_graph_dashboard" {
	title       = "{{uniq}}"
	description = "Created using the Datadog provider in Terraform"
	layout_type = "ordered"
	widget {
		split_graph_definition {
			title = "Terraform Split Graph Widget"
			source_widget_definition {
				timeseries_definition {
					title_size  = "16"
					title_align = "left"
					title       = "system.cpu.user"
					request {
						query {
							metric_query {
								data_source = "metrics"
								query       = "avg:system.cpu.user{*}"
								name        = "my_query_1"
							}
						}
						style {
							line_width = "thin"
							palette    = "dog_classic"
							line_type  = "solid"
						}
						display_type = "line"
					}
				}
			}
			split_config {
				split_dimensions {
					one_graph_per = "service"
				}
				limit = 24
				sort {
					compute {
						aggregation = "sum"
						metric      = "system.cpu.user"
					}
					order = "desc"
				}
				static_splits {
					split_vector {
						tag_key    = "service"
						tag_values = ["cassandra"]
					}
					split_vector {
						tag_key    = "datacenter"
						tag_values = []
					}
				}
				static_splits {
					split_vector {
						tag_key    = "demo"
						tag_values = ["env"]
					}
				}
			}
			size               = "md"
			has_uniform_y_axes = true
			live_span          = "5m"
		}
	}
}
`

var datadogDashboardSplitGraphAssertsWithStaticSplits = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"title = {{uniq}}",
	"widget.0.split_graph_definition.0.title = Terraform Split Graph Widget",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.title_size = 16",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.title_align = left",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.title = system.cpu.user",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*}",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.style.0.palette = dog_classic",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.display_type = line",
	"widget.0.split_graph_definition.0.split_config.0.split_dimensions.0.one_graph_per = service",
	"widget.0.split_graph_definition.0.split_config.0.limit = 24",
	"widget.0.split_graph_definition.0.split_config.0.sort.0.compute.0.aggregation = sum",
	"widget.0.split_graph_definition.0.split_config.0.sort.0.compute.0.metric = system.cpu.user",
	"widget.0.split_graph_definition.0.split_config.0.sort.0.order = desc",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.0.split_vector.0.tag_key = service",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.0.split_vector.0.tag_values.0 = cassandra",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.0.split_vector.0.tag_values.# = 1",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.0.split_vector.1.tag_key = datacenter",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.0.split_vector.1.tag_values.# = 0",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.1.split_vector.0.tag_key = demo",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.1.split_vector.0.tag_values.0 = env",
	"widget.0.split_graph_definition.0.split_config.0.static_splits.1.split_vector.0.tag_values.# = 1",
	"widget.0.split_graph_definition.0.size = md",
	"widget.0.split_graph_definition.0.has_uniform_y_axes = true",
	"widget.0.split_graph_definition.0.live_span = 5m",
}

func TestAccDatadogDashboardSplitGraphWithStaticSplits(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSplitGraphConfigWithStaticSplits, "datadog_dashboard.split_graph_dashboard", datadogDashboardSplitGraphAssertsWithStaticSplits)
}

func TestAccDatadogDashboardSplitGraphWithStaticSplits_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSplitGraphConfigWithStaticSplits, "datadog_dashboard.split_graph_dashboard")
}

const datadogDashboardSplitGraphConfig = `
resource "datadog_dashboard" "split_graph_dashboard" {
	title       = "{{uniq}}"
	description = "Created using the Datadog provider in Terraform"
	layout_type = "ordered"
	widget {
		split_graph_definition {
			title = "Terraform Split Graph Widget"
			source_widget_definition {
				timeseries_definition {
					title_size  = "16"
					title_align = "left"
					title       = "system.cpu.user"
					request {
						query {
							metric_query {
								data_source = "metrics"
								query       = "avg:system.cpu.user{*}"
								name        = "my_query_1"
							}
						}
						style {
							line_width = "thin"
							palette    = "dog_classic"
							line_type  = "solid"
						}
						display_type = "line"
					}
				}
			}
			split_config {
				split_dimensions {
					one_graph_per = "service"
				}
				limit = 24
				sort {
					order = "asc"
				}
			}
			size      = "md"
			live_span = "5m"
		}
	}
}
`

var datadogDashboardSplitGraphAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"title = {{uniq}}",
	"widget.0.split_graph_definition.0.title = Terraform Split Graph Widget",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.title_size = 16",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.title_align = left",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.title = system.cpu.user",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*}",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.query.0.metric_query.0.name = my_query_1",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.style.0.palette = dog_classic",
	"widget.0.split_graph_definition.0.source_widget_definition.0.timeseries_definition.0.request.0.display_type = line",
	"widget.0.split_graph_definition.0.split_config.0.split_dimensions.0.one_graph_per = service",
	"widget.0.split_graph_definition.0.split_config.0.limit = 24",
	"widget.0.split_graph_definition.0.split_config.0.sort.0.order = asc",
	"widget.0.split_graph_definition.0.size = md",
	"widget.0.split_graph_definition.0.live_span = 5m",
}

func TestAccDatadogDashboardSplitGraph(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSplitGraphConfig, "datadog_dashboard.split_graph_dashboard", datadogDashboardSplitGraphAsserts)
}

func TestAccDatadogDashboardSplitGraphWith_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSplitGraphConfig, "datadog_dashboard.split_graph_dashboard")
}
