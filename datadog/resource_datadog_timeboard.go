package datadog

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/kr/pretty"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogTimeboard() *schema.Resource {
	request := &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"q": {
					Type:     schema.TypeString,
					Required: true,
				},
				"stacked": {
					Type:     schema.TypeBool,
					Optional: true,
					Default:  false,
				},
				"type": {
					Type:     schema.TypeString,
					Optional: true,
					Default:  "line",
				},
				"aggregator": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validateAggregatorMethod,
				},
				"style": {
					Type:     schema.TypeMap,
					Optional: true,
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
				},
			},
		},
	}

	marker := &schema.Schema{
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
					Type:     schema.TypeString,
					Required: true,
				},
				"request": request,
				"marker":  marker,
				"yaxis": {
					Type:     schema.TypeMap,
					Optional: true,
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
				},
				"custom_unit": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Use a custom unit (like 'users')",
				},
				"style": {
					Type:     schema.TypeMap,
					Optional: true,
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
		Create: resourceDatadogTimeboardCreate,
		Update: resourceDatadogTimeboardUpdate,
		Read:   resourceDatadogTimeboardRead,
		Delete: resourceDatadogTimeboardDelete,
		Exists: resourceDatadogTimeboardExists,
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
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"graph":             graph,
			"template_variable": templateVariable,
		},
	}
}

func appendConditionalFormats(datadogRequest *datadog.GraphDefinitionRequest, terraformFormats *[]interface{}) {
	for _, _t := range *terraformFormats {
		t := _t.(map[string]interface{})
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
	for i, _t := range *terraformTemplateVariables {
		t := _t.(map[string]interface{})
		datadogTemplateVariables[i] = datadog.TemplateVariable{
			Name:    datadog.String(t["name"].(string)),
			Prefix:  datadog.String(t["prefix"].(string)),
			Default: datadog.String(t["default"].(string)),
		}
	}
	return &datadogTemplateVariables
}

func appendRequests(datadogGraph *datadog.Graph, terraformRequests *[]interface{}) {
	for _, _t := range *terraformRequests {
		t := _t.(map[string]interface{})
		log.Printf("[DataDog] request: %v", pretty.Sprint(t))
		d := datadog.GraphDefinitionRequest{
			Query:      datadog.String(t["q"].(string)),
			Type:       datadog.String(t["type"].(string)),
			Aggregator: datadog.String(t["aggregator"].(string)),
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
			d.SetExtraCol(v.(string))
		}
		if v, ok := t["order_direction"]; ok {
			d.SetOrderDirection(v.(string))
		}

		if v, ok := t["conditional_format"]; ok {
			_v := v.([]interface{})
			appendConditionalFormats(&d, &_v)
		}

		datadogGraph.Definition.Requests = append(datadogGraph.Definition.Requests, d)
	}
}

func appendEvents(datadogGraph *datadog.Graph, terraformEvents *[]interface{}) {
	for _, _t := range *terraformEvents {
		datadogGraph.Definition.Events = append(datadogGraph.Definition.Events, datadog.GraphEvent{
			Query: datadog.String(_t.(string)),
		})
	}
}

func appendMarkers(datadogGraph *datadog.Graph, terraformMarkers *[]interface{}) {
	for _, _t := range *terraformMarkers {
		t := _t.(map[string]interface{})
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

func buildGraphs(terraformGraphs *[]interface{}) *[]datadog.Graph {
	datadogGraphs := make([]datadog.Graph, len(*terraformGraphs))
	for i, _t := range *terraformGraphs {
		t := _t.(map[string]interface{})

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
		appendRequests(d, &v)
	}
	return &datadogGraphs
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
	return &datadog.Dashboard{
		Id:                datadog.Int(id),
		Title:             datadog.String(d.Get("title").(string)),
		Description:       datadog.String(d.Get("description").(string)),
		ReadOnly:          datadog.Bool(d.Get("read_only").(bool)),
		Graphs:            *buildGraphs(&terraformGraphs),
		TemplateVariables: *buildTemplateVariables(&terraformTemplateVariables),
	}, nil
}

func resourceDatadogTimeboardCreate(d *schema.ResourceData, meta interface{}) error {
	timeboard, err := buildTimeboard(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	timeboard, err = meta.(*datadog.Client).CreateDashboard(timeboard)
	if err != nil {
		return fmt.Errorf("Failed to create timeboard using Datadog API: %s", err.Error())
	}
	d.SetId(strconv.Itoa(timeboard.GetId()))
	return nil
}

func resourceDatadogTimeboardUpdate(d *schema.ResourceData, meta interface{}) error {
	timeboard, err := buildTimeboard(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	if err = meta.(*datadog.Client).UpdateDashboard(timeboard); err != nil {
		return fmt.Errorf("Failed to update timeboard using Datadog API: %s", err.Error())
	}
	return resourceDatadogTimeboardRead(d, meta)
}

func appendTerraformGraphRequests(datadogRequests []datadog.GraphDefinitionRequest, requests *[]map[string]interface{}) {
	for _, datadogRequest := range datadogRequests {
		request := map[string]interface{}{}
		request["q"] = datadogRequest.GetQuery()
		request["stacked"] = datadogRequest.GetStacked()
		request["type"] = datadogRequest.GetType()
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
		conditionalFormats := []map[string]interface{}{}
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

		*requests = append(*requests, request)
	}
}

func buildTerraformGraph(datadogGraph datadog.Graph) map[string]interface{} {
	graph := map[string]interface{}{}
	graph["title"] = datadogGraph.GetTitle()

	definition := datadogGraph.Definition
	graph["viz"] = definition.GetViz()

	events := []string{}
	for _, e := range definition.Events {
		if v, ok := e.GetQueryOk(); ok {
			events = append(events, v)
		}
	}
	if len(events) > 0 {
		graph["events"] = events
	}

	markers := []map[string]interface{}{}
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

	requests := []map[string]interface{}{}
	appendTerraformGraphRequests(definition.Requests, &requests)
	graph["request"] = requests

	return graph
}

func resourceDatadogTimeboardRead(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	timeboard, err := meta.(*datadog.Client).GetDashboard(id)
	if err != nil {
		return err
	}
	log.Printf("[DataDog] timeboard: %v", pretty.Sprint(timeboard))
	if err := d.Set("title", timeboard.GetTitle()); err != nil {
		return err
	}
	if err := d.Set("description", timeboard.GetDescription()); err != nil {
		return err
	}

	graphs := []map[string]interface{}{}
	for _, datadogGraph := range timeboard.Graphs {
		graphs = append(graphs, buildTerraformGraph(datadogGraph))
	}
	log.Printf("[DataDog] graphs: %v", pretty.Sprint(graphs))
	if err := d.Set("graph", graphs); err != nil {
		return err
	}

	templateVariables := []map[string]string{}
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

	return nil
}

func resourceDatadogTimeboardDelete(d *schema.ResourceData, meta interface{}) error {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return err
	}
	if err = meta.(*datadog.Client).DeleteDashboard(id); err != nil {
		return err
	}
	return nil
}

func resourceDatadogTimeboardImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogTimeboardRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceDatadogTimeboardExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return false, err
	}
	if _, err = meta.(*datadog.Client).GetDashboard(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func validateAggregatorMethod(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	validMethods := map[string]struct{}{
		"avg":  {},
		"max":  {},
		"min":  {},
		"sum":  {},
		"last": {},
	}
	if _, ok := validMethods[value]; !ok {
		errors = append(errors, fmt.Errorf(
			`%q contains an invalid method %q. Valid methods are either "avg", "max", "min", "sum", or "last"`, k, value))
	}
	return
}
