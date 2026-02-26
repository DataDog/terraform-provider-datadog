package dashboardmapping

// engine.go
//
// FieldSpec bidirectional mapping engine for the dashboard resource.
// This file contains the generic engine types (FieldSpec, WidgetSpec, FieldType),
// the HCL↔JSON engine functions, path helpers, and the dashboard-level build/flatten
// entry points.

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// FieldType drives serialization/deserialization behavior in the engine.
type FieldType int

const (
	TypeString     FieldType = iota // plain string
	TypeBool                        // bool
	TypeInt                         // int
	TypeFloat                       // float64
	TypeStringList                  // []string (TypeList or TypeSet of strings)
	TypeIntList                     // []int
	TypeBlock                       // single nested block (MaxItems:1 in schema)
	TypeBlockList                   // list of blocks
)

// FieldSpec declares the bidirectional mapping for a single field.
// One declaration covers both the HCL→JSON build direction and the
// JSON→HCL flatten direction.
type FieldSpec struct {
	// HCLKey is the key used in the terraform schema (e.g., "title", "search_query").
	HCLKey string

	// JSONKey overrides the JSON key when it differs from HCLKey.
	// Example: HCLKey="compute_query", JSONKey="compute"
	// If empty, HCLKey is used as the JSON key.
	JSONKey string

	// JSONPath handles fields that nest into a sub-object in JSON but are flat in HCL.
	// Example: HCLKey="live_span", JSONPath="time.live_span"
	//   HCL: live_span = "5m"  →  JSON: {"time": {"live_span": "5m"}}
	// If set, JSONKey is ignored.
	JSONPath string

	// Type drives serialization and deserialization behavior.
	Type FieldType

	// OmitEmpty: if true, the field is omitted from JSON when it is the zero value.
	// Derived by comparing cassette request bodies against HCL configs.
	OmitEmpty bool

	// Children: for TypeBlock and TypeBlockList, the nested field specs.
	Children []FieldSpec

	// ── Terraform schema metadata ──────────────────────────────────────
	// Description is shown in `make docs` output. Required for docs generation.
	Description string

	// Required: true → schema.Required; false (default) → schema.Optional
	Required bool

	// Computed: true for fields set by the API (e.g. widget IDs, URLs)
	Computed bool

	// Default: default value emitted in schema (e.g. true for has_background)
	Default interface{}

	// MaxItems: override for TypeBlockList (default 0 = unlimited)
	// TypeBlock always uses MaxItems: 1 automatically.
	MaxItems int

	// Sensitive: mask this field in logs and UI
	Sensitive bool

	// Deprecated: non-empty string = deprecation message
	Deprecated string

	// ValidValues: valid string values for enum fields.
	// Generates validators.ValidateEnumValue automatically.
	// Use instead of SDK-based enum validators.
	ValidValues []string

	// ConflictsWith: field paths this field conflicts with.
	// Populated post-generation for the rare fields that need it.
	ConflictsWith []string

	// UseSet: use schema.TypeSet instead of schema.TypeList for list fields.
	// Rare — only for fields that require set semantics.
	UseSet bool

	// ForceNew: when true, changing this field forces resource replacement.
	ForceNew bool

	// SchemaOnly: if true, this field is for schema registration only.
	// BuildEngineJSON skips it. Use for fields like dashboard_lists
	// that are managed as side effects, not serialized to the API.
	SchemaOnly bool
}

// effectiveJSONKey returns the JSON key or path root for a FieldSpec.
func (f *FieldSpec) effectiveJSONKey() string {
	if f.JSONPath != "" {
		return strings.SplitN(f.JSONPath, ".", 2)[0]
	}
	if f.JSONKey != "" {
		return f.JSONKey
	}
	return f.HCLKey
}

// effectiveJSONPath returns the full dotted JSON path for a FieldSpec.
func (f *FieldSpec) effectiveJSONPath() string {
	if f.JSONPath != "" {
		return f.JSONPath
	}
	return f.effectiveJSONKey()
}

// WidgetSpec maps an HCL widget definition block to a JSON widget type.
type WidgetSpec struct {
	// HCLKey is the schema key for this widget's definition block.
	// Example: "timeseries_definition"
	HCLKey string

	// JSONType is the "type" string used in the JSON definition object.
	// Example: "timeseries"
	JSONType string

	// Description is shown in `make docs` output for the outer widget definition block.
	Description string

	// Fields are the widget-specific fields.
	// CommonWidgetFields are automatically merged in by the engine.
	Fields []FieldSpec
}

// ============================================================
// Framework attr.Value helpers
// ============================================================

// getStrAttr gets a string attr, returns ("", false) if null/unknown/missing.
func getStrAttr(attrs map[string]attr.Value, key string) (string, bool) {
	raw, ok := attrs[key]
	if !ok {
		return "", false
	}
	sv, ok := raw.(types.String)
	if !ok || sv.IsNull() || sv.IsUnknown() {
		return "", false
	}
	return sv.ValueString(), true
}

// getBoolAttr gets a bool attr, returns (false, false) if null/unknown/missing.
func getBoolAttr(attrs map[string]attr.Value, key string) (bool, bool) {
	raw, ok := attrs[key]
	if !ok {
		return false, false
	}
	bv, ok := raw.(types.Bool)
	if !ok || bv.IsNull() || bv.IsUnknown() {
		return false, false
	}
	return bv.ValueBool(), true
}

// getInt64Attr gets an int64 attr, returns (0, false) if null/unknown/missing.
func getInt64Attr(attrs map[string]attr.Value, key string) (int64, bool) {
	raw, ok := attrs[key]
	if !ok {
		return 0, false
	}
	iv, ok := raw.(types.Int64)
	if !ok || iv.IsNull() || iv.IsUnknown() {
		return 0, false
	}
	return iv.ValueInt64(), true
}

// getFloat64Attr gets a float64 attr, returns (0, false) if null/unknown/missing.
func getFloat64Attr(attrs map[string]attr.Value, key string) (float64, bool) {
	raw, ok := attrs[key]
	if !ok {
		return 0, false
	}
	fv, ok := raw.(types.Float64)
	if !ok || fv.IsNull() || fv.IsUnknown() {
		return 0, false
	}
	return fv.ValueFloat64(), true
}

// getListElems returns the elements of a list/set attr, or nil if missing/null/empty.
func getListElems(attrs map[string]attr.Value, key string) []attr.Value {
	raw, ok := attrs[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case types.List:
		if v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.Elements()
	case types.Set:
		if v.IsNull() || v.IsUnknown() {
			return nil
		}
		return v.Elements()
	}
	return nil
}

// getObjAttrs returns the attributes map of the first element of a list attr,
// or nil if missing/null/empty.
func getObjAttrs(attrs map[string]attr.Value, key string) map[string]attr.Value {
	elems := getListElems(attrs, key)
	if len(elems) == 0 {
		return nil
	}
	obj, ok := elems[0].(types.Object)
	if !ok {
		return nil
	}
	return obj.Attributes()
}

// getStrListElems returns the string values from a list/set of strings attr.
func getStrListElems(attrs map[string]attr.Value, key string) []string {
	elems := getListElems(attrs, key)
	if len(elems) == 0 {
		return nil
	}
	result := make([]string, 0, len(elems))
	for _, elem := range elems {
		if sv, ok := elem.(types.String); ok && !sv.IsNull() {
			result = append(result, sv.ValueString())
		}
	}
	return result
}

// getInt64ListElems returns the int64 values from a list of int64 attrs.
func getInt64ListElems(attrs map[string]attr.Value, key string) []int64 {
	elems := getListElems(attrs, key)
	if len(elems) == 0 {
		return nil
	}
	result := make([]int64, 0, len(elems))
	for _, elem := range elems {
		if iv, ok := elem.(types.Int64); ok && !iv.IsNull() {
			result = append(result, iv.ValueInt64())
		}
	}
	return result
}

// ============================================================
// Generic Engine: HCL → JSON (build direction)
// ============================================================

// BuildEngineJSON converts a framework attrs map to a JSON map.
func BuildEngineJSON(attrs map[string]attr.Value, fields []FieldSpec) map[string]interface{} {
	result := map[string]interface{}{}
	for _, f := range fields {
		if f.SchemaOnly {
			continue
		}
		raw, present := attrs[f.HCLKey]
		switch f.Type {
		case TypeString:
			var strVal string
			if present {
				if sv, ok := raw.(types.String); ok && !sv.IsNull() && !sv.IsUnknown() {
					strVal = sv.ValueString()
				}
			}
			if f.OmitEmpty && strVal == "" {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), strVal)

		case TypeBool:
			boolVal := false
			if present {
				if bv, ok := raw.(types.Bool); ok && !bv.IsNull() && !bv.IsUnknown() {
					boolVal = bv.ValueBool()
				}
			}
			if f.OmitEmpty && !boolVal {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), boolVal)

		case TypeInt:
			var intVal int
			if present {
				if iv, ok := raw.(types.Int64); ok && !iv.IsNull() && !iv.IsUnknown() {
					intVal = int(iv.ValueInt64())
				}
			}
			if f.OmitEmpty && intVal == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), intVal)

		case TypeFloat:
			var floatVal float64
			if present {
				if fv, ok := raw.(types.Float64); ok && !fv.IsNull() && !fv.IsUnknown() {
					floatVal = fv.ValueFloat64()
				}
			}
			if f.OmitEmpty && floatVal == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), floatVal)

		case TypeStringList:
			var strs []string
			if present {
				strs = getStrListElems(attrs, f.HCLKey)
			}
			if f.OmitEmpty && len(strs) == 0 {
				continue
			}
			if strs == nil {
				strs = []string{}
			}
			setAtJSONPath(result, f.effectiveJSONPath(), strs)

		case TypeIntList:
			var ints []int64
			if present {
				ints = getInt64ListElems(attrs, f.HCLKey)
			}
			if f.OmitEmpty && len(ints) == 0 {
				continue
			}
			if ints == nil {
				ints = []int64{}
			}
			setAtJSONPath(result, f.effectiveJSONPath(), ints)

		case TypeBlock:
			elems := getListElems(attrs, f.HCLKey)
			if len(elems) == 0 {
				continue
			}
			childObj, ok := elems[0].(types.Object)
			if !ok {
				continue
			}
			nested := BuildEngineJSON(childObj.Attributes(), f.Children)
			if len(nested) == 0 && f.OmitEmpty {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), nested)

		case TypeBlockList:
			elems := getListElems(attrs, f.HCLKey)
			items := make([]interface{}, 0, len(elems))
			for _, elem := range elems {
				childObj, ok := elem.(types.Object)
				if !ok {
					continue
				}
				items = append(items, BuildEngineJSON(childObj.Attributes(), f.Children))
			}
			if f.OmitEmpty && len(items) == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), items)
		}
	}
	return result
}

// setAtJSONPath sets a value at a dotted path within a nested map.
// Example: setAtJSONPath(m, "time.live_span", "5m") sets m["time"]["live_span"] = "5m"
func setAtJSONPath(m map[string]interface{}, path string, val interface{}) {
	parts := strings.SplitN(path, ".", 2)
	if len(parts) == 1 {
		m[path] = val
		return
	}
	sub, ok := m[parts[0]].(map[string]interface{})
	if !ok {
		sub = map[string]interface{}{}
		m[parts[0]] = sub
	}
	setAtJSONPath(sub, parts[1], val)
}

// ============================================================
// Generic Engine: JSON → HCL (flatten direction)
// ============================================================

// FlattenEngineJSON converts a JSON map into a map[string]interface{} suitable
// for setting on a TypeList field in schema.ResourceData.
func FlattenEngineJSON(fields []FieldSpec, data map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, f := range fields {
		if f.SchemaOnly {
			continue
		}
		jsonVal := getAtJSONPath(data, f.effectiveJSONPath())
		if jsonVal == nil {
			continue
		}

		switch f.Type {
		case TypeString:
			result[f.HCLKey] = fmt.Sprintf("%v", jsonVal)

		case TypeBool:
			if b, ok := jsonVal.(bool); ok {
				result[f.HCLKey] = b
			}

		case TypeInt:
			switch v := jsonVal.(type) {
			case float64:
				result[f.HCLKey] = int(v)
			case int:
				result[f.HCLKey] = v
			case int64:
				result[f.HCLKey] = int(v)
			}

		case TypeFloat:
			switch v := jsonVal.(type) {
			case float64:
				result[f.HCLKey] = v
			case int:
				result[f.HCLKey] = float64(v)
			}

		case TypeStringList:
			strs := toStringSliceFromInterface(jsonVal)
			if f.OmitEmpty && len(strs) == 0 {
				continue
			}
			result[f.HCLKey] = strs

		case TypeBlock:
			if m, ok := jsonVal.(map[string]interface{}); ok {
				result[f.HCLKey] = []interface{}{FlattenEngineJSON(f.Children, m)}
			}

		case TypeBlockList:
			if items, ok := jsonVal.([]interface{}); ok {
				list := make([]interface{}, len(items))
				for i, item := range items {
					if m, ok := item.(map[string]interface{}); ok {
						list[i] = FlattenEngineJSON(f.Children, m)
					}
				}
				result[f.HCLKey] = list
			}
		}
	}
	return result
}

// getAtJSONPath retrieves a value from a nested map following a dotted path.
func getAtJSONPath(m map[string]interface{}, path string) interface{} {
	parts := strings.SplitN(path, ".", 2)
	val, ok := m[parts[0]]
	if !ok || val == nil {
		return nil
	}
	if len(parts) == 1 {
		return val
	}
	sub, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}
	return getAtJSONPath(sub, parts[1])
}

// toStringSliceFromInterface converts an interface{} (expected []interface{}) to []string.
func toStringSliceFromInterface(raw interface{}) []string {
	items, ok := raw.([]interface{})
	if !ok {
		return []string{}
	}
	result := make([]string, len(items))
	for i, item := range items {
		result[i] = fmt.Sprintf("%v", item)
	}
	return result
}

// ============================================================
// Widget Type Dispatch
// ============================================================

// buildWidgetEngineJSON builds the JSON for a single widget by dispatching to the correct WidgetSpec.
// attrs is the attributes map of the widget block.
func buildWidgetEngineJSON(attrs map[string]attr.Value) map[string]interface{} {
	for _, spec := range allWidgetSpecs {
		elems := getListElems(attrs, spec.HCLKey)
		if len(elems) == 0 {
			continue
		}
		defObj, ok := elems[0].(types.Object)
		if !ok {
			continue
		}
		defAttrs := defObj.Attributes()
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := BuildEngineJSON(defAttrs, allFields)
		defJSON["type"] = spec.JSONType

		// Per-widget post-processing: formula/query blocks, type-specific fixups,
		// and Batch C complex widget handling — all consolidated in one hook.
		buildWidgetPostProcess(defAttrs, spec, defJSON)

		widgetJSON := map[string]interface{}{"definition": defJSON}

		// Include widget_layout if present (required for 'free' layout dashboards).
		// HCL: widget_layout { x, y, width, height, is_column_break }
		// JSON: layout { x, y, width, height, is_column_break }
		if layoutElems := getListElems(attrs, "widget_layout"); len(layoutElems) > 0 {
			if layoutObj, ok := layoutElems[0].(types.Object); ok {
				layoutAttrs := layoutObj.Attributes()
				layout := map[string]interface{}{}
				if v, ok := getInt64Attr(layoutAttrs, "x"); ok {
					layout["x"] = int(v)
				}
				if v, ok := getInt64Attr(layoutAttrs, "y"); ok {
					layout["y"] = int(v)
				}
				if v, ok := getInt64Attr(layoutAttrs, "width"); ok {
					layout["width"] = int(v)
				}
				if v, ok := getInt64Attr(layoutAttrs, "height"); ok {
					layout["height"] = int(v)
				}
				if v, ok := getBoolAttr(layoutAttrs, "is_column_break"); ok && v {
					layout["is_column_break"] = v
				}
				widgetJSON["layout"] = layout
			}
		}

		return widgetJSON
	}
	return nil
}

// buildFormulaQueryRequestJSON builds the JSON for a new-style formula/query request.
// These requests use `formulas`, `queries`, and `response_format` instead of the
// old-style `q`, `log_query`, etc.
func buildFormulaQueryRequestJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}

	// on_right_yaxis is always emitted (OmitEmpty: false)
	onRightYaxis, _ := getBoolAttr(attrs, "on_right_yaxis")
	result["on_right_yaxis"] = onRightYaxis

	// display_type (optional)
	if dt, ok := getStrAttr(attrs, "display_type"); ok && dt != "" {
		result["display_type"] = dt
	}

	// formulas — use widgetFormulaFields engine for standard fields.
	// number_format is excluded from widgetFormulaFields (polymorphic structure);
	// it is not present on timeseries formula requests.
	formulaElems := getListElems(attrs, "formula")
	if len(formulaElems) > 0 {
		formulas := make([]interface{}, 0, len(formulaElems))
		for _, elem := range formulaElems {
			if fObj, ok := elem.(types.Object); ok {
				formulas = append(formulas, BuildEngineJSON(fObj.Attributes(), widgetFormulaFields))
			}
		}
		result["formulas"] = formulas
	}

	// queries (polymorphic: each query block has one of metric_query, event_query, etc.)
	queryElems := getListElems(attrs, "query")
	if len(queryElems) > 0 {
		queries := make([]interface{}, 0, len(queryElems))
		for _, elem := range queryElems {
			if qObj, ok := elem.(types.Object); ok {
				queries = append(queries, buildQueryFromAttrs(qObj.Attributes()))
			}
		}
		result["queries"] = queries
	}

	// style (optional, present on timeseries formula requests at the request level)
	// This is request-level style (palette, line_type, line_width), distinct from
	// the per-formula style (palette, palette_index) in widgetFormulaFields.
	if styleAttrs := getObjAttrs(attrs, "style"); styleAttrs != nil {
		style := map[string]interface{}{}
		if v, ok := getStrAttr(styleAttrs, "palette"); ok && v != "" {
			style["palette"] = v
		}
		if v, ok := getStrAttr(styleAttrs, "line_type"); ok && v != "" {
			style["line_type"] = v
		}
		if v, ok := getStrAttr(styleAttrs, "line_width"); ok && v != "" {
			style["line_width"] = v
		}
		if len(style) > 0 {
			result["style"] = style
		}
	}

	// response_format is always "timeseries" for formula/query requests
	if len(formulaElems) > 0 || len(queryElems) > 0 {
		result["response_format"] = "timeseries"
	}

	return result
}

// buildQueryFromAttrs dispatches a query attrs map to the appropriate query builder.
func buildQueryFromAttrs(attrs map[string]attr.Value) map[string]interface{} {
	if metricAttrs := getObjAttrs(attrs, "metric_query"); metricAttrs != nil {
		return buildMetricQueryJSON(metricAttrs)
	}
	if eventAttrs := getObjAttrs(attrs, "event_query"); eventAttrs != nil {
		return buildEventQueryJSON(eventAttrs)
	}
	if procAttrs := getObjAttrs(attrs, "process_query"); procAttrs != nil {
		return buildFormulaProcessQueryJSON(procAttrs)
	}
	if sloAttrs := getObjAttrs(attrs, "slo_query"); sloAttrs != nil {
		return buildSLOQueryJSON(sloAttrs)
	}
	if ccAttrs := getObjAttrs(attrs, "cloud_cost_query"); ccAttrs != nil {
		return buildCloudCostQueryJSON(ccAttrs)
	}
	return nil
}

// isFormulaCapableWidget returns true for widget types that support
// formula/query-style requests (response_format: "scalar" or "timeseries").
func isFormulaCapableWidget(jsonType string) bool {
	switch jsonType {
	case "timeseries", "heatmap", "change", "query_value", "toplist", "sunburst", "geomap", "treemap":
		return true
	}
	return false
}

// isTimeseriesFormulaWidget returns true for widgets whose formula requests
// use response_format: "timeseries" (as opposed to "scalar").
// NOTE: only "timeseries" uses buildFormulaQueryRequestJSON (with on_right_yaxis/display_type);
// "heatmap" uses buildScalarFormulaQueryRequestJSON with "timeseries" response_format.
func isTimeseriesFormulaWidget(jsonType string) bool {
	return jsonType == "timeseries"
}

// formulaResponseFormat returns the response_format string for the given widget type.
func formulaResponseFormat(jsonType string) string {
	switch jsonType {
	case "timeseries":
		return "timeseries"
	case "heatmap":
		return "timeseries"
	default:
		return "scalar"
	}
}

// buildScalarFormulaQueryRequestJSON builds the JSON for a new-style formula/query
// request for scalar-response widgets (change, query_value, sunburst, geomap, treemap, toplist).
// For timeseries, use buildFormulaQueryRequestJSON instead (response_format: "timeseries").
func buildScalarFormulaQueryRequestJSON(attrs map[string]attr.Value, widgetType string) map[string]interface{} {
	result := map[string]interface{}{}

	// change widget has increase_good and show_present even in formula requests
	if widgetType == "change" {
		increaseGood, _ := getBoolAttr(attrs, "increase_good")
		result["increase_good"] = increaseGood
		showPresent, _ := getBoolAttr(attrs, "show_present")
		result["show_present"] = showPresent
		for _, field := range []string{"change_type", "compare_to", "order_by", "order_dir"} {
			if v, ok := getStrAttr(attrs, field); ok && v != "" {
				result[field] = v
			}
		}
	}

	// formulas — use widgetFormulaFields engine for standard fields.
	// number_format is excluded from widgetFormulaFields (polymorphic structure);
	// it is not present on scalar formula requests for non-query_table widgets.
	formulaElems := getListElems(attrs, "formula")
	if len(formulaElems) > 0 {
		formulas := make([]interface{}, 0, len(formulaElems))
		for _, elem := range formulaElems {
			if fObj, ok := elem.(types.Object); ok {
				formulas = append(formulas, BuildEngineJSON(fObj.Attributes(), widgetFormulaFields))
			}
		}
		result["formulas"] = formulas
	}

	// queries (polymorphic)
	queryElems := getListElems(attrs, "query")
	if len(queryElems) > 0 {
		queries := make([]interface{}, 0, len(queryElems))
		for _, elem := range queryElems {
			if qObj, ok := elem.(types.Object); ok {
				queries = append(queries, buildQueryFromAttrs(qObj.Attributes()))
			}
		}
		result["queries"] = queries
	}

	// style (optional, present on sunburst/toplist/etc. formula requests at the request level)
	// This is request-level style (just palette for scalar widgets), distinct from
	// the per-formula style (palette, palette_index) in widgetFormulaFields.
	if styleAttrs := getObjAttrs(attrs, "style"); styleAttrs != nil {
		style := map[string]interface{}{}
		if v, ok := getStrAttr(styleAttrs, "palette"); ok && v != "" {
			style["palette"] = v
		}
		if len(style) > 0 {
			result["style"] = style
		}
	}

	// response_format depends on widget type
	if len(formulaElems) > 0 || len(queryElems) > 0 {
		result["response_format"] = formulaResponseFormat(widgetType)
	}

	return result
}

// buildMetricQueryJSON builds the JSON for a formula metric_query block.
func buildMetricQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := getStrAttr(attrs, "data_source"); ok && v != "" {
		result["data_source"] = v
	} else {
		result["data_source"] = "metrics" // default
	}
	if v, ok := getStrAttr(attrs, "query"); ok {
		result["query"] = v
	}
	if v, ok := getStrAttr(attrs, "name"); ok {
		result["name"] = v
	}
	if v, ok := getStrAttr(attrs, "aggregator"); ok && v != "" {
		result["aggregator"] = v
	}
	if v, ok := getStrAttr(attrs, "semantic_mode"); ok && v != "" {
		result["semantic_mode"] = v
	}
	// cross_org_uuids: list of one string
	if elems := getListElems(attrs, "cross_org_uuids"); len(elems) == 1 {
		if sv, ok := elems[0].(types.String); ok && !sv.IsNull() && sv.ValueString() != "" {
			result["cross_org_uuids"] = []string{sv.ValueString()}
		}
	}
	return result
}

// buildEventQueryJSON builds the JSON for a formula event_query block.
func buildEventQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := getStrAttr(attrs, "data_source"); ok {
		result["data_source"] = v
	}
	if v, ok := getStrAttr(attrs, "name"); ok {
		result["name"] = v
	}
	if v, ok := getStrAttr(attrs, "storage"); ok && v != "" {
		result["storage"] = v
	}

	// compute (required)
	if computeAttrs := getObjAttrs(attrs, "compute"); computeAttrs != nil {
		compute := map[string]interface{}{}
		if v, ok := getStrAttr(computeAttrs, "aggregation"); ok {
			compute["aggregation"] = v
		}
		if v, ok := getInt64Attr(computeAttrs, "interval"); ok && v != 0 {
			compute["interval"] = v
		}
		if v, ok := getStrAttr(computeAttrs, "metric"); ok && v != "" {
			compute["metric"] = v
		}
		result["compute"] = compute
	}

	// search (optional)
	if searchAttrs := getObjAttrs(attrs, "search"); searchAttrs != nil {
		search := map[string]interface{}{}
		if v, ok := getStrAttr(searchAttrs, "query"); ok {
			search["query"] = v
		}
		result["search"] = search
	}

	// indexes (can be empty array)
	if indexes := getStrListElems(attrs, "indexes"); indexes != nil {
		result["indexes"] = indexes
	}

	// group_by (optional)
	groupByElems := getListElems(attrs, "group_by")
	if len(groupByElems) > 0 {
		groupBys := make([]interface{}, 0, len(groupByElems))
		for _, elem := range groupByElems {
			gbObj, ok := elem.(types.Object)
			if !ok {
				continue
			}
			gbAttrs := gbObj.Attributes()
			gb := map[string]interface{}{}
			if v, ok := getStrAttr(gbAttrs, "facet"); ok {
				gb["facet"] = v
			}
			if v, ok := getInt64Attr(gbAttrs, "limit"); ok && v != 0 {
				gb["limit"] = v
			}
			if sortAttrs := getObjAttrs(gbAttrs, "sort"); sortAttrs != nil {
				sort := map[string]interface{}{}
				if v, ok := getStrAttr(sortAttrs, "aggregation"); ok {
					sort["aggregation"] = v
				}
				if v, ok := getStrAttr(sortAttrs, "metric"); ok && v != "" {
					sort["metric"] = v
				}
				if v, ok := getStrAttr(sortAttrs, "order"); ok && v != "" {
					sort["order"] = v
				}
				gb["sort"] = sort
			}
			groupBys = append(groupBys, gb)
		}
		result["group_by"] = groupBys
	}

	// cross_org_uuids
	if elems := getListElems(attrs, "cross_org_uuids"); len(elems) == 1 {
		if sv, ok := elems[0].(types.String); ok && !sv.IsNull() && sv.ValueString() != "" {
			result["cross_org_uuids"] = []string{sv.ValueString()}
		}
	}

	return result
}

// buildFormulaProcessQueryJSON builds JSON for a formula process_query block.
func buildFormulaProcessQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := getStrAttr(attrs, "data_source"); ok {
		result["data_source"] = v
	}
	if v, ok := getStrAttr(attrs, "name"); ok {
		result["name"] = v
	}
	if v, ok := getStrAttr(attrs, "metric"); ok {
		result["metric"] = v
	}
	if v, ok := getStrAttr(attrs, "text_filter"); ok && v != "" {
		result["text_filter"] = v
	}
	if v, ok := getInt64Attr(attrs, "limit"); ok && v != 0 {
		result["limit"] = v
	}
	if v, ok := getStrAttr(attrs, "sort"); ok && v != "" {
		result["sort"] = v
	}
	if isNormalizedCPU, ok := getBoolAttr(attrs, "is_normalized_cpu"); ok && isNormalizedCPU {
		result["is_normalized_cpu"] = true
	}
	if tags := getStrListElems(attrs, "tag_filters"); len(tags) > 0 {
		result["tag_filters"] = tags
	}
	// cross_org_uuids: list of one string
	if elems := getListElems(attrs, "cross_org_uuids"); len(elems) == 1 {
		if sv, ok := elems[0].(types.String); ok && !sv.IsNull() && sv.ValueString() != "" {
			result["cross_org_uuids"] = []string{sv.ValueString()}
		}
	}
	return result
}

// buildSLOQueryJSON builds JSON for a formula slo_query block.
func buildSLOQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := getStrAttr(attrs, "data_source"); ok {
		result["data_source"] = v
	}
	if v, ok := getStrAttr(attrs, "name"); ok {
		result["name"] = v
	}
	if v, ok := getStrAttr(attrs, "slo_id"); ok {
		result["slo_id"] = v
	}
	if v, ok := getStrAttr(attrs, "measure"); ok {
		result["measure"] = v
	}
	if v, ok := getStrAttr(attrs, "group_mode"); ok && v != "" {
		result["group_mode"] = v
	}
	if v, ok := getStrAttr(attrs, "slo_query_type"); ok && v != "" {
		result["slo_query_type"] = v
	}
	if v, ok := getStrAttr(attrs, "additional_query_filters"); ok && v != "" {
		result["additional_query_filters"] = v
	}
	if elems := getListElems(attrs, "cross_org_uuids"); len(elems) == 1 {
		if sv, ok := elems[0].(types.String); ok && !sv.IsNull() && sv.ValueString() != "" {
			result["cross_org_uuids"] = []string{sv.ValueString()}
		}
	}
	return result
}

// buildCloudCostQueryJSON builds JSON for a formula cloud_cost_query block.
func buildCloudCostQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := getStrAttr(attrs, "data_source"); ok {
		result["data_source"] = v
	}
	if v, ok := getStrAttr(attrs, "name"); ok {
		result["name"] = v
	}
	if v, ok := getStrAttr(attrs, "query"); ok {
		result["query"] = v
	}
	if v, ok := getStrAttr(attrs, "aggregator"); ok && v != "" {
		result["aggregator"] = v
	}
	if elems := getListElems(attrs, "cross_org_uuids"); len(elems) == 1 {
		if sv, ok := elems[0].(types.String); ok && !sv.IsNull() && sv.ValueString() != "" {
			result["cross_org_uuids"] = []string{sv.ValueString()}
		}
	}
	return result
}

// flattenWidgetEngineJSON converts the JSON widget object into HCL state.
// Returns nil if the widget type is not recognized (unimplemented widget types pass through).
func flattenWidgetEngineJSON(widgetData map[string]interface{}) map[string]interface{} {
	def, ok := widgetData["definition"].(map[string]interface{})
	if !ok {
		return nil
	}
	widgetType, _ := def["type"].(string)

	for _, spec := range allWidgetSpecs {
		if spec.JSONType != widgetType {
			continue
		}
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defState := FlattenEngineJSON(allFields, def)

		// Per-widget post-processing: formula/query blocks, type-specific fixups,
		// and Batch C complex widget handling — all consolidated in one hook.
		flattenWidgetPostProcess(spec, def, defState)

		return map[string]interface{}{
			spec.HCLKey: []interface{}{defState},
		}
	}
	return nil
}

// flattenFormulaQueryRequestJSON converts a JSON formula/query request into HCL state.
func flattenFormulaQueryRequestJSON(req map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	// on_right_yaxis
	if v, ok := req["on_right_yaxis"].(bool); ok {
		result["on_right_yaxis"] = v
	} else {
		result["on_right_yaxis"] = false
	}

	// display_type
	if v, ok := req["display_type"].(string); ok && v != "" {
		result["display_type"] = v
	}

	// formulas — use widgetFormulaFields engine for standard fields.
	// number_format is not present on timeseries formula requests.
	if formulas, ok := req["formulas"].([]interface{}); ok {
		flatFormulas := make([]interface{}, len(formulas))
		for fi, f := range formulas {
			fm, ok := f.(map[string]interface{})
			if !ok {
				flatFormulas[fi] = map[string]interface{}{}
				continue
			}
			flatFormulas[fi] = FlattenEngineJSON(widgetFormulaFields, fm)
		}
		result["formula"] = flatFormulas
	}

	// queries
	if queries, ok := req["queries"].([]interface{}); ok {
		flatQueries := make([]interface{}, len(queries))
		for qi, q := range queries {
			qm, ok := q.(map[string]interface{})
			if !ok {
				flatQueries[qi] = map[string]interface{}{}
				continue
			}
			flatQueries[qi] = flattenFormulaQueryJSON(qm)
		}
		result["query"] = flatQueries
	}

	// style (present on timeseries formula requests at the request level)
	// Includes palette, line_type, line_width (timeseries-specific style fields).
	// This is request-level style, distinct from per-formula style in widgetFormulaFields.
	if style, ok := req["style"].(map[string]interface{}); ok {
		styleState := map[string]interface{}{}
		if v, ok := style["palette"].(string); ok && v != "" {
			styleState["palette"] = v
		}
		if v, ok := style["line_type"].(string); ok && v != "" {
			styleState["line_type"] = v
		}
		if v, ok := style["line_width"].(string); ok && v != "" {
			styleState["line_width"] = v
		}
		if len(styleState) > 0 {
			result["style"] = []interface{}{styleState}
		}
	}

	return result
}

// flattenFormulaQueryJSON converts a single query JSON object into the HCL query block state.
// Returns a map with one key: the query type name (e.g., "metric_query").
func flattenFormulaQueryJSON(q map[string]interface{}) map[string]interface{} {
	dataSource, _ := q["data_source"].(string)

	switch dataSource {
	case "metrics":
		return map[string]interface{}{
			"metric_query": []interface{}{flattenMetricQueryJSON(q)},
		}
	case "logs", "spans", "profiling", "audit", "rum":
		return map[string]interface{}{
			"event_query": []interface{}{flattenEventQueryJSON(q)},
		}
	case "process":
		return map[string]interface{}{
			"process_query": []interface{}{flattenFormulaProcessQueryFlatJSON(q)},
		}
	case "slo":
		return map[string]interface{}{
			"slo_query": []interface{}{flattenSLOQueryJSON(q)},
		}
	case "cloud_cost":
		return map[string]interface{}{
			"cloud_cost_query": []interface{}{flattenCloudCostQueryJSON(q)},
		}
	default:
		// Unknown query type — check for event_query markers
		if _, hasCompute := q["compute"]; hasCompute {
			return map[string]interface{}{
				"event_query": []interface{}{flattenEventQueryJSON(q)},
			}
		}
		// Try metric_query structure
		if _, hasQuery := q["query"]; hasQuery {
			return map[string]interface{}{
				"metric_query": []interface{}{flattenMetricQueryJSON(q)},
			}
		}
		return map[string]interface{}{}
	}
}

// buildToplistStyleDisplayJSON post-processes the toplist style.display block to
// inject "legend": "automatic" when type == "stacked" (SDK default behavior).
// The HCL schema only has "type", but JSON requires "legend" for the stacked variant.
// display is emitted as a single object by TypeBlock (not array).
func buildToplistStyleDisplayJSON(defJSON map[string]interface{}) {
	style, ok := defJSON["style"].(map[string]interface{})
	if !ok {
		return
	}
	displayObj, ok := style["display"].(map[string]interface{})
	if !ok {
		return
	}
	if t, ok := displayObj["type"].(string); ok && t == "stacked" {
		// Inject "legend": "automatic" for stacked type (SDK default)
		displayObj["legend"] = "automatic"
	}
}

// buildScatterplotTableJSON builds the "table" object within scatterplot "requests"
// for formula/query style scatterplot_table requests.
// defAttrs is the attributes map of the scatterplot definition block.
func buildScatterplotTableJSON(defAttrs map[string]attr.Value, defJSON map[string]interface{}) {
	requestElems := getListElems(defAttrs, "request")
	if len(requestElems) == 0 {
		return
	}
	reqObj, ok := requestElems[0].(types.Object)
	if !ok {
		return
	}
	reqAttrs := reqObj.Attributes()
	tableAttrs := getObjAttrs(reqAttrs, "scatterplot_table")
	if tableAttrs == nil {
		return
	}

	tableObj := map[string]interface{}{}

	// formulas (ScatterplotWidgetFormula: formula_expression → formula, plus dimension and alias)
	formulaElems := getListElems(tableAttrs, "formula")
	if len(formulaElems) > 0 {
		formulas := make([]interface{}, 0, len(formulaElems))
		for _, fElem := range formulaElems {
			fObj, ok := fElem.(types.Object)
			if !ok {
				continue
			}
			fAttrs := fObj.Attributes()
			formula := map[string]interface{}{}
			if expr, ok := getStrAttr(fAttrs, "formula_expression"); ok {
				formula["formula"] = expr
			}
			if dim, ok := getStrAttr(fAttrs, "dimension"); ok {
				formula["dimension"] = dim
			}
			if alias, ok := getStrAttr(fAttrs, "alias"); ok {
				formula["alias"] = alias
			}
			formulas = append(formulas, formula)
		}
		tableObj["formulas"] = formulas
	}

	// queries
	queryElems := getListElems(tableAttrs, "query")
	if len(queryElems) > 0 {
		queries := make([]interface{}, 0, len(queryElems))
		for _, qElem := range queryElems {
			qObj, ok := qElem.(types.Object)
			if !ok {
				continue
			}
			queries = append(queries, buildQueryFromAttrs(qObj.Attributes()))
		}
		tableObj["queries"] = queries
	}

	if len(formulaElems) > 0 || len(queryElems) > 0 {
		tableObj["response_format"] = "scalar"
	}

	if len(tableObj) > 0 {
		// Inject "table" into the "requests" object, creating it if necessary
		requestsObj, ok := defJSON["requests"].(map[string]interface{})
		if !ok {
			requestsObj = map[string]interface{}{}
			defJSON["requests"] = requestsObj
		}
		requestsObj["table"] = tableObj
	}
}

// flattenScatterplotTableState reconstructs the scatterplot_table HCL block
// from the JSON "requests.table" object.
func flattenScatterplotTableState(def map[string]interface{}, defState map[string]interface{}) {
	requests, ok := def["requests"].(map[string]interface{})
	if !ok {
		return
	}
	tableObj, ok := requests["table"].(map[string]interface{})
	if !ok {
		return
	}

	tableState := map[string]interface{}{}

	// formulas
	if formulas, ok := tableObj["formulas"].([]interface{}); ok {
		flatFormulas := make([]interface{}, len(formulas))
		for fi, f := range formulas {
			fm, ok := f.(map[string]interface{})
			if !ok {
				flatFormulas[fi] = map[string]interface{}{}
				continue
			}
			flatF := map[string]interface{}{}
			if v, ok := fm["formula"].(string); ok {
				flatF["formula_expression"] = v
			}
			if v, ok := fm["dimension"].(string); ok && v != "" {
				flatF["dimension"] = v
			}
			if v, ok := fm["alias"].(string); ok && v != "" {
				flatF["alias"] = v
			}
			flatFormulas[fi] = flatF
		}
		tableState["formula"] = flatFormulas
	}

	// queries
	if queries, ok := tableObj["queries"].([]interface{}); ok {
		flatQueries := make([]interface{}, len(queries))
		for qi, q := range queries {
			qm, ok := q.(map[string]interface{})
			if !ok {
				flatQueries[qi] = map[string]interface{}{}
				continue
			}
			flatQueries[qi] = flattenFormulaQueryJSON(qm)
		}
		tableState["query"] = flatQueries
	}

	if len(tableState) > 0 {
		// Inject scatterplot_table into the request block
		if reqState, ok := defState["request"].([]interface{}); ok && len(reqState) > 0 {
			if reqMap, ok := reqState[0].(map[string]interface{}); ok {
				reqMap["scatterplot_table"] = []interface{}{tableState}
			}
		}
	}
}

// flattenSunburstLegendState disambiguates the sunburst legend JSON object
// (which maps to either legend_inline or legend_table in HCL based on legend type).
// The FlattenEngineJSON engine would incorrectly set both; this function
// fixes the state by removing the wrong one.
func flattenSunburstLegendState(def map[string]interface{}, defState map[string]interface{}) {
	legend, ok := def["legend"].(map[string]interface{})
	if !ok {
		// No legend in response — remove both blocks from state
		delete(defState, "legend_inline")
		delete(defState, "legend_table")
		return
	}
	legendType, _ := legend["type"].(string)
	// SunburstWidgetLegendTableType: "table" or "none"
	// SunburstWidgetLegendInlineAutomaticType: "inline" or "automatic"
	switch legendType {
	case "table", "none":
		// legend_table wins — remove legend_inline
		delete(defState, "legend_inline")
	case "inline", "automatic":
		// legend_inline wins — remove legend_table
		delete(defState, "legend_table")
	default:
		delete(defState, "legend_inline")
		delete(defState, "legend_table")
	}
}

// flattenScalarFormulaQueryRequestJSON flattens a scalar formula/query request
// (for change, query_value, toplist, sunburst, geomap, treemap widgets).
func flattenScalarFormulaQueryRequestJSON(req map[string]interface{}, widgetType string) map[string]interface{} {
	result := map[string]interface{}{}

	// change widget has widget-specific fields in the request
	if widgetType == "change" {
		if v, ok := req["increase_good"].(bool); ok {
			result["increase_good"] = v
		} else {
			result["increase_good"] = false
		}
		if v, ok := req["show_present"].(bool); ok {
			result["show_present"] = v
		} else {
			result["show_present"] = false
		}
		for _, field := range []string{"change_type", "compare_to", "order_by", "order_dir"} {
			if v, ok := req[field].(string); ok && v != "" {
				result[field] = v
			}
		}
	}

	// formulas — use widgetFormulaFields engine for standard fields.
	// number_format is not present on scalar formula requests for non-query_table widgets.
	if formulas, ok := req["formulas"].([]interface{}); ok {
		flatFormulas := make([]interface{}, len(formulas))
		for fi, f := range formulas {
			fm, ok := f.(map[string]interface{})
			if !ok {
				flatFormulas[fi] = map[string]interface{}{}
				continue
			}
			flatFormulas[fi] = FlattenEngineJSON(widgetFormulaFields, fm)
		}
		result["formula"] = flatFormulas
	}

	// queries
	if queries, ok := req["queries"].([]interface{}); ok {
		flatQueries := make([]interface{}, len(queries))
		for qi, q := range queries {
			qm, ok := q.(map[string]interface{})
			if !ok {
				flatQueries[qi] = map[string]interface{}{}
				continue
			}
			flatQueries[qi] = flattenFormulaQueryJSON(qm)
		}
		result["query"] = flatQueries
	}

	// style (optional, present on sunburst/toplist/etc. formula requests at the request level)
	// This is request-level style (just palette for scalar widgets), distinct from
	// the per-formula style (palette, palette_index) in widgetFormulaFields.
	if style, ok := req["style"].(map[string]interface{}); ok {
		styleState := map[string]interface{}{}
		if v, ok := style["palette"].(string); ok && v != "" {
			styleState["palette"] = v
		}
		if len(styleState) > 0 {
			result["style"] = []interface{}{styleState}
		}
	}

	return result
}

// flattenOldStyleRequestJSON flattens an old-style (non-formula) request
// for non-timeseries formula-capable widgets. It finds the request field spec
// from the widget spec and uses FlattenEngineJSON.
func flattenOldStyleRequestJSON(req map[string]interface{}, spec WidgetSpec) map[string]interface{} {
	// Find the request FieldSpec within the widget spec
	for _, f := range spec.Fields {
		if f.HCLKey == "request" {
			return FlattenEngineJSON(f.Children, req)
		}
	}
	return map[string]interface{}{}
}

func flattenMetricQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := q["data_source"].(string); ok {
		result["data_source"] = v
	}
	if v, ok := q["query"].(string); ok {
		result["query"] = v
	}
	if v, ok := q["name"].(string); ok {
		result["name"] = v
	}
	if v, ok := q["aggregator"].(string); ok && v != "" {
		result["aggregator"] = v
	}
	if v, ok := q["semantic_mode"].(string); ok && v != "" {
		result["semantic_mode"] = v
	}
	if uuids, ok := q["cross_org_uuids"].([]interface{}); ok && len(uuids) > 0 {
		result["cross_org_uuids"] = uuids
	}
	return result
}

func flattenEventQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := q["data_source"].(string); ok {
		result["data_source"] = v
	}
	if v, ok := q["name"].(string); ok {
		result["name"] = v
	}
	if v, ok := q["storage"].(string); ok && v != "" {
		result["storage"] = v
	}
	// compute
	if compute, ok := q["compute"].(map[string]interface{}); ok {
		c := map[string]interface{}{}
		if v, ok := compute["aggregation"].(string); ok {
			c["aggregation"] = v
		}
		if v, ok := compute["interval"]; ok {
			switch iv := v.(type) {
			case float64:
				c["interval"] = int(iv)
			case int:
				c["interval"] = iv
			}
		}
		if v, ok := compute["metric"].(string); ok && v != "" {
			c["metric"] = v
		}
		result["compute"] = []interface{}{c}
	}
	// search
	if search, ok := q["search"].(map[string]interface{}); ok {
		s := map[string]interface{}{}
		if v, ok := search["query"].(string); ok {
			s["query"] = v
		}
		result["search"] = []interface{}{s}
	}
	// indexes
	if indexes, ok := q["indexes"].([]interface{}); ok {
		idxStrs := make([]string, len(indexes))
		for i, idx := range indexes {
			idxStrs[i] = fmt.Sprintf("%v", idx)
		}
		result["indexes"] = idxStrs
	}
	// group_by
	if groupBys, ok := q["group_by"].([]interface{}); ok {
		flatGBs := make([]interface{}, len(groupBys))
		for gi, gb := range groupBys {
			gbMap, ok := gb.(map[string]interface{})
			if !ok {
				flatGBs[gi] = map[string]interface{}{}
				continue
			}
			flatGB := map[string]interface{}{}
			if v, ok := gbMap["facet"].(string); ok {
				flatGB["facet"] = v
			}
			if v, ok := gbMap["limit"]; ok {
				switch lv := v.(type) {
				case float64:
					flatGB["limit"] = int(lv)
				case int:
					flatGB["limit"] = lv
				}
			}
			if sort, ok := gbMap["sort"].(map[string]interface{}); ok {
				s := map[string]interface{}{}
				if v, ok := sort["aggregation"].(string); ok {
					s["aggregation"] = v
				}
				if v, ok := sort["metric"].(string); ok && v != "" {
					s["metric"] = v
				}
				if v, ok := sort["order"].(string); ok && v != "" {
					s["order"] = v
				}
				flatGB["sort"] = []interface{}{s}
			}
			flatGBs[gi] = flatGB
		}
		result["group_by"] = flatGBs
	}
	if uuids, ok := q["cross_org_uuids"].([]interface{}); ok && len(uuids) > 0 {
		result["cross_org_uuids"] = uuids
	}
	return result
}

func flattenFormulaProcessQueryFlatJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := q["data_source"].(string); ok {
		result["data_source"] = v
	}
	if v, ok := q["name"].(string); ok {
		result["name"] = v
	}
	if v, ok := q["metric"].(string); ok {
		result["metric"] = v
	}
	if v, ok := q["text_filter"].(string); ok && v != "" {
		result["text_filter"] = v
	}
	if v, ok := q["limit"]; ok {
		switch lv := v.(type) {
		case float64:
			result["limit"] = int(lv)
		case int:
			result["limit"] = lv
		}
	}
	if v, ok := q["sort"].(string); ok && v != "" {
		result["sort"] = v
	}
	if v, ok := q["is_normalized_cpu"].(bool); ok {
		result["is_normalized_cpu"] = v
	}
	if tags, ok := q["tag_filters"].([]interface{}); ok && len(tags) > 0 {
		tagStrs := make([]string, len(tags))
		for i, t := range tags {
			tagStrs[i] = fmt.Sprintf("%v", t)
		}
		result["tag_filters"] = tagStrs
	}
	if uuids, ok := q["cross_org_uuids"].([]interface{}); ok && len(uuids) > 0 {
		result["cross_org_uuids"] = uuids
	}
	return result
}

func flattenSLOQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "name", "slo_id", "measure", "group_mode", "slo_query_type", "additional_query_filters"} {
		if v, ok := q[field].(string); ok && v != "" {
			result[field] = v
		}
	}
	if uuids, ok := q["cross_org_uuids"].([]interface{}); ok && len(uuids) > 0 {
		result["cross_org_uuids"] = uuids
	}
	return result
}

func flattenCloudCostQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "name", "query", "aggregator"} {
		if v, ok := q[field].(string); ok && v != "" {
			result[field] = v
		}
	}
	if uuids, ok := q["cross_org_uuids"].([]interface{}); ok && len(uuids) > 0 {
		result["cross_org_uuids"] = uuids
	}
	return result
}

// ============================================================
// Widget Post-Processing Hooks
// ============================================================
// buildWidgetPostProcess and flattenWidgetPostProcess are the unified per-widget
// post-processing hooks called by buildWidgetEngineJSON / flattenWidgetEngineJSON.
// All per-widget special-cases live here so the engine calls exactly one hook per widget.
//
// Also contains complex widget helper functions for widgets that require custom
// build/flatten logic beyond what the FieldSpec engine provides:
//   - Query table: formula requests, number_format, text_formats (2D array)
//   - Split graph: source_widget_definition, static_splits (2D array)
//   - Group: recursive nested widget list
//   - Scatterplot: scatterplot_table formula/query injection
//   - Sunburst: polymorphic legend disambiguation

// buildWidgetPostProcess runs all per-widget post-processing in the build direction.
// It covers:
//   - Formula/query blocks (formula-capable widgets)
//   - Scatterplot table
//   - Toplist style display fix-up
//   - Treemap color_by injection
//   - Query table requests override
//   - Split graph source widget and static_splits
//   - Group nested widgets
func buildWidgetPostProcess(defAttrs map[string]attr.Value, spec WidgetSpec, defJSON map[string]interface{}) {
	// ---- Formula/query blocks ----
	if isFormulaCapableWidget(spec.JSONType) {
		requestElems := getListElems(defAttrs, "request")
		if len(requestElems) > 0 {
			requests := make([]interface{}, 0, len(requestElems))
			for ri, reqElem := range requestElems {
				reqObj, ok := reqElem.(types.Object)
				if !ok {
					requests = append(requests, map[string]interface{}{})
					continue
				}
				reqAttrs := reqObj.Attributes()
				formulaCount := len(getListElems(reqAttrs, "formula"))
				queryCount := len(getListElems(reqAttrs, "query"))
				if formulaCount > 0 || queryCount > 0 {
					// New-style formula/query request
					if isTimeseriesFormulaWidget(spec.JSONType) {
						requests = append(requests, buildFormulaQueryRequestJSON(reqAttrs))
					} else {
						requests = append(requests, buildScalarFormulaQueryRequestJSON(reqAttrs, spec.JSONType))
					}
				} else {
					// Old-style request (already built by FieldSpec engine)
					if existingReqs, ok := defJSON["requests"].([]interface{}); ok && ri < len(existingReqs) {
						requests = append(requests, existingReqs[ri])
					}
				}
			}
			defJSON["requests"] = requests
		}
	}

	// ---- Scatterplot table ----
	if spec.JSONType == "scatterplot" {
		buildScatterplotTableJSON(defAttrs, defJSON)
	}

	// ---- Toplist style display ----
	if spec.JSONType == "toplist" {
		buildToplistStyleDisplayJSON(defJSON)
	}

	// ---- Treemap color_by ----
	if spec.JSONType == "treemap" {
		defJSON["color_by"] = "user"
	}

	// ---- Query table requests override ----
	if spec.JSONType == "query_table" {
		requests := buildQueryTableRequestsJSON(defAttrs)
		defJSON["requests"] = requests
	}

	// ---- Split graph source widget + static_splits ----
	if spec.JSONType == "split_group" {
		srcDefJSON := buildSplitGraphSourceWidgetJSON(defAttrs)
		if srcDefJSON != nil {
			defJSON["source_widget_definition"] = srcDefJSON
		}
		if splitConfigMap, ok := defJSON["split_config"].(map[string]interface{}); ok {
			if splitConfigAttrs := getObjAttrs(defAttrs, "split_config"); splitConfigAttrs != nil {
				staticSplits := buildSplitConfigStaticSplitsJSON(splitConfigAttrs)
				if len(staticSplits) > 0 {
					splitConfigMap["static_splits"] = staticSplits
				}
			}
		}
	}

	// ---- Group nested widgets ----
	if spec.JSONType == "group" {
		widgets := buildGroupWidgetsJSON(defAttrs)
		defJSON["widgets"] = widgets
	}
}

// flattenWidgetPostProcess runs all per-widget post-processing in the flatten direction.
// It covers:
//   - Formula/query request flattening (formula-capable widgets)
//   - Sunburst legend disambiguation
//   - Scatterplot table flattening
//   - Query table requests override
//   - Split graph source widget and static_splits
//   - Group nested widgets
func flattenWidgetPostProcess(spec WidgetSpec, def map[string]interface{}, defState map[string]interface{}) {
	// ---- Formula/query request flattening ----
	if isFormulaCapableWidget(spec.JSONType) {
		if requests, ok := def["requests"].([]interface{}); ok {
			flatRequests := make([]interface{}, len(requests))
			for ri, req := range requests {
				reqMap, ok := req.(map[string]interface{})
				if !ok {
					flatRequests[ri] = map[string]interface{}{}
					continue
				}
				_, hasFormulas := reqMap["formulas"]
				_, hasQueries := reqMap["queries"]
				if hasFormulas || hasQueries {
					if isTimeseriesFormulaWidget(spec.JSONType) {
						// Timeseries formula/query request (response_format: timeseries)
						flatRequests[ri] = flattenFormulaQueryRequestJSON(reqMap)
					} else {
						// Scalar formula/query request (change, query_value, toplist, etc.)
						flatRequests[ri] = flattenScalarFormulaQueryRequestJSON(reqMap, spec.JSONType)
					}
				} else {
					// Old-style request — already flattened by FieldSpec engine
					if existingRequests, ok := defState["request"].([]interface{}); ok && ri < len(existingRequests) {
						flatRequests[ri] = existingRequests[ri]
					} else {
						if spec.JSONType == "timeseries" {
							flatRequests[ri] = FlattenEngineJSON(timeseriesWidgetRequestFields, reqMap)
						} else {
							flatRequests[ri] = flattenOldStyleRequestJSON(reqMap, spec)
						}
					}
				}
			}
			defState["request"] = flatRequests
		}
	}

	// ---- Sunburst legend disambiguation ----
	if spec.JSONType == "sunburst" {
		flattenSunburstLegendState(def, defState)
	}

	// ---- Scatterplot table ----
	if spec.JSONType == "scatterplot" {
		flattenScatterplotTableState(def, defState)
	}

	// ---- Query table requests override ----
	if spec.JSONType == "query_table" {
		if requests, ok := def["requests"].([]interface{}); ok {
			flatRequests := make([]interface{}, len(requests))
			for ri, req := range requests {
				reqMap, ok := req.(map[string]interface{})
				if !ok {
					flatRequests[ri] = map[string]interface{}{}
					continue
				}
				flatRequests[ri] = flattenQueryTableRequestJSON(reqMap)
			}
			defState["request"] = flatRequests
		}
	}

	// ---- Split graph source widget + static_splits ----
	if spec.JSONType == "split_group" {
		if srcDef, ok := def["source_widget_definition"].(map[string]interface{}); ok {
			flatSrc := flattenSplitGraphSourceWidgetJSON(srcDef)
			if flatSrc != nil {
				defState["source_widget_definition"] = []interface{}{flatSrc}
			}
		}
		if splitConfigArr, ok := defState["split_config"].([]interface{}); ok && len(splitConfigArr) > 0 {
			if splitConfigMap, ok := splitConfigArr[0].(map[string]interface{}); ok {
				if rawSplitConfig, ok := def["split_config"].(map[string]interface{}); ok {
					if staticSplits, ok := rawSplitConfig["static_splits"].([]interface{}); ok && len(staticSplits) > 0 {
						splitConfigMap["static_splits"] = flattenSplitConfigStaticSplitsJSON(staticSplits)
					}
				}
			}
		}
	}

	// ---- Group nested widgets ----
	if spec.JSONType == "group" {
		if widgets, ok := def["widgets"].([]interface{}); ok {
			defState["widget"] = flattenGroupWidgetsJSON(widgets)
		}
	}
}

// ============================================================
// Complex Widget Helpers
// ============================================================

// buildQueryTableRequestsJSON builds the "requests" array for query_table.
// It handles both old-style and formula-style requests.
func buildQueryTableRequestsJSON(defAttrs map[string]attr.Value) []interface{} {
	requestElems := getListElems(defAttrs, "request")
	if len(requestElems) == 0 {
		return []interface{}{}
	}
	requests := make([]interface{}, 0, len(requestElems))
	for _, reqElem := range requestElems {
		reqObj, ok := reqElem.(types.Object)
		if !ok {
			requests = append(requests, map[string]interface{}{})
			continue
		}
		reqAttrs := reqObj.Attributes()
		formulaCount := len(getListElems(reqAttrs, "formula"))
		queryCount := len(getListElems(reqAttrs, "query"))
		if formulaCount > 0 || queryCount > 0 {
			// Formula/query style request
			requests = append(requests, buildQueryTableFormulaRequestJSON(reqAttrs))
		} else {
			// Old-style request — build via FieldSpec engine
			req := BuildEngineJSON(reqAttrs, queryTableOldRequestFields)
			// text_formats needs custom handling (it's a 2D array)
			buildQueryTableTextFormatsJSON(reqAttrs, req)
			requests = append(requests, req)
		}
	}
	return requests
}

// buildQueryTableFormulaRequestJSON builds a formula/query request for query_table.
// Unlike the generic scalar formula handler, query_table formulas can have
// cell_display_mode, cell_display_mode_options, conditional_formats, number_format
// per formula.
func buildQueryTableFormulaRequestJSON(reqAttrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}

	// formulas
	formulaElems := getListElems(reqAttrs, "formula")
	if len(formulaElems) > 0 {
		formulas := make([]interface{}, 0, len(formulaElems))
		for _, fElem := range formulaElems {
			fObj, ok := fElem.(types.Object)
			if !ok {
				continue
			}
			fAttrs := fObj.Attributes()
			formula := map[string]interface{}{}

			// formula_expression → formula (JSON key)
			if expr, ok := getStrAttr(fAttrs, "formula_expression"); ok {
				formula["formula"] = expr
			}
			if alias, ok := getStrAttr(fAttrs, "alias"); ok {
				formula["alias"] = alias
			}
			// limit
			if limitAttrs := getObjAttrs(fAttrs, "limit"); limitAttrs != nil {
				limitObj := map[string]interface{}{}
				if count, ok := getInt64Attr(limitAttrs, "count"); ok {
					limitObj["count"] = count
				}
				if order, ok := getStrAttr(limitAttrs, "order"); ok {
					limitObj["order"] = order
				}
				if len(limitObj) > 0 {
					formula["limit"] = limitObj
				}
			}
			// cell_display_mode (TypeString on formula)
			if v, ok := getStrAttr(fAttrs, "cell_display_mode"); ok && v != "" {
				formula["cell_display_mode"] = v
			}
			// cell_display_mode_options (TypeBlock MaxItems:1)
			if cdmoAttrs := getObjAttrs(fAttrs, "cell_display_mode_options"); cdmoAttrs != nil {
				opts := map[string]interface{}{}
				if v, ok := getStrAttr(cdmoAttrs, "trend_type"); ok {
					opts["trend_type"] = v
				}
				if v, ok := getStrAttr(cdmoAttrs, "y_scale"); ok {
					opts["y_scale"] = v
				}
				if len(opts) > 0 {
					formula["cell_display_mode_options"] = opts
				}
			}
			// conditional_formats on the formula
			cfElems := getListElems(fAttrs, "conditional_formats")
			if len(cfElems) > 0 {
				cfs := make([]interface{}, 0, len(cfElems))
				for _, cfElem := range cfElems {
					if cfObj, ok := cfElem.(types.Object); ok {
						cfs = append(cfs, BuildEngineJSON(cfObj.Attributes(), widgetConditionalFormatFields))
					}
				}
				formula["conditional_formats"] = cfs
			}
			// number_format (TypeBlock MaxItems:1)
			if nfAttrs := getObjAttrs(fAttrs, "number_format"); nfAttrs != nil {
				nf := buildQueryTableNumberFormatJSON(nfAttrs)
				if len(nf) > 0 {
					formula["number_format"] = nf
				}
			}
			formulas = append(formulas, formula)
		}
		result["formulas"] = formulas
	}

	// queries
	queryElems := getListElems(reqAttrs, "query")
	if len(queryElems) > 0 {
		queries := make([]interface{}, 0, len(queryElems))
		for _, qElem := range queryElems {
			qObj, ok := qElem.(types.Object)
			if !ok {
				continue
			}
			qAttrs := qObj.Attributes()
			var queryJSON map[string]interface{}
			if metricAttrs := getObjAttrs(qAttrs, "metric_query"); metricAttrs != nil {
				queryJSON = buildMetricQueryJSON(metricAttrs)
			} else if eventAttrs := getObjAttrs(qAttrs, "event_query"); eventAttrs != nil {
				queryJSON = buildEventQueryJSON(eventAttrs)
			} else if procAttrs := getObjAttrs(qAttrs, "process_query"); procAttrs != nil {
				queryJSON = buildFormulaProcessQueryJSON(procAttrs)
			} else if apmDepAttrs := getObjAttrs(qAttrs, "apm_dependency_stats_query"); apmDepAttrs != nil {
				queryJSON = buildApmDependencyStatsQueryJSON(apmDepAttrs)
			} else if apmResAttrs := getObjAttrs(qAttrs, "apm_resource_stats_query"); apmResAttrs != nil {
				queryJSON = buildApmResourceStatsQueryJSON(apmResAttrs)
			} else if sloAttrs := getObjAttrs(qAttrs, "slo_query"); sloAttrs != nil {
				queryJSON = buildSLOQueryJSON(sloAttrs)
			} else if ccAttrs := getObjAttrs(qAttrs, "cloud_cost_query"); ccAttrs != nil {
				queryJSON = buildCloudCostQueryJSON(ccAttrs)
			}
			queries = append(queries, queryJSON)
		}
		result["queries"] = queries
	}

	// response_format for query_table formulas is always "scalar"
	if len(formulaElems) > 0 || len(queryElems) > 0 {
		result["response_format"] = "scalar"
	}

	return result
}

// buildQueryTableNumberFormatJSON builds the number_format JSON object for a formula.
// The HCL structure has unit { canonical { unit_name, per_unit_name } } or unit { custom { label } }.
// The JSON structure has unit { type, unit_name, per_unit_name } or unit { type, label }.
func buildQueryTableNumberFormatJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	// unit block (TypeList MaxItems:1)
	if unitAttrs := getObjAttrs(attrs, "unit"); unitAttrs != nil {
		unit := map[string]interface{}{}
		// canonical sub-block
		if canonAttrs := getObjAttrs(unitAttrs, "canonical"); canonAttrs != nil {
			unit["type"] = "canonical_unit"
			if v, ok := getStrAttr(canonAttrs, "unit_name"); ok {
				unit["unit_name"] = v
			}
			if v, ok := getStrAttr(canonAttrs, "per_unit_name"); ok {
				unit["per_unit_name"] = v
			}
		}
		// custom sub-block
		if customAttrs := getObjAttrs(unitAttrs, "custom"); customAttrs != nil {
			unit["type"] = "custom_unit_label"
			if v, ok := getStrAttr(customAttrs, "label"); ok {
				unit["label"] = v
			}
		}
		if len(unit) > 0 {
			result["unit"] = unit
		}
	}
	// unit_scale block
	if unitScaleAttrs := getObjAttrs(attrs, "unit_scale"); unitScaleAttrs != nil {
		unitScale := map[string]interface{}{}
		if v, ok := getStrAttr(unitScaleAttrs, "unit_name"); ok {
			unitScale["unit_name"] = v
		}
		if len(unitScale) > 0 {
			result["unit_scale"] = unitScale
		}
	}
	return result
}

// buildQueryTableTextFormatsJSON handles the text_formats 2D array for query_table
// old-style requests. It reads text_formats from the HCL attrs and injects them into
// the request JSON.
func buildQueryTableTextFormatsJSON(reqAttrs map[string]attr.Value, req map[string]interface{}) {
	tfElems := getListElems(reqAttrs, "text_formats")
	if len(tfElems) == 0 {
		return
	}
	textFormats := make([]interface{}, 0, len(tfElems))
	for _, tfElem := range tfElems {
		tfObj, ok := tfElem.(types.Object)
		if !ok {
			textFormats = append(textFormats, []interface{}{})
			continue
		}
		tfAttrs := tfObj.Attributes()
		innerElems := getListElems(tfAttrs, "text_format")
		innerFormats := make([]interface{}, 0, len(innerElems))
		for _, innerElem := range innerElems {
			innerObj, ok := innerElem.(types.Object)
			if !ok {
				innerFormats = append(innerFormats, map[string]interface{}{})
				continue
			}
			innerAttrs := innerObj.Attributes()
			rule := map[string]interface{}{}
			// match (TypeBlock MaxItems:1, Required)
			if matchAttrs := getObjAttrs(innerAttrs, "match"); matchAttrs != nil {
				match := map[string]interface{}{}
				if v, ok := getStrAttr(matchAttrs, "type"); ok {
					match["type"] = v
				}
				if v, ok := getStrAttr(matchAttrs, "value"); ok {
					match["value"] = v
				}
				rule["match"] = match
			}
			// palette (optional)
			if v, ok := getStrAttr(innerAttrs, "palette"); ok && v != "" {
				rule["palette"] = v
			}
			// custom_bg_color (optional)
			if v, ok := getStrAttr(innerAttrs, "custom_bg_color"); ok && v != "" {
				rule["custom_bg_color"] = v
			}
			// custom_fg_color (optional)
			if v, ok := getStrAttr(innerAttrs, "custom_fg_color"); ok && v != "" {
				rule["custom_fg_color"] = v
			}
			// replace (TypeBlock MaxItems:1, Optional)
			if replaceAttrs := getObjAttrs(innerAttrs, "replace"); replaceAttrs != nil {
				replace := map[string]interface{}{}
				if v, ok := getStrAttr(replaceAttrs, "type"); ok {
					replace["type"] = v
				}
				if v, ok := getStrAttr(replaceAttrs, "with"); ok {
					replace["with"] = v
				}
				if v, ok := getStrAttr(replaceAttrs, "substring"); ok && v != "" {
					replace["substring"] = v
				}
				rule["replace"] = replace
			}
			innerFormats = append(innerFormats, rule)
		}
		textFormats = append(textFormats, innerFormats)
	}
	req["text_formats"] = textFormats
}

// buildApmDependencyStatsQueryJSON builds JSON for an apm_dependency_stats_query formula query.
func buildApmDependencyStatsQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "env", "name", "operation_name", "resource_name", "service", "stat"} {
		if v, ok := getStrAttr(attrs, field); ok {
			result[field] = v
		}
	}
	if v, ok := getBoolAttr(attrs, "is_upstream"); ok {
		result["is_upstream"] = v
	}
	if v, ok := getStrAttr(attrs, "primary_tag_name"); ok && v != "" {
		result["primary_tag_name"] = v
	}
	if v, ok := getStrAttr(attrs, "primary_tag_value"); ok && v != "" {
		result["primary_tag_value"] = v
	}
	return result
}

// buildApmResourceStatsQueryJSON builds JSON for an apm_resource_stats_query formula query.
func buildApmResourceStatsQueryJSON(attrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "env", "name", "service", "stat"} {
		if v, ok := getStrAttr(attrs, field); ok {
			result[field] = v
		}
	}
	if v, ok := getStrAttr(attrs, "operation_name"); ok && v != "" {
		result["operation_name"] = v
	}
	if v, ok := getStrAttr(attrs, "resource_name"); ok && v != "" {
		result["resource_name"] = v
	}
	if v, ok := getStrAttr(attrs, "primary_tag_name"); ok && v != "" {
		result["primary_tag_name"] = v
	}
	if v, ok := getStrAttr(attrs, "primary_tag_value"); ok && v != "" {
		result["primary_tag_value"] = v
	}
	// group_by: []string
	if strs := getStrListElems(attrs, "group_by"); len(strs) > 0 {
		result["group_by"] = strs
	}
	return result
}

// flattenQueryTableRequestJSON flattens a single query_table request JSON.
// Handles both formula-style (has "formulas"/"queries") and old-style requests.
func flattenQueryTableRequestJSON(req map[string]interface{}) map[string]interface{} {
	_, hasFormulas := req["formulas"]
	_, hasQueries := req["queries"]
	if hasFormulas || hasQueries {
		return flattenQueryTableFormulaRequestJSON(req)
	}
	// Old-style request
	result := FlattenEngineJSON(queryTableOldRequestFields, req)
	// text_formats (2D array) needs special handling
	if textFormats, ok := req["text_formats"].([]interface{}); ok && len(textFormats) > 0 {
		result["text_formats"] = flattenQueryTableTextFormatsJSON(textFormats)
	}
	return result
}

// flattenQueryTableFormulaRequestJSON flattens a formula/query-style query_table request.
func flattenQueryTableFormulaRequestJSON(req map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	// formulas
	if formulas, ok := req["formulas"].([]interface{}); ok {
		flatFormulas := make([]interface{}, len(formulas))
		for fi, f := range formulas {
			fm, ok := f.(map[string]interface{})
			if !ok {
				flatFormulas[fi] = map[string]interface{}{}
				continue
			}
			flatF := map[string]interface{}{}
			// JSON "formula" → HCL "formula_expression"
			if v, ok := fm["formula"].(string); ok {
				flatF["formula_expression"] = v
			}
			if v, ok := fm["alias"].(string); ok && v != "" {
				flatF["alias"] = v
			}
			// limit
			if limitMap, ok := fm["limit"].(map[string]interface{}); ok {
				limitState := map[string]interface{}{}
				if v, ok := limitMap["count"]; ok {
					switch cv := v.(type) {
					case float64:
						limitState["count"] = int(cv)
					case int64:
						limitState["count"] = int(cv)
					case int:
						limitState["count"] = cv
					}
				}
				if v, ok := limitMap["order"].(string); ok {
					limitState["order"] = v
				}
				flatF["limit"] = []interface{}{limitState}
			}
			// cell_display_mode (TypeString)
			if v, ok := fm["cell_display_mode"].(string); ok && v != "" {
				flatF["cell_display_mode"] = v
			}
			// cell_display_mode_options
			if opts, ok := fm["cell_display_mode_options"].(map[string]interface{}); ok {
				optsState := map[string]interface{}{}
				if v, ok := opts["trend_type"].(string); ok && v != "" {
					optsState["trend_type"] = v
				}
				if v, ok := opts["y_scale"].(string); ok && v != "" {
					optsState["y_scale"] = v
				}
				if len(optsState) > 0 {
					flatF["cell_display_mode_options"] = []interface{}{optsState}
				}
			}
			// conditional_formats
			if cfs, ok := fm["conditional_formats"].([]interface{}); ok && len(cfs) > 0 {
				flatCFs := make([]interface{}, len(cfs))
				for ci, cf := range cfs {
					if cfMap, ok := cf.(map[string]interface{}); ok {
						flatCFs[ci] = FlattenEngineJSON(widgetConditionalFormatFields, cfMap)
					}
				}
				flatF["conditional_formats"] = flatCFs
			}
			// number_format
			if nf, ok := fm["number_format"].(map[string]interface{}); ok {
				nfState := flattenQueryTableNumberFormatJSON(nf)
				if len(nfState) > 0 {
					flatF["number_format"] = []interface{}{nfState}
				}
			}
			flatFormulas[fi] = flatF
		}
		result["formula"] = flatFormulas
	}

	// queries
	if queries, ok := req["queries"].([]interface{}); ok {
		flatQueries := make([]interface{}, len(queries))
		for qi, q := range queries {
			qm, ok := q.(map[string]interface{})
			if !ok {
				flatQueries[qi] = map[string]interface{}{}
				continue
			}
			flatQueries[qi] = flattenQueryTableQueryJSON(qm)
		}
		result["query"] = flatQueries
	}

	return result
}

// flattenQueryTableQueryJSON converts a single query JSON object into the HCL query block state.
// Handles apm_dependency_stats and apm_resource_stats in addition to the standard query types.
func flattenQueryTableQueryJSON(q map[string]interface{}) map[string]interface{} {
	dataSource, _ := q["data_source"].(string)
	switch dataSource {
	case "metrics":
		return map[string]interface{}{
			"metric_query": []interface{}{flattenMetricQueryJSON(q)},
		}
	case "logs", "spans", "profiling", "audit", "rum":
		return map[string]interface{}{
			"event_query": []interface{}{flattenEventQueryJSON(q)},
		}
	case "process":
		return map[string]interface{}{
			"process_query": []interface{}{flattenFormulaProcessQueryFlatJSON(q)},
		}
	case "slo":
		return map[string]interface{}{
			"slo_query": []interface{}{flattenSLOQueryJSON(q)},
		}
	case "cloud_cost":
		return map[string]interface{}{
			"cloud_cost_query": []interface{}{flattenCloudCostQueryJSON(q)},
		}
	case "apm_dependency_stats":
		return map[string]interface{}{
			"apm_dependency_stats_query": []interface{}{flattenApmDependencyStatsQueryJSON(q)},
		}
	case "apm_resource_stats":
		return map[string]interface{}{
			"apm_resource_stats_query": []interface{}{flattenApmResourceStatsQueryJSON(q)},
		}
	default:
		// fallback to generic
		if _, hasCompute := q["compute"]; hasCompute {
			return map[string]interface{}{
				"event_query": []interface{}{flattenEventQueryJSON(q)},
			}
		}
		if _, hasQuery := q["query"]; hasQuery {
			return map[string]interface{}{
				"metric_query": []interface{}{flattenMetricQueryJSON(q)},
			}
		}
		return map[string]interface{}{}
	}
}

// flattenApmDependencyStatsQueryJSON flattens an apm_dependency_stats_query JSON object.
func flattenApmDependencyStatsQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "env", "name", "operation_name", "resource_name", "service", "stat", "primary_tag_name", "primary_tag_value"} {
		if v, ok := q[field].(string); ok && v != "" {
			result[field] = v
		}
	}
	if v, ok := q["is_upstream"].(bool); ok {
		result["is_upstream"] = v
	}
	return result
}

// flattenApmResourceStatsQueryJSON flattens an apm_resource_stats_query JSON object.
func flattenApmResourceStatsQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "env", "name", "service", "stat", "operation_name", "resource_name", "primary_tag_name", "primary_tag_value"} {
		if v, ok := q[field].(string); ok && v != "" {
			result[field] = v
		}
	}
	if groupBy, ok := q["group_by"].([]interface{}); ok && len(groupBy) > 0 {
		strs := make([]string, len(groupBy))
		for i, g := range groupBy {
			strs[i] = fmt.Sprintf("%v", g)
		}
		result["group_by"] = strs
	}
	return result
}

// flattenQueryTableNumberFormatJSON flattens the number_format JSON into HCL state.
// The JSON has unit { type, unit_name, per_unit_name } or unit { type, label }.
// The HCL has unit { canonical { unit_name, per_unit_name } } or unit { custom { label } }.
func flattenQueryTableNumberFormatJSON(nf map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if unit, ok := nf["unit"].(map[string]interface{}); ok {
		unitState := map[string]interface{}{}
		unitType, _ := unit["type"].(string)
		switch unitType {
		case "canonical_unit":
			canonState := map[string]interface{}{}
			if v, ok := unit["unit_name"].(string); ok && v != "" {
				canonState["unit_name"] = v
			}
			if v, ok := unit["per_unit_name"].(string); ok && v != "" {
				canonState["per_unit_name"] = v
			}
			unitState["canonical"] = []interface{}{canonState}
		case "custom_unit_label":
			customState := map[string]interface{}{}
			if v, ok := unit["label"].(string); ok && v != "" {
				customState["label"] = v
			}
			unitState["custom"] = []interface{}{customState}
		}
		if len(unitState) > 0 {
			result["unit"] = []interface{}{unitState}
		}
	}
	if unitScale, ok := nf["unit_scale"].(map[string]interface{}); ok {
		unitScaleState := map[string]interface{}{}
		if v, ok := unitScale["unit_name"].(string); ok && v != "" {
			unitScaleState["unit_name"] = v
		}
		if len(unitScaleState) > 0 {
			result["unit_scale"] = []interface{}{unitScaleState}
		}
	}
	return result
}

// flattenQueryTableTextFormatsJSON flattens the 2D text_formats array.
func flattenQueryTableTextFormatsJSON(textFormats []interface{}) []interface{} {
	result := make([]interface{}, len(textFormats))
	for i, tf := range textFormats {
		rules, ok := tf.([]interface{})
		if !ok {
			result[i] = map[string]interface{}{"text_format": []interface{}{}}
			continue
		}
		flatRules := make([]interface{}, len(rules))
		for j, rule := range rules {
			rm, ok := rule.(map[string]interface{})
			if !ok {
				flatRules[j] = map[string]interface{}{}
				continue
			}
			flatRule := map[string]interface{}{}
			if match, ok := rm["match"].(map[string]interface{}); ok {
				matchState := map[string]interface{}{}
				if v, ok := match["type"].(string); ok {
					matchState["type"] = v
				}
				if v, ok := match["value"].(string); ok {
					matchState["value"] = v
				}
				flatRule["match"] = []interface{}{matchState}
			}
			if v, ok := rm["palette"].(string); ok && v != "" {
				flatRule["palette"] = v
			}
			if v, ok := rm["custom_bg_color"].(string); ok && v != "" {
				flatRule["custom_bg_color"] = v
			}
			if v, ok := rm["custom_fg_color"].(string); ok && v != "" {
				flatRule["custom_fg_color"] = v
			}
			if replace, ok := rm["replace"].(map[string]interface{}); ok {
				replaceState := map[string]interface{}{}
				if v, ok := replace["type"].(string); ok {
					replaceState["type"] = v
				}
				if v, ok := replace["with"].(string); ok {
					replaceState["with"] = v
				}
				if v, ok := replace["substring"].(string); ok && v != "" {
					replaceState["substring"] = v
				}
				flatRule["replace"] = []interface{}{replaceState}
			}
			flatRules[j] = flatRule
		}
		result[i] = map[string]interface{}{"text_format": flatRules}
	}
	return result
}

// buildSplitConfigStaticSplitsJSON builds the static_splits 2D JSON array for split_config.
// The HCL structure is:
//
//	static_splits { split_vector { tag_key, tag_values } split_vector { ... } }
//
// The JSON structure is: [[{"tag_key":"...","tag_values":[...]},...],...]
// splitConfigAttrs is the attributes map of the split_config block (not the outer attrs).
func buildSplitConfigStaticSplitsJSON(splitConfigAttrs map[string]attr.Value) []interface{} {
	staticSplitElems := getListElems(splitConfigAttrs, "static_splits")
	if len(staticSplitElems) == 0 {
		return nil
	}
	result := make([]interface{}, 0, len(staticSplitElems))
	for _, ssElem := range staticSplitElems {
		ssObj, ok := ssElem.(types.Object)
		if !ok {
			result = append(result, []interface{}{})
			continue
		}
		ssAttrs := ssObj.Attributes()
		vectorElems := getListElems(ssAttrs, "split_vector")
		innerArray := make([]interface{}, 0, len(vectorElems))
		for _, vElem := range vectorElems {
			vObj, ok := vElem.(types.Object)
			if !ok {
				innerArray = append(innerArray, map[string]interface{}{})
				continue
			}
			vAttrs := vObj.Attributes()
			vec := map[string]interface{}{}
			if v, ok := getStrAttr(vAttrs, "tag_key"); ok {
				vec["tag_key"] = v
			}
			// tag_values: always emit (even empty array)
			tagValues := getStrListElems(vAttrs, "tag_values")
			if tagValues == nil {
				tagValues = []string{}
			}
			vec["tag_values"] = tagValues
			innerArray = append(innerArray, vec)
		}
		result = append(result, innerArray)
	}
	return result
}

// flattenSplitConfigStaticSplitsJSON flattens the static_splits 2D JSON array into HCL state.
// The JSON structure is: [[{"tag_key":"...","tag_values":[...]},...],...]
// The HCL structure is: [{split_vector: [{tag_key, tag_values}, ...]}]
func flattenSplitConfigStaticSplitsJSON(staticSplits []interface{}) []interface{} {
	result := make([]interface{}, len(staticSplits))
	for i, entry := range staticSplits {
		innerArray, ok := entry.([]interface{})
		if !ok {
			result[i] = map[string]interface{}{"split_vector": []interface{}{}}
			continue
		}
		vectors := make([]interface{}, len(innerArray))
		for j, vec := range innerArray {
			vm, ok := vec.(map[string]interface{})
			if !ok {
				vectors[j] = map[string]interface{}{}
				continue
			}
			vecState := map[string]interface{}{}
			if v, ok := vm["tag_key"].(string); ok {
				vecState["tag_key"] = v
			}
			if tvs, ok := vm["tag_values"].([]interface{}); ok {
				tagValues := make([]string, len(tvs))
				for k, tv := range tvs {
					tagValues[k] = fmt.Sprintf("%v", tv)
				}
				vecState["tag_values"] = tagValues
			} else {
				vecState["tag_values"] = []string{}
			}
			vectors[j] = vecState
		}
		result[i] = map[string]interface{}{"split_vector": vectors}
	}
	return result
}

// buildSplitGraphSourceWidgetJSON builds the source_widget_definition JSON for split_graph.
// This requires type dispatch since source_widget_definition can be any of several widget types.
// It reuses allWidgetSpecs directly (no separate hardcoded list) — any widget type registered
// in allWidgetSpecs is automatically supported as a source widget.
// defAttrs is the attributes map of the split_graph definition block.
func buildSplitGraphSourceWidgetJSON(defAttrs map[string]attr.Value) map[string]interface{} {
	sourceWidgetAttrs := getObjAttrs(defAttrs, "source_widget_definition")
	if sourceWidgetAttrs == nil {
		return nil
	}
	// source_widget_definition is like a miniature widget: try each registered spec.
	for _, spec := range allWidgetSpecs {
		innerElems := getListElems(sourceWidgetAttrs, spec.HCLKey)
		if len(innerElems) == 0 {
			continue
		}
		innerObj, ok := innerElems[0].(types.Object)
		if !ok {
			continue
		}
		innerAttrs := innerObj.Attributes()
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := BuildEngineJSON(innerAttrs, allFields)
		defJSON["type"] = spec.JSONType
		// Apply the same per-widget post-processing as the main engine
		buildWidgetPostProcess(innerAttrs, spec, defJSON)
		return defJSON
	}
	return nil
}

// flattenSplitGraphSourceWidgetJSON flattens the source_widget_definition JSON
// for a split_graph widget response.
func flattenSplitGraphSourceWidgetJSON(srcDef map[string]interface{}) map[string]interface{} {
	widgetType, _ := srcDef["type"].(string)
	for _, spec := range allWidgetSpecs {
		if spec.JSONType != widgetType {
			continue
		}
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defState := FlattenEngineJSON(allFields, srcDef)
		// Apply the same per-widget post-processing as the main flatten engine
		flattenWidgetPostProcess(spec, srcDef, defState)
		return map[string]interface{}{
			spec.HCLKey: []interface{}{defState},
		}
	}
	return nil
}

// buildGroupWidgetsJSON builds the "widgets" array for a group widget.
func buildGroupWidgetsJSON(defAttrs map[string]attr.Value) []interface{} {
	widgetElems := getListElems(defAttrs, "widget")
	widgets := make([]interface{}, 0, len(widgetElems))
	for _, elem := range widgetElems {
		if wObj, ok := elem.(types.Object); ok {
			w := buildWidgetEngineJSON(wObj.Attributes())
			if w == nil {
				w = map[string]interface{}{}
			}
			widgets = append(widgets, w)
		}
	}
	return widgets
}

// flattenGroupWidgetsJSON flattens the "widgets" array from a group widget JSON response.
func flattenGroupWidgetsJSON(widgets []interface{}) []interface{} {
	result := make([]interface{}, len(widgets))
	for i, w := range widgets {
		widgetData, ok := w.(map[string]interface{})
		if !ok {
			result[i] = map[string]interface{}{}
			continue
		}
		flattened := flattenWidgetEngineJSON(widgetData)
		if flattened == nil {
			flattened = map[string]interface{}{}
		}
		// Handle widget IDs for child widgets
		if def, ok := widgetData["definition"].(map[string]interface{}); ok {
			if widgetID, ok := def["id"]; ok {
				switch v := widgetID.(type) {
				case float64:
					flattened["id"] = int(v)
				case int:
					flattened["id"] = v
				}
			}
		}
		if widgetID, ok := widgetData["id"]; ok {
			switch v := widgetID.(type) {
			case float64:
				flattened["id"] = int(v)
			case int:
				flattened["id"] = v
			}
		}
		// Handle widget_layout from "layout" object
		if layout, ok := widgetData["layout"].(map[string]interface{}); ok {
			layoutState := map[string]interface{}{}
			for _, key := range []string{"x", "y", "width", "height"} {
				if v, ok := layout[key]; ok {
					switch iv := v.(type) {
					case float64:
						layoutState[key] = int(iv)
					case int:
						layoutState[key] = iv
					}
				}
			}
			if v, ok := layout["is_column_break"].(bool); ok && v {
				layoutState["is_column_break"] = v
			}
			if len(layoutState) > 0 {
				flattened["widget_layout"] = []interface{}{layoutState}
			}
		}
		result[i] = flattened
	}
	return result
}

// ============================================================
// Dashboard-Level Build and Flatten
// ============================================================

// DashboardAPIPath is the Datadog API path for dashboards.
const DashboardAPIPath = "/api/v1/dashboard"

// BuildDashboardEngineJSON builds the full dashboard JSON body from a framework attrs map.
func BuildDashboardEngineJSON(attrs map[string]attr.Value, id string) map[string]interface{} {
	result := BuildEngineJSON(attrs, DashboardTopLevelFields)

	// Both POST and PUT bodies need this field set to the current ID.
	result["id"] = id

	// Handle widgets with type dispatch.
	widgetElems := getListElems(attrs, "widget")
	widgets := make([]interface{}, 0, len(widgetElems))
	for _, elem := range widgetElems {
		if wObj, ok := elem.(types.Object); ok {
			w := buildWidgetEngineJSON(wObj.Attributes())
			if w != nil {
				widgets = append(widgets, w)
			}
		}
	}
	result["widgets"] = widgets

	return result
}

// FlattenTemplateVariables converts the JSON template_variables list to HCL state.
func FlattenTemplateVariables(tvs []interface{}) []interface{} {
	result := make([]interface{}, len(tvs))
	for i, tv := range tvs {
		m, ok := tv.(map[string]interface{})
		if !ok {
			result[i] = map[string]interface{}{}
			continue
		}
		item := map[string]interface{}{}
		if v, ok := m["name"]; ok {
			item["name"] = fmt.Sprintf("%v", v)
		}
		if v, ok := m["prefix"]; ok && v != nil {
			item["prefix"] = fmt.Sprintf("%v", v)
		}
		if v, ok := m["defaults"].([]interface{}); ok && len(v) > 0 {
			defaults := make([]string, len(v))
			for j, d := range v {
				defaults[j] = fmt.Sprintf("%v", d)
			}
			item["defaults"] = defaults
		} else if v, ok := m["default"]; ok && v != nil {
			item["default"] = fmt.Sprintf("%v", v)
		}
		if v, ok := m["available_values"].([]interface{}); ok {
			availVals := make([]string, len(v))
			for j, av := range v {
				availVals[j] = fmt.Sprintf("%v", av)
			}
			item["available_values"] = availVals
		}
		if v, ok := m["type"]; ok && v != nil {
			item["type"] = fmt.Sprintf("%v", v)
		}
		result[i] = item
	}
	return result
}

// FlattenTemplateVariablePresets converts the JSON template_variable_presets list to HCL state.
func FlattenTemplateVariablePresets(tvps []interface{}) []interface{} {
	result := make([]interface{}, len(tvps))
	for i, tvp := range tvps {
		m, ok := tvp.(map[string]interface{})
		if !ok {
			result[i] = map[string]interface{}{}
			continue
		}
		item := map[string]interface{}{}
		if v, ok := m["name"]; ok {
			item["name"] = fmt.Sprintf("%v", v)
		}
		if tvVals, ok := m["template_variables"].([]interface{}); ok {
			vals := make([]interface{}, len(tvVals))
			for j, tvv := range tvVals {
				vm, ok := tvv.(map[string]interface{})
				if !ok {
					vals[j] = map[string]interface{}{}
					continue
				}
				valItem := map[string]interface{}{}
				if v, ok := vm["name"]; ok {
					valItem["name"] = fmt.Sprintf("%v", v)
				}
				if v, ok := vm["values"].([]interface{}); ok && len(v) > 0 {
					valStrs := make([]string, len(v))
					for k, vs := range v {
						valStrs[k] = fmt.Sprintf("%v", vs)
					}
					valItem["values"] = valStrs
				} else if v, ok := vm["value"]; ok && v != nil {
					valItem["value"] = fmt.Sprintf("%v", v)
				}
				vals[j] = valItem
			}
			item["template_variable"] = vals
		}
		result[i] = item
	}
	return result
}

// BuildWidgetEngineJSON is the exported entry point for building a single widget's JSON map
// from a framework attrs map. Returns nil if no recognized widget type is found.
func BuildWidgetEngineJSON(attrs map[string]attr.Value) map[string]interface{} {
	return buildWidgetEngineJSON(attrs)
}

// FlattenWidgetEngineJSON is the exported entry point for flattening a single widget's JSON map
// into HCL state suitable for setting on a TypeList field in schema.ResourceData.
// Returns nil if the widget type is not recognized.
func FlattenWidgetEngineJSON(widgetData map[string]interface{}) map[string]interface{} {
	return flattenWidgetEngineJSON(widgetData)
}

// MarshalDashboardJSON marshals the dashboard JSON body from a framework attrs map.
// The trailing newline matches behavior of json.NewEncoder used when cassettes were recorded.
func MarshalDashboardJSON(attrs map[string]attr.Value, id string) (string, error) {
	body, err := json.Marshal(BuildDashboardEngineJSON(attrs, id))
	if err != nil {
		return "", fmt.Errorf("error marshaling dashboard JSON: %s", err)
	}
	return string(body) + "\n", nil
}

// ============================================================
// Framework State Conversion (JSON → attr.Value)
// ============================================================

// RawToAttrValue converts a raw interface{} value from JSON to the appropriate attr.Value
// based on the target attr.Type. Returns a null value of the correct type if v is nil.
func RawToAttrValue(v interface{}, t attr.Type) attr.Value {
	if v == nil {
		return nullValue(t)
	}
	switch ct := t.(type) {
	case basetypes.StringType:
		return types.StringValue(fmt.Sprintf("%v", v))
	case basetypes.BoolType:
		if b, ok := v.(bool); ok {
			return types.BoolValue(b)
		}
		return types.BoolNull()
	case basetypes.Int64Type:
		switch iv := v.(type) {
		case float64:
			return types.Int64Value(int64(iv))
		case int:
			return types.Int64Value(int64(iv))
		case int64:
			return types.Int64Value(iv)
		}
		return types.Int64Null()
	case basetypes.Float64Type:
		switch fv := v.(type) {
		case float64:
			return types.Float64Value(fv)
		case int:
			return types.Float64Value(float64(fv))
		}
		return types.Float64Null()
	case basetypes.ListType:
		// Handle both []interface{} (from JSON) and []string/[]int (from FlattenEngineJSON)
		var elems []attr.Value
		switch items := v.(type) {
		case []interface{}:
			elems = make([]attr.Value, len(items))
			for i, item := range items {
				elems[i] = RawToAttrValue(item, ct.ElemType)
			}
		case []string:
			elems = make([]attr.Value, len(items))
			for i, item := range items {
				elems[i] = RawToAttrValue(item, ct.ElemType)
			}
		case []int:
			elems = make([]attr.Value, len(items))
			for i, item := range items {
				elems[i] = RawToAttrValue(item, ct.ElemType)
			}
		default:
			return types.ListNull(ct.ElemType)
		}
		lv, diags := types.ListValue(ct.ElemType, elems)
		if diags.HasError() {
			return types.ListNull(ct.ElemType)
		}
		return lv
	case basetypes.SetType:
		var elems []attr.Value
		switch items := v.(type) {
		case []interface{}:
			elems = make([]attr.Value, len(items))
			for i, item := range items {
				elems[i] = RawToAttrValue(item, ct.ElemType)
			}
		case []string:
			elems = make([]attr.Value, len(items))
			for i, item := range items {
				elems[i] = RawToAttrValue(item, ct.ElemType)
			}
		case []int:
			elems = make([]attr.Value, len(items))
			for i, item := range items {
				elems[i] = RawToAttrValue(item, ct.ElemType)
			}
		default:
			return types.SetNull(ct.ElemType)
		}
		sv, diags := types.SetValue(ct.ElemType, elems)
		if diags.HasError() {
			return types.SetNull(ct.ElemType)
		}
		return sv
	case basetypes.ObjectType:
		m, ok := v.(map[string]interface{})
		if !ok {
			return types.ObjectNull(ct.AttrTypes)
		}
		attrVals := MapToAttrValues(m, ct.AttrTypes)
		obj, diags := types.ObjectValue(ct.AttrTypes, attrVals)
		if diags.HasError() {
			return types.ObjectNull(ct.AttrTypes)
		}
		return obj
	}
	return nullValue(t)
}

// nullValue returns a null attr.Value of the correct type.
func nullValue(t attr.Type) attr.Value {
	switch ct := t.(type) {
	case basetypes.StringType:
		return types.StringNull()
	case basetypes.BoolType:
		return types.BoolNull()
	case basetypes.Int64Type:
		return types.Int64Null()
	case basetypes.Float64Type:
		return types.Float64Null()
	case basetypes.ListType:
		return types.ListNull(ct.ElemType)
	case basetypes.SetType:
		return types.SetNull(ct.ElemType)
	case basetypes.ObjectType:
		return types.ObjectNull(ct.AttrTypes)
	}
	return types.StringNull()
}

// MapToAttrValues converts a map[string]interface{} to map[string]attr.Value
// using the provided attrTypes for type guidance.
// For missing keys, uses SDKv2-compatible zero values to match v1 behavior.
func MapToAttrValues(data map[string]interface{}, attrTypes map[string]attr.Type) map[string]attr.Value {
	result := make(map[string]attr.Value, len(attrTypes))
	for key, t := range attrTypes {
		if v, ok := data[key]; ok {
			result[key] = RawToAttrValue(v, t)
		} else {
			result[key] = zeroValue(t)
		}
	}
	return result
}

// zeroValue returns the SDKv2-compatible zero value for a type.
// Matches TypeString default "", TypeInt default 0, TypeBool default false, TypeList default [].
func zeroValue(t attr.Type) attr.Value {
	switch ct := t.(type) {
	case basetypes.StringType:
		return types.StringValue("")
	case basetypes.BoolType:
		return types.BoolValue(false)
	case basetypes.Int64Type:
		return types.Int64Value(0)
	case basetypes.Float64Type:
		return types.Float64Value(0)
	case basetypes.ListType:
		return types.ListValueMust(ct.ElemType, []attr.Value{})
	case basetypes.SetType:
		return types.SetValueMust(ct.ElemType, []attr.Value{})
	case basetypes.ObjectType:
		attrs := make(map[string]attr.Value, len(ct.AttrTypes))
		for k, at := range ct.AttrTypes {
			attrs[k] = zeroValue(at)
		}
		obj, diags := types.ObjectValue(ct.AttrTypes, attrs)
		if diags.HasError() {
			return types.ObjectNull(ct.AttrTypes)
		}
		return obj
	}
	return nullValue(t)
}

// FlattenWidgetsToFW converts the "widgets" array from a dashboard API response
// into a types.List of widget objects suitable for use with AllWidgetAttrTypes.
// excludePowerpackOnly controls whether powerpack/split_group definitions are included.
func FlattenWidgetsToFW(widgets []interface{}, excludePowerpackOnly bool) (types.List, error) {
	widgetAttrTypes := AllWidgetAttrTypes(excludePowerpackOnly)
	widgetObjType := types.ObjectType{AttrTypes: widgetAttrTypes}

	if len(widgets) == 0 {
		return types.ListValueMust(widgetObjType, []attr.Value{}), nil
	}

	elems := make([]attr.Value, 0, len(widgets))
	for _, w := range widgets {
		widgetData, ok := w.(map[string]interface{})
		if !ok {
			continue
		}

		// Flatten the widget definition
		flattened := flattenWidgetEngineJSON(widgetData)
		if flattened == nil {
			flattened = map[string]interface{}{}
		}

		// Handle widget ID
		if def, ok := widgetData["definition"].(map[string]interface{}); ok {
			if widgetID, ok := def["id"]; ok {
				switch v := widgetID.(type) {
				case float64:
					flattened["id"] = int(v)
				case int:
					flattened["id"] = v
				}
			}
		}
		if widgetID, ok := widgetData["id"]; ok {
			switch v := widgetID.(type) {
			case float64:
				flattened["id"] = int(v)
			case int:
				flattened["id"] = v
			}
		}

		// Flatten widget_layout from JSON "layout" object
		if layout, ok := widgetData["layout"].(map[string]interface{}); ok {
			layoutState := map[string]interface{}{}
			for _, key := range []string{"x", "y", "width", "height"} {
				if v, ok := layout[key]; ok {
					switch iv := v.(type) {
					case float64:
						layoutState[key] = int(iv)
					case int:
						layoutState[key] = iv
					}
				}
			}
			if v, ok := layout["is_column_break"].(bool); ok && v {
				layoutState["is_column_break"] = v
			}
			if len(layoutState) > 0 {
				flattened["widget_layout"] = []interface{}{layoutState}
			}
		}

		// Convert flattened map to framework attr.Value map
		attrVals := MapToAttrValues(flattened, widgetAttrTypes)
		widgetObj, diags := types.ObjectValue(widgetAttrTypes, attrVals)
		if diags.HasError() {
			continue
		}
		elems = append(elems, widgetObj)
	}

	widgetList, diags := types.ListValue(widgetObjType, elems)
	if diags.HasError() {
		return types.ListNull(widgetObjType), fmt.Errorf("error creating widget list")
	}
	return widgetList, nil
}

// FlattenListToFW converts a []interface{} of flat maps to a types.List
// of objects using the provided attrTypes.
func FlattenListToFW(items []interface{}, attrTypes map[string]attr.Type) (types.List, error) {
	objType := types.ObjectType{AttrTypes: attrTypes}
	elems := make([]attr.Value, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		attrVals := MapToAttrValues(m, attrTypes)
		obj, diags := types.ObjectValue(attrTypes, attrVals)
		if diags.HasError() {
			continue
		}
		elems = append(elems, obj)
	}
	list, diags := types.ListValue(objType, elems)
	if diags.HasError() {
		return types.ListNull(objType), fmt.Errorf("error creating list value")
	}
	return list, nil
}
