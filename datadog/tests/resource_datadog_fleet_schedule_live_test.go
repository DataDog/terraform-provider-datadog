package test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
)

// TestAccDatadogFleetSchedule_LiveOrg2Lifecycle is intentionally live-only.
// Run it with the org2 staging credentials exported through the standard
// DD_TEST_CLIENT_* variables and RECORD=none. The schedule is inactive for its
// entire lifetime and targets a deliberately nonexistent non-production host.
func TestAccDatadogFleetSchedule_LiveOrg2Lifecycle(t *testing.T) {
	if os.Getenv("RECORD") != "none" {
		t.Skip("live Fleet Automation lifecycle requires RECORD=none and org2 staging credentials")
	}
	testAccDatadogFleetScheduleLifecycle(t, true)
}

func TestAccDatadogFleetSchedule_Lifecycle(t *testing.T) {
	if os.Getenv("RECORD") == "none" {
		t.Skip("cassette-backed Fleet Automation lifecycle is not run in live-only mode")
	}
	testAccDatadogFleetScheduleLifecycle(t, false)
}

func testAccDatadogFleetScheduleLifecycle(t *testing.T, verifyLiveDeletion bool) {
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	name := uniqueEntityName(ctx, t)
	updatedName := name + "-updated"
	query := fmt.Sprintf("env:terraform-fleet-test host:nonexistent-%d", clockFromContext(ctx).Now().Unix())
	updatedQuery := query + " service:provider-test"
	var capturedID string

	// Keep an API-level cleanup in addition to Terraform's normal destroy so a
	// failed intermediate assertion cannot leak a schedule in org2 staging.
	t.Cleanup(func() {
		if capturedID == "" || providers.frameworkProvider.DatadogApiInstances == nil {
			return
		}
		api := providers.frameworkProvider.DatadogApiInstances.GetFleetAutomationApiV2()
		httpResponse, err := api.DeleteFleetSchedule(providers.frameworkProvider.Auth, capturedID)
		if err != nil && (httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound) {
			t.Errorf("cleaning up Fleet Automation schedule %s: %v", capturedID, err)
		}
	})

	testCase := resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetScheduleConfig(name, query, []string{"Tue"}, 60, "03:00", "UTC", 1),
				Check: resource.ComposeTestCheckFunc(
					captureFleetScheduleID(&capturedID),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "status", "inactive"),
					resource.TestCheckResourceAttr("data.datadog_fleet_schedule.test", "name", name),
					resource.TestCheckResourceAttr("data.datadog_fleet_schedule.test", "rule.start_maintenance_window", "03:00"),
					checkFleetScheduleInPluralDataSource(&capturedID),
				),
			},
			{
				Config: testAccFleetScheduleConfig(updatedName, updatedQuery, []string{"Wed", "Fri"}, 90, "04:30", "America/New_York", 2),
				Check: resource.ComposeTestCheckFunc(
					checkFleetScheduleIDUnchanged(&capturedID),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "name", updatedName),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "query", updatedQuery),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "status", "inactive"),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "version_to_latest", "2"),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "rule.maintenance_window_duration", "90"),
					resource.TestCheckResourceAttr("datadog_fleet_schedule.test", "rule.start_maintenance_window", "04:30"),
					resource.TestCheckResourceAttr("data.datadog_fleet_schedule.test", "name", updatedName),
					checkFleetScheduleInPluralDataSource(&capturedID),
				),
			},
			{
				ResourceName:      "datadog_fleet_schedule.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	}
	// Cassette replay must not depend on Terraform's internal refresh count:
	// older Terraform versions can consume an earlier identical GET response
	// before this callback reaches the cassette's final 404. The live test keeps
	// the strict API-level deletion check, while unit tests cover 404 handling.
	if verifyLiveDeletion {
		testCase.CheckDestroy = testAccCheckFleetScheduleDestroyed(providers.frameworkProvider, &capturedID)
	}
	resource.Test(t, testCase)
}

func testAccFleetScheduleConfig(name, query string, days []string, duration int, start, timezone string, version int) string {
	return fmt.Sprintf(`
resource "datadog_fleet_schedule" "test" {
  name              = %q
  query             = %q
  status            = "inactive"
  version_to_latest = %d

  rule = {
    days_of_week                = [%s]
    maintenance_window_duration = %d
    start_maintenance_window    = %q
    timezone                    = %q
  }
}

data "datadog_fleet_schedule" "test" {
  id = datadog_fleet_schedule.test.id
}

data "datadog_fleet_schedules" "all" {
  depends_on = [datadog_fleet_schedule.test]
}
`, name, query, version, quoteHCLStrings(days), duration, start, timezone)
}

func quoteHCLStrings(values []string) string {
	result := ""
	for i, value := range values {
		if i > 0 {
			result += ", "
		}
		result += fmt.Sprintf("%q", value)
	}
	return result
}

func captureFleetScheduleID(capturedID *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources["datadog_fleet_schedule.test"]
		if !ok || resourceState.Primary.ID == "" {
			return fmt.Errorf("Fleet Automation schedule ID was not set")
		}
		*capturedID = resourceState.Primary.ID
		return nil
	}
}

func checkFleetScheduleIDUnchanged(capturedID *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources["datadog_fleet_schedule.test"]
		if !ok {
			return fmt.Errorf("Fleet Automation schedule resource not found in state")
		}
		if resourceState.Primary.ID != *capturedID {
			return fmt.Errorf("Fleet Automation schedule ID changed from %s to %s", *capturedID, resourceState.Primary.ID)
		}
		return nil
	}
}

func checkFleetScheduleInPluralDataSource(capturedID *string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		dataState, ok := state.RootModule().Resources["data.datadog_fleet_schedules.all"]
		if !ok {
			return fmt.Errorf("plural Fleet Automation schedules data source not found in state")
		}
		count, err := strconv.Atoi(dataState.Primary.Attributes["schedules.#"])
		if err != nil {
			return fmt.Errorf("invalid plural schedule count: %w", err)
		}
		for i := 0; i < count; i++ {
			if dataState.Primary.Attributes[fmt.Sprintf("schedules.%d.id", i)] == *capturedID {
				return nil
			}
		}
		return fmt.Errorf("schedule %s was not returned by the plural data source", *capturedID)
	}
}

func testAccCheckFleetScheduleDestroyed(provider *fwprovider.FrameworkProvider, capturedID *string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if *capturedID == "" {
			return nil
		}
		api := provider.DatadogApiInstances.GetFleetAutomationApiV2()
		_, httpResponse, err := api.GetFleetScheduleV2(provider.Auth, *capturedID)
		if err == nil || httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound {
			return fmt.Errorf("schedule %s still exists after deletion", *capturedID)
		}
		list, _, err := api.ListFleetSchedulesV2(provider.Auth)
		if err != nil {
			return fmt.Errorf("listing schedules after deletion: %w", err)
		}
		for _, schedule := range list.GetData() {
			if schedule.GetId() == *capturedID {
				return fmt.Errorf("schedule %s still appears in list results", *capturedID)
			}
		}
		return nil
	}
}
