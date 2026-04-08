package test

import (
	"testing"
)

// datadogDashboardV2ToplistDisplayConfig tests the TypeOneOf display block for toplist.
// This uses the new v2-only HCL syntax:
//
//	display { stacked { legend = "automatic" } }
//	display { flat {} }
const datadogDashboardV2ToplistDisplayConfig = `
resource "datadog_dashboard_v2" "toplist_display_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"

    widget {
        toplist_definition {
            title      = "Toplist with flat style"
            title_size = "16"
            title_align = "right"
            live_span  = "1w"
            request {
                q = "top(avg:system.cpu.user{*} by {service,app_id}, 10, 'sum', 'desc')"
                conditional_formats {
                    comparator = ">"
                    value      = 15000
                    palette    = "white_on_red"
                }
            }
            custom_link {
                link  = "https://app.datadoghq.com/dashboard/lists"
                label = "Test Custom Link label"
            }
            custom_link {
                link           = "https://app.datadoghq.com/dashboard/lists"
                is_hidden      = true
                override_label = "logs"
            }
            style {
                display {
                    flat {}
                }
            }
        }
    }

    widget {
        toplist_definition {
            title       = "Toplist with stacked style"
            title_size  = "16"
            title_align = "right"
            live_span   = "1w"
            request {
                q = "top(avg:system.cpu.user{*} by {service,app_id}, 10, 'sum', 'desc')"
                conditional_formats {
                    comparator = ">"
                    value      = 15000
                    palette    = "white_on_red"
                }
            }
            custom_link {
                link  = "https://app.datadoghq.com/dashboard/lists"
                label = "Test Custom Link label"
            }
            custom_link {
                link           = "https://app.datadoghq.com/dashboard/lists"
                is_hidden      = true
                override_label = "logs"
            }
            style {
                display {
                    stacked {
                        legend = "automatic"
                    }
                }
                palette = "datadog16"
                scaling = "relative"
            }
        }
    }
}
`

var datadogDashboardV2ToplistDisplayAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",

	"widget.0.toplist_definition.0.title = Toplist with flat style",
	"widget.0.toplist_definition.0.title_size = 16",
	"widget.0.toplist_definition.0.title_align = right",
	"widget.0.toplist_definition.0.live_span = 1w",
	"widget.0.toplist_definition.0.request.0.q = top(avg:system.cpu.user{*} by {service,app_id}, 10, 'sum', 'desc')",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.comparator = >",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.value = 15000",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.toplist_definition.0.custom_link.# = 2",
	"widget.0.toplist_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.toplist_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.toplist_definition.0.custom_link.1.override_label = logs",
	"widget.0.toplist_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.0.toplist_definition.0.custom_link.1.is_hidden = true",
	"widget.0.toplist_definition.0.style.# = 1",
	"widget.0.toplist_definition.0.style.0.display.# = 1",
	"widget.0.toplist_definition.0.style.0.display.0.flat.# = 1",

	"widget.1.toplist_definition.0.title = Toplist with stacked style",
	"widget.1.toplist_definition.0.title_size = 16",
	"widget.1.toplist_definition.0.title_align = right",
	"widget.1.toplist_definition.0.live_span = 1w",
	"widget.1.toplist_definition.0.request.0.q = top(avg:system.cpu.user{*} by {service,app_id}, 10, 'sum', 'desc')",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.comparator = >",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.value = 15000",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"widget.1.toplist_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.1.toplist_definition.0.custom_link.# = 2",
	"widget.1.toplist_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.toplist_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.toplist_definition.0.custom_link.1.override_label = logs",
	"widget.1.toplist_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.toplist_definition.0.custom_link.1.is_hidden = true",
	"widget.1.toplist_definition.0.style.# = 1",
	"widget.1.toplist_definition.0.style.0.display.# = 1",
	"widget.1.toplist_definition.0.style.0.display.0.stacked.# = 1",
	"widget.1.toplist_definition.0.style.0.display.0.stacked.0.legend = automatic",
	"widget.1.toplist_definition.0.style.0.palette = datadog16",
	"widget.1.toplist_definition.0.style.0.scaling = relative",
}

func TestAccDatadogDashboardV2ToplistDisplay(t *testing.T) {
	config, name := datadogDashboardV2ToplistDisplayConfig, "datadog_dashboard_v2.toplist_display_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2ToplistDisplay", config, name, datadogDashboardV2ToplistDisplayAsserts)
}

func TestAccDatadogDashboardV2ToplistDisplay_import(t *testing.T) {
	config, name := datadogDashboardV2ToplistDisplayConfig, "datadog_dashboard_v2.toplist_display_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2ToplistDisplay_import", config, name)
}
