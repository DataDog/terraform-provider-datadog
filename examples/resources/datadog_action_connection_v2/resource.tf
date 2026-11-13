resource "datadog_action_connection_v2" "example" {
  name = "AWS connection"
  tags = ["env:staging"]
  integration = {
    aws_integration = {
      type = "AWS"
      credentials = {
        aws_assume_role = {
          type       = "AWSAssumeRole"
          account_id = "111222333444"
          role       = "DatadogWorkflowRole"
        }
      }
    }
  }
}
