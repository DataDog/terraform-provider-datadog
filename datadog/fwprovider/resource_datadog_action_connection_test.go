package fwprovider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestActionConnectionTagValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tag   string
		valid bool
	}{
		{name: "key value", tag: "env:prod", valid: true},
		{name: "permitted punctuation", tag: "team/name:action-platform_v2.0", valid: true},
		{name: "reserved default key", tag: "default:connection", valid: false},
		{name: "missing value", tag: "env:", valid: false},
		{name: "missing separator", tag: "env", valid: false},
		{name: "multiple separators", tag: "env:prod:us1", valid: false},
		{name: "space", tag: "team:action platform", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := validator.StringRequest{
				Path:        path.Root("tags"),
				ConfigValue: types.StringValue(test.tag),
			}
			response := &validator.StringResponse{}
			actionConnectionTagValidator{}.ValidateString(context.Background(), request, response)

			if valid := !response.Diagnostics.HasError(); valid != test.valid {
				t.Fatalf("valid = %t, want %t", valid, test.valid)
			}
		})
	}
}
