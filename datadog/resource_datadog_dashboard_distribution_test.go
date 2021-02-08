package datadog

import (
	"testing"
)

const datadogDashboardDistributionConfig = `
resource "datadog_dashboard" "distribution_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	
	widget {
		distribution_definition {
			title = "Avg of system.cpu.user over account:prod by service,account"
			title_align = "left"
			title_size = "16"
			show_legend = "true"
			legend_size = "2"
			live_span = "1h"
			request {
				q = "avg:system.cpu.user{account:prod} by {service,account}"
				style {
					palette = "purple"
				}
			}
		}
	}
	widget {
		distribution_definition {
			title = "Avg of system.cpu.user over account:prod by service,account"
			title_align = "left"
			title_size = "16"
			show_legend = "true"
			legend_size = "2"
			time = {
				live_span = "1h"
			}
			request {
				q = "avg:system.cpu.user{account:prod} by {service,account}"
				style {
					palette = "purple"
				}
			}
		}
	}
}
`

const datadogDashboardDistributionConfigImport = `
resource "datadog_dashboard" "distribution_dashboard" {
	title         = "{{uniq}}"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"
	
	widget {
		distribution_definition {
			title = "Avg of system.cpu.user over account:prod by service,account"
			title_align = "left"
			title_size = "16"
			show_legend = "true"
			legend_size = "2"
			live_span = "1h"
			request {
				q = "avg:system.cpu.user{account:prod} by {service,account}"
				style {
					palette = "purple"
				}
			}
		}
	}
}
`

var datadogDashboardDistributionAsserts = []string{
	"title = {{uniq}}",
	"widget.0.distribution_definition.0.live_span = 1h",
	"widget.0.distribution_definition.0.title = Avg of system.cpu.user over account:prod by service,account",
	"widget.0.distribution_definition.0.title_size = 16",
	"widget.0.distribution_definition.0.title_align = left",
	"widget.0.distribution_definition.0.show_legend = true",
	"widget.0.distribution_definition.0.legend_size = 2",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.distribution_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service,account}",
	"widget.0.distribution_definition.0.request.0.style.0.palette = purple",
	"layout_type = ordered",
	"is_read_only = true",
	// Deprecated widget
	"widget.1.distribution_definition.0.time.live_span = 1h",
	"widget.1.distribution_definition.0.title = Avg of system.cpu.user over account:prod by service,account",
	"widget.1.distribution_definition.0.title_size = 16",
	"widget.1.distribution_definition.0.title_align = left",
	"widget.1.distribution_definition.0.show_legend = true",
	"widget.1.distribution_definition.0.legend_size = 2",
	"widget.1.distribution_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service,account}",
	"widget.1.distribution_definition.0.request.0.style.0.palette = purple",
}

func TestAccDatadogDashboardDistribution(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardDistributionConfig, "datadog_dashboard.distribution_dashboard", datadogDashboardDistributionAsserts)
}

func TestAccDatadogDashboardDistribution_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardDistributionConfigImport, "datadog_dashboard.distribution_dashboard")
}
