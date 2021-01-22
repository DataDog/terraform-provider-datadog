---
page_title: "datadog_monitor Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.
---

# Resource `datadog_monitor`

Provides a Datadog monitor resource. This can be used to create and manage Datadog monitors.

## Example Usage

```terraform
# Create a new Datadog monitor
resource "datadog_monitor" "foo" {
  name               = "Name for monitor foo"
  type               = "metric alert"
  message            = "Monitor triggered. Notify: @hipchat-channel"
  escalation_message = "Escalation message @pagerduty"

  query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"

  monitor_thresholds {
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

  # ignore any changes in silenced value; using silenced is deprecated in favor of downtimes
  lifecycle {
    ignore_changes = [silenced]
  }

  tags = ["foo:bar", "baz"]
}
```

## Schema

### Required

- **message** (String, Required) A message to include with notifications for this monitor.

Email notifications can be sent to specific users by using the same `@username` notation as events.
- **name** (String, Required) Name of Datadog monitor.
- **query** (String, Required) The monitor query to notify on. Note this is not the same query you see in the UI and the syntax is different depending on the monitor type, please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. `terraform plan` will validate query contents unless `validate` is set to `false`.
- **type** (String, Required) The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation page](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor). Note: The monitor type cannot be changed after a monitor is created.

### Optional

- **enable_logs_sample** (Boolean, Optional) A boolean indicating whether or not to include a list of log values which triggered the alert. This is only used by log monitors. Defaults to `false`.
- **escalation_message** (String, Optional) A message to include with a re-notification. Supports the `@username` notification allowed elsewhere.
- **evaluation_delay** (Number, Optional) (Only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.

For example, if the value is set to `300` (5min), the `timeframe` is set to `last_5m` and the time is 7:00, the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled metrics to ensure the monitor will always have data during evaluation.
- **force_delete** (Boolean, Optional) A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. SLO, composite monitor).
- **id** (String, Optional) The ID of this resource.
- **include_tags** (Boolean, Optional) A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title. Defaults to `true`.
- **locked** (Boolean, Optional) A boolean indicating whether changes to to this monitor should be restricted to the creator or admins. Defaults to `false`.
- **monitor_threshold_windows** (Block List, Max: 1) A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors. (see [below for nested schema](#nestedblock--monitor_threshold_windows))
- **monitor_thresholds** (Block List, Max: 1) Alert thresholds of the monitor. (see [below for nested schema](#nestedblock--monitor_thresholds))
- **new_host_delay** (Number, Optional) Time (in seconds) to allow a host to boot and applications to fully start before starting the evaluation of monitor results. Should be a non negative integer. Defaults to `300`.
- **no_data_timeframe** (Number, Optional) The number of minutes before a monitor will notify when data stops reporting. Provider defaults to 10 minutes.

We recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.
- **notify_audit** (Boolean, Optional) A boolean indicating whether tagged users will be notified on changes to this monitor. Defaults to `false`.
- **notify_no_data** (Boolean, Optional) A boolean indicating whether this monitor will notify when data stops reporting. Defaults to `false`.
- **priority** (Number, Optional) Integer from 1 (high) to 5 (low) indicating alert severity.
- **renotify_interval** (Number, Optional) The number of minutes after the last notification before a monitor will re-notify on the current status. It will only re-notify if it's not resolved.
- **require_full_window** (Boolean, Optional) A boolean indicating whether this monitor needs a full window of data before it's evaluated.

We highly recommend you set this to `false` for sparse metrics, otherwise some evaluations will be skipped. Default: `true` for `on average`, `at all times` and `in total` aggregation. `false` otherwise.
- **silenced** (Map of Number, Optional, Deprecated) Each scope will be muted until the given POSIX timestamp or forever if the value is `0`. Use `-1` if you want to unmute the scope. Deprecated: the silenced parameter is being deprecated in favor of the downtime resource. This will be removed in the next major version of the Terraform Provider.
- **tags** (Set of String, Optional) A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API
- **threshold_windows** (Map of String, Optional, Deprecated) A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. Can only be used for, and are required for, anomaly monitors.
- **thresholds** (Map of String, Optional, Deprecated) Alert thresholds of the monitor.
- **timeout_h** (Number, Optional) The number of hours of the monitor not reporting data before it will automatically resolve from a triggered state.
- **validate** (Boolean, Optional) If set to `false`, skip the validation call done during plan.

<a id="nestedblock--monitor_threshold_windows"></a>
### Nested Schema for `monitor_threshold_windows`

Optional:

- **recovery_window** (String, Optional) Describes how long an anomalous metric must be normal before the alert recovers.
- **trigger_window** (String, Optional) Describes how long a metric must be anomalous before an alert triggers.


<a id="nestedblock--monitor_thresholds"></a>
### Nested Schema for `monitor_thresholds`

Optional:

- **critical** (String, Optional) The monitor `CRITICAL` recovery threshold. Must be a number.
- **critical_recovery** (String, Optional) The monitor `CRITICAL` recovery threshold. Must be a number.
- **ok** (String, Optional) The monitor `OK` threshold. Must be a number.
- **unknown** (String, Optional) The monitor `UNKNOWN` threshold. Must be a number.
- **warning** (String, Optional) The monitor `WARNING` threshold. Must be a number.
- **warning_recovery** (String, Optional) The monitor `WARNING` recovery threshold. Must be a number.

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_monitor.bytes_received_localhost 2081
```
