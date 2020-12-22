# Amazon Web Services integrations can be imported using their account ID and role name separated with a colon (:), while the external_id should be passed by setting an environment variable called EXTERNAL_ID

EXTERNAL_ID=${external_id} terraform import datadog_integration_aws.test ${account_id}:${role_name}

