package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardNoteConfig = `
resource "datadog_dashboard" "note_dashboard" {
    title         = "Acceptance Test Notes Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = "true"

    widget {
		note_definition {
			tick_pos= "50%"
			show_tick = true
			tick_edge = "bottom"
			text_align = "center"
			content = "This is a note widget"
			font_size = "18"
			background_color = "green"
		}
    }
}
`

var datadogDashboardNoteAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.note_definition.0.content = This is a note widget",
	"title = Acceptance Test Notes Widget Dashboard",
	"widget.0.note_definition.0.font_size = 18",
	"widget.0.note_definition.0.text_align = center",
	"widget.0.note_definition.0.show_tick = true",
	"widget.0.note_definition.0.tick_edge = bottom",
	"layout_type = ordered",
	"is_read_only = true",
	"widget.0.note_definition.0.tick_pos = 50%",
	"widget.0.note_definition.0.background_color = green",
}

func TestAccDatadogDashboardNote(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardNoteConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.note_dashboard", checkDashboardExists(accProvider), datadogDashboardNoteAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardNote_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardNoteConfig,
			},
			{
				ResourceName:      "datadog_dashboard.note_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
