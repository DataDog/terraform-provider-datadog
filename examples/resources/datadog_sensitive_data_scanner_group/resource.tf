
# Create new sensitive_data_scanner_group resource

resource "datadog_sensitive_data_scanner_group" "mygroup" {
  name        = "My new scanning group"
  description = "A relevant description"
  filter {
    query = "service:my-service"
  }
  is_enabled   = true
  product_list = ["apm", "logs"]
  samplings {
    product = "apm"
    rate    = 100
  }
  samplings {
    product = "logs"
    rate    = 10
  }
}
