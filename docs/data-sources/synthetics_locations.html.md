---
layout: "datadog"
page_title: "Datadog: datadog_synthetics_locations"
sidebar_current: "docs-datadog-datasource-synthetics-locations"
description: |-
  Get information on Datadog's Synthetics locations for Synthetics tests
---

# datadog_synthetics_locations

Use this data source to retrieve information about Datadog's Synthetics Locations.
## Example Usage

```hcl
data "datadog_synthetics_locations" "test" {}
```

## Attributes Reference

 * `locations` - An Array of available Synthetics locations for Synthetics tests.
