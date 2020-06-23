package datadog

import "testing"

// JSON export used as test scenario
//{
//    "notify_list": [],
//    "description": "",
//    "author_name": "--redacted--",
//    "id": "--redacted--",
//    "url": "--redacted--",
//    "template_variables": [],
//    "is_read_only": false,
//    "title": "TF - Service Map Example",
//    "created_at": "2020-06-09T13:32:03.535027+00:00",
//    "modified_at": "2020-06-09T13:32:50.224757+00:00",
//    "author_handle": "--redacted--",
//    "widgets": [
//        {
//            "definition": {
//                "title_size": "16",
//                "service": "master-db",
//                "title": "env: prod, datacenter:us1.prod.dog, service: master-db",
//                "title_align": "left",
//                "filters": [
//                    "env:prod",
//                    "datacenter:us1.prod.dog"
//                ],
//                "type": "servicemap"
//            },
//            "layout": {
//                "y": 3,
//                "x": -1,
//                "height": 15,
//                "width": 47
//            },
//            "id": 0
//        }
//    ],
//    "layout_type": "free"
//}

const datadogDashboardServiceMapConfig = `
resource "datadog_dashboard" "service_map_dashboard" {
  title         = "Acceptance Test Service Map Widget Dashboard"
  description   = "Created using the Datadog provider in Terraform"
  layout_type   = "free"
  is_read_only  = "true"

  widget {
		servicemap_definition {
			service = "master-db"
			filters = ["env:prod","datacenter:us1.prod.dog"]
			title = "env: prod, datacenter:us1.prod.dog, service: master-db"
			title_size = "16"
			title_align = "left"
			
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

var datadogDashboardServiceMapAsserts = []string{
	"title = Acceptance Test Service Map Widget Dashboard",
	"widget.0.layout.width = 32",
	"widget.0.servicemap_definition.0.title_align = left",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.layout.x = 5",
	"widget.0.servicemap_definition.0.filters.0 = env:prod",
	"widget.0.servicemap_definition.0.title_size = 16",
	"layout_type = free",
	"widget.0.servicemap_definition.0.service = master-db",
	"is_read_only = true",
	"widget.0.layout.y = 5",
	"widget.0.servicemap_definition.0.title = env: prod, datacenter:us1.prod.dog, service: master-db",
	"widget.0.layout.height = 43",
	"widget.0.servicemap_definition.0.filters.1 = datacenter:us1.prod.dog",
}

func TestAccDatadogDashboardServiceMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardServiceMapConfig, "datadog_dashboard.service_map_dashboard", datadogDashboardServiceMapAsserts)
}

func TestAccDatadogDashboardServiceMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardServiceMapConfig, "datadog_dashboard.service_map_dashboard")
}
