---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_downtime_schedule Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog DowntimeSchedule resource. This can be used to create and manage Datadog downtimes.
---

# datadog_downtime_schedule (Resource)

Provides a Datadog DowntimeSchedule resource. This can be used to create and manage Datadog downtimes.

## Example Usage

```terraform
# Create new downtime_schedule resource


resource "datadog_downtime_schedule" "downtime_schedule_example" {
  scope = "env:us9-prod7 AND team:test123"
  monitor_identifier {
    monitor_tags = ["test:123", "data:test"]
  }
  recurring_schedule {
    recurrence {
      duration = "1h"
      rrule    = "FREQ=DAILY;INTERVAL=1"
      start    = "2050-01-02T03:04:05"
    }
    timezone = "America/New_York"
  }
  display_timezone                 = "America/New_York"
  message                          = "Message about the downtime"
  mute_first_recovery_notification = true
  notify_end_states                = ["alert", "warn"]
  notify_end_types                 = ["canceled", "expired"]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `scope` (String) The scope to which the downtime applies. Must follow the [common search syntax](https://docs.datadoghq.com/logs/explorer/search_syntax/).

### Optional

- `display_timezone` (String) The timezone in which to display the downtime's start and end times in Datadog applications. This is not used as an offset for scheduling.
- `message` (String) A message to include with notifications for this downtime. Email notifications can be sent to specific users by using the same `@username` notation as events.
- `monitor_identifier` (Block, Optional) (see [below for nested schema](#nestedblock--monitor_identifier))
- `mute_first_recovery_notification` (Boolean) If the first recovery notification during a downtime should be muted.
- `notify_end_states` (Set of String) States that will trigger a monitor notification when the `notify_end_types` action occurs.
- `notify_end_types` (Set of String) Actions that will trigger a monitor notification if the downtime is in the `notify_end_types` state.
- `one_time_schedule` (Block, Optional) (see [below for nested schema](#nestedblock--one_time_schedule))
- `recurring_schedule` (Block, Optional) (see [below for nested schema](#nestedblock--recurring_schedule))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--monitor_identifier"></a>
### Nested Schema for `monitor_identifier`

Optional:

- `monitor_id` (Number) ID of the monitor to prevent notifications.
- `monitor_tags` (Set of String) A list of monitor tags. For example, tags that are applied directly to monitors, not tags that are used in monitor queries (which are filtered by the scope parameter), to which the downtime applies. The resulting downtime applies to monitors that match **all** provided monitor tags. Setting `monitor_tags` to `[*]` configures the downtime to mute all monitors for the given scope.


<a id="nestedblock--one_time_schedule"></a>
### Nested Schema for `one_time_schedule`

Optional:

- `end` (String) ISO-8601 Datetime to end the downtime. Must include a UTC offset of zero. If not provided, the downtime never ends.
- `start` (String) ISO-8601 Datetime to start the downtime. Must include a UTC offset of zero. If not provided, the downtime starts the moment it is created.


<a id="nestedblock--recurring_schedule"></a>
### Nested Schema for `recurring_schedule`

Optional:

- `recurrence` (Block List) (see [below for nested schema](#nestedblock--recurring_schedule--recurrence))
- `timezone` (String) The timezone in which to schedule the downtime.

<a id="nestedblock--recurring_schedule--recurrence"></a>
### Nested Schema for `recurring_schedule.recurrence`

Required:

- `duration` (String) The length of the downtime. Must begin with an integer and end with one of 'm', 'h', d', or 'w'.
- `rrule` (String) The `RRULE` standard for defining recurring events. For example, to have a recurring event on the first day of each month, set the type to `rrule` and set the `FREQ` to `MONTHLY` and `BYMONTHDAY` to `1`. Most common `rrule` options from the [iCalendar Spec](https://tools.ietf.org/html/rfc5545) are supported.  **Note**: Attributes specifying the duration in `RRULE` are not supported (for example, `DTSTART`, `DTEND`, `DURATION`). More examples available in this [downtime guide](https://docs.datadoghq.com/monitors/guide/suppress-alert-with-downtimes/?tab=api).

Optional:

- `start` (String) ISO-8601 Datetime to start the downtime. Must not include a UTC offset. If not provided, the downtime starts the moment it is created.

## Import

Import is supported using the following syntax:

The [`terraform import` command](https://developer.hashicorp.com/terraform/cli/commands/import) can be used, for example:

```shell
terraform import datadog_downtime_schedule.new_list "00e000000-0000-1234-0000-000000000000"
```
