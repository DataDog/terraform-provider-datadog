package test

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/terraform-providers/terraform-provider-datadog/datadog"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
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
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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
						"datadog_monitor.foo", "query", `"custom.check".by("environment:foo").last(2).count_by_status()`),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "new_host_delay", "600"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "evaluation_delay", "700"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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

func TestAccDatadogMonitor_BasicNoTresholdOrPriority(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorConfigNoThresholdsOrPriority(monitorName),
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
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "bar:baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", ""),
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
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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
						"datadog_monitor.foo", "timeout_h", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
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
						"datadog_monitor.foo", "notification_preset_name", "show_all"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "40"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "0"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#"),
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
						"datadog_monitor.foo", "timeout_h", "10"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
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
						"datadog_monitor.foo", "notification_preset_name", "hide_query"),
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
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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
						"datadog_monitor.foo", "timeout_h", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "true"),
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
						"datadog_monitor.foo", "notification_preset_name", "show_all"),
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
						"datadog_monitor.foo", "renotify_occurrences", "0"),
					resource.TestCheckNoResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#"),
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
						"datadog_monitor.foo", "timeout_h", "10"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "include_tags", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "require_full_window", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "priority", ""),
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
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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
						"datadog_monitor.foo", "on_missing_data", ""),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "group_retention_duration", ""),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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

func TestAccDatadogMonitor_LogMultiAlert(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorConfigLogMultiAlert(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "message", "some message Notify: @hipchat-channel"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "type", "log alert"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "query", "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").by(\"source,status\").last(\"5m\") > 2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_no_data", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "on_missing_data", "show_no_data"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "group_retention_duration", "2d"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_interval", "60"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.warning", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "monitor_thresholds.0.critical", "2"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "enable_logs_sample", "true"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "groupby_simple_monitor", "false"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_by.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "notify_by.0", "status"),
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
						"datadog_monitor.foo", "renotify_occurrences", "5"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "renotify_statuses.#", "2"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "alert"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "renotify_statuses.*", "warn"),
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
						"datadog_monitor.foo", "timeout_h", "1"),
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

func TestAccDatadogMonitor_FormulaFunction(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorFormulaFunction(monitorName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "variables.#", "1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "variables.0.event_query.#", "2"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorDestroy(accProvider func() (*schema.Provider, error)) func(*terraform.State) error {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := destroyMonitorHelper(auth, s, apiInstances); err != nil {
			return err
		}
		return nil
	}
}

func testAccCheckDatadogMonitorExists(accProvider func() (*schema.Provider, error)) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		provider, _ := accProvider()
		providerConf := provider.Meta().(*datadog.ProviderConfiguration)
		apiInstances := providerConf.DatadogApiInstances
		auth := providerConf.Auth

		if err := existsHelper(auth, s, apiInstances); err != nil {
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
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "restricted_roles.#", "0"),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_SchedulingOptions(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorWithSchedulingOptions(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.evaluation_window.0.day_starts", "04:00"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.evaluation_window.0.month_starts", "1"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorWithSchedulingOptions(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "a message"
  priority = 3

  query = "avg(current_1mo):avg:system.load.5{*} > 0.5"

  monitor_thresholds {
	critical = "0.5"
  }

  scheduling_options {
	evaluation_window {
	  day_starts   = "04:00"
	  month_starts = 1
	}
  }
}`, uniq)
}

func testAccCheckDatadogMonitorWithEmptySchedulingOptions(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%[1]s"
  type = "metric alert"
  message = "%[1]s"
  priority = 3

  query = "avg(last_1h):avg:system.load.5{*} > 4"

  monitor_thresholds {
	critical = "4"
  }

  scheduling_options {}
}`, uniq)
}

func TestAccDatadogMonitor_EmptySchedulingOptions(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorWithEmptySchedulingOptions(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
				),
			},
		},
	})
}

func TestAccDatadogMonitor_SchedulingOptionsHourStart(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorWithSchedulingOptionsHourStart(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.evaluation_window.0.hour_starts", "0"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorWithSchedulingOptionsHourStart(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "a message"
  priority = 3

  query = "avg(current_1h):avg:system.load.5{*} > 0.5"

  monitor_thresholds {
	critical = "0.5"
  }

  scheduling_options {
	evaluation_window {
	  hour_starts = "0"
	}
  }
}`, uniq)
}

func TestAccDatadogMonitor_SchedulingOptionsCustomSchedule(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorWithSchedulingOptionsCustomSchedule(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.timezone", "America/New_York"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.start", "2023-11-10T12:31:00"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorWithSchedulingOptionsCustomSchedule(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "a message"
  priority = 3

  query = "avg(last_1h):avg:system.load.5{*} > 0.5"

  monitor_thresholds {
	critical = "0.5"
  }

  scheduling_options {
	custom_schedule {
		recurrence {
			rrule = "FREQ=DAILY;INTERVAL=1"
			timezone = "America/New_York"
			start = "2023-11-10T12:31:00"
		}
	}
  }
}`, uniq)
}

func TestAccDatadogMonitor_SchedulingOptionsCustomScheduleNoStart(t *testing.T) {
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
				Config: testAccCheckDatadogMonitorWithSchedulingOptionsCustomSchedule(monitorName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatadogMonitorExists(accProvider),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "name", monitorName),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.rrule", "FREQ=DAILY;INTERVAL=1"),
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "scheduling_options.0.custom_schedule.0.recurrence.0.timezone", "America/New_York"),
				),
			},
		},
	})
}

func testAccCheckDatadogMonitorWithSchedulingOptionsCustomScheduleNoStart(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "metric alert"
  message = "a message"
  priority = 3
  
  query = "avg(last_1h):avg:system.load.5{*} > 0.5"

  monitor_thresholds {
	critical = "0.5"
  }

  scheduling_options {
	custom_schedule {
		recurrence {
			rrule = "FREQ=DAILY;INTERVAL=1"
			timezone = "America/New_York"
		}
	}
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
  priority = 3

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  monitor_thresholds {
	warning = "1.0"
	critical = "2.0"
	warning_recovery = "0.5"
	critical_recovery = "1.5"
  }

  renotify_interval = 60
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  new_group_delay = 500
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  tags = ["foo:bar", "baz"]
  notification_preset_name = "hide_query"
}`, uniq)
}

func testAccCheckDatadogMonitorConfigDuplicateTags(uniq string) string {
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
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  new_group_delay = 500
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  tags = ["foo:bar", "baz", "foo:thebar"]
  notification_preset_name = "hide_query"
}`, uniq)
}

func testAccCheckDatadogMonitorConfigNoTag(uniq string) string {
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
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  new_group_delay = 500
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  notification_preset_name = "hide_query"
}`, uniq)
}

func testAccCheckDatadogMonitorConfigNoThresholdsOrPriority(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2"

  notify_no_data = false
  renotify_interval = 60
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  include_tags = true
  require_full_window = true
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

  query = "\"custom.check\".by(\"environment:foo\").last(2).count_by_status()"

  monitor_thresholds {
	warning = 1
	critical = 1
	unknown = 1
	ok = 1
  }

  renotify_interval = 60
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
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
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit        = false
  timeout_h           = 1
  include_tags        = true
  require_full_window = true

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
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit        = false
  timeout_h           = 1
  include_tags        = true
  require_full_window = true

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
  timeout_h = 10
  include_tags = false
  require_full_window = false
  tags = ["baz:qux", "quux"]
  notification_preset_name = "show_all"
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
  timeout_h = 10
  include_tags = false
  require_full_window = false
  notification_preset_name = "show_all"
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
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
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
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  tags = ["foo:bar", "baz"]
  enable_logs_sample = true
  groupby_simple_monitor = true
}`, uniq)
}

func testAccCheckDatadogMonitorConfigLogMultiAlert(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "log alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "logs(\"service:foo AND type:error\").index(\"main\").rollup(\"count\").by(\"source,status\").last(\"5m\") > 2"

  monitor_thresholds {
	warning = "1.0"
	critical = "2.0"
  }

  renotify_interval = 60
  renotify_occurrences = 5
  renotify_statuses = ["alert", "warn"]

  notify_audit = false
  timeout_h = 1
  group_retention_duration = "2d"
  on_missing_data = "show_no_data"
  new_host_delay = 600
  evaluation_delay = 700
  include_tags = true
  require_full_window = true
  tags = ["foo:bar", "baz"]
  enable_logs_sample = true
  notify_by = ["status"]
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
	renotify_occurrences = 5
	renotify_statuses = ["alert", "warn"]

	notify_audit = false
	timeout_h = 1
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
	renotify_occurrences = 5
	renotify_statuses = ["alert", "warn"]

	notify_audit = false
	timeout_h = 1
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

func testAccCheckDatadogMonitorFormulaFunction(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "rum alert"
  message = "test"

  query = "formula(\"var1 + var2\").last(\"5m\") > 100"

  variables {
  	event_query {
		data_source = "rum"
		name        = "var1"
		indexes     = ["*"]
		compute {
			aggregation = "count"
		}
		search {
			query = "@browser.name:Firefox"
		}
	}
	event_query {
		data_source = "rum"
		name        = "var2"
		indexes     = ["*"]
		compute {
			aggregation = "count"
		}
		search {
			query = "@browser.name:Chrome"
		}
	}
  }

  monitor_thresholds {
      critical = 100
  }
}
`, uniq)
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

func destroyMonitorHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_monitor" {
			continue
		}

		err := utils.Retry(2, 10, func() error {
			i, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
			_, httpresp, err := apiInstances.GetMonitorsApiV1().GetMonitor(ctx, i)
			if err != nil {
				if httpresp != nil && httpresp.StatusCode == 404 {
					return nil
				}
				return &utils.RetryableError{Prob: fmt.Sprintf("received an error retrieving Monitor %s", err)}
			}
			return &utils.RetryableError{Prob: "Monitor still exists"}
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func existsHelper(ctx context.Context, s *terraform.State, apiInstances *utils.ApiInstances) error {
	for _, r := range s.RootModule().Resources {
		if r.Type != "datadog_monitor" {
			continue
		}
		i, _ := strconv.ParseInt(r.Primary.ID, 10, 64)
		_, httpresp, err := apiInstances.GetMonitorsApiV1().GetMonitor(ctx, i)
		if err != nil {
			return utils.TranslateClientError(err, httpresp, "error retrieving monitor")
		}
	}
	return nil
}

func TestAccDatadogMonitor_DefaultTags(t *testing.T) {
	t.Parallel()
	ctx, accProviders := testAccProviders(context.Background(), t)
	accProvider := testAccProvider(t, accProviders)
	resource.Test(t, resource.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{ // New tags are correctly added and duplicates are kept
				Config: testAccCheckDatadogMonitorConfigDuplicateTags(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "4"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:thebar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "default_key:default_value"),
				),
			},
			{ // Resource tags take precedence over default tags and duplicates stay
				Config: testAccCheckDatadogMonitorConfigDuplicateTags(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"foo": "not_bar",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:thebar"),
				),
			},
			{ // Resource tags take precedence over default tags, but new tags are added
				Config: testAccCheckDatadogMonitorConfig(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"foo":     "not_bar",
						"new_tag": "new_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "new_tag:new_value"),
				),
			},
			{ // Tags without any value work correctly
				Config: testAccCheckDatadogMonitorConfig(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"no_value": "",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "no_value"),
				),
			},
			{ // Tags with colons in the value work correctly
				Config: testAccCheckDatadogMonitorConfig(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"repo_url": "https://github.com/repo/path",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"datadog_monitor.foo", "tags.#", "3"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "baz"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "foo:bar"),
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "repo_url:https://github.com/repo/path"),
				),
			},
			{ // Works with monitors without a tag attribute
				Config: testAccCheckDatadogMonitorConfigNoTag(uniqueEntityName(ctx, t)),
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"datadog": withDefaultTags(accProvider, map[string]interface{}{
						"default_key": "default_value",
					}),
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttr(
						"datadog_monitor.foo", "tags.*", "default_key:default_value"),
				),
			},
		},
	})
}
