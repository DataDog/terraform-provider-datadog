package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

const datadogDashboardHostMapConfig = `
resource "datadog_dashboard" "hostmap_dashboard" {
    title         = "Acceptance Test Host Map Widget Dashboard"
    description   = "Created using the Datadog provider in Terraform"
    layout_type   = "ordered"
    is_read_only  = "true"

    widget {
		hostmap_definition {
			style {
				fill_min = "10"
                fill_max = "30"
                palette = "YlOrRd"
                palette_flip = true
			}
            node_type = "host"
            no_metric_hosts = "true"
            group = ["region"]
            request {
				size {
					q = "max:system.cpu.user{env:prod} by {host}"
				}
				fill {
					q = "avg:system.cpu.idle{env:prod} by {host}"
				}
			}
			no_group_hosts = "true"
			scope = ["env:prod"]
			title = "system.cpu.idle, system.cpu.user"
            title_align = "right"
			title_size = "16"
		}
    }
}
`

var datadogDashboardHostMapAsserts = []string{
	"widget.0.hostmap_definition.0.style.0.palette_flip = true",
	"widget.0.hostmap_definition.0.request.0.fill.0.q = avg:system.cpu.idle{env:prod} by {host}",
	"widget.0.hostmap_definition.0.title = system.cpu.idle, system.cpu.user",
	"widget.0.hostmap_definition.0.node_type = host",
	"widget.0.hostmap_definition.0.title_align = right",
	"widget.0.hostmap_definition.0.no_metric_hosts = true",
	"description = Created using the Datadog provider in Terraform",
	"layout_type = ordered",
	"widget.0.hostmap_definition.0.style.0.palette = YlOrRd",
	"widget.0.hostmap_definition.0.scope.0 = env:prod",
	"widget.0.hostmap_definition.0.title_size = 16",
	"widget.0.hostmap_definition.0.style.0.fill_max = 30",
	"widget.0.hostmap_definition.0.style.0.fill_min = 10",
	"widget.0.hostmap_definition.0.no_group_hosts = true",
	"widget.0.hostmap_definition.0.request.0.size.0.q = max:system.cpu.user{env:prod} by {host}",
	"is_read_only = true",
	"title = Acceptance Test Host Map Widget Dashboard",
	"widget.0.hostmap_definition.0.group.0 = region",
}

func TestAccDatadogDashboardHostMap(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardHostMapConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrs("datadog_dashboard.hostmap_dashboard", checkDashboardExists(accProvider), datadogDashboardHostMapAsserts)...,
				),
			},
		},
	})
}

func TestAccDatadogDashboardHostMap_import(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: checkDashboardDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardHostMapConfig,
			},
			{
				ResourceName:      "datadog_dashboard.hostmap_dashboard",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
