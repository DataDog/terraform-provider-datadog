# Import existing APM retention filter 
terraform import datadog_apm_retention_filter.foo <filter_id>

# Import default APM retention filter.
# Note: default filter name and query cannot be updated

# terraform version < 1.5.0
# Import using default filter id
terraform import datadog_apm_retention_filter.foo <filter_id>

# terraform version >= 1.5.0
# Generate terraform configuration file using terraform plan command and import block
# See: https://developer.hashicorp.com/terraform/language/import
: '
# main.tf
```
import {
  to = datadog_apm_retention_filter.error_default
  id = "<default_filter_id>"
}
```

# Generate terraform configuration
'
terraform plan -generate-config-out=generated.tf
terraform apply
