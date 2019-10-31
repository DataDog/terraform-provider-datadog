package datadog

import (
	"fmt"

	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	datadog "github.com/zorkian/go-datadog-api"
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
					Schema: getWidgetSchema(),
				},
			},
			"layout_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The layout type of the dashboard, either 'free' or 'ordered'.",
				ValidateFunc: validateDashboardLayoutType,
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
	return resourceDatadogDashboardRead(d, meta)
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

	if err = d.Set("title", dashboard.GetTitle()); err != nil {
		return err
	}
	if err = d.Set("layout_type", dashboard.GetLayoutType()); err != nil {
		return err
	}
	if err = d.Set("description", dashboard.GetDescription()); err != nil {
		return err
	}
	if err = d.Set("is_read_only", dashboard.GetIsReadOnly()); err != nil {
		return err
	}

	// Set widgets
	terraformWidgets, err := buildTerraformWidgets(&dashboard.Widgets)
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

// The generic widget schema is a combination of the schema for a non-group widget
// and the schema for a Group Widget (which can contains only non-group widgets)
func getWidgetSchema() map[string]*schema.Schema {
	widgetSchema := getNonGroupWidgetSchema()
	widgetSchema["group_definition"] = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		MaxItems:    1,
		Description: "The definition for a Group widget",
		Elem: &schema.Resource{
			Schema: getGroupDefinitionSchema(),
		},
	}
	return widgetSchema
}

func getNonGroupWidgetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"layout": {
			Type:        schema.TypeMap,
			Optional:    true,
			Description: "The layout of the widget on a 'free' dashboard",
			Elem: &schema.Resource{
				Schema: getWidgetLayoutSchema(),
			},
		},
		// A widget should implement exactly one of the following definitions
		"alert_graph_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Alert Graph widget",
			Elem: &schema.Resource{
				Schema: getAlertGraphDefinitionSchema(),
			},
		},
		"alert_value_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Alert Value widget",
			Elem: &schema.Resource{
				Schema: getAlertValueDefinitionSchema(),
			},
		},
		"change_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Change  widget",
			Elem: &schema.Resource{
				Schema: getChangeDefinitionSchema(),
			},
		},
		"check_status_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Check Status widget",
			Elem: &schema.Resource{
				Schema: getCheckStatusDefinitionSchema(),
			},
		},
		"distribution_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Distribution widget",
			Elem: &schema.Resource{
				Schema: getDistributionDefinitionSchema(),
			},
		},
		"event_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Event Stream widget",
			Elem: &schema.Resource{
				Schema: getEventStreamDefinitionSchema(),
			},
		},
		"event_timeline_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Event Timeline widget",
			Elem: &schema.Resource{
				Schema: getEventTimelineDefinitionSchema(),
			},
		},
		"free_text_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Free Text widget",
			Elem: &schema.Resource{
				Schema: getFreeTextDefinitionSchema(),
			},
		},
		"heatmap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Heatmap widget",
			Elem: &schema.Resource{
				Schema: getHeatmapDefinitionSchema(),
			},
		},
		"hostmap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Hostmap widget",
			Elem: &schema.Resource{
				Schema: getHostmapDefinitionSchema(),
			},
		},
		"iframe_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Iframe widget",
			Elem: &schema.Resource{
				Schema: getIframeDefinitionSchema(),
			},
		},
		"image_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Image widget",
			Elem: &schema.Resource{
				Schema: getImageDefinitionSchema(),
			},
		},
		"log_stream_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Log Stream widget",
			Elem: &schema.Resource{
				Schema: getLogStreamDefinitionSchema(),
			},
		},
		"manage_status_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for an Manage Status widget",
			Elem: &schema.Resource{
				Schema: getManageStatusDefinitionSchema(),
			},
		},

		"note_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Note widget",
			Elem: &schema.Resource{
				Schema: getNoteDefinitionSchema(),
			},
		},
		"query_value_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Query Value widget",
			Elem: &schema.Resource{
				Schema: getQueryValueDefinitionSchema(),
			},
		},
		"scatterplot_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Scatterplot widget",
			Elem: &schema.Resource{
				Schema: getScatterplotDefinitionSchema(),
			},
		},
		"service_level_objective_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Service Level Objective widget",
			Elem: &schema.Resource{
				Schema: getServiceLevelObjectiveDefinitionSchema(),
			},
		},
		"timeseries_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Timeseries widget",
			Elem: &schema.Resource{
				Schema: getTimeseriesDefinitionSchema(),
			},
		},
		"toplist_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Toplist widget",
			Elem: &schema.Resource{
				Schema: getToplistDefinitionSchema(),
			},
		},
		"trace_service_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Trace Service widget",
			Elem: &schema.Resource{
				Schema: getTraceServiceDefinitionSchema(),
			},
		},
	}
}

func buildDatadogWidgets(terraformWidgets *[]interface{}) (*[]datadog.BoardWidget, error) {
	datadogWidgets := make([]datadog.BoardWidget, len(*terraformWidgets))
	for i, terraformWidget := range *terraformWidgets {
		datadogWidget, err := buildDatadogWidget(terraformWidget.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		datadogWidgets[i] = *datadogWidget
	}
	return &datadogWidgets, nil
}

// Helper to build a Datadog widget from a Terraform widget
func buildDatadogWidget(terraformWidget map[string]interface{}) (*datadog.BoardWidget, error) {
	datadogWidget := datadog.BoardWidget{}

	// Build widget layout
	if v, ok := terraformWidget["layout"].(map[string]interface{}); ok && len(v) != 0 {
		datadogWidget.SetLayout(buildDatadogWidgetLayout(v))
	}

	// Build widget Definition
	if _def, ok := terraformWidget["group_definition"].([]interface{}); ok && len(_def) > 0 {
		if groupDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogDefinition, err := buildDatadogGroupDefinition(groupDefinition)
			if err != nil {
				return nil, err
			}
			datadogWidget.Definition = datadogDefinition
		}
	} else if _def, ok := terraformWidget["alert_graph_definition"].([]interface{}); ok && len(_def) > 0 {
		if alertGraphDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogAlertGraphDefinition(alertGraphDefinition)
		}
	} else if _def, ok := terraformWidget["alert_value_definition"].([]interface{}); ok && len(_def) > 0 {
		if alertValueDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogAlertValueDefinition(alertValueDefinition)
		}
	} else if _def, ok := terraformWidget["change_definition"].([]interface{}); ok && len(_def) > 0 {
		if changeDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogChangeDefinition(changeDefinition)
		}
	} else if _def, ok := terraformWidget["check_status_definition"].([]interface{}); ok && len(_def) > 0 {
		if checkStatusDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogCheckStatusDefinition(checkStatusDefinition)
		}
	} else if _def, ok := terraformWidget["distribution_definition"].([]interface{}); ok && len(_def) > 0 {
		if distributionDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogDistributionDefinition(distributionDefinition)
		}
	} else if _def, ok := terraformWidget["event_stream_definition"].([]interface{}); ok && len(_def) > 0 {
		if eventStreamDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogEventStreamDefinition(eventStreamDefinition)
		}
	} else if _def, ok := terraformWidget["event_timeline_definition"].([]interface{}); ok && len(_def) > 0 {
		if eventTimelineDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogEventTimelineDefinition(eventTimelineDefinition)
		}
	} else if _def, ok := terraformWidget["free_text_definition"].([]interface{}); ok && len(_def) > 0 {
		if freeTextDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogFreeTextDefinition(freeTextDefinition)
		}
	} else if _def, ok := terraformWidget["heatmap_definition"].([]interface{}); ok && len(_def) > 0 {
		if heatmapDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogHeatmapDefinition(heatmapDefinition)
		}
	} else if _def, ok := terraformWidget["hostmap_definition"].([]interface{}); ok && len(_def) > 0 {
		if hostDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogHostmapDefinition(hostDefinition)
		}
	} else if _def, ok := terraformWidget["iframe_definition"].([]interface{}); ok && len(_def) > 0 {
		if iframeDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogIframeDefinition(iframeDefinition)
		}
	} else if _def, ok := terraformWidget["image_definition"].([]interface{}); ok && len(_def) > 0 {
		if imageDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogImageDefinition(imageDefinition)
		}
	} else if _def, ok := terraformWidget["log_stream_definition"].([]interface{}); ok && len(_def) > 0 {
		if logStreamDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogLogStreamDefinition(logStreamDefinition)
		}
	} else if _def, ok := terraformWidget["manage_status_definition"].([]interface{}); ok && len(_def) > 0 {
		if manageStatusDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogManageStatusDefinition(manageStatusDefinition)
		}
	} else if _def, ok := terraformWidget["note_definition"].([]interface{}); ok && len(_def) > 0 {
		if noteDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogNoteDefinition(noteDefinition)
		}
	} else if _def, ok := terraformWidget["query_value_definition"].([]interface{}); ok && len(_def) > 0 {
		if queryValueDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogQueryValueDefinition(queryValueDefinition)
		}
	} else if _def, ok := terraformWidget["scatterplot_definition"].([]interface{}); ok && len(_def) > 0 {
		if scatterplotDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogScatterplotDefinition(scatterplotDefinition)
		}
	} else if _def, ok := terraformWidget["service_level_objective_definition"].([]interface{}); ok && len(_def) > 0 {
		if serviceLevelObjectiveDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogServiceLevelObjectiveDefinition(serviceLevelObjectiveDefinition)
		}
	} else if _def, ok := terraformWidget["timeseries_definition"].([]interface{}); ok && len(_def) > 0 {
		if timeseriesDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogTimeseriesDefinition(timeseriesDefinition)
		}
	} else if _def, ok := terraformWidget["toplist_definition"].([]interface{}); ok && len(_def) > 0 {
		if toplistDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogToplistDefinition(toplistDefinition)
		}
	} else if _def, ok := terraformWidget["trace_service_definition"].([]interface{}); ok && len(_def) > 0 {
		if traceServiceDefinition, ok := _def[0].(map[string]interface{}); ok {
			datadogWidget.Definition = buildDatadogTraceServiceDefinition(traceServiceDefinition)
		}
	} else {
		return nil, fmt.Errorf("Failed to find valid definition in widget configuration")
	}

	return &datadogWidget, nil
}

// Helper to build a list of Terraform widgets from a list of Datadog widgets
func buildTerraformWidgets(datadogWidgets *[]datadog.BoardWidget) (*[]map[string]interface{}, error) {
	terraformWidgets := make([]map[string]interface{}, len(*datadogWidgets))
	for i, datadogWidget := range *datadogWidgets {
		terraformWidget, err := buildTerraformWidget(datadogWidget)
		if err != nil {
			return nil, err
		}
		terraformWidgets[i] = terraformWidget
	}
	return &terraformWidgets, nil
}

// Helper to build a Terraform widget from a Datadog widget
func buildTerraformWidget(datadogWidget datadog.BoardWidget) (map[string]interface{}, error) {
	terraformWidget := map[string]interface{}{}

	// Build layout
	if datadogWidget.Layout != nil {
		terraformWidget["layout"] = buildTerraformWidgetLayout(*datadogWidget.Layout)
	}

	// Build definition
	widgetType, err := datadogWidget.GetWidgetType()
	if err != nil {
		return nil, err
	}
	switch widgetType {
	case datadog.GROUP_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.GroupDefinition)
		terraformDefinition := buildTerraformGroupDefinition(datadogDefinition)
		terraformWidget["group_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.ALERT_GRAPH_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.AlertGraphDefinition)
		terraformDefinition := buildTerraformAlertGraphDefinition(datadogDefinition)
		terraformWidget["alert_graph_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.ALERT_VALUE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.AlertValueDefinition)
		terraformDefinition := buildTerraformAlertValueDefinition(datadogDefinition)
		terraformWidget["alert_value_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.CHANGE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.ChangeDefinition)
		terraformDefinition := buildTerraformChangeDefinition(datadogDefinition)
		terraformWidget["change_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.CHECK_STATUS_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.CheckStatusDefinition)
		terraformDefinition := buildTerraformCheckStatusDefinition(datadogDefinition)
		terraformWidget["check_status_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.DISTRIBUTION_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.DistributionDefinition)
		terraformDefinition := buildTerraformDistributionDefinition(datadogDefinition)
		terraformWidget["distribution_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.EVENT_STREAM_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.EventStreamDefinition)
		terraformDefinition := buildTerraformEventStreamDefinition(datadogDefinition)
		terraformWidget["event_stream_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.EVENT_TIMELINE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.EventTimelineDefinition)
		terraformDefinition := buildTerraformEventTimelineDefinition(datadogDefinition)
		terraformWidget["event_timeline_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.FREE_TEXT_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.FreeTextDefinition)
		terraformDefinition := buildTerraformFreeTextDefinition(datadogDefinition)
		terraformWidget["free_text_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.HEATMAP_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.HeatmapDefinition)
		terraformDefinition := buildTerraformHeatmapDefinition(datadogDefinition)
		terraformWidget["heatmap_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.HOSTMAP_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.HostmapDefinition)
		terraformDefinition := buildTerraformHostmapDefinition(datadogDefinition)
		terraformWidget["hostmap_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.IFRAME_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.IframeDefinition)
		terraformDefinition := buildTerraformIframeDefinition(datadogDefinition)
		terraformWidget["iframe_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.IMAGE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.ImageDefinition)
		terraformDefinition := buildTerraformImageDefinition(datadogDefinition)
		terraformWidget["image_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.LOG_STREAM_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.LogStreamDefinition)
		terraformDefinition := buildTerraformLogStreamDefinition(datadogDefinition)
		terraformWidget["log_stream_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.MANAGE_STATUS_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.ManageStatusDefinition)
		terraformDefinition := buildTerraformManageStatusDefinition(datadogDefinition)
		terraformWidget["manage_status_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.NOTE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.NoteDefinition)
		terraformDefinition := buildTerraformNoteDefinition(datadogDefinition)
		terraformWidget["note_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.QUERY_VALUE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.QueryValueDefinition)
		terraformDefinition := buildTerraformQueryValueDefinition(datadogDefinition)
		terraformWidget["query_value_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.SCATTERPLOT_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.ScatterplotDefinition)
		terraformDefinition := buildTerraformScatterplotDefinition(datadogDefinition)
		terraformWidget["scatterplot_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.SERVICE_LEVEL_OBJECTIVE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.ServiceLevelObjectiveDefinition)
		terraformDefinition := buildTerraformServiceLevelObjectiveDefinition(datadogDefinition)
		terraformWidget["service_level_objective_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.TIMESERIES_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.TimeseriesDefinition)
		terraformDefinition := buildTerraformTimeseriesDefinition(datadogDefinition)
		terraformWidget["timeseries_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.TOPLIST_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.ToplistDefinition)
		terraformDefinition := buildTerraformToplistDefinition(datadogDefinition)
		terraformWidget["toplist_definition"] = []map[string]interface{}{terraformDefinition}
	case datadog.TRACE_SERVICE_WIDGET:
		datadogDefinition := datadogWidget.Definition.(datadog.TraceServiceDefinition)
		terraformDefinition := buildTerraformTraceServiceDefinition(datadogDefinition)
		terraformWidget["trace_service_definition"] = []map[string]interface{}{terraformDefinition}
	default:
		return nil, fmt.Errorf("Unsupported widget type: %s - %s", widgetType, datadog.TIMESERIES_WIDGET)
	}

	return terraformWidget, nil
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

//
// Group Widget helpers
//

func getGroupDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"widget": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The list of widgets in this group.",
			Elem: &schema.Resource{
				Schema: getNonGroupWidgetSchema(),
			},
		},
		"layout_type": {
			Type:         schema.TypeString,
			Required:     true,
			Description:  "The layout type of the group, only 'ordered' for now.",
			ValidateFunc: validateGroupWidgetLayoutType,
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

func buildTerraformGroupDefinition(datadogGroupDefinition datadog.GroupDefinition) map[string]interface{} {
	terraformGroupDefinition := map[string]interface{}{}

	groupWidgets := []map[string]interface{}{}
	for _, datadogGroupWidgets := range datadogGroupDefinition.Widgets {
		newGroupWidget, _ := buildTerraformWidget(datadogGroupWidgets)
		groupWidgets = append(groupWidgets, newGroupWidget)
	}
	terraformGroupDefinition["widget"] = groupWidgets

	if v, ok := datadogGroupDefinition.GetLayoutTypeOk(); ok {
		terraformGroupDefinition["layout_type"] = v
	}
	if v, ok := datadogGroupDefinition.GetTitleOk(); ok {
		terraformGroupDefinition["title"] = v
	}

	return terraformGroupDefinition
}

//
// Alert Graph Widget Definition helpers
//

func getAlertGraphDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alert_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"viz_type": {
			Type:     schema.TypeString,
			Required: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogAlertGraphDefinition(terraformDefinition map[string]interface{}) *datadog.AlertGraphDefinition {
	datadogDefinition := &datadog.AlertGraphDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.ALERT_GRAPH_WIDGET)
	datadogDefinition.AlertId = datadog.String(terraformDefinition["alert_id"].(string))
	datadogDefinition.VizType = datadog.String(terraformDefinition["viz_type"].(string))
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleAlign = datadog.String(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}

func buildTerraformAlertGraphDefinition(datadogDefinition datadog.AlertGraphDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["alert_id"] = *datadogDefinition.AlertId
	terraformDefinition["viz_type"] = *datadogDefinition.VizType
	// Optional params
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

//
// Alert Value Widget Definition helpers
//

func getAlertValueDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"alert_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"precision": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"unit": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"text_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogAlertValueDefinition(terraformDefinition map[string]interface{}) *datadog.AlertValueDefinition {
	datadogDefinition := &datadog.AlertValueDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.ALERT_VALUE_WIDGET)
	datadogDefinition.AlertId = datadog.String(terraformDefinition["alert_id"].(string))
	// Optional params
	if v, ok := terraformDefinition["precision"].(int); ok && v != 0 {
		datadogDefinition.SetPrecision(v)
	}
	if v, ok := terraformDefinition["unit"].(string); ok && len(v) != 0 {
		datadogDefinition.SetUnit(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(v)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	return datadogDefinition
}

func buildTerraformAlertValueDefinition(datadogDefinition datadog.AlertValueDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["alert_id"] = *datadogDefinition.AlertId
	// Optional params
	if datadogDefinition.Precision != nil {
		terraformDefinition["precision"] = *datadogDefinition.Precision
	}
	if datadogDefinition.Unit != nil {
		terraformDefinition["unit"] = *datadogDefinition.Unit
	}
	if datadogDefinition.TextAlign != nil {
		terraformDefinition["text_align"] = *datadogDefinition.TextAlign
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	return terraformDefinition
}

//
// Change Widget Definition helpers
//

func getChangeDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getChangeRequestSchema(),
			},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}
func buildDatadogChangeDefinition(terraformDefinition map[string]interface{}) *datadog.ChangeDefinition {
	datadogDefinition := &datadog.ChangeDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.CHANGE_WIDGET)
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogChangeRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformChangeDefinition(datadogDefinition datadog.ChangeDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformChangeRequests(&datadogDefinition.Requests)
	// Optional params
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

func getChangeRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to Change requests
		"change_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"compare_to": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"increase_good": {
			Type:     schema.TypeBool,
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
		"show_present": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}
func buildDatadogChangeRequests(terraformRequests *[]interface{}) *[]datadog.ChangeRequest {
	datadogRequests := make([]datadog.ChangeRequest, len(*terraformRequests))
	for i, _request := range *terraformRequests {
		terraformRequest := _request.(map[string]interface{})
		// Build ChangeRequest
		datadogChangeRequest := datadog.ChangeRequest{}
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetMetricQuery(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogChangeRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogChangeRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogChangeRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}

		if v, ok := terraformRequest["change_type"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetChangeType(v)
		}
		if v, ok := terraformRequest["compare_to"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetCompareTo(v)
		}
		if v, ok := terraformRequest["increase_good"].(bool); ok {
			datadogChangeRequest.SetIncreaseGood(v)
		}
		if v, ok := terraformRequest["order_by"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetOrderBy(v)
		}
		if v, ok := terraformRequest["order_dir"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetOrderDir(v)
		}
		if v, ok := terraformRequest["show_present"].(bool); ok {
			datadogChangeRequest.SetShowPresent(v)
		}

		datadogRequests[i] = datadogChangeRequest
	}
	return &datadogRequests
}
func buildTerraformChangeRequests(datadogChangeRequests *[]datadog.ChangeRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogChangeRequests))
	for i, datadogRequest := range *datadogChangeRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.MetricQuery != nil {
			terraformRequest["q"] = *datadogRequest.MetricQuery
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}

		if datadogRequest.ChangeType != nil {
			terraformRequest["change_type"] = *datadogRequest.ChangeType
		}
		if datadogRequest.CompareTo != nil {
			terraformRequest["compare_to"] = *datadogRequest.CompareTo
		}
		if datadogRequest.IncreaseGood != nil {
			terraformRequest["increase_good"] = *datadogRequest.IncreaseGood
		}
		if datadogRequest.OrderBy != nil {
			terraformRequest["order_by"] = *datadogRequest.OrderBy
		}
		if datadogRequest.OrderDir != nil {
			terraformRequest["order_dir"] = *datadogRequest.OrderDir
		}
		if datadogRequest.ShowPresent != nil {
			terraformRequest["show_present"] = *datadogRequest.ShowPresent
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Distribution Widget Definition helpers
//

func getDistributionDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getDistributionRequestSchema(),
			},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}
func buildDatadogDistributionDefinition(terraformDefinition map[string]interface{}) *datadog.DistributionDefinition {
	datadogDefinition := &datadog.DistributionDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.DISTRIBUTION_WIDGET)
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogDistributionRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformDistributionDefinition(datadogDefinition datadog.DistributionDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformDistributionRequests(&datadogDefinition.Requests)
	// Optional params
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

func getDistributionRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to Distribution requests
		"style": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
}
func buildDatadogDistributionRequests(terraformRequests *[]interface{}) *[]datadog.DistributionRequest {
	datadogRequests := make([]datadog.DistributionRequest, len(*terraformRequests))
	for i, _request := range *terraformRequests {
		terraformRequest := _request.(map[string]interface{})
		// Build DistributionRequest
		datadogDistributionRequest := datadog.DistributionRequest{}
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogDistributionRequest.SetMetricQuery(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}
		if _style, ok := terraformRequest["style"].([]interface{}); ok && len(_style) > 0 {
			if v, ok := _style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogDistributionRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}

		datadogRequests[i] = datadogDistributionRequest
	}
	return &datadogRequests
}
func buildTerraformDistributionRequests(datadogDistributionRequests *[]datadog.DistributionRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogDistributionRequests))
	for i, datadogRequest := range *datadogDistributionRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.MetricQuery != nil {
			terraformRequest["q"] = *datadogRequest.MetricQuery
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}
		if datadogRequest.Style != nil {
			_style := buildTerraformWidgetRequestStyle(*datadogRequest.Style)
			terraformRequest["style"] = []map[string]interface{}{_style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Event Stream Widget Definition helpers
//

func getEventStreamDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"query": {
			Type:     schema.TypeString,
			Required: true,
		},
		"event_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogEventStreamDefinition(terraformDefinition map[string]interface{}) *datadog.EventStreamDefinition {
	datadogDefinition := &datadog.EventStreamDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.EVENT_STREAM_WIDGET)
	datadogDefinition.Query = datadog.String(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["event_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetEventSize(v)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}

func buildTerraformEventStreamDefinition(datadogDefinition datadog.EventStreamDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = *datadogDefinition.Query
	// Optional params
	if datadogDefinition.EventSize != nil {
		terraformDefinition["event_size"] = *datadogDefinition.EventSize
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

//
// Event Timeline Widget Definition helpers
//

func getEventTimelineDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"query": {
			Type:     schema.TypeString,
			Required: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogEventTimelineDefinition(terraformDefinition map[string]interface{}) *datadog.EventTimelineDefinition {
	datadogDefinition := &datadog.EventTimelineDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.EVENT_TIMELINE_WIDGET)
	datadogDefinition.Query = datadog.String(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}

func buildTerraformEventTimelineDefinition(datadogDefinition datadog.EventTimelineDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = *datadogDefinition.Query
	// Optional params
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

//
// Check Status Widget Definition helpers
//

func getCheckStatusDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"check": {
			Type:     schema.TypeString,
			Required: true,
		},
		"grouping": {
			Type:     schema.TypeString,
			Required: true,
		},
		"group": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"group_by": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"tags": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogCheckStatusDefinition(terraformDefinition map[string]interface{}) *datadog.CheckStatusDefinition {
	datadogDefinition := &datadog.CheckStatusDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.CHECK_STATUS_WIDGET)
	datadogDefinition.Check = datadog.String(terraformDefinition["check"].(string))
	datadogDefinition.Grouping = datadog.String(terraformDefinition["grouping"].(string))
	// Optional params
	if v, ok := terraformDefinition["group"].(string); ok && len(v) != 0 {
		datadogDefinition.SetGroup(v)
	}
	if terraformGroupBys, ok := terraformDefinition["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]string, len(terraformGroupBys))
		for i, groupBy := range terraformGroupBys {
			datadogGroupBys[i] = groupBy.(string)
		}
		datadogDefinition.GroupBy = datadogGroupBys
	}
	if terraformTags, ok := terraformDefinition["tags"].([]interface{}); ok && len(terraformTags) > 0 {
		datadogTags := make([]string, len(terraformTags))
		for i, tag := range terraformTags {
			datadogTags[i] = tag.(string)
		}
		datadogDefinition.Tags = datadogTags
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}

func buildTerraformCheckStatusDefinition(datadogDefinition datadog.CheckStatusDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["check"] = *datadogDefinition.Check
	terraformDefinition["grouping"] = *datadogDefinition.Grouping
	// Optional params
	if datadogDefinition.Group != nil {
		terraformDefinition["group"] = *datadogDefinition.Group
	}
	if datadogDefinition.GroupBy != nil {
		terraformGroupBys := make([]string, len(datadogDefinition.GroupBy))
		for i, datadogGroupBy := range datadogDefinition.GroupBy {
			terraformGroupBys[i] = datadogGroupBy
		}
		terraformDefinition["group_by"] = terraformGroupBys
	}
	if datadogDefinition.Tags != nil {
		terraformTags := make([]string, len(datadogDefinition.Tags))
		for i, datadogTag := range datadogDefinition.Tags {
			terraformTags[i] = datadogTag
		}
		terraformDefinition["tags"] = terraformTags
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

//
// Free Text Definition helpers
//

func getFreeTextDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"text": {
			Type:     schema.TypeString,
			Required: true,
		},
		"color": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"font_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"text_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogFreeTextDefinition(terraformDefinition map[string]interface{}) *datadog.FreeTextDefinition {
	datadogDefinition := &datadog.FreeTextDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.FREE_TEXT_WIDGET)
	datadogDefinition.SetText(terraformDefinition["text"].(string))
	// Optional params
	if v, ok := terraformDefinition["color"].(string); ok && len(v) != 0 {
		datadogDefinition.SetColor(v)
	}
	if v, ok := terraformDefinition["font_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetFontSize(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(v)
	}
	return datadogDefinition
}

func buildTerraformFreeTextDefinition(datadogDefinition datadog.FreeTextDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["text"] = *datadogDefinition.Text
	// Optional params
	if datadogDefinition.Color != nil {
		terraformDefinition["color"] = *datadogDefinition.Color
	}
	if datadogDefinition.FontSize != nil {
		terraformDefinition["font_size"] = *datadogDefinition.FontSize
	}
	if datadogDefinition.TextAlign != nil {
		terraformDefinition["text_align"] = *datadogDefinition.TextAlign
	}
	return terraformDefinition
}

//
// Heatmap Widget Definition helpers
//

func getHeatmapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getHeatmapRequestSchema(),
			},
		},
		"yaxis": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}
func buildDatadogHeatmapDefinition(terraformDefinition map[string]interface{}) *datadog.HeatmapDefinition {
	datadogDefinition := &datadog.HeatmapDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.HEATMAP_WIDGET)
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogHeatmapRequests(&terraformRequests)
	// Optional params
	if _axis, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(_axis) > 0 {
		if v, ok := _axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(v)
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleAlign = datadog.String(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}
func buildTerraformHeatmapDefinition(datadogDefinition datadog.HeatmapDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformHeatmapRequests(&datadogDefinition.Requests)
	// Optional params
	if datadogDefinition.Yaxis != nil {
		_axis := buildTerraformWidgetAxis(*datadogDefinition.Yaxis)
		terraformDefinition["yaxis"] = []map[string]interface{}{_axis}
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

func getHeatmapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to Heatmap requests
		"style": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
}
func buildDatadogHeatmapRequests(terraformRequests *[]interface{}) *[]datadog.HeatmapRequest {
	datadogRequests := make([]datadog.HeatmapRequest, len(*terraformRequests))
	for i, _request := range *terraformRequests {
		terraformRequest := _request.(map[string]interface{})
		// Build HeatmapRequest
		datadogHeatmapRequest := datadog.HeatmapRequest{}
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogHeatmapRequest.SetMetricQuery(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}
		if _style, ok := terraformRequest["style"].([]interface{}); ok && len(_style) > 0 {
			if v, ok := _style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogHeatmapRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}
		datadogRequests[i] = datadogHeatmapRequest
	}
	return &datadogRequests
}
func buildTerraformHeatmapRequests(datadogHeatmapRequests *[]datadog.HeatmapRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogHeatmapRequests))
	for i, datadogRequest := range *datadogHeatmapRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.MetricQuery != nil {
			terraformRequest["q"] = *datadogRequest.MetricQuery
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}
		if datadogRequest.Style != nil {
			_style := buildTerraformWidgetRequestStyle(*datadogRequest.Style)
			terraformRequest["style"] = []map[string]interface{}{_style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Hostmap Widget Definition helpers
//

func getHostmapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"fill": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: getHostmapRequestSchema(),
						},
					},
					"size": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: getHostmapRequestSchema(),
						},
					},
				},
			},
		},
		"node_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"no_metric_hosts": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"no_group_hosts": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"group": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"scope": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"style": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"palette": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"palette_flip": {
						Type:     schema.TypeBool,
						Optional: true,
					},
					"fill_min": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"fill_max": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogHostmapDefinition(terraformDefinition map[string]interface{}) *datadog.HostmapDefinition {

	// Required params
	datadogDefinition := &datadog.HostmapDefinition{}
	datadogDefinition.SetType(datadog.HOSTMAP_WIDGET)
	if v, ok := terraformDefinition["request"].([]interface{}); ok && len(v) > 0 {
		terraformRequests := v[0].(map[string]interface{})
		datadogRequests := datadog.HostmapRequests{}
		if terraformFillArray, ok := terraformRequests["fill"].([]interface{}); ok && len(terraformFillArray) > 0 {
			terraformFill := terraformFillArray[0].(map[string]interface{})
			datadogRequests.Fill = buildDatadogHostmapRequest(terraformFill)
		}
		if terraformSizeArray, ok := terraformRequests["size"].([]interface{}); ok && len(terraformSizeArray) > 0 {
			terraformSize := terraformSizeArray[0].(map[string]interface{})
			datadogRequests.Size = buildDatadogHostmapRequest(terraformSize)
		}
		datadogDefinition.SetRequests(datadogRequests)
	}

	// Optional params
	if v, ok := terraformDefinition["node_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetNodeType(v)
	}
	if v, ok := terraformDefinition["no_metric_hosts"].(bool); ok {
		datadogDefinition.SetNoMetricHosts(v)
	}
	if v, ok := terraformDefinition["no_group_hosts"].(bool); ok {
		datadogDefinition.SetNoGroupHosts(v)
	}
	if terraformGroups, ok := terraformDefinition["group"].([]interface{}); ok && len(terraformGroups) > 0 {
		datadogGroups := make([]string, len(terraformGroups))
		for i, group := range terraformGroups {
			datadogGroups[i] = group.(string)
		}
		datadogDefinition.Group = datadogGroups
	}
	if terraformScopes, ok := terraformDefinition["scope"].([]interface{}); ok && len(terraformScopes) > 0 {
		datadogScopes := make([]string, len(terraformScopes))
		for i, Scope := range terraformScopes {
			datadogScopes[i] = Scope.(string)
		}
		datadogDefinition.Scope = datadogScopes
	}
	if _style, ok := terraformDefinition["style"].([]interface{}); ok && len(_style) > 0 {
		if v, ok := _style[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Style = buildDatadogHostmapRequestStyle(v)
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	return datadogDefinition
}
func buildTerraformHostmapDefinition(datadogDefinition datadog.HostmapDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformRequests := map[string]interface{}{}
	if datadogDefinition.Requests.Size != nil {
		terraformSize := buildTerraformHostmapRequest(datadogDefinition.Requests.Size)
		terraformRequests["size"] = []map[string]interface{}{*terraformSize}
	}
	if datadogDefinition.Requests.Fill != nil {
		terraformFill := buildTerraformHostmapRequest(datadogDefinition.Requests.Fill)
		terraformRequests["fill"] = []map[string]interface{}{*terraformFill}
	}
	terraformDefinition["request"] = []map[string]interface{}{terraformRequests}
	// Optional params
	if datadogDefinition.NodeType != nil {
		terraformDefinition["node_type"] = *datadogDefinition.NodeType
	}
	if datadogDefinition.NoMetricHosts != nil {
		terraformDefinition["no_metric_hosts"] = *datadogDefinition.NoMetricHosts
	}
	if datadogDefinition.NoGroupHosts != nil {
		terraformDefinition["no_group_hosts"] = *datadogDefinition.NoGroupHosts
	}
	if datadogDefinition.Group != nil {
		terraformGroups := make([]string, len(datadogDefinition.Group))
		for i, datadogGroup := range datadogDefinition.Group {
			terraformGroups[i] = datadogGroup
		}
		terraformDefinition["group"] = terraformGroups
	}
	if datadogDefinition.Scope != nil {
		terraformScopes := make([]string, len(datadogDefinition.Scope))
		for i, datadogScope := range datadogDefinition.Scope {
			terraformScopes[i] = datadogScope
		}
		terraformDefinition["scope"] = terraformScopes
	}
	if datadogDefinition.Style != nil {
		_style := buildTerraformHostmapRequestStyle(*datadogDefinition.Style)
		terraformDefinition["style"] = []map[string]interface{}{_style}
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	return terraformDefinition
}

func getHostmapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement at least one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
	}
}
func buildDatadogHostmapRequest(terraformRequest map[string]interface{}) *datadog.HostmapRequest {

	datadogHostmapRequest := &datadog.HostmapRequest{}
	if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
		datadogHostmapRequest.SetMetricQuery(v)
	} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
		apmQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
	} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
		logQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
	} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
		processQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
	}

	return datadogHostmapRequest
}
func buildTerraformHostmapRequest(datadogHostmapRequest *datadog.HostmapRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}
	if datadogHostmapRequest.MetricQuery != nil {
		terraformRequest["q"] = *datadogHostmapRequest.MetricQuery
	} else if datadogHostmapRequest.ApmQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogHostmapRequest.ApmQuery)
		terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogHostmapRequest.LogQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogHostmapRequest.LogQuery)
		terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogHostmapRequest.ProcessQuery != nil {
		terraformQuery := buildTerraformProcessQuery(*datadogHostmapRequest.ProcessQuery)
		terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
	}
	return &terraformRequest
}

//
// Iframe Definition helpers
//

func getIframeDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}

func buildDatadogIframeDefinition(terraformDefinition map[string]interface{}) *datadog.IframeDefinition {
	datadogDefinition := &datadog.IframeDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.IFRAME_WIDGET)
	datadogDefinition.SetUrl(terraformDefinition["url"].(string))
	return datadogDefinition
}

func buildTerraformIframeDefinition(datadogDefinition datadog.IframeDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["url"] = *datadogDefinition.Url
	return terraformDefinition
}

//
// Image Widget Definition helpers
//

func getImageDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"url": {
			Type:     schema.TypeString,
			Required: true,
		},
		"sizing": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"margin": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogImageDefinition(terraformDefinition map[string]interface{}) *datadog.ImageDefinition {
	datadogDefinition := &datadog.ImageDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.IMAGE_WIDGET)
	datadogDefinition.Url = datadog.String(terraformDefinition["url"].(string))
	// Optional params
	if v, ok := terraformDefinition["sizing"].(string); ok && len(v) != 0 {
		datadogDefinition.Sizing = datadog.String(v)
	}
	if v, ok := terraformDefinition["margin"].(string); ok && len(v) != 0 {
		datadogDefinition.Margin = datadog.String(v)
	}
	return datadogDefinition
}

func buildTerraformImageDefinition(datadogDefinition datadog.ImageDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["url"] = *datadogDefinition.Url
	// Optional params
	if datadogDefinition.Sizing != nil {
		terraformDefinition["sizing"] = *datadogDefinition.Sizing
	}
	if datadogDefinition.Margin != nil {
		terraformDefinition["margin"] = *datadogDefinition.Margin
	}
	return terraformDefinition
}

//
// Log Stream Widget Definition helpers
//

func getLogStreamDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"logset": {
			Type:     schema.TypeString,
			Required: true,
		},
		"query": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"columns": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogLogStreamDefinition(terraformDefinition map[string]interface{}) *datadog.LogStreamDefinition {
	datadogDefinition := &datadog.LogStreamDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.LOG_STREAM_WIDGET)
	datadogDefinition.Logset = datadog.String(terraformDefinition["logset"].(string))
	// Optional params
	if v, ok := terraformDefinition["query"].(string); ok && len(v) != 0 {
		datadogDefinition.Query = datadog.String(v)
	}
	if terraformColumns, ok := terraformDefinition["columns"].([]interface{}); ok && len(terraformColumns) > 0 {
		datadogColumns := make([]string, len(terraformColumns))
		for i, column := range terraformColumns {
			datadogColumns[i] = column.(string)
		}
		datadogDefinition.Columns = datadogColumns
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleAlign = datadog.String(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}

func buildTerraformLogStreamDefinition(datadogDefinition datadog.LogStreamDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["logset"] = *datadogDefinition.Logset
	// Optional params
	if datadogDefinition.Query != nil {
		terraformDefinition["query"] = *datadogDefinition.Query
	}
	if datadogDefinition.Columns != nil {
		terraformColumns := make([]string, len(datadogDefinition.Columns))
		for i, datadogColumn := range datadogDefinition.Columns {
			terraformColumns[i] = datadogColumn
		}
		terraformDefinition["columns"] = terraformColumns
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

//
// Manage Status Widget Definition helpers
//

func getManageStatusDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"query": {
			Type:     schema.TypeString,
			Required: true,
		},
		"sort": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"count": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"start": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"display_format": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"color_preference": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"hide_zero_counts": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogManageStatusDefinition(terraformDefinition map[string]interface{}) *datadog.ManageStatusDefinition {
	datadogDefinition := &datadog.ManageStatusDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.MANAGE_STATUS_WIDGET)
	datadogDefinition.Query = datadog.String(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["sort"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSort(v)
	}
	if v, ok := terraformDefinition["count"].(int); ok {
		datadogDefinition.SetCount(v)
	}
	if v, ok := terraformDefinition["start"].(int); ok {
		datadogDefinition.SetStart(v)
	}
	if v, ok := terraformDefinition["display_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetDisplayFormat(v)
	}
	if v, ok := terraformDefinition["color_preference"].(string); ok && len(v) != 0 {
		datadogDefinition.SetColorPreference(v)
	}
	if v, ok := terraformDefinition["hide_zero_counts"].(bool); ok {
		datadogDefinition.SetHideZeroCounts(v)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	return datadogDefinition
}

func buildTerraformManageStatusDefinition(datadogDefinition datadog.ManageStatusDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = *datadogDefinition.Query
	// Optional params
	if datadogDefinition.Sort != nil {
		terraformDefinition["sort"] = *datadogDefinition.Sort
	}
	if datadogDefinition.Count != nil {
		terraformDefinition["count"] = *datadogDefinition.Count
	}
	if datadogDefinition.Start != nil {
		terraformDefinition["start"] = *datadogDefinition.Start
	}
	if datadogDefinition.DisplayFormat != nil {
		terraformDefinition["display_format"] = *datadogDefinition.DisplayFormat
	}
	if datadogDefinition.ColorPreference != nil {
		terraformDefinition["color_preference"] = *datadogDefinition.ColorPreference
	}
	if datadogDefinition.HideZeroCounts != nil {
		terraformDefinition["hide_zero_counts"] = *datadogDefinition.HideZeroCounts
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	return terraformDefinition
}

//
// Note Widget Definition helpers
//

func getNoteDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"content": {
			Type:     schema.TypeString,
			Required: true,
		},
		"background_color": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"font_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"text_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"show_tick": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"tick_pos": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"tick_edge": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogNoteDefinition(terraformDefinition map[string]interface{}) *datadog.NoteDefinition {
	datadogDefinition := &datadog.NoteDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.NOTE_WIDGET)
	datadogDefinition.Content = datadog.String(terraformDefinition["content"].(string))
	// Optional params
	if v, ok := terraformDefinition["background_color"].(string); ok && len(v) != 0 {
		datadogDefinition.BackgroundColor = datadog.String(v)
	}
	if v, ok := terraformDefinition["font_size"].(string); ok && len(v) != 0 {
		datadogDefinition.FontSize = datadog.String(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.TextAlign = datadog.String(v)
	}
	if v, ok := terraformDefinition["show_tick"]; ok {
		datadogDefinition.ShowTick = datadog.Bool(v.(bool))
	}
	if v, ok := terraformDefinition["tick_pos"].(string); ok && len(v) != 0 {
		datadogDefinition.TickPos = datadog.String(v)
	}
	if v, ok := terraformDefinition["tick_edge"].(string); ok && len(v) != 0 {
		datadogDefinition.TickEdge = datadog.String(v)
	}
	return datadogDefinition
}

func buildTerraformNoteDefinition(datadogDefinition datadog.NoteDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["content"] = *datadogDefinition.Content
	// Optional params
	if datadogDefinition.BackgroundColor != nil {
		terraformDefinition["background_color"] = *datadogDefinition.BackgroundColor
	}
	if datadogDefinition.FontSize != nil {
		terraformDefinition["font_size"] = *datadogDefinition.FontSize
	}
	if datadogDefinition.TextAlign != nil {
		terraformDefinition["text_align"] = *datadogDefinition.TextAlign
	}
	if datadogDefinition.ShowTick != nil {
		terraformDefinition["show_tick"] = *datadogDefinition.ShowTick
	}
	if datadogDefinition.TickPos != nil {
		terraformDefinition["tick_pos"] = *datadogDefinition.TickPos
	}
	if datadogDefinition.TickEdge != nil {
		terraformDefinition["tick_edge"] = *datadogDefinition.TickEdge
	}
	return terraformDefinition
}

//
// Query Value Widget Definition helpers
//

func getQueryValueDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getQueryValueRequestSchema(),
			},
		},
		"autoscale": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"custom_unit": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"precision": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"text_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}
func buildDatadogQueryValueDefinition(terraformDefinition map[string]interface{}) *datadog.QueryValueDefinition {
	datadogDefinition := &datadog.QueryValueDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.QUERY_VALUE_WIDGET)
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogQueryValueRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["autoscale"].(bool); ok {
		datadogDefinition.SetAutoscale(v)
	}
	if v, ok := terraformDefinition["custom_unit"].(string); ok && len(v) != 0 {
		datadogDefinition.SetCustomUnit(v)
	}
	if v, ok := terraformDefinition["precision"].(int); ok {
		datadogDefinition.SetPrecision(v)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformQueryValueDefinition(datadogDefinition datadog.QueryValueDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformQueryValueRequests(&datadogDefinition.Requests)
	// Optional params
	if datadogDefinition.Autoscale != nil {
		terraformDefinition["autoscale"] = *datadogDefinition.Autoscale
	}
	if datadogDefinition.CustomUnit != nil {
		terraformDefinition["custom_unit"] = *datadogDefinition.CustomUnit
	}
	if datadogDefinition.Precision != nil {
		terraformDefinition["precision"] = *datadogDefinition.Precision
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TextAlign != nil {
		terraformDefinition["text_align"] = *datadogDefinition.TextAlign
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

func getQueryValueRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to QueryValue requests
		"conditional_formats": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetConditionalFormatSchema(),
			},
		},
		"aggregator": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogQueryValueRequests(terraformRequests *[]interface{}) *[]datadog.QueryValueRequest {
	datadogRequests := make([]datadog.QueryValueRequest, len(*terraformRequests))
	for i, _request := range *terraformRequests {
		terraformRequest := _request.(map[string]interface{})
		// Build QueryValueRequest
		datadogQueryValueRequest := datadog.QueryValueRequest{}
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogQueryValueRequest.SetMetricQuery(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}

		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogQueryValueRequest.ConditionalFormats = *buildDatadogWidgetConditionalFormat(&v)
		}
		if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
			datadogQueryValueRequest.SetAggregator(v)
		}

		datadogRequests[i] = datadogQueryValueRequest
	}
	return &datadogRequests
}
func buildTerraformQueryValueRequests(datadogQueryValueRequests *[]datadog.QueryValueRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogQueryValueRequests))
	for i, datadogRequest := range *datadogQueryValueRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.MetricQuery != nil {
			terraformRequest["q"] = *datadogRequest.MetricQuery
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}

		if datadogRequest.ConditionalFormats != nil {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(&datadogRequest.ConditionalFormats)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}

		if datadogRequest.Aggregator != nil {
			terraformRequest["aggregator"] = *datadogRequest.Aggregator
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Scatterplot Widget Definition helpers
//

func getScatterplotDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			MinItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"x": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: getScatterplotRequestSchema(),
						},
					},
					"y": {
						Type:     schema.TypeList,
						Optional: true,
						Elem: &schema.Resource{
							Schema: getScatterplotRequestSchema(),
						},
					},
				},
			},
		},
		"xaxis": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"yaxis": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"color_by_groups": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}
func buildDatadogScatterplotDefinition(terraformDefinition map[string]interface{}) *datadog.ScatterplotDefinition {
	datadogDefinition := &datadog.ScatterplotDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.SCATTERPLOT_WIDGET)

	if v, ok := terraformDefinition["request"].([]interface{}); ok && len(v) > 0 {
		terraformRequests := v[0].(map[string]interface{})
		datadogRequests := datadog.ScatterplotRequests{}
		if terraformXArray, ok := terraformRequests["x"].([]interface{}); ok && len(terraformXArray) > 0 {
			terraformX := terraformXArray[0].(map[string]interface{})
			datadogRequests.X = buildDatadogScatterplotRequest(terraformX)
		}
		if terraformYArray, ok := terraformRequests["y"].([]interface{}); ok && len(terraformYArray) > 0 {
			terraformY := terraformYArray[0].(map[string]interface{})
			datadogRequests.Y = buildDatadogScatterplotRequest(terraformY)
		}
		datadogDefinition.SetRequests(datadogRequests)
	}

	// Optional params
	if _axis, ok := terraformDefinition["xaxis"].([]interface{}); ok && len(_axis) > 0 {
		if v, ok := _axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Xaxis = buildDatadogWidgetAxis(v)
		}
	}
	if _axis, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(_axis) > 0 {
		if v, ok := _axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(v)
		}
	}
	if terraformColorByGroups, ok := terraformDefinition["color_by_groups"].([]interface{}); ok && len(terraformColorByGroups) > 0 {
		datadogColorByGroups := make([]string, len(terraformColorByGroups))
		for i, colorByGroup := range terraformColorByGroups {
			datadogColorByGroups[i] = colorByGroup.(string)
		}
		datadogDefinition.ColorByGroups = datadogColorByGroups
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformScatterplotDefinition(datadogDefinition datadog.ScatterplotDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformRequests := map[string]interface{}{}
	if datadogDefinition.Requests.X != nil {
		terraformX := buildTerraformScatterplotRequest(datadogDefinition.Requests.X)
		terraformRequests["x"] = []map[string]interface{}{*terraformX}
	}
	if datadogDefinition.Requests.Y != nil {
		terraformY := buildTerraformScatterplotRequest(datadogDefinition.Requests.Y)
		terraformRequests["y"] = []map[string]interface{}{*terraformY}
	}
	terraformDefinition["request"] = []map[string]interface{}{terraformRequests}

	// Optional params
	if datadogDefinition.Xaxis != nil {
		_axis := buildTerraformWidgetAxis(*datadogDefinition.Xaxis)
		terraformDefinition["xaxis"] = []map[string]interface{}{_axis}
	}
	if datadogDefinition.Yaxis != nil {
		_axis := buildTerraformWidgetAxis(*datadogDefinition.Yaxis)
		terraformDefinition["yaxis"] = []map[string]interface{}{_axis}
	}

	if datadogDefinition.ColorByGroups != nil {
		terraformColorByGroups := make([]string, len(datadogDefinition.ColorByGroups))
		for i, datadogColorByGroup := range datadogDefinition.ColorByGroups {
			terraformColorByGroups[i] = datadogColorByGroup
		}
		terraformDefinition["color_by_groups"] = terraformColorByGroups
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

func getScatterplotRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to Scatterplot requests
		"aggregator": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogScatterplotRequest(terraformRequest map[string]interface{}) *datadog.ScatterplotRequest {

	datadogScatterplotRequest := &datadog.ScatterplotRequest{}
	if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
		datadogScatterplotRequest.SetMetricQuery(v)
	} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
		apmQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
	} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
		logQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
	} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
		processQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
	}

	if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
		datadogScatterplotRequest.SetAggregator(v)
	}

	return datadogScatterplotRequest
}
func buildTerraformScatterplotRequest(datadogScatterplotRequest *datadog.ScatterplotRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}
	if datadogScatterplotRequest.MetricQuery != nil {
		terraformRequest["q"] = *datadogScatterplotRequest.MetricQuery
	} else if datadogScatterplotRequest.ApmQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogScatterplotRequest.ApmQuery)
		terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogScatterplotRequest.LogQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogScatterplotRequest.LogQuery)
		terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogScatterplotRequest.ProcessQuery != nil {
		terraformQuery := buildTerraformProcessQuery(*datadogScatterplotRequest.ProcessQuery)
		terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
	}

	if datadogScatterplotRequest.Aggregator != nil {
		terraformRequest["aggregator"] = *datadogScatterplotRequest.Aggregator
	}
	return &terraformRequest
}

//
// Service Level Objective Widget Definition helpers
//

func getServiceLevelObjectiveDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"view_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"slo_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"show_error_budget": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"view_mode": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time_windows": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
}

func buildDatadogServiceLevelObjectiveDefinition(terraformDefinition map[string]interface{}) *datadog.ServiceLevelObjectiveDefinition {
	datadogDefinition := &datadog.ServiceLevelObjectiveDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.SERVICE_LEVEL_OBJECTIVE_WIDGET)

	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["view_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetViewType(v)
	}
	if v, ok := terraformDefinition["slo_id"].(string); ok && len(v) != 0 {
		datadogDefinition.SetServiceLevelObjectiveID(v)
	}
	if v, ok := terraformDefinition["show_error_budget"].(bool); ok {
		datadogDefinition.SetShowErrorBudget(v)
	}
	if v, ok := terraformDefinition["view_mode"].(string); ok && len(v) != 0 {
		datadogDefinition.SetViewMode(v)
	}
	if terraformTimeWindows, ok := terraformDefinition["time_windows"].([]interface{}); ok && len(terraformTimeWindows) > 0 {
		datadogTimeWindows := make([]string, len(terraformTimeWindows))
		for i, timeWindows := range terraformTimeWindows {
			datadogTimeWindows[i] = timeWindows.(string)
		}
		datadogDefinition.TimeWindows = datadogTimeWindows
	}
	return datadogDefinition
}

func buildTerraformServiceLevelObjectiveDefinition(datadogDefinition datadog.ServiceLevelObjectiveDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	// Optional params
	if title, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = title
	}
	if titleSize, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = titleSize
	}
	if titleAlign, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = titleAlign
	}
	if viewType, ok := datadogDefinition.GetViewTypeOk(); ok {
		terraformDefinition["view_type"] = viewType
	}
	if sloID, ok := datadogDefinition.GetServiceLevelObjectiveIDOk(); ok {
		terraformDefinition["slo_id"] = sloID
	}
	if showErrorBudget, ok := datadogDefinition.GetShowErrorBudgetOk(); ok {
		terraformDefinition["show_error_budget"] = showErrorBudget
	}
	if viewMode, ok := datadogDefinition.GetViewModeOk(); ok {
		terraformDefinition["view_mode"] = viewMode
	}
	if datadogDefinition.TimeWindows != nil {
		terraformTimeWindows := make([]string, len(datadogDefinition.TimeWindows))
		for i, datadogTimeWindow := range datadogDefinition.TimeWindows {
			terraformTimeWindows[i] = datadogTimeWindow
		}
		terraformDefinition["time_windows"] = terraformTimeWindows
	}
	return terraformDefinition
}

//
// Timeseries Widget Definition helpers
//

func getTimeseriesDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getTimeseriesRequestSchema(),
			},
		},
		"marker": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetMarkerSchema(),
			},
		},
		"event": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetEventSchema(),
			},
		},
		"yaxis": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetAxisSchema(),
			},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"show_legend": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"legend_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogTimeseriesDefinition(terraformDefinition map[string]interface{}) *datadog.TimeseriesDefinition {
	datadogDefinition := &datadog.TimeseriesDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.TIMESERIES_WIDGET)
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogTimeseriesRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["marker"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Markers = *buildDatadogWidgetMarkers(&v)
	}
	if v, ok := terraformDefinition["event"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Events = *buildDatadogWidgetEvents(&v)
	}
	if v, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(v) > 0 {
		if _axis, ok := v[0].(map[string]interface{}); ok && len(_axis) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(_axis)
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleAlign = datadog.String(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.ShowLegend = datadog.Bool(v)
	}
	return datadogDefinition
}

func buildTerraformTimeseriesDefinition(datadogDefinition datadog.TimeseriesDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformTimeseriesRequests(&datadogDefinition.Requests)
	// Optional params
	if datadogDefinition.Markers != nil {
		terraformDefinition["marker"] = buildTerraformWidgetMarkers(&datadogDefinition.Markers)
	}
	if datadogDefinition.Events != nil {
		terraformDefinition["event"] = buildTerraformWidgetEvents(&datadogDefinition.Events)
	}
	if datadogDefinition.Yaxis != nil {
		_axis := buildTerraformWidgetAxis(*datadogDefinition.Yaxis)
		terraformDefinition["yaxis"] = []map[string]interface{}{_axis}
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	if datadogDefinition.ShowLegend != nil {
		terraformDefinition["show_legend"] = *datadogDefinition.ShowLegend
	}
	return terraformDefinition
}

func getTimeseriesRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to Timeseries requests
		"style": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"palette": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"line_type": {
						Type:     schema.TypeString,
						Optional: true,
					},
					"line_width": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"metadata": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"expression": {
						Type:     schema.TypeString,
						Required: true,
					},
					"alias_name": {
						Type:     schema.TypeString,
						Optional: true,
					},
				},
			},
		},
		"display_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogTimeseriesRequests(terraformRequests *[]interface{}) *[]datadog.TimeseriesRequest {
	datadogRequests := make([]datadog.TimeseriesRequest, len(*terraformRequests))
	for i, _request := range *terraformRequests {
		terraformRequest := _request.(map[string]interface{})
		// Build TimeseriesRequest
		datadogTimeseriesRequest := datadog.TimeseriesRequest{}
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogTimeseriesRequest.SetMetricQuery(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}
		if _style, ok := terraformRequest["style"].([]interface{}); ok && len(_style) > 0 {
			if v, ok := _style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogTimeseriesRequest.Style = buildDatadogTimeseriesRequestStyle(v)
			}
		}
		// Metadata
		if terraformMetadataList, ok := terraformRequest["metadata"].([]interface{}); ok && len(terraformMetadataList) > 0 {
			datadogMetadataList := make([]datadog.WidgetMetadata, len(terraformMetadataList))
			for i, _metadata := range terraformMetadataList {
				metadata := _metadata.(map[string]interface{})
				// Expression
				datadogMetadata := datadog.WidgetMetadata{
					Expression: datadog.String(metadata["expression"].(string)),
				}
				// AliasName
				if v, ok := metadata["alias_name"].(string); ok && len(v) != 0 {
					datadogMetadata.AliasName = datadog.String(v)
				}
				datadogMetadataList[i] = datadogMetadata
			}
			datadogTimeseriesRequest.Metadata = datadogMetadataList
		}
		if v, ok := terraformRequest["display_type"].(string); ok && len(v) != 0 {
			datadogTimeseriesRequest.DisplayType = datadog.String(v)
		}
		datadogRequests[i] = datadogTimeseriesRequest
	}
	return &datadogRequests
}
func buildTerraformTimeseriesRequests(datadogTimeseriesRequests *[]datadog.TimeseriesRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogTimeseriesRequests))
	for i, datadogRequest := range *datadogTimeseriesRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.MetricQuery != nil {
			terraformRequest["q"] = *datadogRequest.MetricQuery
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}
		if datadogRequest.Style != nil {
			_style := buildTerraformTimeseriesRequestStyle(*datadogRequest.Style)
			terraformRequest["style"] = []map[string]interface{}{_style}
		}
		// Metadata
		if datadogRequest.Metadata != nil {
			terraformMetadataList := make([]map[string]interface{}, len(datadogRequest.Metadata))
			for i, metadata := range datadogRequest.Metadata {
				// Expression
				terraformMetadata := map[string]interface{}{
					"expression": *metadata.Expression,
				}
				// AliasName
				if metadata.AliasName != nil {
					terraformMetadata["alias_name"] = *metadata.AliasName
				}

				terraformMetadataList[i] = terraformMetadata
			}
			terraformRequest["metadata"] = &terraformMetadataList
		}
		if datadogRequest.DisplayType != nil {
			terraformRequest["display_type"] = *datadogRequest.DisplayType
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Toplist Widget Definition helpers
//

func getToplistDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getToplistRequestSchema(),
			},
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}
func buildDatadogToplistDefinition(terraformDefinition map[string]interface{}) *datadog.ToplistDefinition {
	datadogDefinition := &datadog.ToplistDefinition{}
	// Required params
	datadogDefinition.SetType(datadog.TOPLIST_WIDGET)
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogToplistRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadog.String(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleAlign = datadog.String(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}
func buildTerraformToplistDefinition(datadogDefinition datadog.ToplistDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformToplistRequests(&datadogDefinition.Requests)
	// Optional params
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

func getToplistRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":             getMetricQuerySchema(),
		"apm_query":     getApmOrLogQuerySchema(),
		"log_query":     getApmOrLogQuerySchema(),
		"process_query": getProcessQuerySchema(),
		// Settings specific to Toplist requests
		"conditional_formats": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetConditionalFormatSchema(),
			},
		},
		"style": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetRequestStyle(),
			},
		},
	}
}
func buildDatadogToplistRequests(terraformRequests *[]interface{}) *[]datadog.ToplistRequest {
	datadogRequests := make([]datadog.ToplistRequest, len(*terraformRequests))
	for i, _request := range *terraformRequests {
		terraformRequest := _request.(map[string]interface{})
		// Build ToplistRequest
		datadogToplistRequest := datadog.ToplistRequest{}
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogToplistRequest.SetMetricQuery(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogToplistRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogToplistRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogToplistRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}
		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogToplistRequest.ConditionalFormats = *buildDatadogWidgetConditionalFormat(&v)
		}
		if _style, ok := terraformRequest["style"].([]interface{}); ok && len(_style) > 0 {
			if v, ok := _style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogToplistRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}
		datadogRequests[i] = datadogToplistRequest
	}
	return &datadogRequests
}
func buildTerraformToplistRequests(datadogToplistRequests *[]datadog.ToplistRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogToplistRequests))
	for i, datadogRequest := range *datadogToplistRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.MetricQuery != nil {
			terraformRequest["q"] = *datadogRequest.MetricQuery
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		}

		if datadogRequest.ConditionalFormats != nil {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(&datadogRequest.ConditionalFormats)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}
		if datadogRequest.Style != nil {
			_style := buildTerraformWidgetRequestStyle(*datadogRequest.Style)
			terraformRequest["style"] = []map[string]interface{}{_style}
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Trace Service Widget Definition helpers
//

func getTraceServiceDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"env": {
			Type:     schema.TypeString,
			Required: true,
		},
		"service": {
			Type:     schema.TypeString,
			Required: true,
		},
		"span_name": {
			Type:     schema.TypeString,
			Required: true,
		},
		"show_hits": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"show_errors": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"show_latency": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"show_breakdown": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"show_distribution": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"show_resource_list": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"size_format": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"display_format": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_size": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"title_align": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"time": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetTimeSchema(),
			},
		},
	}
}

func buildDatadogTraceServiceDefinition(terraformDefinition map[string]interface{}) *datadog.TraceServiceDefinition {
	datadogDefinition := &datadog.TraceServiceDefinition{}
	// Required params
	datadogDefinition.Type = datadog.String(datadog.TRACE_SERVICE_WIDGET)
	datadogDefinition.Env = datadog.String(terraformDefinition["env"].(string))
	datadogDefinition.Service = datadog.String(terraformDefinition["service"].(string))
	datadogDefinition.SpanName = datadog.String(terraformDefinition["span_name"].(string))
	// Optional params
	if v, ok := terraformDefinition["show_hits"].(bool); ok {
		datadogDefinition.SetShowHits(v)
	}
	if v, ok := terraformDefinition["show_errors"].(bool); ok {
		datadogDefinition.SetShowErrors(v)
	}
	if v, ok := terraformDefinition["show_latency"].(bool); ok {
		datadogDefinition.SetShowLatency(v)
	}
	if v, ok := terraformDefinition["show_breakdown"].(bool); ok {
		datadogDefinition.SetShowBreakdown(v)
	}
	if v, ok := terraformDefinition["show_distribution"].(bool); ok {
		datadogDefinition.SetShowDistribution(v)
	}
	if v, ok := terraformDefinition["show_resource_list"].(bool); ok {
		datadogDefinition.SetShowResourceList(v)
	}
	if v, ok := terraformDefinition["size_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSizeFormat(v)
	}
	if v, ok := terraformDefinition["display_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetDisplayFormat(v)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(v)
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}

func buildTerraformTraceServiceDefinition(datadogDefinition datadog.TraceServiceDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["env"] = *datadogDefinition.Env
	terraformDefinition["service"] = *datadogDefinition.Service
	terraformDefinition["span_name"] = *datadogDefinition.SpanName
	// Optional params
	if datadogDefinition.ShowHits != nil {
		terraformDefinition["show_hits"] = *datadogDefinition.ShowHits
	}
	if datadogDefinition.ShowErrors != nil {
		terraformDefinition["show_errors"] = *datadogDefinition.ShowErrors
	}
	if datadogDefinition.ShowLatency != nil {
		terraformDefinition["show_latency"] = *datadogDefinition.ShowLatency
	}
	if datadogDefinition.ShowBreakdown != nil {
		terraformDefinition["show_breakdown"] = *datadogDefinition.ShowBreakdown
	}
	if datadogDefinition.ShowDistribution != nil {
		terraformDefinition["show_distribution"] = *datadogDefinition.ShowDistribution
	}
	if datadogDefinition.ShowResourceList != nil {
		terraformDefinition["show_resource_list"] = *datadogDefinition.ShowResourceList
	}
	if datadogDefinition.SizeFormat != nil {
		terraformDefinition["size_format"] = *datadogDefinition.SizeFormat
	}
	if datadogDefinition.DisplayFormat != nil {
		terraformDefinition["display_format"] = *datadogDefinition.DisplayFormat
	}
	if datadogDefinition.Title != nil {
		terraformDefinition["title"] = *datadogDefinition.Title
	}
	if datadogDefinition.TitleSize != nil {
		terraformDefinition["title_size"] = *datadogDefinition.TitleSize
	}
	if datadogDefinition.TitleAlign != nil {
		terraformDefinition["title_align"] = *datadogDefinition.TitleAlign
	}
	if datadogDefinition.Time != nil {
		terraformDefinition["time"] = buildTerraformWidgetTime(*datadogDefinition.Time)
	}
	return terraformDefinition
}

// Widget Conditional Format helpers
func getWidgetConditionalFormatSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"comparator": {
			Type:     schema.TypeString,
			Required: true,
		},
		"value": {
			Type:     schema.TypeFloat,
			Required: true,
		},
		"palette": {
			Type:     schema.TypeString,
			Required: true,
		},
		"custom_bg_color": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"custom_fg_color": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"image_url": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"hide_value": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"timeframe": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogWidgetConditionalFormat(terraformWidgetConditionalFormat *[]interface{}) *[]datadog.WidgetConditionalFormat {
	datadogWidgetConditionalFormat := make([]datadog.WidgetConditionalFormat, len(*terraformWidgetConditionalFormat))
	for i, _conditionalFormat := range *terraformWidgetConditionalFormat {
		terraformConditionalFormat := _conditionalFormat.(map[string]interface{})
		datadogConditionalFormat := datadog.WidgetConditionalFormat{}
		// Required
		datadogConditionalFormat.SetComparator(terraformConditionalFormat["comparator"].(string))
		datadogConditionalFormat.SetValue(terraformConditionalFormat["value"].(float64))
		datadogConditionalFormat.SetPalette(terraformConditionalFormat["palette"].(string))
		// Optional
		if v, ok := terraformConditionalFormat["custom_bg_color"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetCustomBgColor(v)
		}
		if v, ok := terraformConditionalFormat["custom_fg_color"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetCustomFgColor(v)
		}
		if v, ok := terraformConditionalFormat["image_url"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetImageUrl(v)
		}
		if v, ok := terraformConditionalFormat["hide_value"].(bool); ok {
			datadogConditionalFormat.SetHideValue(v)
		}
		if v, ok := terraformConditionalFormat["timeframe"].(string); ok && len(v) != 0 {
			datadogConditionalFormat.SetTimeframe(v)
		}
		datadogWidgetConditionalFormat[i] = datadogConditionalFormat
	}
	return &datadogWidgetConditionalFormat
}
func buildTerraformWidgetConditionalFormat(datadogWidgetConditionalFormat *[]datadog.WidgetConditionalFormat) *[]map[string]interface{} {
	terraformWidgetConditionalFormat := make([]map[string]interface{}, len(*datadogWidgetConditionalFormat))
	for i, datadogConditionalFormat := range *datadogWidgetConditionalFormat {
		terraformConditionalFormat := map[string]interface{}{}
		// Required params
		terraformConditionalFormat["comparator"] = *datadogConditionalFormat.Comparator
		terraformConditionalFormat["value"] = *datadogConditionalFormat.Value
		terraformConditionalFormat["palette"] = *datadogConditionalFormat.Palette
		// Optional params
		if datadogConditionalFormat.CustomBgColor != nil {
			terraformConditionalFormat["custom_bg_color"] = *datadogConditionalFormat.CustomBgColor
		}
		if datadogConditionalFormat.CustomFgColor != nil {
			terraformConditionalFormat["custom_fg_color"] = *datadogConditionalFormat.CustomFgColor
		}
		if datadogConditionalFormat.ImageUrl != nil {
			terraformConditionalFormat["image_url"] = *datadogConditionalFormat.ImageUrl
		}
		if datadogConditionalFormat.HideValue != nil {
			terraformConditionalFormat["hide_value"] = *datadogConditionalFormat.HideValue
		}
		if datadogConditionalFormat.Timeframe != nil {
			terraformConditionalFormat["timeframe"] = *datadogConditionalFormat.Timeframe
		}
		terraformWidgetConditionalFormat[i] = terraformConditionalFormat
	}
	return &terraformWidgetConditionalFormat
}

// Widget Event helpers

func getWidgetEventSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"q": {
			Type:     schema.TypeString,
			Required: true,
		},
	}
}
func buildDatadogWidgetEvents(terraformWidgetEvents *[]interface{}) *[]datadog.WidgetEvent {
	datadogWidgetEvents := make([]datadog.WidgetEvent, len(*terraformWidgetEvents))
	for i, _event := range *terraformWidgetEvents {
		terraformEvent := _event.(map[string]interface{})
		datadogWidgetEvent := datadog.WidgetEvent{}
		// Required params
		datadogWidgetEvent.Query = datadog.String(terraformEvent["q"].(string))
		datadogWidgetEvents[i] = datadogWidgetEvent
	}

	return &datadogWidgetEvents
}
func buildTerraformWidgetEvents(datadogWidgetEvents *[]datadog.WidgetEvent) *[]map[string]string {
	terraformWidgetEvents := make([]map[string]string, len(*datadogWidgetEvents))
	for i, datadogWidget := range *datadogWidgetEvents {
		terraformWidget := map[string]string{}
		// Required params
		terraformWidget["q"] = *datadogWidget.Query
		terraformWidgetEvents[i] = terraformWidget
	}
	return &terraformWidgetEvents
}

// Widget Time helpers

func getWidgetTimeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"live_span": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogWidgetTime(terraformWidgetTime map[string]interface{}) *datadog.WidgetTime {
	datadogWidgetTime := &datadog.WidgetTime{}
	if v, ok := terraformWidgetTime["live_span"].(string); ok && len(v) != 0 {
		datadogWidgetTime.LiveSpan = datadog.String(v)
	}
	return datadogWidgetTime
}
func buildTerraformWidgetTime(datadogWidgetTime datadog.WidgetTime) map[string]string {
	terraformWidgetTime := map[string]string{}
	if datadogWidgetTime.LiveSpan != nil {
		terraformWidgetTime["live_span"] = *datadogWidgetTime.LiveSpan
	}
	return terraformWidgetTime
}

// Widget Marker helpers
func getWidgetMarkerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"value": {
			Type:     schema.TypeString,
			Required: true,
		},
		"display_type": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"label": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogWidgetMarkers(terraformWidgetMarkers *[]interface{}) *[]datadog.WidgetMarker {
	datadogWidgetMarkers := make([]datadog.WidgetMarker, len(*terraformWidgetMarkers))
	for i, _marker := range *terraformWidgetMarkers {
		terraformMarker := _marker.(map[string]interface{})
		// Required
		datadogMarker := datadog.WidgetMarker{
			Value: datadog.String(terraformMarker["value"].(string)),
		}
		// Optional
		if v, ok := terraformMarker["display_type"].(string); ok && len(v) != 0 {
			datadogMarker.DisplayType = datadog.String(v)
		}
		if v, ok := terraformMarker["label"].(string); ok && len(v) != 0 {
			datadogMarker.Label = datadog.String(v)
		}
		datadogWidgetMarkers[i] = datadogMarker
	}
	return &datadogWidgetMarkers
}
func buildTerraformWidgetMarkers(datadogWidgetMarkers *[]datadog.WidgetMarker) *[]map[string]string {
	terraformWidgetMarkers := make([]map[string]string, len(*datadogWidgetMarkers))
	for i, datadogMarker := range *datadogWidgetMarkers {
		terraformMarker := map[string]string{}
		// Required params
		terraformMarker["value"] = *datadogMarker.Value
		// Optional params
		if datadogMarker.DisplayType != nil {
			terraformMarker["display_type"] = *datadogMarker.DisplayType
		}
		if datadogMarker.Label != nil {
			terraformMarker["label"] = *datadogMarker.Label
		}
		terraformWidgetMarkers[i] = terraformMarker
	}
	return &terraformWidgetMarkers
}

//
// Widget Query helpers
//

// Metric Query
func getMetricQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
	}
}

// APM or Log Query
func getApmOrLogQuerySchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"index": {
					Type:     schema.TypeString,
					Required: true,
				},
				"compute": &schema.Schema{
					Type:     schema.TypeMap,
					Required: true,
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
								Type:     schema.TypeInt,
								Optional: true,
							},
						},
					},
				},
				"search": &schema.Schema{
					Type:     schema.TypeMap,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"query": {
								Type:     schema.TypeString,
								Required: true,
							},
						},
					},
				},
				"group_by": &schema.Schema{
					Type:     schema.TypeList,
					Optional: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"facet": {
								Type:     schema.TypeString,
								Optional: true,
							},
							"limit": {
								Type:     schema.TypeInt,
								Optional: true,
							},
							"sort": {
								Type:     schema.TypeMap,
								Optional: true,
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
}
func buildDatadogApmOrLogQuery(terraformQuery map[string]interface{}) *datadog.WidgetApmOrLogQuery {
	// Index
	datadogQuery := datadog.WidgetApmOrLogQuery{
		Index: datadog.String(terraformQuery["index"].(string)),
	}
	// Compute
	terraformCompute := terraformQuery["compute"].(map[string]interface{})
	datadogCompute := datadog.ApmOrLogQueryCompute{}
	if aggr, ok := terraformCompute["aggregation"].(string); ok && len(aggr) != 0 {
		datadogCompute.Aggregation = datadog.String(aggr)
	}
	if facet, ok := terraformCompute["facet"].(string); ok && len(facet) != 0 {
		datadogCompute.Facet = datadog.String(facet)
	}
	if interval, ok := terraformCompute["interval"].(string); ok {
		if v, err := strconv.ParseInt(interval, 10, 64); err == nil {
			datadogCompute.Interval = datadog.Int(int(v))
		}
	}
	datadogQuery.Compute = &datadogCompute
	// Search
	if terraformSearch, ok := terraformQuery["search"].(map[string]interface{}); ok && len(terraformSearch) > 0 {
		datadogQuery.Search = &datadog.ApmOrLogQuerySearch{
			Query: datadog.String(terraformSearch["query"].(string)),
		}
	}
	// GroupBy
	if terraformGroupBys, ok := terraformQuery["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]datadog.ApmOrLogQueryGroupBy, len(terraformGroupBys))
		for i, _groupBy := range terraformGroupBys {
			groupBy := _groupBy.(map[string]interface{})
			// Facet
			datadogGroupBy := datadog.ApmOrLogQueryGroupBy{
				Facet: datadog.String(groupBy["facet"].(string)),
			}
			// Limit
			if v, ok := groupBy["limit"].(int); ok && v != 0 {
				datadogGroupBy.Limit = &v
			}
			// Sort
			if sort, ok := groupBy["sort"].(map[string]interface{}); ok && len(sort) > 0 {

				datadogGroupBy.Sort = &datadog.ApmOrLogQueryGroupBySort{}
				if aggr, ok := sort["aggregation"].(string); ok && len(aggr) > 0 {
					datadogGroupBy.Sort.Aggregation = datadog.String(aggr)
				}
				if order, ok := sort["order"].(string); ok && len(order) > 0 {
					datadogGroupBy.Sort.Order = datadog.String(order)
				}
				if facet, ok := sort["facet"].(string); ok && len(facet) > 0 {
					datadogGroupBy.Sort.Facet = datadog.String(facet)
				}
			}
			datadogGroupBys[i] = datadogGroupBy
		}
		datadogQuery.GroupBy = datadogGroupBys
	}
	return &datadogQuery
}
func buildTerraformApmOrLogQuery(datadogQuery datadog.WidgetApmOrLogQuery) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	// Index
	terraformQuery["index"] = *datadogQuery.Index
	// Compute
	terraformCompute := map[string]interface{}{
		"aggregation": *datadogQuery.Compute.Aggregation,
	}
	if datadogQuery.Compute.Facet != nil {
		terraformCompute["facet"] = *datadogQuery.Compute.Facet
	}
	if datadogQuery.Compute.Interval != nil {
		terraformCompute["interval"] = strconv.FormatInt(int64(*datadogQuery.Compute.Interval), 10)
	}
	terraformQuery["compute"] = terraformCompute
	// Search
	if datadogQuery.Search != nil {
		terraformQuery["search"] = map[string]interface{}{
			"query": *datadogQuery.Search.Query,
		}
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
				sort := map[string]string{
					"aggregation": *groupBy.Sort.Aggregation,
					"order":       *groupBy.Sort.Order,
				}
				if groupBy.Sort.Facet != nil {
					sort["facet"] = *groupBy.Sort.Facet
				}
				terraformGroupBy["sort"] = sort
			}

			terraformGroupBys[i] = terraformGroupBy
		}
		terraformQuery["group_by"] = &terraformGroupBys
	}
	return terraformQuery
}

// Process Query
func getProcessQuerySchema() *schema.Schema {
	return &schema.Schema{
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
}
func buildDatadogProcessQuery(terraformQuery map[string]interface{}) *datadog.WidgetProcessQuery {
	datadogQuery := datadog.WidgetProcessQuery{}
	if v, ok := terraformQuery["metric"].(string); ok && len(v) != 0 {
		datadogQuery.SetMetric(v)
	}
	if v, ok := terraformQuery["search_by"].(string); ok && len(v) != 0 {
		datadogQuery.SetSearchBy(v)
	}

	if terraformFilterBys, ok := terraformQuery["filter_by"].([]interface{}); ok && len(terraformFilterBys) > 0 {
		datadogFilterbys := make([]string, len(terraformFilterBys))
		for i, filtrBy := range terraformFilterBys {
			datadogFilterbys[i] = filtrBy.(string)
		}
		datadogQuery.FilterBy = datadogFilterbys
	}

	if v, ok := terraformQuery["limit"].(int); ok && v != 0 {
		datadogQuery.SetLimit(v)
	}

	return &datadogQuery
}

func buildTerraformProcessQuery(datadogQuery datadog.WidgetProcessQuery) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	if datadogQuery.Metric != nil {
		terraformQuery["metric"] = *datadogQuery.Metric
	}
	if datadogQuery.SearchBy != nil {
		terraformQuery["search_by"] = *datadogQuery.SearchBy
	}
	if datadogQuery.FilterBy != nil {
		terraformFilterBys := make([]string, len(datadogQuery.FilterBy))
		for i, datadogFilterBy := range datadogQuery.FilterBy {
			terraformFilterBys[i] = datadogFilterBy
		}
		terraformQuery["filter_by"] = terraformFilterBys
	}
	if datadogQuery.Limit != nil {
		terraformQuery["limit"] = *datadogQuery.Limit
	}

	return terraformQuery
}

// Widget Axis helpers

func getWidgetAxisSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"label": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"scale": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"min": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"max": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"include_zero": {
			Type:     schema.TypeBool,
			Optional: true,
		},
	}
}
func buildDatadogWidgetAxis(terraformWidgetAxis map[string]interface{}) *datadog.WidgetAxis {
	datadogWidgetAxis := &datadog.WidgetAxis{}
	if v, ok := terraformWidgetAxis["label"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetLabel(v)
	}
	if v, ok := terraformWidgetAxis["scale"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetScale(v)
	}
	if v, ok := terraformWidgetAxis["min"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetMin(v)
	}
	if v, ok := terraformWidgetAxis["max"].(string); ok && len(v) != 0 {
		datadogWidgetAxis.SetMax(v)
	}
	if v, ok := terraformWidgetAxis["include_zero"].(bool); ok {
		datadogWidgetAxis.SetIncludeZero(v)
	}
	return datadogWidgetAxis
}
func buildTerraformWidgetAxis(datadogWidgetAxis datadog.WidgetAxis) map[string]interface{} {
	terraformWidgetAxis := map[string]interface{}{}
	if datadogWidgetAxis.Label != nil {
		terraformWidgetAxis["label"] = *datadogWidgetAxis.Label
	}
	if datadogWidgetAxis.Scale != nil {
		terraformWidgetAxis["scale"] = *datadogWidgetAxis.Scale
	}
	if datadogWidgetAxis.Min != nil {
		terraformWidgetAxis["min"] = *datadogWidgetAxis.Min
	}
	if datadogWidgetAxis.Max != nil {
		terraformWidgetAxis["max"] = *datadogWidgetAxis.Max
	}
	if datadogWidgetAxis.IncludeZero != nil {
		terraformWidgetAxis["include_zero"] = *datadogWidgetAxis.IncludeZero
	}
	return terraformWidgetAxis
}

// Widget Style helpers

func getWidgetRequestStyle() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"palette": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogWidgetRequestStyle(terraformStyle map[string]interface{}) *datadog.WidgetRequestStyle {
	datadogStyle := &datadog.WidgetRequestStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}

	return datadogStyle
}
func buildTerraformWidgetRequestStyle(datadogStyle datadog.WidgetRequestStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if datadogStyle.Palette != nil {
		terraformStyle["palette"] = *datadogStyle.Palette
	}
	return terraformStyle
}

// Timeseriest Style helpers

func buildDatadogTimeseriesRequestStyle(terraformStyle map[string]interface{}) *datadog.TimeseriesRequestStyle {
	datadogStyle := &datadog.TimeseriesRequestStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}
	if v, ok := terraformStyle["line_type"].(string); ok && len(v) != 0 {
		datadogStyle.SetLineType(v)
	}
	if v, ok := terraformStyle["line_width"].(string); ok && len(v) != 0 {
		datadogStyle.SetLineWidth(v)
	}

	return datadogStyle
}
func buildTerraformTimeseriesRequestStyle(datadogStyle datadog.TimeseriesRequestStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if datadogStyle.Palette != nil {
		terraformStyle["palette"] = *datadogStyle.Palette
	}
	if datadogStyle.LineType != nil {
		terraformStyle["line_type"] = *datadogStyle.LineType
	}
	if datadogStyle.LineWidth != nil {
		terraformStyle["line_width"] = *datadogStyle.LineWidth
	}
	return terraformStyle
}

// Hostmap Style helpers

func buildDatadogHostmapRequestStyle(terraformStyle map[string]interface{}) *datadog.HostmapStyle {
	datadogStyle := &datadog.HostmapStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}
	if v, ok := terraformStyle["palette_flip"].(bool); ok {
		datadogStyle.SetPaletteFlip(v)
	}
	if v, ok := terraformStyle["fill_min"].(string); ok && len(v) != 0 {
		datadogStyle.SetFillMin(v)
	}
	if v, ok := terraformStyle["fill_max"].(string); ok && len(v) != 0 {
		datadogStyle.SetFillMax(v)
	}

	return datadogStyle
}
func buildTerraformHostmapRequestStyle(datadogStyle datadog.HostmapStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if datadogStyle.Palette != nil {
		terraformStyle["palette"] = *datadogStyle.Palette
	}
	if datadogStyle.PaletteFlip != nil {
		terraformStyle["palette_flip"] = *datadogStyle.PaletteFlip
	}
	if datadogStyle.FillMin != nil {
		terraformStyle["fill_min"] = *datadogStyle.FillMin
	}
	if datadogStyle.FillMax != nil {
		terraformStyle["fill_max"] = *datadogStyle.FillMax
	}
	return terraformStyle
}

// Schema validation
func validateDashboardLayoutType(val interface{}, key string) (warns []string, errs []error) {
	value := val.(string)
	switch value {
	case "free", "ordered":
		break
	default:
		errs = append(errs, fmt.Errorf(
			"%q contains an invalid value %q. Valid values are `free` or `ordered`", key, value))
	}
	return
}
func validateGroupWidgetLayoutType(val interface{}, key string) (warns []string, errs []error) {
	value := val.(string)
	switch value {
	case "ordered":
		break
	default:
		errs = append(errs, fmt.Errorf(
			"%q contains an invalid value %q. Only `ordered` is a valid value", key, value))
	}
	return
}
