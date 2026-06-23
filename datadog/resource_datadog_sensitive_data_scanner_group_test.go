package datadog

import (
	"testing"
)

func TestSamplingsEqualOrderInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		old  []interface{}
		new  []interface{}
		want bool
	}{
		{
			name: "same order",
			old: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 25.0},
				map[string]interface{}{"product": "rum", "rate": 25.0},
			},
			new: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 25.0},
				map[string]interface{}{"product": "rum", "rate": 25.0},
			},
			want: true,
		},
		{
			name: "different order",
			old: []interface{}{
				map[string]interface{}{"product": "apm", "rate": 25.0},
				map[string]interface{}{"product": "rum", "rate": 25.0},
				map[string]interface{}{"product": "events", "rate": 25.0},
			},
			new: []interface{}{
				map[string]interface{}{"product": "rum", "rate": 25.0},
				map[string]interface{}{"product": "events", "rate": 25.0},
				map[string]interface{}{"product": "apm", "rate": 25.0},
			},
			want: true,
		},
		{
			name: "different rate",
			old: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 25.0},
			},
			new: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
			},
			want: false,
		},
		{
			name: "different product",
			old: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 25.0},
			},
			new: []interface{}{
				map[string]interface{}{"product": "rum", "rate": 25.0},
			},
			want: false,
		},
		{
			name: "different length",
			old: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 25.0},
			},
			new: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 25.0},
				map[string]interface{}{"product": "rum", "rate": 25.0},
			},
			want: false,
		},
		{
			name: "both empty",
			old:  []interface{}{},
			new:  []interface{}{},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := samplingsEqualOrderInsensitive(tt.old, tt.new); got != tt.want {
				t.Fatalf("samplingsEqualOrderInsensitive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAlignTerraformSamplingsOrder(t *testing.T) {
	t.Parallel()

	apiSamplings := []interface{}{
		map[string]interface{}{"product": "apm", "rate": 25.0},
		map[string]interface{}{"product": "rum", "rate": 25.0},
		map[string]interface{}{"product": "events", "rate": 25.0},
	}
	preferredOrder := []interface{}{
		map[string]interface{}{"product": "rum", "rate": 25.0},
		map[string]interface{}{"product": "events", "rate": 25.0},
		map[string]interface{}{"product": "apm", "rate": 25.0},
	}

	got := alignTerraformSamplingsOrder(apiSamplings, preferredOrder)
	if len(got) != len(preferredOrder) {
		t.Fatalf("alignTerraformSamplingsOrder() length = %d, want %d", len(got), len(preferredOrder))
	}
	for i, sampling := range got {
		samplingMap := sampling.(map[string]interface{})
		preferredMap := preferredOrder[i].(map[string]interface{})
		if samplingMap["product"] != preferredMap["product"] {
			t.Fatalf("alignTerraformSamplingsOrder()[%d].product = %q, want %q", i, samplingMap["product"], preferredMap["product"])
		}
	}
}

func TestValidateSamplingsProductUniqueness(t *testing.T) {
	t.Parallel()

	err := validateSamplingsProductUniqueness([]interface{}{
		map[string]interface{}{"product": "logs", "rate": 25.0},
		map[string]interface{}{"product": "logs", "rate": 50.0},
	})
	if err == nil {
		t.Fatal("expected duplicate product validation error")
	}
}
