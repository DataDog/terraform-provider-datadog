package datadog

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccDatadogDowntime_Basic(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfig,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
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
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	clock := testClock(t)
	start := clock.Now().Local().Add(time.Hour * time.Duration(3))
	end := start.Add(time.Hour * time.Duration(1))

	config := testAccCheckDatadogDowntimeConfigWithMonitor(start.Unix(), end.Unix())

	resource.Test(t, resource.TestCase{
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
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	clock := testClock(t)
	start := clock.Now().Local().Add(time.Hour * time.Duration(3))
	end := start.Add(time.Hour * time.Duration(1))

	config := testAccCheckDatadogDowntimeConfigWithMonitorTags(start.Unix(), end.Unix())

	resource.Test(t, resource.TestCase{
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
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigMultiScope,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicNoRecurrence(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigNoRecurrence,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogDowntimeExists(accProvider, "datadog_downtime.foo"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "scope.0", "host:NoRecurrence"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "start", "1735707600"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "end", "1735765200"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicUntilDateRecurrence(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigUntilDateRecurrence,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_BasicUntilOccurrencesRecurrence(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigUntilOccurrencesRecurrence,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_WeekDayRecurring(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigWeekDaysRecurrence,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_Updated(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfig,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2925324183", "app:webserver"),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.188033227", "service"),
				),
			},
			{
				Config: testAccCheckDatadogDowntimeConfigUpdated,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntime_TrimWhitespace(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigWhitespace,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func testAccCheckDatadogDowntimeDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
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
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := datadogDowntimeExistsHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func TestAccDatadogDowntimeDates(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogDowntimeConfigDates,
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
						"datadog_downtime.foo", "message", "Example Datadog downtime message."),
					resource.TestCheckResourceAttr(
						"datadog_downtime.foo", "monitor_tags.2679715827", "*"),
				),
			},
		},
	})
}

func TestAccDatadogDowntimeDatesConflict(t *testing.T) {
	accProviders, cleanup := testAccProviders(t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogDowntimeDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config:      testAccCheckDatadogDowntimeConfigDatesConflict,
				ExpectError: regexp.MustCompile("\"start_date\": conflicts with start"),
			},
			{
				Config:      testAccCheckDatadogDowntimeConfigDatesConflict,
				ExpectError: regexp.MustCompile("\"end_date\": conflicts with end"),
			},
		},
	})
}

const testAccCheckDatadogDowntimeConfigDates = `
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start_date = "2099-10-31T11:11:00+01:00"
  end_date = "2099-10-31T21:00:00+01:00"

  recurrence {
    type   = "days"
    period = 1
  }

	message = "Example Datadog downtime message."
  monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigDatesConflict = `
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

	message = "Example Datadog downtime message."
  monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfig = `
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 1
  }

	message = "Example Datadog downtime message."
  monitor_tags = ["service", "app:webserver"]
}
`

func testAccCheckDatadogDowntimeConfigWithMonitor(start int64, end int64) string {
	//When scheduling downtime, Datadog switches the silenced property of monitor to the "end" property of downtime.
	//If that is omitted, the plan doesn't become empty after removing the downtime.
	return fmt.Sprintf(`
resource "datadog_monitor" "downtime_monitor" {
  name = "name for monitor foo"
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

  message = "Example Datadog downtime message."
  monitor_id = "${datadog_monitor.downtime_monitor.id}"
}
`, end, start, end)
}

func testAccCheckDatadogDowntimeConfigWithMonitorTags(start int64, end int64) string {
	//When scheduling downtime, Datadog switches the silenced property of monitor to the "end" property of downtime.
	//If that is omitted, the plan doesn't become empty after removing the downtime.
	return fmt.Sprintf(`
resource "datadog_monitor" "downtime_monitor" {
  name = "name for monitor foo"
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

  message = "Example Datadog downtime message."
	monitor_tags = ["app:webserver"]
}
`, end, start, end)
}

const testAccCheckDatadogDowntimeConfigMultiScope = `
resource "datadog_downtime" "foo" {
  scope = ["host:A", "host:B"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 1
  }

	message = "Example Datadog downtime message."
	monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigNoRecurrence = `
resource "datadog_downtime" "foo" {
  scope = ["host:NoRecurrence"]
  start = 1735707600
  end   = 1735765200
	message = "Example Datadog downtime message."
	monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigUntilDateRecurrence = `
resource "datadog_downtime" "foo" {
  scope = ["host:UntilDateRecurrence"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type       = "days"
    period     = 1
	until_date = 1736226000
  }

	message = "Example Datadog downtime message."
	monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigUntilOccurrencesRecurrence = `
resource "datadog_downtime" "foo" {
  scope = ["host:UntilOccurrencesRecurrence"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type              = "days"
    period            = 1
	until_occurrences = 5
  }

	message = "Example Datadog downtime message."
	monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigWeekDaysRecurrence = `
resource "datadog_downtime" "foo" {
  scope = ["WeekDaysRecurrence"]
  start = 1735646400
  end   = 1735732799

  recurrence {
    period    = 1
	type      = "weeks"
	week_days = ["Sat", "Sun"]
  }

	message = "Example Datadog downtime message."
	monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigUpdated = `
resource "datadog_downtime" "foo" {
  scope = ["Updated"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 3
  }

	message = "Example Datadog downtime message."
	monitor_tags = ["*"]
}
`

const testAccCheckDatadogDowntimeConfigWhitespace = `
resource "datadog_downtime" "foo" {
  scope = ["host:Whitespace"]
  start = 1735707600
  end   = 1735765200

  recurrence {
    type   = "days"
    period = 1
  }

  message = <<EOF
Example Datadog downtime message.
EOF
  monitor_tags = ["*"]
}
`

func TestResourceDatadogDowntimeRecurrenceTypeValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "daily",
			ErrCount: 1,
		},
		{
			Value:    "days",
			ErrCount: 0,
		},
		{
			Value:    "days,weeks",
			ErrCount: 1,
		},
		{
			Value:    "months",
			ErrCount: 0,
		},
		{
			Value:    "years",
			ErrCount: 0,
		},
		{
			Value:    "weeks",
			ErrCount: 0,
		},
	}

	for _, tc := range cases {
		_, errors := validateDatadogDowntimeRecurrenceType(tc.Value, "datadog_downtime_recurrence_type")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected Datadog Downtime Recurrence Type validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func TestResourceDatadogDowntimeRecurrenceWeekDaysValidation(t *testing.T) {
	cases := []struct {
		Value    string
		ErrCount int
	}{
		{
			Value:    "Mon",
			ErrCount: 0,
		},
		{
			Value:    "Mon,",
			ErrCount: 1,
		},
		{
			Value:    "Monday",
			ErrCount: 1,
		},
		{
			Value:    "mon",
			ErrCount: 1,
		},
		{
			Value:    "mon,",
			ErrCount: 1,
		},
		{
			Value:    "monday",
			ErrCount: 1,
		},
		{
			Value:    "mon,Tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon,tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon,Tue",
			ErrCount: 1,
		},
		{
			Value:    "Mon, Tue",
			ErrCount: 1,
		},
	}

	for _, tc := range cases {
		_, errors := validateDatadogDowntimeRecurrenceWeekDays(tc.Value, "datadog_downtime_recurrence_week_days")

		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected Datadog Downtime Recurrence Week Days validation to trigger %d error(s) for value %q - instead saw %d",
				tc.ErrCount, tc.Value, len(errors))
		}
	}
}

func datadogDowntimeDestroyHelper(authV1 context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_downtime" {
			continue
		}

		id, _ := strconv.Atoi(r.Primary.ID)
		dt, _, err := datadogClientV1.DowntimesApi.GetDowntime(authV1, int64(id)).Execute()

		if err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("received an error retrieving downtime %s", err)
		}

		// Datadog only cancels downtime on DELETE
		if !dt.GetActive() {
			continue
		}
		return fmt.Errorf("downtime still exists")
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
