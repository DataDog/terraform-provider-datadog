# Example Security Monitoring Rule JSON
resource "datadog_security_monitoring_rule_json" "security_rule_json" {
  json = <<EOF
{
  "name": "High error rate security monitoring",
  "isEnabled": true,
  "type": "log_detection",
  "message": "High error rate detected in logs",
  "tags": ["env:prod", "security"],
  "cases": [
    {
      "name": "high case",
      "status": "high",
      "condition": "errors > 100 && warnings > 1000",
      "notifications": ["@security-team"]
    }
  ],
  "queries": [
    {
      "name": "errors",
      "query": "status:error",
      "aggregation": "count",
      "dataSource": "logs",
      "groupByFields": ["service", "env"]
    },
    {
      "name": "warnings",
      "query": "status:warning",
      "aggregation": "count",
      "dataSource": "logs",
      "groupByFields": ["service", "env"]
    }
  ],
  "options": {
    "evaluationWindow": 300,
    "keepAlive": 600,
    "maxSignalDuration": 900,
    "detectionMethod": "threshold"
  }
}
EOF
}
