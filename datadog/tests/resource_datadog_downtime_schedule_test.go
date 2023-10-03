package test

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/fwprovider"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

func TestAccDowntimeScheduleBasicRecurring_Import(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeScheduleRecurring(uniq),
			},
			{
				ResourceName:      "datadog_downtime_schedule.t",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

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
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.start", "2042-07-13T01:02:03"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.duration", "1d"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.start", "2042-07-15T01:02:03"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.duration", "1w"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.rrule", "FREQ=DAILY;INTERVAL=12"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.2.start", "2042-07-17T01:02:03"),
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
			{
				Config: testAccCheckDatadogDowntimeScheduleRecurringUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "scope", fmt.Sprintf("env:(staging OR %v)", uniq)),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.#", "1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.0", "cat:hat"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.timezone", "UTC"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.#", "2"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.start", "2042-07-13T01:02:03"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.duration", "1d"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.start", "3022-07-15T01:02:03"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.duration", "1m"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.1.rrule", "FREQ=WEEKLY;INTERVAL=123"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "display_timezone", "UTC"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "message", "Updated message"),
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
	currentTime := clockFromContext(ctx).Now()
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeScheduleOneTime(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.#", "2"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.0", "cat:hat"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.1", "mat:sat"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "scope", fmt.Sprintf("env:(staging OR %v)", uniq)),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "one_time_schedule.start", "2050-01-02T03:04:05Z"),
					resource.TestCheckNoResourceAttr("datadog_downtime_schedule.t", "one_time_schedule.end"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_states.#", "3"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_states.*", "alert"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_states.*", "no data"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_states.*", "warn"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_types.#", "0"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "mute_first_recovery_notification", "false"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "display_timezone", "UTC"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeScheduleOneTimeUpdate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "scope", fmt.Sprintf("env:(changed OR %v)", uniq)),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.#", "1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "monitor_identifier.monitor_tags.0", "vat:mat"),
					getCheckNowDate("one_time_schedule.start", "2006-01-02T15:04:05Z", currentTime),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "one_time_schedule.end", "2060-01-02T03:04:05Z"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "mute_first_recovery_notification", "true"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_states.#", "1"),
					resource.TestCheckTypeSetElemAttr("datadog_downtime_schedule.t", "notify_end_states.*", "alert"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_types.#", "1"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "notify_end_types.0", "canceled"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "display_timezone", "UTC"),
				),
			},
		},
	})
}

func TestAccDowntimeScheduleBasicOneTimeWithMonitorID(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDowntimeOneTimeScheduleConfigWithMonitor(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttrSet("datadog_downtime_schedule.t", "monitor_identifier.monitor_id"),
				),
			},
		},
	})
}

func TestAccDowntimeScheduleChanged(t *testing.T) {
	t.Parallel()
	ctx, providers, accProviders := testAccFrameworkMuxProviders(context.Background(), t)
	currentTime := clockFromContext(ctx).Now()
	uniq := uniqueEntityName(ctx, t)

	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: accProviders,
		CheckDestroy:             testAccCheckDatadogDowntimeScheduleDestroy(providers.frameworkProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeScheduleChangeCreate(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "one_time_schedule.start", "2050-01-02T03:04:05Z"),
					resource.TestCheckNoResourceAttr("datadog_downtime_schedule.t", "one_time_schedule.end"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeScheduleChangeUpdate1(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.#", "1"),
					getCheckNowDate("recurring_schedule.recurrence.0.start", "2006-01-02T15:04:05", currentTime),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.duration", "1d"),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "recurring_schedule.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=11"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeScheduleChangeUpdate2(uniq),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeScheduleExists(providers.frameworkProvider),
					getCheckNowDate("one_time_schedule.start", "2006-01-02T15:04:05Z", currentTime),
					resource.TestCheckResourceAttr("datadog_downtime_schedule.t", "one_time_schedule.end", "3000-01-02T03:04:05Z"),
				),
			},
		},
	})
}

func getCheckNowDate(attributeName string, dateFormat string, currentTime time.Time) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check that the calculated start now string is close to now
		startStr := s.RootModule().Resources["datadog_downtime_schedule.t"].Primary.Attributes[attributeName]
		timestamp, _ := time.Parse(dateFormat, startStr)
		timeDifference := math.Abs(float64(currentTime.UTC().Sub(timestamp)))
		if timeDifference >= float64(time.Minute) {
			return fmt.Errorf("Attribute value %v does not match a now date %v", startStr, currentTime.UTC().Format(time.RFC3339))
		}
		return nil
	}
}

func testAccCheckDowntimeOneTimeScheduleConfigWithMonitor(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "downtime_monitor" {
  name = "%s"
  type = "metric alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
    warning = "1.0"
    critical = "2.0"
  }
}

resource "datadog_downtime_schedule" "t" {
    scope = "env:(staging OR %v)"
    monitor_identifier {
      monitor_id = datadog_monitor.downtime_monitor.id
    }
    one_time_schedule {
    	start = "2050-01-02T03:04:05Z"
    }
    notify_end_types = []
}
`, uniq, uniq)
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
		  start = "2042-07-13T01:02:03"
		  duration = "1d"
		  rrule    = "FREQ=DAILY;INTERVAL=1"
		}
		recurrence {
		  start = "2042-07-15T01:02:03"
		  duration = "1w"
		  rrule    = "FREQ=DAILY;INTERVAL=12"
		}
		recurrence {
		  start = "2042-07-17T01:02:03"
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

func testAccCheckDatadogDowntimeScheduleRecurringUpdate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
    scope = "env:(staging OR %v)"
    monitor_identifier {
      monitor_tags = ["cat:hat"]
    }
    recurring_schedule {
        recurrence {
		  start = "2042-07-13T01:02:03"
		  duration = "1d"
		  rrule    = "FREQ=DAILY;INTERVAL=1"
		}
        recurrence {
		  start = "3022-07-15T01:02:03"
		  duration = "1m"
		  rrule    = "FREQ=WEEKLY;INTERVAL=123"
		}
    	timezone = "UTC"
    }
    display_timezone = "UTC"
    message = "Updated message"
    mute_first_recovery_notification = true
    notify_end_states = ["warn", "alert"]
    notify_end_types = ["expired"]
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleOneTime(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
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
resource "datadog_downtime_schedule" "t" {
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
	mute_first_recovery_notification = true
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleChangeCreate(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
    scope = "env:%v"
    monitor_identifier {
      monitor_tags = ["cat:hat"]
    }
    one_time_schedule {
    	start = "2050-01-02T03:04:05Z"
    }
	display_timezone = "UTC"
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleChangeUpdate1(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
    scope = "env:%v"
    monitor_identifier {
      monitor_tags = ["cat:hat"]
    }
    recurring_schedule {
      recurrence {
        duration = "1d"
        rrule    = "FREQ=DAILY;INTERVAL=11"
       }
      timezone = "UTC"
    }
	display_timezone = "UTC"
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleChangeUpdate2(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime_schedule" "t" {
    scope = "env:%v"
    monitor_identifier {
      monitor_tags = ["cat:hat"]
    }
    one_time_schedule {
    	start = null
        end = "3000-01-02T03:04:05Z"
      }
     display_timezone = "UTC"
}`, uniq)
}

func testAccCheckDatadogDowntimeScheduleDestroy(accProvider *fwprovider.FrameworkProvider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		apiInstances := accProvider.DatadogApiInstances
		auth := accProvider.Auth

		if err := downtimeScheduleDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func downtimeScheduleDestroyHelper(auth context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
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
