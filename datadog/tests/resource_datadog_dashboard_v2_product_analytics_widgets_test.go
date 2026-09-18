package test

import "testing"

const datadogDashboardV2ProductAnalyticsWidgetsConfig = `
resource "datadog_dashboard_v2" "product_analytics_widgets_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    product_analytics_funnel_definition {
      title           = "Checkout journey"
      grouped_display = "stacked"

      request {
        request_type       = "user_journey_funnel"
        comparison_segments = ["paid"]

        query {
          data_source = "product_analytics_journey"

          search {
            node_objects = jsonencode({
              step1 = {
                data_source = "product_analytics"
                search      = { query = "@type:view @view.name:/home" }
              }
              step2 = {
                data_source = "product_analytics"
                search      = { query = "@type:action @action.name:checkout" }
              }
            })
            expression = "step1 -> step2"
          }

          compute {
            aggregation = "count"
            metric      = "__dd.conversion_rate"
          }

          group_by {
            facet = "@usr.email"
            sort {
              aggregation = "count"
              order       = "desc"
            }
          }
        }

        comparison_time {
          type = "previous_week"
        }
      }
    }
  }

  widget {
    cohort_definition {
      title = "Weekly retention cohort"

      request {
        request_type = "retention_grid"

        query {
          name        = "cohort_query"
          data_source = "product_analytics_retention"

          search {
            cohort_criteria {
              base_query {
                data_source = "product_analytics"
                search {
                  query = "@type:view @view.name:/signup"
                }
              }
              time_interval {
                type = "calendar"
                value {
                  type     = "week"
                  timezone = "UTC"
                }
              }
            }
            retention_entity = "@usr.id"
            return_condition = "conversion_on_or_after"
            return_criteria {
              base_query {
                data_source = "product_analytics"
                search {
                  query = "@type:action @action.name:purchase"
                }
              }
            }
          }

          compute {
            aggregation = "count"
            metric      = "__dd.retention_rate"
          }
        }
      }
    }
  }

  widget {
    retention_curve_definition {
      title = "Weekly retention curve"

      request {
        request_type = "retention_curve"

        query {
          name        = "curve_query"
          data_source = "product_analytics_retention"

          search {
            cohort_criteria {
              base_query {
                data_source = "product_analytics"
                search {
                  query = "@type:view @view.name:/signup"
                }
              }
              time_interval {
                type = "calendar"
                value {
                  type     = "week"
                  timezone = "UTC"
                }
              }
            }
            retention_entity = "@usr.id"
            return_condition = "conversion_on_or_after"
          }

          compute {
            aggregation = "count"
            metric      = "__dd.retention_rate"
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

var datadogDashboardV2ProductAnalyticsWidgetsAsserts = []string{
	"title = {{uniq}}",
	"widget.0.product_analytics_funnel_definition.0.grouped_display = stacked",
	"widget.0.product_analytics_funnel_definition.0.request.0.request_type = user_journey_funnel",
	"widget.0.product_analytics_funnel_definition.0.request.0.query.0.data_source = product_analytics_journey",
	"widget.0.product_analytics_funnel_definition.0.request.0.query.0.compute.0.metric = __dd.conversion_rate",
	"widget.1.cohort_definition.0.request.0.request_type = retention_grid",
	"widget.1.cohort_definition.0.request.0.query.0.data_source = product_analytics_retention",
	"widget.1.cohort_definition.0.request.0.query.0.search.0.retention_entity = @usr.id",
	"widget.2.retention_curve_definition.0.request.0.request_type = retention_curve",
	"widget.2.retention_curve_definition.0.request.0.query.0.compute.0.metric = __dd.retention_rate",
	"widget.2.retention_curve_definition.0.request.0.style.0.palette = dog_classic",
}

func TestAccDatadogDashboardV2ProductAnalyticsWidgets(t *testing.T) {
	config, name := datadogDashboardV2ProductAnalyticsWidgetsConfig, "datadog_dashboard_v2.product_analytics_widgets_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2ProductAnalyticsWidgets", config, name, datadogDashboardV2ProductAnalyticsWidgetsAsserts)
}

func TestAccDatadogDashboardV2ProductAnalyticsWidgets_import(t *testing.T) {
	config, name := datadogDashboardV2ProductAnalyticsWidgetsConfig, "datadog_dashboard_v2.product_analytics_widgets_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2ProductAnalyticsWidgets_import", config, name)
}
