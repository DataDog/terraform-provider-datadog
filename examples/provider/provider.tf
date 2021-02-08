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



# Terraform 0.12- can be specified as:

# Configure the Datadog provider
# provider "datadog" {
#   api_key = "${var.datadog_api_key}"
#   app_key = "${var.datadog_app_key}"
# }
