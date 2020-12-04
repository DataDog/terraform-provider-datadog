---
page_title: "datadog_synthetics_private_location"
---

# datadog_synthetics_private_location Resource

Provides a Datadog synthetics private location resource. This can be used to create and manage Datadog synthetics private locations.

## Example Usage

```hcl
resource "datadog_synthetics_private_location" "private_location" {
    name = "First private location"
    description = "Description of the private location"
    tags = ["foo:bar", "env:test"]
}
```

## Argument Reference

The following arguments are supported:

-   `name`: (Required) Synthetics private location name.
-   `description`: (Optional) Description of the private location.
-   `tags`: (Required) A list of tags to associate with your synthetics private location.

## Argument Reference

-   `id`: ID of the Datadog synthetics private location.
-   `config`: Configuration skeleton for the private location. See installation instructions of the private location on how to use this configuration.

## Import

Synthetics private locations can be imported using their string ID, e.g.

```
$ terraform import datadog_synthetics_private_location.bar pl:private-location-name-abcdef123456
```
