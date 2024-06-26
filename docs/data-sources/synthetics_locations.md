---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "datadog_synthetics_locations Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).
---

# datadog_synthetics_locations (Data Source)

Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).

## Example Usage

```terraform
data "datadog_synthetics_locations" "test" {}

resource "datadog_synthetics_test" "test_api" {
  type      = "api"
  locations = keys(data.datadog_synthetics_locations.test.locations)
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Read-Only

- `id` (String) The ID of this resource.
- `locations` (Map of String) A map of available Synthetics location IDs to names for Synthetics tests.
