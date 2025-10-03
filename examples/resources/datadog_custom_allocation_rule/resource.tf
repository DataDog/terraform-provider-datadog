# Create new datadog_custom_allocation_rule resource

resource "datadog_custom_allocation_rule" "foo" {
  costs_to_allocate {
    condition = "UPDATE ME"
    tag       = "UPDATE ME"
    value     = "UPDATE ME"
    values    = "UPDATE ME"
  }
  enabled       = "UPDATE ME"
  providernames = [""]
  rule_name     = "UPDATE ME"
  strategy {
    allocated_by {
      allocated_tags {
        key   = "UPDATE ME"
        value = "UPDATE ME"
      }
      percentage = "UPDATE ME"
    }
    allocated_by_filters {
      condition = "UPDATE ME"
      tag       = "UPDATE ME"
      value     = "UPDATE ME"
      values    = "UPDATE ME"
    }
    allocated_by_tag_keys = "UPDATE ME"
    based_on_costs {
      condition = "UPDATE ME"
      tag       = "UPDATE ME"
      value     = "UPDATE ME"
      values    = "UPDATE ME"
    }
    based_on_timeseries {
    }
    evaluate_grouped_by_filters {
      condition = "UPDATE ME"
      tag       = "UPDATE ME"
      value     = "UPDATE ME"
      values    = "UPDATE ME"
    }
    evaluate_grouped_by_tag_keys = "UPDATE ME"
    granularity                  = "UPDATE ME"
    method                       = "UPDATE ME"
  }
  type = "UPDATE ME"
}
