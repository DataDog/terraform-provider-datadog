package datadog

import (
	"encoding/json"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file tests dashboard widget time functionality independently of the cassette testing
// in this project. It does this because it tests inbound json formats of the widget time that
// can only be created externally by another API client (or via the UI).

// Test buildTerraformWidgetTime with UnparsedObject (new format in unparsed)
func TestBuildTerraformWidgetTime_UnparsedNewFormat(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		expectedSpan string
		expectedHide *bool
	}{
		{
			name:         "new format with hide true",
			jsonData:     `{"type":"live","unit":"hour","value":1,"hide_incomplete_cost_data":true}`,
			expectedSpan: "1h",
			expectedHide: boolPtr(true),
		},
		{
			name:         "new format with hide false",
			jsonData:     `{"type":"live","unit":"minute","value":5,"hide_incomplete_cost_data":false}`,
			expectedSpan: "5m",
			expectedHide: nil, // false is the default, not stored in state
		},
		{
			name:         "new format without hide",
			jsonData:     `{"type":"live","unit":"day","value":1}`,
			expectedSpan: "1d",
			expectedHide: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an unparsed WidgetTime by unmarshaling raw JSON into UnparsedObject
			var unparsedMap map[string]interface{}
			err := json.Unmarshal([]byte(tt.jsonData), &unparsedMap)
			require.NoError(t, err)

			widgetTime := &datadogV1.WidgetTime{
				UnparsedObject: unparsedMap,
			}

			terraformDef := make(map[string]interface{})
			buildTerraformWidgetTime(widgetTime, terraformDef)

			// WidgetLiveSpan is a type alias, convert to string for comparison
			if liveSpan, ok := terraformDef["live_span"].(datadogV1.WidgetLiveSpan); ok {
				assert.Equal(t, tt.expectedSpan, string(liveSpan))
			} else if liveSpanStr, ok := terraformDef["live_span"].(string); ok {
				assert.Equal(t, tt.expectedSpan, liveSpanStr)
			} else {
				t.Fatalf("live_span has unexpected type: %T", terraformDef["live_span"])
			}

			if tt.expectedHide != nil {
				assert.Equal(t, *tt.expectedHide, terraformDef["hide_incomplete_cost_data"])
			} else {
				// Should not be set if not present
				_, exists := terraformDef["hide_incomplete_cost_data"]
				assert.False(t, exists)
			}
		})
	}
}

// Test buildTerraformWidgetTime with UnparsedObject (legacy format in unparsed)
func TestBuildTerraformWidgetTime_UnparsedLegacyFormat(t *testing.T) {
	tests := []struct {
		name         string
		jsonData     string
		expectedSpan string
		expectedHide *bool
	}{
		{
			name:         "legacy format with hide true",
			jsonData:     `{"live_span":"1h","hide_incomplete_cost_data":true}`,
			expectedSpan: "1h",
			expectedHide: boolPtr(true),
		},
		{
			name:         "legacy format without hide",
			jsonData:     `{"live_span":"5m"}`,
			expectedSpan: "5m",
			expectedHide: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var unparsedMap map[string]interface{}
			err := json.Unmarshal([]byte(tt.jsonData), &unparsedMap)
			require.NoError(t, err)

			widgetTime := &datadogV1.WidgetTime{
				UnparsedObject: unparsedMap,
			}

			terraformDef := make(map[string]interface{})
			buildTerraformWidgetTime(widgetTime, terraformDef)

			// WidgetLiveSpan is a type alias, convert to string for comparison
			if liveSpan, ok := terraformDef["live_span"].(datadogV1.WidgetLiveSpan); ok {
				assert.Equal(t, tt.expectedSpan, string(liveSpan))
			} else if liveSpanStr, ok := terraformDef["live_span"].(string); ok {
				assert.Equal(t, tt.expectedSpan, liveSpanStr)
			} else {
				t.Fatalf("live_span has unexpected type: %T", terraformDef["live_span"])
			}

			if tt.expectedHide != nil {
				assert.Equal(t, *tt.expectedHide, terraformDef["hide_incomplete_cost_data"])
			} else {
				_, exists := terraformDef["hide_incomplete_cost_data"]
				assert.False(t, exists)
			}
		})
	}
}

// Test isWidgetTimeUnparsedObject
func TestIsWidgetTimeUnparsedObject(t *testing.T) {
	tests := []struct {
		name     string
		obj      interface{}
		expected bool
	}{
		{
			name: "WidgetNewLiveSpan signature",
			obj: map[string]interface{}{
				"type":  "live",
				"unit":  "hour",
				"value": 1,
			},
			expected: true,
		},
		{
			name: "WidgetNewLiveSpan with hide_incomplete_cost_data",
			obj: map[string]interface{}{
				"type":                      "live",
				"unit":                      "hour",
				"value":                     1,
				"hide_incomplete_cost_data": true,
			},
			expected: true,
		},
		{
			name: "WidgetNewFixedSpan signature",
			obj: map[string]interface{}{
				"type": "fixed",
				"from": 1609459200,
				"to":   1609545600,
			},
			expected: true,
		},
		{
			name: "WidgetLegacyLiveSpan signature",
			obj: map[string]interface{}{
				"live_span": "1h",
			},
			expected: true,
		},
		{
			name: "WidgetLegacyLiveSpan with hide_incomplete_cost_data",
			obj: map[string]interface{}{
				"live_span":                 "1h",
				"hide_incomplete_cost_data": true,
			},
			expected: true,
		},
		{
			name: "not a WidgetTime - missing required fields",
			obj: map[string]interface{}{
				"type": "live",
				// missing unit and value
			},
			expected: false,
		},
		{
			name: "not a WidgetTime - wrong structure",
			obj: map[string]interface{}{
				"something": "else",
			},
			expected: false,
		},
		{
			name: "not a WidgetTime - legacy with extra fields",
			obj: map[string]interface{}{
				"live_span": "1h",
				"extra":     "field",
			},
			expected: false,
		},
		{
			name:     "not a map",
			obj:      "string",
			expected: false,
		},
		{
			name:     "nil",
			obj:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isWidgetTimeUnparsedObject(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper function to create bool pointers
func boolPtr(b bool) *bool {
	return &b
}
