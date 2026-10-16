resource "datadog_rum_operation_journey_link" "checkout" {
  operation_id = datadog_rum_operation.checkout.id
  journey_id   = "<journey_id>"
}
