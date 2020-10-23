---
layout: "datadog"
page_title: "Provider: Datadog"
sidebar_current: "docs-datadog-index"
description: |-
  The Datadog provider is used to interact with the resources supported by Datadog. The provider needs to be configured with the proper credentials before it can be used.
---

# Datadog Provider

The [Datadog](https://www.datadoghq.com) provider is used to interact with the
resources supported by Datadog. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the Datadog provider
provider "datadog" {
  api_key = "${var.datadog_api_key}"
  app_key = "${var.datadog_app_key}"
}

# Create a new monitor
resource "datadog_monitor" "default" {
  # ...
}

# Create a new timeboard
resource "datadog_timeboard" "default" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `api_key` - (Required unless `validate` is false) Datadog API key. This can also be set via the `DD_API_KEY` environment variable.
* `app_key` - (Required unless `validate` is false) Datadog APP key. This can also be set via the `DD_APP_KEY` environment variable.
* `api_url` - (Optional) The API Url. This can be also be set via the `DD_HOST` environment variable. Note that this URL must not end with the `/api/` path. For example, `https://api.datadoghq.com/` is a correct value, while `https://api.datadoghq.com/api/` is not. And if you're working with  "EU" version of Datadog, use `https://api.datadoghq.eu/`.
* `validate` - (Optional) Enables validation of the provided API and APP keys during provider initialization. Default is true. When false, `api_key` and `app_key` may only be set to empty strings.
