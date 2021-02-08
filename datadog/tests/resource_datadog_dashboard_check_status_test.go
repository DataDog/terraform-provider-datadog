package test

import (
	"testing"
)

const datadogDashboardCheckStatusConfig = `
resource "datadog_dashboard" "check_status_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		check_status_definition {
			title = "Agent Up"
			title_align = "center"
			title_size = "16"
			group_by = ["app"]
			check = "aws.ec2.host_status"
			tags = ["account:prod"]
			grouping = "cluster"
		}
	}
}
`

var datadogDashboardCheckStatusAsserts = []string{
	"widget.0.check_status_definition.0.group_by.0 = app",
	"widget.0.check_status_definition.0.title_size = 16",
	"widget.0.check_status_definition.0.check = aws.ec2.host_status",
	"is_read_only = true",
	"widget.0.check_status_definition.0.title = Agent Up",
	"widget.0.check_status_definition.0.title_align = center",
	"widget.0.check_status_definition.0.group =",
	"title = {{uniq}}",
	"widget.0.check_status_definition.0.grouping = cluster",
	"widget.0.check_status_definition.0.tags.0 = account:prod",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
}

func TestAccDatadogDashboardCheckStatus(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardCheckStatusConfig, "datadog_dashboard.check_status_dashboard", datadogDashboardCheckStatusAsserts)
}

func TestAccDatadogDashboardCheckStatus_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardCheckStatusConfig, "datadog_dashboard.check_status_dashboard")
}
