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

- **name** (String) Synthetics private location name.

### Optional

- **description** (String) Description of the private location.
- **id** (String) The ID of this resource.
- **tags** (List of String) A list of tags to associate with your synthetics private location.

### Read-only

- **config** (String, Sensitive) Configuration skeleton for the private location. See installation instructions of the private location on how to use this configuration.

## Import

Import is supported using the following syntax:

```shell
# Synthetics private locations can be imported using their string ID, e.g.
terraform import datadog_synthetics_private_location.bar pl:private-location-name-abcdef123456
```
