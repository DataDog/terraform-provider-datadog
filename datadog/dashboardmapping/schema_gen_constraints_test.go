package dashboardmapping

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestFieldSpecListConstraints(t *testing.T) {
	computeSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:   "compute",
		Type:     TypeBlockList,
		MinItems: 1,
		MaxItems: 5,
		Children: []FieldSpec{{
			HCLKey:   "aggregation",
			Type:     TypeString,
			Required: true,
		}},
	})
	if computeSchema.MinItems != 1 || computeSchema.MaxItems != 5 {
		t.Fatalf("block list constraints were not registered: %#v", computeSchema)
	}

	statesSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:      "states",
		Type:        TypeStringList,
		MinItems:    1,
		MaxItems:    2,
		ValidValues: []string{"OPEN", "RESOLVED"},
	})
	if statesSchema.MinItems != 1 || statesSchema.MaxItems != 2 {
		t.Fatalf("string list constraints were not registered: %#v", statesSchema)
	}
	elem, ok := statesSchema.Elem.(*schema.Schema)
	if !ok || elem.ValidateDiagFunc == nil {
		t.Fatalf("string-list enum validation was not registered: %#v", statesSchema.Elem)
	}
	if diagnostics := elem.ValidateDiagFunc("OPEN", cty.Path{}); len(diagnostics) != 0 {
		t.Fatalf("valid string-list enum value was rejected: %#v", diagnostics)
	}
	if diagnostics := elem.ValidateDiagFunc("CLOSED", cty.Path{}); len(diagnostics) == 0 {
		t.Fatal("invalid string-list enum value was accepted")
	}

	thresholdsSchema := FieldSpecToSDKv2(FieldSpec{
		HCLKey:   "thresholds",
		Type:     TypeIntList,
		MinItems: 2,
		MaxItems: 4,
	})
	if thresholdsSchema.MinItems != 2 || thresholdsSchema.MaxItems != 4 {
		t.Fatalf("int list constraints were not registered: %#v", thresholdsSchema)
	}
}
