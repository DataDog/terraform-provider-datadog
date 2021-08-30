package test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestDatadogMonitor_import(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_monitor.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigImported(monitorName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestDatadogMonitor_import_no_recovery(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_monitor.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigImportedNoRecovery(monitorName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestDatadogMonitor_importNoDataTimeFrame(t *testing.T) {
	t.Parallel()
	resourceName := "datadog_monitor.foo"
	ctx, accProviders := testAccProviders(context.Background(), t)
	monitorName := uniqueEntityName(ctx, t)
	accProvider := testAccProvider(t, accProviders)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: accProviders,
		CheckDestroy:      testAccCheckDatadogMonitorDestroy(accProvider),
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigImportedNoDataTimeFrame(monitorName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDatadogMonitorConfigImported(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2.5"

  monitor_thresholds {
    ok = 1.5
    warning           = 2.3
    warning_recovery  = 2.2
    critical          = 2.5
    critical_recovery = 2.4
  }

  notify_no_data = false
  new_group_delay = 500
  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "bar:baz"]
}`, uniq)
}

func testAccCheckDatadogMonitorConfigImportedNoRecovery(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2.5"

  monitor_thresholds {
    ok = 1.5
    warning           = 2.3
    critical          = 2.5
  }

  notify_no_data = false
  new_host_delay = 600
  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "bar:baz"]
}`, uniq)
}

func testAccCheckDatadogMonitorConfigImportedNoDataTimeFrame(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2.5"

  monitor_thresholds {
    ok = 1.5
    warning           = 2.3
    warning_recovery  = 2.2
    critical          = 2.5
    critical_recovery = 2.4
  }

  notify_no_data = true
  no_data_timeframe = 300
  new_host_delay = 600
  renotify_interval = 60

  notify_audit = false
  timeout_h = 60
  include_tags = true
  require_full_window = true
  locked = false
  tags = ["foo:bar", "bar:baz"]
}`, uniq)
}
