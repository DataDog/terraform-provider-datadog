resource "datadog_compliance_custom_framework" "example" {
  version  = "1"
  handle   = "new-terraform-framework-handle"
  name     = "new-terraform-framework"
  icon_url = "https://example.com/icon.png"
  requirements {
    name = "requirement1"
    controls {
      name     = "control1"
      rules_id = ["def-000-k6h", "def-000-u48"]
    }
    controls {
      name     = "control2"
      rules_id = ["def-000-k8u"]
    }
  }
  requirements {
    name = "requirement2"
    controls {
      name     = "control3"
      rules_id = ["def-000-k6h"]
    }
  }
}