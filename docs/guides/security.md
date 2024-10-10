---
subcategory: ""
page_title: "Security Resource Examples"
description: |-
    Security Resource Examples
---

### Security resource examples

This page lists examples of how to create Datadog Security Monitoring resources within Terraform. This list is not exhaustive and will be updated over time to provide more examples.

## Security monitoring rules

### Application security rules

Datadog Application Security Management (ASM) rules protect against application-level attacks aimed at exploiting code-level vulnerabilities. Some examples include Server-Side-Request-Forgery (SSRF), SQL injection, Log4Shell, and Reflected Cross-Site-Scripting (XSS).

```terraform
resource "datadog_security_monitoring_rule" "example_vulnerability_triggered" {
	name = "Example vulnerability triggered"
	query {
		query = "@appsec.security_activity:vulnerability_trigger.example"
		group_by_fields = ["service","env"]
		distinct_fields = []
		aggregation = "count"
		name = "successful_example_trigger"
	}
	message = "The example vulnerability was triggered"
	type = "application_security"
}
```

### Log detection rules

```terraform
resource "datadog_security_monitoring_rule" "example_log_detection_triggered" {
	name = "Example log detection triggered"
	enabled = true
	query {
		query = "source:example service:example-logs-service @evt.name:(example_1 OR example_2)"
		group_by_fields = ["@usr.email"]
		distinct_fields = []
		aggregation = "count"
		name = "example_rule"
	}
	message = "The example log detection rule was triggered"
	type = "log_detection"
}

```

### Signal correlation rules

Signal correlation rules combine multiple signals to create a new signal, allowing you to alert on high-complexity use cases and minimize alert fatigue.

```terraform
resource "datadog_security_monitoring_rule" "signal_correlation_rule" {
	name = "Example signal correlation rule"
	enabled = true
	signal_query {
		group_by_fields = []
		distinct_fields = []
		aggregation = "event_count"
		name = "example_rule_a"
		default_rule_id = "123-456-789"
		correlated_by_fields = ["@userIdentity.arn"]
	}
	signal_query {
		group_by_fields = []
		distinct_fields = []
		aggregation = "event_count"
		name = "example_rule_b"
		default_rule_id = "abc-def-ghi"
		correlated_by_fields = ["@userIdentity.arn"]
	}
	message = "The example signal correlation rule was triggered"
	type = "signal_correlation"
}
```
