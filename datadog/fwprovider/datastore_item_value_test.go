package fwprovider

import (
	"testing"
)

func TestDatastoreItemValueToMap(t *testing.T) {
	got := datastoreItemValueToMap(map[string]interface{}{
		"product": "apm",
		"object":  map[string]interface{}{"threshold": float64(10)},
		"list":    []interface{}{"a", "b"},
		"number":  float64(10),
		"bool":    true,
	})

	want := map[string]string{
		"product": "apm",
		"object":  `{"threshold":10}`,
		"list":    `["a","b"]`,
		"number":  "10",
		"bool":    "true",
	}
	for column, expected := range want {
		if got[column] != expected {
			t.Errorf("column %q: got %q, want %q", column, got[column], expected)
		}
	}
}
