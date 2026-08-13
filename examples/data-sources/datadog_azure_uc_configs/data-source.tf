# Retrieve every Azure Usage Cost configuration in the organization.
data "datadog_azure_uc_configs" "example" {}

# Collect the actual-cost export names, grouped per Cloud Cost Management account.
output "actual_export_names" {
  value = [
    for pair in data.datadog_azure_uc_configs.example.azure_uc_configs :
    [for config in pair.configs : config.export_name if config.dataset_type == "actual"]
  ]
}
