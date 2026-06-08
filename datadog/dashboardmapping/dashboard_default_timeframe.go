package dashboardmapping

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// dashboardDefaultTimeframeLiveFields corresponds to the "live" variant of DashboardDefaultTimeframe.
var dashboardDefaultTimeframeLiveFields = []FieldSpec{
	{HCLKey: "unit", Type: TypeString, OmitEmpty: false, Required: true,
		ValidValues: []string{"minute", "hour", "day", "week", "month", "year"},
		Description: "Unit of the live timeframe span."},
	{HCLKey: "value", Type: TypeInt, OmitEmpty: false, Required: true,
		Description: "Value of the live timeframe span."},
}

// dashboardDefaultTimeframeFixedFields corresponds to the "fixed" variant of DashboardDefaultTimeframe.
var dashboardDefaultTimeframeFixedFields = []FieldSpec{
	{HCLKey: "from", Type: TypeInt, OmitEmpty: false, Required: true,
		Description: "Start time in milliseconds since epoch."},
	{HCLKey: "to", Type: TypeInt, OmitEmpty: false, Required: true,
		Description: "End time in milliseconds since epoch."},
}

// dashboardDefaultTimeframeFields is the flat field list used by the v1 datadog_dashboard resource.
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

// DashboardDefaultTimeframeField returns the TypeOneOf FieldSpec for default_timeframe.
// Used by the v2 datadog_dashboard_v2 resource via DashboardTopLevelFields.
func DashboardDefaultTimeframeField() FieldSpec {
	return FieldSpec{
		HCLKey:      "default_timeframe",
		Type:        TypeOneOf,
		OmitEmpty:   true,
		NullOnClear: true,
		Description: "The default timeframe applied when opening the dashboard. Set to `null` to disable after it has been configured.",
		Discriminator: &OneOfDiscriminator{JSONKey: "type"},
		Children: []FieldSpec{
			{HCLKey: "live", Type: TypeBlock, OmitEmpty: true,
				Discriminator: &OneOfDiscriminator{Value: "live"},
				Description:   "A live timeframe applied when opening the dashboard.",
				Children:      dashboardDefaultTimeframeLiveFields},
			{HCLKey: "fixed", Type: TypeBlock, OmitEmpty: true,
				Discriminator: &OneOfDiscriminator{Value: "fixed"},
				Description:   "A fixed timeframe applied when opening the dashboard.",
				Children:      dashboardDefaultTimeframeFixedFields},
		},
	}
}

// DashboardDefaultTimeframeSchema returns the flat SDKv2 schema for default_timeframe.
// Used by the v1 datadog_dashboard resource.
func DashboardDefaultTimeframeSchema() *schema.Schema {
	return FieldSpecToSDKv2(FieldSpec{
		HCLKey:      "default_timeframe",
		Type:        TypeBlock,
		OmitEmpty:   true,
		Description: "The default timeframe applied when opening the dashboard. Set to `null` to disable after it has been configured.",
		Children:    dashboardDefaultTimeframeFields,
	})
}

// BuildDefaultTimeframeJSONFromMap converts a v1 flat default_timeframe block to API JSON.
// Used by the v1 datadog_dashboard resource.
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

// FlattenDefaultTimeframe converts API default_timeframe JSON to flat Terraform state.
// Used by the v1 datadog_dashboard resource.
func FlattenDefaultTimeframe(api map[string]interface{}) []interface{} {
	if api == nil {
		return nil
	}
	typeVal, _ := api["type"].(string)
	if typeVal == "" {
		return nil
	}

	block := map[string]interface{}{"type": typeVal}
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
