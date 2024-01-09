data "datadog_users" "test" {
  filter        = "user.name@company.com"
  filter_status = "Active,Pending"
}
