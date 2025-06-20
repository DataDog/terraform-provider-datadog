# terraform_examples/main.tf
terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

provider "datadog" {
  api_key = "8ac3d0cc1e3bdba2ee1d5472af7c548"
  app_key = "156541800324d6546b066cf78384ff782057abb4"
}
