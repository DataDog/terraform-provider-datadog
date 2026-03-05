package dashboardmapping

// engine_sdk2.go
//
// SDKv2 build functions that read from map[string]interface{} (SDKv2 native types
// as returned by d.Get() or ResourceData.GetRawState()).
//
// This is the SDKv2 parallel of the build direction in engine.go.
// The flatten direction reuses FlattenEngineJSON, flattenWidgetEngineJSON, and
// flattenWidgetPostProcess directly — they already return map[string]interface{}.
//
// Primary entry points:
//   - MarshalDashboardJSONFromMap: build JSON body string from SDKv2 data map
//   - BuildDashboardEngineJSONFromMap: build JSON body map from SDKv2 data map
//   - BuildEngineJSONFromMap: generic FieldSpec engine for SDKv2 maps
//   - FlattenWidgetsForSDKv2: convert API response widgets to []interface{} for d.Set

import (
	"encoding/json"
	"fmt"
)

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

		case TypeBlockList:
			items := getBlockListFromMap(data, f.HCLKey)
			built := make([]interface{}, 0, len(items))
			for _, item := range items {
				built = append(built, BuildEngineJSONFromMap(item, f.Children))
			}
			if f.OmitEmpty && len(built) == 0 {
				continue
			}
			setAtJSONPath(result, f.effectiveJSONPath(), built)

		case TypeOneOf:
			// The TypeOneOf container is a TypeList, MaxItems:1 in SDKv2.
			// Each child variant is itself a TypeBlock (TypeList, MaxItems:1).
			outer := getBlockFromMap(data, f.HCLKey)
			if outer == nil {
				if f.OmitEmpty {
					continue
				}
				break
			}
			// Find the first populated variant child
			var built map[string]interface{}
			var matchedChild *FieldSpec
			for i := range f.Children {
				child := &f.Children[i]
				if child.Type == TypeBlock {
					inner := getBlockFromMap(outer, child.HCLKey)
					if inner != nil {
						built = BuildEngineJSONFromMap(inner, child.Children)
						matchedChild = child
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
	result["widgets"] = buildWidgetsJSONFromMap(data)

	return result
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
