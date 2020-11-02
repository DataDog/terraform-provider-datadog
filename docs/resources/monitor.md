---
page_title: "datadog_monitor"
---

# datadog_monitor Resource

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

  thresholds = {
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

  # ignore any changes in silenced value; using silenced is deprecated in favor of downtimes
  lifecycle {
    ignore_changes = [silenced]
  }

  tags = ["foo:bar", "baz"]
}
```

## Argument Reference

The following arguments are supported:

-   `type`: (Required) The type of the monitor. The mapping from these types to the types found in the Datadog Web UI can be found in the Datadog API [documentation](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) page. The available options are below. **Note**: The monitor type cannot be changed after a monitor is created.
    -   `metric alert`
    -   `service check`
    -   `event alert`
    -   `query alert`
    -   `composite`
    -   `log alert`
    -   `process alert`
    -   `rum alert`
    -   `synthetics alert`
    -   `trace-analytics alert`
    -   `slo alert`
-   `name`: (Required) Name of Datadog monitor
-   `query`: (Required) The monitor query to notify on. Note this is not the same query you see in the UI and

    the syntax is different depending on the monitor `type` , please see the [API Reference](https://docs.datadoghq.com/api/v1/monitors/#create-a-monitor) for details. **Warning:** `terraform plan` won't perform any validation of the query contents.

-   `message`: (Required) A message to include with notifications for this monitor.

    Email notifications can be sent to specific users by using the same '@username' notation as events.

-   `escalation_message`: (Optional) A message to include with a re-notification. Supports the '@username'

    notification allowed elsewhere.

-   `thresholds`: (Optional)

    -   Metric alerts:

    A dictionary of thresholds by threshold type. Currently we have four threshold types for metric alerts: critical, critical recovery, warning, and warning recovery. Critical is defined in the query, but can also be specified in this option. Warning and recovery thresholds can only be specified using the thresholds option. Example usage:

          thresholds = {
              critical          = 90
              critical_recovery = 85
              warning           = 80
              warning_recovery  = 75
          }

    **Warning:** the `critical` threshold value must match the one contained in the `query` argument. The `threshold` from the previous example is valid along with a query like `avg(last_1h):avg:system.disk.in_use{role:sqlserver} by {host} > 90` but along with something like `avg(last_1h):avg:system.disk.in_use{role:sqlserver} by {host} > 95` would make the Datadog API return a HTTP error 400, complaining "The value provided for parameter 'query' is invalid".

    -   Service checks:

    A dictionary of thresholds by status. Because service checks can have multiple thresholds, we don't define them directly in the query. Default values:

          thresholds = {
              ok       = 1
              critical = 1
              warning  = 1
              unknown  = 1
          }

*   `notify_no_data` (Optional) A boolean indicating whether this monitor will notify when data stops reporting. Defaults

    to false.

*   `new_host_delay` (Optional) Time (in seconds) to allow a host to boot and

    applications to fully start before starting the evaluation of monitor results. Should be a non negative integer. Defaults to 300.

*   `evaluation_delay` (Optional, only applies to metric alert) Time (in seconds) to delay evaluation, as a non-negative integer.

    For example, if the value is set to 300 (5min), the timeframe is set to last_5m and the time is 7:00, the monitor will evaluate data from 6:50 to 6:55. This is useful for AWS CloudWatch and other backfilled metrics to ensure the monitor will always have data during evaluation.

*   `no_data_timeframe` (Optional) The number of minutes before a monitor will notify when data stops reporting. Provider defaults to 10 minutes.

    We recommend at least 2x the monitor timeframe for metric alerts or 2 minutes for service checks.

*   `renotify_interval` (Optional) The number of minutes after the last notification before a monitor will re-notify

    on the current status. It will only re-notify if it's not resolved.

*   `notify_audit` (Optional) A boolean indicating whether tagged users will be notified on changes to this monitor.

    Defaults to false.

*   `timeout_h` (Optional) The number of hours of the monitor not reporting data before it will automatically resolve

    from a triggered state. Defaults to false.

*   `include_tags` (Optional) A boolean indicating whether notifications from this monitor automatically insert its triggering tags into the title. Defaults to true.
*   `enable_logs_sample` (Optional) A boolean indicating whether or not to include a list of log values which triggered the alert. Defaults to false. This is only used by log monitors.

    triggering tags into the title. Defaults to true.

*   `require_full_window` (Optional) A boolean indicating whether this monitor needs a full window of data before it's evaluated.

    We highly recommend you set this to False for sparse metrics, otherwise some evaluations will be skipped. Default: True for "on average", "at all times" and "in total" aggregation. False otherwise.

*   `locked` (Optional) A boolean indicating whether changes to to this monitor should be restricted to the creator or admins. Defaults to False.
*   `tags` (Optional) A list of tags to associate with your monitor. This can help you categorize and filter monitors in the manage monitors page of the UI. Note: it's not currently possible to filter by these tags when querying via the API
*   `force_delete` (Optional) A boolean indicating whether this monitor can be deleted even if it’s referenced by other resources (e.g. SLO, composite monitor).
*   `threshold_windows` (Optional) A mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m` . Can only be used for, and are required for, anomaly monitors.
    -   `recovery_window` describes how long an anomalous metric must be normal before the alert recovers.
    -   `trigger_window` describes how long a metric must be anomalous before an alert triggers.
*   `validate` (Optional) If set to false, skip the validation call done during `plan` .
*   `silenced` (Optional) Each scope will be muted until the given POSIX timestamp or forever if the value is 0. Use `-1` if you want to unmute the scope. **Deprecated** The `silenced` parameter is being deprecated in favor of the downtime resource. This will be removed in the next major version of the Terraform Provider.

    To mute the alert completely:

          silenced = {
            "*" =  0
          }

    To mute role:db for a short time:

          silenced = {
            "role:db" = 1412798116
          }

    Note: due to [HCL limitations](https://github.com/hashicorp/terraform/issues/2042), it is impossible to use interpolations in keys. For example, the following will result in muting the scope `role:${var:role}` (no interpolation is done):

          silenced = {
              "role:${var:role}" = 0
          }

    To workaround this, you can use the [map function](https://www.terraform.io/docs/configuration/functions/map.html) of HCL:

          silenced = ${map("role:${var:role}", 0)}

## Silencing by Hand and by Downtimes

There are two ways how to silence a single monitor:

-   Mute it by hand
-   Create a Downtime

Both of these actions add a new value to the `silenced` map. This can be problematic if the `silenced` attribute doesn't contain them in your Terraform, as they would be removed on next `terraform apply` invocation. In order to prevent that from happening, you can add following to your monitor:

```hcl
lifecycle {
  ignore_changes = [silenced]
}
```

The above will make sure that any changes to the `silenced` attribute are ignored.

This issue doesn't apply to multi-monitor downtimes (those that don't contain `monitor_id` ), as these don't influence contents of the `silenced` attribute.

## Attributes Reference

The following attributes are exported:

-   `id`: ID of the Datadog monitor

## Import

Monitors can be imported using their numeric ID, e.g.

```console
$ terraform import datadog_monitor.bytes_received_localhost 2081
```

## Composite Monitors

You can compose monitors of all types in order to define more specific alert conditions (see the [doc](https://docs.datadoghq.com/monitors/monitor_types/composite/)). You just need to reuse the ID of your `datadog_monitor` resources. You can also compose any monitor with a `datadog_synthetics_test` by passing the computed `monitor_id` attribute in the query.

```hcl
resource "datadog_monitor" "bar" {
  name = "Composite Monitor"
  type = "composite"
  message = "This is a message"

  query = "${datadog_monitor.foo.id} || ${datadog_synthetics_test.foo.monitor_id}"
}
```
