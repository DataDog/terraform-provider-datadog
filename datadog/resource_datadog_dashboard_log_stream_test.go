package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardLogStreamConfig = `
resource "datadog_dashboard" "log_stream_dashboard" {
    title         = "Acceptance Test Log Stream Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "free"
    is_read_only  = "true"

    widget {
		log_stream_definition {
			title = "Log Stream"
            title_align = "right"
			title_size = "16"
			show_message_column = "true"
			message_display = "expanded-md"
			query = "status:error env:prod"
			show_date_column = "true"
			indexes = []
			columns = ["core_host", "core_service"]
			time = {
				live_span = "1d"
			}
			sort {
				column = "time"
				order = "desc"
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

var datadogDashboardLogStreamAsserts = []string{
	"description = Created using the Datadog provider in Terraform",
	"widget.0.log_stream_definition.0.query = status:error env:prod",
	"widget.0.log_stream_definition.0.title_align = right",
	"widget.0.log_stream_definition.0.show_date_column = true",
	"widget.0.log_stream_definition.0.columns.0 = core_host",
	"layout_type = free",
	"widget.0.log_stream_definition.0.show_message_column = true",
	"widget.0.log_stream_definition.0.time.live_span = 1d",
	"widget.0.layout.width = 32",
	"widget.0.layout.x = 5",
	"is_read_only = true",
	"widget.0.log_stream_definition.0.message_display = expanded-md",
	"widget.0.layout.height = 43",
	"title = Acceptance Test Log Stream Widget Dashboard",
	"widget.0.log_stream_definition.0.columns.1 = core_service",
	"widget.0.log_stream_definition.0.title_size = 16",
	"widget.0.log_stream_definition.0.logset =",
	"widget.0.layout.y = 5",
	"widget.0.log_stream_definition.0.sort.0.column = time",
	"widget.0.log_stream_definition.0.title = Log Stream",
	"widget.0.log_stream_definition.0.sort.0.order = desc",
}

func TestAccDatadogDashboardLogStream(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardLogStreamConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.log_stream_dashboard", checkDashboardExists(accProvider), datadogDashboardLogStreamAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardLogStream_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardLogStreamConfig,
			},
			{
				ResourceName:      "datadog_dashboard.log_stream_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
