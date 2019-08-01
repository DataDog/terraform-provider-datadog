---
layout: "datadog"
page_title: "Datadog: datadog_downtime"
sidebar_current: "docs-datadog-resource-downtime"
description: |-
  Provides a Datadog downtime resource. This can be used to create and manage downtimes.
---

# datadog_downtime

Provides a Datadog downtime resource. This can be used to create and manage Datadog downtimes.

## Example Usage

```hcl
# Create a new daily 1700-0900 Datadog downtime
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

## Argument Reference

The following arguments are supported:

* `scope` - (Required) A list of items to apply the downtime to, e.g. host:X
* `active` - (Optional) A flag indicating if the downtime is active now.
* `disabled` - (Optional) A flag indicating if the downtime was disabled.
* `start` - (Optional) POSIX timestamp to start the downtime.
* `start_date` - (Optional) String representing date and time to start the downtime in RFC3339 format.
* `end` - (Optional) POSIX timestamp to end the downtime.
* `end_date` - (Optional) String representing date and time to end the downtime in RFC3339 format.
* `timezone` (Optional) The timezone for the downtime, default UTC. It must be a valid IANA Time Zone.
* `recurrence` - (Optional) A dictionary to configure the downtime to be recurring.
    * `type` - days, weeks, months, or years
    * `period` - How often to repeat as an integer. For example to repeat every 3 days, select a type of days and a period of 3.
    * `week_days` - (Optional) A list of week days to repeat on. Choose from: Mon, Tue, Wed, Thu, Fri, Sat or Sun. Only applicable when type is weeks. First letter must be capitalized.
    * `until_occurrences` - (Optional) How many times the downtime will be rescheduled. `until_occurrences` and `until_date` are mutually exclusive.
    * `until_date` - (Optional) The date at which the recurrence should end as a POSIX timestamp. `until_occurrences` and `until_date` are mutually exclusive.
* `message` - (Optional) A message to include with notifications for this downtime.
* `monitor_tags` - (Optional) A list of monitor tags to match. The resulting downtime applies to monitors that match **all** provided monitor tags. This option conflicts with `monitor_id` as it will match all monitors that match these tags.
* `monitor_id` - (Optional) Reference to which monitor this downtime is applied. When scheduling downtime for a given monitor, datadog changes `silenced` property of the monitor to match the `end` POSIX timestamp. **Note:** this will effectively change the `silenced` attribute of the referenced monitor. If that monitor is also tracked by Terraform and you don't want it to be unmuted on the next `terraform apply`, see [details](/docs/providers/datadog/r/monitor.html#silencing-by-hand-and-by-downtimes) in the monitor resource documentation. This option also conflicts with `monitor_tags` use none or one or the other.

## Attributes Reference

The following attributes are exported:

* `id` - ID of the Datadog downtime. On updates this can sometime change based on API logic. For recurring downtimes it would be recommended to `ignore_changes` on this field.
* `active` - If true this indicates the downtime is currently active.
* `disabled` - If true this indicates the downtime is currently disabled.


## Import

Downtimes can be imported using their numeric ID, e.g.

```
$ terraform import datadog_downtime.bytes_received_localhost 2081
```
