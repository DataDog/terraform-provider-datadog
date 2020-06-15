package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardAlertValueConfig = `
resource "datadog_dashboard" "alert_value_dashboard" {
    title         = "Acceptance Test Alert Value Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = true
    widget {
		alert_value_definition {
			alert_id = "895605"
		}
    }
    widget {
		alert_value_definition {
			alert_id = "895606"
			precision = 1
			unit = "b"
            title_size = "16"
			title_align = "center"
			title = "Widget Title"
			text_align = "center"
		}
    }
}
`

func TestAccDatadogDashboardAlertValue(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardAlertValueConfig,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists(accProvider),
					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "title", "Acceptance Test Alert Value Widget Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "is_read_only", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.#", "2"),
					// Alert Graph widget
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.0.alert_value_definition.0.alert_id", "895605"),

					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.1.alert_value_definition.0.alert_id", "895606"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.1.alert_value_definition.0.precision", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.1.alert_value_definition.0.unit", "b"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.1.alert_value_definition.0.title_size", "16"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.1.alert_value_definition.0.title", "Widget Title"),
					resource.TestCheckResourceAttr("datadog_dashboard.alert_value_dashboard", "widget.1.alert_value_definition.0.text_align", "center"),
				),
			},
		},
	})
}
