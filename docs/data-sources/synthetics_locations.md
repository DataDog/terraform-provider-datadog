---
page_title: "Datadog: datadog_synthetics_locations"
---

# datadog_synthetics_locations Data Source

Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).

## Example Usage

```hcl
data "datadog_synthetics_locations" "test" {}

resource "datadog_synthetics_test" "test_api" {
  type = "api"
  locations = keys(data.datadog_synthetics_locations.test.locations)
}
```

## Attributes Reference

- `locations` : A map of available Synthetics location IDs to names for Synthetics tests.
