package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": "",
//    "author_name": "--redacted--",
//    "id": "--redacted--",
//    "url": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "title": "TF - Distribution Example",
//    "created_at": "2020-06-09T13:22:05.545823+00:00",
//    "modified_at": "2020-06-09T13:22:57.196703+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "13",
//                "title": "Avg of system.cpu.user over account:prod by service,account",
//                "title_align": "center",
//                "time": {
//                    "live_span": "1d"
//                },
//                "requests": [
//                    {
//                        "q": "avg:system.cpu.user{account:prod} by {service,account}",
//                        "style": {
//                            "palette": "purple"
//                        }
//                    }
//                ],
//                "type": "distribution",
//                "show_legend": true,
//                "legend_size": "2"
//            },
//            "layout": {
//                "y": 3,
//                "x": 2,
//                "height": 15,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

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
	"title = Acceptance Test Distribution Widget Dashboard",
	"widget.0.distribution_definition.0.time.live_span = 1h",
	"widget.0.distribution_definition.0.title = Avg of system.cpu.user over account:prod by service,account",
	"widget.0.distribution_definition.0.title_size = 16",
	"widget.0.distribution_definition.0.title_align = left",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.distribution_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service,account}",
	"widget.0.distribution_definition.0.request.0.style.0.palette = purple",
	"layout_type = ordered",
	"is_read_only = true",
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
				Config: datadogDashboardDistributionConfig,
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
