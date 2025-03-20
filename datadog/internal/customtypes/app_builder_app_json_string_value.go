// reference: https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/v0.2.0/jsontypes/normalized_value.go

package customtypes

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure the implementation satisfies the expected interfaces
var _ basetypes.StringValuable = (*AppBuilderAppStringValue)(nil)
var _ basetypes.StringValuableWithSemanticEquals = (*AppBuilderAppStringValue)(nil)

// AppBuilderAppStringValue is an attribute type that represents a JSON string (RFC 7159). Semantic equality logic is defined for AppBuilderAppStringValue
// such that inconsequential differences between JSON strings are ignored (whitespace, property order, etc), similar to jsontypes.Normalized,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
type AppBuilderAppStringValue struct {
	basetypes.StringValue
}

// Type returns an AppBuilderAppStringType.
func (v AppBuilderAppStringValue) Type(ctx context.Context) attr.Type {
	// AppBuilderAppStringType defined in the schema type section
	return AppBuilderAppStringType{}
}

// Equal returns true if the given value is equivalent.
func (v AppBuilderAppStringValue) Equal(o attr.Value) bool {
	other, ok := o.(AppBuilderAppStringValue)

	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true if the given JSON string value is semantically equal to the current JSON string value. When compared,
// these JSON string values are "normalized" by marshalling them to empty Go structs. This prevents Terraform data consistency errors and
// resource drift due to inconsequential differences in the JSON strings (whitespace, property order, etc), similar to jsontypes.Normalized,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
func (v AppBuilderAppStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(AppBuilderAppStringValue)

	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}

	result, err := appJSONEqual(newValue.ValueString(), v.ValueString())

	if err != nil {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected error occurred while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)

		return false, diags
	}

	return result, diags
}

func appJSONEqual(s1, s2 string) (bool, error) {
	s1, err := normalizeAppBuilderAppString(s1)
	if err != nil {
		return false, err
	}

	s2, err = normalizeAppBuilderAppString(s2)
	if err != nil {
		return false, err
	}

	return s1 == s2, nil
}

func normalizeAppBuilderAppString(jsonStr string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonStr))

	// This ensures the JSON decoder will not parse JSON numbers into Go's float64 type; avoiding Go
	// normalizing the JSON number representation or imposing limits on numeric range. See the unit test cases
	// of StringSemanticEquals for examples.
	dec.UseNumber()

	var temp interface{}
	if err := dec.Decode(&temp); err != nil {
		return "", err
	}

	// feature specific to AppBuilderAppStringValue:
	// we only want to compare fields that matter to Create/Update requests when comparing JSON strings
	if jsonMap, ok := temp.(map[string]interface{}); ok {
		fieldsToKeep := []string{"components", "description", "name", "queries", "rootInstanceName"}

		newJsonMap := make(map[string]interface{})
		for _, field := range fieldsToKeep {
			if val, ok := jsonMap[field]; ok {
				newJsonMap[field] = val
			}
		}

		temp = newJsonMap
	}

	jsonBytes, err := json.Marshal(&temp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Unmarshal calls (encoding/json).Unmarshal with the AppBuilderAppStringValue and `target` input. A null or unknown value will produce an error diagnostic.
// See encoding/json docs for more on usage: https://pkg.go.dev/encoding/json#Unmarshal
func (v AppBuilderAppStringValue) Unmarshal(target any) diag.Diagnostics {
	var diags diag.Diagnostics

	if v.IsNull() {
		diags.Append(diag.NewErrorDiagnostic("AppBuilderAppStringValue Unmarshal Error", "json string value is null"))
		return diags
	}

	if v.IsUnknown() {
		diags.Append(diag.NewErrorDiagnostic("AppBuilderAppStringValue Unmarshal Error", "json string value is unknown"))
		return diags
	}

	err := json.Unmarshal([]byte(v.ValueString()), target)
	if err != nil {
		diags.Append(diag.NewErrorDiagnostic("AppBuilderAppStringValue Unmarshal Error", err.Error()))
	}

	return diags
}

// NewAppBuilderAppStringValue creates a AppBuilderAppStringValue with a known value. Access the value via ValueString method.
func NewAppBuilderAppStringValue(value string) AppBuilderAppStringValue {
	return AppBuilderAppStringValue{
		StringValue: basetypes.NewStringValue(value),
	}
}

// NewAppBuilderAppStringValueNull creates a AppBuilderAppStringValue with a null value. Determine whether the value is null via IsNull method.
func NewAppBuilderAppStringValueNull() AppBuilderAppStringValue {
	return AppBuilderAppStringValue{
		StringValue: basetypes.NewStringNull(),
	}
}

// NewAppBuilderAppStringValueUnknown creates a AppBuilderAppStringValue with an unknown value. Determine whether the value is unknown via IsUnknown method.
func NewAppBuilderAppStringValueUnknown() AppBuilderAppStringValue {
	return AppBuilderAppStringValue{
		StringValue: basetypes.NewStringUnknown(),
	}
}
