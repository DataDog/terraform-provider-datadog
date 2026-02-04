# Budget with multiple tag combinations
# Note: Each unique tag combination needs its own budget_line block
resource "datadog_cost_budget" "with_tags" {
  name          = "Multi-Environment Budget"
  metrics_query = "sum:aws.cost.amortized{*} by {environment}"
  start_month   = 202601
  end_month     = 202603

  budget_line {
    amounts = {
      "202601" = 2000
      "202602" = 2200
      "202603" = 2000
    }
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }

  budget_line {
    amounts = {
      "202601" = 1000
      "202602" = 1100
      "202603" = 1000
    }
    tag_filters {
      tag_key   = "environment"
      tag_value = "staging"
    }
  }
}
