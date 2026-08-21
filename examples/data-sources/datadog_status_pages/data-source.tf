data "datadog_status_pages" "all" {}

data "datadog_status_pages" "by_name" {
  name = "abridge.com"
}
