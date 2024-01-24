# Create new sensitive_data_scanner_rule resource in a sensitive_data_scanner_group

resource "datadog_sensitive_data_scanner_group" "mygroup" {
  name        = "My new scanning group"
  description = "A relevant description"
  filter {
    query = "service:my-service"
  }
  is_enabled   = true
  product_list = ["apm"]
}

resource "datadog_sensitive_data_scanner_rule" "myrule" {
  name                = "My new rule"
  description         = "Another description"
  group_id            = datadog_sensitive_data_scanner_group.mygroup.id
  excluded_namespaces = ["username"]
  is_enabled          = true
  pattern             = "myregex"
  tags                = ["sensitive_data:true"]
  text_replacement {
    number_of_chars    = 0
    replacement_string = ""
    type               = "hash"
  }
  included_keyword_configuration {
    keywords        = ["cc", "credit card"]
    character_count = 30
  }
}

data "datadog_sensitive_data_scanner_standard_pattern" "aws_sp" {
  filter = "AWS Access Key ID Scanner"
}

resource "datadog_sensitive_data_scanner_rule" "mylibraryrule" {
  name        = "My library rule"
  description = "A description"
  group_id    = datadog_sensitive_data_scanner_group.mygroup.id
  // As standard_pattern_id is provided, the resource MUST NOT contain the "pattern" attribute
  standard_pattern_id = data.datadog_sensitive_data_scanner_standard_pattern.aws_sp.id
  excluded_namespaces = ["username"]
  is_enabled          = true
  tags                = ["sensitive_data:true"]
}