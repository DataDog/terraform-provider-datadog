// Service Definition with v2.1 Schema Definition
resource "datadog_service_definition_yaml" "service_definition_v2_1" {
  service_definition = <<EOF
schema-version: v2.1
dd-service: shopping-cart
team: e-commerce-team
contacts:
  - name: Support Email
    type: email
    contact: team@shopping.com
  - name: Support Slack
    type: slack
    contact: https://www.slack.com/archives/shopping-cart
description: shopping cart service responsible for managing shopping carts
tier: high
lifecycle: production
application: e-commerce
links:
  - name: shopping-cart runbook
    type: runbook
    url: https://runbook/shopping-cart
  - name: shopping-cart architecture
    type: doc
    provider: gdoc
    url: https://google.drive/shopping-cart-architecture
  - name: shopping-cart service Wiki
    type: doc
    provider: wiki
    url: https://wiki/shopping-cart
  - name: shopping-cart source code
    type: repo
    provider: github
    url: http://github/shopping-cart
tags:
  - business-unit:retail
  - cost-center:engineering
integrations:
  pagerduty: 
    service-url: https://www.pagerduty.com/service-directory/Pshopping-cart
extensions:
  mycompany.com/shopping-cart:
    customField: customValue
EOF
}

// Service Definition with v2 Schema Definition
resource "datadog_service_definition_yaml" "service_definition_v2" {
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

// Service Definition with backstage.io Schema
resource "datadog_service_definition_yaml" "service_definition_backstage" {
  service_definition = <<EOF
apiVersion: backstage.io/v1alpha1
kind: Component
metadata:
  annotations:
    backstage.io/techdocs-ref: http://a/b/c
    some.annotation: value
  namespace: default
  name: shopping-cart
  title: Shopping Cart
  description: A shopping cart service
  tags: ["taga:valuea", "tagb:valueb"]
  links:
    - title: Wiki
      url: https://wiki/shopping-cart
      icon: help
  ignore-attribute:
    id: 1
    value: "value"
spec:
  type: service
  lifecycle: production
  owner: e-commerce
  system: retail
EOF
}
