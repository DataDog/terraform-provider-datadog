# The existing API test another team owns.
data "datadog_synthetics_test" "checkout_api" {
  test_id = "abc-123-xyz"
}

# A browser test over the same journey, kept in lockstep with the API test's
# cadence, coverage, and alerting behavior.
resource "datadog_synthetics_test" "checkout_browser" {
  name      = "Checkout journey (browser)"
  type      = "browser"
  status    = data.datadog_synthetics_test.checkout_api.status
  locations = data.datadog_synthetics_test.checkout_api.locations

  request_definition {
    method = "GET"
    url    = "https://www.example.com/checkout"
  }

  device_ids = ["laptop_large"]

  options_list {
    tick_every          = data.datadog_synthetics_test.checkout_api.options_list[0].tick_every
    min_location_failed = data.datadog_synthetics_test.checkout_api.options_list[0].min_location_failed
    monitor_priority    = data.datadog_synthetics_test.checkout_api.options_list[0].monitor_priority

    retry {
      count    = data.datadog_synthetics_test.checkout_api.options_list[0].retry[0].count
      interval = data.datadog_synthetics_test.checkout_api.options_list[0].retry[0].interval
    }
  }
}
