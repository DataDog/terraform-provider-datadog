---
layout: "datadog"
page_title: "Datadog: monitor_examples"
sidebar_current: "docs-datadog-monitor-examples"
description: |-
  Provides examples for the different types of Datadog monitors. This list isn't exhaustive but serves as a reference for some examples.
---

## Watchdog Monitors

```
resource "datadog_monitor" "Watchdog_Monitor" {
  name               = "Watchdog Monitor TF"
  type               = "event alert"
  message            = "Check out this Watchdog Service Monitor"
  escalation_message = "Escalation message"

  query = "events('priority:all sources:watchdog tags:story_type:service,env:test_env,service:test_service:_aggregate').by('service,resource_name').rollup('count').last('30m') > 0"

  notify_no_data    = false
  renotify_interval = 60

  notify_audit = false
  timeout_h    = 60
  include_tags = true

  tags = ["foo:bar", "baz"]
}
```
