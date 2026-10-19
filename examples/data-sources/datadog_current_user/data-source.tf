data "datadog_current_user" "me" {}

locals {
  greeting = "Hello ${data.datadog_current_user.me.name}, your org id is ${data.datadog_current_user.me.org_id}"
}
