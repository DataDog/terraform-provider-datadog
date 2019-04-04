package datadog

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
)

const orderedDashboardConfig = `
resource "datadog_dashboard" "ordered_dashboard" {
	title         = "Acceptance Test Ordered Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true

	widget {
		definition = <<JSON
{
	"type": "note",
	"content": "some notes",
	"background_color": "yellow"
}JSON
	}

	widget {
		group_definition {
			layout_type = "ordered"
			title = "New Cluster"

			widget {
			definition = <<JSON
{
"type": "alert_graph",
"alert_id": "123",
"viz_type": "timeseries",
"title": "Alert Graph"
}JSON
			}
		}
	}

	template_variable {
		name   = "var_1"
		prefix = "host"
		default = "aws"
	}
	template_variable {
		name   = "var_2"
		prefix = "service_name"
		default = "autoscaling"
	}
}
`

const freeDashboardConfig = `
resource "datadog_dashboard" "free_dashboard" {
	title         = "Acceptance Test Free Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = false

	widget {
		layout {
			x = 10
			y = 10
			height = 20
			width = 40
		}
		definition = <<JSON
{
"title": "Timeseries",
"requests": [{"q": "avg:system.cpu.user{*} by {availability-zone}","display_type": "area"}],
"type": "timeseries"
}JSON
	}

	template_variable {
		name   = "var_1"
		prefix = "host"
		default = "aws"
	}
	template_variable {
		name   = "var_2"
		prefix = "service_name"
		default = "autoscaling"
	}
}
`

func TestAccDatadogDashboard_update(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: orderedDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists,
					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "title", "Acceptance Test Ordered Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "is_read_only", "true"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.#", "2"),
					// Note widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.0.definition",
						"{\n\t\"type\": \"note\",\n\t\"content\": \"some notes\",\n\t\"background_color\": \"yellow\"\n}"),
					// Group widget
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.group_definition.0.layout_type", "ordered"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.group_definition.0.title", "New Cluster"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.group_definition.0.widget.#", "1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "widget.1.group_definition.0.widget.0.definition",
						"{\n\"type\": \"alert_graph\",\n\"alert_id\": \"123\",\n\"viz_type\": \"timeseries\",\n\"title\": \"Alert Graph\"\n}"),
					// Template Variables
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.0.name", "var_1"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.0.prefix", "host"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.0.default", "aws"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.1.name", "var_2"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.1.prefix", "service_name"),
					resource.TestCheckResourceAttr("datadog_dashboard.ordered_dashboard", "template_variable.1.default", "autoscaling"),
				),
			},
			{
				Config: freeDashboardConfig,
				Check: resource.ComposeTestCheckFunc(
					checkDashboardExists,
					// Dashboard metadata
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "title", "Acceptance Test Free Dashboard"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "description", "Created using the Datadog provider in Terraform"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "layout_type", "free"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "is_read_only", "false"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.#", "1"),
					// Note widget
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "widget.0.definition",
						"{\n\"title\": \"Timeseries\",\n\"requests\": [{\"q\": \"avg:system.cpu.user{*} by {availability-zone}\",\"display_type\": \"area\"}],\n\"type\": \"timeseries\"\n}"),
					// Template Variables
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.#", "2"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.0.name", "var_1"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.0.prefix", "host"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.0.default", "aws"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.1.name", "var_2"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.1.prefix", "service_name"),
					resource.TestCheckResourceAttr("datadog_dashboard.free_dashboard", "template_variable.1.default", "autoscaling"),
				),
			},
		},
	})
}

func checkDashboardExists(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		if _, err := client.GetBoard(r.Primary.ID); err != nil {
			return fmt.Errorf("Received an error retrieving dashboard1 %s", err)
		}
	}
	return nil
}

func checkDashboardDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)
	for _, r := range s.RootModule().Resources {
		if _, err := client.GetBoard(r.Primary.ID); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving dashboard2 %s", err)
		}
		return fmt.Errorf("Timeboard still exists")
	}
	return nil
}
