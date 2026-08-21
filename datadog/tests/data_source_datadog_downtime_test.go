package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Seeds a datadog_downtime_schedule (same v2 Downtimes API the data source reads) and looks it
// up by ID, so the generated field mapping runs against a payload with every mapped field set.
func TestAccDatadogDowntimeDatasource(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceDatadogDowntimeConfig(uniq),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.datadog_downtime.foo", "id",
						"datadog_downtime_schedule.t", "id"),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "scope", fmt.Sprintf("env:(staging OR %s)", uniq)),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "message", "Message about the downtime"),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "display_timezone", "America/New_York"),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "mute_first_recovery_notification", "true"),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "notify_end_states.#", "2"),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "notify_end_types.#", "1"),
					resource.TestCheckResourceAttr("data.datadog_downtime.foo", "notify_end_types.0", "expired"),
					resource.TestCheckResourceAttrSet("data.datadog_downtime.foo", "status"),
					resource.TestCheckResourceAttrSet("data.datadog_downtime.foo", "created"),
				),
			},
		},
	})
}

func testAccDatasourceDatadogDowntimeConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
  scope = "env:(staging OR %s)"
  monitor_identifier {
    monitor_tags = ["cat:hat"]
  }
  recurring_schedule {
    recurrence {
      start    = "2042-07-13T01:02:03"
      duration = "1d"
      rrule    = "FREQ=DAILY;INTERVAL=1"
    }
    timezone = "America/New_York"
  }
  display_timezone                 = "America/New_York"
  message                          = "Message about the downtime"
  mute_first_recovery_notification = true
  notify_end_states                = ["warn", "alert"]
  notify_end_types                 = ["expired"]
}

data "datadog_downtime" "foo" {
  id = datadog_downtime_schedule.t.id
}
`, uniq)
}
