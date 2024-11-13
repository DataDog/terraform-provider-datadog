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
  lifecycle {
    // Use this meta-argument to avoid disabling the group when modifying the 
    // `included_keyword_configuration` field
    create_before_destroy = true
  }
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
  priority = 1
}

data "datadog_sensitive_data_scanner_standard_pattern" "aws_sp" {
  filter = "AWS Access Key ID Scanner"
}

resource "datadog_sensitive_data_scanner_rule" "mylibraryrule_with_custom_included_keywords" {
  name        = "My library rule"
  description = "A description"
  group_id    = datadog_sensitive_data_scanner_group.mygroup.id
  // As standard_pattern_id is provided, the resource MUST NOT contain the "pattern" attribute
  standard_pattern_id = data.datadog_sensitive_data_scanner_standard_pattern.aws_sp.id
  excluded_namespaces = ["username"]
  is_enabled          = true
  tags                = ["sensitive_data:true"]

  // SDS will set the recommended keywords by default. If the user doesn't want to use the recommended keywords,
  // they have to create an empty included keyword configuration (with empty keywords)
  included_keyword_configuration {
    keywords        = ["cc", "credit card"]
    character_count = 30
  }
}

resource "datadog_sensitive_data_scanner_rule" "mylibraryrule_with_recommended_keywords" {
  name        = "My library rule"
  description = "A description"
  group_id    = datadog_sensitive_data_scanner_group.mygroup.id
  // As standard_pattern_id is provided, the resource MUST NOT contain the "pattern" attribute
  standard_pattern_id = data.datadog_sensitive_data_scanner_standard_pattern.aws_sp.id
  excluded_namespaces = ["username"]
  is_enabled          = true
  tags                = ["sensitive_data:true"]

  // SDS will set the recommended keywords by default.
}
