package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardFreeTextConfig = `
resource "datadog_dashboard" "free_text_dashboard" {
    title         = "Acceptance Test Free Text Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "free"
    is_read_only  = "true"
    
    widget {
		free_text_definition {
			color = "#eb364b"
			text = "Free Text"
            font_size = "56"
            text_align = "left"
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

var datadogDashboardFreeTextAsserts = []string{
	"widget.0.layout.y = 5",
	"widget.0.free_text_definition.0.text = Free Text",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.free_text_definition.0.font_size = 56",
	"is_read_only = true",
	"widget.0.free_text_definition.0.color = #eb364b",
	"widget.0.layout.width = 32",
	"widget.0.layout.height = 43",
	"widget.0.free_text_definition.0.text_align = left",
	"title = Acceptance Test Free Text Widget Dashboard",
	"widget.0.layout.x = 5",
}

func TestAccDatadogDashboardFreeText(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardFreeTextConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.free_text_dashboard", checkDashboardExists(accProvider), datadogDashboardFreeTextAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardFreeText_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardFreeTextConfig,
			},
			{
				ResourceName:      "datadog_dashboard.free_text_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
