package test

import "testing"

const datadogDashboardV2SankeyAudienceConfig = `
resource "datadog_dashboard_v2" "sankey_audience_dashboard" {
  title       = "{{uniq}}"
  layout_type = "ordered"

  widget {
    sankey_definition {
      title = "Product Analytics audience"

      request {
        rum_request {
          query {
            data_source  = "product_analytics"
            query_string = "@type:view"
            mode         = "source"

            audience_filters {
              user {
                name  = "buyers"
                query = "@usr.plan:pro"
              }

              segment {
                name       = "returning-users"
                segment_id = "segment-id"
              }

              account {
                name  = "enterprise-accounts"
                query = "@account.tier:enterprise"
              }

              filter_condition = "users and segments and accounts"
            }

            occurrences {
              operator = "gt"
              value    = "2"
            }

            join_keys {
              primary   = "session.id"
              secondary = ["usr.id", "account.id"]
            }
          }
        }
      }
    }
  }
}
`

var datadogDashboardV2SankeyAudienceAsserts = []string{
	"title = {{uniq}}",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.data_source = product_analytics",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.user.0.name = buyers",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.user.0.query = @usr.plan:pro",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.segment.0.name = returning-users",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.segment.0.segment_id = segment-id",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.account.0.name = enterprise-accounts",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.account.0.query = @account.tier:enterprise",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.audience_filters.0.filter_condition = users and segments and accounts",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.occurrences.0.operator = gt",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.occurrences.0.value = 2",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.join_keys.0.primary = session.id",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.join_keys.0.secondary.0 = usr.id",
	"widget.0.sankey_definition.0.request.0.rum_request.0.query.0.join_keys.0.secondary.1 = account.id",
}

func TestAccDatadogDashboardV2SankeyAudience(t *testing.T) {
	config, name := datadogDashboardV2SankeyAudienceConfig, "datadog_dashboard_v2.sankey_audience_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2SankeyAudience", config, name, datadogDashboardV2SankeyAudienceAsserts)
}

func TestAccDatadogDashboardV2SankeyAudience_import(t *testing.T) {
	config, name := datadogDashboardV2SankeyAudienceConfig, "datadog_dashboard_v2.sankey_audience_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2SankeyAudience_import", config, name)
}
