package test

import (
	"testing"
)

const datadogPowerpackGeoMapTest = `
resource "datadog_powerpack" "geomap_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
	widget {
		geomap_definition {
		  request {
			q = "avg:system.load.1{*}"
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		  view {
			focus = "WORLD"
		  }
		}
	  }
	  widget {
		geomap_definition {
		  request {
			log_query {
			  index = "*"
			  compute_query {
				aggregation = "count"
			  }
			}
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		  view {
			focus = "WORLD"
		  }
		  custom_link {
			link = "https://app.datadoghq.com/dashboard/lists"
			label = "Test Custom Link label"
		  }
		  custom_link {
			link = "https://app.datadoghq.com/dashboard/lists"
			is_hidden = true
			override_label = "logs"
		  }
		}
	  }
	  widget {
		geomap_definition {
		  request {
			rum_query {
			  index = "*"
			  compute_query {
				aggregation = "count"
			  }
			}
		  }
		  style {
			palette      = "hostmap_blues"
			palette_flip = false
		  }
		  view {
			focus = "WORLD"
		  }
		}
	  }
}
`

var datadogPowerpackGeoMapTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 3",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Geomap widget
	"widget.0.geomap_definition.0.request.0.q = avg:system.load.1{*}",
	"widget.0.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.0.geomap_definition.0.style.0.palette_flip = false",
	"widget.0.geomap_definition.0.view.0.focus = WORLD",
	"widget.1.geomap_definition.0.request.0.log_query.0.compute_query.0.aggregation = count",
	"widget.1.geomap_definition.0.request.0.log_query.0.index = *",
	"widget.1.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.1.geomap_definition.0.style.0.palette_flip = false",
	"widget.1.geomap_definition.0.view.0.focus = WORLD",
	"widget.1.geomap_definition.0.custom_link.# = 2",
	"widget.1.geomap_definition.0.custom_link.0.label = Test Custom Link label",
	"widget.1.geomap_definition.0.custom_link.0.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.geomap_definition.0.custom_link.1.override_label = logs",
	"widget.1.geomap_definition.0.custom_link.1.link = https://app.datadoghq.com/dashboard/lists",
	"widget.1.geomap_definition.0.custom_link.1.is_hidden = true",
	"widget.2.geomap_definition.0.request.0.rum_query.0.compute_query.0.aggregation = count",
	"widget.2.geomap_definition.0.request.0.rum_query.0.index = *",
	"widget.2.geomap_definition.0.style.0.palette = hostmap_blues",
	"widget.2.geomap_definition.0.style.0.palette_flip = false",
	"widget.2.geomap_definition.0.view.0.focus = WORLD",
	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackGeoMap(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackGeoMapTest, "datadog_powerpack.geomap_powerpack", datadogPowerpackGeoMapTestAsserts)
}
