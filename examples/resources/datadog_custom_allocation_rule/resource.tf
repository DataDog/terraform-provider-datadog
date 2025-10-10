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
