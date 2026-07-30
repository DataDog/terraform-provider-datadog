package dashboardmapping

import (
	"testing"

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
}
