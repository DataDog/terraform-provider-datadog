resource "datadog_custom_framework" "example4" {
  version     = "1"
  handle      = "new-terraform-framework-handle"
  name        = "new-terraform-framework"
  icon_url    = "https://example.com/icon.png"
  description = "This is a test I created this resource through terraform"
  requirements {
    name = "requirement1"
    controls {
      name     = "new-control"
      rules_id = ["def-000-be9", "def-000-cea"]
    }
    controls {
      name     = "new-control"
      rules_id = ["def-000-be9"]
    }
  }
}