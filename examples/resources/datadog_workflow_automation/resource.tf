resource "datadog_workflow_automation" "workflow" {
  name        = "Send Email when Monitor Alerts"
  description = "This workflow alerts me by email when my monitor goes off. "
  tags        = ["service:foo","source:alert","team:bar"]
  published   = true

  spec_json = jsonencode(
    {
      "triggers": [
        {
          "startStepNames": [
            "Send_Email"
          ],
          "monitorTrigger": {}
        }
      ],
      "steps": [
        {
          "name": "Send_Email",
          "actionId": "com.datadoghq.email.send",
          "parameters": [
            {
              "name": "to",
              "value": "REPLACE_ME"
            },
            {
              "name": "subject",
              "value": "Monitor \"{{ Source.monitor.name }}\" alerted"
            },
            {
              "name": "message",
              "value": "This message is from {{ WorkflowName }}. \n\nYou can find a link to the monitor here: {{ Source.url }}."
            }
          ],
          "display": {
            "bounds": {
              "x": 0,
              "y": 216
            }
          }
        }
      ],
      "handle": "my-handle"
    }
  )
}