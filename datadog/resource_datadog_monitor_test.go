package datadog

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/zorkian/go-datadog-api"
)

func TestAccDatadogMonitor_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
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
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					// Tags are a TypeSet => use a weird way to access members by their hash
					// TF TypeSet is internally represented as a map that maps computed hashes
					// to actual values. Since the hashes are always the same for one value,
					// this is the way to get them.
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1750285118", "foo:bar"),
				),
			},
		},
	})
}

func TestAccDatadogMonitorServiceCheck_Basic(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorServiceCheckConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
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
					// Tags are a TypeSet => use a weird way to access members by their hash
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1750285118", "foo:bar"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_BasicNoTreshold(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholds(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
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
					// Tags are a TypeSet => use a weird way to access members by their hash
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.3417822676", "bar:baz"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1750285118", "foo:bar"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Updated(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	monitorNameUpdated := monitorName + "-updated"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
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
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2"),
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
					// Tags are a TypeSet => use a weird way to access members by their hash
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1750285118", "foo:bar"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorNameUpdated),
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
						"datadog_monitor.foo", "thresholds.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "3"),
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
					// Tags are a TypeSet => use a weird way to access members by their hash
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1280427750", "baz:qux"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1520885421", "quux"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.complex_metric_alert_example_monitor"),
					// even though this is defined as a metric alert, the API will actually return query alert
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "query alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.complex_query_alert_example_monitor"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_UpdatedToRemoveTags(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	monitorNameUpdated := monitorName + "-updated"
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
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
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2"),
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
					// Tags are a TypeSet => use a weird way to access members by their hash
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.2644851163", "baz"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.1750285118", "foo:bar"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdatedWithTagsRemoved(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorNameUpdated),
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
						"datadog_monitor.foo", "thresholds.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "3"),
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
						"datadog_monitor.foo", "tags.#"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.complex_metric_alert_example_monitor"),
					// even though this is defined as a metric alert, the API will actually return query alert
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "query alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.complex_query_alert_example_monitor"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_TrimWhitespace(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigWhitespace(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
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
						"datadog_monitor.foo", "thresholds.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "1.5"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Basic_float_int(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigInts(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
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
				Config: testAccCheckDatadogMonitorConfigIntsMixed(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "2"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Log(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigLogAlert(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "log alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").last(\"5m\") > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "enable_logs_sample", "true"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_NoThresholdWindows(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholdWindows(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "message", "test"),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_ThresholdWindows(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigThresholdWindows(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query",
						"avg(last_1h):anomalies(avg:system.cpu.system{name:cassandra}, 'basic', 3, direction='above', alert_window='last_5m', interval=20, count_default_zero='true') >= 1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.warning_recovery", "0.25"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "thresholds.critical_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "threshold_windows.recovery_window", "last_5m"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "threshold_windows.trigger_window", "last_5m"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_MuteUnmuteSpecificScopes(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigMuteSpecificScopes(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.%", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.host:myserver", "0"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUnmuteSpecificScopes(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "silenced"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_ComposeWithSyntheticsTest(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorComposeWithSyntheticsTest(monitorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_synthetics_test.foo", "name", monitorName),
					resource.TestCheckResourceAttrSet(
						"datadog_monitor.bar", "query"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorDestroy(accProvider *schema.Provider) func(*terraform.State) error {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient

		if err := destroyHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogMonitorExists(accProvider *schema.Provider, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		providerConf := accProvider.Meta().(*ProviderConfiguration)
		client := providerConf.CommunityClient
		if err := existsHelper(s, client); err != nil {
			return err
		}
		return nil
	}
}

func TestAccDatadogMonitor_SilencedRemove(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorSilenceZero(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "0"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorSilenceUnmute(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "-1"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_SilencedUpdateNoDiff(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorSilenceZero(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "0"),
				),
			},
			{
				Config:             testAccCheckDatadogMonitorSilenceZero(monitorName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccDatadogMonitor_SilencedUpdatePastTimestamp(t *testing.T) {
	accProviders, clock, cleanup := testAccProviders(t, initRecorder(t))
	monitorName := uniqueEntityName(clock, t)
	defer cleanup(t)
	accProvider := testAccProvider(t, accProviders)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    accProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorSilenceZero(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider, "datadog_monitor.foo"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "0"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorSilencePastTimestamp(monitorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "silenced.*", "1559759717",
					),
				),
			},
			{
				Config:             testAccCheckDatadogMonitorSilencePastTimestamp(monitorName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func testAccCheckDatadogMonitorSilenceZero(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
	name = "%s"
	type = "query alert"
	message = "some message Notify: @hipchat-channel"

	query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

	silenced = {
    "*" = 0
  }
}`, uniq)
}

func testAccCheckDatadogMonitorSilenceUnmute(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
	name = "%s"
	type = "query alert"
	message = "some message Notify: @hipchat-channel"

	query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

	silenced = {
    "*" = -1
  }
}`, uniq)
}

func testAccCheckDatadogMonitorSilencePastTimestamp(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
	name = "%s"
	type = "query alert"
	message = "some message Notify: @hipchat-channel"

	query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

	silenced = {
    "*" = 1559759717
  }
}`, uniq)
}

func testAccCheckDatadogMonitorConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  thresholds = {
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
}`, uniq)
}
func testAccCheckDatadogMonitorConfigNoThresholds(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
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
}`, uniq)
}

func testAccCheckDatadogMonitorServiceCheckConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "service check"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "\"custom.check\".over(\"environment:foo\").last(2).count_by_status()"

  thresholds = {
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
}`, uniq)
}

func testAccCheckDatadogMonitorConfigInts(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name               = "%s"
  type               = "query alert"
  message            = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  thresholds = {
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
}`, uniq)
}

func testAccCheckDatadogMonitorConfigIntsMixed(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name               = "%s"
  type               = "query alert"
  message            = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 3"

  thresholds = {
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
}`, uniq)
}

func testAccCheckDatadogMonitorConfigUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "a different message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated! @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"

  thresholds = {
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
  notify_audit = true
  timeout_h = 70
  include_tags = false
  require_full_window = false
  locked = true
  silenced = {
	"*" = 0
  }
  tags = ["baz:qux", "quux"]
}`, uniq)
}

func testAccCheckDatadogMonitorConfigUpdatedWithTagsRemoved(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "a different message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated! @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"

  thresholds = {
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
  notify_audit = true
  timeout_h = 70
  include_tags = false
  require_full_window = false
  locked = true
  silenced = {
	"*" = 0
  }
}`, uniq)
}

func testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "complex_metric_alert_example_monitor" {
  name = "%s"
  type = "metric alert"
  message = "a message"

  query = "change(min(last_1m),last_5m):sum:org.eclipse.jetty.servlet.ServletContextHandler.5xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.4xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.3xx_responses{example,framework:chronos} > 5"
}`, uniq)
}

func testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "complex_query_alert_example_monitor" {
  name = "%s"
  type = "query alert"
  message = "a message"

  query = "change(min(last_1m),last_5m):sum:org.eclipse.jetty.servlet.ServletContextHandler.5xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.4xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.3xx_responses{example,framework:chronos} > 5"
}`, uniq)
}

func testAccCheckDatadogMonitorConfigWhitespace(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
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
  thresholds = {
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
}`, uniq)
}

func testAccCheckDatadogMonitorConfigLogAlert(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "log alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").last(\"5m\") > 2"

  thresholds = {
	warning = "1.0"
	critical = "2.0"
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
}`, uniq)
}

func testAccCheckDatadogMonitorConfigNoThresholdWindows(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
	name = "%s"
	type = "query alert"
	message = "test"
	query = "avg(last_1h):anomalies(avg:system.cpu.system{name:cassandra}, 'basic', 2, direction='above') >= 1"
	thresholds = {
	  ok = "0.0"
	  warning = "0.5"
	  warning_recovery = "0.25"
	  critical = "1.0"
	  critical_recovery = "0.5"
	}

	notify_no_data = false
	renotify_interval = 60

	notify_audit = false
	timeout_h = 60
	include_tags = true
}`, uniq)
}

func testAccCheckDatadogMonitorConfigThresholdWindows(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
	name = "%s"
	type = "query alert"
	message = "some message Notify: @hipchat-channel"
	escalation_message = "the situation has escalated @pagerduty"
	query = "avg(last_1h):anomalies(avg:system.cpu.system{name:cassandra}, 'basic', 3, direction='above', alert_window='last_5m', interval=20, count_default_zero='true') >= 1"
	thresholds = {
	  ok = "0.0"
	  warning = "0.5"
	  warning_recovery = "0.25"
	  critical = "1.0"
	  critical_recovery = "0.5"
	}

	notify_no_data = false
	renotify_interval = 60

	notify_audit = false
	timeout_h = 60
	include_tags = true

	threshold_windows = {
		recovery_window = "last_5m"
		trigger_window = "last_5m"
	}
}`, uniq)
}

func testAccCheckDatadogMonitorConfigMuteSpecificScopes(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
    name = "%s"
    type = "metric alert"
    message = "test"

    query = "avg(last_5m):max:system.load.1{*} by {host} > 100"

    thresholds = {
        critical = 100
    }

    silenced = {
      "host:myserver" = 0
    }
}`, uniq)
}

func testAccCheckDatadogMonitorConfigUnmuteSpecificScopes(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
    name = "%s"
    type = "metric alert"
    message = "test"

    query = "avg(last_5m):max:system.load.1{*} by {host} > 100"

    thresholds = {
        critical = 100
    }
}`, uniq)
}

func testAccCheckDatadogMonitorComposeWithSyntheticsTest(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "test"

  query = "avg(last_5m):max:system.load.1{*} by {host} > 100"

  thresholds = {
      critical = 100
  }
}

resource "datadog_synthetics_test" "foo" {
	type = "api"

	request = {
		method = "GET"
		url = "https://docs.datadoghq.com"
		timeout = 60
	}

	assertions = [
		{
			type = "statusCode"
			operator = "isNot"
			target = "500"
		}
	]

	locations = [ "aws:eu-central-1" ]

	options = {
		tick_every = 900
		min_failure_duration = 10
		min_location_failed = 1
	}

	name = "%s"
	message = "Notify @pagerduty"
	tags = ["foo:bar", "foo", "env:test"]

	status = "live"
}

resource "datadog_monitor" "bar" {
  	name = "%s-composite"
  	type = "composite"
  	message = "test"
	query = "${datadog_monitor.foo.id} || ${datadog_synthetics_test.foo.monitor_id}"
}`, uniq, uniq, uniq)
}

func destroyHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetMonitor(i); err != nil {
			if strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			return fmt.Errorf("received an error retrieving monitor %s", err)
		}
		return fmt.Errorf("monitor still exists")
	}
	return nil
}

func existsHelper(s *terraform.State, client *datadog.Client) error {
	for _, r := range s.RootModule().Resources {
		i, _ := strconv.Atoi(r.Primary.ID)
		if _, err := client.GetMonitor(i); err != nil {
			return fmt.Errorf("received an error retrieving monitor %s", err)
		}
	}
	return nil
}
