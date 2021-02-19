---
page_title: "datadog_downtime Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog downtime resource. This can be used to create and manage Datadog downtimes.
---

# Resource `datadog_downtime`

Provides a Datadog downtime resource. This can be used to create and manage Datadog downtimes.

## Example Usage

```terraform
# Example: downtime for a specific monitor
# Create a new daily 1700-0900 Datadog downtime for a specific monitor id
resource "datadog_downtime" "foo" {
  scope      = ["*"]
  start      = 1483308000
  end        = 1483365600
  monitor_id = 12345

  recurrence {
    type   = "days"
    period = 1
  }
}

# Example: downtime for all monitors
# Create a new daily 1700-0900 Datadog downtime for all monitors
resource "datadog_downtime" "foo" {
  scope = ["*"]
  start = 1483308000
  end   = 1483365600

  recurrence {
    type   = "days"
    period = 1
  }
}
```

## Schema

### Required

- **scope** (List of String, Required) specify the group scope to which this downtime applies. For everything use '*'

### Optional

- **active** (Boolean, Optional) When true indicates this downtime is being actively applied
- **disabled** (Boolean, Optional) When true indicates this downtime is not being applied
- **end** (Number, Optional) Optionally specify an end date when this downtime should expire
- **end_date** (String, Optional) String representing date and time to end the downtime in RFC3339 format.
- **id** (String, Optional) The ID of this resource.
- **message** (String, Optional) An optional message to provide when creating the downtime, can include notification handles
- **monitor_id** (Number, Optional) When specified, this downtime will only apply to this monitor
- **monitor_tags** (Set of String, Optional) A list of monitor tags (up to 25), i.e. tags that are applied directly to monitors to which the downtime applies
- **recurrence** (Block List, Max: 1) Optional recurring schedule for this downtime (see [below for nested schema](#nestedblock--recurrence))
- **start** (Number, Optional) Specify when this downtime should start
- **start_date** (String, Optional) String representing date and time to start the downtime in RFC3339 format.
- **timezone** (String, Optional) The timezone for the downtime, default UTC

<a id="nestedblock--recurrence"></a>
### Nested Schema for `recurrence`

Required:

- **type** (String, Required) One of `days`, `weeks`, `months`, or `years`

Optional:

- **period** (Number, Optional) How often to repeat as an integer. For example to repeat every 3 days, select a `type` of `days` and a `period` of `3`.
- **rrule** (String, Optional) The RRULE standard for defining recurring events. For example, to have a recurring event on the first day of each month, use `FREQ=MONTHLY;INTERVAL=1`. Most common rrule options from the iCalendar Spec are supported. Attributes specifying the duration in RRULE are not supported (for example, `DTSTART`, `DTEND`, `DURATION`).
- **until_date** (Number, Optional) The date at which the recurrence should end as a POSIX timestamp. `until_occurrences` and `until_date` are mutually exclusive.
- **until_occurrences** (Number, Optional) How many times the downtime will be rescheduled. `until_occurrences` and `until_date` are mutually exclusive.
- **week_days** (List of String, Optional) A list of week days to repeat on. Choose from: `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat` or `Sun`. Only applicable when `type` is `weeks`. First letter must be capitalized.

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_downtime.bytes_received_localhost 2081
```
