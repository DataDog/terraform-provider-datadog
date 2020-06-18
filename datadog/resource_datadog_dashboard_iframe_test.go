package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardIFrameConfig = `
resource "datadog_dashboard" "iframe_dashboard" {
    title         = "Acceptance Test IFrame Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "free"
    is_read_only  = "true"

    widget {
		iframe_definition {
			url = "https://en.wikipedia.org/wiki/Datadog"
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

var datadogDashboardIFrameAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.iframe_definition.0.url = https://en.wikipedia.org/wiki/Datadog",
	"widget.0.layout.height = 43",
	"title = Acceptance Test IFrame Widget Dashboard",
	"widget.0.layout.x = 5",
	"widget.0.layout.y = 5",
	"layout_type = free",
	"widget.0.layout.width = 32",
}

func TestAccDatadogDashboardIFrame(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardIFrameConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.iframe_dashboard", checkDashboardExists(accProvider), datadogDashboardIFrameAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardIFrame_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardIFrameConfig,
			},
			{
				ResourceName:      "datadog_dashboard.iframe_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
