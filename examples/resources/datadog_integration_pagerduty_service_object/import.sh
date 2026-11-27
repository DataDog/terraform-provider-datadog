# Pagerduty service object can be imported using the service_name, while the service_key should be passed by setting the environment variable SERVICE_KEY
# When the configuration uses service_key_wo, omit SERVICE_KEY so the key is not written to the state
SERVICE_KEY=${service_key} terraform import datadog_integration_pagerduty_service_object.foo ${service_name}
