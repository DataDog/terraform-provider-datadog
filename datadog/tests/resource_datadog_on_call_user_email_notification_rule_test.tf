
resource "datadog_user" "on_call_user_email_rule_test" {
  email = "USER_EMAIL"
}

resource "datadog_on_call_user_notification_channel" "on_call_user_email_rule_test" {
  user_id = datadog_user.on_call_user_email_rule_test.id

  email {
    address = datadog_user.on_call_user_email_rule_test.email
    formats = ["html"]
  }
}

resource "datadog_on_call_user_notification_rule" "on_call_user_email_rule_test" {
  user_id = datadog_user.on_call_user_email_rule_test.id
  channel_id = datadog_on_call_user_notification_channel.on_call_user_email_rule_test.id

  category = "high_urgency"
  delay_minutes = DELAY_MINUTES
}

