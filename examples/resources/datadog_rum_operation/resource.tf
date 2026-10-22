resource "datadog_rum_operation" "checkout" {
  name           = "checkout"
  display_name   = "Checkout"
  application_id = "00000000-0000-0000-0000-000000000000"
  description    = "From the cart to the order confirmation page."
  category       = "form_submission"
  tags           = ["team:payments"]

  steps = [
    {
      type  = "start"
      nodes = [{ query = "@view.name:/cart @action.name:checkout" }]
    },
    {
      type  = "stop"
      nodes = [{ query = "@view.name:/order/confirmation" }]
    },
    {
      type = "error"
      composite = {
        kind          = "in_order"
        max_window_ms = 60000
        predicates = [
          { query = "@type:action @action.name:pay" },
          { query = "@type:error @error.message:*declined*" },
        ]
      }
    },
  ]
}
