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

func resourceDatadogScreenboard() *schema.Resource {

	tileDefEvent := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"q": {
					Description: "The search query for event overlays.",
					Type:        schema.TypeString,
					Required:    true,
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
					Description: "A label for the line or range.",
					Type:        schema.TypeString,
					Optional:    true,
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
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "Boolean indicating whether to invert color scheme.",
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
					Description: "Name of the index to query",
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
								Description: "The aggregation method.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"facet": {
								Description: "Facet name.",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"interval": {
								Description: "Define a time interval in seconds.",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
				"search": {
					Description: "One nested block is allowed with the structure below.",
					Type:        schema.TypeList,
					Optional:    true,
					MaxItems:    1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Description: "Query to use.",
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
								Description: "Facet name.",
								Type:        schema.TypeString,
								Required:    true,
							},
							"limit": {
								Description: "Maximum number of items in the group.",
								Type:        schema.TypeInt,
								Optional:    true,
							},
							"sort": {
								Description: "One map is allowed with the keys as below.",
								Type:        schema.TypeList,
								Optional:    true,
								MaxItems:    1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"aggregation": {
											Description: "The aggregation method.",
											Type:        schema.TypeString,
											Required:    true,
										},
										"order": {
											Description: "Widget sorting methods.",
											Type:        schema.TypeString,
											Required:    true,
										},
										"facet": {
											Description: "Facet name.",
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

	tileDefProcessQuery := &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"metric": {
					Description: "Your chosen metric.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"search_by": {
					Description: "Your chosen search term.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"filter_by": {
					Description: "List of processes.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"limit": {
					Description: "Max number of items in the filter list.",
					Type:        schema.TypeInt,
					Optional:    true,
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
					Description: "Only for widgets of type `timeseries`, `query_value`, `toplist`, `change`, `hostmap`: The query of the request. Pro tip: Use the JSON tab inside the Datadog UI to help build you query strings.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"log_query":     tileDefApmOrLogQuery,
				"apm_query":     tileDefApmOrLogQuery,
				"process_query": tileDefProcessQuery,
				"type": {
					Description: "Only for widgets of type `timeseries`, `query_value`, `hostmap`: Choose the type of representation to use for this query. For widgets of type `timeseries` and `query_value`, use one of `line`, `bars` or `area`. For widgets of type `hostmap`, use `fill` or `size`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"query_type": {
					Description: "Only for widgets of type `process`. Use `process`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"metric": {
					Description: "Only for widgets of type `process`. The metric you want to use for the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"text_filter": {
					Description: "Only for widgets of type `process`. The search query for the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tag_filters": {
					Description: "Only for widgets of type `process`. Tags to use for filtering.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"limit": {
					Description: "Only for widgets of type `process`. Integer indicating the number of hosts to limit to.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"style": {
					Description: "Only for widgets of type `timeseries`, `query_value`, `toplist`, `process`. How to display the widget. The structure of this block is described below. At most one such block should be present in a given `request` block.",
					Type:        schema.TypeMap,
					Optional:    true,
				},
				"conditional_format": tileDefRequestConditionalFormat,
				"aggregator": {
					Description:  "Only for widgets of type `query_value`. The aggregator to use for time aggregation. One of `avg`, `min`, `max`, `sum`, `last`.",
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validators.ValidateAggregatorMethod,
				},
				"compare_to": {
					Description: "Only for widgets of type `change`. Choose from when to compare current data to. One of `hour_before`, `day_before`, `week_before` or `month_before`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"change_type": {
					Description: "Only for widgets of type `change`. Whether to show absolute or relative change. One of `absolute`, `relative`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"order_by": {
					Description: "Only for widgets of type `change`. One of `change`, `name`, `present` (present value) or `past` (past value).",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"order_dir": {
					Description: "Only for widgets of type `change`. Either `asc` (ascending) or `desc` (descending).",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"extra_col": {
					Description: "Only for widgets of type `change`. If set to `present`, displays current value. Can be left empty otherwise.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"increase_good": {
					Description: "Only for widgets of type `change`. Boolean indicating whether an increase in the value is good (thus displayed in green) or not (thus displayed in red).",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"metadata_json": {
					Description:  "A JSON blob (preferrably created using [jsonencode](https://www.terraform.io/docs/configuration/functions/jsonencode.html?_ga=2.6381362.1091155358.1609189257-888022054.1605547463)) representing mapping of query expressions to alias names. Note that the query expressions in `metadata_json` will be ignored if they're not present in the query.",
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
					Description: "Should be the same as the widget's type. One of `timeseries`, `query_value`, `hostmap`, `change`, `toplist`, `process`.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"custom_unit": {
					Description: "Only for widgets of type `query_value`. The unit for the value displayed in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"autoscale": {
					Description: "Only for widgets of type `query_value`. Boolean indicating whether to automatically scale the tile.",
					Type:        schema.TypeBool,
					Optional:    true,
					Default:     true,
				},
				"precision": {
					Description: "Only for widgets of type `query_value`. The precision to use when displaying the tile.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"text_align": {
					Description: "Only for widgets of type `query_value`. The alignment of the text.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"node_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Only for widgets of type `hostmap`. The type of node used. Either `host` or `container`.",
				},
				"scope": {
					Description: "Only for widgets of type `hostmap`. The list of tags to filter nodes by.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"group": {
					Description: "Only for widgets of type `hostmap`. The list of tags to group nodes by.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"no_group_hosts": {
					Description: "Only for widgets of type `hostmap`. Boolean indicating whether to show ungrouped nodes.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"no_metric_hosts": {
					Description: "Only for widgets of type `hostmap`. Boolean indicating whether to show nodes with no metrics.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"style": {
					Description: "Only for widgets of type `hostmap`. Nested block describing how to display the widget. The structure of this block is described below. At most one such block should be present in a given `tile_def` block.",
					Type:        schema.TypeMap,
					Optional:    true,
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
					Description: "The type of the widget. One of [ 'free_text', 'timeseries', 'query_value', 'toplist', 'change', 'event_timeline', 'event_stream', 'image', 'note', 'alert_graph', 'alert_value', 'iframe', 'check_status', 'trace_service', 'hostmap', 'manage_status', 'log_stream', 'process']",
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
					Type:        schema.TypeString,
					Optional:    true,
					Default:     "auto",
					Description: "Only for widgets of type `alert_value`. The size of the text in the widget.",
				},
				"text_align": {
					Description: "Only for widgets of type `free_text`, `alert_value`, `note`. The alignment of the text in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"color": {
					Description: "Only for widgets of type `free_text`. The color of the text in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"bgcolor": {
					Description: "Only for widgets of type `note`. The color of the background of the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"font_size": {
					Description: "Only for widgets of type `free_text`, `note`. The size of the text in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"unit": {
					Description: "Only for widgets of type `alert_value`. The unit for the value displayed in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"alert_id": {
					Description: "Only for widgets of type `alert_value`, `alert_graph`. The ID of the monitor used by the widget.",
					Type:        schema.TypeInt,
					Optional:    true,
				},
				"auto_refresh": {
					Description: "Only for widgets of type `alert_value`, `alert_graph`. Boolean indicating whether the widget is refreshed automatically.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"legend": {
					Description: "Only for widgets of type `timeseries`, `query_value`, `toplist`. Boolean indicating whether to display a legend in the widget.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"query": {
					Description: "Only for widgets of type `event_timeline`, `event_stream`, `hostmap`, `log_stream`. The query to use in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"legend_size": {
					Description: "Only for widgets of type `timeseries`, `query_value`, `toplist`. The size of the legend displayed in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"url": {
					Description: "Only for widgets of type `image`, `iframe`. The URL to use as a data source for the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"precision": {
					Description: "Only for widgets of type `query_value`. The precision to use when displaying the tile.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tags": {
					Description: "Only for widgets of type `check_status`. List of tags to use in the widget.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"viz_type": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['timeseries', 'toplist']",
				},
				"check": {
					Description: "Only for widgets of type `check_status`. The check to use in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"group": {
					Description: "Only for widgets of type `check_status`. The check group to use in the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"grouping": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "One of: ['cluster', 'check']",
				},
				"group_by": {
					Description: "Only for widgets of type `check_status`. When `grouping = \"cluster\"`, indicates a list of tags to use for grouping.",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"tick_pos": {
					Description: "Only for widgets of type `note`. When `tick = true`, string with a percent sign indicating the position of the tick. Example: use `tick_pos = \"50%\"` for centered alignment.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tick_edge": {
					Description: "Only for widgets of type `note`. When `tick = true`, string indicating on which side of the widget the tick should be displayed. One of `bottom`, `top`, `left`, `right`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"html": {
					Description: "Only for widgets of type `note`. The content of the widget. HTML tags supported.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"tick": {
					Description: "Only for widgets of type `note`. Boolean indicating whether a tick should be displayed on the border of the widget.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"event_size": {
					Description: "Only for widgets of type `event_stream`. The size of the events in the widget. Either `s` (small, title only) or `l` (large, full event).",
					Type:        schema.TypeString,
					Optional:    true,
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
					Description: "Only for widgets of type `timeseries`, `toplist`, `event_timeline`, `event_stream`, `alert_graph`, `check_status`, `trace_service`, `log_stream`. Nested block describing the timeframe to use when displaying the widget. The structure of this block is described below. At most one such block should be present in a given widget.",
					Type:        schema.TypeMap,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"env": {
					Description: "Only for widgets of type `trace_service`. The environment to use.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"service_service": {
					Description: "Only for widgets of type `trace_service`. The trace service to use.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"service_name": {
					Description: "Only for widgets of type `trace_service`. The name of the service to use.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"size_version": {
					Description: "Only for widgets of type `trace_service`. The size of the widget. One of `small`, `medium`, `large`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"layout_version": {
					Description: "Only for widgets of type `trace_service`. The number of columns to use when displaying values. One of `one_column`, `two_column`, `three_column`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"must_show_hits": {
					Description: "Only for widgets of type `trace_service`. Boolean indicating whether to display hits.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"must_show_errors": {
					Description: "Only for widgets of type `trace_service`. Boolean indicating whether to display errors.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"must_show_latency": {
					Description: "Only for widgets of type `trace_service`. Boolean indicating whether to display latency.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"must_show_breakdown": {
					Description: "Only for widgets of type `trace_service`. Boolean indicating whether to display breakdown.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"must_show_distribution": {
					Description: "Only for widgets of type `trace_service`. Boolean indicating whether to display distribution.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"must_show_resource_list": {
					Description: "Only for widgets of type `trace_service` Boolean indicating whether to display resources.",
					Type:        schema.TypeBool,
					Optional:    true,
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
					Description: "Only for widgets of type `manage_status`. Boolean indicating whether to hide empty categories.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"show_last_triggered": {
					Description: "Only for widgets of type `manage_status`. Boolean indicating whether to show when monitors/groups last triggered.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"params": {
					Description: "Only for widgets of type `manage_status`. Nested block describing the monitors to display. The structure of this block is described below. At most one such block should be present in a given widget.",
					Type:        schema.TypeMap,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"manage_status_show_title": {
					Description: "Only for widgets of type `manage_status`. Boolean indicating whether to show a title.",
					Type:        schema.TypeBool,
					Optional:    true,
				},
				"manage_status_title_text": {
					Description: "Only for widgets of type `manage_status`. The title of the widget.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"manage_status_title_size": {
					Description: "Only for widgets of type `manage_status`. The size of the widget's title.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"manage_status_title_align": {
					Description: "Only for widgets of type `manage_status`. The alignment of the widget's title. One of `left`, `center`, or `right`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"columns": {
					Description: "Only for widgets of type `log_stream`. Stringified list of columns to use. Example: `[\"column1\",\"column2\",\"column3\"]`.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"logset": {
					Description: "Only for widgets of type `log_stream`. ID of the logset to use.",
					Type:        schema.TypeString,
					Optional:    true,
				},
				"timeframes": {
					Description: "",
					Type:        schema.TypeList,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"rule": {
					Description: "",
					Type:        schema.TypeList,
					Optional:    true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"threshold": {
								Description: "",
								Type:        schema.TypeFloat,
								Optional:    true,
							},
							"timeframe": {
								Description: "",
								Type:        schema.TypeString,
								Optional:    true,
							},
							"color": {
								Description: "",
								Type:        schema.TypeString,
								Optional:    true,
							},
						},
					},
				},
				"monitor": {
					Description: "",
					Type:        schema.TypeMap,
					Optional:    true,
					Elem:        &schema.Schema{Type: schema.TypeString},
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
		Description:        "Provides a Datadog screenboard resource. This can be used to create and manage Datadog screenboards.",
		DeprecationMessage: "This resource is deprecated. Instead use the Dashboard resource",
		Create:             resourceDatadogScreenboardCreate,
		Read:               resourceDatadogScreenboardRead,
		Update:             resourceDatadogScreenboardUpdate,
		Delete:             resourceDatadogScreenboardDelete,
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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "The read-only status of the screenboard. Default is `false`.",
			},
			"widget":            widget,
			"template_variable": templateVariable,
		},
	}
}

// #######################################################################################
// # Convenience functions to safely pass info from Terraform to the Datadog API wrapper #
// # DEPRECATED - All utils methods are moved to the internal package
// #######################################################################################

func validateMetadataJSON(v interface{}, k string) (ws []string, errors []error) {
	err := utils.GetMetadataFromJSON([]byte(v.(string)), &map[string]datadog.TileDefMetadata{})
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
		utils.GetMetadataFromJSON([]byte(requestMap["metadata_json"].(string)), &metadata)
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
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error creating screenboard")
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
		if strings.Contains(err.Error(), "404 Not Found") {
			d.SetId("")
			return nil
		}
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error getting screenboard")
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
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error updating screenboard")
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
		return utils.TranslateClientError(err, providerConf.CommunityClient.GetBaseUrl(),  "error deleting screenboard")
	}
	return nil
}

func resourceDatadogScreenboardImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogScreenboardRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
