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
