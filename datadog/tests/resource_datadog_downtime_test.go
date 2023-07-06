package test

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccDatadogDowntime_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "app:webserver"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.1", "service"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicWithMonitor(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	start := clockFromContext(ctx).Now().Local().Add(time.Hour * time.Duration(3))
	end := start.Add(time.Hour * time.Duration(1))

	config := testAccCheckDatadogDowntimeConfigWithMonitor(downtimeMessage, start.Unix(), end.Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	start := clockFromContext(ctx).Now().Local().Add(time.Hour * time.Duration(3))
	end := start.Add(time.Hour * time.Duration(1))

	config := testAccCheckDatadogDowntimeConfigWithMonitorTags(downtimeMessage, start.Unix(), end.Unix())

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicNoRecurrence(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicUntilDateRecurrence(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicUntilOccurrencesRecurrence(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_WeekDayRecurring(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigWeekDaysRecurrence(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "scope:WeekDaysRecurrence"),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_RRule(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigRRule(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "scope:RRuleRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735646400"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735732799"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.rrule", "FREQ=MONTHLY;BYSETPOS=3;BYDAY=WE;INTERVAL=1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "app:webserver"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.1", "service"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeConfigUpdated(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "scope:Updated"),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_TrimWhitespace(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_DiffStart(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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

func testAccCheckDatadogDowntimeDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogDowntimeDestroyHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogDowntimeExists(accProvider func() (*schema.Provider, error), n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := datadogDowntimeExistsHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func TestAccDatadogDowntimeDates(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigDates(downtimeMessage),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "*"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start_date", "2099-10-31T10:11:00Z"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end_date", "2099-10-31T20:00:00Z"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.type", "days"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "recurrence.0.period", "1"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", downtimeMessage),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.0", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntimeDatesConflict(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	downtimeMessage := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogDowntimeDestroy(accProvider),
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

resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = %d
  end   = %d

  message = "%s"
  monitor_id = "${datadog_monitor.downtime_monitor.id}"
}
`, uniq, start, end, uniq)
}

func testAccCheckDatadogDowntimeConfigWithMonitorTags(uniq string, start int64, end int64) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "downtime_monitor" {
  name = "%s"
  type = "metric alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"
  tags = ["app:webserver"]

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
    warning = "1.0"
    critical = "2.0"
  }
}

resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = %d
  end   = %d

  message = "%s"
  monitor_tags = ["app:webserver"]
}
`, uniq, start, end, uniq)
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
  scope = ["scope:WeekDaysRecurrence"]
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
  scope = ["scope:RRuleRecurrence"]
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
  scope = ["scope:Updated"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 3
  }

  mute_first_recovery_notification = true

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
  scope = ["new:somescope"]

  monitor_tags = ["*"]
  message = "%s"
}`, uniq)
}

func datadogDowntimeDestroyHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_downtime" {
			continue
		}

		id, _ := strconv.Atoi(r.Primary.ID)

		err := utils.Retry(2, 5, func() error {
			for _, r := range s.RootModule().Resources {
				if r.Primary.ID != "" {
					dt, httpResp, err := apiInstances.GetDowntimesApiV1().GetDowntime(ctx, int64(id))
					if err != nil {
						if httpResp != nil && httpResp.StatusCode == 404 {
							return nil
						}
						return &utils.FatalError{Prob: fmt.Sprintf("received an error retrieving Downtime %s", err)}
					} else if canceled, ok := dt.GetCanceledOk(); !dt.GetActive() || (ok && canceled != nil) {
						// Datadog only cancels downtime on DELETE so if its not Active, its deleted
						return nil
					}
					return &utils.RetryableError{Prob: "Downtime still exists or is active"}
				}
			}
			return nil
		})
		return err
	}
	return nil
}

func datadogDowntimeExistsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_downtime" {
			continue
		}

		id, _ := strconv.Atoi(r.Primary.ID)
		if _, _, err := apiInstances.GetDowntimesApiV1().GetDowntime(ctx, int64(id)); err != nil {
			return fmt.Errorf("received an error retrieving downtime %s", err)
		}
	}
	return nil
}
