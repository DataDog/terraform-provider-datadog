package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDatadogMonitor_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "new_group_delay", "500"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "3"),
				),
			},
		},
	})
}

func TestAccDatadogMonitorServiceCheck_Basic(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorServiceCheckConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.unknown", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "3"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_BasicNoTreshold(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholds(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "bar:baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Updated(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	monitorNameUpdated := monitorName + "-updated"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "new_group_delay", "500"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "300"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
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
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "3"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "2.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "70"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz:qux"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "quux"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "1"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "metric alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_UpdatedToRemoveTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	monitorNameUpdated := monitorName + "-updated"
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfig(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "new_group_delay", "500"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "300"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
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
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigUpdatedWithPriorityRemoved(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "2.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "70"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "locked", "true"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "tags.#"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", "0"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_metric_alert_example_monitor", "type", "metric alert"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(monitorNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.complex_query_alert_example_monitor", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_TrimWhitespace(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigWhitespace(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1.5"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Basic_float_int(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigInts(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "1"),
				),
			},

			{
				Config: testAccCheckDatadogMonitorConfigIntsMixed(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "3"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "2"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_Log(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigLogAlert(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "log alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").by(\"source\").last(\"5m\") > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "enable_logs_sample", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "groupby_simple_monitor", "true"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_NoThresholdWindows(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigNoThresholdWindows(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "message", "test"),
					resource.TestCheckResourceAttr("datadog_monitor.foo", "type", "query alert"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_ThresholdWindows(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigThresholdWindows(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
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
						"datadog_monitor.foo", "monitor_thresholds.0.ok", "0"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning_recovery", "0.25"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical_recovery", "0.5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_audit", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "timeout_h", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_threshold_windows.0.recovery_window", "last_5m"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_threshold_windows.0.trigger_window", "last_5m"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_ComposeWithSyntheticsTest(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
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

func testAccCheckDatadogMonitorDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := destroyHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogMonitorExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		datadogClientV1 := providerConf.DatadogClientV1
		authV1 := providerConf.AuthV1

		if err := existsHelper(authV1, s, datadogClientV1); err != nil {
			return err
		}
		return nil
	}
}

func TestAccDatadogMonitor_ZeroDelay(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigZeroDelay(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "query alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "0"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_RestrictedRoles(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigRestrictedRoles(monitorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "restricted_roles.#", "1"),
				),
			},
			{
				Config: testAccCheckDatadogMonitorConfigRestrictedRolesRemoved(monitorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "restricted_roles"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorConfig(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"
  priority = 3

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
	warning = "1.0"
	critical = "2.0"
	warning_recovery = "0.5"
	critical_recovery = "1.5"
  }

  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  new_group_delay = 500
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
  priority = 3

  query = "\"custom.check\".over(\"environment:foo\").last(2).count_by_status()"

  monitor_thresholds {
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

  monitor_thresholds {
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

  monitor_thresholds {
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
  priority = 1

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"

  monitor_thresholds {
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
  tags = ["baz:qux", "quux"]
}`, uniq)
}

func testAccCheckDatadogMonitorConfigUpdatedWithPriorityRemoved(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "a different message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated! @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:bar,host:bar} by {host} > 3"

  monitor_thresholds {
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
}`, uniq)
}

func testAccCheckDatadogMonitorConfigMetricAlertNotUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "complex_metric_alert_example_monitor" {
  name = "%s"
  type = "metric alert"
  message = "a message"
  priority = 3

  query = "change(min(last_1m),last_5m):sum:org.eclipse.jetty.servlet.ServletContextHandler.5xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.4xx_responses{example,framework:chronos} + sum:org.eclipse.jetty.servlet.ServletContextHandler.3xx_responses{example,framework:chronos} > 5"
}`, uniq)
}

func testAccCheckDatadogMonitorConfigQueryAlertNotUpdated(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "complex_query_alert_example_monitor" {
  name = "%s"
  type = "query alert"
  message = "a message"
  priority = 3

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
  monitor_thresholds {
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

  query = "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").by(\"source\").last(\"5m\") > 2"

  monitor_thresholds {
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
	groupby_simple_monitor = true
}`, uniq)
}

func testAccCheckDatadogMonitorConfigNoThresholdWindows(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
	name = "%s"
	type = "query alert"
	message = "test"
	query = "avg(last_1h):anomalies(avg:system.cpu.system{name:cassandra}, 'basic', 2, direction='above') >= 1"
	monitor_thresholds {
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
	monitor_thresholds {
		ok = 0
		warning = 0.5
		warning_recovery = 0.25
		critical = 1
		critical_recovery = 0.5
	}

	notify_no_data = false
	renotify_interval = 60

	notify_audit = false
	timeout_h = 60
	include_tags = true

	monitor_threshold_windows {
		recovery_window = "last_5m"
		trigger_window = "last_5m"
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

  monitor_thresholds {
      critical = 100
  }
}

resource "datadog_synthetics_test" "foo" {
	type = "api"

	request_definition {
		method = "GET"
		url = "https://docs.datadoghq.com"
		timeout = 60
	}

	assertion {
		type = "statusCode"
		operator = "isNot"
		target = "500"
	}

	locations = [ "aws:eu-central-1" ]

	options_list {
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

func testAccCheckDatadogMonitorConfigZeroDelay(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
	warning = "1.0"
	critical = "2.0"
	warning_recovery = "0.5"
	critical_recovery = "1.5"
  }

  new_host_delay = 0
}`, uniq)
}

func testAccCheckDatadogMonitorConfigRestrictedRoles(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_role" "foo" {
  name      = "%s"
}

resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
	critical = "2.0"
  }

  restricted_roles = ["${datadog_role.foo.id}"]
}`, uniq, uniq)
}

func testAccCheckDatadogMonitorConfigRestrictedRolesRemoved(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
	critical = "2.0"
  }
}`, uniq)
}

func destroyHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	err := utils.Retry(2, 10, func() error {
		for _, r := range s.RootModule().Resources {
			i, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			_, httpresp, err := datadogClientV1.MonitorsApi.GetMonitor(ctx, i)
			if err != nil {
				if httpresp != nil && httpresp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		}
		return nil
	})
	return err
}

func existsHelper(ctx context.Context, s *terraform.State, datadogClientV1 *datadogV1.APIClient) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_monitor" {
			continue
		}
		i, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
		_, httpresp, err := datadogClientV1.MonitorsApi.GetMonitor(ctx, i)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error retrieving monitor")
		}
	}
	return nil
}
