resource "datadog_compliance_custom_framework" "example" {
  version  = "1"
  handle   = "new-terraform-framework-handle"
  name     = "new-terraform-framework"
  icon_url = "https://example.com/icon.png"
  requirements {
    name = "requirement1"
    controls {
      name     = "control1"
      rules_id = ["04w-clb-3io", "0eg-j5g-xip"]
    }
    controls {
      name     = "control2"
      rules_id = ["def-000-11s"]
    }
  }
  requirements {
    name = "requirement2"
    controls {
      name     = "control3"
      rules_id = ["def-000-1im"]
    }
  }
}