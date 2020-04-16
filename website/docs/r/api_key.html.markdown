---
layout: "datadog"
page_title: "Datadog: api_key"
sidebar_current: "docs-datadog-resource-api-key"
description: |-
  Provides a Datadog API key resource. This can be used to create and manage API keys.
---

# datadog_api_key

Provides a Datadog API key resource. This can be used to create and manage Datadog API keys.

## Example Usage

```hcl
# Create a new Datadog API key
resource "datadog_api_key" "foo" {
  name = "My App"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name for the API key

## Attributes Reference

The following attributes are exported:

* `key` - The API key

## Import

Keys can be imported using their key, e.g.

```
$ terraform import datadog_api_key.foo 289d32a8097785306e8c990cd66f416a
```
