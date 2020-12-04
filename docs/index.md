---
page_title: "Provider: Datadog"
---

# Datadog Provider

The [Datadog](https://www.datadoghq.com) provider is used to interact with the resources supported by Datadog. The provider needs to be configured with the proper credentials before it can be used.

Try the [hands-on tutorial](https://learn.hashicorp.com/tutorials/terraform/datadog-provider?in=terraform/use-case?utm_source=WEBSITE&utm_medium=WEB_IO&utm_offer=ARTICLE_PAGE&utm_content=DOCS) on the Datadog provider on the HashiCorp Learn site.

Use the navigation to the left to read about the available resources.

## Example Usage

Terraform 0.13+ uses the Terraform Registry:

```hcl
terraform {
  required_providers {
    datadog = {
      source  = "DataDog/datadog"
    }
  }
}


# Configure the Datadog provider
provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}

# Create a new monitor
resource "datadog_monitor" "default" {
  # ...
}

# Create a new dashboard
resource "datadog_dashboard" "default" {
  # ...
}

```

Terraform 0.12- can be specified as:

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

# Create a new dashboard
resource "datadog_dashboard" "default" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

-   `api_key`: (Required unless `validate` is false) Datadog API key. This can also be set via the `DD_API_KEY` environment variable.
-   `app_key`: (Required unless `validate` is false) Datadog APP key. This can also be set via the `DD_APP_KEY` environment variable.
-   `api_url`: (Optional) The API Url. This can be also be set via the `DD_HOST` environment variable. Note that this URL must not end with the `/api/` path. For example, `https://api.datadoghq.com/` is a correct value, while `https://api.datadoghq.com/api/` is not. And if you're working with "EU" version of Datadog, use `https://api.datadoghq.eu/`.
-   `validate`: (Optional) Enables validation of the provided API and APP keys during provider initialization. Default is true. When false, `api_key` and `app_key`won't be checked.
