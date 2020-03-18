package datadog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

const testAccCheckDatadogDashListConfig = `
resource "datadog_dashboard_list" "new_list" {
	depends_on = [
		"datadog_dashboard.screen",
		"datadog_dashboard.time"
	]

    name = "Terraform newest Created List"
    dash_item {
        type = "custom_timeboard"
        dash_id = "${datadog_dashboard.time.id}"
    }
    dash_item {
        type = "custom_screenboard"
        dash_id = "${datadog_dashboard.screen.id}"
	}
}

resource "datadog_dashboard" "time" {
	title         = "TF Test Layout Dashboard"
	# NOTE: this dependency is present to make sure the dashboards are created in
	# a predictable order and thus the recorder test cassettes always work
	depends_on    = ["datadog_dashboard.screen"]
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = true
	widget {
		alert_graph_definition {
		  alert_id = "1234"
		  viz_type = "timeseries"
		  title = "Widget Title"
		  time = {
			live_span = "1h"
		  }
		}
	  }

}

resource "datadog_dashboard" "screen" {
	title         = "TF Test Free Layout Dashboard"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "free"
	is_read_only  = false
	widget {
		event_stream_definition {
		  query = "*"
		  event_size = "l"
		  title = "Widget Title"
		  title_size = 16
		  title_align = "left"
		  time = {
			live_span = "1h"
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

func TestDatadogDashListImport(t *testing.T) {
	resourceName := "datadog_dashboard_list.new_list"
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	// Getting the hash for a TypeSet element that has dynamic elements isn't possible
	// So instead we use an import test to make sure the resource can be imported properly.
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashListConfig,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogDashListDestroy(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		return datadogDashListDestroyHelper(s, client)
	}
}

func datadogDashListDestroyHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if !strings.Contains(r.Primary.Attributes["name"], "List") {
			continue
		}
		id, _ := strconv.Atoi(r.Primary.ID)
		_, errList := client.GetDashboardList(id)
		if errList != nil {
			if strings.Contains(errList.Error(), "not found") {
				continue
			}
			return fmt.Errorf("received an error retrieving Dash List %s", errList)
		}

		return fmt.Errorf("dashoard List still exists")
	}
	return nil
}

func testAccCheckDatadogDashListExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		return datadogDashListExistsHelper(s, client)
	}
}

func datadogDashListExistsHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		if !strings.Contains(r.Primary.Attributes["name"], "List") {
			continue
		}
		id, _ := strconv.Atoi(r.Primary.ID)
		_, errList := client.GetDashboardList(id)
		if errList != nil {
			if strings.Contains(errList.Error(), "not found") {
				continue
			}
			return fmt.Errorf("received an error retrieving Dash List %s", errList)
		}

		return nil
	}
	return nil
}
