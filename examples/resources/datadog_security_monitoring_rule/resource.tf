resource "datadog_security_monitoring_rule" "myrule" {
  name = "My rule"

  message = "The rule has triggered."
  enabled = true

  query {
    name            = "errors"
    query           = "status:error"
    aggregation     = "count"
    group_by_fields = ["host"]
  }

  query {
    name            = "warnings"
    query           = "status:warning"
    aggregation     = "count"
    group_by_fields = ["host"]
  }

  case {
    status        = "high"
    condition     = "errors > 3 && warnings > 10"
    notifications = ["@user"]
  }

  options {
    evaluation_window   = 300
    keep_alive          = 600
    max_signal_duration = 900
  }

  tags = ["type:dos"]
}
