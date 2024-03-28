# This resource is imported using user_id and role_id seperated by `:`.

terraform import datadog_user_role.foo "${role_id}:${user_id}"