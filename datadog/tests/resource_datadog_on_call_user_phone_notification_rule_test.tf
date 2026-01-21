
resource "datadog_user" "on_call_user_phone_rule_test" {
  email = "USER_EMAIL"
}

resource "datadog_on_call_user_notification_channel" "on_call_user_phone_rule_test" {
  user_id = datadog_user.on_call_user_phone_rule_test.id

  phone {
    number = "USER_PHONE"
  }
}

resource "datadog_on_call_user_notification_rule" "on_call_user_phone_rule_test" {
  user_id = datadog_user.on_call_user_phone_rule_test.id
  channel_id = datadog_on_call_user_notification_channel.on_call_user_phone_rule_test.id

  category = "high_urgency"
  delay_minutes = 1
  phone {
    method = "voice"
  }
}

