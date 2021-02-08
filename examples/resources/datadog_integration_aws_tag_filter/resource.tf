resource "datadog_integration_aws_tag_filter" "foo" {
  account_id = "123456789010"
  namespace = "sqs"
  tag_filter_str = "key:value"
}
