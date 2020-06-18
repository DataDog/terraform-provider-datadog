package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardDistributionConfig = `
resource "datadog_dashboard" "distribution_dashboard" {
    title         = "Acceptance Test Distribution Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = "true"
    
    widget {
		distribution_definition {
			title = "Avg of system.cpu.user over account:prod by service,account"
            title_align = "left"
			title_size = "16"
			//show_legend = "true"
			//legend_size = "2"
			time = {
				live_span = "1h"
			}
			request {
				q = "avg:system.cpu.user{account:prod} by {service,account}"
                style {
					palette = "purple"
				}
			}
		}
    }
}
`

var datadogDashboardDistributionAsserts = []string{
	"title = Acceptance Test Alert Graph Widget Dashboard",
}

func TestAccDatadogDashboardDistribution(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardAlertGraphConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.distribution_dashboard", checkDashboardExists(accProvider), datadogDashboardDistributionAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardDistribution_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardDistributionConfig,
			},
			{
				ResourceName:      "datadog_dashboard.distribution_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
