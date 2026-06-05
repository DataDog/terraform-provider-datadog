package dashboardmapping

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dashboardDefaultTimeframeFields corresponds to OpenAPI DashboardDefaultTimeframe.
var dashboardDefaultTimeframeFields = []FieldSpec{
	{HCLKey: "type", Type: TypeString, Required: true, OmitEmpty: false,
		ValidValues: []string{"live", "fixed"},
		Description: "The type of timeframe. Valid values are `live`, `fixed`."},
	{HCLKey: "unit", Type: TypeString, OmitEmpty: true,
		ValidValues: []string{"minute", "hour", "day", "week", "month", "year"},
		Description: "Unit of the live timeframe span. Required when `type` is `live`."},
	{HCLKey: "value", Type: TypeInt, OmitEmpty: true,
		Description: "Value of the live timeframe span. Required when `type` is `live`."},
	{HCLKey: "from", Type: TypeInt, OmitEmpty: true,
		Description: "Start time in milliseconds since epoch. Required when `type` is `fixed`."},
	{HCLKey: "to", Type: TypeInt, OmitEmpty: true,
		Description: "End time in milliseconds since epoch. Required when `type` is `fixed`."},
}

// DashboardDefaultTimeframeSchema returns the SDKv2 schema for default_timeframe.
func DashboardDefaultTimeframeSchema() *schema.Schema {
	return FieldSpecToSDKv2(FieldSpec{
		HCLKey:      "default_timeframe",
		Type:        TypeBlock,
		OmitEmpty:   true,
		Description: "The default timeframe applied when opening the dashboard. Set to `null` to disable after it has been configured.",
		Children:    dashboardDefaultTimeframeFields,
	})
}

// CollectDefaultTimeframeData returns the value to store in a dashboard data map for JSON building.
// An empty slice means omit the field; nil means send explicit null.
func CollectDefaultTimeframeData(d *schema.ResourceData) interface{} {
	blocks := d.Get("default_timeframe").([]interface{})
	if len(blocks) > 0 {
		return blocks
	}
	if !d.IsNewResource() && d.HasChange("default_timeframe") {
		return nil
	}
	return []interface{}{}
}

// ApplyDefaultTimeframeToDashboardJSON sets default_timeframe on a dashboard JSON body.
func ApplyDefaultTimeframeToDashboardJSON(result map[string]interface{}, data map[string]interface{}) error {
	tf, ok := data["default_timeframe"]
	if !ok {
		return nil
	}
	if tf == nil {
		result["default_timeframe"] = nil
		return nil
	}
	blocks, ok := tf.([]interface{})
	if !ok || len(blocks) == 0 {
		return nil
	}
	block, ok := blocks[0].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid default_timeframe block")
	}
	built, err := BuildDefaultTimeframeJSONFromMap(block)
	if err != nil {
		return err
	}
	result["default_timeframe"] = built
	return nil
}

// BuildDefaultTimeframeJSONFromMap converts a Terraform default_timeframe block to API JSON.
func BuildDefaultTimeframeJSONFromMap(block map[string]interface{}) (map[string]interface{}, error) {
	typeVal, ok := block["type"].(string)
	if !ok || typeVal == "" {
		return nil, fmt.Errorf("default_timeframe.type is required")
	}

	result := map[string]interface{}{"type": typeVal}
	switch typeVal {
	case "live":
		unit, ok := block["unit"].(string)
		if !ok || unit == "" {
			return nil, fmt.Errorf("default_timeframe.unit is required when type is live")
		}
		if _, ok := block["value"]; !ok {
			return nil, fmt.Errorf("default_timeframe.value is required when type is live")
		}
		result["unit"] = unit
		result["value"] = getIntFromMap(block, "value")
	case "fixed":
		if _, ok := block["from"]; !ok {
			return nil, fmt.Errorf("default_timeframe.from is required when type is fixed")
		}
		if _, ok := block["to"]; !ok {
			return nil, fmt.Errorf("default_timeframe.to is required when type is fixed")
		}
		result["from"] = getIntFromMap(block, "from")
		result["to"] = getIntFromMap(block, "to")
	default:
		return nil, fmt.Errorf("invalid default_timeframe.type %q, must be live or fixed", typeVal)
	}
	return result, nil
}

// FlattenDefaultTimeframe converts API default_timeframe JSON to Terraform state.
func FlattenDefaultTimeframe(api map[string]interface{}) []interface{} {
	if api == nil {
		return nil
	}
	typeVal, _ := api["type"].(string)
	if typeVal == "" {
		return nil
	}

	// Initialize all schema fields explicitly so d.Set receives a complete
	// map. Passing a partial map (missing TypeInt fields) causes Terraform
	// 1.1.x to store 0 for the value field in state.
	block := map[string]interface{}{
		"type":  typeVal,
		"unit":  "",
		"value": 0,
		"from":  0,
		"to":    0,
	}
	switch typeVal {
	case "live":
		if unit, ok := api["unit"].(string); ok {
			block["unit"] = unit
		}
		if _, ok := api["value"]; ok {
			block["value"] = getIntFromMap(api, "value")
		}
	case "fixed":
		if _, ok := api["from"]; ok {
			block["from"] = getIntFromMap(api, "from")
		}
		if _, ok := api["to"]; ok {
			block["to"] = getIntFromMap(api, "to")
		}
	}
	return []interface{}{block}
}
