# Amazon Web Services integrations can be imported using their account ID and role name separated with a colon (:),
# The EXTERNAL_ID variable is optional and allows to set external_id

# Import will be done and the external_id will be set to the value of the EXTERNAL_ID variable
EXTERNAL_ID=${external_id} terraform import datadog_integration_aws.test ${account_id}:${role_name}

# Import will be done and the external_id will not be set
terraform import datadog_integration_aws.test ${account_id}:${role_name}