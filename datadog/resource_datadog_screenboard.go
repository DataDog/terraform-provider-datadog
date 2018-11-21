package datadog

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/kr/pretty"
	datadog "github.com/zorkian/go-datadog-api"
)

func resourceDatadogScreenboard() *schema.Resource {

	tileDefEvent := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"q": &schema.Schema{
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
				"type": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"value": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"label": &schema.Schema{
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
				"palette": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The palette to use if this condition is met.",
				},
				"comparator": &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "Comparator (<, >, etc)",
				},
				"color": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Custom color (e.g., #205081)",
				},
				"value": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Value that is threshold for conditional format",
				},
				"invert": &schema.Schema{
					Type:     schema.TypeBool,
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
				"q": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"type": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"query_type": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"metric": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"text_filter": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"tag_filters": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"limit": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"style": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
				},
				"conditional_format": tileDefRequestConditionalFormat,
				"aggregator": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"compare_to": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"change_type": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"order_by": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"order_dir": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"extra_col": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"increase_good": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
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
				"viz": &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
				"custom_unit": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"autoscale": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
					Default:  true,
				},
				"precision": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"text_align": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"node_type": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['host', 'container']",
				},
				"scope": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"group": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"no_group_hosts": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"no_metric_hosts": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"style": &schema.Schema{
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
				"type": &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The type of the widget. One of [ 'free_text', 'timeseries', 'query_value', 'toplist', 'change', 'event_timeline', 'event_stream', 'image', 'note', 'alert_graph', 'alert_value', 'iframe', 'check_status', 'trace_service', 'hostmap', 'manage_status', 'log_stream', 'uptime', 'process']",
				},
				"title": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The name of the widget.",
				},
				"title_align": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "left",
					Description: "The alignment of the widget's title.",
				},
				"title_size": &schema.Schema{
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     16,
					Description: "The size of the widget's title.",
				},
				"height": &schema.Schema{
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     15,
					Description: "The height of the widget.",
				},
				"width": &schema.Schema{
					Type:        schema.TypeInt,
					Optional:    true,
					Default:     50,
					Description: "The width of the widget.",
				},
				"x": &schema.Schema{
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The position of the widget on the x axis.",
				},
				"y": &schema.Schema{
					Type:        schema.TypeInt,
					Required:    true,
					Description: "The position of the widget on the y axis.",
				},
				"text": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "For widgets of type 'free_text', the text to use.",
				},
				"text_size": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					Default:  "auto",
				},
				"text_align": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"color": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"bgcolor": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"font_size": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"unit": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"alert_id": &schema.Schema{
					Type:     schema.TypeInt,
					Optional: true,
				},
				"auto_refresh": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"legend": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"query": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"legend_size": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"url": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"precision": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"tags": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"viz_type": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['timeseries', 'toplist']",
				},
				"check": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"group": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"grouping": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['cluster', 'check']",
				},
				"group_by": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"tick_pos": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"tick_edge": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"html": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"tick": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"event_size": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"sizing": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['center', 'zoom', 'fit']",
				},
				"margin": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['small', 'large']",
				},
				"tile_def": tileDef,
				"time": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"env": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"service_service": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"service_name": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"size_version": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"layout_version": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"must_show_hits": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_errors": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_latency": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_breakdown": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_distribution": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"must_show_resource_list": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"display_format": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['counts', 'list', 'countsAndList']",
				},
				"color_preference": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['background', 'text']",
				},
				"hide_zero_counts": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"params": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"manage_status_show_title": &schema.Schema{
					Type:     schema.TypeBool,
					Optional: true,
				},
				"manage_status_title_text": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"manage_status_title_size": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"manage_status_title_align": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"columns": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"logset": &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
				},
				"timeframes": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				"rule": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"threshold": &schema.Schema{
								Type:     schema.TypeFloat,
								Optional: true,
							},
							"timeframe": &schema.Schema{
								Type:     schema.TypeString,
								Optional: true,
							},
							"color": &schema.Schema{
								Type:     schema.TypeString,
								Optional: true,
							},
						},
					},
				},
				"monitor": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeInt},
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
				"name": &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The name of the variable.",
				},
				"prefix": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The tag prefix associated with the variable. Only tags with this prefix will appear in the variable dropdown.",
				},
				"default": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Description: "The default value for the template variable on dashboard load.",
				},
			},
		},
	}

	return &schema.Resource{
		Create: resourceDatadogScreenboardCreate,
		Read:   resourceDatadogScreenboardRead,
		Update: resourceDatadogScreenboardUpdate,
		Delete: resourceDatadogScreenboardDelete,
		Exists: resourceDatadogScreenboardExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogScreenboardImport,
		},

		Schema: map[string]*schema.Schema{
			"title": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the screenboard",
			},
			"height": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Height of the screenboard",
			},
			"width": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Width of the screenboard",
			},
			"shared": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether the screenboard is shared or not",
			},
			"read_only": &schema.Schema{
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
		f := json.Number(strconv.FormatFloat(v.(float64), 'e', -1, 64))
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

	r := []datadog.ConditionalFormat{}
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
				{"invert", &d.Invert},
			}})

		r = append(r, d)
	}
	return r
}

func buildTileDefRequests(source interface{}) []datadog.TileDefRequest {
	requests, ok := source.([]interface{})
	if !ok || len(requests) == 0 {
		return nil
	}

	r := []datadog.TileDefRequest{}
	for _, request := range requests {
		requestMap := request.(map[string]interface{})
		d := datadog.TileDefRequest{}
		batchSetFromDict(batch{
			dict: requestMap,
			matches: []match{
				{"q", &d.Query},
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

	r := []datadog.TileDefEvent{}
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

	r := []datadog.TileDefMarker{}
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
				{"display_format", &d.DisplayFormat},
				{"color_preference", &d.ColorPreference},
				{"hide_zero_counts", &d.HideZeroCounts},
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
	return &datadog.Screenboard{
		Id:                datadog.Int(id),
		Title:             datadog.String(d.Get("title").(string)),
		Height:            datadog.Int(d.Get("height").(int)),
		Width:             datadog.Int(d.Get("width").(int)),
		Shared:            datadog.Bool(d.Get("shared").(bool)),
		ReadOnly:          datadog.Bool(d.Get("read_only").(bool)),
		Widgets:           buildWidgets(&terraformWidgets),
		TemplateVariables: *buildTemplateVariables(&terraformTemplateVariables),
	}, nil
}

func resourceDatadogScreenboardCreate(d *schema.ResourceData, meta interface{}) error {
	screenboard, err := buildScreenboard(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	screenboard, err = meta.(*datadog.Client).CreateScreenboard(screenboard)
	if err != nil {
		return fmt.Errorf("Failed to create screenboard using Datadog API: %s", err.Error())
	}
	d.SetId(strconv.Itoa(screenboard.GetId()))
	return nil
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
	if field != nil {
		v, err := (*field).Float64()
		if err != nil {
			panic(fmt.Sprintf("setJSONNumberToDict(): %v is not convertible to float", *field))
		}
		dict[key] = v
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
				{"q", ddRequest.Query},
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

		// request.conditionalFormats
		tfRequest["conditional_format"] = buildTFTileDefRequestConditionalFormats(ddRequest.ConditionalFormats)

		r[i] = tfRequest
	}

	return r
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
			{"display_format", dw.DisplayFormat},
			{"color_preference", dw.ColorPreference},
			{"hide_zero_counts", dw.HideZeroCounts},
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
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	screenboard, err := meta.(*datadog.Client).GetScreenboard(id)
	if err != nil {
		return err
	}
	log.Printf("[DataDog] screenboard: %v", pretty.Sprint(screenboard))
	if err := d.Set("title", screenboard.GetTitle()); err != nil {
		return err
	}
	if err := d.Set("height", screenboard.GetHeight()); err != nil {
		return err
	}
	if err := d.Set("width", screenboard.GetWidth()); err != nil {
		return err
	}
	if err := d.Set("shared", screenboard.GetShared()); err != nil {
		return err
	}
	if err := d.Set("read_only", screenboard.GetReadOnly()); err != nil {
		return err
	}

	widgets := []map[string]interface{}{}
	for _, datadogWidget := range screenboard.Widgets {
		widgets = append(widgets, buildTFWidget(datadogWidget))
	}
	log.Printf("[DataDog] widgets: %v", pretty.Sprint(widgets))
	if err := d.Set("widget", widgets); err != nil {
		return err
	}

	templateVariables := []map[string]string{}
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

	return nil
}

func resourceDatadogScreenboardUpdate(d *schema.ResourceData, meta interface{}) error {
	screenboard, err := buildScreenboard(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	if err = meta.(*datadog.Client).UpdateScreenboard(screenboard); err != nil {
		return fmt.Errorf("Failed to update screenboard using Datadog API: %s", err.Error())
	}
	return resourceDatadogScreenboardRead(d, meta)
}

func resourceDatadogScreenboardDelete(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	if err = meta.(*datadog.Client).DeleteScreenboard(id); err != nil {
		return err
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
	if _, err = meta.(*datadog.Client).GetScreenboard(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
