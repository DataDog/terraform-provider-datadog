package test

import (
	"testing"
)

const datadogPowerpackAlertGraphTest = `
 resource "datadog_monitor" "downtime_monitor" { 
   name = "monitor"
   type = "metric alert" 
   message = "some message Notify: @hipchat-channel" 
   escalation_message = "the situation has escalated @pagerduty" 
  
   query = "avg(last_1h):avg:aws.ec2.cpu{environment:foo,host:foo} by {host} > 2" 
  
   monitor_thresholds { 
     warning = "1.0" 
     critical = "2.0" 
   } 
 } 

resource "datadog_powerpack" "alert_graph_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
    widget {
      alert_graph_definition {
        alert_id  = "${datadog_monitor.downtime_monitor.id}"
        viz_type  = "timeseries"
        title     = "Widget Title"
        title_align = "center"
        title_size = "20"
      }
    }
}
`

var datadogPowerpackAlertGraphTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 1",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Alert Graph widget
	"widget.0.alert_graph_definition.0.viz_type = timeseries",
	"widget.0.alert_graph_definition.0.title = Widget Title",
	"widget.0.alert_graph_definition.0.title_align = center",
	"widget.0.alert_graph_definition.0.title_size = 20",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackAlertGraph(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackAlertGraphTest, "datadog_powerpack.alert_graph_powerpack", datadogPowerpackAlertGraphTestAsserts)
}
