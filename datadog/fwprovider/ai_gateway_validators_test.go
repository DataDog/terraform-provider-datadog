package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestAiGatewayPositiveFloat64Validator(t *testing.T) {
	for name, tt := range map[string]struct {
		value   types.Float64
		invalid bool
	}{
		"positive": {value: types.Float64Value(0.5)},
		"zero":     {value: types.Float64Value(0), invalid: true},
		"negative": {value: types.Float64Value(-1), invalid: true},
		"null":     {value: types.Float64Null()},
		"unknown":  {value: types.Float64Unknown()},
	} {
		t.Run(name, func(t *testing.T) {
			resp := &validator.Float64Response{}
			aiGatewayPositiveFloat64Validator{}.ValidateFloat64(context.Background(), validator.Float64Request{
				Path:        path.Root("value"),
				ConfigValue: tt.value,
			}, resp)
			if resp.Diagnostics.HasError() != tt.invalid {
				t.Errorf("error = %v, want invalid = %t", resp.Diagnostics, tt.invalid)
			}
		})
	}
}
