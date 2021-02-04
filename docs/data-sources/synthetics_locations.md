---
page_title: "datadog_synthetics_locations Data Source - terraform-provider-datadog"
subcategory: ""
description: |-
  Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).
---

# Data Source `datadog_synthetics_locations`

Use this data source to retrieve Datadog's Synthetics Locations (to be used in Synthetics tests).

## Example Usage

```terraform
data "datadog_synthetics_locations" "test" {}

resource "datadog_synthetics_test" "test_api" {
  type = "api"
  locations = keys(data.datadog_synthetics_locations.test.locations)
}
```

## Schema

### Optional

- **id** (String, Optional) The ID of this resource.

### Read-only

- **locations** (Map of String, Read-only) A map of available Synthetics location IDs to names for Synthetics tests.


