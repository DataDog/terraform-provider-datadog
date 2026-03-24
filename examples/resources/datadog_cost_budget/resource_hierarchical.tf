# Hierarchical budget with parent/child tag structure
# Note: Order in "by {tag1,tag2}" determines hierarchy (parent,child)
# Each unique parent+child combination needs its own budget_line block
resource "datadog_cost_budget" "hierarchical" {
  name          = "Team-Based AWS Budget"
  metrics_query = "sum:aws.cost.amortized{*} by {team,environment}"
  start_month   = 202601
  end_month     = 202603

  budget_line {
    amounts = {
      "202601" = 1500
      "202602" = 1600
      "202603" = 1500
    }
    parent_tag_filters {
      tag_key   = "team"
      tag_value = "backend"
    }
    child_tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }

  budget_line {
    amounts = {
      "202601" = 500
      "202602" = 550
      "202603" = 500
    }
    parent_tag_filters {
      tag_key   = "team"
      tag_value = "frontend"
    }
    child_tag_filters {
      tag_key   = "environment"
      tag_value = "staging"
    }
  }
}
