resource "datadog_logs_custom_destination" "sample_destination" {
  name    = "sample destination"
  query   = "service:my-service"
  enabled = true

  http_destination {
    endpoint = "https://example.org"
    basic_auth {
      username = "my-username"
      password = "my-password"
    }
  }
}
