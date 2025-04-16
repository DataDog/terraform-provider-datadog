resource "datadog_custom_framework" "example" {
  version     = "1.0.0"
  handle      = "example_handle"
  name        = "example_name"
  icon_url    = "https://example.com/icon.png"
  description = "description"
  requirements {
    name = "requirement1"
    controls {
      name     = "control1"
      rules_id = ["rule1", "rule2"]
    }
  }
}


