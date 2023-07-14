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

func TestAccDowntimeScheduleBasicRecurring(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeScheduleRecurring(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "scope", fmt.Sprintf("env:(staging OR %v)", uniq)),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.#", "1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.0", "cat:hat"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.timezone", "America/New_York"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.#", "3"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.start", "2022-07-13T01:02:03"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.duration", "1d"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.start", "2022-07-15T01:02:03"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.duration", "1w"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.rrule", "FREQ=DAILY;INTERVAL=12"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.2.duration", "1m"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.2.rrule", "FREQ=DAILY;INTERVAL=123"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "display_timezone", "America/New_York"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "message", "Message about the downtime"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "mute_first_recovery_notification", "true"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_states.#", "2"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_states.*", "alert"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_states.*", "warn"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_types.#", "1"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_types.*", "expired"),
				),
			},
		},
	})
}

func TestAccDowntimeScheduleBasicOneTime(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeScheduleOneTime(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "monitor_identifier.monitor_tags.#", "2"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "monitor_identifier.monitor_tags.0", "cat:hat"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "monitor_identifier.monitor_tags.1", "mat:sat"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "scope", fmt.Sprintf("env:(staging OR %v)", uniq)),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "one_time_schedule.start", "2050-01-02T03:04:05Z"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "notify_end_states.#", "3"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t1", "notify_end_states.*", "alert"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t1", "notify_end_states.*", "no data"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t1", "notify_end_states.*", "warn"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t1", "notify_end_types.#", "0"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeScheduleOneTimeUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet("datadog_downtime_schedule.t1", "one_time_schedule.start"), // Is there a way to check a subset of the string is now?
				),
			},
		},
	})
}

func testAccCheckDatadogDowntimeScheduleRecurring(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
    scope = "env:(staging OR %v)"
    monitor_identifier {
      monitor_tags = ["cat:hat"]
    }
    recurring_schedule {
		recurrence {
		  start = "2022-07-13T01:02:03"
		  duration = "1d"
		  rrule    = "FREQ=DAILY;INTERVAL=1"
		}
		recurrence {
		  start = "2022-07-15T01:02:03"
		  duration = "1w"
		  rrule    = "FREQ=DAILY;INTERVAL=12"
		}
		recurrence {
		  start = "2022-07-15T01:02:03"
		  duration = "1m"
		  rrule    = "FREQ=DAILY;INTERVAL=123"
		}

    	timezone = "America/New_York"
    }
    display_timezone = "America/New_York"
    message = "Message about the downtime"
    mute_first_recovery_notification = true
    notify_end_states = ["warn", "alert"]
    notify_end_types = ["expired"]
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleOneTime(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t1" {
    scope = "env:(staging OR %v)"
    monitor_identifier {
      monitor_tags = ["cat:hat", "mat:sat"]
    }
    one_time_schedule {
    	start = "2050-01-02T03:04:05Z"
    }
    notify_end_types = []
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleOneTimeUpdate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t1" {
    scope = "env:(changed OR %v)"
    monitor_identifier {
      monitor_tags = ["vat:mat"]
    }
    one_time_schedule {
    	start = null
		end = "2060-01-02T03:04:05Z"
    }
    message = "updated"
	notify_end_states = ["alert"]
    notify_end_types = ["canceled"]
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
