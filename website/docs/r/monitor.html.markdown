---
layout: "datadog"
page_title: "Datadog: datadog_monitor"
sidebar_current: "docs-datadog-resource-monitor"
description: |-
  Provides a Datadog monitor resource. This can be used to create and manage monitors.
---

# datadog_monitor

Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.

## Example Usage

```hcl
# Create a new Datadog monitor
resource "datadog_monitor" "foo" {
  name               = "Name for monitor foo"
  type               = "metric alert"
  message            = "Monitor triggered. Notify: @hipchat-channel"
  escalation_message = "Escalation message @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"

  thresholds {
    ok                = 0
    warning           = 2
    warning_recovery  = 1
    critical          = 4
    critical_recovery = 3
  }

  notify_no_data    = false
  renotify_interval = 60

  notify_audit = false
  timeout_h    = 60
  include_tags = true

  silenced {
    "*" = 0
  }

  tags = ["foo:bar", "baz"]
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The type of the monitor, chosen from:
    * `metric alert`
    * `service check`
    * `event alert`
    * `query alert`
    * `composite`
    * `log alert`    
* `name` - (Required) Name of Datadog monitor
* `query` - (Required) The monitor query to notify on. Note this is not the same query you see in the UI and
    the syntax is different depending on the monitor `type`, please see the [API Reference](https://docs.datadoghq.com/api/?lang=python#create-a-monitor) for details. **Warning:** `terraform plan` won't perform any validation of the query contents.
* `message` - (Required) A message to include with notifications for this monitor.
    Email notifications can be sent to specific users by using the same '@username' notation as events.
* `escalation_message` - (Optional) A message to include with a re-notification. Supports the '@username'
    notification allowed elsewhere.
* `thresholds` - (Optional)
    * Metric alerts:
    A dictionary of thresholds by threshold type. Currently we have four threshold types for metric alerts: critical, critical recovery, warning, and warning recovery. Critical is defined in the query, but can also be specified in this option. Warning and recovery thresholds can only be specified using the thresholds option.
    Example usage:
        ```
        thresholds {
            critical          = 90
            critical_recovery = 85
            warning           = 80
            warning_recovery  = 75
        }
        ```
    **Warning:** the `critical` threshold value must match the one contained in the `query` argument. The `threshold` from the previous example is valid along with a query like `avg(last_1h):avg:system.disk.in_use{role:sqlserver} by {host} > 90` but
    along with something like `avg(last_1h):avg:system.disk.in_use{role:sqlserver} by {host} > 95` would make the Datadog API return a HTTP error 400, complaining "The value provided for parameter 'query' is invalid".
    * Service checks:
    A dictionary of thresholds by status. Because service checks can have multiple thresholds, we don't define them directly in the query.
    Default values:
        ```
        thresholds {
            ok       = 1
            critical = 1
            warning  = 1
            unknown  = 1
        }
        ```

* `notify_no_data` (Optional) A boolean indicating whether this monitor will notify when data stops reporting. Defaults
    to false.
* `new_host_delay` (Optional) Time (in seconds) to allow a host to boot and
    applications to fully start before starting the evaluation of monitor
    results. Should be a non negative integer. Defaults to 300.
* `evaluation_delay` (Optional, only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.
    For example, if the value is set to 300 (5min), the timeframe is set to last_5m and the time is 7:00,
    the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled
    metrics to ensure the monitor will always have data during evaluation.
* `no_data_timeframe` (Optional) The number of minutes before a monitor will notify when data stops reporting. Must be at
    least 2x the monitor timeframe for metric alerts or 2 minutes for service checks. Default: 2x timeframe for
    metric alerts, 2 minutes for service checks.
* `renotify_interval` (Optional) The number of minutes after the last notification before a monitor will re-notify
    on the current status. It will only re-notify if it's not resolved.
* `notify_audit` (Optional) A boolean indicating whether tagged users will be notified on changes to this monitor.
    Defaults to false.
* `timeout_h` (Optional) The number of hours of the monitor not reporting data before it will automatically resolve
    from a triggered state. Defaults to false.
* `include_tags` (Optional) A boolean indicating whether notifications from this monitor will automatically insert its
* `enable_logs_sample` (Optional) A boolean indicating whether or not to include a list of log values which triggered the alert. Defaults to false. This is only used by log monitors.
    triggering tags into the title. Defaults to true.
* `require_full_window` (Optional) A boolean indicating whether this monitor needs a full window of data before it's evaluated.
    We highly recommend you set this to False for sparse metrics, otherwise some evaluations will be skipped.
    Default: True for "on average", "at all times" and "in total" aggregation. False otherwise.
* `locked` (Optional) A boolean indicating whether changes to to this monitor should be restricted to the creator or admins. Defaults to False.
* `tags` (Optional) A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API
* `silenced` (Optional) Each scope will be muted until the given POSIX timestamp or forever if the value is 0.
    To mute the alert completely:

        silenced {
          "*" =  0
        }

    To mute role:db for a short time:

        silenced {
          "role:db" = 1412798116
        }

## Attributes Reference

The following attributes are exported:

* `id` - ID of the Datadog monitor

## Import

Monitors can be imported using their numeric ID, e.g.

```
$ terraform import datadog_monitor.bytes_received_localhost 2081
```
