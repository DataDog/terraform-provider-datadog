# Microsoft Azure integrations can be imported using their `tenant name` and `client` id separated with a colon (`:`).
terraform import datadog_integration_azure.sandbox ${tenant_name}:${client_id}

