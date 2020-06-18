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

var datadogDashboardAlertValueAsserts = []string{
	"widget.0.alert_value_definition.0.title_align =",
	"widget.1.alert_value_definition.0.title_align = center",
	"widget.1.alert_value_definition.0.text_align = center",
	"widget.1.layout.% = 0",
	"title = Acceptance Test Alert Value Widget Dashboard",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.alert_value_definition.0.precision = 0",
	"widget.1.alert_value_definition.0.title_size = 16",
	"widget.1.alert_value_definition.0.precision = 1",
	"widget.0.alert_value_definition.0.title_size =",
	"widget.1.alert_value_definition.0.alert_id = 895606",
	"widget.0.alert_value_definition.0.text_align =",
	"layout_type = ordered",
	"widget.0.alert_value_definition.0.title =",
	"widget.0.alert_value_definition.0.unit =",
	"widget.1.alert_value_definition.0.title = Widget Title",
	"widget.0.alert_value_definition.0.alert_id = 895605",
	"widget.1.alert_value_definition.0.unit = b",
	"is_read_only = true",
}

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
					testCheckResourceAttrs("datadog_dashboard.alert_value_dashboard", checkDashboardExists(accProvider), datadogDashboardAlertValueAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardAlertValue_import(t *testing.T) {
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
			},
			{
				ResourceName:      "datadog_dashboard.alert_value_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
