package datadog

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/zorkian/go-datadog-api"
)

func resourceDatadogDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatadogDashboardCreate,
		Update: resourceDatadogDashboardUpdate,
		Read:   resourceDatadogDashboardRead,
		Delete: resourceDatadogDashboardDelete,
		Exists: resourceDatadogDashboardExists,
		Importer: &schema.ResourceImporter{
			State: resourceDatadogDashboardImport,
		},
		Schema: map[string]*schema.Schema{
			"title": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The title of the dashboard.",
			},
			"widget": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "The list of widgets to display on the dashboard.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"definition": {
							Type:          schema.TypeString,
							Optional:      true,
							Description:   "The definition for non-Group widget.",
							ConflictsWith: []string{"widget.group_definition"},
						},
						"group_definition": {
							Type:          schema.TypeList,
							Optional:      true,
							MaxItems:      1,
							Description:   "The definition for a Group widget.",
							ConflictsWith: []string{"widget.definition"},
							Elem: &schema.Resource{
								Schema: getGroupDefinitionSchema(),
							},
						},
						"layout": {
							Type:        schema.TypeMap,
							Optional:    true,
							Description: "The layout of the widget on a 'free' dashboard.",
							Elem: &schema.Resource{
								Schema: getWidgetLayoutSchema(),
							},
						},
					},
				},
			},
			"layout_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The layout type of the dashboard, either 'free' or 'ordered'.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the dashboard.",
			},
			"is_read_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether this dashboard is read-only.",
			},
			"template_variable": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of template variables for this dashboard.",
				Elem: &schema.Resource{
					Schema: getTemplateVariableSchema(),
				},
			},
			"notify_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of handles of users to notify when changes are made to this dashboard.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func buildDatadogDashboard(d *schema.ResourceData) (*datadog.Board, error) {
	var dashboard datadog.Board

	dashboard.SetId(d.Id())

	if v, ok := d.GetOk("title"); ok {
		dashboard.SetTitle(v.(string))
	}
	if v, ok := d.GetOk("layout_type"); ok {
		dashboard.SetLayoutType(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		dashboard.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("is_read_only"); ok {
		dashboard.SetIsReadOnly(v.(bool))
	}

	// Build Widgets
	terraformWidgets := d.Get("widget").([]interface{})
	datadogWidgets, err := buildDatadogWidgets(&terraformWidgets)
	if err != nil {
		return nil, err
	}
	dashboard.Widgets = *datadogWidgets

	// Build NotifyList
	notifyList := d.Get("notify_list").([]interface{})
	dashboard.NotifyList = *buildDatadogNotifyList(&notifyList)

	// Build TemplateVariables
	templateVariables := d.Get("template_variable").([]interface{})
	dashboard.TemplateVariables = *buildDatadogTemplateVariables(&templateVariables)

	return &dashboard, nil
}

func resourceDatadogDashboardCreate(d *schema.ResourceData, meta interface{}) error {
	dashboard, err := buildDatadogDashboard(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	dashboard, err = meta.(*datadog.Client).CreateBoard(dashboard)
	if err != nil {
		return fmt.Errorf("Failed to create dashboard using Datadog API: %s", err.Error())
	}
	d.SetId(*dashboard.Id)
	return nil
}

func resourceDatadogDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	dashboard, err := buildDatadogDashboard(d)
	if err != nil {
		return fmt.Errorf("Failed to parse resource configuration: %s", err.Error())
	}
	if err = meta.(*datadog.Client).UpdateBoard(dashboard); err != nil {
		return fmt.Errorf("Failed to update dashboard using Datadog API: %s", err.Error())
	}
	return resourceDatadogDashboardRead(d, meta)
}

func resourceDatadogDashboardRead(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	dashboard, err := meta.(*datadog.Client).GetBoard(id)
	if err != nil {
		return err
	}

	if err := d.Set("title", dashboard.GetTitle()); err != nil {
		return err
	}
	if err := d.Set("layout_type", dashboard.GetLayoutType()); err != nil {
		return err
	}
	if err := d.Set("description", dashboard.GetDescription()); err != nil {
		return err
	}
	if err := d.Set("is_read_only", dashboard.GetIsReadOnly()); err != nil {
		return err
	}

	// Set widgets
	currentTerraformWidgets := d.Get("widget").([]interface{})
	terraformWidgets, err := buildTerraformWidgets(&dashboard.Widgets, currentTerraformWidgets)
	if err != nil {
		return err
	}
	if err := d.Set("widget", terraformWidgets); err != nil {
		return err
	}

	// Set template variables
	templateVariables := buildTerraformTemplateVariables(&dashboard.TemplateVariables)
	if err := d.Set("template_variable", templateVariables); err != nil {
		return err
	}

	// Set notify list
	notifyList := buildTerraformNotifyList(&dashboard.NotifyList)
	if err := d.Set("notify_list", notifyList); err != nil {
		return err
	}

	return nil
}

func resourceDatadogDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	id := d.Id()
	if err := meta.(*datadog.Client).DeleteBoard(id); err != nil {
		return err
	}
	return nil
}

func resourceDatadogDashboardImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceDatadogDashboardRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceDatadogDashboardExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	id := d.Id()
	if _, err := meta.(*datadog.Client).GetBoard(id); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

//
// Template Variable helpers
//

func getTemplateVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
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
	}
}

func buildDatadogTemplateVariables(terraformTemplateVariables *[]interface{}) *[]datadog.TemplateVariable {
	datadogTemplateVariables := make([]datadog.TemplateVariable, len(*terraformTemplateVariables))
	for i, _terraformTemplateVariable := range *terraformTemplateVariables {
		terraformTemplateVariable := _terraformTemplateVariable.(map[string]interface{})
		var datadogTemplateVariable datadog.TemplateVariable
		if v, ok := terraformTemplateVariable["name"].(string); ok && len(v) != 0 {
			datadogTemplateVariable.SetName(v)
		}
		if v, ok := terraformTemplateVariable["prefix"].(string); ok && len(v) != 0 {
			datadogTemplateVariable.SetPrefix(v)
		}
		if v, ok := terraformTemplateVariable["default"].(string); ok && len(v) != 0 {
			datadogTemplateVariable.SetDefault(v)
		}
		datadogTemplateVariables[i] = datadogTemplateVariable
	}
	return &datadogTemplateVariables
}

func buildTerraformTemplateVariables(datadogTemplateVariables *[]datadog.TemplateVariable) *[]map[string]string {
	terraformTemplateVariables := make([]map[string]string, len(*datadogTemplateVariables))
	for i, templateVariable := range *datadogTemplateVariables {
		terraformTemplateVariable := map[string]string{}
		if v, ok := templateVariable.GetNameOk(); ok {
			terraformTemplateVariable["name"] = v
		}
		if v, ok := templateVariable.GetPrefixOk(); ok {
			terraformTemplateVariable["prefix"] = v
		}
		if v, ok := templateVariable.GetDefaultOk(); ok {
			terraformTemplateVariable["default"] = v
		}
		terraformTemplateVariables[i] = terraformTemplateVariable
	}
	return &terraformTemplateVariables
}

//
// Notify List helpers
//

func buildDatadogNotifyList(terraformNotifyList *[]interface{}) *[]string {
	datadogNotifyList := make([]string, len(*terraformNotifyList))
	for i, authorHandle := range *terraformNotifyList {
		datadogNotifyList[i] = authorHandle.(string)
	}
	return &datadogNotifyList
}

func buildTerraformNotifyList(datadogNotifyList *[]string) *[]string {
	terraformNotifyList := make([]string, len(*datadogNotifyList))
	for i, authorHandle := range *datadogNotifyList {
		terraformNotifyList[i] = authorHandle
	}
	return &terraformNotifyList
}

//
// Widget helpers
//

func buildDatadogWidgets(terraformWidgets *[]interface{}) (*[]datadog.BoardWidget, error) {
	datadogWidgets := make([]datadog.BoardWidget, len(*terraformWidgets))
	for i, _terraformWidget := range *terraformWidgets {
		terraformWidget := _terraformWidget.(map[string]interface{})
		var datadogWidget datadog.BoardWidget

		// Build widget definition or group definition
		if v, ok := terraformWidget["definition"].(string); ok && len(v) != 0 {
			datadogDefinition, err := unmarshalWidgetDefinition([]byte(v))
			if err != nil {
				return nil, err
			}
			datadogWidget.Definition = datadogDefinition
		} else if _v, ok := terraformWidget["group_definition"].([]interface{}); ok && len(_v) != 0 {
			if v, ok := _v[0].(map[string]interface{}); ok && len(v) != 0 {
				datadogGroupDefinition, err := buildDatadogGroupDefinition(v)
				if err != nil {
					return nil, err
				}
				datadogWidget.Definition = datadogGroupDefinition
			}
		}

		// Build widget layout
		if v, ok := terraformWidget["layout"].(map[string]interface{}); ok && len(v) != 0 {
			datadogWidget.SetLayout(buildDatadogWidgetLayout(v))
		}

		datadogWidgets[i] = datadogWidget
	}
	return &datadogWidgets, nil
}

func buildTerraformWidgets(datadogWidgets *[]datadog.BoardWidget, currentTerraformWidgets []interface{}) (*[]map[string]interface{}, error) {
	terraformWidgets := make([]map[string]interface{}, len(*datadogWidgets))
	for i, datadogWidget := range *datadogWidgets {
		terraformWidget := map[string]interface{}{}

		// Get the widget type
		widgetType, err := datadogWidget.GetWidgetType()
		if err != nil {
			return nil, err
		}
		// If this is a Group widget, build attribute "group_definition", otherwise build attribute "definition"
		if widgetType == datadog.GROUP_WIDGET {
			currentGroupWidgets := []interface{}{}
			if i < len(currentTerraformWidgets) {
				terraformDefinition := currentTerraformWidgets[i].(map[string]interface{})["group_definition"].([]interface{})
				currentGroupWidgets = terraformDefinition[0].(map[string]interface{})["widget"].([]interface{})
			}
			terraformGroupDefinition, err := buildTerraformGroupDefinition(datadogWidget.Definition.(datadog.GroupDefinition), currentGroupWidgets)
			if err != nil {
				return nil, err
			}
			terraformWidget["group_definition"] = []map[string]interface{}{terraformGroupDefinition}
		} else {
			var terraformDefinition string
			if i < len(currentTerraformWidgets) {
				// If we have an existing widget in Terraform at that index, get its definition as JSON,
				// unmarshal it to get a Go struct representing the definition and compare this struct to the one
				// coming from the Datadog library.
				// If the structs are equal, then keep the Terraform definition, otherwise marshal the Datadong definition.
				// This is to unsure we don't ask for an update when the 2 definitions represent the same object but have
				// been serialized into different strings in the state (because of attributes order, whitespaces, ...).
				terraformDefinition = currentTerraformWidgets[i].(map[string]interface{})["definition"].(string)
				unmarshalledDefinition, err := unmarshalWidgetDefinition([]byte(terraformDefinition))
				if err != nil {
					return nil, err
				}
				if reflect.DeepEqual(unmarshalledDefinition, datadogWidget.Definition) == false {
					marshalledDatadogDefinition, _ := json.Marshal(datadogWidget.Definition)
					terraformDefinition = string(marshalledDatadogDefinition)
				}
			} else {
				marshalledDatadogDefinition, _ := json.Marshal(datadogWidget.Definition)
				terraformDefinition = string(marshalledDatadogDefinition)
			}
			terraformWidget["definition"] = terraformDefinition
		}

		// Build widget layout
		if v, ok := datadogWidget.GetLayoutOk(); ok {
			terraformWidget["layout"] = buildTerraformWidgetLayout(v)
		}

		terraformWidgets[i] = terraformWidget
	}
	return &terraformWidgets, nil
}

// unmarshalWidgetDefinition is a custom unmarshal for Terraform widget definition. If first tries to unmarshal the definition
// against a light struct to get the widget type. Then based on the widget type, it will try to unmarshal the definition
// against the right definition struct.
func unmarshalWidgetDefinition(terraformDefinition []byte) (interface{}, error) {
	var definitionHandler struct {
		Type *string `json:"type"`
	}
	if err := json.Unmarshal(terraformDefinition, &definitionHandler); err != nil {
		return nil, err
	}
	switch *definitionHandler.Type {
	case datadog.ALERT_GRAPH_WIDGET:
		var definition datadog.AlertGraphDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.ALERT_VALUE_WIDGET:
		var definition datadog.AlertValueDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.CHANGE_WIDGET:
		var definition datadog.ChangeDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.CHECK_STATUS_WIDGET:
		var definition datadog.CheckStatusDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.DISTRIBUTION_WIDGET:
		var definition datadog.DistributionDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.EVENT_STREAM_WIDGET:
		var definition datadog.EventStreamDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.EVENT_TIMELINE_WIDGET:
		var definition datadog.EventTimelineDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.FREE_TEXT_WIDGET:
		var definition datadog.FreeTextDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.HEATMAP_WIDGET:
		var definition datadog.HeatmapDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.HOSTMAP_WIDGET:
		var definition datadog.HostmapDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.IFRAME_WIDGET:
		var definition datadog.IframeDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.IMAGE_WIDGET:
		var definition datadog.ImageDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.LOG_STREAM_WIDGET:
		var definition datadog.LogStreamDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.MANAGE_STATUS_WIDGET:
		var definition datadog.ManageStatusDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.NOTE_WIDGET:
		var definition datadog.NoteDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.QUERY_VALUE_WIDGET:
		var definition datadog.QueryValueDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.SCATTERPLOT_WIDGET:
		var definition datadog.ScatterplotDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.TIMESERIES_WIDGET:
		var definition datadog.TimeseriesDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.TOPLIST_WIDGET:
		var definition datadog.ToplistDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	case datadog.TRACE_SERVICE_WIDGET:
		var definition datadog.TraceServiceDefinition
		if err := json.Unmarshal(terraformDefinition, &definition); err != nil {
			return nil, err
		}
		return definition, nil
	default:
		return nil, fmt.Errorf("Cannot unmarshal widget definition of type: %s", *definitionHandler.Type)
	}
}

//
// Group Widget helpers
//

func getGroupDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"widget": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The list of widgets to display in this group.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"definition": {
						Type:        schema.TypeString,
						Description: "The definition of the widget.",
						Required:    true,
					},
					"layout": {
						Type:        schema.TypeMap,
						Optional:    true,
						Description: "The layout of the widget on a 'free' dashboard.",
						Elem: &schema.Resource{
							Schema: getWidgetLayoutSchema(),
						},
					},
				},
			},
		},
		"layout_type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The layout type of the group, only 'ordered' for now.",
		},
		"title": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The title of the group.",
		},
	}
}

func buildDatadogGroupDefinition(terraformGroupDefinition map[string]interface{}) (*datadog.GroupDefinition, error) {
	datadogGroupDefinition := datadog.GroupDefinition{}
	datadogGroupDefinition.SetType(datadog.GROUP_WIDGET)

	if v, ok := terraformGroupDefinition["widget"].([]interface{}); ok && len(v) != 0 {
		datadogWidgets, err := buildDatadogWidgets(&v)
		if err != nil {
			return nil, err
		}
		datadogGroupDefinition.Widgets = *datadogWidgets
	}
	if v, ok := terraformGroupDefinition["layout_type"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetLayoutType(v)
	}
	if v, ok := terraformGroupDefinition["title"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetTitle(v)
	}

	return &datadogGroupDefinition, nil
}

func buildTerraformGroupDefinition(datadogGroupDefinition datadog.GroupDefinition, currentGroupWidgets []interface{}) (map[string]interface{}, error) {
	terraformGroupDefinition := map[string]interface{}{}

	terraformWidgets, err := buildTerraformWidgets(&datadogGroupDefinition.Widgets, currentGroupWidgets)
	if err != nil {
		return nil, err
	}
	terraformGroupDefinition["widget"] = terraformWidgets

	if v, ok := datadogGroupDefinition.GetLayoutTypeOk(); ok {
		terraformGroupDefinition["layout_type"] = v
	}
	if v, ok := datadogGroupDefinition.GetTitleOk(); ok {
		terraformGroupDefinition["title"] = v
	}

	return terraformGroupDefinition, nil
}

//
// Widget Layout helpers
//

func getWidgetLayoutSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"x": {
			Type:     schema.TypeFloat,
			Required: true,
		},
		"y": {
			Type:     schema.TypeFloat,
			Required: true,
		},
		"width": {
			Type:     schema.TypeFloat,
			Required: true,
		},
		"height": {
			Type:     schema.TypeFloat,
			Required: true,
		},
	}
}

func buildDatadogWidgetLayout(terraformLayout map[string]interface{}) datadog.WidgetLayout {
	datadogLayout := datadog.WidgetLayout{}

	if _v, ok := terraformLayout["x"].(string); ok && len(_v) != 0 {
		if v, err := strconv.ParseFloat(_v, 64); err == nil {
			datadogLayout.SetX(v)
		}
	}
	if _v, ok := terraformLayout["y"].(string); ok && len(_v) != 0 {
		if v, err := strconv.ParseFloat(_v, 64); err == nil {
			datadogLayout.SetY(v)
		}
	}
	if _v, ok := terraformLayout["height"].(string); ok && len(_v) != 0 {
		if v, err := strconv.ParseFloat(_v, 64); err == nil {
			datadogLayout.SetHeight(v)
		}
	}
	if _v, ok := terraformLayout["width"].(string); ok && len(_v) != 0 {
		if v, err := strconv.ParseFloat(_v, 64); err == nil {
			datadogLayout.SetWidth(v)
		}
	}
	return datadogLayout
}

func buildTerraformWidgetLayout(datadogLayout datadog.WidgetLayout) map[string]string {
	terraformLayout := map[string]string{}

	if v, ok := datadogLayout.GetXOk(); ok {
		terraformLayout["x"] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	if v, ok := datadogLayout.GetYOk(); ok {
		terraformLayout["y"] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	if v, ok := datadogLayout.GetHeightOk(); ok {
		terraformLayout["height"] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	if v, ok := datadogLayout.GetWidthOk(); ok {
		terraformLayout["width"] = strconv.FormatFloat(v, 'f', -1, 64)
	}
	return terraformLayout
}
