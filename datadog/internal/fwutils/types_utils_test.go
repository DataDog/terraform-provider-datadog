package fwutils

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestToTerraformFloat64(t *testing.T) {
	zero := 0.0
	value := 2.5

	cases := map[string]struct {
		v        *float64
		ok       bool
		expected types.Float64
	}{
		"set":              {&value, true, types.Float64Value(2.5)},
		"set to zero":      {&zero, true, types.Float64Value(0)},
		"not ok":           {&value, false, types.Float64Null()},
		"nil pointer":      {nil, true, types.Float64Null()},
		"nil pointer, !ok": {nil, false, types.Float64Null()},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := ToTerraformFloat64(tc.v, tc.ok); !got.Equal(tc.expected) {
				t.Errorf("ToTerraformFloat64() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestSetOptFloat64(t *testing.T) {
	cases := map[string]struct {
		f          types.Float64
		wantCalled bool
		wantValue  float64
	}{
		"set":         {types.Float64Value(2.5), true, 2.5},
		"set to zero": {types.Float64Value(0), true, 0},
		"null":        {types.Float64Null(), false, 0},
		"unknown":     {types.Float64Unknown(), false, 0},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			called := false
			var got float64
			SetOptFloat64(tc.f, func(v float64) {
				called = true
				got = v
			})
			if called != tc.wantCalled {
				t.Fatalf("setter called = %v, want %v", called, tc.wantCalled)
			}
			if called && got != tc.wantValue {
				t.Errorf("setter received %v, want %v", got, tc.wantValue)
			}
		})
	}
}
