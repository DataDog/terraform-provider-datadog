package test

import (
	"testing"
)

const datadogPowerpackCheckStatusTest = `
resource "datadog_powerpack" "check_status_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	
    widget {
      check_status_definition {
        check     = "aws.ecs.agent_connected"
        grouping  = "cluster"
        group_by  = ["account", "cluster"]
        tags      = ["account:demo", "cluster:awseb-ruthebdog-env-8-dn3m6u3gvk"]
        title     = "Widget Title!!"
      }
    }
}
`

var datadogPowerpackCheckStatusTestAsserts = []string{
	"tags.# = 1",
	"tags.0 = tag:foo1",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.check_status_definition.0.check = aws.ecs.agent_connected",
	"widget.0.check_status_definition.0.grouping = cluster",
	"widget.0.check_status_definition.0.group_by.# = 2",
	"widget.0.check_status_definition.0.group_by.0 = account",
	"widget.0.check_status_definition.0.group_by.1 = cluster",
	"widget.0.check_status_definition.0.tags.# = 2",
	"widget.0.check_status_definition.0.tags.0 = account:demo",
	"widget.0.check_status_definition.0.tags.1 = cluster:awseb-ruthebdog-env-8-dn3m6u3gvk",
	"widget.0.check_status_definition.0.title = Widget Title!!",
}

func TestAccDatadogPowerpackCheckStatus(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackCheckStatusTest, "datadog_powerpack.check_status_powerpack", datadogPowerpackCheckStatusTestAsserts)
}
