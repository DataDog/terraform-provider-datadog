package dashboardmapping

// engine.go
//
// FieldSpec bidirectional mapping engine for the dashboard and powerpack resources.
// Contains:
//   - Generic types: FieldSpec, WidgetSpec, FieldType, FormulaRequestConfig
//   - JSON flatten direction (HCL state ← API JSON)
//   - JSON build direction (HCL state → API JSON) using SDKv2 map[string]interface{}
//   - Shared helpers: path utilities, formula request builder/flattener

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
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
	TypeJSON                        // arbitrary JSON — HCL string ↔ JSON object/array
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
// Shared OneOf Discriminator Helpers
// ============================================================

// matchOneOfVariant finds the child variant matching the discriminator value
// in a JSON object. Used by both TypeOneOf and discriminated TypeBlockList.
func matchOneOfVariant(f FieldSpec, jsonObj map[string]interface{}) *FieldSpec {
	var discriminatorValue string
	if f.Discriminator != nil && f.Discriminator.JSONKey != "" {
		discriminatorValue, _ = jsonObj[f.Discriminator.JSONKey].(string)
	}
	for i := range f.Children {
		child := &f.Children[i]
		if child.Discriminator == nil {
			continue
		}
		if child.Discriminator.Value != "" && child.Discriminator.Value == discriminatorValue {
			return child
		}
		for _, v := range child.Discriminator.Values {
			if v == discriminatorValue {
				return child
			}
		}
	}
	// Fall back to DefaultVariant
	for i := range f.Children {
		if f.Children[i].Discriminator != nil && f.Children[i].Discriminator.DefaultVariant {
			return &f.Children[i]
		}
	}
	return nil
}

// findPopulatedVariant finds the first populated variant block in an HCL state map.
// Returns the matched FieldSpec and the inner block data. Used by both TypeOneOf and
// discriminated TypeBlockList build direction.
func findPopulatedVariant(f FieldSpec, item map[string]interface{}) (*FieldSpec, map[string]interface{}) {
	for i := range f.Children {
		child := &f.Children[i]
		if child.Type == TypeBlock {
			if inner := getBlockFromMap(item, child.HCLKey); inner != nil {
				return child, inner
			}
		}
	}
	return nil, nil
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

		case TypeJSON:
			if jsonVal != nil {
				if bytes, err := json.Marshal(jsonVal); err == nil {
					result[f.HCLKey] = string(bytes)
				}
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
			if f.Discriminator != nil {
				// Discriminated union per list item — each JSON array element
				// is dispatched to a variant child based on the discriminator.
				if items, ok := jsonVal.([]interface{}); ok {
					list := make([]interface{}, len(items))
					for i, item := range items {
						m, ok := item.(map[string]interface{})
						if !ok {
							list[i] = map[string]interface{}{}
							continue
						}
						if matched := matchOneOfVariant(f, m); matched != nil {
							list[i] = map[string]interface{}{
								matched.HCLKey: []interface{}{FlattenEngineJSON(matched.Children, m)},
							}
						} else {
							list[i] = map[string]interface{}{}
						}
					}
					result[f.HCLKey] = list
				}
			} else {
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

		case TypeOneOf:
			jsonObj, ok := jsonVal.(map[string]interface{})
			if !ok {
				continue
			}
			matchedChild := matchOneOfVariant(f, jsonObj)
			if matchedChild == nil {
				continue
			}
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

// conditionalFormatsExtraFields are the request-level extra fields for widgets
// that support conditional_formats in formula-style requests (query_value, toplist).
var conditionalFormatsExtraFields = []FieldSpec{
	{
		HCLKey:      "conditional_formats",
		Type:        TypeBlockList,
		OmitEmpty:   true,
		Description: "Conditional formats allow you to set the color of your widget content or background depending on the rule applied to your data.",
		Children:    widgetConditionalFormatFields,
	},
}

var scalarWithConditionalFormatsConfig = FormulaRequestConfig{
	ResponseFormat: "scalar",
	StyleFields:    widgetRequestStyleFields,
	IncludeSort:    true,
	ExtraFields:    conditionalFormatsExtraFields,
}

var queryTableFormulaRequestConfig = FormulaRequestConfig{
	ResponseFormat: "scalar",
	ExtraFields:    queryTableRequestExtraFields,
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
	case "query_value", "toplist", "bar_chart":
		return scalarWithConditionalFormatsConfig
	default:
		return scalarFormulaRequestConfig
	}
}

// buildFormulaRequest is the unified formula/query request builder.
// It replaces buildFormulaQueryRequestJSON, buildScalarFormulaQueryRequestJSON,
// and buildQueryTableFormulaRequestJSON.

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

// isFormulaCapableWidget returns true for widget types that support
// formula/query-style requests (response_format: "scalar" or "timeseries").
func isFormulaCapableWidget(jsonType string) bool {
	switch jsonType {
	case "timeseries", "heatmap", "change", "query_value", "toplist", "sunburst", "geomap", "treemap", "bar_chart":
		return true
	}
	return false
}

// buildWidgetSortByJSON builds the JSON for a WidgetSortBy block (sort { count, order_by { ... } }).
// Each order_by element is a discriminated union: formula_sort or group_sort.

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
	// group_by_fields: API returns group_by as a map when group_by_fields variant is used
	if gbf, ok := q["group_by"].(map[string]interface{}); ok {
		flat := map[string]interface{}{}
		if fields, ok := gbf["fields"].([]interface{}); ok {
			strs := make([]string, len(fields))
			for i, f := range fields {
				strs[i] = fmt.Sprintf("%v", f)
			}
			flat["fields"] = strs
		}
		if v, ok := gbf["limit"]; ok {
			switch lv := v.(type) {
			case float64:
				flat["limit"] = int(lv)
			case int:
				flat["limit"] = lv
			}
		}
		if sort, ok := gbf["sort"].(map[string]interface{}); ok {
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
			flat["sort"] = []interface{}{s}
		}
		result["group_by_fields"] = []interface{}{flat}
	}

	if uuids, ok := q["cross_org_uuids"].([]interface{}); ok && len(uuids) > 0 {
		result["cross_org_uuids"] = uuids
	}
	return result
}

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

	// ---- Wildcard requests (multiple formula types + list stream) ----
	if spec.JSONType == "wildcard" {
		if requests, ok := def["requests"].([]interface{}); ok {
			flatRequests := make([]interface{}, len(requests))
			for ri, req := range requests {
				reqMap, ok := req.(map[string]interface{})
				if !ok {
					flatRequests[ri] = map[string]interface{}{}
					continue
				}
				flatRequests[ri] = flattenWildcardRequestJSON(reqMap)
			}
			defState["request"] = flatRequests
		}
	}
}

func flattenQueryTableRequestJSON(req map[string]interface{}) map[string]interface{} {
	_, hasFormulas := req["formulas"]
	_, hasQueries := req["queries"]
	if hasFormulas || hasQueries {
		result := flattenFormulaRequest(req, queryTableFormulaRequestConfig)
		// text_formats (2D array) needs special handling
		if textFormats, ok := req["text_formats"].([]interface{}); ok && len(textFormats) > 0 {
			result["text_formats"] = flattenQueryTableTextFormatsJSON(textFormats)
		}
		return result
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

// FlattenWidgetEngineJSON is the exported entry point for flattening a single widget's JSON map
// into HCL state suitable for setting on a TypeList field in schema.ResourceData.
// Returns nil if the widget type is not recognized.
func FlattenWidgetEngineJSON(widgetData map[string]interface{}) map[string]interface{} {
	return flattenWidgetEngineJSON(widgetData)
}

// ============================================================
// SDKv2 Build Direction (map[string]interface{} input)
// ============================================================

// ============================================================
// Map Helpers — read SDKv2 native types from map[string]interface{}
// ============================================================

// toSlice converts a value to []interface{}. Handles:
//   - []interface{} (TypeList from d.Get)
//   - *schema.Set (TypeSet from d.Get) via the List() interface
func toSlice(v interface{}) []interface{} {
	if items, ok := v.([]interface{}); ok {
		return items
	}
	// *schema.Set implements List() []interface{} — use interface assertion
	// to avoid importing the schema package.
	if lister, ok := v.(interface{ List() []interface{} }); ok {
		return lister.List()
	}
	return nil
}

// getStringFromMap returns a string value from a SDKv2 data map.
// SDKv2 stores TypeString as string (never nil in ResourceData).
func getStringFromMap(data map[string]interface{}, key string) string {
	if data == nil {
		return ""
	}
	if v, ok := data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// getBoolFromMap returns a bool value from a SDKv2 data map.
func getBoolFromMap(data map[string]interface{}, key string) bool {
	if data == nil {
		return false
	}
	if v, ok := data[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

// getIntFromMap returns an int value from a SDKv2 data map.
// SDKv2 stores TypeInt as int.
func getIntFromMap(data map[string]interface{}, key string) int {
	if data == nil {
		return 0
	}
	if v, ok := data[key]; ok {
		switch iv := v.(type) {
		case int:
			return iv
		case float64:
			return int(iv)
		case int64:
			return int(iv)
		}
	}
	return 0
}

// getFloat64FromMap returns a float64 value from a SDKv2 data map.
func getFloat64FromMap(data map[string]interface{}, key string) float64 {
	if data == nil {
		return 0
	}
	if v, ok := data[key]; ok {
		switch fv := v.(type) {
		case float64:
			return fv
		case int:
			return float64(fv)
		}
	}
	return 0
}

// getStringListFromMap returns a []string from a SDKv2 TypeList/TypeSet of strings.
// SDKv2 stores TypeList of strings as []interface{} where each element is string.
// TypeSet values arrive as *schema.Set (which implements List() []interface{}).
func getStringListFromMap(data map[string]interface{}, key string) []string {
	if data == nil {
		return nil
	}
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	items := toSlice(v)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			result = append(result, s)
		} else if item != nil {
			result = append(result, fmt.Sprintf("%v", item))
		}
	}
	return result
}

// getIntListFromMap returns a []int from a SDKv2 TypeList/TypeSet of ints.
func getIntListFromMap(data map[string]interface{}, key string) []int {
	if data == nil {
		return nil
	}
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	items := toSlice(v)
	if len(items) == 0 {
		return nil
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		switch iv := item.(type) {
		case int:
			result = append(result, iv)
		case float64:
			result = append(result, int(iv))
		}
	}
	return result
}

// getBlockFromMap returns the first element of a TypeList/TypeBlock (MaxItems:1) as a map.
// SDKv2 stores TypeList of blocks as []interface{} where each element is map[string]interface{}.
func getBlockFromMap(data map[string]interface{}, key string) map[string]interface{} {
	if data == nil {
		return nil
	}
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	items := toSlice(v)
	if len(items) == 0 {
		return nil
	}
	switch item := items[0].(type) {
	case map[string]interface{}:
		return item
	case nil:
		// SDKv2 returns nil for blocks where all fields are zero-valued.
		return map[string]interface{}{}
	}
	return nil
}

// getBlockListFromMap returns all elements of a TypeList of blocks as []map[string]interface{}.
func getBlockListFromMap(data map[string]interface{}, key string) []map[string]interface{} {
	if data == nil {
		return nil
	}
	v, ok := data[key]
	if !ok || v == nil {
		return nil
	}
	items := toSlice(v)
	if len(items) == 0 {
		return nil
	}
	result := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		switch v := item.(type) {
		case map[string]interface{}:
			result = append(result, v)
		case nil:
			// SDKv2 returns nil for blocks where all fields are at their zero value.
			// Treat as an empty map so Required fields are still serialized.
			result = append(result, map[string]interface{}{})
		}
	}
	return result
}

// ============================================================
// Generic Engine: SDKv2 map → JSON (build direction)
// ============================================================

// BuildEngineJSONFromMap converts a SDKv2 data map to a JSON map using FieldSpec declarations.
// This is the SDKv2 parallel of BuildEngineJSON (which reads from map[string]attr.Value).
func BuildEngineJSONFromMap(data map[string]interface{}, fields []FieldSpec) map[string]interface{} {
	result := map[string]interface{}{}
	for _, f := range fields {
		if f.SchemaOnly {
			continue
		}

		switch f.Type {
		case TypeString:
			strVal := getStringFromMap(data, f.HCLKey)
			if f.OmitEmpty && strVal == "" {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), strVal)

		case TypeBool:
			boolVal := getBoolFromMap(data, f.HCLKey)
			if f.OmitEmpty && !boolVal {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), boolVal)

		case TypeInt:
			intVal := getIntFromMap(data, f.HCLKey)
			if f.OmitEmpty && intVal == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), intVal)

		case TypeFloat:
			floatVal := getFloat64FromMap(data, f.HCLKey)
			if f.OmitEmpty && floatVal == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), floatVal)

		case TypeStringList:
			strs := getStringListFromMap(data, f.HCLKey)
			if f.OmitEmpty && len(strs) == 0 {
				continue
			}
			if strs == nil {
				strs = []string{}
			}
			setAtJSONPath(result, f.effectiveJSONPath(), strs)

		case TypeIntList:
			ints := getIntListFromMap(data, f.HCLKey)
			if f.OmitEmpty && len(ints) == 0 {
				continue
			}
			if ints == nil {
				ints = []int{}
			}
			setAtJSONPath(result, f.effectiveJSONPath(), ints)

		case TypeBlock:
			child := getBlockFromMap(data, f.HCLKey)
			if child == nil {
				continue
			}
			nested := BuildEngineJSONFromMap(child, f.Children)
			if len(nested) == 0 && f.OmitEmpty {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), nested)

		case TypeJSON:
			if str, ok := data[f.HCLKey].(string); ok && str != "" {
				var obj interface{}
				if err := json.Unmarshal([]byte(str), &obj); err == nil {
					setAtJSONPath(result, f.effectiveJSONPath(), obj)
				}
			}

		case TypeBlockList:
			if f.Discriminator != nil {
				// Discriminated union per list item
				items := getBlockListFromMap(data, f.HCLKey)
				built := make([]interface{}, 0, len(items))
				for _, item := range items {
					matched, inner := findPopulatedVariant(f, item)
					if matched == nil {
						built = append(built, map[string]interface{}{})
						continue
					}
					resultJSON := BuildEngineJSONFromMap(inner, matched.Children)
					if matched.Discriminator != nil && matched.Discriminator.Value != "" &&
						f.Discriminator.JSONKey != "" {
						resultJSON[f.Discriminator.JSONKey] = matched.Discriminator.Value
					}
					built = append(built, resultJSON)
				}
				if f.OmitEmpty && len(built) == 0 {
					continue
				}
				setAtJSONPath(result, f.effectiveJSONPath(), built)
			} else {
				items := getBlockListFromMap(data, f.HCLKey)
				built := make([]interface{}, 0, len(items))
				for _, item := range items {
					built = append(built, BuildEngineJSONFromMap(item, f.Children))
				}
				if f.OmitEmpty && len(built) == 0 {
					continue
				}
				setAtJSONPath(result, f.effectiveJSONPath(), built)
			}

		case TypeOneOf:
			outer := getBlockFromMap(data, f.HCLKey)
			if outer == nil {
				if f.OmitEmpty {
					continue
				}
				break
			}
			matched, inner := findPopulatedVariant(f, outer)
			var built map[string]interface{}
			if matched != nil && inner != nil {
				built = BuildEngineJSONFromMap(inner, matched.Children)
			}
			if built == nil {
				if f.OmitEmpty {
					continue
				}
				built = map[string]interface{}{}
			}
			if matched != nil && matched.Discriminator != nil &&
				matched.Discriminator.Value != "" &&
				f.Discriminator != nil && f.Discriminator.JSONKey != "" {
				built[f.Discriminator.JSONKey] = matched.Discriminator.Value
			}
			setAtJSONPath(result, f.effectiveJSONPath(), built)
		}
	}
	return result
}

// ============================================================
// Widget Build — SDKv2 map → JSON
// ============================================================

// BuildWidgetEngineJSONFromMap builds the JSON for a single widget from a SDKv2 data map.
// This is the exported entry point for callers outside the dashboardmapping package
// (e.g., the powerpack SDKv2 resource).
func BuildWidgetEngineJSONFromMap(widget map[string]interface{}) map[string]interface{} {
	return buildWidgetEngineJSONFromMap(widget)
}

// buildWidgetEngineJSONFromMap builds the JSON for a single widget from a SDKv2 map.
// Parallel to buildWidgetEngineJSON but reads from map[string]interface{}.
func buildWidgetEngineJSONFromMap(widget map[string]interface{}) map[string]interface{} {
	for _, spec := range allWidgetSpecs {
		defList, ok := widget[spec.HCLKey].([]interface{})
		if !ok || len(defList) == 0 {
			continue
		}
		defMap, ok := defList[0].(map[string]interface{})
		if !ok {
			continue
		}

		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := BuildEngineJSONFromMap(defMap, allFields)
		defJSON["type"] = spec.JSONType

		// Per-widget post-processing
		buildWidgetPostProcessFromMap(defMap, spec, defJSON)

		widgetJSON := map[string]interface{}{"definition": defJSON}

		// Include widget_layout if present
		if layoutList, ok := widget["widget_layout"].([]interface{}); ok && len(layoutList) > 0 {
			if layoutMap, ok := layoutList[0].(map[string]interface{}); ok {
				layout := map[string]interface{}{}
				for _, key := range []string{"x", "y", "width", "height"} {
					if v := getIntFromMap(layoutMap, key); v != 0 {
						layout[key] = v
					}
				}
				if getBoolFromMap(layoutMap, "is_column_break") {
					layout["is_column_break"] = true
				}
				if len(layout) > 0 {
					widgetJSON["layout"] = layout
				}
			}
		}

		// Include widget id if present
		if idVal := getIntFromMap(widget, "id"); idVal != 0 {
			widgetJSON["id"] = idVal
		}

		return widgetJSON
	}
	return nil
}

// buildWidgetsJSONFromMap builds the "widgets" array from a SDKv2 data map.
func buildWidgetsJSONFromMap(data map[string]interface{}) []interface{} {
	widgetList, ok := data["widget"].([]interface{})
	if !ok || len(widgetList) == 0 {
		return []interface{}{}
	}
	widgets := make([]interface{}, 0, len(widgetList))
	for _, w := range widgetList {
		widgetMap, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		built := buildWidgetEngineJSONFromMap(widgetMap)
		if built != nil {
			widgets = append(widgets, built)
		}
	}
	return widgets
}

// ============================================================
// Widget Post-Processing — SDKv2 map version
// ============================================================

// buildFormulaRequestFromMap builds a formula/query request from a SDKv2 map.
// Parallel to buildFormulaRequest in engine.go.
func buildFormulaRequestFromMap(reqMap map[string]interface{}, cfg FormulaRequestConfig) map[string]interface{} {
	result := map[string]interface{}{}

	// Widget-specific extra fields (e.g. on_right_yaxis, increase_good)
	if len(cfg.ExtraFields) > 0 {
		for k, v := range BuildEngineJSONFromMap(reqMap, cfg.ExtraFields) {
			result[k] = v
		}
	}

	// Formulas
	formulaList := getBlockListFromMap(reqMap, "formula")
	if len(formulaList) > 0 {
		formulas := make([]interface{}, 0, len(formulaList))
		for _, fMap := range formulaList {
			formulas = append(formulas, BuildEngineJSONFromMap(fMap, widgetFormulaFields))
		}
		result["formulas"] = formulas
	}

	// Queries
	queryList := getBlockListFromMap(reqMap, "query")
	if len(queryList) > 0 {
		queries := make([]interface{}, 0, len(queryList))
		for _, qMap := range queryList {
			built := buildQueryFromMapAttrs(qMap)
			if built != nil {
				queries = append(queries, built)
			}
		}
		result["queries"] = queries
	}

	// Request-level style
	if cfg.StyleFields != nil {
		if styleMap := getBlockFromMap(reqMap, "style"); styleMap != nil {
			style := BuildEngineJSONFromMap(styleMap, cfg.StyleFields)
			if len(style) > 0 {
				result["style"] = style
			}
		}
	}

	// Sort
	if cfg.IncludeSort {
		if sortMap := getBlockFromMap(reqMap, "sort"); sortMap != nil {
			sortJSON := buildWidgetSortByJSONFromMap(sortMap)
			if len(sortJSON) > 0 {
				result["sort"] = sortJSON
			}
		}
	}

	// response_format
	if len(formulaList) > 0 || len(queryList) > 0 {
		result["response_format"] = cfg.ResponseFormat
	}

	return result
}

// buildQueryFromMapAttrs dispatches a query map to the appropriate query builder.
// Parallel to buildQueryFromAttrs in engine.go.
func buildQueryFromMapAttrs(qMap map[string]interface{}) map[string]interface{} {
	for _, child := range formulaAndFunctionQueryFields {
		inner := getBlockFromMap(qMap, child.HCLKey)
		if inner == nil {
			continue
		}
		if child.HCLKey == "event_query" {
			return buildEventQueryJSONFromMap(inner)
		}
		return BuildEngineJSONFromMap(inner, child.Children)
	}
	return nil
}

// buildEventQueryJSONFromMap builds the JSON for a formula event_query block from a SDKv2 map.
// Parallel to buildEventQueryJSON in engine.go.
func buildEventQueryJSONFromMap(attrs map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}

	if v := getStringFromMap(attrs, "data_source"); v != "" {
		result["data_source"] = v
	}
	if v := getStringFromMap(attrs, "name"); v != "" {
		result["name"] = v
	}
	if v := getStringFromMap(attrs, "storage"); v != "" {
		result["storage"] = v
	}

	// compute (required)
	if computeMap := getBlockFromMap(attrs, "compute"); computeMap != nil {
		compute := map[string]interface{}{}
		if v := getStringFromMap(computeMap, "aggregation"); v != "" {
			compute["aggregation"] = v
		}
		if v := getIntFromMap(computeMap, "interval"); v != 0 {
			compute["interval"] = v
		}
		if v := getStringFromMap(computeMap, "metric"); v != "" {
			compute["metric"] = v
		}
		result["compute"] = compute
	}

	// search (optional)
	if searchMap := getBlockFromMap(attrs, "search"); searchMap != nil {
		search := map[string]interface{}{}
		if v := getStringFromMap(searchMap, "query"); v != "" {
			search["query"] = v
		}
		result["search"] = search
	}

	// indexes
	if indexes := getStringListFromMap(attrs, "indexes"); indexes != nil {
		result["indexes"] = indexes
	}

	// group_by (optional)
	groupByList := getBlockListFromMap(attrs, "group_by")
	if len(groupByList) > 0 {
		groupBys := make([]interface{}, 0, len(groupByList))
		for _, gbMap := range groupByList {
			gb := map[string]interface{}{}
			if v := getStringFromMap(gbMap, "facet"); v != "" {
				gb["facet"] = v
			}
			if v := getIntFromMap(gbMap, "limit"); v != 0 {
				gb["limit"] = v
			}
			if sortMap := getBlockFromMap(gbMap, "sort"); sortMap != nil {
				sort := map[string]interface{}{}
				if v := getStringFromMap(sortMap, "aggregation"); v != "" {
					sort["aggregation"] = v
				}
				if v := getStringFromMap(sortMap, "metric"); v != "" {
					sort["metric"] = v
				}
				if v := getStringFromMap(sortMap, "order"); v != "" {
					sort["order"] = v
				}
				gb["sort"] = sort
			}
			groupBys = append(groupBys, gb)
		}
		result["group_by"] = groupBys
	}

	// group_by_fields (alternative to group_by — groups by multiple facets)
	if gbfMap := getBlockFromMap(attrs, "group_by_fields"); gbfMap != nil {
		gbf := map[string]interface{}{}
		if fields := getStringListFromMap(gbfMap, "fields"); len(fields) > 0 {
			gbf["fields"] = fields
		}
		if v := getIntFromMap(gbfMap, "limit"); v != 0 {
			gbf["limit"] = v
		}
		if sortMap := getBlockFromMap(gbfMap, "sort"); sortMap != nil {
			sort := map[string]interface{}{}
			if v := getStringFromMap(sortMap, "aggregation"); v != "" {
				sort["aggregation"] = v
			}
			if v := getStringFromMap(sortMap, "metric"); v != "" {
				sort["metric"] = v
			}
			if v := getStringFromMap(sortMap, "order"); v != "" {
				sort["order"] = v
			}
			gbf["sort"] = sort
		}
		result["group_by"] = gbf
	}

	// cross_org_uuids
	if uuids := getStringListFromMap(attrs, "cross_org_uuids"); len(uuids) == 1 && uuids[0] != "" {
		result["cross_org_uuids"] = []string{uuids[0]}
	}

	return result
}

// buildWidgetSortByJSONFromMap builds the JSON for a WidgetSortBy block from a SDKv2 map.
// Parallel to buildWidgetSortByJSON in engine.go.
func buildWidgetSortByJSONFromMap(sortMap map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	if count := getIntFromMap(sortMap, "count"); count > 0 {
		result["count"] = count
	}
	orderByList := getBlockListFromMap(sortMap, "order_by")
	if len(orderByList) > 0 {
		orderBys := make([]interface{}, 0, len(orderByList))
		for _, obMap := range orderByList {
			entry := map[string]interface{}{}
			if fsMap := getBlockFromMap(obMap, "formula_sort"); fsMap != nil {
				entry["type"] = "formula"
				if idx := getIntFromMap(fsMap, "index"); idx != 0 {
					entry["index"] = idx
				}
				if ord := getStringFromMap(fsMap, "order"); ord != "" {
					entry["order"] = ord
				}
			} else if gsMap := getBlockFromMap(obMap, "group_sort"); gsMap != nil {
				entry["type"] = "group"
				if name := getStringFromMap(gsMap, "name"); name != "" {
					entry["name"] = name
				}
				if ord := getStringFromMap(gsMap, "order"); ord != "" {
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

// buildQueryTableRequestsJSONFromMap builds the "requests" array for query_table from a SDKv2 map.
// Parallel to buildQueryTableRequestsJSON in engine.go.
func buildQueryTableRequestsJSONFromMap(defMap map[string]interface{}) []interface{} {
	requestList := getBlockListFromMap(defMap, "request")
	if len(requestList) == 0 {
		return []interface{}{}
	}
	requests := make([]interface{}, 0, len(requestList))
	for _, reqMap := range requestList {
		formulaCount := len(getBlockListFromMap(reqMap, "formula"))
		queryCount := len(getBlockListFromMap(reqMap, "query"))
		if formulaCount > 0 || queryCount > 0 {
			req := buildFormulaRequestFromMap(reqMap, queryTableFormulaRequestConfig)
			buildQueryTableTextFormatsJSONFromMap(reqMap, req)
			requests = append(requests, req)
		} else {
			req := BuildEngineJSONFromMap(reqMap, queryTableOldRequestFields)
			buildQueryTableTextFormatsJSONFromMap(reqMap, req)
			requests = append(requests, req)
		}
	}
	return requests
}

// buildQueryTableTextFormatsJSONFromMap handles the text_formats 2D array for query_table.
// Parallel to buildQueryTableTextFormatsJSON in engine.go.
func buildQueryTableTextFormatsJSONFromMap(reqMap map[string]interface{}, req map[string]interface{}) {
	tfList := getBlockListFromMap(reqMap, "text_formats")
	if len(tfList) == 0 {
		return
	}
	textFormats := make([]interface{}, 0, len(tfList))
	for _, tfMap := range tfList {
		innerList := getBlockListFromMap(tfMap, "text_format")
		innerFormats := make([]interface{}, 0, len(innerList))
		for _, innerMap := range innerList {
			rule := map[string]interface{}{}
			if matchMap := getBlockFromMap(innerMap, "match"); matchMap != nil {
				match := map[string]interface{}{}
				if v := getStringFromMap(matchMap, "type"); v != "" {
					match["type"] = v
				}
				if v := getStringFromMap(matchMap, "value"); v != "" {
					match["value"] = v
				}
				rule["match"] = match
			}
			if v := getStringFromMap(innerMap, "palette"); v != "" {
				rule["palette"] = v
			}
			if v := getStringFromMap(innerMap, "custom_bg_color"); v != "" {
				rule["custom_bg_color"] = v
			}
			if v := getStringFromMap(innerMap, "custom_fg_color"); v != "" {
				rule["custom_fg_color"] = v
			}
			if replaceMap := getBlockFromMap(innerMap, "replace"); replaceMap != nil {
				replace := map[string]interface{}{}
				if v := getStringFromMap(replaceMap, "type"); v != "" {
					replace["type"] = v
				}
				if v := getStringFromMap(replaceMap, "with"); v != "" {
					replace["with"] = v
				}
				if v := getStringFromMap(replaceMap, "substring"); v != "" {
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

// buildScatterplotTableJSONFromMap builds the scatterplot "requests.table" from a SDKv2 map.
// Parallel to buildScatterplotTableJSON in engine.go.
func buildScatterplotTableJSONFromMap(defMap map[string]interface{}, defJSON map[string]interface{}) {
	reqMap := getBlockFromMap(defMap, "request")
	if reqMap == nil {
		return
	}
	tableMap := getBlockFromMap(reqMap, "scatterplot_table")
	if tableMap == nil {
		return
	}

	tableObj := map[string]interface{}{}

	formulaList := getBlockListFromMap(tableMap, "formula")
	if len(formulaList) > 0 {
		formulas := make([]interface{}, 0, len(formulaList))
		for _, fMap := range formulaList {
			formula := map[string]interface{}{}
			if expr := getStringFromMap(fMap, "formula_expression"); expr != "" {
				formula["formula"] = expr
			}
			if dim := getStringFromMap(fMap, "dimension"); dim != "" {
				formula["dimension"] = dim
			}
			if alias := getStringFromMap(fMap, "alias"); alias != "" {
				formula["alias"] = alias
			}
			formulas = append(formulas, formula)
		}
		tableObj["formulas"] = formulas
	}

	queryList := getBlockListFromMap(tableMap, "query")
	if len(queryList) > 0 {
		queries := make([]interface{}, 0, len(queryList))
		for _, qMap := range queryList {
			built := buildQueryFromMapAttrs(qMap)
			if built != nil {
				queries = append(queries, built)
			}
		}
		tableObj["queries"] = queries
	}

	if len(formulaList) > 0 || len(queryList) > 0 {
		tableObj["response_format"] = "scalar"
	}

	if len(tableObj) > 0 {
		requestsObj, ok := defJSON["requests"].(map[string]interface{})
		if !ok {
			requestsObj = map[string]interface{}{}
			defJSON["requests"] = requestsObj
		}
		requestsObj["table"] = tableObj
	}
}

// buildSplitGraphSourceWidgetJSONFromMap builds the source_widget_definition JSON for split_graph.
// Parallel to buildSplitGraphSourceWidgetJSON in engine.go.
func buildSplitGraphSourceWidgetJSONFromMap(defMap map[string]interface{}) map[string]interface{} {
	srcWidgetMap := getBlockFromMap(defMap, "source_widget_definition")
	if srcWidgetMap == nil {
		return nil
	}
	for _, spec := range allWidgetSpecs {
		innerList, ok := srcWidgetMap[spec.HCLKey].([]interface{})
		if !ok || len(innerList) == 0 {
			continue
		}
		innerMap, ok := innerList[0].(map[string]interface{})
		if !ok {
			continue
		}
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := BuildEngineJSONFromMap(innerMap, allFields)
		defJSON["type"] = spec.JSONType
		buildWidgetPostProcessFromMap(innerMap, spec, defJSON)
		return defJSON
	}
	return nil
}

// buildSplitConfigStaticSplitsJSONFromMap builds the static_splits 2D JSON array.
// Parallel to buildSplitConfigStaticSplitsJSON in engine.go.
func buildSplitConfigStaticSplitsJSONFromMap(splitConfigMap map[string]interface{}) []interface{} {
	staticSplitList := getBlockListFromMap(splitConfigMap, "static_splits")
	if len(staticSplitList) == 0 {
		return nil
	}
	result := make([]interface{}, 0, len(staticSplitList))
	for _, ssMap := range staticSplitList {
		vectorList := getBlockListFromMap(ssMap, "split_vector")
		innerArray := make([]interface{}, 0, len(vectorList))
		for _, vMap := range vectorList {
			vec := map[string]interface{}{}
			if v := getStringFromMap(vMap, "tag_key"); v != "" {
				vec["tag_key"] = v
			}
			tagValues := getStringListFromMap(vMap, "tag_values")
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

// buildGroupWidgetsJSONFromMap builds the "widgets" array for a group widget.
// Parallel to buildGroupWidgetsJSON in engine.go.
func buildGroupWidgetsJSONFromMap(defMap map[string]interface{}) []interface{} {
	widgetList, ok := defMap["widget"].([]interface{})
	if !ok || len(widgetList) == 0 {
		return []interface{}{}
	}
	widgets := make([]interface{}, 0, len(widgetList))
	for _, w := range widgetList {
		widgetMap, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		built := buildWidgetEngineJSONFromMap(widgetMap)
		if built == nil {
			built = map[string]interface{}{}
		}
		widgets = append(widgets, built)
	}
	return widgets
}

// buildWidgetPostProcessFromMap runs all per-widget post-processing in the build direction.
// Parallel to buildWidgetPostProcess in engine.go but reads from map[string]interface{}.
func buildWidgetPostProcessFromMap(defMap map[string]interface{}, spec WidgetSpec, defJSON map[string]interface{}) {
	// ---- Formula/query blocks ----
	if isFormulaCapableWidget(spec.JSONType) {
		requestList := getBlockListFromMap(defMap, "request")
		if len(requestList) > 0 {
			requests := make([]interface{}, 0, len(requestList))
			for ri, reqMap := range requestList {
				formulaCount := len(getBlockListFromMap(reqMap, "formula"))
				queryCount := len(getBlockListFromMap(reqMap, "query"))
				if formulaCount > 0 || queryCount > 0 {
					requests = append(requests, buildFormulaRequestFromMap(reqMap, formulaRequestConfigForWidget(spec.JSONType)))
				} else {
					// Old-style request — already built by FieldSpec engine
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
		buildScatterplotTableJSONFromMap(defMap, defJSON)
	}

	// ---- Treemap color_by ----
	if spec.JSONType == "treemap" {
		defJSON["color_by"] = "user"
	}

	// ---- Query table requests override ----
	if spec.JSONType == "query_table" {
		defJSON["requests"] = buildQueryTableRequestsJSONFromMap(defMap)
	}

	// ---- Split graph source widget + static_splits ----
	if spec.JSONType == "split_group" {
		srcDefJSON := buildSplitGraphSourceWidgetJSONFromMap(defMap)
		if srcDefJSON != nil {
			defJSON["source_widget_definition"] = srcDefJSON
		}
		if splitConfigMap, ok := defJSON["split_config"].(map[string]interface{}); ok {
			if splitConfigHCL := getBlockFromMap(defMap, "split_config"); splitConfigHCL != nil {
				staticSplits := buildSplitConfigStaticSplitsJSONFromMap(splitConfigHCL)
				if len(staticSplits) > 0 {
					splitConfigMap["static_splits"] = staticSplits
				}
			}
		}
	}

	// ---- Group nested widgets ----
	if spec.JSONType == "group" {
		defJSON["widgets"] = buildGroupWidgetsJSONFromMap(defMap)
	}

	// ---- Funnel request_type injection ----
	if spec.JSONType == "funnel" {
		if requests, ok := defJSON["requests"].([]interface{}); ok {
			for _, req := range requests {
				if reqMap, ok := req.(map[string]interface{}); ok {
					reqMap["request_type"] = "funnel"
				}
			}
		}
	}

	// ---- Wildcard requests ----
	if spec.JSONType == "wildcard" {
		defJSON["requests"] = buildWildcardRequestsJSONFromMap(defMap)
	}
}

// ============================================================
// Wildcard Widget Build/Flatten Helpers
// ============================================================

// wildcardFormulaRequestConfig is the FormulaRequestConfig for wildcard formula requests.
// ResponseFormat is empty — it's read from the request data, not injected.
var wildcardScalarConfig = FormulaRequestConfig{
	ResponseFormat: "scalar",
	StyleFields:    widgetRequestStyleFields,
	IncludeSort:    true,
}

var wildcardTimeseriesConfig = FormulaRequestConfig{
	ResponseFormat: "timeseries",
	StyleFields:    widgetRequestStyleFields,
	ExtraFields: []FieldSpec{
		{HCLKey: "display_type", Type: TypeString, OmitEmpty: true,
			Description: "How the data points are displayed on the graph."},
	},
}

// flattenWildcardRequestJSON flattens a single wildcard request JSON object.
// Dispatches based on response_format to determine which flattening style to use.
func flattenWildcardRequestJSON(req map[string]interface{}) map[string]interface{} {
	responseFormat, _ := req["response_format"].(string)
	var result map[string]interface{}
	switch responseFormat {
	case "timeseries":
		result = flattenFormulaRequest(req, wildcardTimeseriesConfig)
	case "event_list":
		// List stream request
		result = FlattenEngineJSON(listStreamRequestFields, req)
	default:
		// scalar (TreeMap-style) or distribution
		result = flattenFormulaRequest(req, wildcardScalarConfig)
	}
	// Always preserve response_format in state
	if responseFormat != "" {
		result["response_format"] = responseFormat
	}
	return result
}

// buildWildcardRequestsJSONFromMap builds the wildcard requests JSON array from HCL.
func buildWildcardRequestsJSONFromMap(defMap map[string]interface{}) []interface{} {
	requestList := getBlockListFromMap(defMap, "request")
	requests := make([]interface{}, 0, len(requestList))
	for _, reqMap := range requestList {
		responseFormat, _ := reqMap["response_format"].(string)
		formulaCount := len(getBlockListFromMap(reqMap, "formula"))
		queryCount := len(getBlockListFromMap(reqMap, "query"))

		if formulaCount > 0 || queryCount > 0 {
			var cfg FormulaRequestConfig
			switch responseFormat {
			case "timeseries":
				cfg = wildcardTimeseriesConfig
			default:
				cfg = wildcardScalarConfig
			}
			requests = append(requests, buildFormulaRequestFromMap(reqMap, cfg))
		} else if responseFormat == "event_list" {
			// List stream request
			reqJSON := BuildEngineJSONFromMap(reqMap, listStreamRequestFields)
			reqJSON["response_format"] = "event_list"
			requests = append(requests, reqJSON)
		}
	}
	return requests
}

// ============================================================
// Dashboard-Level Build and Marshal
// ============================================================

// BuildDashboardEngineJSONFromMap builds the full dashboard JSON body from a SDKv2 data map.
// Parallel to BuildDashboardEngineJSON in engine.go.
func BuildDashboardEngineJSONFromMap(data map[string]interface{}, id string) map[string]interface{} {
	result := BuildEngineJSONFromMap(data, DashboardTopLevelFields)

	// Both POST and PUT bodies need the id set.
	result["id"] = id

	// Build widgets with type dispatch.
	widgets := buildWidgetsJSONFromMap(data)
	result["widgets"] = widgets

	// Build tabs with @N → widget ID resolution.
	if tabs := buildTabsJSONFromMap(data, widgets, id); len(tabs) > 0 {
		result["tabs"] = tabs
	}

	return result
}

// buildTabsJSONFromMap builds the "tabs" array from HCL tab blocks.
// Passes @N widget references through to the API (the API resolves them server-side).
// Generates deterministic UUID v5 IDs for tabs that don't have one (matching v1 behavior).
func buildTabsJSONFromMap(data map[string]interface{}, builtWidgets []interface{}, dashID string) []interface{} {
	tabList := getBlockListFromMap(data, "tab")
	if len(tabList) == 0 {
		return nil
	}

	tabs := make([]interface{}, 0, len(tabList))
	for i, tabMap := range tabList {
		tab := map[string]interface{}{}

		// Tab name (required)
		if name, ok := tabMap["name"].(string); ok && name != "" {
			tab["name"] = name
		}

		// Tab ID: use existing or generate deterministic UUID v5 (matches v1 behavior)
		if id, ok := tabMap["id"].(string); ok && id != "" {
			tab["id"] = id
		} else {
			tab["id"] = uuid.NewSHA1(uuid.Nil, []byte(fmt.Sprintf("tab-%d", i))).String()
		}

		// Pass @N widget references through as-is — the API resolves them server-side
		if widgetIDs, ok := tabMap["widget_ids"].([]interface{}); ok {
			tab["widget_ids"] = widgetIDs
		}

		tabs = append(tabs, tab)
	}
	return tabs
}

// MarshalDashboardJSONFromMap marshals the dashboard JSON body from a SDKv2 data map.
// Parallel to MarshalDashboardJSON in engine.go.
// The trailing newline matches behavior of json.NewEncoder used when cassettes were recorded.
func MarshalDashboardJSONFromMap(data map[string]interface{}, id string) (string, error) {
	body, err := json.Marshal(BuildDashboardEngineJSONFromMap(data, id))
	if err != nil {
		return "", fmt.Errorf("error marshaling dashboard JSON: %s", err)
	}
	return string(body) + "\n", nil
}

// ============================================================
// Flatten — SDKv2 compatible (reuses existing FlattenEngineJSON)
// ============================================================

// FlattenWidgetsForSDKv2 converts the "widgets" array from a dashboard API response
// into a []interface{} suitable for d.Set("widget", ...).
// It reuses the existing flattenWidgetEngineJSON function which already returns
// map[string]interface{} compatible with SDKv2's d.Set().
func FlattenWidgetsForSDKv2(apiWidgets []interface{}) []interface{} {
	result := make([]interface{}, 0, len(apiWidgets))
	for _, w := range apiWidgets {
		widgetData, ok := w.(map[string]interface{})
		if !ok {
			continue
		}

		// Flatten the widget definition using the existing engine function.
		flattened := FlattenWidgetEngineJSON(widgetData)
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

		result = append(result, flattened)
	}
	return result
}

// ============================================================
// Flatten Tabs — reverse-map widget IDs to @N positional references
// ============================================================

// FlattenTabs converts the "tabs" array from a dashboard API response
// into a []interface{} suitable for d.Set("tab", ...).
// Widget IDs in each tab are reverse-mapped to @N positional references
// based on the order of widgets in the flattened widget list.
func FlattenTabs(apiTabs []interface{}, apiWidgets []interface{}) []interface{} {
	// Build widgetID → 1-indexed position map.
	// Widget IDs are large integers that arrive as float64 from JSON;
	// normalize to int64 strings for consistent lookup.
	widgetIDToPosition := map[string]int{}
	for i, w := range apiWidgets {
		wMap, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		// Widget IDs can be at the top level or inside definition.
		var wID interface{}
		if id, ok := wMap["id"]; ok {
			wID = id
		} else if def, ok := wMap["definition"].(map[string]interface{}); ok {
			wID = def["id"]
		}
		if wID != nil {
			widgetIDToPosition[normalizeNumericID(wID)] = i + 1
		}
	}

	result := make([]interface{}, 0, len(apiTabs))
	for _, t := range apiTabs {
		tabData, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		flattened := map[string]interface{}{}

		if id, ok := tabData["id"].(string); ok {
			flattened["id"] = id
		}
		if name, ok := tabData["name"].(string); ok {
			flattened["name"] = name
		}

		// Convert widget IDs to @N references.
		if widgetIDs, ok := tabData["widget_ids"].([]interface{}); ok {
			refs := make([]interface{}, 0, len(widgetIDs))
			for _, wID := range widgetIDs {
				if pos, ok := widgetIDToPosition[normalizeNumericID(wID)]; ok {
					refs = append(refs, fmt.Sprintf("@%d", pos))
				}
			}
			flattened["widget_ids"] = refs
		}

		result = append(result, flattened)
	}
	return result
}

// normalizeNumericID converts a JSON numeric value (typically float64) to a
// consistent integer string representation for use as a map key.
func normalizeNumericID(v interface{}) string {
	switch id := v.(type) {
	case float64:
		return strconv.FormatInt(int64(id), 10)
	case int:
		return strconv.Itoa(id)
	case int64:
		return strconv.FormatInt(id, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ============================================================
// Validation — check FieldSpec ConflictsWith at plan time
// ============================================================

// ValidateWidgetConflicts walks the widget tree and checks ConflictsWith
// constraints on request-level fields. Returns a list of human-readable
// error strings for any violations found.
//
// This is driven by the ConflictsWith annotations on FieldSpec declarations
// (e.g., the "q" field conflicts with "query" and "formula").
func ValidateWidgetConflicts(data map[string]interface{}) []string {
	widgetList, ok := data["widget"].([]interface{})
	if !ok {
		return nil
	}
	var errs []string
	for wi, w := range widgetList {
		wMap, ok := w.(map[string]interface{})
		if !ok {
			continue
		}
		for _, spec := range allWidgetSpecs {
			defList, ok := wMap[spec.HCLKey].([]interface{})
			if !ok || len(defList) == 0 {
				continue
			}
			defMap, ok := defList[0].(map[string]interface{})
			if !ok {
				continue
			}

			// Check ConflictsWith constraints on request fields
			var requestFields []FieldSpec
			for _, f := range spec.Fields {
				if f.HCLKey == "request" {
					requestFields = f.Children
					break
				}
			}
			if requestFields != nil {
				reqList := getBlockListFromMap(defMap, "request")
				for ri, reqMap := range reqList {
					for _, f := range requestFields {
						if len(f.ConflictsWith) == 0 {
							continue
						}
						if !fieldIsSetInMap(reqMap, f) {
							continue
						}
						for _, conflicting := range f.ConflictsWith {
							for _, cf := range requestFields {
								if cf.HCLKey == conflicting && fieldIsSetInMap(reqMap, cf) {
									errs = append(errs, fmt.Sprintf(
										"widget[%d].%s.request[%d]: %q conflicts with %q — use one query style per request, not both",
										wi, spec.HCLKey, ri, f.HCLKey, cf.HCLKey,
									))
								}
							}
						}
					}
				}
			}

			// Recurse into group/split_graph nested widgets
			if spec.HCLKey == "group_definition" || spec.HCLKey == "split_graph_definition" {
				if nestedErrs := ValidateWidgetConflicts(defMap); len(nestedErrs) > 0 {
					for _, e := range nestedErrs {
						errs = append(errs, fmt.Sprintf("widget[%d].%s.%s", wi, spec.HCLKey, e))
					}
				}
			}
		}
	}
	return errs
}

// fieldIsSetInMap returns true if the field has a non-zero value in the map.
func fieldIsSetInMap(data map[string]interface{}, f FieldSpec) bool {
	v, ok := data[f.HCLKey]
	if !ok || v == nil {
		return false
	}
	switch f.Type {
	case TypeString:
		return v.(string) != ""
	case TypeBlockList:
		if list, ok := v.([]interface{}); ok {
			return len(list) > 0
		}
	case TypeBlock:
		if list, ok := v.([]interface{}); ok {
			return len(list) > 0
		}
	}
	return false
}
