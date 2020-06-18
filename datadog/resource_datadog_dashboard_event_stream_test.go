package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardEventStreamConfig = `
resource "datadog_dashboard" "event_stream_dashboard" {
    title         = "Acceptance Test Event Stream Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "free"
    is_read_only  = "true"
    
    widget {
		event_stream_definition {
			title = "Widget Title"
            title_align = "right"
			title_size = "16"
			tags_execution = "and"
			query = "*"
			event_size = "l"
			time = {
				live_span = "4h"
			}
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

var datadogDashboardEventStreamAsserts = []string{
	"widget.0.layout.x = 5",
	"widget.0.event_stream_definition.0.title_size = 16",
	"widget.0.event_stream_definition.0.tags_execution = and",
	"title = Acceptance Test Event Stream Widget Dashboard",
	"widget.0.layout.y = 5",
	"widget.0.event_stream_definition.0.title_align = right",
	"widget.0.event_stream_definition.0.time.live_span = 4h",
	"widget.0.layout.width = 32",
	"widget.0.event_stream_definition.0.event_size = l",
	"layout_type = free",
	"description = Created using the Datadog provider in Terraform",
	"is_read_only = true",
	"widget.0.event_stream_definition.0.query = *",
	"widget.0.event_stream_definition.0.title = Widget Title",
	"widget.0.layout.height = 43",
}

func TestAccDatadogDashboardEventStream(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardEventStreamConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.event_stream_dashboard", checkDashboardExists(accProvider), datadogDashboardEventStreamAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardEventStream_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardEventStreamConfig,
			},
			{
				ResourceName:      "datadog_dashboard.event_stream_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
