data "datadog_hosts" "foo" {
  include_muted_hosts_data = true
  include_hosts_metadata   = true
}
