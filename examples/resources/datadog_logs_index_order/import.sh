# The Datadog Terraform Provider does not support the creation and deletion of index orders. There must be at most one `datadog_logs_index_order` resource
# `<name>` can be whatever you specify in your code. Datadog does not store the name on the server.
terraform import <datadog_logs_index_order.name> <name>
