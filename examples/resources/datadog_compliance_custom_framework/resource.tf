resource "datadog_compliance_custom_framework" "framework" {
  name    = "my-custom-framework-terraform-2"
  version = "2.0.0"
  handle  = "my-custom-framework-terraform-2"

  requirements {
    name = "requirement2"
    controls {
      name     = "control2"
      rules_id = ["def-000-h9o", "def-000-b6i", "def-000-yed", "def-000-h5a", "def-000-aw5"]
    }
    controls {
      name     = "control1"
      rules_id = ["def-000-j9v", "def-000-465", "def-000-vq1", "def-000-4hf", "def-000-s2d", "def-000-vnl"]
    }
  }

  requirements {
    name = "requirement1"
    controls {
      name     = "control2"
      rules_id = ["def-000-wuf", "def-000-7og"]
    }
    controls {
      name     = "control5"
      rules_id = ["def-000-mdt", "def-000-zrx", "def-000-z6k"]
    }
  }
}