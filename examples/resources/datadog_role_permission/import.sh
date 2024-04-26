# This resource is imported using user_id and role_id seperated by `:`.

terraform import datadog_role_permission.role_with_monitors_write "${role_id}:${permission_id}"