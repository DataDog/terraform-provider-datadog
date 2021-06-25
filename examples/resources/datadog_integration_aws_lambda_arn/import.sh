# Amazon Web Services Lambda ARN integrations can be imported using their account_id and lambda_arn separated with a space (` `).
terraform import datadog_integration_aws_lambda_arn.test "1234567890 arn:aws:lambda:us-east-1:1234567890:function:datadog-forwarder-Forwarder"
