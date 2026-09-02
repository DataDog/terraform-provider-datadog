package datadog

import (
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/stretchr/testify/assert"
)

func TestBuildDataQualityMonitorOptions(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want *datadogV1.MonitorFormulaAndFunctionDataQualityMonitorOptions
	}{
		{
			name: "no options set",
			data: map[string]interface{}{},
			want: nil,
		},
		{
			name: "empty values are ignored",
			data: map[string]interface{}{
				"custom_sql":       "",
				"custom_where":     "",
				"group_by_columns": []interface{}{},
				"sensitivity":      0.0,
			},
			want: nil,
		},
		{
			name: "sensitivity only",
			data: map[string]interface{}{"sensitivity": 2.5},
			want: func() *datadogV1.MonitorFormulaAndFunctionDataQualityMonitorOptions {
				o := datadogV1.NewMonitorFormulaAndFunctionDataQualityMonitorOptions()
				o.SetSensitivity(2.5)
				return o
			}(),
		},
		{
			name: "all options set",
			data: map[string]interface{}{
				"custom_sql":          "SELECT 1",
				"custom_where":        "database = 'production'",
				"group_by_columns":    []interface{}{"column1", "column2"},
				"crontab_override":    "0 0 * * *",
				"model_type_override": "freshness",
				"sensitivity":         4.0,
			},
			want: func() *datadogV1.MonitorFormulaAndFunctionDataQualityMonitorOptions {
				o := datadogV1.NewMonitorFormulaAndFunctionDataQualityMonitorOptions()
				o.SetCustomSql("SELECT 1")
				o.SetCustomWhere("database = 'production'")
				o.SetGroupByColumns([]string{"column1", "column2"})
				o.SetCrontabOverride("0 0 * * *")
				o.SetModelTypeOverride(datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYMODELTYPEOVERRIDE_FRESHNESS)
				o.SetSensitivity(4.0)
				return o
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildDataQualityMonitorOptions(tt.data))
		})
	}
}

func TestBuildTerraformDataQualityMonitorVariablesRoundTrip(t *testing.T) {
	query := map[string]interface{}{
		"name":        "query1",
		"data_source": "data_quality_metrics",
		"measure":     "row_count",
		"filter":      "search for column where `database:production`",
		"monitor_options": []interface{}{map[string]interface{}{
			"sensitivity":         2.5,
			"model_type_override": "percentage",
		}},
	}

	definition := buildMonitorFormulaAndFunctionDataQualityQuery(query)
	terraformVariables := buildTerraformDataQualityMonitorVariables([]datadogV1.MonitorFormulaAndFunctionQueryDefinition{*definition})

	queries := terraformVariables[0]["data_quality_query"].([]map[string]interface{})
	options := queries[0]["monitor_options"].([]map[string]interface{})[0]

	assert.Equal(t, 2.5, *options["sensitivity"].(*float64))
	assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYMODELTYPEOVERRIDE_PERCENTAGE, *options["model_type_override"].(*datadogV1.MonitorFormulaAndFunctionDataQualityModelTypeOverride))
}
