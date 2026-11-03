package datadog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// metricQueryRaw mirrors the schema-decoded shape of a single time-slice /
// count `query` block containing one `metric_query`.
func metricQueryRaw(dataSource string) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"formula": []interface{}{
				map[string]interface{}{"formula_expression": "query1"},
			},
			"query": []interface{}{
				map[string]interface{}{
					"metric_query": []interface{}{
						map[string]interface{}{
							"data_source": dataSource,
							"name":        "query1",
							"query":       "avg:system.cpu.user{*}",
						},
					},
				},
			},
		},
	}
}

// TestBuildSLOTimeSliceQueryStruct_InvalidDataSource guards against a
// regression where an unsupported data_source (e.g. "rum") caused a nil
// pointer dereference: NewFormulaAndFunctionMetricDataSourceFromValue returns
// (nil, err) for anything other than "metrics", the error was discarded, and
// dereferencing the nil pointer crashed the whole provider process at apply.
func TestBuildSLOTimeSliceQueryStruct_InvalidDataSource(t *testing.T) {
	require.NotPanics(t, func() {
		_, err := buildSLOTimeSliceQueryStruct(metricQueryRaw("rum"))
		assert.Error(t, err, "invalid data_source must return an error, not panic")
	})
}

func TestBuildSLOTimeSliceQueryStruct_ValidDataSource(t *testing.T) {
	q, err := buildSLOTimeSliceQueryStruct(metricQueryRaw("metrics"))
	require.NoError(t, err)
	require.NotNil(t, q)
	require.Len(t, q.GetQueries(), 1)
}

// TestBuildSLOCountSpec_InvalidDataSource covers the same nil-deref pattern in
// the count-SLI code path.
func TestBuildSLOCountSpec_InvalidDataSource(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{
			"good_events_formula":  "query1",
			"total_events_formula": "query1",
			"queries": []interface{}{
				map[string]interface{}{
					"metric_query": []interface{}{
						map[string]interface{}{
							"data_source": "rum",
							"name":        "query1",
							"query":       "avg:system.cpu.user{*}",
						},
					},
				},
			},
		},
	}
	require.NotPanics(t, func() {
		_, err := buildSLOCountSpec(raw)
		assert.Error(t, err, "invalid data_source must return an error, not panic")
	})
}
