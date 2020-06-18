package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardEventTimelineConfig = `
resource "datadog_dashboard" "event_timeline_dashboard" {
    title         = "Acceptance Test Event Timeline Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "free"
    is_read_only  = "true"
    
    widget {
		event_timeline_definition {
			title = "Widget Title"
            title_align = "right"
			title_size = "16"
			tags_execution = "and"
            query = "status:error"
			time = {
				live_span = "1h"
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

var datadogDashboardEventTimelineAsserts = []string{
	"widget.0.layout.y = 5",
	"widget.0.event_timeline_definition.0.title_align = right",
	"widget.0.layout.x = 5",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.event_timeline_definition.0.time.live_span = 1h",
	"title = Acceptance Test Event Timeline Widget Dashboard",
	"is_read_only = true",
	"widget.0.layout.width = 32",
	"widget.0.event_timeline_definition.0.title_size = 16",
	"layout_type = free",
	"widget.0.event_timeline_definition.0.query = status:error",
	"widget.0.event_timeline_definition.0.title = Widget Title",
	"widget.0.event_timeline_definition.0.tags_execution = and",
	"widget.0.layout.height = 43",
}

func TestAccDatadogDashboardEventTimeline(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardEventTimelineConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.event_timeline_dashboard", checkDashboardExists(accProvider), datadogDashboardEventTimelineAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardEventTimeline_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardEventTimelineConfig,
			},
			{
				ResourceName:      "datadog_dashboard.event_timeline_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
