# Microsoft Azure integrations can be imported using their `tenant name` and `client` id separated with a colon (`:`).
# The client_secret should be passed by setting the environment variable CLIENT_SECRET
terraform import datadog_integration_azure.sandbox ${tenant_name}:${client_id}
