package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDowntimeScheduleBasic(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeSchedule(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr(
						"datadog_downtime_schedule.foo", "display_timezone", "America/New_York"),
					resource.TestCheckResourceAttr(
						"datadog_downtime_schedule.foo", "message", "Message about the downtime"),
					resource.TestCheckResourceAttr(
						"datadog_downtime_schedule.foo", "mute_first_recovery_notification", "false"),
					resource.TestCheckResourceAttr(
						"datadog_downtime_schedule.foo", "scope", "env:(staging OR prod)"),
				),
			},
		},
	})
}

func testAccCheckDatadogDowntimeSchedule(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "test" {
    display_timezone = "America/New_York"
    message = "Message about the downtime"
    monitor_identifier {
      monitor_tags: ["cat:hat", "mat:sat"]
    }
    mute_first_recovery_notification = true
    notify_end_states = ["warn"]
    notify_end_types = ["canceled", "expired"]
    recurring_schedule {
		recurrence {
		  start = "2022-07-13T01:02:03"
		  duration = "1d"
		  rrule    = "FREQ=DAILY;INTERVAL=123"
		}
		recurrence {
		  start = "2022-07-15T01:02:03"
		  duration = "1w"
		  rrule    = "FREQ=DAILY;INTERVAL=12"
		}
    	timezone = "America/New_York"
  }
    scope = "env:(staging OR %v)"
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := DowntimeScheduleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func DowntimeScheduleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			if r.Type != "resource_datadog_downtime_schedule" {
				continue
			}
			id := r.Primary.ID
			_, httpResp, err := apiInstances.GetDowntimesApiV2().GetDowntime(auth, id)
			if err != nil {
				if httpResp != nil && httpResp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving DowntimeSchedule %s", err)}
			}
			return &utils.RetryableError{Prob: "DowntimeSchedule still exists"}
		}
		return nil
	})
	return err
}

func testAccCheckDatadogDowntimeScheduleExists(accProvider *fwprovider.FrameworkProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := downtimeScheduleExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func downtimeScheduleExistsHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "resource_datadog_downtime_schedule" {
			continue
		}
		id := r.Primary.ID
		_, httpResp, err := apiInstances.GetDowntimesApiV2().GetDowntime(auth, id)
		if err != nil {
			return utils.TranslateClientError(err, httpResp, "error retrieving DowntimeSchedule")
		}
	}
	return nil
}
