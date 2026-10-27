package datadog

import (
	"reflect"
	"testing"
)

func TestParseStringArray(t *testing.T) {
	testCases := []struct {
		name     string
		input    []interface{}
		expected []string
	}{
		{
			name:     "normal_strings",
			input:    []interface{}{"@user", "@admin"},
			expected: []string{"@user", "@admin"},
		},
		{
			name:     "empty_array",
			input:    []interface{}{},
			expected: []string{},
		},
		{
			name:     "nil_values_filtered",
			input:    []interface{}{"@user", nil, "@admin"},
			expected: []string{"@user", "@admin"},
		},
		{
			name:     "all_nil_values",
			input:    []interface{}{nil, nil, nil},
			expected: []string{},
		},
		{
			name:     "empty_string_kept",
			input:    []interface{}{"@user", "", "@admin"},
			expected: []string{"@user", "", "@admin"},
		},
		{
			name:     "only_empty_string",
			input:    []interface{}{""},
			expected: []string{""},
		},
		{
			name:     "mixed_nil_and_empty",
			input:    []interface{}{nil, "", nil, "@user"},
			expected: []string{"", "@user"},
		},
		{
			name:     "single_valid_string",
			input:    []interface{}{"@slack-channel"},
			expected: []string{"@slack-channel"},
		},
		{
			name:     "single_nil",
			input:    []interface{}{nil},
			expected: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseStringArray(tc.input)
			if !reflect.DeepEqual(result, tc.expected) {
				t.Errorf("parseStringArray(%v) = %v, expected %v", tc.input, result, tc.expected)
			}
		})
	}
}

// TestParseStringArrayNoPanic ensures the function doesn't panic on nil values
// This is the specific fix for GitHub issue #3322
func TestParseStringArrayNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("parseStringArray panicked with input containing nil: %v", r)
		}
	}()

	// These inputs previously caused a panic before the fix
	panicInputs := [][]interface{}{
		{nil},
		{"", nil},
		{nil, nil, nil},
		{"@user", nil, "@admin"},
	}

	for _, input := range panicInputs {
		parseStringArray(input)
	}
}

