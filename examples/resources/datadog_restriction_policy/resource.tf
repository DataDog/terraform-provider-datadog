
# Create new restriction_policy resource


resource "datadog_restriction_policy" "foo" {
  resource_id = "security-rule:abc-def-ghi"
  bindings {
    principals = ["role:00000000-0000-1111-0000-000000000000"]
    relation   = "editor"
  }
  bindings {
    principals = ["org:10000000-0000-1111-0000-000000000000"]
    relation   = "viewer"
  }
}
