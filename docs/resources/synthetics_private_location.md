---
page_title: "datadog_synthetics_private_location Resource - terraform-provider-datadog"
subcategory: ""
description: |-
  Provides a Datadog synthetics private location resource. This can be used to create and manage Datadog synthetics private locations.
---

# Resource `datadog_synthetics_private_location`

Provides a Datadog synthetics private location resource. This can be used to create and manage Datadog synthetics private locations.

## Example Usage

```terraform
resource "datadog_synthetics_private_location" "private_location" {
    name = "First private location"
    description = "Description of the private location"
    tags = ["foo:bar", "env:test"]
}
```

## Schema

### Required

- **name** (String, Required)

### Optional

- **description** (String, Optional)
- **id** (String, Optional) The ID of this resource.
- **tags** (List of String, Optional)

### Read-only

- **config** (String, Read-only)

## Import

Import is supported using the following syntax:

```shell
# Synthetics private locations can be imported using their string ID, e.g.
terraform import datadog_synthetics_private_location.bar pl:private-location-name-abcdef123456
```
