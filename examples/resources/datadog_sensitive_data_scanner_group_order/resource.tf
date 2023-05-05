# Create new sensitive_data_scanner_group_order resource

resource "datadog_sensitive_data_scanner_group_order" "mygrouporder" {
  group_ids = [
    "group-id-1",
    "group-id-2",
    "group-id-3"
  ]
}
