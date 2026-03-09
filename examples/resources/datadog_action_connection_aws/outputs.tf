output "connection_id" {
  value = datadog_action_connection.aws.id
}

output "external_id" {
  value     = datadog_action_connection.aws.aws.assume_role.external_id
  sensitive = false
}

output "principal_id" {
  value = datadog_action_connection.aws.aws.assume_role.principal_id
}
