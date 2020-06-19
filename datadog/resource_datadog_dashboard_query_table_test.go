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
//    "title": "TF - Query Table Example",
//    "created_at": "2020-06-09T11:53:33.269271+00:00",
//    "modified_at": "2020-06-09T11:57:11.580865+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "system.cpu.user, system.load.1",
//                "title_align": "right",
//                "time": {
//                    "live_span": "1d"
//                },
//                "requests": [
//                    {
//                        "aggregator": "max",
//                        "conditional_formats": [
//                            {
//                                "palette": "white_on_green",
//                                "value": 90,
//                                "comparator": "<"
//                            },
//                            {
//                                "palette": "white_on_red",
//                                "value": 90,
//                                "comparator": ">="
//                            }
//                        ],
//                        "q": "avg:system.cpu.user{account:prod} by {service, team}",
//                        "alias": "cpu user",
//                        "limit": 25,
//                        "order": "desc"
//                    },
//                    {
//                        "q": "avg:system.load.1{*} by {service, team}",
//                        "aggregator": "last",
//                        "conditional_formats": [
//                            {
//                                "palette": "custom_bg",
//                                "value": 50,
//                                "comparator": ">"
//                            }
//                        ],
//                        "alias": "system load"
//                    }
//                ],
//                "type": "query_table"
//            },
//            "layout": {
//                "y": 1,
//                "x": 1,
//                "height": 32,
//                "width": 54
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardQueryTableConfig = `
resource "datadog_dashboard" "query_table_dashboard" {
	title         = "Acceptance Test Query Table Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		query_table_definition {
			title_size = "16"
			title = "system.cpu.user, system.load.1"
			title_align = "right"
			time = {
				live_span = "1d"
			}
			request {
				aggregator = "max"
				conditional_formats {
					palette = "white_on_green"
					value = 90
					comparator = "<"
				}
				conditional_formats {
					palette = "white_on_red"
					value = 90
					comparator = ">="
				}
				q = "avg:system.cpu.user{account:prod} by {service, team}"
				alias = "cpu user"
				limit = 25
				order = "desc"
			}
			request {
				q = "avg:system.load.1{*} by {service, team}"
				aggregator = "last"
				conditional_formats {
					palette = "custom_bg"
					value = 50
					comparator = ">"
				}
				alias = "system load"
			}
		}
	}
}
`

var datadogDashboardQueryTableAsserts = []string{
	"widget.0.query_table_definition.0.time.live_span = 1d",
	"widget.0.query_table_definition.0.request.1.order =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.timeframe =",
	"widget.0.query_table_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service, team}",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.custom_fg_color =",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.comparator = >",
	"widget.0.query_table_definition.0.title_size = 16",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.comparator = <",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.value = 90",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.image_url =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.hide_value = false",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.timeframe =",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.palette = white_on_green",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.q = avg:system.load.1{*} by {service, team}",
	"title = Acceptance Test Query Table Widget Dashboard",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.value = 90",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.aggregator = last",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.custom_fg_color =",
	"widget.0.query_table_definition.0.request.1.limit = 0",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.hide_value = false",
	"widget.0.query_table_definition.0.request.0.aggregator = max",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.palette = custom_bg",
	"widget.0.query_table_definition.0.request.1.alias = system load",
	"widget.0.query_table_definition.0.request.0.order = desc",
	"widget.0.query_table_definition.0.request.0.conditional_formats.0.image_url =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.comparator = >=",
	"widget.0.query_table_definition.0.request.0.alias = cpu user",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.palette = white_on_red",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.value = 50",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.custom_bg_color =",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.image_url =",
	"widget.0.query_table_definition.0.request.1.conditional_formats.0.hide_value = false",
	"is_read_only = true",
	"widget.0.query_table_definition.0.request.0.limit = 25",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.timeframe =",
	"widget.0.query_table_definition.0.request.0.conditional_formats.1.custom_fg_color =",
	"layout_type = ordered",
	"widget.0.query_table_definition.0.title = system.cpu.user, system.load.1",
	"widget.0.query_table_definition.0.title_align = right",
}

func TestAccDatadogDashboardQueryTable(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardQueryTableConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.query_table_dashboard", checkDashboardExists(accProvider), datadogDashboardQueryTableAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardQueryTable_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardQueryTableConfig,
			},
			{
				ResourceName:      "datadog_dashboard.query_table_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
