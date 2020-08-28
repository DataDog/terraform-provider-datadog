---
page_title: "datadog_monitor"
---

# datadog_monitor Data Source

Use this data source to retrieve information about an existing monitor for use in other resources.

## Example Usage

```
data "datadog_monitor" "test" {
  name_filter = "My awesome monitor"
  monitor_tags_filter = ["foo:bar"]
}
```

## Argument Reference

- `name_filter`: (Optional) A monitor name to limit the search.
- `tags_filter`: (Optional) A list of tags to limit the search. This filters on the monitor scope.
- `monitor_tags_filter`: (Optional) A list of monitor tags to limit the search. This filters on the tags set on the monitor itself.

~> **NOTE:** If more or less than a single match is returned by the search, Terraform will fail. Ensure that your search is specific enough to return a single monitor.

## Attributes Reference

- `id` : ID of the Datadog monitor
- `name` : Name of the monitor.
- `type` : Type of the monitor.
- `query` : Query of the monitor.
- `message` : Message included with notifications for this monitor.
- `escalation_message` : Message included with a re-notification for this monitor.
- `thresholds` : Alert thresholds of the monitor.
- `notify_no_data` : Whether or not this monitor notifies when data stops reporting.
- `new_host_delay` : Time (in seconds) allowing a host to boot and applications to fully start before starting the evaluation of monitor results.
- `evaluation_delay` : Time (in seconds) for which evaluation is delayed. This is only used by metric monitors.
- `no_data_timeframe` : The number of minutes before the monitor notifies when data stops reporting.
- `renotify_interval` : The number of minutes after the last notification before the monitor re-notifies on the current status.
- `notify_audit` : Whether or not tagged users are notified on changes to the monitor.
- `timeout_h` : Number of hours of the monitor not reporting data before it automatically resolves from a triggered state.
- `include_tags` : Whether or not notifications from the monitor automatically inserts its triggering tags into the title.
- `enable_logs_sample` : Whether or not a list of log values which triggered the alert is included. This is only used by log monitors.
- `require_full_window` : Whether or not the monitor needs a full window of data before it is evaluated.
- `locked` : Whether or not changes to the monitor are restricted to the creator or admins.
- `tags` : List of tags associated with the monitor.
- `threshold_windows` : Mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. This is only used by anomaly monitors.
