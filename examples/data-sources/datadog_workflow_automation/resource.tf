resource "datadog_workflow_automation" "workflow" {
  name        = "My Workflow"
  description = "This is my workflow"
  tags        = ["team:my-team", "service:my-service", "env:dev", "foo:bar"]
  published   = true

  spec_json = jsonencode(
    {
      "triggers" : [
        {
          "startStepNames" : [
            "No_op"
          ],
          "appTrigger" : {}
        }
      ],
      "steps" : [
        {
          "name" : "No_op",
          "actionId" : "com.datadoghq.core.noop",
          "display" : {
            "bounds" : {
              "x" : 0,
              "y" : 228
            }
          }
        }
      ]
    }
  )
}