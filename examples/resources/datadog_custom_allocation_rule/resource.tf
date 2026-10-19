# Create new datadog_custom_allocation_rule resource

resource "datadog_custom_allocation_rule" "my_allocation_rule" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "ec2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "my-allocation-rule"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "env"
      value     = "prod"
    }
    granularity = "daily"
    method      = "even"
  }
}

# A "Dynamic by metric" rule, which splits costs by each destination's share of a
# metric rather than by spend. The query is supplied as JSON using Datadog's
# formulas-and-functions request format.
resource "datadog_custom_allocation_rule" "my_timeseries_allocation_rule" {
  costs_to_allocate {
    condition = "is"
    tag       = "azure_product_family"
    value     = "dbforpostgresql"
  }
  enabled       = true
  providernames = ["azure"]
  rule_name     = "postgres-by-query-time"
  strategy {
    granularity = "daily"
    method      = "proportional_timeseries"

    # Partition the allocation. Each tag key listed here must also appear in the
    # query's group-by.
    evaluate_grouped_by_tag_keys = ["env"]

    based_on_timeseries {
      json = jsonencode({
        response_format = "timeseries"
        queries = [{
          name        = "query1"
          data_source = "metrics"
          query       = "sum:postgresql.queries.time{*} by {user,env}.as_count()"
        }]
        formulas = [{ formula = "query1" }]
      })
    }
  }
}
