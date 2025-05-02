package customtypes

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable                    = (*BackwardRFC3339DateType)(nil)
	_ basetypes.StringValuable                   = (*BackwardRFC3339Date)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*BackwardRFC3339Date)(nil)
)

// BackwardRFC3339DateType is a type that represents a date in the RFC3339 format.
// that can only move backward in time.
// This is for example used for the on-call schedule effective_date. A user supplied
// effective date from a terraform config _might_ be clamped to a more recent date (~now) by the Datadog API.
// This is done to prevent creating on-call shifts for the past.
// In such case (when the supplied date is move _forward_ we decide to keep the original value and store the clamped value in a different attribute)
type BackwardRFC3339DateType struct {
	basetypes.StringType
}

type BackwardRFC3339Date struct {
	basetypes.StringValue
}

func NewBackwardRFC3339Date(value string) BackwardRFC3339Date {
	return BackwardRFC3339Date{
		StringValue: basetypes.NewStringValue(value),
	}
}

// Type returns a BackwardRFC3339DateType.
func (v BackwardRFC3339Date) Type(_ context.Context) attr.Type {
	return BackwardRFC3339DateType{}
}

// Equal returns true if the given value is equivalent.
func (v BackwardRFC3339Date) Equal(o attr.Value) bool {
	other, ok := o.(BackwardRFC3339Date)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v BackwardRFC3339Date) StringSemanticEquals(_ context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// The framework should always pass the correct value type, but always check
	newValue, ok := newValuable.(BackwardRFC3339Date)

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

	// Skipping error checking if CustomStringValue already implemented RFC3339 validation
	priorTime, _ := time.Parse(time.RFC3339, v.StringValue.ValueString())

	// Skipping error checking if CustomStringValue already implemented RFC3339 validation
	newTime, _ := time.Parse(time.RFC3339, newValue.ValueString())

	// If the new time is after the prior time, keep the prior value
	return priorTime.Before(newTime) || priorTime.Equal(newTime), diags
}

// String returns a human readable string of the type name.
func (t BackwardRFC3339DateType) String() string {
	return "BackwardRFC3339DateType"
}

// ValueType returns the Value type.
func (t BackwardRFC3339DateType) ValueType(ctx context.Context) attr.Value {
	return BackwardRFC3339Date{}
}

// Equal returns true if the given type is equivalent.
func (t BackwardRFC3339DateType) Equal(o attr.Type) bool {
	other, ok := o.(BackwardRFC3339DateType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

// ValueFromString returns a StringValuable type given a StringValue.
func (t BackwardRFC3339DateType) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	return BackwardRFC3339Date{
		StringValue: in,
	}, nil
}

// ValueFromTerraform returns a Value given a tftypes.Value.  This is meant to convert the tftypes.Value into a more convenient Go type
// for the provider to consume the data with.
func (t BackwardRFC3339DateType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)
	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}
