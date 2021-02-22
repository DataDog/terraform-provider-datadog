package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func testAccCheckDatadogDashListConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_list" "new_list" {
	depends_on = [
		"datadog_dashboard.screen",
		"datadog_dashboard.time"
	]

	name = "%s"
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
	title        = "%s-time"
	# NOTE: this dependency is present to make sure the dashboards are created in
	# a predictable order and thus the recorder test cassettes always work
	depends_on   = ["datadog_dashboard.screen"]
	description  = "Created using the Datadog provider in Terraform"
	layout_type  = "ordered"
	is_read_only = true
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
	title        = "%s-screen"
	description  = "Created using the Datadog provider in Terraform"
	layout_type  = "free"
	is_read_only = false
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
}`, uniq, uniq, uniq)
}

func testAccCheckDatadogDashListConfigInDashboard(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_list" "new_list" {
	name = "%s"
}

resource "datadog_dashboard" "time" {
	title        = "%s-time"
	description  = "Created using the Datadog provider in Terraform"
	layout_type  = "ordered"
	is_read_only = true
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
	dashboard_lists = ["${datadog_dashboard_list.new_list.id}"]
}`, uniq, uniq)
}

func testAccCheckDatadogDashListConfigRemoveFromDashboard(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_dashboard_list" "new_list" {
	name = "%s"
}

resource "datadog_dashboard" "time" {
	title        = "%s-time"
	description  = "Created using the Datadog provider in Terraform"
	layout_type  = "ordered"
	is_read_only = true
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
}`, uniq, uniq)
}

func TestDatadogDashListImport(t *testing.T) {
	resourceName := "datadog_dashboard_list.new_list"
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	// Getting the hash for a TypeSet element that has dynamic elements isn't possible
	// So instead we use an import test to make sure the resource can be imported properly.

	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashListConfig(uniqueName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestDatadogDashListInDashboard(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	t.Parallel()
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDashListDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDashListConfigInDashboard(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard.time", "dashboard_lists.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_dashboard.time", "dashboard_lists_removed.#", "0"),
				),
				// The plan is non empty, because in this case the list is the same file
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccCheckDatadogDashListConfigRemoveFromDashboard(uniqueName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_dashboard.time", "dashboard_lists.#", "0"),
					resource.TestCheckResourceAttr(
						"datadog_dashboard.time", "dashboard_lists_removed.#", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogDashListDestroy(accProvider *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
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
