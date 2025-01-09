package fwutils

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// RemoveDefaultIfConflictingSet remove default value from attr if conflicting paths are set
// explicitly sets the default value to null if any of the conflicting paths are set.
// This is useful for nested blocks which have oneOf attributes.
func RemoveDefaultIfConflictingSet(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse, defaultPath path.Expression, conflictingPaths ...path.Expression) {
	isConflictingSet := false
	for _, conflictingPath := range conflictingPaths {
		conflictingPathMatches, _ := request.Config.PathMatches(ctx, conflictingPath)
		var conflictingValue attr.Value
		request.Config.GetAttribute(ctx, conflictingPathMatches[0], &conflictingValue)
		schema, _ := request.Config.Schema.AttributeAtPath(ctx, conflictingPathMatches[0])
		if (schema.IsComputed() && !conflictingValue.IsUnknown()) || (!schema.IsComputed() && !conflictingValue.IsNull()) {
			isConflictingSet = true
			break
		}
	}

	if isConflictingSet {
		defaultPathMatches, _ := request.Config.PathMatches(ctx, defaultPath)
		schema, _ := request.Config.Schema.AttributeAtPath(ctx, defaultPathMatches[0])
		var defaultValue interface{}
		switch schema.GetType() {
		case types.StringType:
			defaultValue = types.StringNull()
		case types.BoolType:
			defaultValue = types.BoolNull()
		case types.NumberType:
			defaultValue = types.NumberNull()
		case types.Float64Type:
			defaultValue = types.Float64Null()
		case types.Int64Type:
			defaultValue = types.Int64Null()
		case types.ListType{ElemType: types.StringType}:
			defaultValue = types.ListNull(types.StringType)
		case types.ListType{ElemType: types.BoolType}:
			defaultValue = types.ListNull(types.BoolType)
		case types.ListType{ElemType: types.NumberType}:
			defaultValue = types.ListNull(types.NumberType)
		case types.ListType{ElemType: types.Float64Type}:
			defaultValue = types.ListNull(types.Float64Type)
		case types.ListType{ElemType: types.Int64Type}:
			defaultValue = types.ListNull(types.Int64Type)
		default:
			response.Diagnostics.AddError("unsupported type for default value", fmt.Sprintf("Unsupported type: %s", defaultPathMatches[0].String()))
			return
		}

		response.Plan.SetAttribute(ctx, defaultPathMatches[0], defaultValue)
	}

}
