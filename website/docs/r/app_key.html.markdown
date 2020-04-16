---
layout: "datadog"
page_title: "Datadog: app_key"
sidebar_current: "docs-datadog-resource-app-key"
description: |-
  Provides a Datadog app key resource. This can be used to create and manage application keys.
---

# datadog_app_key

Provides a Datadog app key resource. This can be used to create and manage Datadog application keys.

## Example Usage

```hcl
# Create a new Datadog app key
resource "datadog_app_key" "foo" {
  name = "My App"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name for the application key

## Attributes Reference

The following attributes are exported:

* `hash` - The application key

## Import

Application keys can be imported using their hash, e.g.

```
$ terraform import datadog_app_key.foo 10d7ea9c5803590730411eea434758ffcf24d280
```
