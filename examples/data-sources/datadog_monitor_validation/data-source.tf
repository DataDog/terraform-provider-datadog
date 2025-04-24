locals {
  foo_monitor_query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 4"
  foo_monitor_type  = "metric alert"
}

data "datadog_monitor_validation" "foo" {
  query = local.foo_monitor_query
  type  = local.foo_monitor_type

  lifecycle {
    postcondition {
      condition     = self.valid == true
      error_message = join(", ", self.validation_errors)
    }
  }
}

resource "datadog_monitor" "foo" {
  name               = "foo monitor"
  type               = local.foo_monitor_query
  message            = "Monitor triggered. Notify: @hipchat-channel"
  escalation_message = "Escalation message @pagerduty"
  query              = local.foo_monitor_query

  monitor_thresholds {
    warning  = 2
    critical = 4
  }

  include_tags = true
  tags         = ["foo:bar", "team:fooBar"]
}
