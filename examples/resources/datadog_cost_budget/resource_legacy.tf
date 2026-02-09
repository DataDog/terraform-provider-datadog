# Legacy entries with tag filters (deprecated - use budget_line instead)
# Note: Each unique tag combination must have entries for all months
resource "datadog_cost_budget" "legacy_with_tags" {
  name          = "Production Budget (Legacy)"
  metrics_query = "sum:aws.cost.amortized{*} by {environment}"
  start_month   = 202601
  end_month     = 202603

  entries {
    month  = 202601
    amount = 2000
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }
  entries {
    month  = 202602
    amount = 2200
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }
  entries {
    month  = 202603
    amount = 2000
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }
}

