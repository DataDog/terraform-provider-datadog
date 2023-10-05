# Create new powerpack resource

resource "datadog_powerpack" "foo" {
  description = "Powerpack for ABC"
  group_widget {
    definition {
      layout_type = "ordered"
      show_title  = True
      title       = "Sample Powerpack"
      type        = "group"
      widgets {
        definition {
        }
        layout {
          height = "UPDATE ME"
          width  = "UPDATE ME"
          x      = "UPDATE ME"
          y      = "UPDATE ME"
        }
      }
    }
    layout {
      height = "UPDATE ME"
      width  = "UPDATE ME"
      x      = "UPDATE ME"
      y      = "UPDATE ME"
    }
  }
  name = "Sample Powerpack"
  tags = ["tag:foo1"]
  template_variables {
    defaults = "UPDATE ME"
    name     = "datacenter"
  }
}