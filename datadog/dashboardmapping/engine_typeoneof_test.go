package dashboardmapping

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// makeStrAttr builds a types.String attr.Value from a Go string.
func makeStrAttr(s string) attr.Value {
	return types.StringValue(s)
}

// makeListAttr builds a types.List holding a single types.Object with the given attrs.
func makeListAttr(attrTypes map[string]attr.Type, objAttrs map[string]attr.Value) attr.Value {
	objType := types.ObjectType{AttrTypes: attrTypes}
	obj, _ := types.ObjectValue(attrTypes, objAttrs)
	lst, _ := types.ListValue(objType, []attr.Value{obj})
	return lst
}

// makeEmptyListAttr builds an empty types.List with the given element type.
func makeEmptyListAttr(attrTypes map[string]attr.Type) attr.Value {
	objType := types.ObjectType{AttrTypes: attrTypes}
	lst, _ := types.ListValue(objType, nil)
	return lst
}

// TestTypeOneOf_Build_NumberFormatUnit_Canonical verifies that BuildEngineJSON
// correctly handles a TypeOneOf field (unit) by injecting the discriminator
// and serializing the matched variant's children.
func TestTypeOneOf_Build_NumberFormatUnit_Canonical(t *testing.T) {
	// Build the canonical variant attrs:
	// unit { canonical { unit_name = "byte" } }
	canonAttrTypes := FieldSpecsToAttrTypes(numberFormatUnitCanonicalFields)
	customAttrTypes := FieldSpecsToAttrTypes(numberFormatUnitCustomFields)

	canonAttrs := map[string]attr.Value{
		"unit_name":     makeStrAttr("byte"),
		"per_unit_name": makeStrAttr(""),
	}
	canonBlock := makeListAttr(canonAttrTypes, canonAttrs)
	customBlock := makeEmptyListAttr(customAttrTypes)

	// The TypeOneOf "unit" block object
	unitAttrTypes := FieldSpecsToAttrTypes(widgetNumberFormatFields[0].Children)
	unitAttrs := map[string]attr.Value{
		"canonical": canonBlock,
		"custom":    customBlock,
	}
	unitBlock := makeListAttr(unitAttrTypes, unitAttrs)

	// Build the number_format block attrs (unit + unit_scale)
	unitScaleAttrTypes := FieldSpecsToAttrTypes(numberFormatUnitScaleFields)
	emptyUnitScale := makeEmptyListAttr(unitScaleAttrTypes)

	nfAttrTypes := FieldSpecsToAttrTypes(widgetNumberFormatFields)
	nfAttrs := map[string]attr.Value{
		"unit":       unitBlock,
		"unit_scale": emptyUnitScale,
	}

	_ = nfAttrTypes // used for construction verification

	result := BuildEngineJSON(nfAttrs, widgetNumberFormatFields)

	// Verify JSON output
	got, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	unitJSON, ok := result["unit"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result[\"unit\"] to be map, got: %s", got)
	}
	if unitJSON["type"] != "canonical_unit" {
		t.Errorf("expected type=canonical_unit, got %v", unitJSON["type"])
	}
	if unitJSON["unit_name"] != "byte" {
		t.Errorf("expected unit_name=byte, got %v", unitJSON["unit_name"])
	}
	if _, hasCustom := unitJSON["custom"]; hasCustom {
		t.Error("unexpected 'custom' key in unit JSON")
	}
}

// TestTypeOneOf_Build_NumberFormatUnit_Custom verifies the custom variant.
func TestTypeOneOf_Build_NumberFormatUnit_Custom(t *testing.T) {
	canonAttrTypes := FieldSpecsToAttrTypes(numberFormatUnitCanonicalFields)
	customAttrTypes := FieldSpecsToAttrTypes(numberFormatUnitCustomFields)

	customAttrs := map[string]attr.Value{
		"label": makeStrAttr("bytes"),
	}
	customBlock := makeListAttr(customAttrTypes, customAttrs)
	canonBlock := makeEmptyListAttr(canonAttrTypes)

	unitAttrTypes := FieldSpecsToAttrTypes(widgetNumberFormatFields[0].Children)
	unitAttrs := map[string]attr.Value{
		"canonical": canonBlock,
		"custom":    customBlock,
	}
	unitBlock := makeListAttr(unitAttrTypes, unitAttrs)

	unitScaleAttrTypes := FieldSpecsToAttrTypes(numberFormatUnitScaleFields)
	emptyUnitScale := makeEmptyListAttr(unitScaleAttrTypes)

	nfAttrs := map[string]attr.Value{
		"unit":       unitBlock,
		"unit_scale": emptyUnitScale,
	}

	result := BuildEngineJSON(nfAttrs, widgetNumberFormatFields)

	unitJSON, ok := result["unit"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result[\"unit\"] to be map")
	}
	if unitJSON["type"] != "custom_unit_label" {
		t.Errorf("expected type=custom_unit_label, got %v", unitJSON["type"])
	}
	if unitJSON["label"] != "bytes" {
		t.Errorf("expected label=bytes, got %v", unitJSON["label"])
	}
}

// TestTypeOneOf_Flatten_NumberFormatUnit_Canonical verifies that FlattenEngineJSON
// correctly dispatches on the discriminator field and populates only the matched variant.
func TestTypeOneOf_Flatten_NumberFormatUnit_Canonical(t *testing.T) {
	jsonData := map[string]interface{}{
		"unit": map[string]interface{}{
			"type":      "canonical_unit",
			"unit_name": "byte",
		},
	}

	result := FlattenEngineJSON(widgetNumberFormatFields, jsonData)

	unitState, ok := result["unit"].([]interface{})
	if !ok || len(unitState) == 0 {
		t.Fatalf("expected unit state list, got: %v", result["unit"])
	}
	unitMap, ok := unitState[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected unit state map, got: %T", unitState[0])
	}
	canonList, ok := unitMap["canonical"].([]interface{})
	if !ok || len(canonList) == 0 {
		t.Fatalf("expected canonical block in unit state, got: %v", unitMap)
	}
	canonMap, ok := canonList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected canonical map")
	}
	if canonMap["unit_name"] != "byte" {
		t.Errorf("expected unit_name=byte, got %v", canonMap["unit_name"])
	}
	// custom variant should not be populated
	if _, ok := unitMap["custom"]; ok {
		t.Error("unexpected 'custom' key in flattened unit state")
	}
}

// TestTypeOneOf_Flatten_NumberFormatUnit_Custom verifies the custom variant flatten.
func TestTypeOneOf_Flatten_NumberFormatUnit_Custom(t *testing.T) {
	jsonData := map[string]interface{}{
		"unit": map[string]interface{}{
			"type":  "custom_unit_label",
			"label": "bytes",
		},
	}

	result := FlattenEngineJSON(widgetNumberFormatFields, jsonData)

	unitState := result["unit"].([]interface{})
	unitMap := unitState[0].(map[string]interface{})

	customList, ok := unitMap["custom"].([]interface{})
	if !ok || len(customList) == 0 {
		t.Fatalf("expected custom block in unit state, got: %v", unitMap)
	}
	customMap := customList[0].(map[string]interface{})
	if customMap["label"] != "bytes" {
		t.Errorf("expected label=bytes, got %v", customMap["label"])
	}
	if _, ok := unitMap["canonical"]; ok {
		t.Error("unexpected 'canonical' key in flattened unit state")
	}
}

// TestTypeOneOf_Flatten_UnknownDiscriminator verifies that an unknown discriminator
// produces empty state (no panic, no match).
func TestTypeOneOf_Flatten_UnknownDiscriminator(t *testing.T) {
	jsonData := map[string]interface{}{
		"unit": map[string]interface{}{
			"type":      "unknown_future_type",
			"some_field": "value",
		},
	}

	result := FlattenEngineJSON(widgetNumberFormatFields, jsonData)

	// Unknown discriminator: no DefaultVariant on either child, so unit should not appear
	if _, ok := result["unit"]; ok {
		t.Error("expected no unit state for unknown discriminator, but got one")
	}
}

// TestTypeOneOf_Flatten_MultiValue verifies multi-value discriminator matching.
// Uses a minimal test FieldSpec to simulate the SunburstLegend pattern.
func TestTypeOneOf_Flatten_MultiValue(t *testing.T) {
	// A minimal TypeOneOf that uses Values (not Value) for matching
	variantAFields := []FieldSpec{
		{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true, Description: "type"},
	}
	variantBFields := []FieldSpec{
		{HCLKey: "type", Type: TypeString, OmitEmpty: false, Required: true, Description: "type"},
		{HCLKey: "extra", Type: TypeString, OmitEmpty: true, Description: "extra"},
	}
	oneOfField := FieldSpec{
		HCLKey: "legend",
		Type:   TypeOneOf,
		Discriminator: &OneOfDiscriminator{JSONKey: "type"},
		Children: []FieldSpec{
			{
				HCLKey:        "legend_table",
				Type:          TypeBlock,
				OmitEmpty:     true,
				Discriminator: &OneOfDiscriminator{Values: []string{"table", "none"}},
				Children:      variantAFields,
			},
			{
				HCLKey:        "legend_inline",
				Type:          TypeBlock,
				OmitEmpty:     true,
				Discriminator: &OneOfDiscriminator{Values: []string{"inline", "automatic"}},
				Children:      variantBFields,
			},
		},
	}

	fields := []FieldSpec{oneOfField}

	// Test "none" type → legend_table
	result := FlattenEngineJSON(fields, map[string]interface{}{
		"legend": map[string]interface{}{"type": "none"},
	})

	legendState := result["legend"].([]interface{})
	legendMap := legendState[0].(map[string]interface{})
	if _, ok := legendMap["legend_table"]; !ok {
		t.Error("expected legend_table to be populated for type=none")
	}
	if _, ok := legendMap["legend_inline"]; ok {
		t.Error("unexpected legend_inline for type=none")
	}

	// Test "automatic" type → legend_inline
	result2 := FlattenEngineJSON(fields, map[string]interface{}{
		"legend": map[string]interface{}{"type": "automatic", "extra": "foo"},
	})
	legendState2 := result2["legend"].([]interface{})
	legendMap2 := legendState2[0].(map[string]interface{})
	if _, ok := legendMap2["legend_inline"]; !ok {
		t.Error("expected legend_inline to be populated for type=automatic")
	}
}

// TestTypeOneOf_DefaultVariant verifies that DefaultVariant is used when no discriminator
// field exists in the JSON (e.g. WidgetLegacyLiveSpan pattern).
func TestTypeOneOf_DefaultVariant(t *testing.T) {
	legacyFields := []FieldSpec{
		{HCLKey: "live_span", Type: TypeString, OmitEmpty: true, Description: "legacy span"},
	}
	newLiveFields := []FieldSpec{
		{HCLKey: "value", Type: TypeInt, OmitEmpty: false, Required: true, Description: "value"},
		{HCLKey: "unit", Type: TypeString, OmitEmpty: false, Required: true, Description: "unit"},
	}
	oneOfField := FieldSpec{
		HCLKey: "time",
		Type:   TypeOneOf,
		Discriminator: &OneOfDiscriminator{JSONKey: "type"},
		Children: []FieldSpec{
			{
				HCLKey:        "legacy",
				Type:          TypeBlock,
				OmitEmpty:     true,
				Discriminator: &OneOfDiscriminator{DefaultVariant: true},
				Children:      legacyFields,
			},
			{
				HCLKey:        "live",
				Type:          TypeBlock,
				OmitEmpty:     true,
				Discriminator: &OneOfDiscriminator{Value: "live"},
				Children:      newLiveFields,
			},
		},
	}

	fields := []FieldSpec{oneOfField}

	// JSON with no "type" field → should match DefaultVariant (legacy)
	result := FlattenEngineJSON(fields, map[string]interface{}{
		"time": map[string]interface{}{"live_span": "5m"},
	})

	timeState := result["time"].([]interface{})
	timeMap := timeState[0].(map[string]interface{})
	if _, ok := timeMap["legacy"]; !ok {
		t.Error("expected legacy block to be populated for no-type JSON")
	}
	if _, ok := timeMap["live"]; ok {
		t.Error("unexpected live block for no-type JSON")
	}

	// JSON with type="live" → should match live variant
	result2 := FlattenEngineJSON(fields, map[string]interface{}{
		"time": map[string]interface{}{"type": "live", "value": float64(4), "unit": "hour"},
	})
	timeState2 := result2["time"].([]interface{})
	timeMap2 := timeState2[0].(map[string]interface{})
	if _, ok := timeMap2["live"]; !ok {
		t.Error("expected live block to be populated for type=live JSON")
	}
	if _, ok := timeMap2["legacy"]; ok {
		t.Error("unexpected legacy block for type=live JSON")
	}
}
