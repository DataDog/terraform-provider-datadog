package datadog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogMonitor_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "600"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.0", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1", "baz"),
				),
			},
		},
	})
}

func TestAccDatadogMonitorServiceCheck_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorServiceCheckConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "service check"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", `"custom.check".over("environment:foo").last(2).count_by_status()`),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "600"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.unknown", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.ok", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.0", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1", "baz"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_BasicNoTreshold(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholds,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.0", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1", "bar:baz"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Updated(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "600"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.0", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1", "baz"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "a different message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated! @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "900"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "800"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "no_data_timeframe", "20"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "40"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.ok", "0.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "3.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "2.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "70"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.0", "baz:qux"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1", "quux"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.complex_metric_alert_example_monitor"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "metric alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.complex_query_alert_example_monitor"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_UpdatedToRemoveTags(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "600"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.0", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1", "baz"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdatedWithTagsRemoved,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "a different message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "escalation_message", "the situation has escalated! @pagerduty"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "900"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "800"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "no_data_timeframe", "20"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "40"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.ok", "0.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "3.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "2.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "70"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "true"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "tags.0"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "tags.1"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.complex_metric_alert_example_monitor"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "metric alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.complex_query_alert_example_monitor"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_TrimWhitespace(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigWhitespace,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.ok", "0.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Basic_float_int(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig_ints,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1"),
				),
			},

			{
				Config: testAccCheckDatadogMonitorConfig_ints_mixed,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "3.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "2.0"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Log(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigLogAlert,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists("datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", "name for monitor foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "log alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").last(\"5m\") > 100"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.ok", "0.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2.0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "enable_logs_sample", "true"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*datadog.Client)

	if err := destroyHelper(s, client); err != nil {
		return err
	}
	return nil
}

func testAccCheckDatadogMonitorExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := testAccProvider.Meta().(*datadog.Client)
		if err := existsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

const testAccCheckDatadogMonitorConfig = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  thresholds {
	warning = "1.0"
	critical = "2.0"
	warning_recovery = "0.5"
	critical_recovery = "1.5"
  }

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "baz"]
}
`
const testAccCheckDatadogMonitorConfigNoThresholds = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  notify_no_data = false
  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "bar:baz"]
}
`

const testAccCheckDatadogMonitorServiceCheckConfig = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
  type = "service check"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "\"custom.check\".over(\"environment:foo\").last(2).count_by_status()"

  thresholds {
	warning = 1
	critical = 1
	unknown = 1
	ok = 1
  }

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "baz"]
}
`

const testAccCheckDatadogMonitorConfig_ints = `
resource "datadog_monitor" "foo" {
  name               = "name for monitor foo"
  type               = "query alert"
  message            = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  thresholds {
	warning           = 1
	warning_recovery  = 0
	critical          = 2
	critical_recovery = 1
  }

  notify_no_data    = false
  renotify_interval = 60

  notify_audit        = false
  timeout_h           = 60
  include_tags        = true
  require_full_window = true
  locked              = false

  tags = ["foo:bar", "baz"]
}
`

const testAccCheckDatadogMonitorConfig_ints_mixed = `
resource "datadog_monitor" "foo" {
  name               = "name for monitor foo"
  type               = "query alert"
  message            = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 3"

  thresholds {
	warning           = 1
	warning_recovery  = 0.5
	critical          = 3.0
	critical_recovery = 2
  }

  notify_no_data    = false
  renotify_interval = 60

  notify_audit        = false
  timeout_h           = 60
  include_tags        = true
  require_full_window = true
  locked              = false

  tags = ["foo:bar", "baz"]
}
`

const testAccCheckDatadogMonitorConfigUpdated = `
resource "datadog_monitor" "foo" {
  name = "name for monitor bar"
  type = "query alert"
  message = "a different message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"

  thresholds {
	ok                = "0.0"
	warning           = "1.0"
	warning_recovery  = "0.5"
	critical          = "3.0"
	critical_recovery = "2.5"
  }

  notify_no_data = true
  new_host_delay = 900
  evaluation_delay = 800
  no_data_timeframe = 20
  renotify_interval = 40
  escalation_message = "the situation has escalated! @pagerduty"
  notify_audit = true
  timeout_h = 70
  include_tags = false
  require_full_window = false
  locked = true
  silenced {
	"*" = 0
  }
  tags = ["baz:qux", "quux"]
}
`

const testAccCheckDatadogMonitorConfigUpdatedWithTagsRemoved = `
resource "datadog_monitor" "foo" {
  name = "name for monitor bar"
  type = "query alert"
  message = "a different message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"

  thresholds {
	ok                = "0.0"
	warning           = "1.0"
	warning_recovery  = "0.5"
	critical          = "3.0"
	critical_recovery = "2.5"
  }

  notify_no_data = true
  new_host_delay = 900
  evaluation_delay = 800
  no_data_timeframe = 20
  renotify_interval = 40
  escalation_message = "the situation has escalated! @pagerduty"
  notify_audit = true
  timeout_h = 70
  include_tags = false
  require_full_window = false
  locked = true
  silenced {
	"*" = 0
  }
}
`

const testAccCheckDatadogMonitorConfigMetricAlertNotUpdated = `
resource "datadog_monitor" "complex_metric_alert_example_monitor" {
  name = "Terraform provider datadog complex metrics example Monitor"
  type = "metric alert"
  message = "a message"

  query = "change(min(last_1m),last_5m):sum:org.eclipse.jetty.servlet.ServletContextHandler.5xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.4xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.3xx_responses{example,framework:chronos} > 5"
}
`

const testAccCheckDatadogMonitorConfigQueryAlertNotUpdated = `
resource "datadog_monitor" "complex_query_alert_example_monitor" {
  name = "Terraform provider datadog complex query example Monitor"
  type = "query alert"
  message = "a message"

  query = "change(min(last_1m),last_5m):sum:org.eclipse.jetty.servlet.ServletContextHandler.5xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.4xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.3xx_responses{example,framework:chronos} > 5"
}
`

const testAccCheckDatadogMonitorConfigWhitespace = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
  type = "query alert"
  message = <<EOF
some message Notify: @hipchat-channel
EOF
  escalation_message = <<EOF
the situation has escalated @pagerduty
EOF
  query = <<EOF
avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2
EOF
  thresholds {
	ok = "0.0"
	warning = "1.0"
	warning_recovery = "0.5"
	critical = "2.0"
	critical_recovery = "1.5"
  }

  notify_no_data = false
  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  include_tags = true
}
`

const testAccCheckDatadogMonitorConfigLogAlert = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
  type = "log alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").last(\"5m\") > 2"

  thresholds {
	warning = "1.0"
	critical = "2.0"
	warning_recovery = "0.5"
	critical_recovery = "1.5"
  }

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "baz"]
	enable_logs_sample = true
}
`

func destroyHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetMonitor(i); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("Received an error retrieving monitor %s", err)
		}
		return fmt.Errorf("Monitor still exists")
	}
	return nil
}

func existsHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetMonitor(i); err != nil {
			return fmt.Errorf("Received an error retrieving monitor %s", err)
		}
	}
	return nil
}
