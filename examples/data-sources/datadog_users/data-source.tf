data "datadog_users" "test" {
  filter        = "user.name@company.com"
  filter_status = "Active,Pending"
}

data "datadog_users" "by_handle" {
  filter_handle = "user.name@company.com"
}
