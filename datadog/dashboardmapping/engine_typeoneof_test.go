package dashboardmapping

import (
	"testing"
)

// TestTypeOneOf_Build_NumberFormatUnit_Canonical verifies that BuildEngineJSONFromMap
// correctly handles a TypeOneOf field (unit) by injecting the discriminator
// and serializing the matched variant's children.
func TestTypeOneOf_Build_NumberFormatUnit_Canonical(t *testing.T) {
	// Simulate SDKv2 map: unit { canonical { unit_name = "byte" } }
	nfData := map[string]interface{}{
		"unit": []interface{}{
			map[string]interface{}{
				"canonical": []interface{}{
					map[string]interface{}{
						"unit_name":     "byte",
						"per_unit_name": "",
					},
				},
				"custom": []interface{}{},
			},
		},
		"unit_scale": []interface{}{},
	}

	result := BuildEngineJSONFromMap(nfData, widgetNumberFormatFields)

	unitJSON, ok := result["unit"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result[\"unit\"] to be map, got: %T", result["unit"])
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
	nfData := map[string]interface{}{
		"unit": []interface{}{
			map[string]interface{}{
				"canonical": []interface{}{},
				"custom": []interface{}{
					map[string]interface{}{
						"label": "bytes",
					},
				},
			},
		},
		"unit_scale": []interface{}{},
	}

	result := BuildEngineJSONFromMap(nfData, widgetNumberFormatFields)

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
			"type":       "unknown_future_type",
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
		HCLKey:        "legend",
		Type:          TypeOneOf,
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

// TestTypeOneOf_Build_ToplistDisplay_Stacked verifies that the stacked variant injects
// the discriminator "type":"stacked" and includes the optional "legend" field.
func TestTypeOneOf_Build_ToplistDisplay_Stacked(t *testing.T) {
	// display { stacked { legend = "automatic" } }
	styleData := map[string]interface{}{
		"display": []interface{}{
			map[string]interface{}{
				"stacked": []interface{}{
					map[string]interface{}{"legend": "automatic"},
				},
				"flat": []interface{}{},
			},
		},
		"palette": "",
		"scaling": "",
	}

	result := BuildEngineJSONFromMap(styleData, toplistWidgetStyleFields)
	displayJSON, ok := result["display"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected display to be a map, got: %T (%v)", result["display"], result["display"])
	}
	if displayJSON["type"] != "stacked" {
		t.Errorf("expected type=stacked, got %v", displayJSON["type"])
	}
	if displayJSON["legend"] != "automatic" {
		t.Errorf("expected legend=automatic, got %v", displayJSON["legend"])
	}
}

// TestTypeOneOf_Build_ToplistDisplay_Flat verifies that the flat variant injects
// the discriminator "type":"flat" and emits no other fields.
func TestTypeOneOf_Build_ToplistDisplay_Flat(t *testing.T) {
	// display { flat {} }
	styleData := map[string]interface{}{
		"display": []interface{}{
			map[string]interface{}{
				"stacked": []interface{}{},
				"flat": []interface{}{
					map[string]interface{}{},
				},
			},
		},
		"palette": "",
		"scaling": "",
	}

	result := BuildEngineJSONFromMap(styleData, toplistWidgetStyleFields)
	displayJSON, ok := result["display"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected display to be a map, got: %T", result["display"])
	}
	if displayJSON["type"] != "flat" {
		t.Errorf("expected type=flat, got %v", displayJSON["type"])
	}
	if _, hasLegend := displayJSON["legend"]; hasLegend {
		t.Error("unexpected 'legend' key in flat display JSON")
	}
}

// TestTypeOneOf_Flatten_ToplistDisplay_Stacked verifies flatten of stacked display.
func TestTypeOneOf_Flatten_ToplistDisplay_Stacked(t *testing.T) {
	jsonData := map[string]interface{}{
		"display": map[string]interface{}{
			"type":   "stacked",
			"legend": "automatic",
		},
	}

	result := FlattenEngineJSON(toplistWidgetStyleFields, jsonData)
	displayState, ok := result["display"].([]interface{})
	if !ok || len(displayState) == 0 {
		t.Fatalf("expected display state list, got: %v", result["display"])
	}
	displayMap, ok := displayState[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected display state map")
	}
	stackedList, ok := displayMap["stacked"].([]interface{})
	if !ok || len(stackedList) == 0 {
		t.Fatalf("expected stacked block in display state, got: %v", displayMap)
	}
	stackedMap, ok := stackedList[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected stacked map")
	}
	if stackedMap["legend"] != "automatic" {
		t.Errorf("expected legend=automatic, got %v", stackedMap["legend"])
	}
	if _, ok := displayMap["flat"]; ok {
		t.Error("unexpected 'flat' key in display state for stacked variant")
	}
}

// TestTypeOneOf_Flatten_ToplistDisplay_Flat verifies flatten of flat display.
func TestTypeOneOf_Flatten_ToplistDisplay_Flat(t *testing.T) {
	jsonData := map[string]interface{}{
		"display": map[string]interface{}{
			"type": "flat",
		},
	}

	result := FlattenEngineJSON(toplistWidgetStyleFields, jsonData)
	displayState, ok := result["display"].([]interface{})
	if !ok || len(displayState) == 0 {
		t.Fatalf("expected display state list, got: %v", result["display"])
	}
	displayMap, ok := displayState[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected display state map")
	}
	flatList, ok := displayMap["flat"].([]interface{})
	if !ok || len(flatList) == 0 {
		t.Fatalf("expected flat block in display state, got: %v", displayMap)
	}
	if _, ok := displayMap["stacked"]; ok {
		t.Error("unexpected 'stacked' key in display state for flat variant")
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
		HCLKey:        "time",
		Type:          TypeOneOf,
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
