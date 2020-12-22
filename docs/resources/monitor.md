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

## Schema

### Required

- **message** (String, Required)
- **name** (String, Required)
- **query** (String, Required)
- **type** (String, Required)

### Optional

- **enable_logs_sample** (Boolean, Optional)
- **escalation_message** (String, Optional)
- **evaluation_delay** (Number, Optional)
- **force_delete** (Boolean, Optional)
- **id** (String, Optional) The ID of this resource.
- **include_tags** (Boolean, Optional)
- **locked** (Boolean, Optional)
- **new_host_delay** (Number, Optional)
- **no_data_timeframe** (Number, Optional)
- **notify_audit** (Boolean, Optional)
- **notify_no_data** (Boolean, Optional)
- **priority** (Number, Optional)
- **renotify_interval** (Number, Optional)
- **require_full_window** (Boolean, Optional)
- **silenced** (Map of Number, Optional, Deprecated)
- **tags** (Set of String, Optional)
- **threshold_windows** (Map of String, Optional)
- **thresholds** (Map of String, Optional)
- **timeout_h** (Number, Optional)
- **validate** (Boolean, Optional)

## Import

Import is supported using the following syntax:

```shell
terraform import datadog_monitor.bytes_received_localhost 2081
```
