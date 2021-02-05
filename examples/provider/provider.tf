# Terraform 0.13+ uses the Terraform Registry:

terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
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




# Terraform 0.12- can be specified as:

# Configure the Datadog provider
provider "datadog" {
  api_key = "<DD_API_KEY>"
  app_key = "<DD_APP_KEY>"
}

# Create a new monitor
resource "datadog_monitor" "default" {
  # ...
}

# Create a new dashboard
resource "datadog_dashboard" "default" {
  # ...
}
