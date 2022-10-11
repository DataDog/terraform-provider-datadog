package test

import (
	"testing"
)

const datadogDashboardSloListConfig = `
resource "datadog_dashboard" "slo_list_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		slo_list_definition {
			request {
				request_type = "slo_list"
				query {
					query_string = "env:prod AND service:my-app"
					limit = 30
				}
			}
			title = "my title"
			title_size = "16"
			title_align = "center"
		}
	}
}
`

var datadogDashboardSloListAsserts = []string{
	"title = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.0.slo_list_definition.0.request.0.query.0.query_string = env:prod AND service:my-app",
	"widget.0.slo_list_definition.0.request.0.query.0.limit = 30",
	"widget.0.slo_list_definition.0.title = my title",
	"widget.0.slo_list_definition.0.title_size = 16",
	"widget.0.slo_list_definition.0.title_align = center",
}

func TestAccDatadogDashboardSloList(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSloListConfig, "datadog_dashboard.slo_list_dashboard", datadogDashboardSloListAsserts)
}

func TestAccDatadogDashboardSloList_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSloListConfig, "datadog_dashboard.slo_list_dashboard")
}
