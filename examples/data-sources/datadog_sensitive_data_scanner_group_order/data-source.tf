data "datadog_sensitive_data_scanner_group_order" "foo" {}

resource "datadog_sensitive_data_scanner_group_order" "foobar" {
  group_ids = data.datadog_sensitive_data_scanner_group_order.foo.group_ids
}
