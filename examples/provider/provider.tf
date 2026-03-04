terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
  required_version = ">= 1.1.5"
}

# Configure the Datadog provider
provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}
