package test

import (
	"strings"
	"testing"
)

// dashboardV2Config takes a v1 dashboard config and returns a v2 config by replacing
// the resource type and resource name. This enables v2 tests to reuse v1 configs
// and cassettes, proving that the FieldSpec engine is backward-compatible.
func dashboardV2Config(v1Config string, v1ResourceName string) (v2Config string, v2ResourceName string) {
	v2ResourceName = strings.Replace(v1ResourceName, "datadog_dashboard.", "datadog_dashboard_v2.", 1)
	// Extract the resource local name (e.g. "timeseries_dashboard")
	localName := strings.TrimPrefix(v1ResourceName, "datadog_dashboard.")
	v2Config = strings.Replace(v1Config,
		`resource "datadog_dashboard" "`+localName+`"`,
		`resource "datadog_dashboard_v2" "`+localName+`"`, 1)
	return v2Config, v2ResourceName
}

// Timeseries — complex widget with legacy queries, markers, events, custom_links, yaxis
func TestAccDatadogDashboardV2Timeseries(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardTimeseriesConfig, "datadog_dashboard.timeseries_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardTimeseries", config, name, datadogDashboardTimeseriesAsserts)
}

func TestAccDatadogDashboardV2Timeseries_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardTimeseriesConfigImport, "datadog_dashboard.timeseries_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardTimeseries_import", config, name)
}

// Note — simple widget (no separate import config)
func TestAccDatadogDashboardV2Note(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardNoteConfig, "datadog_dashboard.note_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardNote", config, name, datadogDashboardNoteAsserts)
}

func TestAccDatadogDashboardV2Note_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardNoteConfig, "datadog_dashboard.note_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardNote_import", config, name)
}

// QueryValue — formula queries, conditional formats
func TestAccDatadogDashboardV2QueryValue(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardQueryValueConfig, "datadog_dashboard.query_value_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardQueryValue", config, name, datadogDashboardQueryValueAsserts)
}

func TestAccDatadogDashboardV2QueryValue_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardQueryValueConfigImport, "datadog_dashboard.query_value_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardQueryValue_import", config, name)
}

// TopList — legacy and formula queries
func TestAccDatadogDashboardV2TopList(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardTopListConfig, "datadog_dashboard.top_list_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardTopList", config, name, datadogDashboardTopListAsserts)
}

func TestAccDatadogDashboardV2TopList_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardTopListConfigImport, "datadog_dashboard.top_list_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardTopList_import", config, name)
}

// HeatMap — yaxis, events
func TestAccDatadogDashboardV2HeatMap(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardHeatMapConfig, "datadog_dashboard.heatmap_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardHeatMap", config, name, datadogDashboardHeatMapAsserts)
}

func TestAccDatadogDashboardV2HeatMap_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardHeatMapConfigImport, "datadog_dashboard.heatmap_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardHeatMap_import", config, name)
}

// QueryTable — complex requests with formulas
func TestAccDatadogDashboardV2QueryTable(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardQueryTableConfig, "datadog_dashboard.query_table_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardQueryTable", config, name, datadogDashboardQueryTableAsserts)
}

func TestAccDatadogDashboardV2QueryTable_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardQueryTableConfigImport, "datadog_dashboard.query_table_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardQueryTable_import", config, name)
}

// HostMap — free-form widget with nested request structure (no separate import config)
func TestAccDatadogDashboardV2HostMap(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardHostMapConfig, "datadog_dashboard.hostmap_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardHostMap", config, name, datadogDashboardHostMapAsserts)
}

func TestAccDatadogDashboardV2HostMap_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardHostMapConfig, "datadog_dashboard.hostmap_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardHostMap_import", config, name)
}

// Change — compare_to, change_type, increase_good fields
func TestAccDatadogDashboardV2Change(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardChangeConfig, "datadog_dashboard.change_dashboard")
	testAccDatadogDashboardV2WidgetUtil(t, "TestAccDatadogDashboardChange", config, name, datadogDashboardChangeAsserts)
}

func TestAccDatadogDashboardV2Change_import(t *testing.T) {
	config, name := dashboardV2Config(datadogDashboardChangeConfigImport, "datadog_dashboard.change_dashboard")
	testAccDatadogDashboardV2WidgetUtilImport(t, "TestAccDatadogDashboardChange_import", config, name)
}
