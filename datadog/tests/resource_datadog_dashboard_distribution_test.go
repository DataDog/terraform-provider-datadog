package test

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
}
`

const datadogDashboardDistributionApmStatsQueryConfig = `
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
				apm_stats_query {
					service = "service"
					env = "env"
					primary_tag = "tag:*"
					name = "name"
					row_type = "resource"
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
}

var datadogDashboardDistributionApmStatsQueryAsserts = []string{
	"title = {{uniq}}",
	"widget.0.distribution_definition.0.request.0.apm_stats_query.0.service = service",
	"widget.0.distribution_definition.0.request.0.apm_stats_query.0.env = env",
	"widget.0.distribution_definition.0.request.0.apm_stats_query.0.primary_tag = tag:*",
	"widget.0.distribution_definition.0.request.0.apm_stats_query.0.name = name",
	"widget.0.distribution_definition.0.request.0.apm_stats_query.0.row_type = resource",
}

func TestAccDatadogDashboardDistribution(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardDistributionConfig, "datadog_dashboard.distribution_dashboard", datadogDashboardDistributionAsserts)
}

func TestAccDatadogDashboardApmStatsQueryDistribution(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardDistributionApmStatsQueryConfig, "datadog_dashboard.distribution_dashboard", datadogDashboardDistributionApmStatsQueryAsserts)
}

func TestAccDatadogDashboardDistribution_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardDistributionConfigImport, "datadog_dashboard.distribution_dashboard")
}
