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
	TypeOneOf                       // polymorphic block — exactly one child variant should be set
)

// OneOfDiscriminator configures how a TypeOneOf field determines which variant is active.
// Set on the TypeOneOf parent field (JSONKey) and on each child variant field (Value/Values).
type OneOfDiscriminator struct {
	// JSONKey is the field in the JSON object that identifies the variant.
	// Set on the parent TypeOneOf FieldSpec. Used to read the discriminator on flatten
	// and as the key to inject on build when a child has a non-empty Value.
	JSONKey string

	// Value is the expected discriminator value for this variant.
	// Set on child FieldSpecs. If non-empty, this value is injected into the built JSON
	// object on build (using the parent's JSONKey). Also used for flatten matching.
	Value string

	// Values lists multiple JSON discriminator values that all map to this variant.
	// Used for flatten-direction matching only (no build injection). Set on child FieldSpecs.
	// Example: SunburstLegend "table" variant matches both "table" and "none".
	Values []string

	// DefaultVariant marks this child as the fallback when no discriminator field
	// exists in the JSON or no Value/Values match. Used for legacy variants that
	// predate the discriminator field (e.g. WidgetLegacyLiveSpan has no "type" field).
	DefaultVariant bool
}

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

	// Discriminator configures polymorphic oneOf behavior for TypeOneOf fields.
	// Set on the TypeOneOf parent (JSONKey) and on each child variant (Value/Values/DefaultVariant).
	Discriminator *OneOfDiscriminator
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

		case TypeOneOf:
			// Get the outer TypeOneOf block (MaxItems:1 container)
			elems := getListElems(attrs, f.HCLKey)
			if len(elems) == 0 {
				if f.OmitEmpty {
					continue
				}
				break
			}
			outerObj, ok := elems[0].(types.Object)
			if !ok {
				continue
			}
			outerAttrs := outerObj.Attributes()

			// Find the first populated variant child
			var built map[string]interface{}
			var matchedChild *FieldSpec
			for i := range f.Children {
				child := &f.Children[i]
				if child.Type == TypeBlock {
					childElems := getListElems(outerAttrs, child.HCLKey)
					if len(childElems) > 0 {
						if childObj, ok := childElems[0].(types.Object); ok {
							built = BuildEngineJSON(childObj.Attributes(), child.Children)
							matchedChild = child
						}
					}
				}
				if matchedChild != nil {
					break
				}
			}
			if built == nil {
				if f.OmitEmpty {
					continue
				}
				built = map[string]interface{}{}
			}
			// Inject discriminator into built JSON if the matched child specifies a value
			if matchedChild != nil && matchedChild.Discriminator != nil &&
				matchedChild.Discriminator.Value != "" &&
				f.Discriminator != nil && f.Discriminator.JSONKey != "" {
				built[f.Discriminator.JSONKey] = matchedChild.Discriminator.Value
			}
			setAtJSONPath(result, f.effectiveJSONPath(), built)
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

		case TypeOneOf:
			jsonObj, ok := jsonVal.(map[string]interface{})
			if !ok {
				continue
			}
			// Read discriminator value from JSON
			var discriminatorValue string
			if f.Discriminator != nil && f.Discriminator.JSONKey != "" {
				discriminatorValue, _ = jsonObj[f.Discriminator.JSONKey].(string)
			}
			// Find the matching variant child
			var matchedChild *FieldSpec
			for i := range f.Children {
				child := &f.Children[i]
				if child.Discriminator == nil {
					continue
				}
				if child.Discriminator.Value != "" && child.Discriminator.Value == discriminatorValue {
					matchedChild = child
					break
				}
				for _, v := range child.Discriminator.Values {
					if v == discriminatorValue {
						matchedChild = child
						break
					}
				}
				if matchedChild != nil {
					break
				}
			}
			// Fall back to DefaultVariant if no explicit match
			if matchedChild == nil {
				for i := range f.Children {
					if f.Children[i].Discriminator != nil && f.Children[i].Discriminator.DefaultVariant {
						matchedChild = &f.Children[i]
						break
					}
				}
			}
			if matchedChild == nil {
				continue
			}
			// Flatten the matched variant into a state map for the outer TypeOneOf block
			variantState := map[string]interface{}{}
			if matchedChild.Type == TypeBlock {
				variantState[matchedChild.HCLKey] = []interface{}{FlattenEngineJSON(matchedChild.Children, jsonObj)}
			}
			result[f.HCLKey] = []interface{}{variantState}
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

// ============================================================
// Formula Request Config — unified formula/query request builder
// ============================================================

// FormulaRequestConfig describes how to build and flatten a formula/query request
// for a specific widget type. The three previously separate builders
// (timeseries, scalar, query_table) are unified through this struct.
type FormulaRequestConfig struct {
	// ResponseFormat is the value injected as "response_format" in JSON.
	ResponseFormat string
	// StyleFields defines the request-level style block (palette, line_type, etc.).
	// nil = no style block for this widget type.
	StyleFields []FieldSpec
	// ExtraFields are widget-specific request-level fields emitted before formulas.
	// Example: on_right_yaxis + display_type for timeseries; change-widget fields.
	ExtraFields []FieldSpec
	// IncludeSort: when true, build/flatten the sort block (toplist, geomap, etc.).
	IncludeSort bool
}

// Per-widget FormulaRequestConfig declarations.
var timeseriesFormulaRequestConfig = FormulaRequestConfig{
	ResponseFormat: "timeseries",
	StyleFields:    timeseriesWidgetRequestStyleFields,
	ExtraFields: []FieldSpec{
		{HCLKey: "on_right_yaxis", Type: TypeBool, OmitEmpty: false,
			Description: "Boolean indicating whether to display a timeseries on the right side of the widget."},
		{HCLKey: "display_type", Type: TypeString, OmitEmpty: true,
			Description: "How the data points are displayed on the graph."},
	},
}

var heatmapFormulaRequestConfig = FormulaRequestConfig{
	ResponseFormat: "timeseries", // heatmap uses timeseries response_format
	StyleFields:    widgetRequestStyleFields,
}

var changeFormulaRequestConfig = FormulaRequestConfig{
	ResponseFormat: "scalar",
	StyleFields:    widgetRequestStyleFields,
	ExtraFields: []FieldSpec{
		{HCLKey: "increase_good", Type: TypeBool, OmitEmpty: false,
			Description: "Boolean indicating whether an increase in the value is good (depicted in green) or bad (depicted in red)."},
		{HCLKey: "show_present", Type: TypeBool, OmitEmpty: false,
			Description: "If set to `true`, displays current value."},
		{HCLKey: "change_type", Type: TypeString, OmitEmpty: true,
			Description: "Whether to show absolute or relative change."},
		{HCLKey: "compare_to", Type: TypeString, OmitEmpty: true,
			Description: "Timeframe used for the change comparison."},
		{HCLKey: "order_by", Type: TypeString, OmitEmpty: true,
			Description: "What to order by."},
		{HCLKey: "order_dir", Type: TypeString, OmitEmpty: true,
			Description: "Widget sorting methods."},
	},
}

var scalarFormulaRequestConfig = FormulaRequestConfig{
	ResponseFormat: "scalar",
	StyleFields:    widgetRequestStyleFields,
	IncludeSort:    true,
}

var queryTableFormulaRequestConfig = FormulaRequestConfig{
	ResponseFormat: "scalar",
}

// formulaRequestConfigForWidget returns the FormulaRequestConfig for a given widget type.
func formulaRequestConfigForWidget(jsonType string) FormulaRequestConfig {
	switch jsonType {
	case "timeseries":
		return timeseriesFormulaRequestConfig
	case "heatmap":
		return heatmapFormulaRequestConfig
	case "change":
		return changeFormulaRequestConfig
	default:
		return scalarFormulaRequestConfig
	}
}

// buildFormulaRequest is the unified formula/query request builder.
// It replaces buildFormulaQueryRequestJSON, buildScalarFormulaQueryRequestJSON,
// and buildQueryTableFormulaRequestJSON.
func buildFormulaRequest(attrs map[string]attr.Value, cfg FormulaRequestConfig) map[string]interface{} {
	result := map[string]interface{}{}
	// Widget-specific extra fields (e.g. on_right_yaxis, increase_good)
	if len(cfg.ExtraFields) > 0 {
		for k, v := range BuildEngineJSON(attrs, cfg.ExtraFields) {
			result[k] = v
		}
	}
	// Formulas
	formulaElems := getListElems(attrs, "formula")
	if len(formulaElems) > 0 {
		formulas := make([]interface{}, 0, len(formulaElems))
		for _, fElem := range formulaElems {
			if fObj, ok := fElem.(types.Object); ok {
				formulas = append(formulas, BuildEngineJSON(fObj.Attributes(), widgetFormulaFields))
			}
		}
		result["formulas"] = formulas
	}
	// Queries
	queryElems := getListElems(attrs, "query")
	if len(queryElems) > 0 {
		queries := make([]interface{}, 0, len(queryElems))
		for _, qElem := range queryElems {
			if qObj, ok := qElem.(types.Object); ok {
				queries = append(queries, buildQueryFromAttrs(qObj.Attributes()))
			}
		}
		result["queries"] = queries
	}
	// Request-level style (palette, line_type, line_width — varies by widget)
	if cfg.StyleFields != nil {
		if styleAttrs := getObjAttrs(attrs, "style"); styleAttrs != nil {
			style := BuildEngineJSON(styleAttrs, cfg.StyleFields)
			if len(style) > 0 {
				result["style"] = style
			}
		}
	}
	// Sort (toplist, geomap, etc.)
	if cfg.IncludeSort {
		if sortAttrs := getObjAttrs(attrs, "sort"); sortAttrs != nil {
			if sortJSON := buildWidgetSortByJSON(sortAttrs); len(sortJSON) > 0 {
				result["sort"] = sortJSON
			}
		}
	}
	// response_format
	if len(formulaElems) > 0 || len(queryElems) > 0 {
		result["response_format"] = cfg.ResponseFormat
	}
	return result
}

// flattenFormulaRequest is the unified formula/query request flattener.
// It replaces flattenFormulaQueryRequestJSON, flattenScalarFormulaQueryRequestJSON,
// and flattenQueryTableFormulaRequestJSON.
func flattenFormulaRequest(req map[string]interface{}, cfg FormulaRequestConfig) map[string]interface{} {
	result := map[string]interface{}{}
	// Widget-specific extra fields
	if len(cfg.ExtraFields) > 0 {
		for k, v := range FlattenEngineJSON(cfg.ExtraFields, req) {
			result[k] = v
		}
	}
	// Formulas
	if formulas, ok := req["formulas"].([]interface{}); ok {
		flat := make([]interface{}, len(formulas))
		for i, f := range formulas {
			if fm, ok := f.(map[string]interface{}); ok {
				flat[i] = FlattenEngineJSON(widgetFormulaFields, fm)
			} else {
				flat[i] = map[string]interface{}{}
			}
		}
		result["formula"] = flat
	}
	// Queries
	if queries, ok := req["queries"].([]interface{}); ok {
		flat := make([]interface{}, len(queries))
		for i, q := range queries {
			if qm, ok := q.(map[string]interface{}); ok {
				flat[i] = flattenFormulaQueryJSON(qm)
			} else {
				flat[i] = map[string]interface{}{}
			}
		}
		result["query"] = flat
	}
	// Request-level style
	if cfg.StyleFields != nil {
		if style, ok := req["style"].(map[string]interface{}); ok {
			s := FlattenEngineJSON(cfg.StyleFields, style)
			if len(s) > 0 {
				result["style"] = []interface{}{s}
			}
		}
	}
	// Sort
	if cfg.IncludeSort {
		if sortObj, ok := req["sort"].(map[string]interface{}); ok {
			if s := flattenWidgetSortByJSON(sortObj); len(s) > 0 {
				result["sort"] = []interface{}{s}
			}
		}
	}
	return result
}

// dataSourceToQueryType maps JSON data_source values to HCL query block keys.
// Used by flattenFormulaQueryJSON to route flattened queries to the right block.
var dataSourceToQueryType = map[string]string{
	"metrics":              "metric_query",
	"logs":                 "event_query",
	"spans":                "event_query",
	"profiling":            "event_query",
	"audit":                "event_query",
	"rum":                  "event_query",
	"process":              "process_query",
	"slo":                  "slo_query",
	"cloud_cost":           "cloud_cost_query",
	"apm_dependency_stats": "apm_dependency_stats_query",
	"apm_resource_stats":   "apm_resource_stats_query",
}

// buildQueryFromAttrs dispatches a query attrs map to the appropriate query builder.
// event_query uses custom logic (nested compute/search/group_by); all others use
// BuildEngineJSON with the matching FieldSpec children from formulaAndFunctionQueryFields.
func buildQueryFromAttrs(attrs map[string]attr.Value) map[string]interface{} {
	for _, child := range formulaAndFunctionQueryFields {
		if childAttrs := getObjAttrs(attrs, child.HCLKey); childAttrs != nil {
			if child.HCLKey == "event_query" {
				return buildEventQueryJSON(childAttrs)
			}
			return BuildEngineJSON(childAttrs, child.Children)
		}
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

// buildWidgetSortByJSON builds the JSON for a WidgetSortBy block (sort { count, order_by { ... } }).
// Each order_by element is a discriminated union: formula_sort or group_sort.
func buildWidgetSortByJSON(sortAttrs map[string]attr.Value) map[string]interface{} {
	result := map[string]interface{}{}
	if count, ok := getInt64Attr(sortAttrs, "count"); ok && count > 0 {
		result["count"] = int(count)
	}
	orderByElems := getListElems(sortAttrs, "order_by")
	if len(orderByElems) > 0 {
		orderBys := make([]interface{}, 0, len(orderByElems))
		for _, obElem := range orderByElems {
			obObj, ok := obElem.(types.Object)
			if !ok {
				continue
			}
			obAttrs := obObj.Attributes()
			entry := map[string]interface{}{}
			if fsAttrs := getObjAttrs(obAttrs, "formula_sort"); fsAttrs != nil {
				entry["type"] = "formula"
				if idx, ok := getInt64Attr(fsAttrs, "index"); ok {
					entry["index"] = int(idx)
				}
				if ord, ok := getStrAttr(fsAttrs, "order"); ok && ord != "" {
					entry["order"] = ord
				}
			} else if gsAttrs := getObjAttrs(obAttrs, "group_sort"); gsAttrs != nil {
				entry["type"] = "group"
				if name, ok := getStrAttr(gsAttrs, "name"); ok && name != "" {
					entry["name"] = name
				}
				if ord, ok := getStrAttr(gsAttrs, "order"); ok && ord != "" {
					entry["order"] = ord
				}
			}
			if len(entry) > 0 {
				orderBys = append(orderBys, entry)
			}
		}
		if len(orderBys) > 0 {
			result["order_by"] = orderBys
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

// flattenFormulaQueryJSON converts a single query JSON object into the HCL query block state.
// Returns a map with one key: the query type name (e.g., "metric_query").
// event_query uses custom logic; all others use FlattenEngineJSON with the matching FieldSpec.
func flattenFormulaQueryJSON(q map[string]interface{}) map[string]interface{} {
	dataSource, _ := q["data_source"].(string)
	targetKey, ok := dataSourceToQueryType[dataSource]
	if !ok {
		// Fallback heuristics for unknown data_source
		if _, hasCompute := q["compute"]; hasCompute {
			targetKey = "event_query"
		} else if _, hasQuery := q["query"]; hasQuery {
			targetKey = "metric_query"
		} else {
			return map[string]interface{}{}
		}
	}
	for _, child := range formulaAndFunctionQueryFields {
		if child.HCLKey == targetKey {
			if child.HCLKey == "event_query" {
				return map[string]interface{}{targetKey: []interface{}{flattenEventQueryJSON(q)}}
			}
			return map[string]interface{}{targetKey: []interface{}{FlattenEngineJSON(child.Children, q)}}
		}
	}
	return map[string]interface{}{}
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

// flattenWidgetSortByJSON flattens the WidgetSortBy JSON object into HCL state.
func flattenWidgetSortByJSON(sortObj map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if count, ok := sortObj["count"].(float64); ok && count > 0 {
		result["count"] = int(count)
	}
	if orderBys, ok := sortObj["order_by"].([]interface{}); ok && len(orderBys) > 0 {
		flatOBs := make([]interface{}, 0, len(orderBys))
		for _, ob := range orderBys {
			obMap, ok := ob.(map[string]interface{})
			if !ok {
				continue
			}
			obState := map[string]interface{}{}
			sortType, _ := obMap["type"].(string)
			switch sortType {
			case "formula":
				fs := map[string]interface{}{}
				if idx, ok := obMap["index"].(float64); ok {
					fs["index"] = int(idx)
				}
				if ord, ok := obMap["order"].(string); ok {
					fs["order"] = ord
				}
				obState["formula_sort"] = []interface{}{fs}
			case "group":
				gs := map[string]interface{}{}
				if name, ok := obMap["name"].(string); ok {
					gs["name"] = name
				}
				if ord, ok := obMap["order"].(string); ok {
					gs["order"] = ord
				}
				obState["group_sort"] = []interface{}{gs}
			}
			if len(obState) > 0 {
				flatOBs = append(flatOBs, obState)
			}
		}
		if len(flatOBs) > 0 {
			result["order_by"] = flatOBs
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
					// New-style formula/query request (unified via FormulaRequestConfig)
					requests = append(requests, buildFormulaRequest(reqAttrs, formulaRequestConfigForWidget(spec.JSONType)))
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

	// ---- Funnel request_type injection ----
	// The OpenAPI FunnelWidgetRequest requires request_type = "funnel" but we don't
	// expose it in HCL (it's always "funnel"). Inject it into each request object.
	if spec.JSONType == "funnel" {
		if requests, ok := defJSON["requests"].([]interface{}); ok {
			for _, req := range requests {
				if reqMap, ok := req.(map[string]interface{}); ok {
					reqMap["request_type"] = "funnel"
				}
			}
		}
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
					// New-style formula/query request (unified via FormulaRequestConfig)
					flatRequests[ri] = flattenFormulaRequest(reqMap, formulaRequestConfigForWidget(spec.JSONType))
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
			requests = append(requests, buildFormulaRequest(reqAttrs, queryTableFormulaRequestConfig))
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

// flattenQueryTableRequestJSON flattens a single query_table request JSON.
// Handles both formula-style (has "formulas"/"queries") and old-style requests.
func flattenQueryTableRequestJSON(req map[string]interface{}) map[string]interface{} {
	_, hasFormulas := req["formulas"]
	_, hasQueries := req["queries"]
	if hasFormulas || hasQueries {
		return flattenFormulaRequest(req, queryTableFormulaRequestConfig)
	}
	// Old-style request
	result := FlattenEngineJSON(queryTableOldRequestFields, req)
	// text_formats (2D array) needs special handling
	if textFormats, ok := req["text_formats"].([]interface{}); ok && len(textFormats) > 0 {
		result["text_formats"] = flattenQueryTableTextFormatsJSON(textFormats)
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
