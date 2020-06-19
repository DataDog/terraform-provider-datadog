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
//    "title": "TF - Check Status Example",
//    "created_at": "2020-06-09T12:59:18.872351+00:00",
//    "modified_at": "2020-06-09T13:00:12.684566+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "title": "Agent Up",
//                "title_align": "center",
//                "group_by": [
//                    "app"
//                ],
//                "type": "check_status",
//                "check": "aws.ec2.host_status",
//                "tags": [
//                    "account:prod"
//                ],
//                "grouping": "cluster"
//            },
//            "layout": {
//                "y": 1,
//                "x": 7,
//                "height": 8,
//                "width": 15
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardCheckStatusConfig = `
resource "datadog_dashboard" "check_status_dashboard" {
	title         = "Acceptance Test Check Status Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		check_status_definition {
			title = "Agent Up"
			title_align = "center"
			title_size = "16"
			group_by = ["app"]
			check = "aws.ec2.host_status"
			tags = ["account:prod"]
			grouping = "cluster"
		}
	}
}
`

var datadogDashboardCheckStatusAsserts = []string{
	"widget.0.check_status_definition.0.group_by.0 = app",
	"widget.0.check_status_definition.0.title_size = 16",
	"widget.0.check_status_definition.0.check = aws.ec2.host_status",
	"is_read_only = true",
	"widget.0.check_status_definition.0.title = Agent Up",
	"widget.0.check_status_definition.0.title_align = center",
	"widget.0.check_status_definition.0.group =",
	"title = Acceptance Test Check Status Widget Dashboard",
	"widget.0.check_status_definition.0.grouping = cluster",
	"widget.0.check_status_definition.0.tags.0 = account:prod",
	"layout_type = ordered",
	"description = Created using the Datadog provider in Terraform",
}

func TestAccDatadogDashboardCheckStatus(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardCheckStatusConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.check_status_dashboard", checkDashboardExists(accProvider), datadogDashboardCheckStatusAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardCheckStatus_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardCheckStatusConfig,
			},
			{
				ResourceName:      "datadog_dashboard.check_status_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
