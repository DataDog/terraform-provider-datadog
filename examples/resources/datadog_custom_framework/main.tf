terraform {
  required_providers {
    datadog = {
      source = "DataDog/datadog"
    }
  }
}

provider "datadog" {
  api_key = var.datadog_api_key
  app_key = var.datadog_app_key
}


resource "datadog_custom_framework" "example" {
  version     = "1.0.0"
  handle      = "terraform-created-framework-handle"
  name        = "terraform-created-framework"
  icon_url    = "https://example.com/icon.png"
  description = "This is a test I created this resource through terraform"
  requirements {
    name = "requirement2"
    controls {
      name     = "control2"
      rules_id = ["def-000-cea"]
    }
  }
  requirements {
    name = "requirement1"
    controls {
      name     = "changedControlName"
      rules_id = ["def-000-cea", "def-000-be9"]
    }
  }
}



