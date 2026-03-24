
resource "datadog_user" "user_email_channel_test" {
  email = "USER_EMAIL"
}

resource "datadog_on_call_user_notification_channel" "user_email_channel_test" {
  user_id = datadog_user.user_email_channel_test.id

  email {
    address = datadog_user.user_email_channel_test.email
    formats = ["html"]
  }
}

