
# Create new restriction_policy resource

data "datadog_current_user" "me" {}

resource "datadog_restriction_policy" "foo" {
  resource_id = "security-rule:abc-def-ghi"
  bindings {
    principals = [
      "user:${data.datadog_current_user.me.id}",
      "team:00000000-0000-1111-0000-000000000000",
    ]
    relation = "editor"
  }
  bindings {
    principals = ["org:${data.datadog_current_user.me.org_id}"]
    relation   = "viewer"
  }
}
