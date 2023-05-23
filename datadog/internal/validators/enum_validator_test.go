package validators

import (
	"context"
	"testing"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestEnumValidatorString(t *testing.T) {
	t.Parallel()

	type testCaseString struct {
		val         types.String
		enumFunc    interface{}
		expectError bool
	}
	stringTests := map[string]testCaseString{
		"valid string enum": {
			val:         types.StringValue("admins"),
			enumFunc:    datadogV2.NewTeamPermissionSettingValueFromValue,
			expectError: false,
		},
		"invalid string enum": {
			val:         types.StringValue("non-existent"),
			enumFunc:    datadogV2.NewTeamPermissionSettingValueFromValue,
			expectError: true,
		},
		"unknown string enum": {
			val:         types.StringUnknown(),
			enumFunc:    datadogV2.NewTeamPermissionSettingValueFromValue,
			expectError: false,
		},
		"null string enum": {
			val:         types.StringUnknown(),
			enumFunc:    datadogV2.NewTeamPermissionSettingValueFromValue,
			expectError: false,
		},
	}

	for name, test := range stringTests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			resp := validator.StringResponse{}
			NewEnumValidator[validator.String](test.enumFunc).ValidateString(context.TODO(), req, &resp)
			if !resp.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if resp.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}

func TestEnumValidatorInt64(t *testing.T) {
	t.Parallel()

	type testCaseString struct {
		val         types.Int64
		enumFunc    interface{}
		expectError bool
	}
	int64Tests := map[string]testCaseString{
		"valid Int64 enum": {
			val:         types.Int64Value(1),
			enumFunc:    datadogV1.NewSyntheticsPlayingTabFromValue,
			expectError: false,
		},
		"invalid Int64 enum": {
			val:         types.Int64Value(100),
			enumFunc:    datadogV1.NewSyntheticsPlayingTabFromValue,
			expectError: true,
		},
		"unknown Int64 enum": {
			val:         types.Int64Unknown(),
			enumFunc:    datadogV1.NewSyntheticsPlayingTabFromValue,
			expectError: false,
		},
		"null Int64 enum": {
			val:         types.Int64Unknown(),
			enumFunc:    datadogV1.NewSyntheticsPlayingTabFromValue,
			expectError: false,
		},
	}

	for name, test := range int64Tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			req := validator.Int64Request{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			resp := validator.Int64Response{}
			NewEnumValidator[validator.Int64](test.enumFunc).ValidateInt64(context.TODO(), req, &resp)
			if !resp.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if resp.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", resp.Diagnostics)
			}
		})
	}
}
