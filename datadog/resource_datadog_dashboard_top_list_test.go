package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardTopListConfig = `
resource "datadog_dashboard" "top_list_dashboard" {
    title         = "Acceptance Test Top List Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = "true"

    widget {
		toplist_definition {
			title_size = "16"
			title = "Avg of system.core.user over account:prod by service,app"
			title_align = "right"
			time = {
				live_span = "1w"
			}
			request {
				q = "top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')"
				conditional_formats {
					palette = "white_on_red"
					value = 15000
					comparator = ">"
				}
			}
		}
    }
}
`

var datadogDashboardTopListAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.image_url =",
	"layout_type = ordered",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.comparator = >",
	"title = Acceptance Test Top List Widget Dashboard",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"is_read_only = true",
	"widget.0.toplist_definition.0.time.live_span = 1w",
	"widget.0.toplist_definition.0.time.% = 1",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.toplist_definition.0.request.0.q = top(avg:system.core.user{account:prod} by {service,app}, 10, 'sum', 'desc')",
	"widget.0.toplist_definition.0.title_size = 16",
	"widget.0.toplist_definition.0.title_align = right",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.value = 15000",
	"widget.0.toplist_definition.0.title = Avg of system.core.user over account:prod by service,app",
	"widget.0.toplist_definition.0.request.0.conditional_formats.0.custom_fg_color =",
}

func TestAccDatadogDashboardTopList(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardTopListConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.top_list_dashboard", checkDashboardExists(accProvider), datadogDashboardTopListAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardTopList_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardTopListConfig,
			},
			{
				ResourceName:      "datadog_dashboard.top_list_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
