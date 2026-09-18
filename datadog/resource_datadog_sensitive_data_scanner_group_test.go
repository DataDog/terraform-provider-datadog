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

func TestReconcileTerraformSamplings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		apiSamplings       []interface{}
		referenceSamplings []interface{}
		want               []interface{}
	}{
		{
			name: "reorder when product sets match",
			apiSamplings: []interface{}{
				map[string]interface{}{"product": "apm", "rate": 25.0},
				map[string]interface{}{"product": "rum", "rate": 25.0},
				map[string]interface{}{"product": "events", "rate": 25.0},
			},
			referenceSamplings: []interface{}{
				map[string]interface{}{"product": "rum", "rate": 25.0},
				map[string]interface{}{"product": "events", "rate": 25.0},
				map[string]interface{}{"product": "apm", "rate": 25.0},
			},
			want: []interface{}{
				map[string]interface{}{"product": "rum", "rate": 25.0},
				map[string]interface{}{"product": "events", "rate": 25.0},
				map[string]interface{}{"product": "apm", "rate": 25.0},
			},
		},
		{
			name: "drop implicit 100% for unconfigured product",
			apiSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
				map[string]interface{}{"product": "apm", "rate": 100.0},
			},
			referenceSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
			},
			want: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
			},
		},
		{
			name: "keep explicitly configured product at 100%",
			apiSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
				map[string]interface{}{"product": "apm", "rate": 100.0},
			},
			referenceSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
			},
			want: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
			},
		},
		{
			name:         "surface configured product omitted by API as implicit 100%",
			apiSamplings: []interface{}{},
			referenceSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
			},
			want: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
			},
		},
		{
			name: "keep API rate for configured products and surface omitted ones as 100%",
			apiSamplings: []interface{}{
				map[string]interface{}{"product": "apm", "rate": 50.0},
			},
			referenceSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
				map[string]interface{}{"product": "apm", "rate": 50.0},
			},
			want: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
				map[string]interface{}{"product": "apm", "rate": 50.0},
			},
		},
		{
			name: "surface non-100% drift for unconfigured product",
			apiSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
				map[string]interface{}{"product": "apm", "rate": 10.0},
			},
			referenceSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
			},
			want: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 50.0},
				map[string]interface{}{"product": "apm", "rate": 10.0},
			},
		},
		{
			name: "empty reference drops all implicit 100% entries",
			apiSamplings: []interface{}{
				map[string]interface{}{"product": "logs", "rate": 100.0},
				map[string]interface{}{"product": "apm", "rate": 100.0},
			},
			referenceSamplings: []interface{}{},
			want:               []interface{}{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := reconcileTerraformSamplings(tt.apiSamplings, tt.referenceSamplings)
			if len(got) != len(tt.want) {
				t.Fatalf("reconcileTerraformSamplings() length = %d, want %d", len(got), len(tt.want))
			}
			for i, sampling := range got {
				gotMap := sampling.(map[string]interface{})
				wantMap := tt.want[i].(map[string]interface{})
				if gotMap["product"] != wantMap["product"] {
					t.Fatalf("reconcileTerraformSamplings()[%d].product = %q, want %q", i, gotMap["product"], wantMap["product"])
				}
				if gotMap["rate"] != wantMap["rate"] {
					t.Fatalf("reconcileTerraformSamplings()[%d].rate = %v, want %v", i, gotMap["rate"], wantMap["rate"])
				}
			}
		})
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
