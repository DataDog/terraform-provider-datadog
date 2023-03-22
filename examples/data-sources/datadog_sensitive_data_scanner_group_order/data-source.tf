data "datadog_sensitive_data_scanner_group_order" "foo" {}

resource "datadog_sensitive_data_scanner_group_order" "foobar" {
    groups = data.datadog_sensitive_data_scanner_group_order.foo.groups
}
