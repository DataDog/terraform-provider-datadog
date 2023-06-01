resource "datadog_service_account" "example" {
  name  = "A Test Service Account"
  email = "example-service-account@datadoghq.com"
}

resource "datadog_service_account_application_key" "example" {
  name               = "A Test Service Account Application Key"
  service_account_id = datadog_service_account.example.id
}