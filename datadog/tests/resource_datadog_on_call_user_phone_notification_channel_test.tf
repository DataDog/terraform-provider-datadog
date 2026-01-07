
resource "datadog_user" "user_phone_channel_test" {
  email = "USER_EMAIL"
}

resource "datadog_on_call_user_notification_channel" "user_phone_channel_test" {
  user_id = datadog_user.user_phone_channel_test.id

  phone {
    number = "USER_PHONE"
  }
}

