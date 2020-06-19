package datadog

//import (
//	"testing"
//
//	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
//)
//
//// JSON export used as test scenario
////{
////    "notify_list": [],
////    "description": "",
////    "author_name": "--redacted--",
////    "id": "--redacted--",
////    "url": "--redacted--",
////    "template_variables": [],
////    "is_read_only": false,
////    "title": "TF - Service Map Example",
////    "created_at": "2020-06-09T13:32:03.535027+00:00",
////    "modified_at": "2020-06-09T13:32:50.224757+00:00",
////    "author_handle": "--redacted--",
////    "widgets": [
////        {
////            "definition": {
////                "title_size": "16",
////                "service": "master-db",
////                "title": "env: prod, datacenter:us1.prod.dog, service: master-db",
////                "title_align": "left",
////                "filters": [
////                    "env:prod",
////                    "datacenter:us1.prod.dog"
////                ],
////                "type": "servicemap"
////            },
////            "layout": {
////                "y": 3,
////                "x": -1,
////                "height": 15,
////                "width": 47
////            },
////            "id": 0
////        }
////    ],
////    "layout_type": "free"
////}
//
//const datadogDashboardServiceMapConfig = `
//resource "datadog_dashboard" "service_map_dashboard" {
//   title         = "Acceptance Test Service Map Widget Dashboard"
//   description   = "Created using the Datadog provider in Terraform"
//   layout_type   = "ordered"
//   is_read_only  = "true"
//
//   widget {
//		servicemap_definition {
//			title_size = "16"
//			service = "master-db"
//			title = "env: prod, datacenter:us1.prod.dog, service: master-db"
//			title_align = "left"
//			filters = ["env:prod","datacenter:us1.prod.dog"]
//		}
//   }
//}
//`
//
//var datadogDashboardServiceMapAsserts = []string{
//}
//
//func TestAccDatadogDashboardServiceMap(t *testing.T) {
//	accProviders, cleanup := testAccProviders(t)
//	defer cleanup(t)
//	accProvider := testAccProvider(t, accProviders)
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    accProviders,
//		CheckDestroy: checkDashboardDestroy(accProvider),
//		Steps: []resource.TestStep{
//			{
//				Config: datadogDashboardServiceMapConfig,
//				Check: resource.ComposeTestCheckFunc(
//					testCheckResourceAttrs("datadog_dashboard.service_map_dashboard", checkDashboardExists(accProvider), datadogDashboardServiceMapAsserts)...,
//				),
//			},
//		},
//	})
//}
//
//func TestAccDatadogDashboardServiceMap_import(t *testing.T) {
//	accProviders, cleanup := testAccProviders(t)
//	defer cleanup(t)
//	accProvider := testAccProvider(t, accProviders)
//
//	resource.Test(t, resource.TestCase{
//		PreCheck:     func() { testAccPreCheck(t) },
//		Providers:    accProviders,
//		CheckDestroy: checkDashboardDestroy(accProvider),
//		Steps: []resource.TestStep{
//			{
//				Config: datadogDashboardServiceMapConfig,
//			},
//			{
//				ResourceName:      "datadog_dashboard.service_map_dashboard",
//				ImportState:       true,
//				ImportStateVerify: true,
//			},
//		},
//	})
//}
