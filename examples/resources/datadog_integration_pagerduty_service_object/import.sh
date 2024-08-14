# Pagerduty service object can be imported using the service_name, while the service_key should be passed by setting the environment variable SERVICE_KEY
SERVICE_KEY=${service_key} terraform import datadog_integration_pagerduty_service_object.foo ${service_name}
