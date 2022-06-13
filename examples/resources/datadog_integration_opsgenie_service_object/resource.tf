resource "datadog_integration_opsgenie_service_object" "fake_service_name" {
  name = "fake_service_name"
  opsgenie_api_key  = "00000000-0000-0000-0000-000000000000"
  region = "us"
}

resource "datadog_integration_opsgenie_service_object" "fake_service_name_2" {
  name = "fake_service_name_2"
  opsgenie_api_key  = "11111111-1111-1111-1111-111111111111"
  region = "eu"
}
