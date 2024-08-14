# This resource is imported using user_id and role_id seperated by `:`.

terraform import datadog_user_role.user_with_admin_role "${role_id}:${user_id}"