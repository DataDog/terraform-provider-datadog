package test

import (
	"context"
	"fmt"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogDowntime_Basic(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfig(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "*"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2925324183", "app:webserver"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.188033227", "service"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicWithMonitor(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	start := clockFromContext(ctx).Now().Local().Add(time.Hour * time.Duration(3))
	end := start.Add(time.Hour * time.Duration(1))

	config := testAccCheckDatadogDowntimeConfigWithMonitor(downtimeMessage, start.Unix(), end.Unix())

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicWithMonitorTags(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	start := clockFromContext(ctx).Now().Local().Add(time.Hour * time.Duration(3))
	end := start.Add(time.Hour * time.Duration(1))

	config := testAccCheckDatadogDowntimeConfigWithMonitorTags(downtimeMessage, start.Unix(), end.Unix())

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicMultiScope(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigMultiScope(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "host:A"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.1", "host:B"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicNoRecurrence(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigNoRecurrence(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "host:NoRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicUntilDateRecurrence(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigUntilDateRecurrence(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "host:UntilDateRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.until_date", "1736226000"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicUntilOccurrencesRecurrence(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigUntilOccurrencesRecurrence(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "host:UntilOccurrencesRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.until_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_WeekDayRecurring(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigWeekDaysRecurrence(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "WeekDaysRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735646400"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735732799"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "weeks"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.week_days.0", "Sat"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.week_days.1", "Sun"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_RRule(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigRRule(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "RRuleRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735646400"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735732799"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.rrule", "FREQ=MONTHLY;BYSETPOS=3;BYDAY=WE;INTERVAL=1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_Updated(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfig(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "*"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2925324183", "app:webserver"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.188033227", "service"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeConfigUpdated(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "Updated"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "3"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_TrimWhitespace(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigWhitespace(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "host:Whitespace"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_DiffStart(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigDiffStart(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
				),
			},
		},
	})
}

func testAccCheckDatadogDowntimeDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := datadogDowntimeDestroyHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogDowntimeExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := datadogDowntimeExistsHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func TestAccDatadogDowntimeDates(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigDates(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "*"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start_date", "2099-10-31T11:11:00+01:00"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end_date", "2099-10-31T21:00:00+01:00"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntimeDatesConflict(t *testing.T) {
	ctx, accProviders := testAccProviders(context.Background(), t, initRecorder(t))
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	parallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogDowntimeConfigDatesConflict(downtimeMessage),
				ExpectError: regexp.MustCompile("\"start_date\": conflicts with start"),
			},
			{
				Config:      testAccCheckDatadogDowntimeConfigDatesConflict(downtimeMessage),
				ExpectError: regexp.MustCompile("\"end_date\": conflicts with end"),
			},
		},
	})
}

func testAccCheckDatadogDowntimeConfigDates(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start_date = "2099-10-31T11:11:00+01:00"
  end_date = "2099-10-31T21:00:00+01:00"

  recurrence {
    type   = "days"
    period = 1
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigDatesConflict(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start_date = "2099-10-31T11:11:00+01:00"
  start = 1735707600
  end_date = "2099-10-31T11:11:00+01:00"
  end = 1735707600

  recurrence {
    type   = "days"
    period = 1
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 1
  }

  message = "%s"
  monitor_tags = ["service", "app:webserver"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigWithMonitor(uniq string, start int64, end int64) string {
	//When scheduling downtime, Datadog switches the silenced property of monitor to the "end" property of downtime.
	//If that is omitted, the plan doesn't become empty after removing the downtime.
	return fmt.Sprintf(`
resource "datadog_monitor" "downtime_monitor" {
  name = "%s"
  type = "metric alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  thresholds = {
    warning = "1.0"
    critical = "2.0"
  }
  silenced = {
    "*" = %d
  }
}

resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = %d
  end   = %d

  message = "%s"
  monitor_id = "${datadog_monitor.downtime_monitor.id}"
}
`, uniq, end, start, end, uniq)
}

func testAccCheckDatadogDowntimeConfigWithMonitorTags(uniq string, start int64, end int64) string {
	//When scheduling downtime, Datadog switches the silenced property of monitor to the "end" property of downtime.
	//If that is omitted, the plan doesn't become empty after removing the downtime.
	return fmt.Sprintf(`
resource "datadog_monitor" "downtime_monitor" {
  name = "%s"
  type = "metric alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"
  tags = ["app:webserver"]

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  thresholds = {
    warning = "1.0"
    critical = "2.0"
  }
  silenced = {
    "*" = %d
  }
}

resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = %d
  end   = %d

  message = "%s"
  monitor_tags = ["app:webserver"]
}
`, uniq, end, start, end, uniq)
}

func testAccCheckDatadogDowntimeConfigMultiScope(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["host:A", "host:B"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 1
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigNoRecurrence(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["host:NoRecurrence"]
  start = 1735707600
  end   = 1735765200
  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigUntilDateRecurrence(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["host:UntilDateRecurrence"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type       = "days"
    period     = 1
    until_date = 1736226000
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigUntilOccurrencesRecurrence(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["host:UntilOccurrencesRecurrence"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type              = "days"
    period            = 1
    until_occurrences = 5
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigWeekDaysRecurrence(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["WeekDaysRecurrence"]
  start = 1735646400
  end   = 1735732799

  recurrence {
    period    = 1
    type      = "weeks"
    week_days = ["Sat", "Sun"]
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigRRule(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["RRuleRecurrence"]
  start = 1735646400
  end   = 1735732799

  recurrence {
    type     = "rrule"
    rrule    = "FREQ=MONTHLY;BYSETPOS=3;BYDAY=WE;INTERVAL=1"
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["Updated"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 3
  }

  message = "%s"
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigWhitespace(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["host:Whitespace"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 1
  }

  message = <<EOF
%s
EOF
  monitor_tags = ["*"]
}`, uniq)
}

func testAccCheckDatadogDowntimeConfigDiffStart(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_downtime" "foo" {
  scope = ["somescope"]

  monitor_tags = ["*"]
  active = true
  message = "%s"
}`, uniq)
}

func datadogDowntimeDestroyHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_downtime" {
			continue
		}

		id, _ := strconv.Atoi(r.Primary.ID)

		err := utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Primary.ID != "" {
					if dt, httpResp, err := datadogClientV1.DowntimesApi.GetDowntime(authV1, int64(id)).Execute(); err != nil {
						if httpResp != nil && httpResp.StatusCode == 404 {
							return nil
						}
						return &utils.FatalError{Prob: fmt.Sprintf("received an error retrieving Downtime %s", err)}
					} else {
						// Datadog only cancels downtime on DELETE so if its not Active, its deleted
						if !dt.GetActive() {
							return nil
						}
					}
					return &utils.RetryableError{Prob: fmt.Sprintf("Downtime still exists or is active")}
				}
			}
			return nil
		})
		return err
	}
	return nil
}

func datadogDowntimeExistsHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_downtime" {
			continue
		}

		id, _ := strconv.Atoi(r.Primary.ID)
		if _, _, err := datadogClientV1.DowntimesApi.GetDowntime(authV1, int64(id)).Execute(); err != nil {
			return fmt.Errorf("received an error retrieving downtime %s", err)
		}
	}
	return nil
}
