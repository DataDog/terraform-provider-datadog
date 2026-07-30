package test

import "testing"

const datadogDashboardV2ProductAnalyticsQueriesConfig = `
resource "datadog_dashboard_v2" "product_analytics_queries_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    query_value_definition {
      title = "Product Analytics Extended"
      request {
        formula {
          formula_expression = "extended_query"
        }
        query {
          product_analytics_extended_query {
            data_source = "product_analytics_extended"
            name        = "extended_query"
            indexes     = ["*"]
            query {
              data_source = "product_analytics"
              search {
                query = "@type:view @view.name:/checkout"
              }
            }
            compute {
              aggregation = "count"
              name        = "views"
            }
            group_by {
              facet = "@geo.country"
              limit = 10
              sort {
                aggregation = "count"
                order       = "desc"
              }
            }
            audience_filters {
              user {
                name  = "buyers"
                query = "@usr.plan:paid"
              }
            }
          }
        }
      }
    }
  }

  widget {
    query_value_definition {
      title = "User Journey"
      request {
        formula {
          formula_expression = "journey_query"
        }
        query {
          user_journey_query {
            data_source = "product_analytics_journey"
            name        = "journey_query"
            search {
              node_objects = jsonencode({
                node_0 = {
                  data_source = "product_analytics"
                  search      = { query = "@type:view @view.name:/home" }
                }
                node_1 = {
                  data_source = "product_analytics"
                  search      = { query = "@type:action @action.name:checkout" }
                }
              })
              expression = "node_0 -> node_1"
              step_aliases = jsonencode({
                node_0 = "Home"
                node_1 = "Checkout"
              })
              join_keys {
                primary   = "@session.id"
                secondary = ["@usr.id"]
              }
            }
            compute {
              aggregation = "count"
              metric      = "__dd.conversion_rate"
              target {
                type  = "step"
                value = "node_1"
              }
            }
          }
        }
      }
    }
  }

  widget {
    query_value_definition {
      title = "Retention"
      request {
        formula {
          formula_expression = "retention_query"
        }
        query {
          retention_query {
            data_source = "product_analytics_retention"
            name        = "retention_query"
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
            group_by {
              facet  = "@geo.country"
              target = "cohort"
              sort {
                order = "desc"
              }
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2ProductAnalyticsQueriesAsserts = []string{
	"title = {{uniq}}",
	"widget.0.query_value_definition.0.request.0.query.0.product_analytics_extended_query.0.data_source = product_analytics_extended",
	"widget.0.query_value_definition.0.request.0.query.0.product_analytics_extended_query.0.compute.0.aggregation = count",
	"widget.0.query_value_definition.0.request.0.query.0.product_analytics_extended_query.0.group_by.0.facet = @geo.country",
	"widget.1.query_value_definition.0.request.0.query.0.user_journey_query.0.data_source = product_analytics_journey",
	"widget.1.query_value_definition.0.request.0.query.0.user_journey_query.0.search.0.expression = node_0 -> node_1",
	"widget.1.query_value_definition.0.request.0.query.0.user_journey_query.0.compute.0.metric = __dd.conversion_rate",
	"widget.2.query_value_definition.0.request.0.query.0.retention_query.0.data_source = product_analytics_retention",
	"widget.2.query_value_definition.0.request.0.query.0.retention_query.0.search.0.retention_entity = @usr.id",
	"widget.2.query_value_definition.0.request.0.query.0.retention_query.0.compute.0.metric = __dd.retention_rate",
}

func TestAccDatadogDashboardV2ProductAnalyticsQueries(t *testing.T) {
	config, name := datadogDashboardV2ProductAnalyticsQueriesConfig, "datadog_dashboard_v2.product_analytics_queries_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2ProductAnalyticsQueries", config, name, datadogDashboardV2ProductAnalyticsQueriesAsserts)
}

func TestAccDatadogDashboardV2ProductAnalyticsQueries_import(t *testing.T) {
	config, name := datadogDashboardV2ProductAnalyticsQueriesConfig, "datadog_dashboard_v2.product_analytics_queries_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2ProductAnalyticsQueries_import", config, name)
}
