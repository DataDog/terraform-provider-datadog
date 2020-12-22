# Create a Datadog logs archive:

resource "datadog_logs_archive" "my_s3_archive" {
  name  = "my s3 archive"
  query = "service:myservice"
  s3 = {
    bucket     = "my-bucket"
    path       = "/path/foo"
    account_id = "001234567888"
    role_name  = "my-role-name"
  }
}
