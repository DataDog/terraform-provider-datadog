data "datadog_synthetics_locations" "test" {}

resource "datadog_synthetics_test" "test_api" {
  type      = "api"
  locations = keys(data.datadog_synthetics_locations.test.locations)
}
