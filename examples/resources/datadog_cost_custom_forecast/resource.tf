# A budget with two monthly entries for the same tag combination, plus a
# custom forecast override for the first month. The entry's (month, tag_filters)
# must match one of the budget's own entries.
resource "datadog_cost_budget" "example" {
  name          = "Engineering Q1 Budget"
  metrics_query = "sum:aws.cost.amortized{service:ec2} by {service}"
  start_month   = 202601
  end_month     = 202602

  entries {
    amount = 1000
    month  = 202601
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 1200
    month  = 202602
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}

resource "datadog_cost_custom_forecast" "example" {
  budget_uid = datadog_cost_budget.example.id

  entries {
    amount = 900
    month  = 202601
    tag_filters {
      tag_key   = "service"
      tag_value = "ec2"
    }
  }
}
