package test

import (
	"testing"
)

const datadogDashboardV2TimeseriesAnomalyDetectionConfig = `
resource "datadog_dashboard_v2" "timeseries_anomaly_detection_dashboard" {
    title       = "{{uniq}}"
    layout_type = "ordered"

    widget {
        timeseries_definition {
            title = "CPU with anomaly detection"

            anomaly_detection {
                detection_sensitivity = "never_detect"
            }

            request {
                formula {
                    formula_expression = "query1"
                }

                query {
                    metric_query {
                        data_source = "metrics"
                        name        = "query1"
                        query       = "avg:system.cpu.user{*}"
                    }
                }
            }
        }
    }
}
`

var datadogDashboardV2TimeseriesAnomalyDetectionAsserts = []string{
	"title = {{uniq}}",
	"widget.0.timeseries_definition.0.title = CPU with anomaly detection",
	"widget.0.timeseries_definition.0.anomaly_detection.0.detection_sensitivity = never_detect",
	"widget.0.timeseries_definition.0.request.0.formula.0.formula_expression = query1",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.data_source = metrics",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.name = query1",
	"widget.0.timeseries_definition.0.request.0.query.0.metric_query.0.query = avg:system.cpu.user{*}",
}

func TestAccDatadogDashboardV2TimeseriesAnomalyDetection(t *testing.T) {
	config, name := datadogDashboardV2TimeseriesAnomalyDetectionConfig, "datadog_dashboard_v2.timeseries_anomaly_detection_dashboard"
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardV2TimeseriesAnomalyDetection", config, name, datadogDashboardV2TimeseriesAnomalyDetectionAsserts)
}

func TestAccDatadogDashboardV2TimeseriesAnomalyDetection_import(t *testing.T) {
	config, name := datadogDashboardV2TimeseriesAnomalyDetectionConfig, "datadog_dashboard_v2.timeseries_anomaly_detection_dashboard"
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardV2TimeseriesAnomalyDetection_import", config, name)
}
