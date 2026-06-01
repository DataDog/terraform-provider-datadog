package datadog

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubMonitorResource is a minimal utils.Resource implementation used to drive
// buildMonitorStruct in unit tests. It returns values from `values` for both
// Get and GetOk, signalling "set" via GetOk when the key is present and the
// value is non-nil and non-zero. GetRawConfigAt is not used by
// buildMonitorStruct, so we just return a null value.
type stubMonitorResource struct {
	values map[string]interface{}
}

func (s *stubMonitorResource) Get(key string) interface{} {
	if v, ok := s.values[key]; ok {
		return v
	}
	// Mirror ResourceData.Get's zero-value behavior for keys we did not
	// explicitly set. buildMonitorStruct only calls plain Get (with
	// hard-coded type assertions) on a small set of bool / int / string
	// fields whose zero value is the right SDK-default behavior here.
	switch key {
	case "include_tags", "notify_no_data", "require_full_window", "force_delete":
		return false
	case "new_host_delay":
		return 0
	}
	return ""
}

func (s *stubMonitorResource) GetOk(key string) (interface{}, bool) {
	v, ok := s.values[key]
	if !ok {
		return nil, false
	}
	// Match SDK semantics: GetOk reports !ok for zero values.
	switch typed := v.(type) {
	case nil:
		return nil, false
	case string:
		return typed, typed != ""
	case []interface{}:
		return typed, len(typed) > 0
	case bool:
		return typed, typed
	case int:
		return typed, typed != 0
	}
	return v, true
}

func (s *stubMonitorResource) GetRawConfigAt(_ cty.Path) (cty.Value, diag.Diagnostics) {
	return cty.NullVal(cty.DynamicPseudoType), nil
}

// TestBuildMonitorStruct_NilVariableEntry verifies that buildMonitorStruct
// does not panic when the `variables` list contains a nil element. This can
// happen during PlanResourceChange when a `variables {}` block is produced
// by a Terraform dynamic block whose inner content resolves to nothing
// (see GitHub issue #3149). Before the fix, the unsafe type assertion
// `v.(map[string]interface{})` panicked with
// `interface conversion: interface {} is nil, not map[string]interface {}`.
func TestBuildMonitorStruct_NilVariableEntry(t *testing.T) {
	stub := &stubMonitorResource{
		values: map[string]interface{}{
			"name":    "test-monitor",
			"type":    "cost alert",
			"message": "test",
			"query":   `formula("query1").last("30d") > 6`,
			// The bug trigger: a list with a single nil entry, which
			// mirrors what the SDK produces during plan when an empty
			// `variables {}` block comes from a dynamic block with empty
			// nested content.
			"variables": []interface{}{nil},
		},
	}

	require.NotPanics(t, func() {
		m, u := buildMonitorStruct(stub)
		// Sanity checks: the monitor should still be built with the
		// other fields, and the variables list should be unset (the
		// nil entry must be skipped, not propagated).
		require.NotNil(t, m)
		require.NotNil(t, u)
		assert.Equal(t, "test-monitor", m.GetName())
		assert.False(t, m.Options.HasVariables(), "variables should not be set when the only entry is nil")
	})
}

// TestBuildMonitorStruct_NilCloudCostQueryEntry verifies that buildMonitorStruct
// does not panic when an individual `cloud_cost_query` entry within
// `variables` is nil. This guards against the same panic class as
// TestBuildMonitorStruct_NilVariableEntry but at the nested-list level.
func TestBuildMonitorStruct_NilCloudCostQueryEntry(t *testing.T) {
	stub := &stubMonitorResource{
		values: map[string]interface{}{
			"name":    "test-monitor",
			"type":    "cost alert",
			"message": "test",
			"query":   `formula("query1").last("30d") > 6`,
			"variables": []interface{}{
				map[string]interface{}{
					"cloud_cost_query": []interface{}{nil},
				},
			},
		},
	}

	require.NotPanics(t, func() {
		m, _ := buildMonitorStruct(stub)
		require.NotNil(t, m)
		assert.False(t, m.Options.HasVariables(), "variables should not be set when the only cloud_cost_query entry is nil")
	})
}

// TestBuildMonitorStruct_NilEventQueryEntry verifies the same nil-guard for
// `event_query` entries.
func TestBuildMonitorStruct_NilEventQueryEntry(t *testing.T) {
	stub := &stubMonitorResource{
		values: map[string]interface{}{
			"name":    "test-monitor",
			"type":    "rum alert",
			"message": "test",
			"query":   `formula("var1").last("5m") > 100`,
			"variables": []interface{}{
				map[string]interface{}{
					"event_query": []interface{}{nil},
				},
			},
		},
	}

	require.NotPanics(t, func() {
		m, _ := buildMonitorStruct(stub)
		require.NotNil(t, m)
		assert.False(t, m.Options.HasVariables(), "variables should not be set when the only event_query entry is nil")
	})
}
