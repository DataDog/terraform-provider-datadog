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
//    "title": "TF - Image Example",
//    "created_at": "2020-06-09T13:35:49.700883+00:00",
//    "modified_at": "2020-06-09T13:36:10.777106+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "url": "https://i.picsum.photos/id/826/200/300.jpg",
//                "sizing": "fit",
//                "margin": "small",
//                "type": "image"
//            },
//            "layout": {
//                "y": 2,
//                "x": 8,
//                "height": 12,
//                "width": 12
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardImageConfig = `
resource "datadog_dashboard" "image_dashboard" {
	title         = "Acceptance Test Image Widget Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = "true"

	widget {
		image_definition {
			url = "https://i.picsum.photos/id/826/200/300.jpg"
			sizing = "fit"
			margin = "small"
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

var datadogDashboardImageAsserts = []string{
	"widget.0.image_definition.0.sizing = fit",
	"title = Acceptance Test Image Widget Dashboard",
	"widget.0.layout.y = 5",
	"widget.0.layout.x = 5",
	"widget.0.image_definition.0.margin = small",
	"widget.0.layout.height = 43",
	"layout_type = free",
	"widget.0.layout.width = 32",
	"is_read_only = true",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.image_definition.0.url = https://i.picsum.photos/id/826/200/300.jpg",
}

func TestAccDatadogDashboardImage(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardImageConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.image_dashboard", checkDashboardExists(accProvider), datadogDashboardImageAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardImage_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardImageConfig,
			},
			{
				ResourceName:      "datadog_dashboard.image_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
