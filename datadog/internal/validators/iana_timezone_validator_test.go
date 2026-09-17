package validators

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestIANATimezoneValidator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		value     types.String
		wantError bool
	}{
		{name: "regional", value: types.StringValue("America/New_York")},
		{name: "utc", value: types.StringValue("UTC")},
		{name: "invalid", value: types.StringValue("America/Not_A_Place"), wantError: true},
		{name: "local is not IANA", value: types.StringValue("Local"), wantError: true},
		{name: "empty", value: types.StringValue(""), wantError: true},
		{name: "null", value: types.StringNull()},
		{name: "unknown", value: types.StringUnknown()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var response validator.StringResponse
			IANATimezoneValidator().ValidateString(context.Background(), validator.StringRequest{ConfigValue: test.value}, &response)
			if got := response.Diagnostics.HasError(); got != test.wantError {
				t.Fatalf("HasError() = %v, want %v; diagnostics: %v", got, test.wantError, response.Diagnostics)
			}
		})
	}
}
