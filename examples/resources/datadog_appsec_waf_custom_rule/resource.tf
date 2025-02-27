# Create a new WAF custom rule to block a custom IoC
resource "datadog_appsec_waf_custom_rule" "ioc000" {
  name = "Block requests from a bad actor"

  blocking = true
  enabled  = true

  tags = {
    category = "attack_attempt"
    type     = "custom_ioc"
  }

  path_glob = "/db/*"

  conditions {
    operator = "match_regex"
    parameters {
      inputs {
        address = "server.db.statement"
      }
      regex = "stmt.*"
    }
  }

  action {
    action = "redirect_request"
    parameters {
      status_code = 302
      location    = "/blocking"
    }
  }
}


# Create a WAF custom rule to track business logic events
resource "datadog_appsec_waf_custom_rule" "biz000" {
  name = "Track payments"

  blocking = false
  enabled  = true

  tags = {
    category = "business_logic"
    type     = "payment.checkout"
  }

  path_glob = "/cart/*"

  conditions {
    operator = "capture_data"
    parameters {
      inputs {
        address  = "server.request.query"
        key_path = ["payment_id"]
      }
      value = "payment"
    }
  }

  scope {
    env     = "prod"
    service = "paymentsvc"
  }
}
