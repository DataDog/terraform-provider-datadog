package dashboardmapping

// widgets_complex.go — Batch C
//
// Complex/structural widget types: query_table, list_stream, slo, slo_list,
// split_graph, group, powerpack.
//
// Batch C depends on Batch B having merged first, since these widgets
// may reuse field groups (e.g. widgetConditionalFormatFields) defined in field_groups_requests.go.
// New field groups specific to Batch C go in field_groups_complex.go.

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// ============================================================
// Query Table Widget
// ============================================================

// QueryTableWidgetSpec corresponds to OpenAPI
// components/schemas/TableWidgetDefinition.
// Formula requests are handled by post-processing (buildQueryTableFormulaRequestsJSON).
var QueryTableWidgetSpec = WidgetSpec{
	HCLKey:      "query_table_definition",
	JSONType:    "query_table",
	Description: "The definition for a Query Table widget.",
	Fields: []FieldSpec{
		// has_search_bar: OmitEmpty — only present when explicitly set
		{
			HCLKey:      "has_search_bar",
			Type:        TypeString,
			OmitEmpty:   true,
			Description: "Controls the display of the search bar.",
			ValidValues: []string{"always", "never", "auto"},
		},
		// request: HCL singular "request" → JSON plural "requests"
		// OmitEmpty: false — always emit even if empty
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Description: "A nested block describing the request to use when displaying the widget. Multiple `request` blocks are allowed using the structure below (exactly one of `q`, `apm_query`, `log_query`, `rum_query`, `security_query`, `apm_stats_query` or `process_query` is required within the `request` block).",
			Children:    queryTableOldRequestFields,
		},
	},
}

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

// ============================================================
// List Stream Widget
// ============================================================

// ListStreamWidgetSpec corresponds to OpenAPI
// components/schemas/ListStreamWidgetDefinition.
var ListStreamWidgetSpec = WidgetSpec{
	HCLKey:      "list_stream_definition",
	JSONType:    "list_stream",
	Description: "The definition for a List Stream widget.",
	Fields: []FieldSpec{
		// request: HCL singular "request" → JSON plural "requests"
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Required:    true,
			Description: "Nested block describing the requests to use when displaying the widget. Multiple `request` blocks are allowed with the structure below.",
			Children:    listStreamRequestFields,
		},
	},
}

// ============================================================
// SLO Widget
// ============================================================

// SLOWidgetSpec corresponds to OpenAPI
// components/schemas/SLOWidgetDefinition.
var SLOWidgetSpec = WidgetSpec{
	HCLKey:      "service_level_objective_definition",
	JSONType:    "slo",
	Description: "The definition for a Service Level Objective widget.",
	Fields: []FieldSpec{
		// Required fields
		{
			HCLKey:      "slo_id",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The ID of the service level objective used by the widget.",
		},
		{
			HCLKey:      "view_type",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The type of view to use when displaying the widget. Only `detail` is supported.",
		},
		{
			HCLKey:      "view_mode",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The view mode for the widget.",
			ValidValues: []string{"overall", "component", "both"},
		},
		{
			HCLKey:      "time_windows",
			Type:        TypeStringList,
			OmitEmpty:   false,
			Required:    true,
			Description: "A list of time windows to display in the widget.",
			ValidValues: []string{"7d", "30d", "90d", "week_to_date", "previous_week", "month_to_date", "previous_month", "global_time"},
		},
		// Optional fields
		{HCLKey: "show_error_budget", Type: TypeBool, OmitEmpty: true, Description: "Whether to show the error budget or not."},
		{HCLKey: "global_time_target", Type: TypeString, OmitEmpty: true, Description: "The global time target of the widget."},
		{HCLKey: "additional_query_filters", Type: TypeString, OmitEmpty: true, Description: "Additional filters applied to the SLO query."},
	},
}

// ============================================================
// SLO List Widget
// ============================================================

// SLOListWidgetSpec corresponds to OpenAPI
// components/schemas/SLOListWidgetDefinition.
var SLOListWidgetSpec = WidgetSpec{
	HCLKey:      "slo_list_definition",
	JSONType:    "slo_list",
	Description: "The definition for a SLO List widget.",
	Fields: []FieldSpec{
		// request: HCL singular "request" → JSON plural "requests"
		{
			HCLKey:      "request",
			JSONKey:     "requests",
			Type:        TypeBlockList,
			OmitEmpty:   false,
			Required:    true,
			Description: "A nested block describing the request to use when displaying the widget. Exactly one `request` block is allowed.",
			Children:    sloListRequestFields,
		},
	},
}

// ============================================================
// Split Graph Widget
// ============================================================

// SplitGraphWidgetSpec corresponds to OpenAPI
// components/schemas/SplitGraphWidgetDefinition.
// The JSON type is "split_group" (not "split_graph").
// source_widget_definition is handled by post-processing (buildSplitGraphSourceWidgetJSON).
var SplitGraphWidgetSpec = WidgetSpec{
	HCLKey:      "split_graph_definition",
	JSONType:    "split_group",
	Description: "The definition for a Split Graph widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "size",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "Size of the individual graphs in the split.",
		},
		// has_uniform_y_axes: always emitted (OmitEmpty: false) — cassette shows false when not set
		{HCLKey: "has_uniform_y_axes", Type: TypeBool, OmitEmpty: false, Description: "Normalize y axes across graphs."},
		// split_config: TypeBlock (MaxItems:1, Required)
		{
			HCLKey:      "split_config",
			Type:        TypeBlock,
			OmitEmpty:   false,
			Required:    true,
			Description: "Encapsulates all user choices about how to split a graph.",
			Children:    splitConfigFields,
		},
		// source_widget_definition: handled by post-processing
		// title is in CommonWidgetFields
	},
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

// ============================================================
// Group Widget
// ============================================================

// GroupWidgetSpec corresponds to OpenAPI
// components/schemas/GroupWidgetDefinition.
// The nested "widget" list is handled by post-processing (buildGroupWidgetsJSON).
var GroupWidgetSpec = WidgetSpec{
	HCLKey:      "group_definition",
	JSONType:    "group",
	Description: "The definition for a Group widget.",
	Fields: []FieldSpec{
		{
			HCLKey:      "layout_type",
			Type:        TypeString,
			OmitEmpty:   false,
			Required:    true,
			Description: "The layout type of the group.",
			ValidValues: []string{"ordered"},
		},
		{HCLKey: "background_color", Type: TypeString, OmitEmpty: true, Description: "The background color of the group title, options: `vivid_blue`, `vivid_purple`, `vivid_pink`, `vivid_orange`, `vivid_yellow`, `vivid_green`, `blue`, `purple`, `pink`, `orange`, `yellow`, `green`, `gray` or `white`"},
		{HCLKey: "banner_img", Type: TypeString, OmitEmpty: true, Description: "The image URL to display as a banner for the group."},
		// Default: true — preserved from original schema (show_title defaults to visible)
		{HCLKey: "show_title", Type: TypeBool, OmitEmpty: false, Default: true, Description: "Whether to show the title or not."},
		// "widget" is handled by post-processing
		// title is in CommonWidgetFields
	},
}

// buildGroupWidgetsJSON builds the "widgets" array for a group widget.
// It iterates over the child widget list and delegates to buildWidgetEngineJSON,
// which now accepts a path string and handles all widget types uniformly.
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
// Powerpack Widget
// ============================================================

// PowerpackWidgetSpec corresponds to OpenAPI
// components/schemas/PowerpackWidgetDefinition.
var PowerpackWidgetSpec = WidgetSpec{
	HCLKey:      "powerpack_definition",
	JSONType:    "powerpack",
	Description: "The definition for a Powerpack widget.",
	Fields: []FieldSpec{
		{HCLKey: "powerpack_id", Type: TypeString, OmitEmpty: false, Required: true, Description: "UUID of the associated powerpack."},
		{HCLKey: "background_color", Type: TypeString, OmitEmpty: true, Description: "The background color of the powerpack title."},
		{HCLKey: "banner_img", Type: TypeString, OmitEmpty: true, Description: "URL of image to display as a banner for the powerpack."},
		{HCLKey: "show_title", Type: TypeBool, OmitEmpty: true, Description: "Whether to show the title of the powerpack."},
		// template_variables: TypeBlock (MaxItems:1) containing two TypeBlockLists
		{
			HCLKey:      "template_variables",
			Type:        TypeBlock,
			OmitEmpty:   true,
			Description: "The list of template variables for this powerpack.",
			Children:    powerpackTemplateVariableFields,
		},
		// title is in CommonWidgetFields
	},
}

// ============================================================
// Unified post-processing hooks — called by buildWidgetEngineJSON / flattenWidgetEngineJSON
// ============================================================
// All per-widget special-cases (Batch B inline checks + Batch C complex widgets)
// are consolidated here so the engine calls exactly one hook per widget.

// buildWidgetPostProcess runs all per-widget post-processing in the build direction.
// It covers:
//   - Formula/query blocks (Batch A/B: formula-capable widgets)
//   - Scatterplot table (Batch B)
//   - Toplist style display fix-up (Batch B)
//   - Treemap color_by injection (Batch B)
//   - Query table requests override (Batch C)
//   - Split graph source widget and static_splits (Batch C)
//   - Group nested widgets (Batch C)
func buildWidgetPostProcess(d *schema.ResourceData, spec WidgetSpec, defHCLPath string, defJSON map[string]interface{}) {
	// ---- Batch A/B: formula/query blocks ----
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

	// ---- Batch B: scatterplot table ----
	if spec.JSONType == "scatterplot" {
		buildScatterplotTableJSON(d, defHCLPath, defJSON)
	}

	// ---- Batch B: toplist style display ----
	if spec.JSONType == "toplist" {
		buildToplistStyleDisplayJSON(defJSON)
	}

	// ---- Batch B: treemap color_by ----
	if spec.JSONType == "treemap" {
		defJSON["color_by"] = "user"
	}

	// ---- Batch C: query_table requests override ----
	if spec.JSONType == "query_table" {
		requests := buildQueryTableRequestsJSON(d, defHCLPath)
		defJSON["requests"] = requests
	}

	// ---- Batch C: split_graph source widget + static_splits ----
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

	// ---- Batch C: group nested widgets ----
	if spec.JSONType == "group" {
		widgets := buildGroupWidgetsJSON(d, defHCLPath)
		defJSON["widgets"] = widgets
	}
}

// flattenWidgetPostProcess runs all per-widget post-processing in the flatten direction.
// It covers:
//   - Formula/query request flattening (Batch A/B: formula-capable widgets)
//   - Sunburst legend disambiguation (Batch B)
//   - Scatterplot table flattening (Batch B)
//   - Query table requests override (Batch C)
//   - Split graph source widget and static_splits (Batch C)
//   - Group nested widgets (Batch C)
func flattenWidgetPostProcess(spec WidgetSpec, def map[string]interface{}, defState map[string]interface{}) {
	// ---- Batch A/B: formula/query request flattening ----
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

	// ---- Batch B: sunburst legend disambiguation ----
	if spec.JSONType == "sunburst" {
		flattenSunburstLegendState(def, defState)
	}

	// ---- Batch B: scatterplot table ----
	if spec.JSONType == "scatterplot" {
		flattenScatterplotTableState(def, defState)
	}

	// ---- Batch C: query_table requests override ----
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

	// ---- Batch C: split_group source widget + static_splits ----
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

	// ---- Batch C: group nested widgets ----
	if spec.JSONType == "group" {
		if widgets, ok := def["widgets"].([]interface{}); ok {
			defState["widget"] = flattenGroupWidgetsJSON(widgets)
		}
	}
}

// ============================================================
// Widget Spec Registry
// ============================================================

// complexWidgetSpecs is populated by Batch C implementation.
var complexWidgetSpecs = []WidgetSpec{
	QueryTableWidgetSpec,
	ListStreamWidgetSpec,
	SLOWidgetSpec,
	SLOListWidgetSpec,
	SplitGraphWidgetSpec,
	GroupWidgetSpec,
	PowerpackWidgetSpec,
}
