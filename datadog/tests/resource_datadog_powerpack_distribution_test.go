package test

import (
	"testing"
)

const datadogPowerpackDistributionTest = `
resource "datadog_powerpack" "distribution_powerpack" {
	name         = "{{uniq}}"
    tags = ["tag:foo1"]
	description   = "Created using the Datadog provider in Terraform"
	template_variables {
		defaults = ["defaults"]
		name     = "datacenter"
	}
	widget {
		distribution_definition {
			title = "Avg of system.cpu.user over account:prod by service,account"
			title_align = "left"
			title_size = "16"
			show_legend = "true"
			legend_size = "2"
			request {
				q = "avg:system.cpu.user{account:prod} by {service,account}"
				style {
					palette = "purple"
				}
			}
		}
	}
	widget {
		distribution_definition {
			title = "Avg of system.cpu.user over account:prod by service,account"
			title_align = "left"
			title_size = "16"
			show_legend = "true"
			legend_size = "2"
			request {
				apm_stats_query {
					service = "service"
					env = "env"
					primary_tag = "tag:*"
					name = "name"
					row_type = "resource"
				}
			}
		}
	}
}
`

var datadogPowerpackDistributionTestAsserts = []string{
	// Powerpack metadata
	"name = {{uniq}}",
	"description = Created using the Datadog provider in Terraform",
	"widget.# = 2",
	"tags.# = 1",
	"tags.0 = tag:foo1",
	// Distribution widget
	"widget.0.distribution_definition.0.title = Avg of system.cpu.user over account:prod by service,account",
	"widget.0.distribution_definition.0.title_size = 16",
	"widget.0.distribution_definition.0.title_align = left",
	"widget.0.distribution_definition.0.show_legend = true",
	"widget.0.distribution_definition.0.legend_size = 2",
	"widget.0.distribution_definition.0.request.0.q = avg:system.cpu.user{account:prod} by {service,account}",
	"widget.0.distribution_definition.0.request.0.style.0.palette = purple",
	"widget.1.distribution_definition.0.title = Avg of system.cpu.user over account:prod by service,account",
	"widget.1.distribution_definition.0.title_size = 16",
	"widget.1.distribution_definition.0.title_align = left",
	"widget.1.distribution_definition.0.show_legend = true",
	"widget.1.distribution_definition.0.legend_size = 2",
	"widget.1.distribution_definition.0.request.0.apm_stats_query.0.service = service",
	"widget.1.distribution_definition.0.request.0.apm_stats_query.0.env = env",
	"widget.1.distribution_definition.0.request.0.apm_stats_query.0.primary_tag = tag:*",
	"widget.1.distribution_definition.0.request.0.apm_stats_query.0.name = name",
	"widget.1.distribution_definition.0.request.0.apm_stats_query.0.row_type = resource",

	// Template Variables
	"template_variables.# = 1",
	"template_variables.0.name = datacenter",
	"template_variables.0.defaults.# = 1",
	"template_variables.0.defaults.0 = defaults",
}

func TestAccDatadogPowerpackDistribution(t *testing.T) {
	testAccDatadogPowerpackWidgetUtil(t, datadogPowerpackDistributionTest, "datadog_powerpack.distribution_powerpack", datadogPowerpackDistributionTestAsserts)
}
