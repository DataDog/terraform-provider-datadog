package datadog

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestDatadogMonitor_import(t *testing.T) {
	resourceName := "datadog_monitor.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigImported,
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

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckDatadogMonitorDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckDatadogMonitorConfigImportedNoRecovery,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

const testAccCheckDatadogMonitorConfigImported = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
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
}
`

const testAccCheckDatadogMonitorConfigImportedNoRecovery = `
resource "datadog_monitor" "foo" {
  name = "name for monitor foo"
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
}
`
