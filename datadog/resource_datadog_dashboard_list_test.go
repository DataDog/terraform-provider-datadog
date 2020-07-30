package datadog

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
	accProviders, cleanup := testAccProviders(t, initRecorder(t))
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	// Getting the hash for a TypeSet element that has dynamic elements isn't possible
	// So instead we use an import test to make sure the resource can be imported properly.
	resource.ParallelTest(t, resource.TestCase{
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
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return datadogDashListDestroyHelper(s, authV1, datadogClientV1)
	}
}

func datadogDashListDestroyHelper(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if !strings.Contains(r.Primary.Attributes["name"], "List") {
			continue
		}
		id, _ := strconv.Atoi(r.Primary.ID)
		_, _, errList := datadogClientV1.DashboardListsApi.GetDashboardList(authV1, int64(id)).Execute()
		if errList != nil {
			if strings.Contains(strings.ToLower(errList.Error()), "not found") {
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
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		return datadogDashListExistsHelper(s, authV1, datadogClientV1)
	}
}

func datadogDashListExistsHelper(s *terraform.State, authV1 context.Context, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if !strings.Contains(r.Primary.Attributes["name"], "List") {
			continue
		}
		id, _ := strconv.Atoi(r.Primary.ID)
		_, _, errList := datadogClientV1.DashboardListsApi.GetDashboardList(authV1, int64(id)).Execute()
		if errList != nil {
			if strings.Contains(strings.ToLower(errList.Error()), "not found") {
				continue
			}
			return fmt.Errorf("received an error retrieving Dash List %s", errList)
		}

		return nil
	}
	return nil
}
