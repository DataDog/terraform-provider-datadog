package test

import (
	"testing"
)

const datadogDashboardChangeConfigImport = `
resource "datadog_dashboard" "change_dashboard" {
   	title         = "{{uniq}}"
   	description   = "Created using the Datadog provider in Terraform"
   	layout_type   = "ordered"
   	is_read_only  = true
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
			}
		}
	}

	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
				compare_to = "day_before"
				increase_good = "false"
				order_by = "change"
				change_type = "absolute"
				order_dir = "desc"
				show_present = "true"
			}
			title = "Sum of system.cpu.user over * by service,account"
			title_size = "16"
			title_align = "left"
			live_span = "1h"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

const datadogDashboardChangeConfig = `
resource "datadog_dashboard" "change_dashboard" {
   	title         = "{{uniq}}"
   	description   = "Created using the Datadog provider in Terraform"
   	layout_type   = "ordered"
   	is_read_only  = true
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
			}
		}
	}

	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
				compare_to = "day_before"
				increase_good = "false"
				order_by = "change"
				change_type = "absolute"
				order_dir = "desc"
				show_present = "true"
			}
			title = "Sum of system.cpu.user over * by service,account"
			title_size = "16"
			title_align = "left"
			live_span = "1h"
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
	widget {
		change_definition {
			request {
				q = "sum:system.cpu.user{*} by {service,account}"
				compare_to = "day_before"
				increase_good = "false"
				order_by = "change"
				change_type = "absolute"
				order_dir = "desc"
				show_present = "true"
			}
			title = "Sum of system.cpu.user over * by service,account"
			title_size = "16"
			title_align = "left"
			time = {
				live_span = "1h"
			}
			custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
			}
		}
	}
}
`

var datadogDashboardChangeAsserts = []string{
	"widget.0.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.title_align = left",
	"widget.1.change_definition.0.request.0.change_type = absolute",
	"widget.0.change_definition.0.request.0.order_dir =",
	"widget.0.change_definition.0.title_size =",
	"title = {{uniq}}",
	"widget.0.change_definition.0.request.0.change_type =",
	"widget.1.change_definition.0.title = Sum of system.cpu.user over * by service,account",
	"widget.1.change_definition.0.title_size = 16",
	"widget.1.change_definition.0.request.0.compare_to = day_before",
	"is_read_only = true",
	"widget.0.change_definition.0.title_align =",
	"widget.0.change_definition.0.title =",
	"widget.1.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.request.0.show_present = true",
	"widget.1.change_definition.0.request.0.order_by = change",
	"layout_type = ordered",
	"widget.1.change_definition.0.request.0.order_dir = desc",
	"widget.0.change_definition.0.request.0.increase_good = false",
	"widget.1.change_definition.0.request.0.increase_good = false",
	"widget.0.change_definition.0.request.0.show_present = false",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.change_definition.0.request.0.order_by =",
	"widget.1.change_definition.0.live_span = 1h",
	"widget.0.change_definition.0.request.0.compare_to =",
	"widget.1.change_definition.0.custom_link.# = 1",
	"widget.1.change_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.change_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	// Deprecated widget
	"widget.2.change_definition.0.title_align = left",
	"widget.2.change_definition.0.request.0.change_type = absolute",
	"widget.2.change_definition.0.title = Sum of system.cpu.user over * by service,account",
	"widget.2.change_definition.0.title_size = 16",
	"widget.2.change_definition.0.request.0.compare_to = day_before",
	"widget.2.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.2.change_definition.0.request.0.show_present = true",
	"widget.2.change_definition.0.request.0.order_by = change",
	"widget.2.change_definition.0.request.0.order_dir = desc",
	"widget.2.change_definition.0.request.0.increase_good = false",
	"widget.2.change_definition.0.time.live_span = 1h",
	"widget.2.change_definition.0.custom_link.# = 1",
	"widget.2.change_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.2.change_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardChange(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardChangeConfig, "datadog_dashboard.change_dashboard", datadogDashboardChangeAsserts)
}

func TestAccDatadogDashboardChange_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardChangeConfigImport, "datadog_dashboard.change_dashboard")
}
