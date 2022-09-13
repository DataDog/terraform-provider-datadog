resource "datadog_service_definition_yaml" "service_definition" {
  service_definition = <<EOF
schema-version: v2
dd-service: shopping-cart
team: E Commerce
contacts:
  - name: Support Email
    type: email
    contact: team@shopping.com
  - name: Support Slack
    type: slack
    contact: https://www.slack.com/archives/shopping-cart
repos:
  - name: shopping-cart source code
    provider: github
    url: http://github/shopping-cart
docs:
  - name: shopping-cart architecture
    provider: gdoc
    url: https://google.drive/shopping-cart-architecture
  - name: shopping-cart service Wiki
    provider: wiki
    url: https://wiki/shopping-cart
links:
  - name: shopping-cart runbook
    type: runbook
    url: https://runbook/shopping-cart
tags:
  - business-unit:retail
  - cost-center:engineering
integrations:
  pagerduty: https://www.pagerduty.com/service-directory/Pshopping-cart
extensions:
  datadoghq.com/shopping-cart:
    customField: customValue
EOF
}
