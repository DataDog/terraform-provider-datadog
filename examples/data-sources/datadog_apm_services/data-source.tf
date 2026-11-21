# List all APM services reporting in the "prod" environment
data "datadog_apm_services" "prod" {
  filter_env = "prod"
}

# List APM services across all environments
data "datadog_apm_services" "all" {
  filter_env = "*"
}

output "traced_services" {
  value = [for service in data.datadog_apm_services.all.services : service.name if service.is_traced]
}
