resource "datadog_custom_allocation_rule" "rule_1" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonEC2"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "my-custom-rule-1"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonEC2"
    }
    method = "even"
  }
}

resource "datadog_custom_allocation_rule" "rule_2" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonS3"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "my-custom-rule-2"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonS3"
    }
    method = "even"
  }
}

resource "datadog_custom_allocation_rule" "rule_3" {
  costs_to_allocate {
    condition = "is"
    tag       = "aws_product"
    value     = "AmazonRDS"
  }
  enabled       = true
  providernames = ["aws"]
  rule_name     = "my-custom-rule-3"
  strategy {
    allocated_by_tag_keys = ["team"]
    based_on_costs {
      condition = "is"
      tag       = "aws_product"
      value     = "AmazonRDS"
    }
    method = "even"
  }
}

# Manage the order of custom allocation rules
resource "datadog_custom_allocation_rules" "order" {
  rule_ids = [
    datadog_custom_allocation_rule.rule_1.id,
    datadog_custom_allocation_rule.rule_2.id,
    datadog_custom_allocation_rule.rule_3.id
  ]
}
