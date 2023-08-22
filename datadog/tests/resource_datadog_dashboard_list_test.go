package test

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
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
			live_span = "1h"
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
            live_span = "1h"
		}
		widget_layout {
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
            live_span = "1h"
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
            live_span = "1h"
		}
	}
}`, uniq, uniq)
}

func TestDatadogDashListImport(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_dashboard_list.new_list"
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)

	// Getting the hash for a TypeSet element that has dynamic elements isn't possible
	// So instead we use an import test to make sure the resource can be imported properly.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashListDestroyWithFw(providers.frameworkProvider),
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
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniqueName := uniqueEntityName(ctx, t)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDashListDestroyWithFw(providers.frameworkProvider),
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

// Migrate this to fw when other dashboard resources are migrated
func testAccCheckDatadogDashListDestroy(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		return datadogDashListDestroyHelper(auth, s, apiInstances)
	}
}

// remove this when all the dashboard resources are migrated to framework
func testAccCheckDatadogDashListDestroyWithFw(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth
		return datadogDashListDestroyHelper(auth, s, apiInstances)
	}
}

func datadogDashListDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if !strings.Contains(r.Primary.Attributes["name"], "List") {
			continue
		}
		id, _ := strconv.Atoi(r.Primary.ID)
		_, _, errList := apiInstances.GetDashboardListsApiV1().GetDashboardList(ctx, int64(id))
		if errList != nil {
			if strings.Contains(strings.ToLower(errList.Error()), "not found") {
				continue
			}
			return fmt.Errorf("received an error retrieving Dash List %s", errList)
		}
		return fmt.Errorf("dashoard list still exists")
	}
	return nil
}
