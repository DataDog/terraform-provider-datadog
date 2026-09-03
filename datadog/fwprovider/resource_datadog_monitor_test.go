package fwprovider

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func stringList(ctx context.Context, values ...string) types.List {
	list, _ := types.ListValueFrom(ctx, types.StringType, values)
	return list
}

func dataQualityQueryWithOptions(opts DataQualityMonitorOptions) []DataQualityQuery {
	return []DataQualityQuery{{
		Name:           types.StringValue("query1"),
		DataSource:     types.StringValue("data_quality_metrics"),
		Measure:        types.StringValue("row_count"),
		Filter:         types.StringValue("search for column where `database:production`"),
		GroupBy:        types.ListNull(types.StringType),
		MonitorOptions: []DataQualityMonitorOptions{opts},
	}}
}

// buildOptions runs the terraform model through the request builder and returns the
// monitor_options the API would receive.
func buildOptions(t *testing.T, opts DataQualityMonitorOptions) *datadogV1.MonitorFormulaAndFunctionDataQualityMonitorOptions {
	t.Helper()
	r := &monitorResource{}
	defs := r.buildDataQualityQueryStruct(context.Background(), dataQualityQueryWithOptions(opts))
	if !assert.Len(t, defs, 1) {
		return nil
	}
	def := defs[0].MonitorFormulaAndFunctionDataQualityQueryDefinition
	if !assert.NotNil(t, def) {
		return nil
	}
	return def.MonitorOptions
}

func TestBuildDataQualityQueryStruct_Sensitivity(t *testing.T) {
	opts := buildOptions(t, DataQualityMonitorOptions{Sensitivity: types.Float64Value(2.5)})
	assert.Equal(t, 2.5, opts.GetSensitivity())

	// A null sensitivity is omitted rather than sent as zero, which the API would
	// read as a zero-width bound.
	opts = buildOptions(t, DataQualityMonitorOptions{Sensitivity: types.Float64Null()})
	assert.False(t, opts.HasSensitivity())
}

func TestBuildDataQualityQueryStruct_SourceToTargetConfig(t *testing.T) {
	ctx := context.Background()

	opts := buildOptions(t, DataQualityMonitorOptions{
		SourceToTargetConfig: []DataQualitySourceToTargetConfig{{
			DiffType:   types.StringValue("diff_percent"),
			EntityType: types.StringValue("table"),
			Source: []DataQualityEntityMetricConfig{{
				EntityId:       types.StringValue("src-entity"),
				EntityType:     types.StringValue("table"),
				CustomWhere:    types.StringValue("env = 'prod'"),
				GroupByColumns: stringList(ctx, "env"),
			}},
			Target: []DataQualityEntityMetricConfig{{
				EntityId:       types.StringValue("tgt-entity"),
				EntityType:     types.StringValue("table"),
				GroupByColumns: types.ListNull(types.StringType),
			}},
		}},
	})

	cfg := opts.GetSourceToTargetConfig()
	assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDIFFTYPE_DIFF_PERCENT, cfg.GetDiffType())
	assert.Equal(t, "table", cfg.GetEntityType())

	source := cfg.GetSource()
	assert.Equal(t, "src-entity", source.GetEntityId())
	assert.Equal(t, "env = 'prod'", source.GetCustomWhere())
	assert.Equal(t, []string{"env"}, source.GetGroupByColumns())
	assert.False(t, source.HasCustomSql())

	target := cfg.GetTarget()
	assert.Equal(t, "tgt-entity", target.GetEntityId())
	assert.False(t, target.HasCustomWhere())
	assert.False(t, target.HasGroupByColumns())
}

func TestBuildDataQualityQueryStruct_ModelConfiguration(t *testing.T) {
	opts := buildOptions(t, DataQualityMonitorOptions{
		ModelConfiguration: []DataQualityModelConfiguration{{
			AutoResolveDays:         types.Int32Value(7),
			EnableFlatlineDetection: types.BoolValue(true),
			Function:                types.StringValue("DIFF"),
			MinLowerBoundSize:       types.Float64Value(10),
			MinUpperBoundSize:       types.Float64Value(20),
			ModelBoundsOverride:     types.StringValue("UPPER_ONLY"),
		}},
	})

	cfg := opts.GetModelConfiguration()
	assert.Equal(t, int32(7), cfg.GetAutoResolveDays())
	assert.True(t, cfg.GetEnableFlatlineDetection())
	assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDIFFFUNCTION_DIFF, cfg.GetFunction())
	assert.Equal(t, 10.0, cfg.GetMinLowerBoundSize())
	assert.Equal(t, 20.0, cfg.GetMinUpperBoundSize())
	assert.Equal(t, datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYMODELBOUNDSOVERRIDE_UPPER_ONLY, cfg.GetModelBoundsOverride())
}

// The framework can represent null, so unlike the SDKv2 path it does not need to
// default enable_flatline_detection to true: an omitted attribute is simply not
// sent, and an explicit false is preserved rather than read as "unset".
func TestBuildDataQualityQueryStruct_FlatlineDetectionTriState(t *testing.T) {
	tests := []struct {
		name    string
		value   types.Bool
		present bool
		want    bool
	}{
		{name: "omitted is not sent", value: types.BoolNull(), present: false},
		{name: "explicit false is preserved", value: types.BoolValue(false), present: true, want: false},
		{name: "explicit true is sent", value: types.BoolValue(true), present: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := buildOptions(t, DataQualityMonitorOptions{
				ModelConfiguration: []DataQualityModelConfiguration{{
					EnableFlatlineDetection: tt.value,
					AutoResolveDays:         types.Int32Null(),
					MinLowerBoundSize:       types.Float64Null(),
					MinUpperBoundSize:       types.Float64Null(),
				}},
			})
			cfg := opts.GetModelConfiguration()
			assert.Equal(t, tt.present, cfg.HasEnableFlatlineDetection())
			if tt.present {
				assert.Equal(t, tt.want, cfg.GetEnableFlatlineDetection())
			}
		})
	}
}

func TestBuildDataQualityQueryState_MonitorOptions(t *testing.T) {
	ctx := context.Background()
	r := &monitorResource{}

	source := datadogV1.NewMonitorFormulaAndFunctionDataQualityEntityMetricConfig("src-entity", "table")
	source.SetGroupByColumns([]string{"env"})
	target := datadogV1.NewMonitorFormulaAndFunctionDataQualityEntityMetricConfig("tgt-entity", "table")

	modelCfg := datadogV1.NewMonitorFormulaAndFunctionDataQualityModelConfiguration()
	modelCfg.SetAutoResolveDays(7)
	modelCfg.SetEnableFlatlineDetection(false)
	modelCfg.SetModelBoundsOverride(datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYMODELBOUNDSOVERRIDE_LOWER_ONLY)

	apiOpts := datadogV1.NewMonitorFormulaAndFunctionDataQualityMonitorOptions()
	apiOpts.SetSensitivity(2.5)
	apiOpts.SetSourceToTargetConfig(*datadogV1.NewMonitorFormulaAndFunctionDataQualitySourceToTargetConfig(
		datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDIFFTYPE_ABSOLUTE, "table", *source, *target))
	apiOpts.SetModelConfiguration(*modelCfg)

	def := datadogV1.NewMonitorFormulaAndFunctionDataQualityQueryDefinition(
		datadogV1.MONITORFORMULAANDFUNCTIONDATAQUALITYDATASOURCE_DATA_QUALITY_METRICS,
		"search for column where `database:production`", "row_count", "query1")
	def.SetMonitorOptions(*apiOpts)

	state := r.buildDataQualityQueryState(ctx, def)

	if !assert.Len(t, state.MonitorOptions, 1) {
		return
	}
	opts := state.MonitorOptions[0]
	assert.Equal(t, types.Float64Value(2.5), opts.Sensitivity)

	if assert.Len(t, opts.SourceToTargetConfig, 1) {
		cfg := opts.SourceToTargetConfig[0]
		assert.Equal(t, types.StringValue("absolute"), cfg.DiffType)
		assert.Equal(t, types.StringValue("table"), cfg.EntityType)
		if assert.Len(t, cfg.Source, 1) {
			assert.Equal(t, types.StringValue("src-entity"), cfg.Source[0].EntityId)
			assert.Equal(t, stringList(ctx, "env"), cfg.Source[0].GroupByColumns)
		}
		if assert.Len(t, cfg.Target, 1) {
			assert.Equal(t, types.StringValue("tgt-entity"), cfg.Target[0].EntityId)
			// Absent optionals come back null rather than as empty strings.
			assert.True(t, cfg.Target[0].CustomWhere.IsNull())
		}
	}

	if assert.Len(t, opts.ModelConfiguration, 1) {
		cfg := opts.ModelConfiguration[0]
		assert.Equal(t, types.Int32Value(7), cfg.AutoResolveDays)
		assert.Equal(t, types.BoolValue(false), cfg.EnableFlatlineDetection)
		assert.Equal(t, types.StringValue("LOWER_ONLY"), cfg.ModelBoundsOverride)
		assert.True(t, cfg.Function.IsNull())
	}
}

// A value set in configuration must survive the build/read cycle unchanged, or
// terraform reports a permanent diff.
func TestDataQualityMonitorOptionsRoundTrip(t *testing.T) {
	ctx := context.Background()
	r := &monitorResource{}

	original := DataQualityMonitorOptions{
		Sensitivity: types.Float64Value(2.5),
		SourceToTargetConfig: []DataQualitySourceToTargetConfig{{
			DiffType:   types.StringValue("absolute"),
			EntityType: types.StringValue("table"),
			Source: []DataQualityEntityMetricConfig{{
				EntityId:       types.StringValue("src"),
				EntityType:     types.StringValue("table"),
				GroupByColumns: types.ListNull(types.StringType),
			}},
			Target: []DataQualityEntityMetricConfig{{
				EntityId:       types.StringValue("tgt"),
				EntityType:     types.StringValue("table"),
				GroupByColumns: types.ListNull(types.StringType),
			}},
		}},
		ModelConfiguration: []DataQualityModelConfiguration{{
			AutoResolveDays:         types.Int32Value(7),
			EnableFlatlineDetection: types.BoolValue(false),
			MinLowerBoundSize:       types.Float64Null(),
			MinUpperBoundSize:       types.Float64Null(),
		}},
	}

	defs := r.buildDataQualityQueryStruct(ctx, dataQualityQueryWithOptions(original))
	state := r.buildDataQualityQueryState(ctx, defs[0].MonitorFormulaAndFunctionDataQualityQueryDefinition)

	if !assert.Len(t, state.MonitorOptions, 1) {
		return
	}
	got := state.MonitorOptions[0]
	assert.Equal(t, original.Sensitivity, got.Sensitivity)
	assert.Equal(t, original.SourceToTargetConfig, got.SourceToTargetConfig)
	assert.Equal(t, original.ModelConfiguration, got.ModelConfiguration)
}
