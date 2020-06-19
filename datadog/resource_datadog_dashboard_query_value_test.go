package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": null,
//    "author_name": "--redacted--",
//    "id": "--redacted--",
//    "url": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "title": "TF - Query Value example",
//    "created_at": "2020-06-09T11:41:15.788029+00:00",
//    "modified_at": "2020-06-09T11:44:13.755796+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "Avg of system.mem.free over account:prod",
//                "title_align": "center",
//                "custom_unit": "Gib",
//                "precision": 3,
//                "time": {
//                    "live_span": "1h"
//                },
//                "autoscale": true,
//                "requests": [
//                    {
//                        "q": "avg:system.mem.free{account:prod}",
//                        "aggregator": "max",
//                        "conditional_formats": [
//                            {
//                                "palette": "white_on_red",
//                                "value": 9,
//                                "comparator": "<"
//                            },
//                            {
//                                "palette": "white_on_green",
//                                "value": 9,
//                                "comparator": ">="
//                            }
//                        ]
//                    }
//                ],
//                "type": "query_value"
//            },
//            "layout": {
//                "y": 2,
//                "x": 2,
//                "height": 15,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardQueryValueConfig = `
resource "datadog_dashboard" "query_value_dashboard" {
	title         = "Acceptance Test Query Value Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		query_value_definition {
			title = "Avg of system.mem.free over account:prod"
			title_align = "center"
			title_size = "16"
			custom_unit = "Gib"
			precision = "3"
			autoscale = "true"
			request {
				q = "avg:system.mem.free{account:prod}"
				aggregator = "max"
				conditional_formats {
					palette = "white_on_red"
					value = "9"
					comparator = "<"
				}
				conditional_formats {
					palette = "white_on_green"
					value = "9"
					comparator = ">="
				}
			}
			time = {
				live_span = "1h"
			}
		}
	}
}
`

var datadogDashboardQueryValueAsserts = []string{
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.query_value_definition.0.time.live_span = 1h",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.palette = white_on_red",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.image_url =",
	"widget.0.query_value_definition.0.precision = 3",
	"widget.0.query_value_definition.0.request.0.aggregator = max",
	"layout_type = ordered",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.palette = white_on_green",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.custom_fg_color =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.value = 9",
	"widget.0.query_value_definition.0.autoscale = true",
	"widget.0.query_value_definition.0.request.0.q = avg:system.mem.free{account:prod}",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.comparator = >=",
	"widget.0.query_value_definition.0.title_size = 16",
	"widget.0.query_value_definition.0.custom_unit = Gib",
	"widget.0.query_value_definition.0.title_align = center",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.value = 9",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.image_url =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.hide_value = false",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.timeframe =",
	"widget.0.query_value_definition.0.text_align =",
	"widget.0.query_value_definition.0.title = Avg of system.mem.free over account:prod",
	"widget.0.query_value_definition.0.request.0.conditional_formats.1.custom_bg_color =",
	"widget.0.query_value_definition.0.request.0.conditional_formats.# = 2",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.query_value_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"is_read_only = true",
}

func TestAccDatadogDashboardQueryValue(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardQueryValueConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.query_value_dashboard", checkDashboardExists(accProvider), datadogDashboardQueryValueAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardQueryValue_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardQueryValueConfig,
			},
			{
				ResourceName:      "datadog_dashboard.query_value_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
