package datadog

// resource_datadog_dashboard_engine.go
//
// FieldSpec bidirectional mapping engine for the dashboard resource.
// This file contains the generic engine types (FieldSpec, WidgetSpec, FieldType)
// and all reusable FieldSpec groups mirroring OpenAPI schema components.
//
// Each reusable []FieldSpec variable is named after the OpenAPI
// components/schemas/ entry it corresponds to (camelCase). A comment
// identifies the OpenAPI schema and which widget types use it.
//
// Phase 1 implements: timeseries widget.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
)

// FieldType drives serialization/deserialization behavior in the engine.
type FieldType int

const (
	TypeString    FieldType = iota // plain string
	TypeBool                       // bool
	TypeInt                        // int
	TypeFloat                      // float64
	TypeStringList                 // []string (TypeList or TypeSet of strings)
	TypeIntList                    // []int
	TypeBlock                      // single nested block (MaxItems:1 in schema)
	TypeBlockList                  // list of blocks
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

	// Fields are the widget-specific fields.
	// CommonWidgetFields are automatically merged in by the engine.
	Fields []FieldSpec
}

// ============================================================
// Reusable FieldSpec Groups (mirroring OpenAPI $ref schemas)
// ============================================================

// widgetCustomLinkFields corresponds to OpenAPI components/schemas/WidgetCustomLink.
// Used by: timeseries, toplist, query_value, change, distribution, heatmap, hostmap,
//          geomap, scatterplot, service_map, sunburst, table, topology_map, treemap,
//          run_workflow (15 widget types).
// HCL key: "custom_link" (singular Terraform convention)
// JSON key: "custom_links" (plural, matching OpenAPI)
var widgetCustomLinkFields = []FieldSpec{
	{HCLKey: "label", Type: TypeString, OmitEmpty: true},
	{HCLKey: "link", Type: TypeString, OmitEmpty: false},
	{HCLKey: "is_hidden", Type: TypeBool, OmitEmpty: true},
	{HCLKey: "override_label", Type: TypeString, OmitEmpty: true},
}

// widgetTimeField corresponds to OpenAPI components/schemas/WidgetLegacyLiveSpan
// (the live_span variant of WidgetTime, which is the form used by HCL).
// HCL flattens this to a single "live_span" string field on the widget definition,
// which maps to {"time": {"live_span": "..."}} in JSON via JSONPath.
// Used by: 21+ widget types.
var widgetTimeField = FieldSpec{
	HCLKey:    "live_span",
	JSONPath:  "time.live_span",
	Type:      TypeString,
	OmitEmpty: true,
}

// widgetAxisFields corresponds to OpenAPI components/schemas/WidgetAxis.
// Used by: timeseries (yaxis + right_yaxis), distribution, heatmap, scatterplot.
var widgetAxisFields = []FieldSpec{
	{HCLKey: "label", Type: TypeString, OmitEmpty: true},
	{HCLKey: "min", Type: TypeString, OmitEmpty: true},
	{HCLKey: "max", Type: TypeString, OmitEmpty: true},
	{HCLKey: "scale", Type: TypeString, OmitEmpty: true},
	// include_zero is always emitted even when false (OmitEmpty: false)
	// confirmed by cassette: "include_zero": false appears in right_yaxis
	{HCLKey: "include_zero", Type: TypeBool, OmitEmpty: false},
}

// widgetMarkerFields corresponds to OpenAPI components/schemas/WidgetMarker.
// Used by: timeseries, distribution, heatmap.
// HCL key: "marker" (singular), JSON key: "markers" (plural).
var widgetMarkerFields = []FieldSpec{
	{HCLKey: "value", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
	{HCLKey: "display_type", Type: TypeString, OmitEmpty: true},
	{HCLKey: "label", Type: TypeString, OmitEmpty: true},
}

// widgetEventFields corresponds to OpenAPI components/schemas/WidgetEvent.
// Used by: timeseries, heatmap.
// HCL key: "event" (singular), JSON key: "events" (plural).
var widgetEventFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: false},             // required in OpenAPI
	{HCLKey: "tags_execution", Type: TypeString, OmitEmpty: true}, // omit when empty
}

// logQueryDefinitionGroupBySortFields corresponds to OpenAPI
// components/schemas/LogQueryDefinitionGroupBySort.
var logQueryDefinitionGroupBySortFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false},
	{HCLKey: "order", Type: TypeString, OmitEmpty: false},
	{HCLKey: "facet", Type: TypeString, OmitEmpty: true},
}

// logQueryDefinitionGroupByFields corresponds to OpenAPI
// components/schemas/LogQueryDefinitionGroupBy.
var logQueryDefinitionGroupByFields = []FieldSpec{
	{HCLKey: "facet", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
	// HCL key: "sort_query" (disambiguates from other sort fields in HCL)
	// JSON key: "sort" (OpenAPI property name)
	{HCLKey: "sort_query", JSONKey: "sort", Type: TypeBlock, OmitEmpty: true,
		Children: logQueryDefinitionGroupBySortFields},
}

// logsQueryComputeFields corresponds to OpenAPI components/schemas/LogsQueryCompute.
var logsQueryComputeFields = []FieldSpec{
	{HCLKey: "aggregation", Type: TypeString, OmitEmpty: false}, // required in OpenAPI
	{HCLKey: "facet", Type: TypeString, OmitEmpty: true},
	{HCLKey: "interval", Type: TypeInt, OmitEmpty: true},
}

// logQueryDefinitionFields corresponds to OpenAPI components/schemas/LogQueryDefinition.
// Used by request fields: log_query, apm_query, rum_query, network_query,
//                         security_query, audit_query, event_query, profile_metrics_query.
// That is: the same FieldSpec is reused for all 8 query-type fields on a request.
// HCL flattens "search.query" to "search_query" via JSONPath.
// HCL uses "compute_query" instead of "compute" (disambiguates from other uses); JSONKey: "compute".
var logQueryDefinitionFields = []FieldSpec{
	{HCLKey: "index", Type: TypeString, OmitEmpty: false},
	// search_query (flat HCL) → {"search": {"query": "..."}} (nested JSON) via JSONPath
	{HCLKey: "search_query", JSONPath: "search.query", Type: TypeString, OmitEmpty: false},
	// HCL "compute_query" → JSON "compute" (renamed to avoid ambiguity in HCL)
	{HCLKey: "compute_query", JSONKey: "compute", Type: TypeBlock, OmitEmpty: true,
		Children: logsQueryComputeFields},
	// multi_compute → JSON "multi_compute" (same key, list of compute objects)
	{HCLKey: "multi_compute", Type: TypeBlockList, OmitEmpty: true,
		Children: logsQueryComputeFields},
	// HCL key: "group_by" (same in HCL and JSON — no pluralization applied to this field)
	{HCLKey: "group_by", Type: TypeBlockList, OmitEmpty: true,
		Children: logQueryDefinitionGroupByFields},
}

// processQueryDefinitionFields corresponds to OpenAPI
// components/schemas/ProcessQueryDefinition.
// Used by timeseries and other widgets that support process metrics.
var processQueryDefinitionFields = []FieldSpec{
	{HCLKey: "metric", Type: TypeString, OmitEmpty: false},    // required in OpenAPI
	{HCLKey: "search_by", Type: TypeString, OmitEmpty: true},  // omit when empty
	{HCLKey: "filter_by", Type: TypeStringList, OmitEmpty: true},
	{HCLKey: "limit", Type: TypeInt, OmitEmpty: true},
}

// ============================================================
// Common Widget Fields
// ============================================================

// commonWidgetFields are the FieldSpecs shared by most widget definition types.
// They are merged automatically into every WidgetSpec by the engine.
var commonWidgetFields = []FieldSpec{
	// Inline properties on widget definitions (no OpenAPI $ref, common by convention)
	{HCLKey: "title", Type: TypeString, OmitEmpty: true},
	{HCLKey: "title_size", Type: TypeString, OmitEmpty: true},
	{HCLKey: "title_align", Type: TypeString, OmitEmpty: true},
	// WidgetTime: live_span (HCL) → {"time": {"live_span": "..."}} (JSON)
	widgetTimeField,
	// WidgetCustomLink: HCL "custom_link" (singular) → JSON "custom_links" (plural)
	{HCLKey: "custom_link", JSONKey: "custom_links", Type: TypeBlockList, OmitEmpty: true,
		Children: widgetCustomLinkFields},
}

// ============================================================
// Timeseries Widget
// ============================================================

// timeseriesWidgetRequestStyleFields corresponds to OpenAPI
// components/schemas/WidgetRequestStyle (inline on TimeseriesWidgetRequest).
var timeseriesWidgetRequestStyleFields = []FieldSpec{
	{HCLKey: "palette", Type: TypeString, OmitEmpty: true},
	{HCLKey: "line_type", Type: TypeString, OmitEmpty: true},
	{HCLKey: "line_width", Type: TypeString, OmitEmpty: true},
	{HCLKey: "order_by", Type: TypeString, OmitEmpty: true},
}

// timeseriesWidgetMetadataFields corresponds to the inline metadata object
// on TimeseriesWidgetRequest (no standalone OpenAPI $ref; defined inline).
var timeseriesWidgetMetadataFields = []FieldSpec{
	{HCLKey: "expression", Type: TypeString, OmitEmpty: false},
	{HCLKey: "alias_name", Type: TypeString, OmitEmpty: true},
}

// timeseriesWidgetRequestFields corresponds to OpenAPI
// components/schemas/TimeseriesWidgetRequest.
// HCL key: "request" (singular), JSON key: "requests" (plural).
var timeseriesWidgetRequestFields = []FieldSpec{
	{HCLKey: "q", Type: TypeString, OmitEmpty: true},
	{HCLKey: "display_type", Type: TypeString, OmitEmpty: true},
	// on_right_yaxis is always emitted even when false — cassette confirms both true and false appear
	{HCLKey: "on_right_yaxis", Type: TypeBool, OmitEmpty: false},
	{HCLKey: "style", Type: TypeBlock, OmitEmpty: true, Children: timeseriesWidgetRequestStyleFields},
	{HCLKey: "metadata", Type: TypeBlockList, OmitEmpty: true, Children: timeseriesWidgetMetadataFields},
	// The following 7 fields all use logQueryDefinitionFields (same OpenAPI $ref,
	// different JSON key per query source type):
	{HCLKey: "log_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "apm_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "rum_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "network_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "security_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "audit_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	{HCLKey: "profile_metrics_query", Type: TypeBlock, OmitEmpty: true, Children: logQueryDefinitionFields},
	// ProcessQueryDefinition
	{HCLKey: "process_query", Type: TypeBlock, OmitEmpty: true, Children: processQueryDefinitionFields},
}

// timeseriesWidgetSpec corresponds to OpenAPI
// components/schemas/TimeseriesWidgetDefinition.
var timeseriesWidgetSpec = WidgetSpec{
	HCLKey:   "timeseries_definition",
	JSONType: "timeseries",
	Fields: []FieldSpec{
		// show_legend is always emitted even when false — cassette confirms false appears when not set
		{HCLKey: "show_legend", Type: TypeBool, OmitEmpty: false},
		{HCLKey: "legend_size", Type: TypeString, OmitEmpty: true},
		{HCLKey: "legend_layout", Type: TypeString, OmitEmpty: true},
		{HCLKey: "legend_columns", Type: TypeStringList, OmitEmpty: true},
		// WidgetAxis — used twice for the two y-axes
		{HCLKey: "yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		{HCLKey: "right_yaxis", Type: TypeBlock, OmitEmpty: true, Children: widgetAxisFields},
		// WidgetMarker: HCL singular "marker" → JSON plural "markers"
		{HCLKey: "marker", JSONKey: "markers", Type: TypeBlockList, OmitEmpty: true, Children: widgetMarkerFields},
		// WidgetEvent: HCL singular "event" → JSON plural "events"
		{HCLKey: "event", JSONKey: "events", Type: TypeBlockList, OmitEmpty: true, Children: widgetEventFields},
		// TimeseriesWidgetRequest: HCL singular "request" → JSON plural "requests"
		// OmitEmpty: false — always emit even if empty (matches SDK behavior in cassettes)
		{HCLKey: "request", JSONKey: "requests", Type: TypeBlockList, OmitEmpty: false,
			Children: timeseriesWidgetRequestFields},
	},
}

// allWidgetSpecs is the registry of all implemented WidgetSpecs.
// Phase 1: timeseries only.
var allWidgetSpecs = []WidgetSpec{
	timeseriesWidgetSpec,
}

// ============================================================
// Dashboard Top Level
// ============================================================

// templateVariableFields corresponds to OpenAPI DashboardTemplateVariable.
// HCL key: "template_variable" (singular), JSON key: "template_variables" (plural).
var templateVariableFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: false},
	{HCLKey: "prefix", Type: TypeString, OmitEmpty: true},
	{HCLKey: "default", Type: TypeString, OmitEmpty: true},
	{HCLKey: "defaults", Type: TypeStringList, OmitEmpty: true},
	{HCLKey: "available_values", Type: TypeStringList, OmitEmpty: true},
}

// templateVariablePresetValueFields corresponds to OpenAPI DashboardTemplateVariablePresetValue.
// Used inside template_variable_preset blocks.
// HCL key: "template_variable" (singular), JSON key: "template_variables" (plural).
var templateVariablePresetValueFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: true},
	{HCLKey: "value", Type: TypeString, OmitEmpty: true},
	{HCLKey: "values", Type: TypeStringList, OmitEmpty: true},
}

// templateVariablePresetFields corresponds to OpenAPI DashboardTemplateVariablePreset.
// HCL key: "template_variable_preset" (singular), JSON key: "template_variable_presets" (plural).
var templateVariablePresetFields = []FieldSpec{
	{HCLKey: "name", Type: TypeString, OmitEmpty: true},
	// template_variable (singular HCL) → template_variables (plural JSON)
	{HCLKey: "template_variable", JSONKey: "template_variables", Type: TypeBlockList, OmitEmpty: true,
		Children: templateVariablePresetValueFields},
}

// dashboardTopLevelFields are the top-level fields of the Dashboard object.
var dashboardTopLevelFields = []FieldSpec{
	{HCLKey: "title", Type: TypeString, OmitEmpty: false},
	{HCLKey: "description", Type: TypeString, OmitEmpty: false},
	{HCLKey: "layout_type", Type: TypeString, OmitEmpty: false},
	{HCLKey: "reflow_type", Type: TypeString, OmitEmpty: true},
	// notify_list: always send [], never omit (OmitEmpty: false)
	{HCLKey: "notify_list", Type: TypeStringList, OmitEmpty: false},
	// tags: always send [], never omit
	{HCLKey: "tags", Type: TypeStringList, OmitEmpty: false},
	// template_variable (HCL singular) → template_variables (JSON plural)
	{HCLKey: "template_variable", JSONKey: "template_variables", Type: TypeBlockList, OmitEmpty: false,
		Children: templateVariableFields},
	// template_variable_preset (HCL singular) → template_variable_presets (JSON plural)
	{HCLKey: "template_variable_preset", JSONKey: "template_variable_presets", Type: TypeBlockList, OmitEmpty: false,
		Children: templateVariablePresetFields},
	// restricted_roles: omit when empty
	{HCLKey: "restricted_roles", Type: TypeStringList, OmitEmpty: true},
	// is_read_only: kept in schema for backward compat; omit when false
	{HCLKey: "is_read_only", Type: TypeBool, OmitEmpty: true},
}

// ============================================================
// Generic Engine: HCL → JSON (build direction)
// ============================================================

// buildEngineJSON converts schema.ResourceData at the given HCL path prefix to a JSON map.
// hclPrefix is the dotted path to the parent block, e.g., "widget.0.timeseries_definition.0"
func buildEngineJSON(d *schema.ResourceData, hclPrefix string, fields []FieldSpec) map[string]interface{} {
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
			nested := buildEngineJSON(d, hclPath+".0", f.Children)
			if len(nested) == 0 && f.OmitEmpty {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), nested)

		case TypeBlockList:
			count, _ := d.Get(hclPath + ".#").(int)
			items := make([]interface{}, count)
			for i := 0; i < count; i++ {
				items[i] = buildEngineJSON(d, fmt.Sprintf("%s.%d", hclPath, i), f.Children)
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

// flattenEngineJSON converts a JSON map into a map[string]interface{} suitable
// for setting on a TypeList field in schema.ResourceData.
func flattenEngineJSON(fields []FieldSpec, data map[string]interface{}) map[string]interface{} {
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

		case TypeStringList:
			result[f.HCLKey] = toStringSliceFromInterface(jsonVal)

		case TypeBlock:
			if m, ok := jsonVal.(map[string]interface{}); ok {
				result[f.HCLKey] = []interface{}{flattenEngineJSON(f.Children, m)}
			}

		case TypeBlockList:
			if items, ok := jsonVal.([]interface{}); ok {
				list := make([]interface{}, len(items))
				for i, item := range items {
					if m, ok := item.(map[string]interface{}); ok {
						list[i] = flattenEngineJSON(f.Children, m)
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
func buildWidgetEngineJSON(d *schema.ResourceData, widgetIdx int) map[string]interface{} {
	widgetPath := fmt.Sprintf("widget.%d", widgetIdx)
	for _, spec := range allWidgetSpecs {
		defPath := fmt.Sprintf("%s.%s.#", widgetPath, spec.HCLKey)
		count, _ := d.Get(defPath).(int)
		if count == 0 {
			continue
		}
		defHCLPath := fmt.Sprintf("%s.%s.0", widgetPath, spec.HCLKey)
		allFields := make([]FieldSpec, 0, len(commonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, commonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defJSON := buildEngineJSON(d, defHCLPath, allFields)
		defJSON["type"] = spec.JSONType

		// For timeseries widgets, post-process requests to handle formula/query blocks.
		// The FieldSpec engine handles old-style requests (q, log_query, etc.).
		// New-style formula/query requests need special handling.
		if spec.JSONType == "timeseries" {
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
						requests[ri] = buildFormulaQueryRequestJSON(d, reqPath)
					} else {
						// Old-style request (already built by FieldSpec engine)
						// Get it from the already-built requests array
						if existingReqs, ok := defJSON["requests"].([]interface{}); ok && ri < len(existingReqs) {
							requests[ri] = existingReqs[ri]
						}
					}
				}
				defJSON["requests"] = requests
			}
		}

		return map[string]interface{}{"definition": defJSON}
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
			formulas[fi] = formula
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

	// response_format is always "timeseries" for formula/query requests
	if formulaCount > 0 || queryCount > 0 {
		result["response_format"] = "timeseries"
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
		allFields := make([]FieldSpec, 0, len(commonWidgetFields)+len(spec.Fields))
		allFields = append(allFields, commonWidgetFields...)
		allFields = append(allFields, spec.Fields...)
		defState := flattenEngineJSON(allFields, def)

		// For timeseries widgets, post-process requests to handle formula/query
		if spec.JSONType == "timeseries" {
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
						// New-style formula/query request
						flatRequests[ri] = flattenFormulaQueryRequestJSON(reqMap)
					} else {
						// Old-style request — already flattened by FieldSpec engine
						if existingRequests, ok := defState["request"].([]interface{}); ok && ri < len(existingRequests) {
							flatRequests[ri] = existingRequests[ri]
						} else {
							flatRequests[ri] = flattenEngineJSON(timeseriesWidgetRequestFields, reqMap)
						}
					}
				}
				defState["request"] = flatRequests
			}
		}

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
			flatQueries[qi] = flattenFormulaQueryJSON(qm)
		}
		result["query"] = flatQueries
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

const dashboardAPIPath = "/api/v1/dashboard"

// buildDashboardEngineJSON builds the full dashboard JSON body from ResourceData.
func buildDashboardEngineJSON(d *schema.ResourceData) map[string]interface{} {
	result := buildEngineJSON(d, "", dashboardTopLevelFields)

	// SDK sends "id": "" in POST bodies — replicate for cassette compatibility.
	// On create d.Id() is "" (zero value), on update d.Id() is the real ID.
	// Both POST and PUT bodies need this field set to the current ID.
	result["id"] = d.Id()

	// Handle widgets with type dispatch.
	widgetCount, _ := d.Get("widget.#").(int)
	widgets := make([]interface{}, widgetCount)
	for i := 0; i < widgetCount; i++ {
		widgets[i] = buildWidgetEngineJSON(d, i)
	}
	result["widgets"] = widgets

	return result
}

// updateDashboardEngineState sets ResourceData state from the dashboard API response map.
func updateDashboardEngineState(d *schema.ResourceData, resp map[string]interface{}) diag.Diagnostics {
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

// ============================================================
// CRUD Functions (raw HTTP, replacing SDK-based versions)
// ============================================================

func resourceDatadogDashboardCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	body, err := json.Marshal(buildDashboardEngineJSON(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling dashboard JSON: %s", err))
	}
	// The SDK uses json.NewEncoder which appends \n; cassettes were recorded with that behavior.
	// Add \n to match cassette body format.
	bodyStr := string(body) + "\n"

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "POST", dashboardAPIPath, &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error creating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	id, ok := respMap["id"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving id from response"))
	}
	d.SetId(fmt.Sprintf("%v", id))

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	var httpResponse *http.Response
	retryErr := retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		_, httpResponse, err = utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardAPIPath+"/"+d.Id(), nil)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode == 404 {
				return retry.RetryableError(fmt.Errorf("dashboard not created yet"))
			}
			return retry.NonRetryableError(err)
		}
		// We only log the error, as failing to update the list shouldn't fail dashboard creation
		updateDashboardLists(d, providerConf, d.Id(), fmt.Sprintf("%v", layoutType))
		return nil
	})
	if retryErr != nil {
		return diag.FromErr(retryErr)
	}

	return updateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "GET", dashboardAPIPath+"/"+d.Id(), nil)
	if err != nil {
		if httpresp != nil && httpresp.StatusCode == 404 {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientErrorDiag(err, httpresp, "error getting dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	return updateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	body, err := json.Marshal(buildDashboardEngineJSON(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("error marshaling dashboard JSON: %s", err))
	}
	// The SDK uses json.NewEncoder which appends \n; cassettes were recorded with that behavior.
	// Add \n to match cassette body format.
	bodyStr := string(body) + "\n"

	respByte, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "PUT", dashboardAPIPath+"/"+d.Id(), &bodyStr)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error updating dashboard")
	}

	respMap, err := utils.ConvertResponseByteToMap(respByte)
	if err != nil {
		return diag.FromErr(err)
	}

	layoutType, ok := respMap["layout_type"]
	if !ok {
		return diag.FromErr(errors.New("error retrieving layout_type from response"))
	}

	updateDashboardLists(d, providerConf, d.Id(), fmt.Sprintf("%v", layoutType))

	return updateDashboardEngineState(d, respMap)
}

func resourceDatadogDashboardDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	providerConf := meta.(*ProviderConfiguration)
	apiInstances := providerConf.DatadogApiInstances
	auth := providerConf.Auth

	_, httpresp, err := utils.SendRequest(auth, apiInstances.HttpClient, "DELETE", dashboardAPIPath+"/"+d.Id(), nil)
	if err != nil {
		return utils.TranslateClientErrorDiag(err, httpresp, "error deleting dashboard")
	}

	return nil
}
