resource "datadog_cost_budget" "example" {
  name          = "My AWS Cost Budget"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month   = 202401
  end_month     = 202412

  entries {
    amount = 1000
    month  = 202401
  }
  entries {
    amount = 1000
    month  = 202402
  }
  entries {
    amount = 1000
    month  = 202403
  }
  entries {
    amount = 1000
    month  = 202404
  }
  entries {
    amount = 1000
    month  = 202405
  }
  entries {
    amount = 1000
    month  = 202406
  }
  entries {
    amount = 1000
    month  = 202407
  }
  entries {
    amount = 1000
    month  = 202408
  }
  entries {
    amount = 1000
    month  = 202409
  }
  entries {
    amount = 1000
    month  = 202410
  }
  entries {
    amount = 1000
    month  = 202411
  }
  entries {
    amount = 1000
    month  = 202412
  }
}

# Budget with tag filters
resource "datadog_cost_budget" "example_with_filters" {
  name          = "My Filtered Budget"
  metrics_query = "sum:aws.cost.amortized{*}"
  start_month   = 202401
  end_month     = 202412

  entries {
    amount = 500
    month  = 202401
    tag_filters {
      tag_key   = "account"
      tag_value = "ec2"
    }
  }
  entries {
    amount = 500
    month  = 202402
    tag_filters {
      tag_key   = "account"
      tag_value = "ec2"
    }
  }
}

