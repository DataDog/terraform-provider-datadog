package datadog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/kr/pretty"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogScreenboard() *schema.Resource {

	tileDefEvent := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"q": {
					Type:     schema.TypeString,
					Required: true,
				},
			},
		},
	}

	tileDefMarker := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:     schema.TypeString,
					Required: true,
				},
				"value": {
					Type:     schema.TypeString,
					Required: true,
				},
				"label": {
					Type:     schema.TypeString,
					Optional: true,
				},
			},
		},
	}

	tileDefRequestConditionalFormat := &schema.Schema{
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
				"color": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Custom color (e.g., #205081)",
				},
				"value": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Value that is threshold for conditional format",
				},
				"custom_bg_color": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Custom  background color (e.g., #205081)",
				},
				"invert": {
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
	}

	tileDefApmOrLogQuery := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"index": {
					Type:     schema.TypeString,
					Required: true,
				},
				"compute": {
					Type:     schema.TypeList,
					Required: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"aggregation": {
								Type:     schema.TypeString,
								Required: true,
							},
							"facet": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"interval": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"search": {
					Type:     schema.TypeList,
					Optional: true,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"group_by": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"facet": {
								Type:     schema.TypeString,
								Required: true,
							},
							"limit": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"sort": {
								Type:     schema.TypeList,
								Optional: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"aggregation": {
											Type:     schema.TypeString,
											Required: true,
										},
										"order": {
											Type:     schema.TypeString,
											Required: true,
										},
										"facet": {
											Type:     schema.TypeString,
											Optional: true,
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

	tileDefProcessQuery := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric": {
					Type:     schema.TypeString,
					Required: true,
				},
				"search_by": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"filter_by": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
			},
		},
	}

	tileDefRequest := &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"q": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"log_query":     tileDefApmOrLogQuery,
				"apm_query":     tileDefApmOrLogQuery,
				"process_query": tileDefProcessQuery,
				"type": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"query_type": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"metric": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"text_filter": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"tag_filters": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"limit": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"style": {
					Type:     schema.TypeMap,
					Optional: true,
				},
				"conditional_format": tileDefRequestConditionalFormat,
				"aggregator": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateAggregatorMethod,
				},
				"compare_to": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"change_type": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"order_by": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"order_dir": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"extra_col": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"increase_good": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"metadata_json": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateMetadataJSON,
				},
			},
		},
	}

	tileDef := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"event":   tileDefEvent,
				"marker":  tileDefMarker,
				"request": tileDefRequest,
				"viz": {
					Type:     schema.TypeString,
					Required: true,
				},
				"custom_unit": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"autoscale": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"precision": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"text_align": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"node_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['host', 'container']",
				},
				"scope": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"group": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"no_group_hosts": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"no_metric_hosts": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"style": {
					Type:     schema.TypeMap,
					Optional: true,
				},
			}},
	}

	widget := &schema.Schema{
		Type:        schema.TypeList,
		Required:    true,
		Description: "A list of widget definitions.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"type": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "The type of the widget. One of [ 'free_text', 'timeseries', 'query_value', 'toplist', 'change', 'event_timeline', 'event_stream', 'image', 'note', 'alert_graph', 'alert_value', 'iframe', 'check_status', 'trace_service', 'hostmap', 'manage_status', 'log_stream', 'uptime', 'process']",
				},
				"title": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name of the widget.",
				},
				"title_align": {
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "left",
					Description: "The alignment of the widget's title.",
				},
				"title_size": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     16,
					Description: "The size of the widget's title.",
				},
				"height": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     15,
					Description: "The height of the widget.",
				},
				"width": {
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     50,
					Description: "The width of the widget.",
				},
				"x": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The position of the widget on the x axis.",
				},
				"y": {
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The position of the widget on the y axis.",
				},
				"text": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "For widgets of type 'free_text', the text to use.",
				},
				"text_size": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "auto",
				},
				"text_align": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"color": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"bgcolor": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"font_size": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"unit": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"alert_id": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"auto_refresh": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"legend": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"query": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"legend_size": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"url": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"precision": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"tags": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"viz_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['timeseries', 'toplist']",
				},
				"check": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"group": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"grouping": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['cluster', 'check']",
				},
				"group_by": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"tick_pos": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"tick_edge": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"html": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"tick": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"event_size": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"sizing": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['center', 'zoom', 'fit']",
				},
				"margin": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['small', 'large']",
				},
				"tile_def": tileDef,
				"time": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"env": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"service_service": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"service_name": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"size_version": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"layout_version": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"must_show_hits": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_errors": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_latency": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_breakdown": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_distribution": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_resource_list": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"summary_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['monitors', 'groups', 'combined']",
					ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
						v := val.(string)
						summaryTypes := []string{"monitors", "groups", "combined"}
						for _, t := range summaryTypes {
							if v == t {
								return
							}
						}
						errs = append(errs, fmt.Errorf("%q must be one of: %q, got: %q", key, summaryTypes, v))
						return
					},
				},
				"display_format": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['counts', 'list', 'countsAndList']",
				},
				"color_preference": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['background', 'text']",
				},
				"hide_zero_counts": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"show_last_triggered": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"params": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"manage_status_show_title": {
					Type:     schema.TypeBool,
					Optional: true,
				},
				"manage_status_title_text": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"manage_status_title_size": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"manage_status_title_align": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"columns": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"logset": {
					Type:     schema.TypeString,
					Optional: true,
				},
				"timeframes": {
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"rule": {
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"threshold": {
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"timeframe": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"color": {
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"monitor": {
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
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
		DeprecationMessage: "This resource is deprecated. Instead use the Dashboard resource",
		Create:             resourceDatadogScreenboardCreate,
		Read:               resourceDatadogScreenboardRead,
		Update:             resourceDatadogScreenboardUpdate,
		Delete:             resourceDatadogScreenboardDelete,
		Exists:             resourceDatadogScreenboardExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogScreenboardImport,
		},

		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the screenboard",
			},
			"height": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Height of the screenboard",
			},
			"width": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Width of the screenboard",
			},
			"shared": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the screenboard is shared or not",
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"widget":            widget,
			"template_variable": templateVariable,
		},
	}
}

// #######################################################################################
// # Convenience functions to safely pass info from Terraform to the Datadog API wrapper #
// #######################################################################################

func getMetadataFromJSON(jsonBytes []byte, unmarshalled interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(jsonBytes))
	// make sure we return errors on attributes that we don't expect in metadata
	decoder.DisallowUnknownFields()
	err := decoder.Decode(unmarshalled)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata_json: %s", err)
	}
	return nil
}

func validateMetadataJSON(v interface{}, k string) (ws []string, errors []error) {
	err := getMetadataFromJSON([]byte(v.(string)), &map[string]datadog.TileDefMetadata{})
	if err != nil {
		errors = append(errors, fmt.Errorf("%q contains an invalid JSON: %s", k, err))
	}
	return
}

func setStringFromDict(dict map[string]interface{}, key string, field **string) {
	if v, ok := dict[key]; ok && v != nil {
		*field = datadog.String(v.(string))
	}
}

func setIntFromDict(dict map[string]interface{}, key string, field **int) {
	if v, ok := dict[key]; ok {
		*field = datadog.Int(v.(int))
	}
}

func setBoolFromDict(dict map[string]interface{}, key string, field **bool) {
	if v, ok := dict[key]; ok {
		*field = datadog.Bool(v.(bool))
	}
}

// For setJSONNumberFromDict, dict[key] is expected to be a float64
func setJSONNumberFromDict(dict map[string]interface{}, key string, field **json.Number) {
	if v, ok := dict[key]; ok {
		// style fields can be numbers or strings so we need to handle both types
		var number string

		if val, ok := v.(float64); ok {
			number = strconv.FormatFloat(val, 'e', -1, 64)
		} else {
			number = v.(string)
		}
		f := json.Number(number)
		*field = &f
	}
}

// For setPrecisionTFromDict, dict[key] is expected to be a int or a string like "100%" or "*"
func setPrecisionTFromDict(dict map[string]interface{}, key string, field **datadog.PrecisionT) {
	iface, ok := dict[key]
	if !ok {
		return
	}

	switch value := iface.(type) {
	case int:
		f := datadog.PrecisionT(strconv.FormatInt(int64(value), 10))
		*field = &f
	case string:
		f := datadog.PrecisionT(value)
		*field = &f
	}
}

func setStringListFromDict(dict map[string]interface{}, key string, field *[]*string) {
	if v, ok := dict[key].([]interface{}); ok {
		*field = []*string{}
		for _, text := range v {
			*field = append(*field, datadog.String(text.(string)))
		}
	}
}

func setMetadataFromDict(dict map[string]interface{}, key string, field *map[string]datadog.TileDefMetadata) {
	if v, ok := dict[key].(map[string]datadog.TileDefMetadata); ok {
		*field = v
	}
}

func setFromDict(dict map[string]interface{}, key string, field interface{}) {
	switch field.(type) {
	case **string:
		setStringFromDict(dict, key, field.(**string))
	case **int:
		setIntFromDict(dict, key, field.(**int))
	case **bool:
		setBoolFromDict(dict, key, field.(**bool))
	case **json.Number:
		setJSONNumberFromDict(dict, key, field.(**json.Number))
	case **datadog.PrecisionT:
		setPrecisionTFromDict(dict, key, field.(**datadog.PrecisionT))
	case *map[string]datadog.TileDefMetadata:
		setMetadataFromDict(dict, key, field.(*map[string]datadog.TileDefMetadata))
	case *[]*string:
		setStringListFromDict(dict, key, field.(*[]*string))
	default:
		panic(fmt.Sprintf("Cannot call setFromDict() for unsupported type %T (key: %v)", field, key))
	}
}

type match struct {
	key   string
	field interface{}
}

type batch struct {
	dict    map[string]interface{}
	matches []match
}

func batchSetFromDict(b batch) {
	for _, match := range b.matches {
		setFromDict(b.dict, match.key, match.field)
	}
}

// ####################################################################################
// # Handle passing info from Terraform to the Datadog API wrapper                    #
// ####################################################################################

func buildTileDefRequestsConditionalFormats(source interface{}) []datadog.ConditionalFormat {
	formats, ok := source.([]interface{})
	if !ok || len(formats) == 0 {
		return nil
	}

	var r []datadog.ConditionalFormat
	for _, format := range formats {
		formatMap := format.(map[string]interface{})
		d := datadog.ConditionalFormat{}

		batchSetFromDict(batch{
			dict: formatMap,
			matches: []match{
				{"comparator", &d.Comparator},
				{"palette", &d.Palette},
				{"color", &d.Color},
				{"value", &d.Value},
				{"custom_bg_color", &d.CustomBgColor},
				{"invert", &d.Invert},
			}})

		r = append(r, d)
	}
	return r
}

func buildTileDefRequestsGroupBys(source interface{}) []datadog.TileDefApmOrLogQueryGroupBy {
	groupBys, ok := source.([]interface{})
	if !ok || len(groupBys) == 0 {
		return nil
	}

	var r []datadog.TileDefApmOrLogQueryGroupBy
	for _, groupBy := range groupBys {
		groupByMap := groupBy.(map[string]interface{})
		d := datadog.TileDefApmOrLogQueryGroupBy{}

		batchSetFromDict(batch{
			dict: groupByMap,
			matches: []match{
				{"facet", &d.Facet},
				{"limit", &d.Limit},
			}})
		if groupBySort, ok := groupByMap["sort"].([]interface{}); ok && len(groupBySort) > 0 {
			terraformSort := groupBySort[0].(map[string]interface{})
			s := datadog.TileDefApmOrLogQueryGroupBySort{
				Aggregation: datadog.String(terraformSort["aggregation"].(string)),
				Order:       datadog.String(terraformSort["order"].(string)),
			}
			if facet, ok := terraformSort["facet"].(string); ok && len(facet) > 0 {
				s.Facet = datadog.String(facet)
			}
			d.Sort = &s
		}
		r = append(r, d)
	}
	return r
}

func buildTileDefRequestsApmOrLogQuery(source interface{}) *datadog.TileDefApmOrLogQuery {
	datadogQuery := source.(map[string]interface{})

	// Index
	d := datadog.TileDefApmOrLogQuery{
		Index: datadog.String(datadogQuery["index"].(string)),
	}

	// Compute
	if compute, ok := datadogQuery["compute"].([]interface{}); ok && len(compute) > 0 {
		terraformCompute := compute[0].(map[string]interface{})
		datadogCompute := datadog.TileDefApmOrLogQueryCompute{}
		if aggr, ok := terraformCompute["aggregation"].(string); ok && len(aggr) != 0 {
			datadogCompute.SetAggregation(aggr)
		}
		if facet, ok := terraformCompute["facet"].(string); ok && len(facet) != 0 {
			datadogCompute.SetFacet(facet)
		}
		if v, ok := terraformCompute["interval"].(string); ok && len(v) != 0 {
			datadogCompute.SetInterval(v)
		}
		d.Compute = &datadogCompute
	}

	// Search
	if search, ok := datadogQuery["search"].([]interface{}); ok && len(search) > 0 {
		terraformSearch := search[0].(map[string]interface{})
		s := datadog.TileDefApmOrLogQuerySearch{
			Query: datadog.String(terraformSearch["query"].(string)),
		}
		d.Search = &s
	}

	// GroupBy
	d.GroupBy = buildTileDefRequestsGroupBys(datadogQuery["group_by"])
	return &d
}

func buildTileDefRequestsProcessQuery(source interface{}) *datadog.TileDefProcessQuery {
	datadogQuery := source.(map[string]interface{})
	d := datadog.TileDefProcessQuery{}
	batchSetFromDict(batch{
		dict: datadogQuery,
		matches: []match{
			{"metric", &d.Metric},
			{"search_by", &d.SearchBy},
			{"limit", &d.Limit},
		}})

	if v, ok := datadogQuery["filter_by"].([]interface{}); ok && len(v) != 0 {
		filters := make([]string, len(v))
		for i, filter := range v {
			filters[i] = filter.(string)
		}
		d.FilterBy = filters
	}
	return &d
}

func buildTileDefRequests(source interface{}) []datadog.TileDefRequest {
	requests, ok := source.([]interface{})
	if !ok || len(requests) == 0 {
		return nil
	}

	var r []datadog.TileDefRequest
	for _, request := range requests {
		requestMap := request.(map[string]interface{})
		metadata := map[string]datadog.TileDefMetadata{}
		getMetadataFromJSON([]byte(requestMap["metadata_json"].(string)), &metadata)
		requestMap["metadata"] = metadata
		d := datadog.TileDefRequest{}
		batchSetFromDict(batch{
			dict: requestMap,
			matches: []match{
				{"type", &d.Type},
				{"query_type", &d.QueryType},
				{"metric", &d.Metric},
				{"text_filter", &d.TextFilter},
				{"limit", &d.Limit},
				{"aggregator", &d.Aggregator},
				{"compare_to", &d.CompareTo},
				{"change_type", &d.ChangeType},
				{"order_by", &d.OrderBy},
				{"order_dir", &d.OrderDir},
				{"extra_col", &d.ExtraCol},
				{"increase_good", &d.IncreaseGood},
				{"tag_filters", &d.TagFilters},
				{"metadata", &d.Metadata},
			}})

		// request.style
		if v, ok := requestMap["style"].(map[string]interface{}); ok {
			d.Style = &datadog.TileDefRequestStyle{}
			batchSetFromDict(batch{
				dict: v,
				matches: []match{
					{"palette", &d.Style.Palette},
					{"type", &d.Style.Type},
					{"width", &d.Style.Width},
				},
			})
		}

		if v, ok := requestMap["q"].(string); ok && len(v) != 0 {
			d.Query = &v
		} else if v, ok := requestMap["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			d.LogQuery = buildTileDefRequestsApmOrLogQuery(logQuery)
		} else if v, ok := requestMap["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			d.ApmQuery = buildTileDefRequestsApmOrLogQuery(apmQuery)
		} else if v, ok := requestMap["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			d.ProcessQuery = buildTileDefRequestsProcessQuery(processQuery)
		}

		// request.conditionalFormats
		d.ConditionalFormats = buildTileDefRequestsConditionalFormats(requestMap["conditional_format"])

		r = append(r, d)
	}
	return r
}

func buildTileDefEvents(source interface{}) []datadog.TileDefEvent {
	events, ok := source.([]interface{})
	if !ok || len(events) == 0 {
		return nil
	}

	var r []datadog.TileDefEvent
	for _, event := range events {
		eventMap := event.(map[string]interface{})
		d := datadog.TileDefEvent{}

		setFromDict(eventMap, "q", &d.Query)
		r = append(r, d)
	}
	return r
}

func buildTileDefMarkers(source interface{}) []datadog.TileDefMarker {
	markers, ok := source.([]interface{})
	if !ok || len(markers) == 0 {
		return nil
	}

	var r []datadog.TileDefMarker
	for _, marker := range markers {
		markerMap := marker.(map[string]interface{})
		d := datadog.TileDefMarker{}

		batchSetFromDict(batch{
			dict: markerMap,
			matches: []match{
				{"type", &d.Type},
				{"value", &d.Value},
				{"label", &d.Label},
			}})
		r = append(r, d)
	}
	return r
}

func buildTileDefStyle(source interface{}) *datadog.TileDefStyle {
	styleMap, ok := source.(map[string]interface{})
	if !ok {
		return nil
	}

	r := &datadog.TileDefStyle{}
	batchSetFromDict(batch{
		dict: styleMap,
		matches: []match{
			{"palette", &r.Palette},
			{"palette_flip", &r.PaletteFlip},
			{"fill_min", &r.FillMin},
			{"fill_max", &r.FillMax},
		}})
	return r
}

func buildTileDef(source interface{}) *datadog.TileDef {
	tileDefs, ok := source.([]interface{})
	if !ok || len(tileDefs) == 0 {
		return nil
	}

	tileDef := tileDefs[0].(map[string]interface{})
	r := &datadog.TileDef{}

	batchSetFromDict(batch{
		dict: tileDef,
		matches: []match{
			{"viz", &r.Viz},
			{"custom_unit", &r.CustomUnit},
			{"autoscale", &r.Autoscale},
			{"precision", &r.Precision},
			{"text_align", &r.TextAlign},
			{"node_type", &r.NodeType},
			{"no_group_hosts", &r.NoGroupHosts},
			{"no_metric_hosts", &r.NoMetricHosts},
			{"scope", &r.Scope},
			{"group", &r.Group},
		}})

	r.Requests = buildTileDefRequests(tileDef["request"])
	r.Events = buildTileDefEvents(tileDef["event"])
	r.Markers = buildTileDefMarkers(tileDef["marker"])
	r.Style = buildTileDefStyle(tileDef["style"])

	return r
}

func buildWidgets(tfWidgets *[]interface{}) []datadog.Widget {
	ddWidgets := make([]datadog.Widget, len(*tfWidgets))
	for i, widget := range *tfWidgets {
		widgetMap := widget.(map[string]interface{})

		ddWidgets[i] = datadog.Widget{}
		d := &ddWidgets[i]

		if _, ok := widgetMap["title"]; ok {
			d.Title = datadog.Bool(true)
		}

		batchSetFromDict(batch{
			dict: widgetMap,
			matches: []match{
				{"title", &d.TitleText},
				{"title_align", &d.TitleAlign},
				{"title_size", &d.TitleSize},
				{"height", &d.Height},
				{"width", &d.Width},
				{"x", &d.X},
				{"y", &d.Y},
				{"type", &d.Type},
				{"text", &d.Text},
				{"text_size", &d.TextSize},
				{"text_align", &d.TextAlign},
				{"bgcolor", &d.Bgcolor},
				{"color", &d.Color},
				{"font_size", &d.FontSize},
				{"unit", &d.Unit},
				{"alert_id", &d.AlertID},
				{"auto_refresh", &d.AutoRefresh},
				{"legend", &d.Legend},
				{"query", &d.Query},
				{"legend_size", &d.LegendSize},
				{"url", &d.URL},
				{"precision", &d.Precision},
				{"viz_type", &d.VizType},
				{"check", &d.Check},
				{"group", &d.Group},
				{"grouping", &d.Grouping},
				{"tick_pos", &d.TickPos},
				{"tick_edge", &d.TickEdge},
				{"html", &d.HTML},
				{"tick", &d.Tick},
				{"event_size", &d.EventSize},
				{"sizing", &d.Sizing},
				{"margin", &d.Margin},
				{"env", &d.Env},
				{"service_service", &d.ServiceService},
				{"service_name", &d.ServiceName},
				{"size_version", &d.SizeVersion},
				{"layout_version", &d.LayoutVersion},
				{"must_show_hits", &d.MustShowHits},
				{"must_show_errors", &d.MustShowErrors},
				{"must_show_latency", &d.MustShowLatency},
				{"must_show_breakdown", &d.MustShowBreakdown},
				{"must_show_distribution", &d.MustShowDistribution},
				{"must_show_resource_list", &d.MustShowResourceList},
				{"summary_type", &d.SummaryType},
				{"display_format", &d.DisplayFormat},
				{"color_preference", &d.ColorPreference},
				{"hide_zero_counts", &d.HideZeroCounts},
				{"show_last_triggered", &d.ShowLastTriggered},
				{"manage_status_show_title", &d.ManageStatusShowTitle},
				{"manage_status_title_text", &d.ManageStatusTitleText},
				{"manage_status_title_size", &d.ManageStatusTitleSize},
				{"manage_status_title_align", &d.ManageStatusTitleAlign},
				{"columns", &d.Columns},
				{"logset", &d.Logset},
				{"timeframes", &d.Timeframes},
				{"tags", &d.Tags},
				{"group_by", &d.GroupBy},
			}})

		// widget.params
		if v, ok := widgetMap["params"].(map[string]interface{}); ok {
			d.Params = &datadog.Params{}
			batchSetFromDict(batch{
				dict: v,
				matches: []match{
					{"sort", &d.Params.Sort},
					{"text", &d.Params.Text},
					// The count and start params are deprecated for the monitor summary widget
					{"count", &d.Params.Count},
					{"start", &d.Params.Start},
				}})
		}

		// widget.rules
		if v, ok := widgetMap["rule"].([]interface{}); ok && len(v) != 0 {
			d.Rules = map[string]*datadog.Rule{}
			for i, w := range v {
				if x, ok := w.(map[string]interface{}); ok {
					rule := &datadog.Rule{}
					batchSetFromDict(batch{
						dict: x,
						matches: []match{
							{"threshold", &rule.Threshold},
							{"timeframe", &rule.Timeframe},
							{"color", &rule.Color},
						}})
					d.Rules[strconv.Itoa(i)] = rule
				}
			}
		}

		// widget.monitor
		if v, ok := widgetMap["monitor"].(map[string]interface{}); ok {
			d.Monitor = &datadog.ScreenboardMonitor{}

			if w, ok := v["id"]; ok {
				if id, err := strconv.Atoi(w.(string)); err == nil {
					d.Monitor.Id = datadog.Int(id)
				}
			}
		}

		// widget.time
		if v, ok := widgetMap["time"].(map[string]interface{}); ok {
			d.Time = &datadog.Time{}
			setFromDict(v, "live_span", &d.Time.LiveSpan)
		}

		// widget.tile_def
		d.TileDef = buildTileDef(widgetMap["tile_def"])
	}
	return ddWidgets
}

func buildScreenboard(d *schema.ResourceData) (*datadog.Screenboard, error) {
	var id int
	if d.Id() != "" {
		var err error
		id, err = strconv.Atoi(d.Id())
		if err != nil {
			return nil, err
		}
	}
	terraformWidgets := d.Get("widget").([]interface{})
	terraformTemplateVariables := d.Get("template_variable").([]interface{})

	screenboard := &datadog.Screenboard{
		Id:                datadog.Int(id),
		Title:             datadog.String(d.Get("title").(string)),
		Shared:            datadog.Bool(d.Get("shared").(bool)),
		ReadOnly:          datadog.Bool(d.Get("read_only").(bool)),
		Widgets:           buildWidgets(&terraformWidgets),
		TemplateVariables: *buildTemplateVariables(&terraformTemplateVariables),
	}

	if width, err := strconv.ParseInt(d.Get("width").(string), 10, 64); err == nil {
		screenboard.Width = datadog.Int(int(width))
	}

	if height, err := strconv.ParseInt(d.Get("height").(string), 10, 64); err == nil {
		screenboard.Height = datadog.Int(int(height))
	}

	return screenboard, nil
}

func resourceDatadogScreenboardCreate(d *schema.ResourceData, meta interface{}) error {
	screenboard, err := buildScreenboard(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	screenboard, err = client.CreateScreenboard(screenboard)
	if err != nil {
		return translateClientError(err, "error creating screenboard")
	}
	d.SetId(strconv.Itoa(screenboard.GetId()))

	return resourceDatadogScreenboardRead(d, meta)
}

// #######################################################################################
// # Convenience functions to safely pass info from the Datadog API wrapper to Terraform #
// #######################################################################################

func setStringToDict(dict map[string]interface{}, key string, field *string) {
	if field != nil {
		dict[key] = *field
	}
}

func setBoolToDict(dict map[string]interface{}, key string, field *bool) {
	if field != nil {
		dict[key] = *field
	}
}

func setIntToDict(dict map[string]interface{}, key string, field *int) {
	if field != nil {
		dict[key] = *field
	}
}

func setJSONNumberToDict(dict map[string]interface{}, key string, field *json.Number) {
	if field == nil {
		return
	}
	// for fill_min and fill_max, we do not convert to float
	if key == "fill_min" || key == "fill_max" {
		dict[key] = *field
	} else {
		v, err := (*field).Float64()
		if err != nil {
			panic(fmt.Sprintf("setJSONNumberToDict(): %v is not convertible to float", *field))
		}
		dict[key] = v
	}
}

func setPrecisionTToDict(dict map[string]interface{}, key string, field *datadog.PrecisionT) {
	if field != nil {
		dict[key] = *field
	}
}

func setStringListToDict(dict map[string]interface{}, key string, field []*string) {
	if len(field) != 0 {
		s := make([]interface{}, len(field))
		for i := range field {
			s[i] = *field[i]
		}
		dict[key] = s
	}
}

func setToDict(dict map[string]interface{}, key string, field interface{}) {
	switch field.(type) {
	case *string:
		setStringToDict(dict, key, field.(*string))
	case *bool:
		setBoolToDict(dict, key, field.(*bool))
	case *int:
		setIntToDict(dict, key, field.(*int))
	case *json.Number:
		setJSONNumberToDict(dict, key, field.(*json.Number))
	case *datadog.PrecisionT:
		setPrecisionTToDict(dict, key, field.(*datadog.PrecisionT))
	case []*string:
		setStringListToDict(dict, key, field.([]*string))
	default:
		panic(fmt.Sprintf("Cannot call setToDict() for unsupported type %T (key: %v)", field, key))
	}
}

func batchSetToDict(b batch) {
	for _, match := range b.matches {
		setToDict(b.dict, match.key, match.field)
	}
}

// ####################################################################################
// # Handle passing info from the Datadog API wrapper to Terraform                    #
// ####################################################################################

func buildTFTileDefRequestConditionalFormats(d []datadog.ConditionalFormat) []interface{} {
	l := len(d)
	if l == 0 {
		return nil
	}

	r := make([]interface{}, l)
	for i, ddConditionalFormat := range d {
		tfConditionalFormat := map[string]interface{}{}

		batchSetToDict(batch{
			dict: tfConditionalFormat,
			matches: []match{
				{"comparator", ddConditionalFormat.Comparator},
				{"palette", ddConditionalFormat.Palette},
				{"color", ddConditionalFormat.Color},
				{"value", ddConditionalFormat.Value},
				{"custom_bg_color", ddConditionalFormat.CustomBgColor},
				{"invert", ddConditionalFormat.Invert},
			}})
		r[i] = tfConditionalFormat
	}

	return r
}

func buildTFTileDefRequests(d []datadog.TileDefRequest) []interface{} {
	l := len(d)
	if l == 0 {
		return nil
	}
	r := make([]interface{}, l)
	for i, ddRequest := range d {
		tfRequest := map[string]interface{}{}
		batchSetToDict(batch{
			dict: tfRequest,
			matches: []match{
				{"type", ddRequest.Type},
				{"query_type", ddRequest.QueryType},
				{"metric", ddRequest.Metric},
				{"text_filter", ddRequest.TextFilter},
				{"limit", ddRequest.Limit},
				{"aggregator", ddRequest.Aggregator},
				{"compare_to", ddRequest.CompareTo},
				{"change_type", ddRequest.ChangeType},
				{"order_by", ddRequest.OrderBy},
				{"order_dir", ddRequest.OrderDir},
				{"extra_col", ddRequest.ExtraCol},
				{"increase_good", ddRequest.IncreaseGood},
				{"tag_filters", ddRequest.TagFilters},
			}})
		if ddRequest.Metadata != nil {
			res, _ := json.Marshal(ddRequest.Metadata)
			tfRequest["metadata_json"] = string(res)
		}

		// request.style
		if ddRequest.Style != nil {
			tfStyle := map[string]interface{}{}
			batchSetToDict(batch{
				dict: tfStyle,
				matches: []match{
					{"palette", ddRequest.Style.Palette},
					{"type", ddRequest.Style.Type},
					{"width", ddRequest.Style.Width},
				}})
			tfRequest["style"] = tfStyle
		}

		if ddRequest.Query != nil {
			tfRequest["q"] = *ddRequest.Query
		} else if ddRequest.ApmQuery != nil {
			terraformQuery := buildTFTileDefApmOrLogQuery(*ddRequest.ApmQuery)
			tfRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if ddRequest.LogQuery != nil {
			terraformQuery := buildTFTileDefApmOrLogQuery(*ddRequest.LogQuery)
			tfRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if ddRequest.ProcessQuery != nil {
			terraformQuery := buildTFTileDefProcessQuery(*ddRequest.ProcessQuery)
			tfRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}
		// request.conditionalFormats
		tfRequest["conditional_format"] = buildTFTileDefRequestConditionalFormats(ddRequest.ConditionalFormats)

		r[i] = tfRequest
	}

	return r
}

func buildTFTileDefApmOrLogQuery(datadogQuery datadog.TileDefApmOrLogQuery) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	// Index
	terraformQuery["index"] = *datadogQuery.Index
	// Compute
	terraformCompute := map[string]interface{}{}
	if datadogQuery.Compute.Aggregation != nil {
		terraformCompute["aggregation"] = *datadogQuery.Compute.Aggregation
	}
	if datadogQuery.Compute.Facet != nil {
		terraformCompute["facet"] = *datadogQuery.Compute.Facet
	}
	if datadogQuery.Compute.Interval != nil {
		terraformCompute["interval"] = *datadogQuery.Compute.Interval
	}
	terraformQuery["compute"] = []map[string]interface{}{terraformCompute}
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

func buildTFTileDefProcessQuery(datadogQuery datadog.TileDefProcessQuery) map[string]interface{} {
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

func buildTFTileDefEvents(d []datadog.TileDefEvent) []interface{} {
	l := len(d)
	if l == 0 {
		return nil
	}

	r := make([]interface{}, l)
	for i, ddEvent := range d {
		tfEvent := map[string]interface{}{}
		setToDict(tfEvent, "q", ddEvent.Query)
		r[i] = tfEvent
	}

	return r
}

func buildTFTileDefMarkers(d []datadog.TileDefMarker) []interface{} {
	l := len(d)
	if l == 0 {
		return nil
	}

	r := make([]interface{}, l)
	for i, ddMarker := range d {
		tfMarker := map[string]interface{}{}
		batchSetToDict(batch{
			dict: tfMarker,
			matches: []match{
				{"type", ddMarker.Type},
				{"value", ddMarker.Value},
				{"label", ddMarker.Label},
			}})
		r[i] = tfMarker
	}

	return r
}

func buildTFTileDefStyle(d *datadog.TileDefStyle) map[string]interface{} {
	if d == nil {
		return nil
	}

	r := map[string]interface{}{}
	batchSetToDict(batch{
		dict: r,
		matches: []match{
			{"palette", d.Palette},
			{"palette_flip", d.PaletteFlip},
			{"fill_min", d.FillMin},
			{"fill_max", d.FillMax},
		}})
	return r
}

func buildTFTileDef(d *datadog.TileDef) []interface{} {
	if d == nil {
		return nil
	}

	tfTileDef := map[string]interface{}{}

	batchSetToDict(batch{
		dict: tfTileDef,
		matches: []match{
			{"viz", d.Viz},
			{"custom_unit", d.CustomUnit},
			{"autoscale", d.Autoscale},
			{"precision", d.Precision},
			{"text_align", d.TextAlign},
			{"node_type", d.NodeType},
			{"no_group_hosts", d.NoGroupHosts},
			{"no_metric_hosts", d.NoMetricHosts},
			{"scope", d.Scope},
			{"group", d.Group},
		}})

	tfTileDef["request"] = buildTFTileDefRequests(d.Requests)
	tfTileDef["event"] = buildTFTileDefEvents(d.Events)
	tfTileDef["marker"] = buildTFTileDefMarkers(d.Markers)
	tfTileDef["style"] = buildTFTileDefStyle(d.Style)

	return []interface{}{tfTileDef}
}

func buildTFWidget(dw datadog.Widget) map[string]interface{} {
	widget := map[string]interface{}{}
	widget["type"] = dw.GetType()

	batchSetToDict(batch{
		dict: widget,
		matches: []match{
			{"title", dw.TitleText},
			{"title_align", dw.TitleAlign},
			{"title_size", dw.TitleSize},
			{"height", dw.Height},
			{"width", dw.Width},
			{"x", dw.X},
			{"y", dw.Y},
			{"type", dw.Type},
			{"text", dw.Text},
			{"text_size", dw.TextSize},
			{"text_align", dw.TextAlign},
			{"bgcolor", dw.Bgcolor},
			{"color", dw.Color},
			{"font_size", dw.FontSize},
			{"unit", dw.Unit},
			{"alert_id", dw.AlertID},
			{"auto_refresh", dw.AutoRefresh},
			{"legend", dw.Legend},
			{"query", dw.Query},
			{"legend_size", dw.LegendSize},
			{"url", dw.URL},
			{"precision", dw.Precision},
			{"viz_type", dw.VizType},
			{"check", dw.Check},
			{"group", dw.Group},
			{"grouping", dw.Grouping},
			{"tick_pos", dw.TickPos},
			{"tick_edge", dw.TickEdge},
			{"html", dw.HTML},
			{"tick", dw.Tick},
			{"event_size", dw.EventSize},
			{"sizing", dw.Sizing},
			{"margin", dw.Margin},
			{"env", dw.Env},
			{"service_service", dw.ServiceService},
			{"service_name", dw.ServiceName},
			{"size_version", dw.SizeVersion},
			{"layout_version", dw.LayoutVersion},
			{"must_show_hits", dw.MustShowHits},
			{"must_show_errors", dw.MustShowErrors},
			{"must_show_latency", dw.MustShowLatency},
			{"must_show_breakdown", dw.MustShowBreakdown},
			{"must_show_distribution", dw.MustShowDistribution},
			{"must_show_resource_list", dw.MustShowResourceList},
			{"summary_type", dw.SummaryType},
			{"display_format", dw.DisplayFormat},
			{"color_preference", dw.ColorPreference},
			{"hide_zero_counts", dw.HideZeroCounts},
			{"show_last_triggered", dw.ShowLastTriggered},
			{"manage_status_show_title", dw.ManageStatusShowTitle},
			{"manage_status_title_text", dw.ManageStatusTitleText},
			{"manage_status_title_size", dw.ManageStatusTitleSize},
			{"manage_status_title_align", dw.ManageStatusTitleAlign},
			{"columns", dw.Columns},
			{"logset", dw.Logset},
			{"timeframes", dw.Timeframes},
			{"tags", dw.Tags},
			{"group_by", dw.GroupBy},
		}})

	// widget.params
	if dw.Params != nil {
		tfParams := map[string]interface{}{}
		batchSetToDict(batch{
			dict: tfParams,
			matches: []match{
				{"sort", dw.Params.Sort},
				{"text", dw.Params.Text},
				{"count", dw.Params.Count},
				{"start", dw.Params.Start},
			}})
		widget["params"] = tfParams
	}

	// widget.rules
	if l := len(dw.Rules); l > 0 {
		tfRules := make([]interface{}, l)
		for i, ddRule := range dw.Rules {
			tfRule := map[string]interface{}{}
			batchSetToDict(batch{
				dict: tfRule,
				matches: []match{
					{"threshold", ddRule.Threshold},
					{"timeframe", ddRule.Timeframe},
					{"color", ddRule.Color},
				}})
			if index, err := strconv.Atoi(i); err == nil {
				tfRules[index] = tfRule
			}
		}
		widget["rule"] = tfRules
	}

	// widget.monitor
	if dw.Monitor != nil {
		tfMonitor := map[string]interface{}{}
		if dw.Monitor.Id != nil {
			id := strconv.Itoa(*dw.Monitor.Id)
			setToDict(tfMonitor, "id", &id)
			widget["monitor"] = tfMonitor
		}
	}

	// widget.time
	if dw.Time != nil {
		tfTime := map[string]interface{}{}
		setToDict(tfTime, "live_span", dw.Time.LiveSpan)
		widget["time"] = tfTime
	}

	widget["tile_def"] = buildTFTileDef(dw.TileDef)

	return widget
}

func resourceDatadogScreenboardRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	screenboard, err := client.GetScreenboard(id)
	if err != nil {
		return translateClientError(err, "error getting screenboard")
	}
	log.Printf("[DataDog] screenboard: %v", pretty.Sprint(screenboard))
	if err := d.Set("title", screenboard.GetTitle()); err != nil {
		return err
	}
	if v, ok := screenboard.GetHeightOk(); ok {
		if err := d.Set("height", strconv.Itoa(v)); err != nil {
			return err
		}
	}
	if v, ok := screenboard.GetWidthOk(); ok {
		if err := d.Set("width", strconv.Itoa(v)); err != nil {
			return err
		}
	}
	if v, ok := screenboard.GetSharedOk(); ok {
		if err := d.Set("shared", v); err != nil {
			return err
		}
	}
	if v, ok := screenboard.GetReadOnlyOk(); ok {
		if err := d.Set("read_only", v); err != nil {
			return err
		}
	}

	var widgets []map[string]interface{}
	for _, datadogWidget := range screenboard.Widgets {
		widgets = append(widgets, buildTFWidget(datadogWidget))
	}
	log.Printf("[DataDog] widgets: %v", pretty.Sprint(widgets))
	if err := d.Set("widget", widgets); err != nil {
		return err
	}

	var templateVariables []map[string]string
	for _, templateVariable := range screenboard.TemplateVariables {
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
	d.SetId(strconv.Itoa(screenboard.GetId()))

	return nil
}

func resourceDatadogScreenboardUpdate(d *schema.ResourceData, meta interface{}) error {
	screenboard, err := buildScreenboard(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if err = client.UpdateScreenboard(screenboard); err != nil {
		return translateClientError(err, "error updating screenboard")
	}
	return resourceDatadogScreenboardRead(d, meta)
}

func resourceDatadogScreenboardDelete(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if err = client.DeleteScreenboard(id); err != nil {
		return translateClientError(err, "error deleting screenboard")
	}
	return nil
}

func resourceDatadogScreenboardImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogScreenboardRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceDatadogScreenboardExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, err
	}
	providerConf := meta.(*ProviderConfiguration)
	client := providerConf.CommunityClient
	if _, err = client.GetScreenboard(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking screenboard exists")
	}
	return true, nil
}
