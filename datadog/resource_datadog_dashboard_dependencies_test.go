package datadog

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

func datadogDashboardDepConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  query = "avg(last_1h):anomalies(avg:system.cpu.system{name:cassandra}, 'basic', 3, direction='above', alert_window='last_5m', interval=20, count_default_zero='true') >= 1"
}

resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "monitor"
  description = "some updated description about foo SLO"

  thresholds {
	timeframe = "7d"
	target = 99.5
	warning = 99.8
  }

  monitor_ids = [
    datadog_monitor.foo.id
  ]
}

resource "datadog_dashboard" "foo" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		service_level_objective_definition {
			time_windows = ["90d","previous_week","month_to_date"]
			title_size = "16"
			show_error_budget = true
			title = ""
			title_align = "center"
			slo_id = datadog_service_level_objective.foo.id
			view_mode = "both"
			view_type = "detail"
		}
	}
}`, uniq, uniq, uniq)
}

func datadogDashboardDepChangedConfig(uniq string) string {
	return fmt.Sprintf(
		`resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "some message Notify: @hipchat-channel"
  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"
}

resource "datadog_service_level_objective" "foo" {
  name = "%s"
  type = "monitor"
  description = "some updated description about foo SLO"

  thresholds {
	timeframe = "7d"
	target = 99.5
	warning = 99.8
  }

  monitor_ids = [
    datadog_monitor.foo.id
  ]
}

resource "datadog_dashboard" "foo" {
	title         = "%s"
	description   = "Created using the Datadog provider in Terraform"
	layout_type   = "ordered"
	is_read_only  = "true"

	widget {
		service_level_objective_definition {
			time_windows = ["90d","previous_week","month_to_date"]
			title_size = "16"
			show_error_budget = true
			title = ""
			title_align = "center"
			slo_id = datadog_service_level_objective.foo.id
			view_mode = "both"
			view_type = "detail"
		}
	}
}`, uniq, uniq, uniq)
}

func TestAccDatadogDashboard_NewMonitorForceRecreate(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	dbName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: accProviders,
		Steps: []resource.TestStep{
			{
				Config: datadogDashboardDepConfig(dbName),
				Check: func(firstState *terraform.State) error {
					firstSloId, _ := getSloIdHelper(firstState, accProvider)
					firstDbId, _ := getDbIdHelper(firstState, accProvider)
					resource.TestCheckResourceAttr("datadog_service_level_objective.foo", "id", firstSloId)
					resource.TestCheckResourceAttr("datadog_dashboard.foo", "id", firstDbId)
					resource.Test(t, resource.TestCase{
						PreCheck:  func() { testAccPreCheck(t) },
						Providers: accProviders,
						Steps: []resource.TestStep{
							{
								Config: datadogDashboardDepChangedConfig(dbName),
								Check: resource.ComposeAggregateTestCheckFunc(
									checkThatSloHasBeenForcedToBeRecreated(accProvider, firstSloId),
									checkThatDashboardHasBeenForcedToBeRecreated(accProvider, firstDbId),
								),
							},
						},
					})
					return nil
				},
			},
		},
	})
}

func checkThatDashboardHasBeenForcedToBeRecreated(accProvider *schema.Provider, previousDbId string) func(state *terraform.State) error {
	return func(state *terraform.State) error {
		secondDbId, _ := getDbIdHelper(state, accProvider)
		if secondDbId == previousDbId {
			return fmt.Errorf("dashboard id may have change if the resource as been recreated")
		}
		return nil
	}
}

func getDbIdHelper(state *terraform.State, accProvider *schema.Provider) (string, error) {
	providerConf := accProvider.Meta().(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	for _, r := range state.RootModule().Resources {
		dbResp, _, err := datadogClientV1.DashboardsApi.GetDashboard(authV1, r.Primary.ID).Execute()
		if err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "not found") {
				continue
			}
			return "", fmt.Errorf("received an error retrieving dashboard  %s", err)
		}
		return dbResp.GetId(), err
	}
	return "", fmt.Errorf("dashboard not found in current state")
}
