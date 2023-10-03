# Create new powerpack resource

resource "datadog_powerpack" "foo" {
  description = "Powerpack for ABC"
  group_widget {
  }
  name = "Sample Powerpack"
  tags = ["tag:foo1"]
  template_variables {
    defaults = "UPDATE ME"
    name     = "datacenter"
  }
}