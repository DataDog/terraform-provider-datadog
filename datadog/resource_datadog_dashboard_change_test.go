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

var datadogDashboardChangeAsserts = []string{
	"widget.0.change_definition.0.request.0.q = sum:system.cpu.user{*} by {service,account}",
	"widget.1.change_definition.0.title_align = left",
	"widget.1.change_definition.0.request.0.change_type = absolute",
	"widget.0.change_definition.0.request.0.order_dir =",
	"widget.0.change_definition.0.title_size =",
	"title = Acceptance Test Change Widget Dashboard",
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
	"widget.1.change_definition.0.time.live_span = 1h",
	"widget.0.change_definition.0.request.0.compare_to =",
}

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
					testCheckResourceAttrs("datadog_dashboard.change_dashboard", checkDashboardExists(accProvider), datadogDashboardChangeAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardChange_import(t *testing.T) {
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
			},
			{
				ResourceName:      "datadog_dashboard.change_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
