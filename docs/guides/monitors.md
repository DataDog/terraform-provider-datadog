---
subcategory: ""
page_title: "Monitor Resource Examples"
description: |-
    Monitor Resource Examples
---

### Monitor Resource Examples

This page lists examples of how to create different Datadog monitor types within Terraform. This list is not exhaustive and will be updated over time to provide more examples.

## Composite Monitors

You can compose monitors of all types in order to define more specific alert conditions (see the [doc](https://docs.datadoghq.com/monitors/monitor_types/composite/)). You just need to reuse the ID of your `datadog_monitor` resources. You can also compose any monitor with a `datadog_synthetics_test` by passing the computed `monitor_id` attribute in the query.

```terraform
resource "datadog_monitor" "bar" {
  name    = "Composite Monitor"
  type    = "composite"
  message = "This is a message"
  query   = "${datadog_monitor.foo.id} || ${datadog_synthetics_test.foo.monitor_id}"
}
```

## Watchdog Monitors

```terraform
resource "datadog_monitor" "watchdog_monitor" {
  name    = "Watchdog detected an anomaly: {{event.title}}"
  type    = "event-v2 alert"
  message = "Watchdog monitor created from Terraform"

  query = "events(\"source:watchdog story_category:apm env:test_env\").rollup(\"count\").by(\"story_key,service,resource_name\").last(\"30m\") > 0"

  tags = ["terraform:true"]
}
```

## Anomaly Monitors

```terraform
resource "datadog_monitor" "cpu_anomalous" {
  name    = "Anomalous CPU usage"
  type    = "query alert"
  message = "CPU utilization is outside normal bounds"
  query   = "avg(last_4h):anomalies(ewma_20(avg:system.cpu.system{env:prod,service:website}.as_rate()), 'robust', 3, direction='below', alert_window='last_30m', interval=60, count_default_zero='true', seasonality='weekly') >= 1"

  monitor_thresholds {
    critical = 1.0
  }

  monitor_threshold_windows {
    trigger_window  = "last_30m"
    recovery_window = "last_30m"
  }

  renotify_interval = 60
}
```

## Process Monitors

```terraform
resource "datadog_monitor" "process_alert_example" {
  name    = "Process Alert Monitor"
  type    = "process alert"
  message = "Multiple Java processes running on example-tag"
  query   = "processes('java').over('example-tag').rollup('count').last('10m') > 1"

  monitor_thresholds {
    critical = 1.0
  }

  renotify_interval = 60
}
```

## Log Alert Monitors

```terraform
resource "datadog_monitor" "log_alert_example" {
  enable_logs_sample = true
  name = "Log Alert Monitor"
  type = "log alert"
  query = logs("source:nginx status:error").index("*")

  message = "Alert on NGINX error logs"

  monitor_thresholds {
    critical = 100
    warning = 90
  }

  renotify_interval = 60
}
```
