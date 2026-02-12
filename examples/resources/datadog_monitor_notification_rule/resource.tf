resource "datadog_monitor_notification_rule" "team_checkout_notification_rule" {
  name       = "Route alerts from checkout team"
  recipients = ["slack-checkout-ops", "jira-checkout"]
  filter {
    tags = ["team:payment"]
  }
}

resource "datadog_monitor_notification_rule" "team_payment_notification_rule" {
  name    = "Routing logic for team payment"
  filter {
    scope = "team:payment AND NOT env:dev AND service:(payment-processing OR payment-gateway)"
  }
  conditional_recipients {
    conditions {
      scope = "priority:p1"
      recipients = ["oncall-payment", "slack-payment"]
    }
    conditions {
      scope = "priority:p5"
      recipients = ["slack-payment"]
    }
    fallback_recipients = ["slack-payment"]
  }
}
