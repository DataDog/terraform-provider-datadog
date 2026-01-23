data "datadog_security_monitoring_critical_assets" "all" {
}

output "critical_assets_count" {
  value = length(data.datadog_security_monitoring_critical_assets.all.critical_assets)
}

output "critical_assets" {
  value = data.datadog_security_monitoring_critical_assets.all.critical_assets
}
