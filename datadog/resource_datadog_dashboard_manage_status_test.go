package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardManageStatusConfig = `
resource "datadog_dashboard" "manage_status_dashboard" {
    title         = "Acceptance Test Manage Status Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "free"
    is_read_only  = "true"

    widget {
		manage_status_definition {
			sort = "triggered,desc"
			count = "50"
			title_size = "20"
			title = ""
			title_align = "center"
			hide_zero_counts = true
			start = "0"
			summary_type = "combined"
			color_preference = "background"
			query = "env:prod group_status:alert"
			show_last_triggered = true
			display_format = "countsAndList"
		}
		layout = {
			height = 43
			width = 32
			x = 5
			y = 5
		}
    }
}
`

var datadogDashboardManageStatusAsserts = []string{
	"widget.0.manage_status_definition.0.color_preference = background",
	"widget.0.layout.x = 5",
	"widget.0.layout.height = 43",
	"layout_type = free",
	"is_read_only = true",
	"widget.0.manage_status_definition.0.display_format = countsAndList",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.layout.width = 32",
	"widget.0.manage_status_definition.0.hide_zero_counts = true",
	"widget.0.layout.y = 5",
	"widget.0.manage_status_definition.0.summary_type = combined",
	"widget.0.manage_status_definition.0.show_last_triggered = true",
	"widget.0.manage_status_definition.0.title =",
	"widget.0.manage_status_definition.0.title_size = 20",
	"widget.0.manage_status_definition.0.sort = triggered,desc",
	"widget.0.manage_status_definition.0.title_align = center",
	"widget.0.manage_status_definition.0.start = 0",
	"widget.0.manage_status_definition.0.count = 50",
	"title = Acceptance Test Manage Status Widget Dashboard",
	"widget.0.manage_status_definition.0.query = env:prod group_status:alert",
}

func TestAccDatadogDashboardManageStatus(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardManageStatusConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.manage_status_dashboard", checkDashboardExists(accProvider), datadogDashboardManageStatusAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardManageStatus_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardManageStatusConfig,
			},
			{
				ResourceName:      "datadog_dashboard.manage_status_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
