package datadog

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/utils"
	"github.com/terraform-providers/terraform-provider-datadog/datadog/internal/validators"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kr/pretty"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogTimeboard() *schema.Resource {
	apmOrLogQuery := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"index": {
					Description: "",
					Type:        schema.TypeString,
					Required:    true,
				},
				"compute": {
					Description: "Exactly one nested block is required with the structure below.",
					Type:        schema.TypeList,
					Required:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"aggregation": {
								Description: "",
								Type:        schema.TypeString,
								Required:    true,
							},
							"facet": {
								Description: "",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"interval": {
								Description: "",
								Type:        schema.TypeInt,
								Optional:    true,
							},
						},
					},
				},
				"search": {
					Description: "Exactly one nested block is allowed with the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Description: "",
								Type:        schema.TypeString,
								Required:    true,
							},
						},
					},
				},
				"group_by": {
					Description: "Multiple nested blocks are allowed with the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"facet": {
								Description: "",
								Type:        schema.TypeString,
								Required:    true,
							},
							"limit": {
								Description: "",
								Type:        schema.TypeInt,
								Optional:    true,
							},
							"sort": {
								Description: "Exactly one nested block is allowed with the structure below.",
								Type:        schema.TypeList,
								Optional:    true,
								MaxItems:    1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"aggregation": {
											Description: "",
											Type:        schema.TypeString,
											Required:    true,
										},
										"order": {
											Description: "",
											Type:        schema.TypeString,
											Required:    true,
										},
										"facet": {
											Description: "",
											Type:        schema.TypeString,
											Optional:    true,
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	processQuery := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric": {
					Description: "",
					Type:        schema.TypeString,
					Required:    true,
				},
				"search_by": {
					Description: "",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"filter_by": {
					Description: "",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"limit": {
					Description: "",
					Type:        schema.TypeInt,
					Optional:    true,
				},
			},
		},
	}

	request := &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"q": {
					Description: "The query of the request. Pro tip: Use the JSON tab inside the Datadog UI to help build you query strings.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"log_query":     apmOrLogQuery,
				"apm_query":     apmOrLogQuery,
				"process_query": processQuery,
				"stacked": {
					Description: "Boolean value to determine if this is this a stacked area graph. Default: `false` (line chart).",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     false,
				},
				"type": {
					Description: "Choose how to draw the graph. For example: `line`, `bars` or `area`. Default: `line`.",
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "line",
				},
				"aggregator": {
					Description:  "The aggregation method used when the number of data points outnumbers the max that can be shown.",
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validators.ValidateAggregatorMethod,
				},
				"style": {
					Description: "Nested block to customize the graph style.",
					Type:        schema.TypeMap,
					Optional:    true,
				},
				"metadata_json": {
					Description: "A JSON blob (preferrably created using [jsonencode](https://www.terraform.io/docs/configuration/functions/jsonencode.html?_ga=2.6381362.1091155358.1609189257-888022054.1605547463)) representing mapping of query expressions to alias names. Note that the query expressions in `metadata_json` will be ignored if they're not present in the query.",
					Type:        schema.TypeString,
					Optional:    true,
					// NOTE: this is using functions from resource_datadog_screenboard
					// since the metadata attribute is the same for both of these boards
					ValidateFunc: validateMetadataJSON,
				},
				"conditional_format": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "A list of conditional formatting rules.",
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"palette": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "The palette to use if this condition is met.",
							},
							"comparator": {
								Type:        schema.TypeString,
								Required:    true,
								Description: "Comparator (<, >, etc)",
							},
							"custom_bg_color": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Custom background color (e.g., #205081)",
							},
							"value": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Value that is threshold for conditional format",
							},
							"custom_fg_color": {
								Type:        schema.TypeString,
								Optional:    true,
								Description: "Custom foreground color (e.g., #59afe1)",
							},
						},
					},
				},
				"change_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Type of change for change graphs.",
				},
				"order_direction": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Sort change graph in ascending or descending order.",
				},
				"compare_to": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The time period to compare change against in change graphs.",
				},
				"increase_good": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Decides whether to represent increases as good or bad in change graphs.",
				},
				"order_by": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The field a change graph will be ordered by.",
				},
				"extra_col": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "If set to 'present', this will include the present values in change graphs.",
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						stringVal := val.(string)
						if stringVal != "" && stringVal != "present" {
							errs = append(errs, fmt.Errorf("'%s' value must be empty or 'present', got: '%s'", key, stringVal))
						}
						return
					},
				},
			},
		},
	}

	marker := &schema.Schema{
		Description: "Nested block describing lines / ranges added to graph for formatting. The structure of this block is described below. Multiple `marker` blocks are allowed within a `graph` block.",
		Type:        schema.TypeList,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Description: "How the marker lines will look. Possible values are one of {`error`, `warning`, `info`, `ok`} combined with one of {`dashed`, `solid`, `bold`}. Example: `error dashed`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"value": {
					Description: "Mathematical expression describing the marker. Examples: `y > 1`, `-5 < y < 0`, `y = 19`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"label": {
					Description: "A label for the line or range. **Warning**: when a label is enabled but left empty through the UI, the Datadog API returns a boolean value, not a string. This makes terraform plan fail with a JSON decoding error.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
	}

	graph := &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		Description: "A list of graph definitions.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"title": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the graph.",
				},
				"events": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "Filter for events to be overlayed on the graph.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"viz": {
					Description: "The type of visualization to use for the graph. Valid choices are `change`, `distribution`, `heatmap`, `hostmap`, `query_value`, `timeseries`, and `toplist`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"request": request,
				"marker":  marker,
				"yaxis": {
					Description: "Nested block describing modifications to the Y-axis rendering. The structure of this block is described below.",
					Type:        schema.TypeMap,
					Optional:    true,
					// `include_zero` and `include_units` are bool but Terraform treats them as strings
					// as part of the `yaxis` map so we suppress the diff when
					// value in the state and value from the api are the same
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						var oldBool, newBool bool
						var err error

						if oldBool, err = strconv.ParseBool(old); err != nil {
							return false
						}

						if newBool, err = strconv.ParseBool(new); err != nil {
							return false
						}

						return oldBool == newBool
					},
				},
				"autoscale": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Automatically scale graphs",
				},
				"text_align": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "How to align text",
				},
				"precision": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "How many digits to show",
					// Suppress the diff shown if the graph is going to be set to a default of 2 iff its not set by the config
					// Since this precision attribute is only valid for certain graph types, we aren't setting a global default
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						if old == "2" && new == "" {
							return true
						}
						return false
					},
				},
				"custom_unit": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Use a custom unit (like 'users')",
				},
				"style": {
					Description: "Nested block describing hostmaps. The structure of this block is described below.",
					Type:        schema.TypeMap,
					Optional:    true,
					// `palette_flip` is bool but Terraform treats it as a string
					// as part of the `style` map so we suppress the diff when
					// value in the state and value from the api are the same
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						var oldBool, newBool bool
						var err error

						if oldBool, err = strconv.ParseBool(old); err != nil {
							return false
						}

						if newBool, err = strconv.ParseBool(new); err != nil {
							return false
						}

						return oldBool == newBool
					},
				},
				"group": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "A list of groupings for hostmap type graphs.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"include_no_metric_hosts": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Include hosts without metrics in hostmap graphs",
				},
				"scope": {
					Type:        schema.TypeList,
					Optional:    true,
					Description: "A list of scope filters for hostmap type graphs.",
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
				"include_ungrouped_hosts": {
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Include ungrouped hosts in hostmap graphs",
				},
				"node_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Type of nodes to show in hostmap graphs (either 'host' or 'container').",
				},
			},
		},
	}

	templateVariable := &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "A list of template variables for using Dashboard templating.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the variable.",
				},
				"prefix": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The tag prefix associated with the variable. Only tags with this prefix will appear in the variable dropdown.",
				},
				"default": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The default value for the template variable on dashboard load.",
				},
			},
		},
	}

	return &schema.Resource{
		Description:        "Provides a Datadog timeboard resource. This can be used to create and manage Datadog timeboards.",
		DeprecationMessage: "This resource is deprecated. Instead use the Dashboard resource",
		Create:             resourceDatadogTimeboardCreate,
		Update:             resourceDatadogTimeboardUpdate,
		Read:               resourceDatadogTimeboardRead,
		Delete:             resourceDatadogTimeboardDelete,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogTimeboardImport,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the dashboard.",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A description of the dashboard's content.",
			},
			"read_only": {
				Description: "The read-only status of the timeboard. Default is false.",
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
			},
			"graph":             graph,
			"template_variable": templateVariable,
		},
	}
}

func appendConditionalFormats(datadogRequest *datadog.GraphDefinitionRequest, terraformFormats *[]interface{}) {
	for _, tf := range *terraformFormats {
		t := tf.(map[string]interface{})
		d := datadog.DashboardConditionalFormat{
			Comparator: datadog.String(t["comparator"].(string)),
		}

		if v, ok := t["palette"]; ok {
			d.SetPalette(v.(string))
		}

		if v, ok := t["custom_bg_color"]; ok {
			d.SetCustomBgColor(v.(string))
		}

		if v, ok := t["custom_fg_color"]; ok {
			d.SetCustomFgColor(v.(string))
		}

		if v, ok := t["value"]; ok {
			d.SetValue(json.Number(v.(string)))
		}

		datadogRequest.ConditionalFormats = append(datadogRequest.ConditionalFormats, d)
	}
}

func buildTemplateVariables(terraformTemplateVariables *[]interface{}) *[]datadog.TemplateVariable {
	datadogTemplateVariables := make([]datadog.TemplateVariable, len(*terraformTemplateVariables))
	for i, tpv := range *terraformTemplateVariables {
		t := tpv.(map[string]interface{})
		datadogTemplateVariables[i] = datadog.TemplateVariable{
			Name:    datadog.String(t["name"].(string)),
			Prefix:  datadog.String(t["prefix"].(string)),
			Default: datadog.String(t["default"].(string)),
		}
	}
	return &datadogTemplateVariables
}

func buildDatadogGraphApmOrLogQuery(source interface{}) *datadog.GraphApmOrLogQuery {
	terraformLogOrApmQuery := source.(map[string]interface{})
	// Index
	datadogQuery := datadog.GraphApmOrLogQuery{
		Index: datadog.String(terraformLogOrApmQuery["index"].(string)),
	}
	// Compute
	if compute, ok := terraformLogOrApmQuery["compute"].([]interface{}); ok && len(compute) > 0 {
		terraformCompute := compute[0].(map[string]interface{})
		datadogCompute := datadog.GraphApmOrLogQueryCompute{}
		if v, ok := terraformCompute["aggregation"].(string); ok && len(v) != 0 {
			datadogCompute.SetAggregation(v)
		}
		if v, ok := terraformCompute["facet"].(string); ok && len(v) != 0 {
			datadogCompute.SetFacet(v)
		}
		if v, ok := terraformCompute["interval"].(int); ok && v != 0 {
			datadogCompute.SetInterval(v)
		}
		datadogQuery.Compute = &datadogCompute
	}
	// Search
	if search, ok := terraformLogOrApmQuery["search"].([]interface{}); ok && len(search) > 0 {
		terraformSearch := search[0].(map[string]interface{})
		datadogQuery.Search = &datadog.GraphApmOrLogQuerySearch{
			Query: datadog.String(terraformSearch["query"].(string)),
		}
	}
	// GroupBy
	if terraformGroupBys, ok := terraformLogOrApmQuery["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]datadog.GraphApmOrLogQueryGroupBy, len(terraformGroupBys))
		for i, g := range terraformGroupBys {
			groupBy := g.(map[string]interface{})
			// Facet
			datadogGroupBy := datadog.GraphApmOrLogQueryGroupBy{
				Facet: datadog.String(groupBy["facet"].(string)),
			}
			// Limit
			if v, ok := groupBy["limit"].(int); ok && v != 0 {
				datadogGroupBy.Limit = &v
			}
			// Sort
			if sort, ok := groupBy["sort"].([]interface{}); ok && len(sort) > 0 {
				terraformSort := sort[0].(map[string]interface{})
				datadogGroupBy.Sort = &datadog.GraphApmOrLogQueryGroupBySort{
					Aggregation: datadog.String(terraformSort["aggregation"].(string)),
					Order:       datadog.String(terraformSort["order"].(string)),
				}
				if facet, ok := terraformSort["facet"].(string); ok && len(facet) != 0 {
					datadogGroupBy.Sort.Facet = datadog.String(terraformSort["facet"].(string))
				}
			}
			datadogGroupBys[i] = datadogGroupBy
		}
		datadogQuery.GroupBy = datadogGroupBys
	}
	return &datadogQuery
}

func buildDatadogGraphProcessQuery(terraformQuery map[string]interface{}) *datadog.GraphProcessQuery {
	datadogQuery := datadog.GraphProcessQuery{}
	if v, ok := terraformQuery["metric"].(string); ok && len(v) != 0 {
		datadogQuery.SetMetric(v)
	}
	if v, ok := terraformQuery["search_by"].(string); ok {
		datadogQuery.SetSearchBy(v)
	}

	if v, ok := terraformQuery["filter_by"].([]interface{}); ok {
		filters := make([]string, len(v))
		for i, filter := range v {
			filters[i] = filter.(string)
		}
		datadogQuery.FilterBy = filters
	}

	if v, ok := terraformQuery["limit"].(int); ok && v != 0 {
		datadogQuery.SetLimit(v)
	}

	return &datadogQuery
}

func appendRequests(datadogGraph *datadog.Graph, terraformRequests *[]interface{}) error {
	for _, r := range *terraformRequests {
		t := r.(map[string]interface{})
		log.Printf("[DataDog] request: %v", pretty.Sprint(t))
		d := datadog.GraphDefinitionRequest{
			Type:       datadog.String(t["type"].(string)),
			Aggregator: datadog.String(t["aggregator"].(string)),
		}

		if v, ok := t["q"].(string); ok && len(v) != 0 {
			d.SetQuery(v)
		} else if v, ok := t["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			d.LogQuery = buildDatadogGraphApmOrLogQuery(logQuery)
		} else if v, ok := t["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			d.ApmQuery = buildDatadogGraphApmOrLogQuery(apmQuery)
		} else if v, ok := t["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			d.ProcessQuery = buildDatadogGraphProcessQuery(processQuery)
		}

		if stacked, ok := t["stacked"]; ok {
			d.SetStacked(stacked.(bool))
		}
		if style, ok := t["style"]; ok {
			s, _ := style.(map[string]interface{})

			style := datadog.GraphDefinitionRequestStyle{}

			if v, ok := s["palette"]; ok {
				style.SetPalette(v.(string))
			}

			if v, ok := s["width"]; ok {
				style.SetWidth(v.(string))
			}

			if v, ok := s["type"]; ok {
				style.SetType(v.(string))
			}

			d.SetStyle(style)
		}

		if v, ok := t["change_type"]; ok {
			d.SetChangeType(v.(string))
		}
		if v, ok := t["compare_to"]; ok {
			d.SetCompareTo(v.(string))
		}
		if v, ok := t["increase_good"]; ok {
			d.SetIncreaseGood(v.(bool))
		}
		if v, ok := t["order_by"]; ok {
			d.SetOrderBy(v.(string))
		}
		if v, ok := t["extra_col"]; ok {
			// additional validation: `extra_col` may only be used for `change` viz
			if viz := datadogGraph.Definition.GetViz(); viz == "change" {
				d.SetExtraCol(v.(string))
			} else if v != nil && v != "" {
				return fmt.Errorf("'extra_col' attribute may only be used for 'change' viz, not '%s'", viz)
			}
		}
		if v, ok := t["order_direction"]; ok {
			d.SetOrderDirection(v.(string))
		}

		if v, ok := t["conditional_format"]; ok {
			value := v.([]interface{})
			appendConditionalFormats(&d, &value)
		}
		if v, ok := t["metadata_json"]; ok {
			d.Metadata = map[string]datadog.GraphDefinitionMetadata{}
			utils.GetMetadataFromJSON([]byte(v.(string)), &d.Metadata)
		}
		datadogGraph.Definition.Requests = append(datadogGraph.Definition.Requests, d)
	}

	return nil
}

func appendEvents(datadogGraph *datadog.Graph, terraformEvents *[]interface{}) {
	for _, e := range *terraformEvents {
		datadogGraph.Definition.Events = append(datadogGraph.Definition.Events, datadog.GraphEvent{
			Query: datadog.String(e.(string)),
		})
	}
}

func appendMarkers(datadogGraph *datadog.Graph, terraformMarkers *[]interface{}) {
	for _, m := range *terraformMarkers {
		t := m.(map[string]interface{})
		d := datadog.GraphDefinitionMarker{
			Type:  datadog.String(t["type"].(string)),
			Value: datadog.String(t["value"].(string)),
		}
		if v, ok := t["label"]; ok {
			d.SetLabel(v.(string))
		}
		datadogGraph.Definition.Markers = append(datadogGraph.Definition.Markers, d)
	}
}

func buildGraphs(terraformGraphs *[]interface{}) (*[]datadog.Graph, error) {
	datadogGraphs := make([]datadog.Graph, len(*terraformGraphs))
	for i, g := range *terraformGraphs {
		t := g.(map[string]interface{})

		datadogGraphs[i] = datadog.Graph{
			Title: datadog.String(t["title"].(string)),
		}

		d := &datadogGraphs[i]
		d.Definition = &datadog.GraphDefinition{}
		d.Definition.SetViz(t["viz"].(string))

		if v, ok := t["yaxis"]; ok {
			yaxis := v.(map[string]interface{})
			if v, ok := yaxis["min"]; ok {
				min, _ := strconv.ParseFloat(v.(string), 64)
				d.Definition.Yaxis.SetMin(min)
			}
			if v, ok := yaxis["max"]; ok {
				max, _ := strconv.ParseFloat(v.(string), 64)
				d.Definition.Yaxis.SetMax(max)
			}
			if v, ok := yaxis["scale"]; ok {
				d.Definition.Yaxis.SetScale(v.(string))
			}
			if v, ok := yaxis["include_zero"]; ok {
				b, _ := strconv.ParseBool(v.(string))
				d.Definition.Yaxis.SetIncludeZero(b)
			}

			if v, ok := yaxis["include_units"]; ok {
				b, _ := strconv.ParseBool(v.(string))
				d.Definition.Yaxis.SetIncludeUnits(b)
			}
		}

		if v, ok := t["autoscale"]; ok {
			d.Definition.SetAutoscale(v.(bool))
		}

		if v, ok := t["text_align"]; ok {
			d.Definition.SetTextAlign(v.(string))
		}

		if precision, ok := t["precision"]; ok {
			val := precision.(string)
			if val != "" {
				d.Definition.SetPrecision(datadog.PrecisionT(val))
			}
		}

		if v, ok := t["custom_unit"]; ok {
			d.Definition.SetCustomUnit(v.(string))
		}

		if style, ok := t["style"]; ok {
			s := style.(map[string]interface{})

			gs := datadog.Style{}

			if v, ok := s["palette"]; ok {
				gs.SetPalette(v.(string))
			}

			if v, ok := s["palette_flip"]; ok {
				pf, _ := strconv.ParseBool(v.(string))
				gs.SetPaletteFlip(pf)
			}

			if v, ok := s["fill_min"]; ok {
				gs.SetFillMin(json.Number(v.(string)))
			}

			if v, ok := s["fill_max"]; ok {
				gs.SetFillMax(json.Number(v.(string)))
			}

			d.Definition.SetStyle(gs)
		}

		if v, ok := t["group"]; ok {
			for _, g := range v.([]interface{}) {
				d.Definition.Groups = append(d.Definition.Groups, g.(string))
			}
		}

		if includeNoMetricHosts, ok := t["include_no_metric_hosts"]; ok {
			d.Definition.SetIncludeNoMetricHosts(includeNoMetricHosts.(bool))
		}

		if v, ok := t["scope"]; ok {
			for _, s := range v.([]interface{}) {
				d.Definition.Scopes = append(d.Definition.Scopes, s.(string))
			}
		}

		if v, ok := t["include_ungrouped_hosts"]; ok {
			d.Definition.SetIncludeUngroupedHosts(v.(bool))
		}

		if v, ok := t["node_type"]; ok {
			d.Definition.SetNodeType(v.(string))
		}

		v := t["marker"].([]interface{})
		appendMarkers(d, &v)

		v = t["events"].([]interface{})
		appendEvents(d, &v)

		v = t["request"].([]interface{})
		err := appendRequests(d, &v)
		if err != nil {
			return nil, err
		}
	}
	return &datadogGraphs, nil
}

func buildTimeboard(d *schema.ResourceData) (*datadog.Dashboard, error) {
	var id int
	if d.Id() != "" {
		var err error
		id, err = strconv.Atoi(d.Id())
		if err != nil {
			return nil, err
		}
	}
	terraformGraphs := d.Get("graph").([]interface{})
	terraformTemplateVariables := d.Get("template_variable").([]interface{})
	graphs, err := buildGraphs(&terraformGraphs)
	if err != nil {
		return nil, err
	}
	return &datadog.Dashboard{
		Id:                datadog.Int(id),
		Title:             datadog.String(d.Get("title").(string)),
		Description:       datadog.String(d.Get("description").(string)),
		ReadOnly:          datadog.Bool(d.Get("read_only").(bool)),
		Graphs:            *graphs,
		TemplateVariables: *buildTemplateVariables(&terraformTemplateVariables),
	}, nil
}

func resourceDatadogTimeboardCreate(d *schema.ResourceData, meta interface{}) error {
	timeboard, err := buildTimeboard(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	timeboard, err = client.CreateDashboard(timeboard)
	if err != nil {
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error creating timeboard")
	}
	d.SetId(strconv.Itoa(timeboard.GetId()))
	return resourceDatadogTimeboardRead(d, meta)
}

func resourceDatadogTimeboardUpdate(d *schema.ResourceData, meta interface{}) error {
	timeboard, err := buildTimeboard(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if err = client.UpdateDashboard(timeboard); err != nil {
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error updating timeboard")
	}
	return resourceDatadogTimeboardRead(d, meta)
}
func buildTFGraphProcessQuery(datadogQuery datadog.GraphProcessQuery) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	if datadogQuery.Metric != nil {
		terraformQuery["metric"] = *datadogQuery.Metric
	}
	if datadogQuery.SearchBy != nil {
		terraformQuery["search_by"] = *datadogQuery.SearchBy
	}
	if datadogQuery.FilterBy != nil {
		terraformFilterBy := make([]string, len(datadogQuery.FilterBy))
		for i, datadogFilterBy := range datadogQuery.FilterBy {
			terraformFilterBy[i] = datadogFilterBy
		}
		terraformQuery["filter_by"] = terraformFilterBy
	}
	if datadogQuery.Limit != nil {
		terraformQuery["limit"] = *datadogQuery.Limit
	}

	return terraformQuery
}

func buildTFGraphApmOrLogQuery(datadogQuery datadog.GraphApmOrLogQuery) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	// Index
	terraformQuery["index"] = *datadogQuery.Index
	// Compute
	if datadogQuery.Compute != nil {
		terraformCompute := map[string]interface{}{
			"aggregation": *datadogQuery.Compute.Aggregation,
		}
		if datadogQuery.Compute.Facet != nil {
			terraformCompute["facet"] = *datadogQuery.Compute.Facet
		}
		if datadogQuery.Compute.Interval != nil {
			terraformCompute["interval"] = *datadogQuery.Compute.Interval
		}
		terraformQuery["compute"] = []map[string]interface{}{terraformCompute}
	}
	// Search
	if datadogQuery.Search != nil {
		terraformSearch := map[string]interface{}{
			"query": *datadogQuery.Search.Query,
		}
		terraformQuery["search"] = []map[string]interface{}{terraformSearch}
	}
	// GroupBy
	if datadogQuery.GroupBy != nil {
		terraformGroupBys := make([]map[string]interface{}, len(datadogQuery.GroupBy))
		for i, groupBy := range datadogQuery.GroupBy {
			// Facet
			terraformGroupBy := map[string]interface{}{
				"facet": *groupBy.Facet,
			}
			// Limit
			if groupBy.Limit != nil {
				terraformGroupBy["limit"] = *groupBy.Limit
			}
			// Sort
			if groupBy.Sort != nil {
				sort := map[string]interface{}{
					"aggregation": *groupBy.Sort.Aggregation,
					"order":       *groupBy.Sort.Order,
				}
				if groupBy.Sort.Facet != nil {
					sort["facet"] = *groupBy.Sort.Facet
				}
				terraformGroupBy["sort"] = []map[string]interface{}{sort}
			}

			terraformGroupBys[i] = terraformGroupBy
		}
		terraformQuery["group_by"] = &terraformGroupBys
	}
	return terraformQuery
}

func appendTerraformGraphRequests(datadogRequests []datadog.GraphDefinitionRequest, requests *[]map[string]interface{}) {
	for _, datadogRequest := range datadogRequests {
		request := map[string]interface{}{}
		if v, ok := datadogRequest.GetQueryOk(); ok {
			request["q"] = v
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTFGraphApmOrLogQuery(*datadogRequest.ApmQuery)
			request["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTFGraphApmOrLogQuery(*datadogRequest.LogQuery)
			request["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTFGraphProcessQuery(*datadogRequest.ProcessQuery)
			request["process_query"] = []map[string]interface{}{terraformQuery}
		}
		if v, ok := datadogRequest.GetStackedOk(); ok {
			request["stacked"] = v
		}
		if v, ok := datadogRequest.GetTypeOk(); ok {
			request["type"] = v
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := map[string]string{}
			if v, ok := v.GetPaletteOk(); ok {
				style["palette"] = v
			}
			if v, ok := v.GetTypeOk(); ok {
				style["type"] = v
			}
			if v, ok := v.GetWidthOk(); ok {
				style["width"] = v
			}
			request["style"] = style
		}
		var conditionalFormats []map[string]interface{}
		for _, cf := range datadogRequest.ConditionalFormats {
			conditionalFormat := map[string]interface{}{}
			if v, ok := cf.GetPaletteOk(); ok {
				conditionalFormat["palette"] = v
			}
			if v, ok := cf.GetComparatorOk(); ok {
				conditionalFormat["comparator"] = v
			}
			if v, ok := cf.GetCustomBgColorOk(); ok {
				conditionalFormat["custom_bg_color"] = v
			}
			if v, ok := cf.GetValueOk(); ok {
				conditionalFormat["value"] = v
			}
			if v, ok := cf.GetCustomFgColorOk(); ok {
				conditionalFormat["custom_fg_color"] = v
			}
			conditionalFormats = append(conditionalFormats, conditionalFormat)
		}
		request["conditional_format"] = conditionalFormats
		if v, ok := datadogRequest.GetAggregatorOk(); ok {
			request["aggregator"] = v
		}
		if v, ok := datadogRequest.GetChangeTypeOk(); ok {
			request["change_type"] = v
		}
		if v, ok := datadogRequest.GetOrderDirectionOk(); ok {
			request["order_direction"] = v
		}
		if v, ok := datadogRequest.GetCompareToOk(); ok {
			request["compare_to"] = v
		}
		if v, ok := datadogRequest.GetIncreaseGoodOk(); ok {
			request["increase_good"] = v
		}
		if v, ok := datadogRequest.GetOrderByOk(); ok {
			request["order_by"] = v
		}
		if v, ok := datadogRequest.GetExtraColOk(); ok {
			request["extra_col"] = v
		}
		if datadogRequest.Metadata != nil {
			res, _ := json.Marshal(datadogRequest.Metadata)
			request["metadata_json"] = string(res)
		}

		*requests = append(*requests, request)
	}
}

func buildTerraformGraph(datadogGraph datadog.Graph) map[string]interface{} {
	graph := map[string]interface{}{}
	graph["title"] = datadogGraph.GetTitle()

	definition := datadogGraph.Definition
	graph["viz"] = definition.GetViz()

	var events []string
	for _, e := range definition.Events {
		if v, ok := e.GetQueryOk(); ok {
			events = append(events, v)
		}
	}
	if len(events) > 0 {
		graph["events"] = events
	}

	var markers []map[string]interface{}
	for _, datadogMarker := range definition.Markers {
		marker := map[string]interface{}{}
		if v, ok := datadogMarker.GetTypeOk(); ok {
			marker["type"] = v
		}
		if v, ok := datadogMarker.GetValueOk(); ok {
			marker["value"] = v
		}
		if v, ok := datadogMarker.GetLabelOk(); ok {
			marker["label"] = v
		}
		markers = append(markers, marker)
	}
	graph["marker"] = markers

	yaxis := map[string]string{}

	if v, ok := definition.Yaxis.GetMinOk(); ok {
		yaxis["min"] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	if v, ok := definition.Yaxis.GetMaxOk(); ok {
		yaxis["max"] = strconv.FormatFloat(v, 'f', -1, 64)
	}

	if v, ok := definition.Yaxis.GetScaleOk(); ok {
		yaxis["scale"] = v
	}

	if v, ok := definition.Yaxis.GetIncludeZeroOk(); ok {
		yaxis["include_zero"] = strconv.FormatBool(v)
	}

	if v, ok := definition.Yaxis.GetIncludeUnitsOk(); ok {
		yaxis["include_units"] = strconv.FormatBool(v)
	}

	graph["yaxis"] = yaxis

	if v, ok := definition.GetAutoscaleOk(); ok {
		graph["autoscale"] = v
	}
	if v, ok := definition.GetTextAlignOk(); ok {
		graph["text_align"] = v
	}
	if v, ok := definition.GetPrecisionOk(); ok {
		graph["precision"] = v
	}
	if v, ok := definition.GetCustomUnitOk(); ok {
		graph["custom_unit"] = v
	}

	if v, ok := definition.GetStyleOk(); ok {
		style := map[string]string{}
		if v, ok := v.GetPaletteOk(); ok {
			style["palette"] = v
		}
		if v, ok := v.GetPaletteFlipOk(); ok {
			style["palette_flip"] = strconv.FormatBool(v)
		}
		if v, ok := v.GetFillMinOk(); ok {
			style["fill_min"] = string(v)
		}
		if v, ok := v.GetFillMaxOk(); ok {
			style["fill_max"] = string(v)
		}
		graph["style"] = style
	}
	if definition.Groups != nil {
		graph["group"] = definition.Groups
	}
	if definition.Scopes != nil {
		graph["scope"] = definition.Scopes
	}
	if v, ok := definition.GetIncludeNoMetricHostsOk(); ok {
		graph["include_no_metric_hosts"] = v
	}
	if v, ok := definition.GetIncludeUngroupedHostsOk(); ok {
		graph["include_ungrouped_hosts"] = v
	}
	if v, ok := definition.GetNodeTypeOk(); ok {
		graph["node_type"] = v
	}

	var requests []map[string]interface{}
	appendTerraformGraphRequests(definition.Requests, &requests)
	graph["request"] = requests

	return graph
}

func resourceDatadogTimeboardRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	timeboard, err := client.GetDashboard(id)
	if err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error getting timeboard")
	}
	log.Printf("[DataDog] timeboard: %v", pretty.Sprint(timeboard))
	if err := d.Set("title", timeboard.GetTitle()); err != nil {
		return err
	}
	if err := d.Set("description", timeboard.GetDescription()); err != nil {
		return err
	}

	var graphs []map[string]interface{}
	for _, datadogGraph := range timeboard.Graphs {
		graphs = append(graphs, buildTerraformGraph(datadogGraph))
	}
	log.Printf("[DataDog] graphs: %v", pretty.Sprint(graphs))
	if err := d.Set("graph", graphs); err != nil {
		return err
	}

	var templateVariables []map[string]string
	for _, templateVariable := range timeboard.TemplateVariables {
		tv := map[string]string{}
		if v, ok := templateVariable.GetNameOk(); ok {
			tv["name"] = v
		}
		if v, ok := templateVariable.GetPrefixOk(); ok {
			tv["prefix"] = v
		}
		if v, ok := templateVariable.GetDefaultOk(); ok {
			tv["default"] = v
		}
		templateVariables = append(templateVariables, tv)
	}
	if err := d.Set("template_variable", templateVariables); err != nil {
		return err
	}
	// Ensure the ID saved in the state is always the legacy ID returned from the API
	// and not the ID passed to the import statement which could be in the new ID format
	d.SetId(strconv.Itoa(timeboard.GetId()))
	return nil
}

func resourceDatadogTimeboardDelete(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if err = client.DeleteDashboard(id); err != nil {
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error deleting timeboard")
	}
	return nil
}

func resourceDatadogTimeboardImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogTimeboardRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
