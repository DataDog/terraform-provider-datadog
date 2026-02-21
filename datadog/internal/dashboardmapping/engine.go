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

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

	// ValidateDiag: arbitrary validator (escape hatch for complex cases).
	// Use ValidValues for simple enums; ValidateDiag for everything else.
	ValidateDiag schema.SchemaValidateDiagFunc

	// ConflictsWith: field paths this field conflicts with.
	// Populated post-generation for the rare fields that need it.
	ConflictsWith []string

	// UseSet: use schema.TypeSet instead of schema.TypeList for list fields.
	// Rare — only for fields that require set semantics.
	UseSet bool
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
// Generic Engine: HCL → JSON (build direction)
// ============================================================

// BuildEngineJSON converts schema.ResourceData at the given HCL path prefix to a JSON map.
// hclPrefix is the dotted path to the parent block, e.g., "widget.0.timeseries_definition.0"
func BuildEngineJSON(d *schema.ResourceData, hclPrefix string, fields []FieldSpec) map[string]interface{} {
	result := map[string]interface{}{}
	for _, f := range fields {
		var hclPath string
		if hclPrefix == "" {
			hclPath = f.HCLKey
		} else {
			hclPath = hclPrefix + "." + f.HCLKey
		}

		switch f.Type {
		case TypeString:
			val, ok := d.GetOk(hclPath)
			if !ok {
				if f.OmitEmpty {
					continue
				}
				// Not OmitEmpty: include zero value
				val = ""
			}
			strVal := fmt.Sprintf("%v", val)
			if f.OmitEmpty && strVal == "" {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), strVal)

		case TypeBool:
			raw := d.Get(hclPath)
			boolVal, ok := raw.(bool)
			if !ok {
				boolVal = false
			}
			if f.OmitEmpty && !boolVal {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), boolVal)

		case TypeInt:
			raw := d.Get(hclPath)
			var intVal int
			switch v := raw.(type) {
			case int:
				intVal = v
			case float64:
				intVal = int(v)
			}
			if f.OmitEmpty && intVal == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), intVal)

		case TypeFloat:
			raw := d.Get(hclPath)
			var floatVal float64
			switch v := raw.(type) {
			case float64:
				floatVal = v
			case int:
				floatVal = float64(v)
			}
			if f.OmitEmpty && floatVal == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), floatVal)

		case TypeStringList:
			// Handle both TypeList and TypeSet of strings
			raw := d.Get(hclPath)
			var strs []string
			switch v := raw.(type) {
			case []interface{}:
				strs = make([]string, len(v))
				for i, s := range v {
					strs[i] = fmt.Sprintf("%v", s)
				}
			case *schema.Set:
				list := v.List()
				strs = make([]string, len(list))
				for i, s := range list {
					strs[i] = fmt.Sprintf("%v", s)
				}
			}
			if f.OmitEmpty && len(strs) == 0 {
				continue
			}
			if strs == nil {
				strs = []string{}
			}
			setAtJSONPath(result, f.effectiveJSONPath(), strs)

		case TypeBlock:
			count, _ := d.Get(hclPath + ".#").(int)
			if count == 0 {
				continue
			}
			nested := BuildEngineJSON(d, hclPath+".0", f.Children)
			if len(nested) == 0 && f.OmitEmpty {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), nested)

		case TypeBlockList:
			count, _ := d.Get(hclPath + ".#").(int)
			items := make([]interface{}, count)
			for i := 0; i < count; i++ {
				items[i] = BuildEngineJSON(d, fmt.Sprintf("%s.%d", hclPath, i), f.Children)
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
			result[f.HCLKey] = toStringSliceFromInterface(jsonVal)

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
// widgetPath is the dotted HCL path to the widget block, e.g. "widget.0" or
// "widget.0.group_definition.0.widget.2".
func buildWidgetEngineJSON(d *schema.ResourceData, widgetPath string) map[string]interface{} {
	for _, spec := range allWidgetSpecs {
		defPath := fmt.Sprintf("%s.%s.#", widgetPath, spec.HCLKey)
		count, _ := d.Get(defPath).(int)
		if count == 0 {
			continue
		}
		defHCLPath := fmt.Sprintf("%s.%s.0", widgetPath, spec.HCLKey)
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := BuildEngineJSON(d, defHCLPath, allFields)
		defJSON["type"] = spec.JSONType

		// Per-widget post-processing: formula/query blocks, type-specific fixups,
		// and Batch C complex widget handling — all consolidated in one hook.
		buildWidgetPostProcess(d, spec, defHCLPath, defJSON)

		widgetJSON := map[string]interface{}{"definition": defJSON}

		// Include widget_layout if present (required for 'free' layout dashboards).
		// HCL: widget_layout { x, y, width, height, is_column_break }
		// JSON: layout { x, y, width, height, is_column_break }
		layoutPath := widgetPath + ".widget_layout.#"
		if layoutCount, _ := d.Get(layoutPath).(int); layoutCount > 0 {
			lPath := widgetPath + ".widget_layout.0"
			layout := map[string]interface{}{}
			if v, ok := d.GetOk(lPath + ".x"); ok {
				layout["x"] = v
			}
			if v, ok := d.GetOk(lPath + ".y"); ok {
				layout["y"] = v
			}
			if v, ok := d.GetOk(lPath + ".width"); ok {
				layout["width"] = v
			}
			if v, ok := d.GetOk(lPath + ".height"); ok {
				layout["height"] = v
			}
			if v, ok := d.Get(lPath + ".is_column_break").(bool); ok && v {
				layout["is_column_break"] = v
			}
			widgetJSON["layout"] = layout
		}

		return widgetJSON
	}
	return nil
}

// buildFormulaQueryRequestJSON builds the JSON for a new-style formula/query request.
// These requests use `formulas`, `queries`, and `response_format` instead of the
// old-style `q`, `log_query`, etc.
func buildFormulaQueryRequestJSON(d *schema.ResourceData, reqPath string) map[string]interface{} {
	result := map[string]interface{}{}

	// on_right_yaxis is always emitted (OmitEmpty: false)
	onRightYaxis, _ := d.Get(reqPath + ".on_right_yaxis").(bool)
	result["on_right_yaxis"] = onRightYaxis

	// display_type (optional)
	if dt, ok := d.GetOk(reqPath + ".display_type"); ok {
		result["display_type"] = fmt.Sprintf("%v", dt)
	}

	// formulas — use widgetFormulaFields engine for standard fields.
	// number_format is excluded from widgetFormulaFields (polymorphic structure);
	// it is not present on timeseries formula requests.
	formulaCount, _ := d.Get(reqPath + ".formula.#").(int)
	if formulaCount > 0 {
		formulas := make([]interface{}, formulaCount)
		for fi := 0; fi < formulaCount; fi++ {
			fPath := fmt.Sprintf("%s.formula.%d", reqPath, fi)
			formulas[fi] = BuildEngineJSON(d, fPath, widgetFormulaFields)
		}
		result["formulas"] = formulas
	}

	// queries (polymorphic: each query block has one of metric_query, event_query, etc.)
	queryCount, _ := d.Get(reqPath + ".query.#").(int)
	if queryCount > 0 {
		queries := make([]interface{}, queryCount)
		for qi := 0; qi < queryCount; qi++ {
			qPath := fmt.Sprintf("%s.query.%d", reqPath, qi)
			var queryJSON map[string]interface{}

			if n, _ := d.Get(qPath + ".metric_query.#").(int); n > 0 {
				queryJSON = buildMetricQueryJSON(d, qPath+".metric_query.0")
			} else if n, _ := d.Get(qPath + ".event_query.#").(int); n > 0 {
				queryJSON = buildEventQueryJSON(d, qPath+".event_query.0")
			} else if n, _ := d.Get(qPath + ".process_query.#").(int); n > 0 {
				queryJSON = buildFormulaProcessQueryJSON(d, qPath+".process_query.0")
			} else if n, _ := d.Get(qPath + ".slo_query.#").(int); n > 0 {
				queryJSON = buildSLOQueryJSON(d, qPath+".slo_query.0")
			} else if n, _ := d.Get(qPath + ".cloud_cost_query.#").(int); n > 0 {
				queryJSON = buildCloudCostQueryJSON(d, qPath+".cloud_cost_query.0")
			}
			queries[qi] = queryJSON
		}
		result["queries"] = queries
	}

	// style (optional, present on timeseries formula requests at the request level)
	// This is request-level style (palette, line_type, line_width), distinct from
	// the per-formula style (palette, palette_index) in widgetFormulaFields.
	if styleCount, _ := d.Get(reqPath + ".style.#").(int); styleCount > 0 {
		style := map[string]interface{}{}
		if v, ok := d.GetOk(reqPath + ".style.0.palette"); ok && v != "" {
			style["palette"] = fmt.Sprintf("%v", v)
		}
		if v, ok := d.GetOk(reqPath + ".style.0.line_type"); ok && v != "" {
			style["line_type"] = fmt.Sprintf("%v", v)
		}
		if v, ok := d.GetOk(reqPath + ".style.0.line_width"); ok && v != "" {
			style["line_width"] = fmt.Sprintf("%v", v)
		}
		if len(style) > 0 {
			result["style"] = style
		}
	}

	// response_format is always "timeseries" for formula/query requests
	if formulaCount > 0 || queryCount > 0 {
		result["response_format"] = "timeseries"
	}

	return result
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
func buildScalarFormulaQueryRequestJSON(d *schema.ResourceData, reqPath string, widgetType string) map[string]interface{} {
	result := map[string]interface{}{}

	// change widget has increase_good and show_present even in formula requests
	if widgetType == "change" {
		increaseGood, _ := d.Get(reqPath + ".increase_good").(bool)
		result["increase_good"] = increaseGood
		showPresent, _ := d.Get(reqPath + ".show_present").(bool)
		result["show_present"] = showPresent
		if v, ok := d.GetOk(reqPath + ".change_type"); ok {
			result["change_type"] = fmt.Sprintf("%v", v)
		}
		if v, ok := d.GetOk(reqPath + ".compare_to"); ok {
			result["compare_to"] = fmt.Sprintf("%v", v)
		}
		if v, ok := d.GetOk(reqPath + ".order_by"); ok {
			result["order_by"] = fmt.Sprintf("%v", v)
		}
		if v, ok := d.GetOk(reqPath + ".order_dir"); ok {
			result["order_dir"] = fmt.Sprintf("%v", v)
		}
	}

	// formulas — use widgetFormulaFields engine for standard fields.
	// number_format is excluded from widgetFormulaFields (polymorphic structure);
	// it is not present on scalar formula requests for non-query_table widgets.
	formulaCount, _ := d.Get(reqPath + ".formula.#").(int)
	if formulaCount > 0 {
		formulas := make([]interface{}, formulaCount)
		for fi := 0; fi < formulaCount; fi++ {
			fPath := fmt.Sprintf("%s.formula.%d", reqPath, fi)
			formulas[fi] = BuildEngineJSON(d, fPath, widgetFormulaFields)
		}
		result["formulas"] = formulas
	}

	// queries (polymorphic)
	queryCount, _ := d.Get(reqPath + ".query.#").(int)
	if queryCount > 0 {
		queries := make([]interface{}, queryCount)
		for qi := 0; qi < queryCount; qi++ {
			qPath := fmt.Sprintf("%s.query.%d", reqPath, qi)
			var queryJSON map[string]interface{}

			if n, _ := d.Get(qPath + ".metric_query.#").(int); n > 0 {
				queryJSON = buildMetricQueryJSON(d, qPath+".metric_query.0")
			} else if n, _ := d.Get(qPath + ".event_query.#").(int); n > 0 {
				queryJSON = buildEventQueryJSON(d, qPath+".event_query.0")
			} else if n, _ := d.Get(qPath + ".process_query.#").(int); n > 0 {
				queryJSON = buildFormulaProcessQueryJSON(d, qPath+".process_query.0")
			} else if n, _ := d.Get(qPath + ".slo_query.#").(int); n > 0 {
				queryJSON = buildSLOQueryJSON(d, qPath+".slo_query.0")
			} else if n, _ := d.Get(qPath + ".cloud_cost_query.#").(int); n > 0 {
				queryJSON = buildCloudCostQueryJSON(d, qPath+".cloud_cost_query.0")
			}
			queries[qi] = queryJSON
		}
		result["queries"] = queries
	}

	// style (optional, present on sunburst/toplist/etc. formula requests at the request level)
	// This is request-level style (just palette for scalar widgets), distinct from
	// the per-formula style (palette, palette_index) in widgetFormulaFields.
	if styleCount, _ := d.Get(reqPath + ".style.#").(int); styleCount > 0 {
		style := map[string]interface{}{}
		if v, ok := d.GetOk(reqPath + ".style.0.palette"); ok && v != "" {
			style["palette"] = fmt.Sprintf("%v", v)
		}
		if len(style) > 0 {
			result["style"] = style
		}
	}

	// response_format depends on widget type
	if formulaCount > 0 || queryCount > 0 {
		result["response_format"] = formulaResponseFormat(widgetType)
	}

	return result
}

// buildMetricQueryJSON builds the JSON for a formula metric_query block.
func buildMetricQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	} else {
		result["data_source"] = "metrics" // default
	}
	if v, ok := d.GetOk(path + ".query"); ok {
		result["query"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".aggregator"); ok && v != "" {
		result["aggregator"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".semantic_mode"); ok && v != "" {
		result["semantic_mode"] = fmt.Sprintf("%v", v)
	}
	// cross_org_uuids: list of one string
	if raw := d.Get(path + ".cross_org_uuids"); raw != nil {
		if items, ok := raw.([]interface{}); ok && len(items) == 1 {
			if s, ok := items[0].(string); ok && s != "" {
				result["cross_org_uuids"] = []string{s}
			}
		}
	}
	return result
}

// buildEventQueryJSON builds the JSON for a formula event_query block.
func buildEventQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".storage"); ok && v != "" {
		result["storage"] = fmt.Sprintf("%v", v)
	}

	// compute (required)
	computeCount, _ := d.Get(path + ".compute.#").(int)
	if computeCount > 0 {
		cPath := path + ".compute.0"
		compute := map[string]interface{}{}
		if v, ok := d.GetOk(cPath + ".aggregation"); ok {
			compute["aggregation"] = fmt.Sprintf("%v", v)
		}
		if v, ok := d.GetOk(cPath + ".interval"); ok {
			if iv, ok := v.(int); ok && iv != 0 {
				compute["interval"] = int64(iv)
			}
		}
		if v, ok := d.GetOk(cPath + ".metric"); ok && v != "" {
			compute["metric"] = fmt.Sprintf("%v", v)
		}
		result["compute"] = compute
	}

	// search (optional)
	searchCount, _ := d.Get(path + ".search.#").(int)
	if searchCount > 0 {
		sPath := path + ".search.0"
		search := map[string]interface{}{}
		if v, ok := d.GetOk(sPath + ".query"); ok {
			search["query"] = fmt.Sprintf("%v", v)
		}
		result["search"] = search
	}

	// indexes (can be empty array)
	rawIndexes := d.Get(path + ".indexes")
	if items, ok := rawIndexes.([]interface{}); ok {
		indexes := make([]string, len(items))
		for i, item := range items {
			indexes[i] = fmt.Sprintf("%v", item)
		}
		result["indexes"] = indexes
	}

	// group_by (optional)
	groupByCount, _ := d.Get(path + ".group_by.#").(int)
	if groupByCount > 0 {
		groupBys := make([]interface{}, groupByCount)
		for gi := 0; gi < groupByCount; gi++ {
			gbPath := fmt.Sprintf("%s.group_by.%d", path, gi)
			gb := map[string]interface{}{}
			if v, ok := d.GetOk(gbPath + ".facet"); ok {
				gb["facet"] = fmt.Sprintf("%v", v)
			}
			if v, ok := d.GetOk(gbPath + ".limit"); ok {
				if iv, ok := v.(int); ok && iv != 0 {
					gb["limit"] = int64(iv)
				}
			}
			sortCount, _ := d.Get(gbPath + ".sort.#").(int)
			if sortCount > 0 {
				sPath := gbPath + ".sort.0"
				sort := map[string]interface{}{}
				if v, ok := d.GetOk(sPath + ".aggregation"); ok {
					sort["aggregation"] = fmt.Sprintf("%v", v)
				}
				if v, ok := d.GetOk(sPath + ".metric"); ok && v != "" {
					sort["metric"] = fmt.Sprintf("%v", v)
				}
				if v, ok := d.GetOk(sPath + ".order"); ok && v != "" {
					sort["order"] = fmt.Sprintf("%v", v)
				}
				gb["sort"] = sort
			}
			groupBys[gi] = gb
		}
		result["group_by"] = groupBys
	}

	// cross_org_uuids
	if raw := d.Get(path + ".cross_org_uuids"); raw != nil {
		if items, ok := raw.([]interface{}); ok && len(items) == 1 {
			if s, ok := items[0].(string); ok && s != "" {
				result["cross_org_uuids"] = []string{s}
			}
		}
	}

	return result
}

// buildFormulaProcessQueryJSON builds JSON for a formula process_query block.
func buildFormulaProcessQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".metric"); ok {
		result["metric"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".text_filter"); ok && v != "" {
		result["text_filter"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".limit"); ok {
		if iv, ok := v.(int); ok && iv != 0 {
			result["limit"] = int64(iv)
		}
	}
	if v, ok := d.GetOk(path + ".sort"); ok && v != "" {
		result["sort"] = fmt.Sprintf("%v", v)
	}
	isNormalizedCPU, _ := d.Get(path + ".is_normalized_cpu").(bool)
	if isNormalizedCPU {
		result["is_normalized_cpu"] = true
	}
	if raw := d.Get(path + ".tag_filters"); raw != nil {
		if items, ok := raw.([]interface{}); ok && len(items) > 0 {
			tags := make([]string, len(items))
			for i, item := range items {
				tags[i] = fmt.Sprintf("%v", item)
			}
			result["tag_filters"] = tags
		}
	}
	return result
}

// buildSLOQueryJSON builds JSON for a formula slo_query block.
func buildSLOQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".slo_id"); ok {
		result["slo_id"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".measure"); ok {
		result["measure"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".group_mode"); ok && v != "" {
		result["group_mode"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".slo_query_type"); ok && v != "" {
		result["slo_query_type"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".additional_query_filters"); ok && v != "" {
		result["additional_query_filters"] = fmt.Sprintf("%v", v)
	}
	return result
}

// buildCloudCostQueryJSON builds JSON for a formula cloud_cost_query block.
func buildCloudCostQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".query"); ok {
		result["query"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".aggregator"); ok && v != "" {
		result["aggregator"] = fmt.Sprintf("%v", v)
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
func buildScatterplotTableJSON(d *schema.ResourceData, defHCLPath string, defJSON map[string]interface{}) {
	requestPath := defHCLPath + ".request"
	if requestCount, _ := d.Get(requestPath + ".#").(int); requestCount == 0 {
		return
	}
	requestInnerPath := requestPath + ".0"
	tableCount, _ := d.Get(requestInnerPath + ".scatterplot_table.#").(int)
	if tableCount == 0 {
		return
	}
	tablePath := requestInnerPath + ".scatterplot_table.0"

	tableObj := map[string]interface{}{}

	// formulas (ScatterplotWidgetFormula: formula_expression → formula, plus dimension and alias)
	formulaCount, _ := d.Get(tablePath + ".formula.#").(int)
	if formulaCount > 0 {
		formulas := make([]interface{}, formulaCount)
		for fi := 0; fi < formulaCount; fi++ {
			fPath := fmt.Sprintf("%s.formula.%d", tablePath, fi)
			formula := map[string]interface{}{}
			if expr, ok := d.GetOk(fPath + ".formula_expression"); ok {
				formula["formula"] = fmt.Sprintf("%v", expr)
			}
			if dim, ok := d.GetOk(fPath + ".dimension"); ok {
				formula["dimension"] = fmt.Sprintf("%v", dim)
			}
			if alias, ok := d.GetOk(fPath + ".alias"); ok {
				formula["alias"] = fmt.Sprintf("%v", alias)
			}
			formulas[fi] = formula
		}
		tableObj["formulas"] = formulas
	}

	// queries
	queryCount, _ := d.Get(tablePath + ".query.#").(int)
	if queryCount > 0 {
		queries := make([]interface{}, queryCount)
		for qi := 0; qi < queryCount; qi++ {
			qPath := fmt.Sprintf("%s.query.%d", tablePath, qi)
			var queryJSON map[string]interface{}
			if n, _ := d.Get(qPath + ".metric_query.#").(int); n > 0 {
				queryJSON = buildMetricQueryJSON(d, qPath+".metric_query.0")
			} else if n, _ := d.Get(qPath + ".event_query.#").(int); n > 0 {
				queryJSON = buildEventQueryJSON(d, qPath+".event_query.0")
			} else if n, _ := d.Get(qPath + ".process_query.#").(int); n > 0 {
				queryJSON = buildFormulaProcessQueryJSON(d, qPath+".process_query.0")
			} else if n, _ := d.Get(qPath + ".slo_query.#").(int); n > 0 {
				queryJSON = buildSLOQueryJSON(d, qPath+".slo_query.0")
			} else if n, _ := d.Get(qPath + ".cloud_cost_query.#").(int); n > 0 {
				queryJSON = buildCloudCostQueryJSON(d, qPath+".cloud_cost_query.0")
			}
			queries[qi] = queryJSON
		}
		tableObj["queries"] = queries
	}

	if formulaCount > 0 || queryCount > 0 {
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
	return result
}

func flattenSLOQueryJSON(q map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for _, field := range []string{"data_source", "name", "slo_id", "measure", "group_mode", "slo_query_type", "additional_query_filters"} {
		if v, ok := q[field].(string); ok && v != "" {
			result[field] = v
		}
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
func buildWidgetPostProcess(d *schema.ResourceData, spec WidgetSpec, defHCLPath string, defJSON map[string]interface{}) {
	// ---- Formula/query blocks ----
	if isFormulaCapableWidget(spec.JSONType) {
		requestsPath := defHCLPath + ".request"
		requestCount, _ := d.Get(requestsPath + ".#").(int)
		if requestCount > 0 {
			requests := make([]interface{}, requestCount)
			for ri := 0; ri < requestCount; ri++ {
				reqPath := fmt.Sprintf("%s.%d", requestsPath, ri)
				formulaCount, _ := d.Get(reqPath + ".formula.#").(int)
				queryCount, _ := d.Get(reqPath + ".query.#").(int)
				if formulaCount > 0 || queryCount > 0 {
					// New-style formula/query request
					if isTimeseriesFormulaWidget(spec.JSONType) {
						requests[ri] = buildFormulaQueryRequestJSON(d, reqPath)
					} else {
						requests[ri] = buildScalarFormulaQueryRequestJSON(d, reqPath, spec.JSONType)
					}
				} else {
					// Old-style request (already built by FieldSpec engine)
					if existingReqs, ok := defJSON["requests"].([]interface{}); ok && ri < len(existingReqs) {
						requests[ri] = existingReqs[ri]
					}
				}
			}
			defJSON["requests"] = requests
		}
	}

	// ---- Scatterplot table ----
	if spec.JSONType == "scatterplot" {
		buildScatterplotTableJSON(d, defHCLPath, defJSON)
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
		requests := buildQueryTableRequestsJSON(d, defHCLPath)
		defJSON["requests"] = requests
	}

	// ---- Split graph source widget + static_splits ----
	if spec.JSONType == "split_group" {
		srcDefJSON := buildSplitGraphSourceWidgetJSON(d, defHCLPath)
		if srcDefJSON != nil {
			defJSON["source_widget_definition"] = srcDefJSON
		}
		if splitConfigMap, ok := defJSON["split_config"].(map[string]interface{}); ok {
			splitConfigPath := defHCLPath + ".split_config.0"
			staticSplits := buildSplitConfigStaticSplitsJSON(d, splitConfigPath)
			if len(staticSplits) > 0 {
				splitConfigMap["static_splits"] = staticSplits
			}
		}
	}

	// ---- Group nested widgets ----
	if spec.JSONType == "group" {
		widgets := buildGroupWidgetsJSON(d, defHCLPath)
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
func buildQueryTableRequestsJSON(d *schema.ResourceData, defHCLPath string) []interface{} {
	requestsPath := defHCLPath + ".request"
	requestCount, _ := d.Get(requestsPath + ".#").(int)
	if requestCount == 0 {
		return []interface{}{}
	}
	requests := make([]interface{}, requestCount)
	for ri := 0; ri < requestCount; ri++ {
		reqPath := fmt.Sprintf("%s.%d", requestsPath, ri)
		formulaCount, _ := d.Get(reqPath + ".formula.#").(int)
		queryCount, _ := d.Get(reqPath + ".query.#").(int)
		if formulaCount > 0 || queryCount > 0 {
			// Formula/query style request
			requests[ri] = buildQueryTableFormulaRequestJSON(d, reqPath)
		} else {
			// Old-style request — build via FieldSpec engine
			req := BuildEngineJSON(d, reqPath, queryTableOldRequestFields)
			// text_formats needs custom handling (it's a 2D array)
			buildQueryTableTextFormatsJSON(d, reqPath, req)
			requests[ri] = req
		}
	}
	return requests
}

// buildQueryTableFormulaRequestJSON builds a formula/query request for query_table.
// Unlike the generic scalar formula handler, query_table formulas can have
// cell_display_mode, cell_display_mode_options, conditional_formats, number_format
// per formula.
func buildQueryTableFormulaRequestJSON(d *schema.ResourceData, reqPath string) map[string]interface{} {
	result := map[string]interface{}{}

	// formulas
	formulaCount, _ := d.Get(reqPath + ".formula.#").(int)
	if formulaCount > 0 {
		formulas := make([]interface{}, formulaCount)
		for fi := 0; fi < formulaCount; fi++ {
			fPath := fmt.Sprintf("%s.formula.%d", reqPath, fi)
			formula := map[string]interface{}{}

			// formula_expression → formula (JSON key)
			if expr, ok := d.GetOk(fPath + ".formula_expression"); ok {
				formula["formula"] = fmt.Sprintf("%v", expr)
			}
			if alias, ok := d.GetOk(fPath + ".alias"); ok {
				formula["alias"] = fmt.Sprintf("%v", alias)
			}
			// limit
			limitCount, _ := d.Get(fPath + ".limit.#").(int)
			if limitCount > 0 {
				limitPath := fmt.Sprintf("%s.limit.0", fPath)
				limitObj := map[string]interface{}{}
				if count, ok := d.GetOk(limitPath + ".count"); ok {
					switch v := count.(type) {
					case int:
						limitObj["count"] = int64(v)
					case float64:
						limitObj["count"] = int64(v)
					}
				}
				if order, ok := d.GetOk(limitPath + ".order"); ok {
					limitObj["order"] = fmt.Sprintf("%v", order)
				}
				if len(limitObj) > 0 {
					formula["limit"] = limitObj
				}
			}
			// cell_display_mode (TypeString on formula)
			if v, ok := d.GetOk(fPath + ".cell_display_mode"); ok {
				if s := fmt.Sprintf("%v", v); s != "" {
					formula["cell_display_mode"] = s
				}
			}
			// cell_display_mode_options (TypeBlock MaxItems:1)
			cdmoCount, _ := d.Get(fPath + ".cell_display_mode_options.#").(int)
			if cdmoCount > 0 {
				optPath := fPath + ".cell_display_mode_options.0"
				opts := map[string]interface{}{}
				if v, ok := d.GetOk(optPath + ".trend_type"); ok {
					opts["trend_type"] = fmt.Sprintf("%v", v)
				}
				if v, ok := d.GetOk(optPath + ".y_scale"); ok {
					opts["y_scale"] = fmt.Sprintf("%v", v)
				}
				if len(opts) > 0 {
					formula["cell_display_mode_options"] = opts
				}
			}
			// conditional_formats on the formula
			cfCount, _ := d.Get(fPath + ".conditional_formats.#").(int)
			if cfCount > 0 {
				cfs := make([]interface{}, cfCount)
				for ci := 0; ci < cfCount; ci++ {
					cfPath := fmt.Sprintf("%s.conditional_formats.%d", fPath, ci)
					cf := BuildEngineJSON(d, cfPath, widgetConditionalFormatFields)
					cfs[ci] = cf
				}
				formula["conditional_formats"] = cfs
			}
			// number_format (TypeBlock MaxItems:1)
			nfCount, _ := d.Get(fPath + ".number_format.#").(int)
			if nfCount > 0 {
				nfPath := fPath + ".number_format.0"
				nf := buildQueryTableNumberFormatJSON(d, nfPath)
				if len(nf) > 0 {
					formula["number_format"] = nf
				}
			}
			formulas[fi] = formula
		}
		result["formulas"] = formulas
	}

	// queries
	queryCount, _ := d.Get(reqPath + ".query.#").(int)
	if queryCount > 0 {
		queries := make([]interface{}, queryCount)
		for qi := 0; qi < queryCount; qi++ {
			qPath := fmt.Sprintf("%s.query.%d", reqPath, qi)
			var queryJSON map[string]interface{}
			if n, _ := d.Get(qPath + ".metric_query.#").(int); n > 0 {
				queryJSON = buildMetricQueryJSON(d, qPath+".metric_query.0")
			} else if n, _ := d.Get(qPath + ".event_query.#").(int); n > 0 {
				queryJSON = buildEventQueryJSON(d, qPath+".event_query.0")
			} else if n, _ := d.Get(qPath + ".process_query.#").(int); n > 0 {
				queryJSON = buildFormulaProcessQueryJSON(d, qPath+".process_query.0")
			} else if n, _ := d.Get(qPath + ".apm_dependency_stats_query.#").(int); n > 0 {
				queryJSON = buildApmDependencyStatsQueryJSON(d, qPath+".apm_dependency_stats_query.0")
			} else if n, _ := d.Get(qPath + ".apm_resource_stats_query.#").(int); n > 0 {
				queryJSON = buildApmResourceStatsQueryJSON(d, qPath+".apm_resource_stats_query.0")
			} else if n, _ := d.Get(qPath + ".slo_query.#").(int); n > 0 {
				queryJSON = buildSLOQueryJSON(d, qPath+".slo_query.0")
			} else if n, _ := d.Get(qPath + ".cloud_cost_query.#").(int); n > 0 {
				queryJSON = buildCloudCostQueryJSON(d, qPath+".cloud_cost_query.0")
			}
			queries[qi] = queryJSON
		}
		result["queries"] = queries
	}

	// response_format for query_table formulas is always "scalar"
	if formulaCount > 0 || queryCount > 0 {
		result["response_format"] = "scalar"
	}

	return result
}

// buildQueryTableNumberFormatJSON builds the number_format JSON object for a formula.
// The HCL structure has unit { canonical { unit_name, per_unit_name } } or unit { custom { label } }.
// The JSON structure has unit { type, unit_name, per_unit_name } or unit { type, label }.
func buildQueryTableNumberFormatJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	// unit block (TypeList MaxItems:1)
	unitCount, _ := d.Get(path + ".unit.#").(int)
	if unitCount > 0 {
		unitPath := path + ".unit.0"
		unit := map[string]interface{}{}
		// canonical sub-block
		canonicalCount, _ := d.Get(unitPath + ".canonical.#").(int)
		if canonicalCount > 0 {
			canonPath := unitPath + ".canonical.0"
			unit["type"] = "canonical_unit"
			if v, ok := d.GetOk(canonPath + ".unit_name"); ok {
				unit["unit_name"] = fmt.Sprintf("%v", v)
			}
			if v, ok := d.GetOk(canonPath + ".per_unit_name"); ok {
				unit["per_unit_name"] = fmt.Sprintf("%v", v)
			}
		}
		// custom sub-block
		customCount, _ := d.Get(unitPath + ".custom.#").(int)
		if customCount > 0 {
			customPath := unitPath + ".custom.0"
			unit["type"] = "custom_unit_label"
			if v, ok := d.GetOk(customPath + ".label"); ok {
				unit["label"] = fmt.Sprintf("%v", v)
			}
		}
		if len(unit) > 0 {
			result["unit"] = unit
		}
	}
	// unit_scale block
	unitScaleCount, _ := d.Get(path + ".unit_scale.#").(int)
	if unitScaleCount > 0 {
		unitScalePath := path + ".unit_scale.0"
		unitScale := map[string]interface{}{}
		if v, ok := d.GetOk(unitScalePath + ".unit_name"); ok {
			unitScale["unit_name"] = fmt.Sprintf("%v", v)
		}
		if len(unitScale) > 0 {
			result["unit_scale"] = unitScale
		}
	}
	return result
}

// buildQueryTableTextFormatsJSON handles the text_formats 2D array for query_table
// old-style requests. It reads text_formats from the HCL and injects them into
// the request JSON.
func buildQueryTableTextFormatsJSON(d *schema.ResourceData, reqPath string, req map[string]interface{}) {
	tfCount, _ := d.Get(reqPath + ".text_formats.#").(int)
	if tfCount == 0 {
		return
	}
	textFormats := make([]interface{}, tfCount)
	for i := 0; i < tfCount; i++ {
		tfPath := fmt.Sprintf("%s.text_formats.%d", reqPath, i)
		innerCount, _ := d.Get(tfPath + ".text_format.#").(int)
		innerFormats := make([]interface{}, innerCount)
		for j := 0; j < innerCount; j++ {
			innerPath := fmt.Sprintf("%s.text_format.%d", tfPath, j)
			rule := map[string]interface{}{}
			// match (TypeList MaxItems:1, Required)
			matchCount, _ := d.Get(innerPath + ".match.#").(int)
			if matchCount > 0 {
				mPath := innerPath + ".match.0"
				match := map[string]interface{}{}
				if v, ok := d.GetOk(mPath + ".type"); ok {
					match["type"] = fmt.Sprintf("%v", v)
				}
				if v, ok := d.GetOk(mPath + ".value"); ok {
					match["value"] = fmt.Sprintf("%v", v)
				}
				rule["match"] = match
			}
			// palette (optional)
			if v, ok := d.GetOk(innerPath + ".palette"); ok {
				if s := fmt.Sprintf("%v", v); s != "" {
					rule["palette"] = s
				}
			}
			// custom_bg_color (optional)
			if v, ok := d.GetOk(innerPath + ".custom_bg_color"); ok {
				if s := fmt.Sprintf("%v", v); s != "" {
					rule["custom_bg_color"] = s
				}
			}
			// custom_fg_color (optional)
			if v, ok := d.GetOk(innerPath + ".custom_fg_color"); ok {
				if s := fmt.Sprintf("%v", v); s != "" {
					rule["custom_fg_color"] = s
				}
			}
			// replace (TypeList MaxItems:1, Optional)
			replaceCount, _ := d.Get(innerPath + ".replace.#").(int)
			if replaceCount > 0 {
				rPath := innerPath + ".replace.0"
				replace := map[string]interface{}{}
				if v, ok := d.GetOk(rPath + ".type"); ok {
					replace["type"] = fmt.Sprintf("%v", v)
				}
				if v, ok := d.GetOk(rPath + ".with"); ok {
					replace["with"] = fmt.Sprintf("%v", v)
				}
				if v, ok := d.GetOk(rPath + ".substring"); ok {
					if s := fmt.Sprintf("%v", v); s != "" {
						replace["substring"] = s
					}
				}
				rule["replace"] = replace
			}
			innerFormats[j] = rule
		}
		textFormats[i] = innerFormats
	}
	req["text_formats"] = textFormats
}

// buildApmDependencyStatsQueryJSON builds JSON for an apm_dependency_stats_query formula query.
func buildApmDependencyStatsQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".env"); ok {
		result["env"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".operation_name"); ok {
		result["operation_name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".resource_name"); ok {
		result["resource_name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".service"); ok {
		result["service"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".stat"); ok {
		result["stat"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.Get(path + ".is_upstream").(bool); ok {
		result["is_upstream"] = v
	}
	if v, ok := d.GetOk(path + ".primary_tag_name"); ok {
		if s := fmt.Sprintf("%v", v); s != "" {
			result["primary_tag_name"] = s
		}
	}
	if v, ok := d.GetOk(path + ".primary_tag_value"); ok {
		if s := fmt.Sprintf("%v", v); s != "" {
			result["primary_tag_value"] = s
		}
	}
	return result
}

// buildApmResourceStatsQueryJSON builds JSON for an apm_resource_stats_query formula query.
func buildApmResourceStatsQueryJSON(d *schema.ResourceData, path string) map[string]interface{} {
	result := map[string]interface{}{}
	if v, ok := d.GetOk(path + ".data_source"); ok {
		result["data_source"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".env"); ok {
		result["env"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".name"); ok {
		result["name"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".service"); ok {
		result["service"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".stat"); ok {
		result["stat"] = fmt.Sprintf("%v", v)
	}
	if v, ok := d.GetOk(path + ".operation_name"); ok {
		if s := fmt.Sprintf("%v", v); s != "" {
			result["operation_name"] = s
		}
	}
	if v, ok := d.GetOk(path + ".resource_name"); ok {
		if s := fmt.Sprintf("%v", v); s != "" {
			result["resource_name"] = s
		}
	}
	if v, ok := d.GetOk(path + ".primary_tag_name"); ok {
		if s := fmt.Sprintf("%v", v); s != "" {
			result["primary_tag_name"] = s
		}
	}
	if v, ok := d.GetOk(path + ".primary_tag_value"); ok {
		if s := fmt.Sprintf("%v", v); s != "" {
			result["primary_tag_value"] = s
		}
	}
	// group_by: []string
	if raw := d.Get(path + ".group_by"); raw != nil {
		if items, ok := raw.([]interface{}); ok && len(items) > 0 {
			strs := make([]string, len(items))
			for i, item := range items {
				strs[i] = fmt.Sprintf("%v", item)
			}
			result["group_by"] = strs
		}
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
func buildSplitConfigStaticSplitsJSON(d *schema.ResourceData, splitConfigPath string) []interface{} {
	staticSplitsPath := splitConfigPath + ".static_splits"
	count, _ := d.Get(staticSplitsPath + ".#").(int)
	if count == 0 {
		return nil
	}
	result := make([]interface{}, count)
	for i := 0; i < count; i++ {
		entryPath := fmt.Sprintf("%s.%d", staticSplitsPath, i)
		vectorCount, _ := d.Get(entryPath + ".split_vector.#").(int)
		innerArray := make([]interface{}, vectorCount)
		for j := 0; j < vectorCount; j++ {
			vecPath := fmt.Sprintf("%s.split_vector.%d", entryPath, j)
			vec := map[string]interface{}{}
			if v, ok := d.GetOk(vecPath + ".tag_key"); ok {
				vec["tag_key"] = fmt.Sprintf("%v", v)
			}
			// tag_values: always emit (even empty array)
			raw := d.Get(vecPath + ".tag_values")
			var tagValues []string
			switch v := raw.(type) {
			case []interface{}:
				tagValues = make([]string, len(v))
				for k, s := range v {
					tagValues[k] = fmt.Sprintf("%v", s)
				}
			}
			if tagValues == nil {
				tagValues = []string{}
			}
			vec["tag_values"] = tagValues
			innerArray[j] = vec
		}
		result[i] = innerArray
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
func buildSplitGraphSourceWidgetJSON(d *schema.ResourceData, defHCLPath string) map[string]interface{} {
	sourceWidgetPath := defHCLPath + ".source_widget_definition"
	count, _ := d.Get(sourceWidgetPath + ".#").(int)
	if count == 0 {
		return nil
	}
	sourcePath := sourceWidgetPath + ".0"
	// source_widget_definition is like a miniature widget: try each registered spec.
	for _, spec := range allWidgetSpecs {
		cnt, _ := d.Get(sourcePath + "." + spec.HCLKey + ".#").(int)
		if cnt == 0 {
			continue
		}
		innerPath := sourcePath + "." + spec.HCLKey + ".0"
		allFields := make([]FieldSpec, 0, len(CommonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, CommonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := BuildEngineJSON(d, innerPath, allFields)
		defJSON["type"] = spec.JSONType
		// Apply the same per-widget post-processing as the main engine
		buildWidgetPostProcess(d, spec, innerPath, defJSON)
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
// It iterates over the child widget list and delegates to buildWidgetEngineJSON,
// which accepts a path string and handles all widget types uniformly.
func buildGroupWidgetsJSON(d *schema.ResourceData, defHCLPath string) []interface{} {
	widgetsPath := defHCLPath + ".widget"
	widgetCount, _ := d.Get(widgetsPath + ".#").(int)
	widgets := make([]interface{}, widgetCount)
	for i := 0; i < widgetCount; i++ {
		childPath := fmt.Sprintf("%s.%d", widgetsPath, i)
		w := buildWidgetEngineJSON(d, childPath)
		if w == nil {
			w = map[string]interface{}{}
		}
		widgets[i] = w
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

// BuildDashboardEngineJSON builds the full dashboard JSON body from ResourceData.
func BuildDashboardEngineJSON(d *schema.ResourceData) map[string]interface{} {
	result := BuildEngineJSON(d, "", dashboardTopLevelFields)

	// SDK sends "id": "" in POST bodies — replicate for cassette compatibility.
	// On create d.Id() is "" (zero value), on update d.Id() is the real ID.
	// Both POST and PUT bodies need this field set to the current ID.
	result["id"] = d.Id()

	// Handle widgets with type dispatch.
	widgetCount, _ := d.Get("widget.#").(int)
	widgets := make([]interface{}, widgetCount)
	for i := 0; i < widgetCount; i++ {
		widgets[i] = buildWidgetEngineJSON(d, fmt.Sprintf("widget.%d", i))
	}
	result["widgets"] = widgets

	return result
}

// UpdateDashboardEngineState sets ResourceData state from the dashboard API response map.
func UpdateDashboardEngineState(d *schema.ResourceData, resp map[string]interface{}) diag.Diagnostics {
	// Set simple top-level fields from response
	if v, ok := resp["title"]; ok {
		if err := d.Set("title", fmt.Sprintf("%v", v)); err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := resp["layout_type"]; ok {
		if err := d.Set("layout_type", fmt.Sprintf("%v", v)); err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := resp["reflow_type"]; ok {
		if err := d.Set("reflow_type", fmt.Sprintf("%v", v)); err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := resp["description"]; ok {
		if err := d.Set("description", fmt.Sprintf("%v", v)); err != nil {
			return diag.FromErr(err)
		}
	}
	if v, ok := resp["url"]; ok {
		if err := d.Set("url", fmt.Sprintf("%v", v)); err != nil {
			return diag.FromErr(err)
		}
	}

	// Handle is_read_only / restricted_roles (same logic as prepResource in json resource)
	// 'restricted_roles' takes precedence over 'is_read_only'
	if restrictedRoles, ok := resp["restricted_roles"].([]interface{}); ok {
		roles := make([]string, len(restrictedRoles))
		for i, r := range restrictedRoles {
			roles[i] = fmt.Sprintf("%v", r)
		}
		if err := d.Set("restricted_roles", roles); err != nil {
			return diag.FromErr(err)
		}
		// Clear is_read_only when restricted_roles is set
		if err := d.Set("is_read_only", false); err != nil {
			return diag.FromErr(err)
		}
	} else {
		// is_read_only defaults to false — set it to avoid continuous diff when not set
		isReadOnly := false
		if v, ok := resp["is_read_only"].(bool); ok {
			isReadOnly = v
		}
		if err := d.Set("is_read_only", isReadOnly); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set notify_list
	if v, ok := resp["notify_list"].([]interface{}); ok {
		notifyList := make([]string, len(v))
		for i, n := range v {
			notifyList[i] = fmt.Sprintf("%v", n)
		}
		if err := d.Set("notify_list", notifyList); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := d.Set("notify_list", []string{}); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set tags
	if v, ok := resp["tags"].([]interface{}); ok {
		tags := make([]string, len(v))
		for i, t := range v {
			tags[i] = fmt.Sprintf("%v", t)
		}
		if err := d.Set("tags", tags); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set template_variables
	if v, ok := resp["template_variables"].([]interface{}); ok {
		tvs := flattenTemplateVariables(v)
		if err := d.Set("template_variable", tvs); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set template_variable_presets
	if v, ok := resp["template_variable_presets"].([]interface{}); ok {
		tvps := flattenTemplateVariablePresets(v)
		if err := d.Set("template_variable_preset", tvps); err != nil {
			return diag.FromErr(err)
		}
	}

	// Set widgets
	if widgets, ok := resp["widgets"].([]interface{}); ok {
		terraformWidgets := make([]interface{}, len(widgets))
		for i, w := range widgets {
			widgetData, ok := w.(map[string]interface{})
			if !ok {
				continue
			}
			flattened := flattenWidgetEngineJSON(widgetData)
			if flattened == nil {
				// Widget type not implemented yet — preserve existing state
				// by trying to get from current state
				existing := d.Get(fmt.Sprintf("widget.%d", i))
				if existing != nil {
					if existingMap, ok := existing.(map[string]interface{}); ok {
						flattened = existingMap
					}
				}
				if flattened == nil {
					flattened = map[string]interface{}{}
				}
			}
			// Widget IDs come FROM the response and are set in state
			// The schema has widget.id as Computed TypeInt
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
			// Also check top-level widget id
			if widgetID, ok := widgetData["id"]; ok {
				switch v := widgetID.(type) {
				case float64:
					flattened["id"] = int(v)
				case int:
					flattened["id"] = v
				}
			}
			// Flatten widget_layout from JSON "layout" object.
			// The API returns layout as {"x":5,"y":5,"width":32,"height":43}.
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
			terraformWidgets[i] = flattened
		}
		if err := d.Set("widget", terraformWidgets); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// flattenTemplateVariables converts the JSON template_variables list to HCL state.
func flattenTemplateVariables(tvs []interface{}) []interface{} {
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
		result[i] = item
	}
	return result
}

// flattenTemplateVariablePresets converts the JSON template_variable_presets list to HCL state.
func flattenTemplateVariablePresets(tvps []interface{}) []interface{} {
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
// from ResourceData. widgetPath is the dotted HCL path to the widget block, e.g. "widget.0".
// Returns nil if no recognized widget type is found at that path.
func BuildWidgetEngineJSON(d *schema.ResourceData, widgetPath string) map[string]interface{} {
	return buildWidgetEngineJSON(d, widgetPath)
}

// FlattenWidgetEngineJSON is the exported entry point for flattening a single widget's JSON map
// into HCL state suitable for setting on a TypeList field in schema.ResourceData.
// Returns nil if the widget type is not recognized.
func FlattenWidgetEngineJSON(widgetData map[string]interface{}) map[string]interface{} {
	return flattenWidgetEngineJSON(widgetData)
}

// MarshalDashboardJSON marshals the dashboard JSON body from ResourceData.
// The trailing newline matches behavior of json.NewEncoder used when cassettes were recorded.
func MarshalDashboardJSON(d *schema.ResourceData) (string, error) {
	body, err := json.Marshal(BuildDashboardEngineJSON(d))
	if err != nil {
		return "", fmt.Errorf("error marshaling dashboard JSON: %s", err)
	}
	return string(body) + "\n", nil
}
