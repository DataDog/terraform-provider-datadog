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

func TestBuildDataQualitySourceToTargetConfig(t *testing.T) {
	config := buildDataQualitySourceToTargetConfig(map[string]interface{}{
		"source": []interface{}{map[string]interface{}{
			"entity_id":        "src-entity",
			"entity_type":      "table",
			"custom_where":     "env = 'prod'",
			"group_by_columns": []interface{}{"env"},
		}},
		"target": []interface{}{map[string]interface{}{
			"entity_id":   "tgt-entity",
			"entity_type": "table",
		}},
		"diff_type":   "diff_percent",
		"entity_type": "table",
	})

	assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDIFFTYPE_DIFF_PERCENT, config.GetDiffType())
	assert.Equal(t, "table", config.GetEntityType())

	source := config.GetSource()
	assert.Equal(t, "src-entity", source.GetEntityId())
	assert.Equal(t, "env = 'prod'", source.GetCustomWhere())
	assert.Equal(t, []string{"env"}, source.GetGroupByColumns())
	// Unset optionals stay absent rather than being written as empty values.
	assert.False(t, source.HasCustomSql())

	target := config.GetTarget()
	assert.Equal(t, "tgt-entity", target.GetEntityId())
	assert.False(t, target.HasCustomWhere())
	assert.False(t, target.HasGroupByColumns())
}

func TestBuildDataQualityModelConfiguration(t *testing.T) {
	tests := []struct {
		name string
		data map[string]interface{}
		want func(*datadogV1.MonitorFormulaAndFunctionDataQualityModelConfiguration)
	}{
		{
			name: "all fields set",
			data: map[string]interface{}{
				"auto_resolve_days":         7,
				"enable_flatline_detection": true,
				"function":                  "DIFF",
				"min_lower_bound_size":      10.0,
				"min_upper_bound_size":      20.0,
				"model_bounds_override":     "UPPER_ONLY",
			},
			want: func(c *datadogV1.MonitorFormulaAndFunctionDataQualityModelConfiguration) {
				assert.Equal(t, int32(7), c.GetAutoResolveDays())
				assert.True(t, c.GetEnableFlatlineDetection())
				assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDIFFFUNCTION_DIFF, c.GetFunction())
				assert.Equal(t, 10.0, c.GetMinLowerBoundSize())
				assert.Equal(t, 20.0, c.GetMinUpperBoundSize())
				assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYMODELBOUNDSOVERRIDE_UPPER_ONLY, c.GetModelBoundsOverride())
			},
		},
		{
			// enable_flatline_detection is schema-defaulted to true, so an explicit
			// false is a real setting and must survive rather than read as unset.
			name: "flatline detection explicitly disabled",
			data: map[string]interface{}{"enable_flatline_detection": false},
			want: func(c *datadogV1.MonitorFormulaAndFunctionDataQualityModelConfiguration) {
				assert.True(t, c.HasEnableFlatlineDetection())
				assert.False(t, c.GetEnableFlatlineDetection())
			},
		},
		{
			name: "zero valued numerics are treated as unset",
			data: map[string]interface{}{
				"auto_resolve_days":    0,
				"min_lower_bound_size": 0.0,
				"min_upper_bound_size": 0.0,
			},
			want: func(c *datadogV1.MonitorFormulaAndFunctionDataQualityModelConfiguration) {
				assert.False(t, c.HasAutoResolveDays())
				assert.False(t, c.HasMinLowerBoundSize())
				assert.False(t, c.HasMinUpperBoundSize())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want(buildDataQualityModelConfiguration(tt.data))
		})
	}
}

func TestBuildTerraformDataQualityMonitorOptionsRoundTrip(t *testing.T) {
	query := map[string]interface{}{
		"name":        "query1",
		"data_source": "data_quality_metrics",
		"measure":     "row_count",
		"filter":      "search for column where `database:production`",
		"monitor_options": []interface{}{map[string]interface{}{
			"source_to_target_config": []interface{}{map[string]interface{}{
				"source":      []interface{}{map[string]interface{}{"entity_id": "src", "entity_type": "table"}},
				"target":      []interface{}{map[string]interface{}{"entity_id": "tgt", "entity_type": "table"}},
				"diff_type":   "absolute",
				"entity_type": "table",
			}},
			"model_configuration": []interface{}{map[string]interface{}{
				"auto_resolve_days":         7,
				"enable_flatline_detection": false,
			}},
		}},
	}

	definition := buildMonitorFormulaAndFunctionDataQualityQuery(query)
	terraformVariables := buildTerraformDataQualityMonitorVariables([]datadogV1.MonitorFormulaAndFunctionQueryDefinition{*definition})

	queries := terraformVariables[0]["data_quality_query"].([]map[string]interface{})
	options := queries[0]["monitor_options"].([]map[string]interface{})[0]

	sourceToTarget := options["source_to_target_config"].([]map[string]interface{})[0]
	assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDIFFTYPE_ABSOLUTE, *sourceToTarget["diff_type"].(*datadogV1.MonitorFormulaAndFunctionDataQualityDiffType))
	source := sourceToTarget["source"].([]map[string]interface{})[0]
	assert.Equal(t, "src", *source["entity_id"].(*string))

	modelConfiguration := options["model_configuration"].([]map[string]interface{})[0]
	assert.Equal(t, int32(7), *modelConfiguration["auto_resolve_days"].(*int32))
	assert.False(t, *modelConfiguration["enable_flatline_detection"].(*bool))
}
