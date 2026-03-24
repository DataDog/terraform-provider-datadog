package test

import (
	"testing"
)

const datadogDashboardSpansConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
  title = "{{uniq}}"
  layout_type = "ordered"
  reflow_type = "auto"
  description = "Created using the Datadog provider in Terraform"

  widget {
    timeseries_definition {
      live_span      = "1h"
	  hide_incomplete_cost_data = true
      title          = "CPU Utilization?"
      request {
        query {
          metric_query {
            data_source     = "metrics"
            name            = "query1"
            query           = "avg:system.cpu.user{*}"
          }
        }
      }
    }
  }
}
`

var datadogDashboardSpansAsserts = []string{
	"title = {{uniq}}",
	"layout_type = ordered",
	"reflow_type = auto",
	"description = Created using the Datadog provider in Terraform",
	"widget.0.timeseries_definition.0.live_span = 1h",
	"widget.0.timeseries_definition.0.hide_incomplete_cost_data = true",
}

func TestAccDatadogDashboardSpans(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSpansConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardSpansAsserts)
}

func TestAccDatadogDashboardSpans_import(t *testing.T) {
	testAccDatadogDashboardWidgetUtilImport(t, datadogDashboardSpansConfig, "datadog_dashboard.timeseries_dashboard")
}

// Test with live_span only (no hide_incomplete_cost_data)
const datadogDashboardSpansNoHideConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
  title = "{{uniq}}"
  layout_type = "ordered"
  reflow_type = "auto"
  description = "Created using the Datadog provider in Terraform"

  widget {
    timeseries_definition {
      live_span      = "1h"
      title          = "CPU Utilization?"
      request {
        query {
          metric_query {
            data_source     = "metrics"
            name            = "query1"
            query           = "avg:system.cpu.user{*}"
          }
        }
      }
    }
  }
}
`

var datadogDashboardSpansNoHideAsserts = []string{
	"title = {{uniq}}",
	"widget.0.timeseries_definition.0.live_span = 1h",
	// hide_incomplete_cost_data not asserted - defaults to false, not stored in state when false
}

func TestAccDatadogDashboardSpans_NoHideIncompleteCostData(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSpansNoHideConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardSpansNoHideAsserts)
}

// Test with hide_incomplete_cost_data = false
const datadogDashboardSpansHideFalseConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
  title = "{{uniq}}"
  layout_type = "ordered"
  reflow_type = "auto"
  description = "Created using the Datadog provider in Terraform"

  widget {
    timeseries_definition {
      live_span      = "1h"
      hide_incomplete_cost_data = false
      title          = "CPU Utilization?"
      request {
        query {
          metric_query {
            data_source     = "metrics"
            name            = "query1"
            query           = "avg:system.cpu.user{*}"
          }
        }
      }
    }
  }
}
`

var datadogDashboardSpansHideFalseAsserts = []string{
	"title = {{uniq}}",
	"widget.0.timeseries_definition.0.live_span = 1h",
	// hide_incomplete_cost_data not asserted - even when explicitly set to false, not sent to API or stored in state
}

func TestAccDatadogDashboardSpans_HideIncompleteCostDataFalse(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSpansHideFalseConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardSpansHideFalseAsserts)
}

// Test with different time spans
const datadogDashboardSpansDifferentUnitsConfig = `
resource "datadog_dashboard" "timeseries_dashboard" {
  title = "{{uniq}}"
  layout_type = "ordered"
  reflow_type = "auto"
  description = "Created using the Datadog provider in Terraform"

  widget {
    timeseries_definition {
      live_span      = "5m"
      hide_incomplete_cost_data = true
      title          = "5 Minutes"
      request {
        query {
          metric_query {
            data_source     = "metrics"
            name            = "query1"
            query           = "avg:system.cpu.user{*}"
          }
        }
      }
    }
  }

  widget {
    timeseries_definition {
      live_span      = "1d"
      hide_incomplete_cost_data = true
      title          = "1 Day"
      request {
        query {
          metric_query {
            data_source     = "metrics"
            name            = "query2"
            query           = "avg:system.mem.used{*}"
          }
        }
      }
    }
  }

  widget {
    timeseries_definition {
      live_span      = "1w"
      hide_incomplete_cost_data = true
      title          = "1 Week"
      request {
        query {
          metric_query {
            data_source     = "metrics"
            name            = "query3"
            query           = "avg:system.disk.used{*}"
          }
        }
      }
    }
  }
}
`

var datadogDashboardSpansDifferentUnitsAsserts = []string{
	"title = {{uniq}}",
	"widget.0.timeseries_definition.0.live_span = 5m",
	"widget.0.timeseries_definition.0.hide_incomplete_cost_data = true",
	"widget.1.timeseries_definition.0.live_span = 1d",
	"widget.1.timeseries_definition.0.hide_incomplete_cost_data = true",
	"widget.2.timeseries_definition.0.live_span = 1w",
	"widget.2.timeseries_definition.0.hide_incomplete_cost_data = true",
}

func TestAccDatadogDashboardSpans_DifferentUnits(t *testing.T) {
	testAccDatadogDashboardWidgetUtil(t, datadogDashboardSpansDifferentUnitsConfig, "datadog_dashboard.timeseries_dashboard", datadogDashboardSpansDifferentUnitsAsserts)
}
