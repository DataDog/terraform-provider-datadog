# Amazon Web Services log filter resource can be imported using their account ID and namespace separated with a colon (:).
terraform import datadog_integration_aws_tag_filter.foo ${account_id}:${namespace}
