package datadog

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestDatadogMonitor_import(t *testing.T) {
	resourceName := "datadog_monitor.foo"
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
	resourceName := "datadog_monitor.foo"
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

func testAccCheckDatadogMonitorConfigImported(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2.5"

  thresholds = {
    ok = 1.5
    warning           = 2.3
    warning_recovery  = 2.2
    critical          = 2.5
    critical_recovery = 2.4
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

func testAccCheckDatadogMonitorConfigImportedNoRecovery(uniq string) string {
	return fmt.Sprintf(`
resource "datadog_monitor" "foo" {
  name = "%s"
  type = "query alert"
  message = "some message Notify: @hipchat-channel"
  escalation_message = "the situation has escalated @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2.5"

  thresholds = {
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
