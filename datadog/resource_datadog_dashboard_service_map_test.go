package datadog

import "testing"

// JSON export used as test scenario
//{
//   "notify_list":[],
//   "description":"Created using the Datadog provider in Terraform",
//   "author_name":"--redacted--",
//   "template_variable_presets":[],
//   "template_variables":[],
//   "is_read_only":true,
//   "id":"--redacted--",
//   "title":"{{uniq}}",
//   "url":"--redacted--",
//   "created_at":"2020-10-07T21:43:57.803160+00:00",
//   "modified_at":"2020-10-07T21:46:43.222901+00:00",
//   "author_handle":"--redacted--",
//   "widgets":[
//      {
//         "definition":{
//            "custom_links":[
//               {
//                  "link":"https://app.datadoghq.com/dashboard/lists",
//                  "label":"Test Custom Link label"
//               }
//            ],
//            "title_size":"16",
//            "service":"master-db",
//            "title":"env: prod, datacenter:us1.prod.dog, service: master-db",
//            "title_align":"left",
//            "filters":[
//               "env:prod",
//               "datacenter:us1.prod.dog"
//            ],
//            "type":"servicemap"
//         },
//         "layout":{
//            "y":5,
//            "width":32,
//            "x":5,
//            "height":43
//         },
//         "id": "--redacted--"
//      }
//   ],
//   "layout_type":"free"
//}

const datadogDashboardServiceMapConfig = `
resource "datadog_dashboard" "service_map_dashboard" {
  title         = "{{uniq}}"
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
      		custom_link {
				link = "https://app.datadoghq.com/dashboard/lists"
				label = "Test Custom Link label"
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

var datadogDashboardServiceMapAsserts = []string{
	"title = {{uniq}}",
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
	"widget.0.servicemap_definition.0.custom_link.# = 1",
	"widget.0.servicemap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.0.servicemap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
}

func TestAccDatadogDashboardServiceMap(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardServiceMapConfig, "datadog_dashboard.service_map_dashboard", datadogDashboardServiceMapAsserts)
}

func TestAccDatadogDashboardServiceMap_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtil_import(t, datadogDashboardServiceMapConfig, "datadog_dashboard.service_map_dashboard")
}
