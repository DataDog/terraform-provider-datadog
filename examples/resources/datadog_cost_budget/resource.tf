# Simple budget without tag filters
# Note: Must provide entries for all months in the budget period
resource "datadog_cost_budget" "simple" {
  name          = "My AWS Cost Budget"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month   = 202501
  end_month     = 202503

  entries {
    month  = 202501
    amount = 1000
  }
  entries {
    month  = 202502
    amount = 1200
  }
  entries {
    month  = 202503
    amount = 1000
  }
}

# Budget with tag filters
# Note: Must provide entries for all months in the budget period
resource "datadog_cost_budget" "with_tag_filters" {
  name          = "Production AWS Budget"
  metrics_query = "sum:aws.cost.amortized{*} by {environment}"
  start_month   = 202501
  end_month     = 202503

  entries {
    month  = 202501
    amount = 2000
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }
  entries {
    month  = 202502
    amount = 2200
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }
  entries {
    month  = 202503
    amount = 2000
    tag_filters {
      tag_key   = "environment"
      tag_value = "production"
    }
  }
}

# Hierarchical budget with multiple tag combinations
# Note: Order of tags in "by {tag1,tag2}" determines UI hierarchy (parent,child)
# Each unique tag combination must have entries for all months in the budget period
resource "datadog_cost_budget" "hierarchical" {
  name          = "Team-Based AWS Budget"
  metrics_query = "sum:aws.cost.amortized{*} by {team,account}"
  start_month   = 202501
  end_month     = 202503

  entries {
    month  = 202501
    amount = 500
    tag_filters {
      tag_key   = "team"
      tag_value = "backend"
    }
    tag_filters {
      tag_key   = "account"
      tag_value = "staging"
    }
  }
  entries {
    month  = 202502
    amount = 500
    tag_filters {
      tag_key   = "team"
      tag_value = "backend"
    }
    tag_filters {
      tag_key   = "account"
      tag_value = "staging"
    }
  }
  entries {
    month  = 202503
    amount = 500
    tag_filters {
      tag_key   = "team"
      tag_value = "backend"
    }
    tag_filters {
      tag_key   = "account"
      tag_value = "staging"
    }
  }

  entries {
    month  = 202501
    amount = 1500
    tag_filters {
      tag_key   = "team"
      tag_value = "backend"
    }
    tag_filters {
      tag_key   = "account"
      tag_value = "production"
    }
  }
  # ... repeat for additional months and tag combinations
}
