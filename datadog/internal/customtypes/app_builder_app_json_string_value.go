// reference: https://github.com/hashicorp/terraform-plugin-framework-jsontypes/blob/4f0b31dbb4d9aba345ce0616029240eaa8e52f6f/jsontypes/normalized_value.go

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
var _ basetypes.StringValuable = (*AppBuilderAppJSONStringValue)(nil)
var _ basetypes.StringValuableWithSemanticEquals = (*AppBuilderAppJSONStringValue)(nil)

// AppBuilderAppJSONStringValue is an attribute type that represents a JSON string (RFC 7159). Semantic equality logic is defined for AppBuilderAppJSONStringValue
// such that inconsequential differences between JSON strings are ignored (whitespace, property order, etc), similar to jsontypes.Normalized,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
type AppBuilderAppJSONStringValue struct {
	basetypes.StringValue
}

// Type returns an AppBuilderAppJSONStringType.
func (v AppBuilderAppJSONStringValue) Type(ctx context.Context) attr.Type {
	// AppBuilderAppJSONStringType defined in the schema type section
	return AppBuilderAppJSONStringType{}
}

// Equal returns true if the given value is equivalent.
func (v AppBuilderAppJSONStringValue) Equal(o attr.Value) bool {
	other, ok := o.(AppBuilderAppJSONStringValue)

	if !ok {
		return false
	}
	return v.StringValue.Equal(other.StringValue)
}

// StringSemanticEquals returns true if the given JSON string value is semantically equal to the current JSON string value. When compared,
// these JSON string values are "normalized" by marshalling them to empty Go structs. This prevents Terraform data consistency errors and
// resource drift due to inconsequential differences in the JSON strings (whitespace, property order, etc), similar to jsontypes.Normalized,
// but also ignores other differences such as the App's ID, which is ignored in the App Builder API.
func (v AppBuilderAppJSONStringValue) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(AppBuilderAppJSONStringValue)

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

	// var prev interface{}
	// var next interface{}
	// yaml.Unmarshal([]byte(v.StringValue.ValueString()), &prev)
	// yaml.Unmarshal([]byte(other.StringValue.ValueString()), &next)
	// return cmp.Equal(prev, next), diags

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
	s1, err := normalizeAppBuilderAppJSONString(s1)
	if err != nil {
		return false, err
	}

	s2, err = normalizeAppBuilderAppJSONString(s2)
	if err != nil {
		return false, err
	}

	return s1 == s2, nil
}

func normalizeAppBuilderAppJSONString(jsonStr string) (string, error) {
	dec := json.NewDecoder(strings.NewReader(jsonStr))

	// This ensures the JSON decoder will not parse JSON numbers into Go's float64 type; avoiding Go
	// normalizing the JSON number representation or imposing limits on numeric range. See the unit test cases
	// of StringSemanticEquals for examples.
	dec.UseNumber()

	var temp interface{}
	if err := dec.Decode(&temp); err != nil {
		return "", err
	}

	// remove the "id" field from the JSON string because we want to ignore the App ID when comparing JSON strings
	if obj, ok := temp.(map[string]interface{}); ok {
		delete(obj, "id")
	}

	jsonBytes, err := json.Marshal(&temp)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
