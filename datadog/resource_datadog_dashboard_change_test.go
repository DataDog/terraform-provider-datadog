package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardChangeConfig = `
resource "datadog_dashboard" "change_dashboard" {
   title         = "Acceptance Test Change Widget Dashboard"
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
			time = {
				live_span = "1h"
			}
		}
	}
}
`

func TestAccDatadogDashboardChange(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardChangeConfig,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "title", "Acceptance Test Change Widget Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "is_read_only", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.#", "2"),
					// Change widget
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.0.change_definition.0.request.0.q", "sum:system.cpu.user{*} by {service,account}"),

					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.q", "sum:system.cpu.user{*} by {service,account}"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.compare_to", "day_before"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.increase_good", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.order_by", "change"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.change_type", "absolute"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.order_dir", "desc"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.request.0.show_present", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.title", "Sum of system.cpu.user over * by service,account"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.title_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.title_align", "left"),
					resource.TestCheckResourceAttr("datadog_dashboard.change_dashboard", "widget.1.change_definition.0.time.live_span", "1h"),
				),
			},
		},
	})
}
