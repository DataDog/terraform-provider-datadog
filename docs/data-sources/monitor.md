---
page_title: "datadog_monitor Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve information about an existing monitor for use in other resources.
---

# Data Source `datadog_monitor`

Use this data source to retrieve information about an existing monitor for use in other resources.

## Example Usage

```terraform
data "datadog_monitor" "test" {
  name_filter = "My awesome monitor"
  monitor_tags_filter = ["foo:bar"]
}
```

## Schema

### Optional

- **id** (String, Optional) The ID of this resource.
- **monitor_tags_filter** (List of String, Optional) A list of monitor tags to limit the search. This filters on the tags set on the monitor itself.
- **name_filter** (String, Optional) A monitor name to limit the search.
- **tags_filter** (List of String, Optional) A list of tags to limit the search. This filters on the monitor scope.

### Read-only

- **enable_logs_sample** (Boolean, Read-only) Whether or not a list of log values which triggered the alert is included. This is only used by log monitors.
- **escalation_message** (String, Read-only) Message included with a re-notification for this monitor.
- **evaluation_delay** (Number, Read-only) Time (in seconds) for which evaluation is delayed. This is only used by metric monitors.
- **include_tags** (Boolean, Read-only) Whether or not notifications from the monitor automatically inserts its triggering tags into the title.
- **locked** (Boolean, Read-only) Whether or not changes to the monitor are restricted to the creator or admins.
- **message** (String, Read-only) Message included with notifications for this monitor
- **monitor_threshold_windows** (List of Object, Read-only) Mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. This is only used by anomaly monitors. (see [below for nested schema](#nestedatt--monitor_threshold_windows))
- **monitor_thresholds** (List of Object, Read-only) Alert thresholds of the monitor. (see [below for nested schema](#nestedatt--monitor_thresholds))
- **name** (String, Read-only) Name of the monitor
- **new_host_delay** (Number, Read-only) Time (in seconds) allowing a host to boot and applications to fully start before starting the evaluation of monitor results.
- **no_data_timeframe** (Number, Read-only) The number of minutes before the monitor notifies when data stops reporting.
- **notify_audit** (Boolean, Read-only) Whether or not tagged users are notified on changes to the monitor.
- **notify_no_data** (Boolean, Read-only) Whether or not this monitor notifies when data stops reporting.
- **query** (String, Read-only) Query of the monitor.
- **renotify_interval** (Number, Read-only) The number of minutes after the last notification before the monitor re-notifies on the current status.
- **require_full_window** (Boolean, Read-only) Whether or not the monitor needs a full window of data before it is evaluated.
- **tags** (Set of String, Read-only) List of tags associated with the monitor.
- **threshold_windows** (Map of String, Read-only, Deprecated) Mapping containing `recovery_window` and `trigger_window` values, e.g. `last_15m`. This is only used by anomaly monitors.
- **thresholds** (Map of String, Read-only, Deprecated) Alert thresholds of the monitor.
- **timeout_h** (Number, Read-only) Number of hours of the monitor not reporting data before it automatically resolves from a triggered state.
- **type** (String, Read-only) Type of the monitor.

<a id="nestedatt--monitor_threshold_windows"></a>
### Nested Schema for `monitor_threshold_windows`

- **recovery_window** (String)
- **trigger_window** (String)


<a id="nestedatt--monitor_thresholds"></a>
### Nested Schema for `monitor_thresholds`

- **critical** (String)
- **critical_recovery** (String)
- **ok** (String)
- **unknown** (String)
- **warning** (String)
- **warning_recovery** (String)


