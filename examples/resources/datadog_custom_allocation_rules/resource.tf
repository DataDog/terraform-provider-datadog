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

# Example 1: Preserve mode (default) - allows unmanaged rules to exist at the end
# This will preserve any existing rules created outside of Terraform as long as they are at the end
resource "datadog_custom_allocation_rules" "preserve_order" {
  # override_ui_defined_resources = false (default)
  rule_ids = [
    datadog_custom_allocation_rule.rule_1.id,
    datadog_custom_allocation_rule.rule_2.id,
    datadog_custom_allocation_rule.rule_3.id
  ]
}

# Example 2: Override mode - deletes all unmanaged rules and maintains strict order
# This will delete any rules not defined in Terraform and enforce the exact order specified
resource "datadog_custom_allocation_rules" "override_order" {
  override_ui_defined_resources = true
  rule_ids = [
    datadog_custom_allocation_rule.rule_1.id,
    datadog_custom_allocation_rule.rule_2.id,
    datadog_custom_allocation_rule.rule_3.id
  ]
}
