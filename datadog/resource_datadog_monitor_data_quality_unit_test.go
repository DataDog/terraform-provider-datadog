package datadog

import (
	"encoding/json"
	"testing"
)

// TestBuildDataQualityMonitorOptions covers the mapping from a data_quality_query
// monitor_options block to the API payload, including the "zero means unset" handling
// for sensitivity that SDKv2 forces on us (an omitted float attribute and one explicitly
// set to 0 are indistinguishable in the resource data map).
func TestBuildDataQualityMonitorOptions(t *testing.T) {
	cases := map[string]struct {
		data     map[string]interface{}
		wantNil  bool
		wantJSON string
	}{
		"empty block": {
			data:    map[string]interface{}{},
			wantNil: true,
		},
		"sensitivity only": {
			data:     map[string]interface{}{"sensitivity": 2.5},
			wantJSON: `{"sensitivity":2.5}`,
		},
		"zero sensitivity is treated as unset": {
			data:    map[string]interface{}{"sensitivity": 0.0},
			wantNil: true,
		},
		"sensitivity alongside other options": {
			data: map[string]interface{}{
				"model_type_override": "freshness",
				"crontab_override":    "0 0 * * *",
				"sensitivity":         1.5,
			},
			wantJSON: `{"crontab_override":"0 0 * * *","model_type_override":"freshness","sensitivity":1.5}`,
		},
		"zero sensitivity does not suppress other options": {
			data: map[string]interface{}{
				"model_type_override": "freshness",
				"sensitivity":         0.0,
			},
			wantJSON: `{"model_type_override":"freshness"}`,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			opts := buildDataQualityMonitorOptions(tc.data)
			if tc.wantNil {
				if opts != nil {
					t.Fatalf("expected nil options, got %+v", opts)
				}
				return
			}
			if opts == nil {
				t.Fatal("expected options, got nil")
			}
			got, err := json.Marshal(opts)
			if err != nil {
				t.Fatalf("marshaling options: %v", err)
			}
			if string(got) != tc.wantJSON {
				t.Errorf("payload = %s, want %s", got, tc.wantJSON)
			}
		})
	}
}
