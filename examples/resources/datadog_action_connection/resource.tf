resource "datadog_action_connection" "aws_connection" {
  name = "My AWS Connection"

  aws {
    assume_role {
      account_id = "123456789012"
      role       = "role2"
    }
  }
}

variable "token1" {
  type      = string
  sensitive = true
}

variable "token2" {
  type      = string
  sensitive = true
}

resource "datadog_action_connection" "http_connection" {
  name = "My HTTP connection with token auth"

  http {
    base_url = "https://catfact.ninja"

    token_auth {
      token {
        type  = "SECRET"
        name  = "token1"
        value = var.token1
      }

      token {
        type  = "SECRET"
        name  = "token2"
        value = var.token2
      }

      header {
        name  = "header-one"
        value = "headerval"
      }

      header {
        name  = "h2"
        value = "{{ token1 }} test"
      }

      url_parameter {
        name  = "param1"
        value = "{{ token1 }}"
      }

      url_parameter {
        name  = "param2"
        value = "paramVal2"
      }

      body {
        content_type = "application/json"
        content = jsonencode({
          key   = "mykey"
          value = "maybe with a secret: {{ token2 }}"
        })
      }
    }
  }
}
