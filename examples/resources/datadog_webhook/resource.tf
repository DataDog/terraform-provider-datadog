# Create a new Datadog webhook

resource "datadog_webhook" "foo" {
  name           = "test-webhook"
  url            = "example.com"
  encode_as      = "json"
  custom_headers = jsonencode({ "custom" : "header" })
  payload        = jsonencode({ "custom" : "payload" })
}
