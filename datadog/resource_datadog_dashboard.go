package datadog

import (
	"fmt"
	"strconv"
	"strings"

	datadogV1 "github.com/DataDog/datadog-api-client-go/api/v1/datadog"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
			"url": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The URL of the dashboard.",
			},
			"template_variable": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of template variables for this dashboard.",
				Elem: &schema.Resource{
					Schema: getTemplateVariableSchema(),
				},
			},
			"template_variable_preset": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The list of selectable template variable presets for this dashboard.",
				Elem: &schema.Resource{
					Schema: getTemplateVariablePresetSchema(),
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
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	dashboardPayload, err := buildDatadogDashboard(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	dashboard, _, err := datadogClientV1.DashboardsApi.CreateDashboard(authV1).Body(*dashboardPayload).Execute()
	if err != nil {
		return translateClientError(err, "error creating dashboard")
	}
	d.SetId(*dashboard.Id)
	return resourceDatadogDashboardRead(d, meta)
}

func resourceDatadogDashboardUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	id := d.Id()
	dashboard, err := buildDatadogDashboard(d)
	if err != nil {
		return fmt.Errorf("failed to parse resource configuration: %s", err.Error())
	}
	if _, _, err = datadogClientV1.DashboardsApi.UpdateDashboard(authV1, id).Body(*dashboard).Execute(); err != nil {
		return translateClientError(err, "error updating dashboard")
	}
	return resourceDatadogDashboardRead(d, meta)
}

func resourceDatadogDashboardRead(d *schema.ResourceData, meta interface{}) error {
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	id := d.Id()
	dashboard, _, err := datadogClientV1.DashboardsApi.GetDashboard(authV1, id).Execute()
	if err != nil {
		return translateClientError(err, "error getting dashboard")
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
	if err = d.Set("url", dashboard.GetUrl()); err != nil {
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

	// Set template variable presets
	templateVariablePresets := buildTerraformTemplateVariablePresets(&dashboard.TemplateVariablePresets)
	if err := d.Set("template_variable_preset", templateVariablePresets); err != nil {
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
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	id := d.Id()
	if _, _, err := datadogClientV1.DashboardsApi.DeleteDashboard(authV1, id).Execute(); err != nil {
		return translateClientError(err, "error deleting dashboard")
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
	providerConf := meta.(*ProviderConfiguration)
	datadogClientV1 := providerConf.DatadogClientV1
	authV1 := providerConf.AuthV1
	id := d.Id()
	if _, _, err := datadogClientV1.DashboardsApi.GetDashboard(authV1, id).Execute(); err != nil {
		if strings.Contains(err.Error(), "404 Not Found") {
			return false, nil
		}
		return false, translateClientError(err, "error checking dashboard exists")
	}
	return true, nil
}

func buildDatadogDashboard(d *schema.ResourceData) (*datadogV1.Dashboard, error) {
	var dashboard datadogV1.Dashboard

	dashboard.SetId(d.Id())

	if v, ok := d.GetOk("title"); ok {
		dashboard.SetTitle(v.(string))
	}
	if v, ok := d.GetOk("layout_type"); ok {
		dashboard.SetLayoutType(datadogV1.DashboardLayoutType(v.(string)))
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
	dashboard.SetWidgets(*datadogWidgets)

	// Build NotifyList
	notifyList := d.Get("notify_list").([]interface{})
	dashboard.NotifyList = *buildDatadogNotifyList(&notifyList)

	// Build TemplateVariables
	templateVariables := d.Get("template_variable").([]interface{})
	dashboard.TemplateVariables = *buildDatadogTemplateVariables(&templateVariables)

	// Build TemplateVariablePresets
	templateVariablePresets := d.Get("template_variable_preset").([]interface{})
	dashboard.TemplateVariablePresets = *buildDatadogTemplateVariablePresets(&templateVariablePresets)

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

func buildDatadogTemplateVariables(terraformTemplateVariables *[]interface{}) *[]datadogV1.DashboardTemplateVariables {
	datadogTemplateVariables := make([]datadogV1.DashboardTemplateVariables, len(*terraformTemplateVariables))
	for i, ttv := range *terraformTemplateVariables {
		terraformTemplateVariable := ttv.(map[string]interface{})
		var datadogTemplateVariable datadogV1.DashboardTemplateVariables
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

func buildTerraformTemplateVariables(datadogTemplateVariables *[]datadogV1.DashboardTemplateVariables) *[]map[string]string {
	terraformTemplateVariables := make([]map[string]string, len(*datadogTemplateVariables))
	for i, templateVariable := range *datadogTemplateVariables {
		terraformTemplateVariable := map[string]string{}
		if v, ok := templateVariable.GetNameOk(); ok {
			terraformTemplateVariable["name"] = *v
		}
		if v, ok := templateVariable.GetPrefixOk(); ok {
			terraformTemplateVariable["prefix"] = *v
		}
		if v, ok := templateVariable.GetDefaultOk(); ok {
			terraformTemplateVariable["default"] = *v
		}
		terraformTemplateVariables[i] = terraformTemplateVariable
	}
	return &terraformTemplateVariables
}

//
// Template Variable Preset Helpers
//

func getTemplateVariablePresetSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the preset.",
		},
		"template_variable": {
			Type:        schema.TypeList,
			Required:    true,
			Description: "The template variable names and assumed values under the given preset",
			Elem: &schema.Resource{
				Schema: getTemplateVariablePresetValueSchema(),
			},
		},
	}
}

func getTemplateVariablePresetValueSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the template variable",
			Required:    true,
		},
		"value": {
			Type:        schema.TypeString,
			Description: "The value that should be assumed by the template variable in this preset",
			Required:    true,
		},
	}
}

func buildDatadogTemplateVariablePresets(terraformTemplateVariablePresets *[]interface{}) *[]datadogV1.DashboardTemplateVariablePreset {
	datadogTemplateVariablePresets := make([]datadogV1.DashboardTemplateVariablePreset, len(*terraformTemplateVariablePresets))

	for i, tvp := range *terraformTemplateVariablePresets {
		templateVariablePreset := tvp.(map[string]interface{})
		var datadogTemplateVariablePreset datadogV1.DashboardTemplateVariablePreset

		if v, ok := templateVariablePreset["name"].(string); ok && len(v) != 0 {
			datadogTemplateVariablePreset.SetName(v)
		}

		if templateVariablePresetValues, ok := templateVariablePreset["template_variable"].([]interface{}); ok && len(templateVariablePresetValues) != 0 {
			datadogTemplateVariablePresetValues := make([]datadogV1.DashboardTemplateVariablePresetValue, len(templateVariablePresetValues))

			for j, tvp := range templateVariablePresetValues {
				templateVariablePresetValue := tvp.(map[string]interface{})
				var datadogTemplateVariablePresetValue datadogV1.DashboardTemplateVariablePresetValue

				if w, ok := templateVariablePresetValue["name"].(string); ok && len(w) != 0 {
					datadogTemplateVariablePresetValue.SetName(w)
				}

				if w, ok := templateVariablePresetValue["value"].(string); ok && len(w) != 0 {
					datadogTemplateVariablePresetValue.SetValue(w)
				}

				datadogTemplateVariablePresetValues[j] = datadogTemplateVariablePresetValue
			}

			datadogTemplateVariablePreset.SetTemplateVariables(datadogTemplateVariablePresetValues)
		}

		datadogTemplateVariablePresets[i] = datadogTemplateVariablePreset
	}

	return &datadogTemplateVariablePresets
}

func buildTerraformTemplateVariablePresets(datadogTemplateVariablePresets *[]datadogV1.DashboardTemplateVariablePreset) *[]map[string]interface{} {
	// Allocate final resting place for tf/hash version
	terraformTemplateVariablePresets := make([]map[string]interface{}, len(*datadogTemplateVariablePresets))

	//iterate over preset objects
	for i, templateVariablePreset := range *datadogTemplateVariablePresets {
		// Allocate for this preset group, a map of string key to obj (string for name, array for preset values
		terraformTemplateVariablePreset := make(map[string]interface{})
		if v, ok := templateVariablePreset.GetNameOk(); ok {
			terraformTemplateVariablePreset["name"] = v
		}

		// allocate for array of preset values (names = name,value, values = name, template variable)

		terraformTemplateVariablePresetValues := make([]map[string]string, len(templateVariablePreset.GetTemplateVariables()))
		for j, templateVariablePresetValue := range templateVariablePreset.GetTemplateVariables() {
			// allocate map for name => name value => value
			terraformTemplateVariablePresetValue := make(map[string]string)
			if v, ok := templateVariablePresetValue.GetNameOk(); ok {
				terraformTemplateVariablePresetValue["name"] = *v
			}
			if v, ok := templateVariablePresetValue.GetValueOk(); ok {
				terraformTemplateVariablePresetValue["value"] = *v
			}

			terraformTemplateVariablePresetValues[j] = terraformTemplateVariablePresetValue
		}

		// Set template_variable to the array of values we just created
		terraformTemplateVariablePreset["template_variable"] = terraformTemplateVariablePresetValues

		// put the preset group into the output var
		terraformTemplateVariablePresets[i] = terraformTemplateVariablePreset
	}

	return &terraformTemplateVariablePresets
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
		"query_table_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Query Table widget",
			Elem: &schema.Resource{
				Schema: getQueryTableDefinitionSchema(),
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
		"servicemap_definition": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "The definition for a Service Map widget",
			Elem: &schema.Resource{
				Schema: getServiceMapDefinitionSchema(),
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

func buildDatadogWidgets(terraformWidgets *[]interface{}) (*[]datadogV1.Widget, error) {
	datadogWidgets := make([]datadogV1.Widget, len(*terraformWidgets))
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
func buildDatadogWidget(terraformWidget map[string]interface{}) (*datadogV1.Widget, error) {
	// Build widget Definition
	var definition datadogV1.WidgetDefinition
	if def, ok := terraformWidget["group_definition"].([]interface{}); ok && len(def) > 0 {
		if groupDefinition, ok := def[0].(map[string]interface{}); ok {
			datadogDefinition, err := buildDatadogGroupDefinition(groupDefinition)
			if err != nil {
				return nil, err
			}
			definition = datadogV1.GroupWidgetDefinitionAsWidgetDefinition(datadogDefinition)
		}
	} else if def, ok := terraformWidget["alert_graph_definition"].([]interface{}); ok && len(def) > 0 {
		if alertGraphDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.AlertGraphWidgetDefinitionAsWidgetDefinition(buildDatadogAlertGraphDefinition(alertGraphDefinition))
		}
	} else if def, ok := terraformWidget["alert_value_definition"].([]interface{}); ok && len(def) > 0 {
		if alertValueDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.AlertValueWidgetDefinitionAsWidgetDefinition(buildDatadogAlertValueDefinition(alertValueDefinition))
		}
	} else if def, ok := terraformWidget["change_definition"].([]interface{}); ok && len(def) > 0 {
		if changeDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ChangeWidgetDefinitionAsWidgetDefinition(buildDatadogChangeDefinition(changeDefinition))
		}
	} else if def, ok := terraformWidget["check_status_definition"].([]interface{}); ok && len(def) > 0 {
		if checkStatusDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.CheckStatusWidgetDefinitionAsWidgetDefinition(buildDatadogCheckStatusDefinition(checkStatusDefinition))
		}
	} else if def, ok := terraformWidget["distribution_definition"].([]interface{}); ok && len(def) > 0 {
		if distributionDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.DistributionWidgetDefinitionAsWidgetDefinition(buildDatadogDistributionDefinition(distributionDefinition))
		}
	} else if def, ok := terraformWidget["event_stream_definition"].([]interface{}); ok && len(def) > 0 {
		if eventStreamDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.EventStreamWidgetDefinitionAsWidgetDefinition(buildDatadogEventStreamDefinition(eventStreamDefinition))
		}
	} else if def, ok := terraformWidget["event_timeline_definition"].([]interface{}); ok && len(def) > 0 {
		if eventTimelineDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.EventTimelineWidgetDefinitionAsWidgetDefinition(buildDatadogEventTimelineDefinition(eventTimelineDefinition))
		}
	} else if def, ok := terraformWidget["free_text_definition"].([]interface{}); ok && len(def) > 0 {
		if freeTextDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.FreeTextWidgetDefinitionAsWidgetDefinition(buildDatadogFreeTextDefinition(freeTextDefinition))
		}
	} else if def, ok := terraformWidget["heatmap_definition"].([]interface{}); ok && len(def) > 0 {
		if heatmapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.HeatMapWidgetDefinitionAsWidgetDefinition(buildDatadogHeatmapDefinition(heatmapDefinition))
		}
	} else if def, ok := terraformWidget["hostmap_definition"].([]interface{}); ok && len(def) > 0 {
		if hostDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.HostMapWidgetDefinitionAsWidgetDefinition(buildDatadogHostmapDefinition(hostDefinition))
		}
	} else if def, ok := terraformWidget["iframe_definition"].([]interface{}); ok && len(def) > 0 {
		if iframeDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.IFrameWidgetDefinitionAsWidgetDefinition(buildDatadogIframeDefinition(iframeDefinition))
		}
	} else if def, ok := terraformWidget["image_definition"].([]interface{}); ok && len(def) > 0 {
		if imageDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ImageWidgetDefinitionAsWidgetDefinition(buildDatadogImageDefinition(imageDefinition))
		}
	} else if def, ok := terraformWidget["log_stream_definition"].([]interface{}); ok && len(def) > 0 {
		if logStreamDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.LogStreamWidgetDefinitionAsWidgetDefinition(buildDatadogLogStreamDefinition(logStreamDefinition))
		}
	} else if def, ok := terraformWidget["manage_status_definition"].([]interface{}); ok && len(def) > 0 {
		if manageStatusDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.MonitorSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogManageStatusDefinition(manageStatusDefinition))
		}
	} else if def, ok := terraformWidget["note_definition"].([]interface{}); ok && len(def) > 0 {
		if noteDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.NoteWidgetDefinitionAsWidgetDefinition(buildDatadogNoteDefinition(noteDefinition))
		}
	} else if def, ok := terraformWidget["query_value_definition"].([]interface{}); ok && len(def) > 0 {
		if queryValueDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.QueryValueWidgetDefinitionAsWidgetDefinition(buildDatadogQueryValueDefinition(queryValueDefinition))
		}
	} else if def, ok := terraformWidget["query_table_definition"].([]interface{}); ok && len(def) > 0 {
		if queryTableDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TableWidgetDefinitionAsWidgetDefinition(buildDatadogQueryTableDefinition(queryTableDefinition))
		}
	} else if def, ok := terraformWidget["scatterplot_definition"].([]interface{}); ok && len(def) > 0 {
		if scatterplotDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ScatterPlotWidgetDefinitionAsWidgetDefinition(buildDatadogScatterplotDefinition(scatterplotDefinition))
		}
	} else if def, ok := terraformWidget["servicemap_definition"].([]interface{}); ok && len(def) > 0 {
		if serviceMapDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ServiceMapWidgetDefinitionAsWidgetDefinition(buildDatadogServiceMapDefinition(serviceMapDefinition))
		}
	} else if def, ok := terraformWidget["service_level_objective_definition"].([]interface{}); ok && len(def) > 0 {
		if serviceLevelObjectiveDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.SLOWidgetDefinitionAsWidgetDefinition(buildDatadogServiceLevelObjectiveDefinition(serviceLevelObjectiveDefinition))
		}
	} else if def, ok := terraformWidget["timeseries_definition"].([]interface{}); ok && len(def) > 0 {
		if timeseriesDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.TimeseriesWidgetDefinitionAsWidgetDefinition(buildDatadogTimeseriesDefinition(timeseriesDefinition))
		}
	} else if def, ok := terraformWidget["toplist_definition"].([]interface{}); ok && len(def) > 0 {
		if toplistDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ToplistWidgetDefinitionAsWidgetDefinition(buildDatadogToplistDefinition(toplistDefinition))
		}
	} else if def, ok := terraformWidget["trace_service_definition"].([]interface{}); ok && len(def) > 0 {
		if traceServiceDefinition, ok := def[0].(map[string]interface{}); ok {
			definition = datadogV1.ServiceSummaryWidgetDefinitionAsWidgetDefinition(buildDatadogTraceServiceDefinition(traceServiceDefinition))
		}
	} else {
		return nil, fmt.Errorf("failed to find valid definition in widget configuration")
	}

	datadogWidget := datadogV1.NewWidget(definition)

	// Build widget layout
	if v, ok := terraformWidget["layout"].(map[string]interface{}); ok && len(v) != 0 {
		datadogWidget.SetLayout(*buildDatadogWidgetLayout(v))
	}

	return datadogWidget, nil
}

// Helper to build a list of Terraform widgets from a list of Datadog widgets
func buildTerraformWidgets(datadogWidgets *[]datadogV1.Widget) (*[]map[string]interface{}, error) {

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
func buildTerraformWidget(datadogWidget datadogV1.Widget) (map[string]interface{}, error) {
	terraformWidget := map[string]interface{}{}

	// Build layout
	if v, ok := datadogWidget.GetLayoutOk(); ok {
		terraformWidget["layout"] = buildTerraformWidgetLayout(*v)
	}

	// Build definition
	widgetDefinition := datadogWidget.GetDefinition()
	if widgetDefinition.GroupWidgetDefinition != nil {
		terraformDefinition := buildTerraformGroupDefinition(*widgetDefinition.GroupWidgetDefinition)
		terraformWidget["group_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.AlertGraphWidgetDefinition != nil {
		terraformDefinition := buildTerraformAlertGraphDefinition(*widgetDefinition.AlertGraphWidgetDefinition)
		terraformWidget["alert_graph_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.AlertValueWidgetDefinition != nil {
		terraformDefinition := buildTerraformAlertValueDefinition(*widgetDefinition.AlertValueWidgetDefinition)
		terraformWidget["alert_value_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ChangeWidgetDefinition != nil {
		terraformDefinition := buildTerraformChangeDefinition(*widgetDefinition.ChangeWidgetDefinition)
		terraformWidget["change_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.CheckStatusWidgetDefinition != nil {
		terraformDefinition := buildTerraformCheckStatusDefinition(*widgetDefinition.CheckStatusWidgetDefinition)
		terraformWidget["check_status_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.DistributionWidgetDefinition != nil {
		terraformDefinition := buildTerraformDistributionDefinition(*widgetDefinition.DistributionWidgetDefinition)
		terraformWidget["distribution_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.EventStreamWidgetDefinition != nil {
		terraformDefinition := buildTerraformEventStreamDefinition(*widgetDefinition.EventStreamWidgetDefinition)
		terraformWidget["event_stream_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.EventTimelineWidgetDefinition != nil {
		terraformDefinition := buildTerraformEventTimelineDefinition(*widgetDefinition.EventTimelineWidgetDefinition)
		terraformWidget["event_timeline_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.FreeTextWidgetDefinition != nil {
		terraformDefinition := buildTerraformFreeTextDefinition(*widgetDefinition.FreeTextWidgetDefinition)
		terraformWidget["free_text_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.HeatMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformHeatmapDefinition(*widgetDefinition.HeatMapWidgetDefinition)
		terraformWidget["heatmap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.HostMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformHostmapDefinition(*widgetDefinition.HostMapWidgetDefinition)
		terraformWidget["hostmap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.IFrameWidgetDefinition != nil {
		terraformDefinition := buildTerraformIframeDefinition(*widgetDefinition.IFrameWidgetDefinition)
		terraformWidget["iframe_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ImageWidgetDefinition != nil {
		terraformDefinition := buildTerraformImageDefinition(*widgetDefinition.ImageWidgetDefinition)
		terraformWidget["image_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.LogStreamWidgetDefinition != nil {
		terraformDefinition := buildTerraformLogStreamDefinition(*widgetDefinition.LogStreamWidgetDefinition)
		terraformWidget["log_stream_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.MonitorSummaryWidgetDefinition != nil {
		terraformDefinition := buildTerraformManageStatusDefinition(*widgetDefinition.MonitorSummaryWidgetDefinition)
		terraformWidget["manage_status_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.NoteWidgetDefinition != nil {
		terraformDefinition := buildTerraformNoteDefinition(*widgetDefinition.NoteWidgetDefinition)
		terraformWidget["note_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.QueryValueWidgetDefinition != nil {
		terraformDefinition := buildTerraformQueryValueDefinition(*widgetDefinition.QueryValueWidgetDefinition)
		terraformWidget["query_value_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.TableWidgetDefinition != nil {
		terraformDefinition := buildTerraformQueryTableDefinition(*widgetDefinition.TableWidgetDefinition)
		terraformWidget["query_table_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ScatterPlotWidgetDefinition != nil {
		terraformDefinition := buildTerraformScatterplotDefinition(*widgetDefinition.ScatterPlotWidgetDefinition)
		terraformWidget["scatterplot_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ServiceMapWidgetDefinition != nil {
		terraformDefinition := buildTerraformServiceMapDefinition(*widgetDefinition.ServiceMapWidgetDefinition)
		terraformWidget["servicemap_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.SLOWidgetDefinition != nil {
		terraformDefinition := buildTerraformServiceLevelObjectiveDefinition(*widgetDefinition.SLOWidgetDefinition)
		terraformWidget["service_level_objective_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.TimeseriesWidgetDefinition != nil {
		terraformDefinition := buildTerraformTimeseriesDefinition(*widgetDefinition.TimeseriesWidgetDefinition)
		terraformWidget["timeseries_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ToplistWidgetDefinition != nil {
		terraformDefinition := buildTerraformToplistDefinition(*widgetDefinition.ToplistWidgetDefinition)
		terraformWidget["toplist_definition"] = []map[string]interface{}{terraformDefinition}
	} else if widgetDefinition.ServiceSummaryWidgetDefinition != nil {
		terraformDefinition := buildTerraformTraceServiceDefinition(*widgetDefinition.ServiceSummaryWidgetDefinition)
		terraformWidget["trace_service_definition"] = []map[string]interface{}{terraformDefinition}
	} else {
		return nil, fmt.Errorf("unsupported widget type: %s", widgetDefinition.GetActualInstance())
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

func buildDatadogWidgetLayout(terraformLayout map[string]interface{}) *datadogV1.WidgetLayout {
	datadogLayout := datadogV1.NewWidgetLayoutWithDefaults()

	if value, ok := terraformLayout["x"].(string); ok && len(value) != 0 {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			datadogLayout.SetX(v)
		}
	}
	if value, ok := terraformLayout["y"].(string); ok && len(value) != 0 {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			datadogLayout.SetY(v)
		}
	}
	if value, ok := terraformLayout["height"].(string); ok && len(value) != 0 {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			datadogLayout.SetHeight(v)
		}
	}
	if value, ok := terraformLayout["width"].(string); ok && len(value) != 0 {
		if v, err := strconv.ParseInt(value, 10, 64); err == nil {
			datadogLayout.SetWidth(v)
		}
	}
	return datadogLayout
}

func buildTerraformWidgetLayout(datadogLayout datadogV1.WidgetLayout) map[string]string {
	terraformLayout := map[string]string{}

	if v, ok := datadogLayout.GetXOk(); ok {
		terraformLayout["x"] = strconv.FormatInt(*v, 10)
	}
	if v, ok := datadogLayout.GetYOk(); ok {
		terraformLayout["y"] = strconv.FormatInt(*v, 10)
	}
	if v, ok := datadogLayout.GetHeightOk(); ok {
		terraformLayout["height"] = strconv.FormatInt(*v, 10)
	}
	if v, ok := datadogLayout.GetWidthOk(); ok {
		terraformLayout["width"] = strconv.FormatInt(*v, 10)
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

func buildDatadogGroupDefinition(terraformGroupDefinition map[string]interface{}) (*datadogV1.GroupWidgetDefinition, error) {
	datadogGroupDefinition := datadogV1.NewGroupWidgetDefinitionWithDefaults()

	if v, ok := terraformGroupDefinition["widget"].([]interface{}); ok && len(v) != 0 {
		datadogWidgets, err := buildDatadogWidgets(&v)
		if err != nil {
			return nil, err
		}
		datadogGroupDefinition.SetWidgets(*datadogWidgets)
	}
	if v, ok := terraformGroupDefinition["layout_type"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetLayoutType(datadogV1.WidgetLayoutType(v))
	}
	if v, ok := terraformGroupDefinition["title"].(string); ok && len(v) != 0 {
		datadogGroupDefinition.SetTitle(v)
	}

	return datadogGroupDefinition, nil
}

func buildTerraformGroupDefinition(datadogGroupDefinition datadogV1.GroupWidgetDefinition) map[string]interface{} {
	terraformGroupDefinition := map[string]interface{}{}

	var groupWidgets []map[string]interface{}
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

func buildDatadogAlertGraphDefinition(terraformDefinition map[string]interface{}) *datadogV1.AlertGraphWidgetDefinition {
	datadogDefinition := datadogV1.NewAlertGraphWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.AlertId = terraformDefinition["alert_id"].(string)
	datadogDefinition.VizType = datadogV1.WidgetVizType(terraformDefinition["viz_type"].(string))
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.Title = datadogV1.PtrString(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.TitleSize = datadogV1.PtrString(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}

func buildTerraformAlertGraphDefinition(datadogDefinition datadogV1.AlertGraphWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["alert_id"] = datadogDefinition.AlertId
	terraformDefinition["viz_type"] = datadogDefinition.VizType
	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
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

func buildDatadogAlertValueDefinition(terraformDefinition map[string]interface{}) *datadogV1.AlertValueWidgetDefinition {
	datadogDefinition := datadogV1.NewAlertValueWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.AlertId = terraformDefinition["alert_id"].(string)
	// Optional params
	if v, ok := terraformDefinition["precision"].(int); ok && v != 0 {
		datadogDefinition.SetPrecision(int64(v))
	}
	if v, ok := terraformDefinition["unit"].(string); ok && len(v) != 0 {
		datadogDefinition.SetUnit(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	return datadogDefinition
}

func buildTerraformAlertValueDefinition(datadogDefinition datadogV1.AlertValueWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["alert_id"] = datadogDefinition.GetAlertId()
	// Optional params
	if v, ok := datadogDefinition.GetPrecisionOk(); ok {
		terraformDefinition["precision"] = *v
	}
	if v, ok := datadogDefinition.GetUnitOk(); ok {
		terraformDefinition["unit"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
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
func buildDatadogChangeDefinition(terraformDefinition map[string]interface{}) *datadogV1.ChangeWidgetDefinition {
	datadogDefinition := datadogV1.NewChangeWidgetDefinitionWithDefaults()
	// Required params
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
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformChangeDefinition(datadogDefinition datadogV1.ChangeWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformChangeRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func getChangeRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
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
func buildDatadogChangeRequests(terraformRequests *[]interface{}) *[]datadogV1.ChangeWidgetRequest {
	datadogRequests := make([]datadogV1.ChangeWidgetRequest, len(*terraformRequests))
	for i, request := range *terraformRequests {
		terraformRequest := request.(map[string]interface{})
		// Build ChangeRequest
		datadogChangeRequest := datadogV1.NewChangeWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogChangeRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogChangeRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogChangeRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogChangeRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogChangeRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}

		if v, ok := terraformRequest["change_type"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetChangeType(datadogV1.WidgetChangeType(v))
		}
		if v, ok := terraformRequest["compare_to"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetCompareTo(datadogV1.WidgetCompareTo(v))
		}
		if v, ok := terraformRequest["increase_good"].(bool); ok {
			datadogChangeRequest.SetIncreaseGood(v)
		}
		if v, ok := terraformRequest["order_by"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetOrderBy(datadogV1.WidgetOrderBy(v))
		}
		if v, ok := terraformRequest["order_dir"].(string); ok && len(v) != 0 {
			datadogChangeRequest.SetOrderDir(datadogV1.WidgetSort(v))
		}
		if v, ok := terraformRequest["show_present"].(bool); ok {
			datadogChangeRequest.SetShowPresent(v)
		}

		datadogRequests[i] = *datadogChangeRequest
	}
	return &datadogRequests
}
func buildTerraformChangeRequests(datadogChangeRequests *[]datadogV1.ChangeWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogChangeRequests))
	for i, datadogRequest := range *datadogChangeRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}

		if v, ok := datadogRequest.GetChangeTypeOk(); ok {
			terraformRequest["change_type"] = *v
		}
		if v, ok := datadogRequest.GetCompareToOk(); ok {
			terraformRequest["compare_to"] = *v
		}
		if v, ok := datadogRequest.GetIncreaseGoodOk(); ok {
			terraformRequest["increase_good"] = *v
		}
		if v, ok := datadogRequest.GetOrderByOk(); ok {
			terraformRequest["order_by"] = *v
		}
		if v, ok := datadogRequest.GetOrderDirOk(); ok {
			terraformRequest["order_dir"] = *v
		}
		if v, ok := datadogRequest.GetShowPresentOk(); ok {
			terraformRequest["show_present"] = *v
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
		"legend_size": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateTimeseriesWidgetLegendSize,
		},
		"show_legend": {
			Type:     schema.TypeBool,
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
func buildDatadogDistributionDefinition(terraformDefinition map[string]interface{}) *datadogV1.DistributionWidgetDefinition {
	datadogDefinition := datadogV1.NewDistributionWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogDistributionRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.SetShowLegend(v)
	}
	if v, ok := terraformDefinition["legend_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendSize(datadogV1.WidgetLegendSize(v))
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformDistributionDefinition(datadogDefinition datadogV1.DistributionWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformDistributionRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetShowLegendOk(); ok {
		terraformDefinition["show_legend"] = *v
	}
	if v, ok := datadogDefinition.GetLegendSizeOk(); ok {
		terraformDefinition["legend_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func getDistributionRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
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
func buildDatadogDistributionRequests(terraformRequests *[]interface{}) *[]datadogV1.DistributionWidgetRequest {
	datadogRequests := make([]datadogV1.DistributionWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build DistributionRequest
		datadogDistributionRequest := datadogV1.NewDistributionWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogDistributionRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogDistributionRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogDistributionRequest.Style = buildDatadogWidgetStyle(v)
			}
		}

		datadogRequests[i] = *datadogDistributionRequest
	}
	return &datadogRequests
}
func buildTerraformDistributionRequests(datadogDistributionRequests *[]datadogV1.DistributionWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogDistributionRequests))
	for i, datadogRequest := range *datadogDistributionRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}
		if datadogRequest.Style != nil {
			style := buildTerraformWidgetStyle(*datadogRequest.Style)
			terraformRequest["style"] = []map[string]interface{}{style}
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
		"tags_execution": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogEventStreamDefinition(terraformDefinition map[string]interface{}) *datadogV1.EventStreamWidgetDefinition {
	datadogDefinition := datadogV1.NewEventStreamWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetQuery(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["event_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetEventSize(datadogV1.WidgetEventSize(v))
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	if v, ok := terraformDefinition["tags_execution"].(string); ok && len(v) > 0 {
		datadogDefinition.SetTagsExecution(v)
	}
	return datadogDefinition
}

func buildTerraformEventStreamDefinition(datadogDefinition datadogV1.EventStreamWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = datadogDefinition.Query
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
	if datadogDefinition.TagsExecution != nil {
		terraformDefinition["tags_execution"] = *datadogDefinition.TagsExecution
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
		"tags_execution": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}

func buildDatadogEventTimelineDefinition(terraformDefinition map[string]interface{}) *datadogV1.EventTimelineWidgetDefinition {
	datadogDefinition := datadogV1.NewEventTimelineWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetQuery(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	if v, ok := terraformDefinition["tags_execution"].(string); ok && len(v) > 0 {
		datadogDefinition.SetTagsExecution(v)
	}
	return datadogDefinition
}

func buildTerraformEventTimelineDefinition(datadogDefinition datadogV1.EventTimelineWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = datadogDefinition.GetQuery()
	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	if v, ok := datadogDefinition.GetTagsExecutionOk(); ok {
		terraformDefinition["tags_execution"] = *v
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

func buildDatadogCheckStatusDefinition(terraformDefinition map[string]interface{}) *datadogV1.CheckStatusWidgetDefinition {
	datadogDefinition := datadogV1.NewCheckStatusWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetCheck(terraformDefinition["check"].(string))
	datadogDefinition.SetGrouping(datadogV1.WidgetGrouping(terraformDefinition["grouping"].(string)))
	// Optional params
	if v, ok := terraformDefinition["group"].(string); ok && len(v) != 0 {
		datadogDefinition.SetGroup(v)
	}
	if terraformGroupBys, ok := terraformDefinition["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]string, len(terraformGroupBys))
		for i, groupBy := range terraformGroupBys {
			datadogGroupBys[i] = groupBy.(string)
		}
		datadogDefinition.SetGroupBy(datadogGroupBys)
	}
	if terraformTags, ok := terraformDefinition["tags"].([]interface{}); ok && len(terraformTags) > 0 {
		datadogTags := make([]string, len(terraformTags))
		for i, tag := range terraformTags {
			datadogTags[i] = tag.(string)
		}
		datadogDefinition.SetTags(datadogTags)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}

func buildTerraformCheckStatusDefinition(datadogDefinition datadogV1.CheckStatusWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["check"] = datadogDefinition.GetCheck()
	terraformDefinition["grouping"] = datadogDefinition.GetGrouping()
	// Optional params
	if v, ok := datadogDefinition.GetGroupOk(); ok {
		terraformDefinition["group"] = *v
	}
	if v, ok := datadogDefinition.GetGroupByOk(); ok {
		terraformGroupBys := make([]string, len(*v))
		for i, datadogGroupBy := range *v {
			terraformGroupBys[i] = datadogGroupBy
		}
		terraformDefinition["group_by"] = terraformGroupBys
	}
	if v, ok := datadogDefinition.GetTagsOk(); ok {
		terraformTags := make([]string, len(*v))
		for i, datadogTag := range *v {
			terraformTags[i] = datadogTag
		}
		terraformDefinition["tags"] = terraformTags
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
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

func buildDatadogFreeTextDefinition(terraformDefinition map[string]interface{}) *datadogV1.FreeTextWidgetDefinition {
	datadogDefinition := datadogV1.NewFreeTextWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetText(terraformDefinition["text"].(string))
	// Optional params
	if v, ok := terraformDefinition["color"].(string); ok && len(v) != 0 {
		datadogDefinition.SetColor(v)
	}
	if v, ok := terraformDefinition["font_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetFontSize(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	return datadogDefinition
}

func buildTerraformFreeTextDefinition(datadogDefinition datadogV1.FreeTextWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["text"] = datadogDefinition.GetText()
	// Optional params
	if v, ok := datadogDefinition.GetColorOk(); ok {
		terraformDefinition["color"] = *v
	}
	if v, ok := datadogDefinition.GetFontSizeOk(); ok {
		terraformDefinition["font_size"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
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
		"event": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetEventSchema(),
			},
		},
		"show_legend": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"legend_size": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateTimeseriesWidgetLegendSize,
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
func buildDatadogHeatmapDefinition(terraformDefinition map[string]interface{}) *datadogV1.HeatMapWidgetDefinition {
	datadogDefinition := datadogV1.NewHeatMapWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogHeatmapRequests(&terraformRequests)
	// Optional params
	if axis, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(v)
		}
	}
	if v, ok := terraformDefinition["event"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Events = buildDatadogWidgetEvents(&v)
	}
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.SetShowLegend(v)
	}
	if v, ok := terraformDefinition["legend_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendSize(datadogV1.WidgetLegendSize(v))
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}
func buildTerraformHeatmapDefinition(datadogDefinition datadogV1.HeatMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformHeatmapRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["yaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetEventsOk(); ok {
		terraformDefinition["event"] = buildTerraformWidgetEvents(v)
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetShowLegendOk(); ok {
		terraformDefinition["show_legend"] = *v
	}
	if v, ok := datadogDefinition.GetLegendSizeOk(); ok {
		terraformDefinition["legend_size"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func getHeatmapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
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
func buildDatadogHeatmapRequests(terraformRequests *[]interface{}) *[]datadogV1.HeatMapWidgetRequest {
	datadogRequests := make([]datadogV1.HeatMapWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build HeatmapRequest
		datadogHeatmapRequest := datadogV1.NewHeatMapWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogHeatmapRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogHeatmapRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogHeatmapRequest.Style = buildDatadogWidgetStyle(v)
			}
		}
		datadogRequests[i] = *datadogHeatmapRequest
	}
	return &datadogRequests
}
func buildTerraformHeatmapRequests(datadogHeatmapRequests *[]datadogV1.HeatMapWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogHeatmapRequests))
	for i, datadogRequest := range *datadogHeatmapRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
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
func buildDatadogHostmapDefinition(terraformDefinition map[string]interface{}) *datadogV1.HostMapWidgetDefinition {

	// Required params
	datadogDefinition := datadogV1.NewHostMapWidgetDefinitionWithDefaults()
	if v, ok := terraformDefinition["request"].([]interface{}); ok && len(v) > 0 {
		terraformRequests := v[0].(map[string]interface{})
		datadogRequests := datadogV1.NewHostMapWidgetDefinitionRequests()
		if terraformFillArray, ok := terraformRequests["fill"].([]interface{}); ok && len(terraformFillArray) > 0 {
			terraformFill := terraformFillArray[0].(map[string]interface{})
			datadogRequests.Fill = buildDatadogHostmapRequest(terraformFill)
		}
		if terraformSizeArray, ok := terraformRequests["size"].([]interface{}); ok && len(terraformSizeArray) > 0 {
			terraformSize := terraformSizeArray[0].(map[string]interface{})
			datadogRequests.Size = buildDatadogHostmapRequest(terraformSize)
		}
		datadogDefinition.SetRequests(*datadogRequests)
	}

	// Optional params
	if v, ok := terraformDefinition["node_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetNodeType(datadogV1.WidgetNodeType(v))
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
		datadogDefinition.Group = &datadogGroups
	}
	if terraformScopes, ok := terraformDefinition["scope"].([]interface{}); ok && len(terraformScopes) > 0 {
		datadogScopes := make([]string, len(terraformScopes))
		for i, Scope := range terraformScopes {
			datadogScopes[i] = Scope.(string)
		}
		datadogDefinition.SetScope(datadogScopes)
	}
	if style, ok := terraformDefinition["style"].([]interface{}); ok && len(style) > 0 {
		if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
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
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	return datadogDefinition
}
func buildTerraformHostmapDefinition(datadogDefinition datadogV1.HostMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformRequests := map[string]interface{}{}
	if v, ok := datadogDefinition.Requests.GetSizeOk(); ok {
		terraformSize := buildTerraformHostmapRequest(v)
		terraformRequests["size"] = []map[string]interface{}{*terraformSize}
	}
	if v, ok := datadogDefinition.Requests.GetFillOk(); ok {
		terraformFill := buildTerraformHostmapRequest(v)
		terraformRequests["fill"] = []map[string]interface{}{*terraformFill}
	}
	terraformDefinition["request"] = []map[string]interface{}{terraformRequests}
	// Optional params
	if v, ok := datadogDefinition.GetNodeTypeOk(); ok {
		terraformDefinition["node_type"] = *v
	}
	if v, ok := datadogDefinition.GetNoMetricHostsOk(); ok {
		terraformDefinition["no_metric_hosts"] = *v
	}
	if v, ok := datadogDefinition.GetNoGroupHostsOk(); ok {
		terraformDefinition["no_group_hosts"] = *v
	}
	if v, ok := datadogDefinition.GetGroupOk(); ok {
		terraformGroups := make([]string, len(*v))
		for i, datadogGroup := range *v {
			terraformGroups[i] = datadogGroup
		}
		terraformDefinition["group"] = terraformGroups
	}
	if v, ok := datadogDefinition.GetScopeOk(); ok {
		terraformScopes := make([]string, len(*v))
		for i, datadogScope := range *v {
			terraformScopes[i] = datadogScope
		}
		terraformDefinition["scope"] = terraformScopes
	}
	if v, ok := datadogDefinition.GetStyleOk(); ok {
		style := buildTerraformHostmapRequestStyle(*v)
		terraformDefinition["style"] = []map[string]interface{}{style}
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	return terraformDefinition
}

func getHostmapRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement at least one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
	}
}
func buildDatadogHostmapRequest(terraformRequest map[string]interface{}) *datadogV1.HostMapRequest {

	datadogHostmapRequest := &datadogV1.HostMapRequest{}
	if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
		datadogHostmapRequest.SetQ(v)
	} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
		apmQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
	} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
		logQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
	} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
		processQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
	} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
		rumQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
	} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
		securityQuery := v[0].(map[string]interface{})
		datadogHostmapRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
	}

	return datadogHostmapRequest
}
func buildTerraformHostmapRequest(datadogHostmapRequest *datadogV1.HostMapRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}
	if v, ok := datadogHostmapRequest.GetQOk(); ok {
		terraformRequest["q"] = v
	} else if v, ok := datadogHostmapRequest.GetApmQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogHostmapRequest.GetLogQueryOk(); ok {
		terraformQuery := buildTerraformApmOrLogQuery(*v)
		terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
	} else if v, ok := datadogHostmapRequest.GetProcessQueryOk(); ok {
		terraformQuery := buildTerraformProcessQuery(*v)
		terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogHostmapRequest.RumQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogHostmapRequest.RumQuery)
		terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogHostmapRequest.SecurityQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogHostmapRequest.SecurityQuery)
		terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
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

func buildDatadogIframeDefinition(terraformDefinition map[string]interface{}) *datadogV1.IFrameWidgetDefinition {
	datadogDefinition := datadogV1.NewIFrameWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetUrl(terraformDefinition["url"].(string))
	return datadogDefinition
}

func buildTerraformIframeDefinition(datadogDefinition datadogV1.IFrameWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["url"] = datadogDefinition.GetUrl()
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

func buildDatadogImageDefinition(terraformDefinition map[string]interface{}) *datadogV1.ImageWidgetDefinition {
	datadogDefinition := datadogV1.NewImageWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetUrl(terraformDefinition["url"].(string))
	// Optional params
	if v, ok := terraformDefinition["sizing"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSizing(datadogV1.WidgetImageSizing(v))
	}
	if v, ok := terraformDefinition["margin"].(string); ok && len(v) != 0 {
		datadogDefinition.SetMargin(datadogV1.WidgetMargin(v))
	}
	return datadogDefinition
}

func buildTerraformImageDefinition(datadogDefinition datadogV1.ImageWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["url"] = datadogDefinition.GetUrl()
	// Optional params
	if v, ok := datadogDefinition.GetSizingOk(); ok {
		terraformDefinition["sizing"] = *v
	}
	if v, ok := datadogDefinition.GetMarginOk(); ok {
		terraformDefinition["margin"] = *v
	}
	return terraformDefinition
}

//
// Log Stream Widget Definition helpers
//

func getLogStreamDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"indexes": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"logset": {
			Type:       schema.TypeString,
			Deprecated: "This parameter has been deprecated. Use 'indexes' instead",
			Optional:   true,
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
		"show_date_column": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"show_message_column": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"message_display": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "One of: ['inline', 'expanded-md', 'expanded-lg']",
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				value := val.(string)
				switch value {
				case "inline", "expanded-md", "expanded-lg":
					break
				default:
					errs = append(errs, fmt.Errorf(
						"%q contains an invalid value %q. Valid values are `inline`, `expanded-md`, or `expanded-lg`", key, value))
				}
				return
			},
		},
		"sort": {
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetFieldSortSchema(),
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

func getWidgetFieldSortSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"column": {
			Type:     schema.TypeString,
			Required: true,
		},
		"order": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				value := val.(string)
				switch value {
				case "asc", "desc":
					break
				default:
					errs = append(errs, fmt.Errorf(
						"%q contains an invalid value %q. Valid values are `asc`, or `desc`", key, value))
				}
				return
			},
		},
	}
}

func buildDatadogLogStreamDefinition(terraformDefinition map[string]interface{}) *datadogV1.LogStreamWidgetDefinition {
	datadogDefinition := datadogV1.NewLogStreamWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetLogset(terraformDefinition["logset"].(string))
	terraformIndexes := terraformDefinition["indexes"].([]interface{})
	datadogIndexes := make([]string, len(terraformIndexes))
	for i, index := range terraformIndexes {
		datadogIndexes[i] = index.(string)
	}
	datadogDefinition.SetIndexes(datadogIndexes)
	// Optional params
	if v, ok := terraformDefinition["query"].(string); ok && len(v) != 0 {
		datadogDefinition.SetQuery(v)
	}
	if terraformColumns, ok := terraformDefinition["columns"].([]interface{}); ok && len(terraformColumns) > 0 {
		datadogColumns := make([]string, len(terraformColumns))
		for i, column := range terraformColumns {
			datadogColumns[i] = column.(string)
		}
		datadogDefinition.SetColumns(datadogColumns)
	}
	if v, ok := terraformDefinition["show_date_column"].(bool); ok {
		datadogDefinition.SetShowDateColumn(v)
	}
	if v, ok := terraformDefinition["show_message_column"].(bool); ok {
		datadogDefinition.SetShowMessageColumn(v)
	}
	if v, ok := terraformDefinition["message_display"].(string); ok && len(v) != 0 {
		datadogDefinition.SetMessageDisplay(datadogV1.WidgetMessageDisplay(v))
	}
	if v, ok := terraformDefinition["sort"].([]interface{}); ok && len(v) > 0 {
		if v, ok := v[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Sort = buildDatadogWidgetFieldSort(v)
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}

func buildDatadogWidgetFieldSort(terraformWidgetFieldSort map[string]interface{}) *datadogV1.WidgetFieldSort {
	datadogWidgetFieldSort := &datadogV1.WidgetFieldSort{}
	if v, ok := terraformWidgetFieldSort["column"].(string); ok && len(v) != 0 {
		datadogWidgetFieldSort.SetColumn(v)
	}
	if v, ok := terraformWidgetFieldSort["order"].(string); ok && len(v) != 0 {
		datadogWidgetFieldSort.SetOrder(datadogV1.WidgetSort(v))
	}
	return datadogWidgetFieldSort
}

func buildTerraformLogStreamDefinition(datadogDefinition datadogV1.LogStreamWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Optional params

	// Indexes is the recommended required field, but we still allow setting logsets instead for backwards compatibility
	if v, ok := datadogDefinition.GetIndexesOk(); ok {
		terraformDefinition["indexes"] = *v
	}

	if v, ok := datadogDefinition.GetLogsetOk(); ok {
		terraformDefinition["logset"] = *v
	}
	if v, ok := datadogDefinition.GetQueryOk(); ok {
		terraformDefinition["query"] = *v
	}
	if v, ok := datadogDefinition.GetColumnsOk(); ok {
		terraformColumns := make([]string, len(*v))
		for i, datadogColumn := range *v {
			terraformColumns[i] = datadogColumn
		}
		terraformDefinition["columns"] = terraformColumns
	}
	if v, ok := datadogDefinition.GetShowDateColumnOk(); ok {
		terraformDefinition["show_date_column"] = *v
	}
	if v, ok := datadogDefinition.GetShowMessageColumnOk(); ok {
		terraformDefinition["show_message_column"] = *v
	}
	if v, ok := datadogDefinition.GetMessageDisplayOk(); ok {
		terraformDefinition["message_display"] = *v
	}
	if v, ok := datadogDefinition.GetSortOk(); ok {
		sort := buildTerraformWidgetFieldSort(*v)
		terraformDefinition["sort"] = []map[string]interface{}{sort}
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func buildTerraformWidgetFieldSort(datadogWidgetFieldSort datadogV1.WidgetFieldSort) map[string]interface{} {
	terraformWidgetFieldSort := map[string]interface{}{}
	if v, ok := datadogWidgetFieldSort.GetColumnOk(); ok {
		terraformWidgetFieldSort["column"] = string(*v)
	}
	if v, ok := datadogWidgetFieldSort.GetOrderOk(); ok {
		terraformWidgetFieldSort["order"] = string(*v)
	}
	return terraformWidgetFieldSort
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
		"sort": {
			Type:     schema.TypeString,
			Optional: true,
		},
		// The count param is deprecated
		"count": {
			Type:       schema.TypeInt,
			Deprecated: "This parameter has been deprecated",
			Optional:   true,
			Default:    50,
		},
		// The start param is deprecated
		"start": {
			Type:       schema.TypeInt,
			Deprecated: "This parameter has been deprecated",
			Optional:   true,
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
		"show_last_triggered": {
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

func buildDatadogManageStatusDefinition(terraformDefinition map[string]interface{}) *datadogV1.MonitorSummaryWidgetDefinition {
	datadogDefinition := datadogV1.NewMonitorSummaryWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetQuery(terraformDefinition["query"].(string))
	// Optional params
	if v, ok := terraformDefinition["summary_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSummaryType(datadogV1.WidgetSummaryType(v))
	}
	if v, ok := terraformDefinition["sort"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSort(datadogV1.WidgetMonitorSummarySort(v))
	}
	if v, ok := terraformDefinition["count"].(int); ok {
		datadogDefinition.SetCount(int64(v))
	}
	if v, ok := terraformDefinition["start"].(int); ok {
		datadogDefinition.SetStart(int64(v))
	}
	if v, ok := terraformDefinition["display_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetDisplayFormat(datadogV1.WidgetMonitorSummaryDisplayFormat(v))
	}
	if v, ok := terraformDefinition["color_preference"].(string); ok && len(v) != 0 {
		datadogDefinition.SetColorPreference(datadogV1.WidgetColorPreference(v))
	}
	if v, ok := terraformDefinition["hide_zero_counts"].(bool); ok {
		datadogDefinition.SetHideZeroCounts(v)
	}
	if v, ok := terraformDefinition["show_last_triggered"].(bool); ok {
		datadogDefinition.SetShowLastTriggered(v)
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	return datadogDefinition
}

func buildTerraformManageStatusDefinition(datadogDefinition datadogV1.MonitorSummaryWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["query"] = datadogDefinition.GetQuery()
	// Optional params
	if v, ok := datadogDefinition.GetSummaryTypeOk(); ok {
		terraformDefinition["summary_type"] = *v
	}
	if v, ok := datadogDefinition.GetSortOk(); ok {
		terraformDefinition["sort"] = *v
	}
	//Below fields are deprecated
	if v, ok := datadogDefinition.GetCountOk(); ok {
		terraformDefinition["count"] = *v
	}
	if v, ok := datadogDefinition.GetStartOk(); ok {
		terraformDefinition["start"] = *v
	}
	if v, ok := datadogDefinition.GetDisplayFormatOk(); ok {
		terraformDefinition["display_format"] = *v
	}
	if v, ok := datadogDefinition.GetColorPreferenceOk(); ok {
		terraformDefinition["color_preference"] = *v
	}
	if v, ok := datadogDefinition.GetHideZeroCountsOk(); ok {
		terraformDefinition["hide_zero_counts"] = *v
	}
	if v, ok := datadogDefinition.GetShowLastTriggeredOk(); ok {
		terraformDefinition["show_last_triggered"] = *v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
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

func buildDatadogNoteDefinition(terraformDefinition map[string]interface{}) *datadogV1.NoteWidgetDefinition {
	datadogDefinition := datadogV1.NewNoteWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetContent(terraformDefinition["content"].(string))
	// Optional params
	if v, ok := terraformDefinition["background_color"].(string); ok && len(v) != 0 {
		datadogDefinition.SetBackgroundColor(v)
	}
	if v, ok := terraformDefinition["font_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetFontSize(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["show_tick"]; ok {
		datadogDefinition.SetShowTick(v.(bool))
	}
	if v, ok := terraformDefinition["tick_pos"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTickPos(v)
	}
	if v, ok := terraformDefinition["tick_edge"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTickEdge(datadogV1.WidgetTickEdge(v))
	}
	return datadogDefinition
}

func buildTerraformNoteDefinition(datadogDefinition datadogV1.NoteWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["content"] = datadogDefinition.GetContent()
	// Optional params
	if v, ok := datadogDefinition.GetBackgroundColorOk(); ok {
		terraformDefinition["background_color"] = *v
	}
	if v, ok := datadogDefinition.GetFontSizeOk(); ok {
		terraformDefinition["font_size"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
	}
	if v, ok := datadogDefinition.GetShowTickOk(); ok {
		terraformDefinition["show_tick"] = *v
	}
	if v, ok := datadogDefinition.GetTickPosOk(); ok {
		terraformDefinition["tick_pos"] = *v
	}
	if v, ok := datadogDefinition.GetTickEdgeOk(); ok {
		terraformDefinition["tick_edge"] = *v
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
func buildDatadogQueryValueDefinition(terraformDefinition map[string]interface{}) *datadogV1.QueryValueWidgetDefinition {
	datadogDefinition := datadogV1.NewQueryValueWidgetDefinitionWithDefaults()
	// Required params
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
		datadogDefinition.SetPrecision(int64(v))
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["text_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTextAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformQueryValueDefinition(datadogDefinition datadogV1.QueryValueWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformQueryValueRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetAutoscaleOk(); ok {
		terraformDefinition["autoscale"] = *v
	}
	if v, ok := datadogDefinition.GetCustomUnitOk(); ok {
		terraformDefinition["custom_unit"] = *v
	}
	if v, ok := datadogDefinition.GetPrecisionOk(); ok {
		terraformDefinition["precision"] = *v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTextAlignOk(); ok {
		terraformDefinition["text_align"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func getQueryValueRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
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
func buildDatadogQueryValueRequests(terraformRequests *[]interface{}) *[]datadogV1.QueryValueWidgetRequest {
	datadogRequests := make([]datadogV1.QueryValueWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build QueryValueRequest
		datadogQueryValueRequest := datadogV1.NewQueryValueWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogQueryValueRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogQueryValueRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		}

		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogQueryValueRequest.ConditionalFormats = buildDatadogWidgetConditionalFormat(&v)
		}
		if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
			datadogQueryValueRequest.SetAggregator(datadogV1.WidgetAggregator(v))
		}

		datadogRequests[i] = *datadogQueryValueRequest
	}
	return &datadogRequests
}
func buildTerraformQueryValueRequests(datadogQueryValueRequests *[]datadogV1.QueryValueWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogQueryValueRequests))
	for i, datadogRequest := range *datadogQueryValueRequests {
		terraformRequest := map[string]interface{}{}
		if datadogRequest.Q != nil {
			terraformRequest["q"] = datadogRequest.GetQ()
		} else if datadogRequest.ApmQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.ApmQuery)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.LogQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.LogQuery)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.ProcessQuery != nil {
			terraformQuery := buildTerraformProcessQuery(*datadogRequest.ProcessQuery)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}

		if datadogRequest.ConditionalFormats != nil {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(datadogRequest.ConditionalFormats)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}

		if v, ok := datadogRequest.GetAggregatorOk(); ok {
			terraformRequest["aggregator"] = *v
		}
		terraformRequests[i] = terraformRequest
	}
	return &terraformRequests
}

//
// Query Table Widget Definition helpers
//
func getQueryTableDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getQueryTableRequestSchema(),
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
func buildDatadogQueryTableDefinition(terraformDefinition map[string]interface{}) *datadogV1.TableWidgetDefinition {
	datadogDefinition := datadogV1.NewTableWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogQueryTableRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformQueryTableDefinition(datadogDefinition datadogV1.TableWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformQueryTableRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func getQueryTableRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
		// Settings specific to QueryTable requests
		"conditional_formats": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Resource{
				Schema: getWidgetConditionalFormatSchema(),
			},
		},
		"alias": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"aggregator": {
			Type:     schema.TypeString,
			Optional: true,
		},
		"limit": {
			Type:     schema.TypeInt,
			Optional: true,
		},
		"order": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogQueryTableRequests(terraformRequests *[]interface{}) *[]datadogV1.TableWidgetRequest {
	datadogRequests := make([]datadogV1.TableWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build QueryTableRequest
		datadogQueryTableRequest := datadogV1.NewTableWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogQueryTableRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		}

		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogQueryTableRequest.ConditionalFormats = buildDatadogWidgetConditionalFormat(&v)
		}
		if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetAggregator(datadogV1.WidgetAggregator(v))
		}
		if v, ok := terraformRequest["alias"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetAlias(v)

		}
		if v, ok := terraformRequest["limit"].(int); ok && v != 0 {
			datadogQueryTableRequest.SetLimit(int64(v))
		}
		if v, ok := terraformRequest["order"].(string); ok && len(v) != 0 {
			datadogQueryTableRequest.SetOrder(datadogV1.WidgetSort(v))
		}
		datadogRequests[i] = *datadogQueryTableRequest
	}
	return &datadogRequests
}
func buildTerraformQueryTableRequests(datadogQueryTableRequests *[]datadogV1.TableWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogQueryTableRequests))
	for i, datadogRequest := range *datadogQueryTableRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}

		if v, ok := datadogRequest.GetConditionalFormatsOk(); ok {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(v)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}

		if v, ok := datadogRequest.GetAggregatorOk(); ok {
			terraformRequest["aggregator"] = *v
		}
		if v, ok := datadogRequest.GetAliasOk(); ok {
			terraformRequest["alias"] = *v
		}
		if v, ok := datadogRequest.GetLimitOk(); ok {
			terraformRequest["limit"] = *v
		}
		if v, ok := datadogRequest.GetOrderOk(); ok {
			terraformRequest["order"] = *v
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
func buildDatadogScatterplotDefinition(terraformDefinition map[string]interface{}) *datadogV1.ScatterPlotWidgetDefinition {
	datadogDefinition := datadogV1.NewScatterPlotWidgetDefinitionWithDefaults()
	// Required params
	if v, ok := terraformDefinition["request"].([]interface{}); ok && len(v) > 0 {
		terraformRequests := v[0].(map[string]interface{})
		datadogRequests := datadogV1.NewScatterPlotWidgetDefinitionRequestsWithDefaults()
		if terraformXArray, ok := terraformRequests["x"].([]interface{}); ok && len(terraformXArray) > 0 {
			terraformX := terraformXArray[0].(map[string]interface{})
			datadogRequests.SetX(*buildDatadogScatterplotRequest(terraformX))
		}
		if terraformYArray, ok := terraformRequests["y"].([]interface{}); ok && len(terraformYArray) > 0 {
			terraformY := terraformYArray[0].(map[string]interface{})
			datadogRequests.SetY(*buildDatadogScatterplotRequest(terraformY))
		}
		datadogDefinition.SetRequests(*datadogRequests)
	}

	// Optional params
	if axis, ok := terraformDefinition["xaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Xaxis = buildDatadogWidgetAxis(v)
		}
	}
	if axis, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(axis) > 0 {
		if v, ok := axis[0].(map[string]interface{}); ok && len(v) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(v)
		}
	}
	if terraformColorByGroups, ok := terraformDefinition["color_by_groups"].([]interface{}); ok && len(terraformColorByGroups) > 0 {
		datadogColorByGroups := make([]string, len(terraformColorByGroups))
		for i, colorByGroup := range terraformColorByGroups {
			datadogColorByGroups[i] = colorByGroup.(string)
		}
		datadogDefinition.ColorByGroups = &datadogColorByGroups
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}
func buildTerraformScatterplotDefinition(datadogDefinition datadogV1.ScatterPlotWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformRequests := map[string]interface{}{}
	if v, ok := datadogDefinition.Requests.GetXOk(); ok {
		terraformX := buildTerraformScatterplotRequest(v)
		terraformRequests["x"] = []map[string]interface{}{*terraformX}
	}
	if v, ok := datadogDefinition.Requests.GetYOk(); ok {
		terraformY := buildTerraformScatterplotRequest(v)
		terraformRequests["y"] = []map[string]interface{}{*terraformY}
	}
	terraformDefinition["request"] = []map[string]interface{}{terraformRequests}

	// Optional params
	if v, ok := datadogDefinition.GetXaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["xaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["yaxis"] = []map[string]interface{}{axis}
	}

	if v, ok := datadogDefinition.GetColorByGroupsOk(); ok {
		terraformColorByGroups := make([]string, len(*v))
		for i, datadogColorByGroup := range *v {
			terraformColorByGroups[i] = datadogColorByGroup
		}
		terraformDefinition["color_by_groups"] = terraformColorByGroups
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	return terraformDefinition
}

func getScatterplotRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
		// Settings specific to Scatterplot requests
		"aggregator": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogScatterplotRequest(terraformRequest map[string]interface{}) *datadogV1.ScatterPlotRequest {

	datadogScatterplotRequest := datadogV1.NewScatterPlotRequest()
	if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
		datadogScatterplotRequest.SetQ(v)
	} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
		apmQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
	} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
		logQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
	} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
		processQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
	} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
		rumQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
	} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
		securityQuery := v[0].(map[string]interface{})
		datadogScatterplotRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
	}

	if v, ok := terraformRequest["aggregator"].(string); ok && len(v) != 0 {
		datadogScatterplotRequest.SetAggregator(datadogV1.WidgetAggregator(v))
	}

	return datadogScatterplotRequest
}
func buildTerraformScatterplotRequest(datadogScatterplotRequest *datadogV1.ScatterPlotRequest) *map[string]interface{} {
	terraformRequest := map[string]interface{}{}
	if datadogScatterplotRequest.Q != nil {
		terraformRequest["q"] = datadogScatterplotRequest.GetQ()
	} else if datadogScatterplotRequest.ApmQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogScatterplotRequest.ApmQuery)
		terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogScatterplotRequest.LogQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogScatterplotRequest.LogQuery)
		terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogScatterplotRequest.ProcessQuery != nil {
		terraformQuery := buildTerraformProcessQuery(*datadogScatterplotRequest.ProcessQuery)
		terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogScatterplotRequest.RumQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogScatterplotRequest.RumQuery)
		terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
	} else if datadogScatterplotRequest.SecurityQuery != nil {
		terraformQuery := buildTerraformApmOrLogQuery(*datadogScatterplotRequest.SecurityQuery)
		terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
	}

	if datadogScatterplotRequest.Aggregator != nil {
		terraformRequest["aggregator"] = *datadogScatterplotRequest.Aggregator
	}
	return &terraformRequest
}

//
// ServiceMap Widget Definition helpers
//

func getServiceMapDefinitionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"service": {
			Type:     schema.TypeString,
			Required: true,
		},
		"filters": {
			Type:     schema.TypeList,
			Required: true,
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
	}
}
func buildDatadogServiceMapDefinition(terraformDefinition map[string]interface{}) *datadogV1.ServiceMapWidgetDefinition {
	datadogDefinition := datadogV1.NewServiceMapWidgetDefinitionWithDefaults()

	// Required params
	datadogDefinition.SetService(terraformDefinition["service"].(string))
	terraformFilters := terraformDefinition["filters"].([]interface{})
	datadogFilters := make([]string, len(terraformFilters))
	for i, terraformFilter := range terraformFilters {
		datadogFilters[i] = terraformFilter.(string)
	}
	datadogDefinition.SetFilters(datadogFilters)

	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}

	return datadogDefinition
}
func buildTerraformServiceMapDefinition(datadogDefinition datadogV1.ServiceMapWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}

	// Required params
	terraformDefinition["service"] = datadogDefinition.GetService()
	terraformDefinition["filters"] = datadogDefinition.GetFilters()

	// Optional params
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}

	return terraformDefinition
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
			Required: true,
		},
		"slo_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"show_error_budget": {
			Type:     schema.TypeBool,
			Optional: true,
		},
		"view_mode": {
			Type:     schema.TypeString,
			Required: true,
		},
		"time_windows": {
			Type:     schema.TypeList,
			Required: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
}

func buildDatadogServiceLevelObjectiveDefinition(terraformDefinition map[string]interface{}) *datadogV1.SLOWidgetDefinition {
	datadogDefinition := datadogV1.NewSLOWidgetDefinitionWithDefaults()
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["view_type"].(string); ok && len(v) != 0 {
		datadogDefinition.SetViewType(v)
	}
	if v, ok := terraformDefinition["slo_id"].(string); ok && len(v) != 0 {
		datadogDefinition.SetSloId(v)
	}
	if v, ok := terraformDefinition["show_error_budget"].(bool); ok {
		datadogDefinition.SetShowErrorBudget(v)
	}
	if v, ok := terraformDefinition["view_mode"].(string); ok && len(v) != 0 {
		datadogDefinition.SetViewMode(datadogV1.WidgetViewMode(v))
	}
	if terraformTimeWindows, ok := terraformDefinition["time_windows"].([]interface{}); ok && len(terraformTimeWindows) > 0 {
		datadogTimeWindows := make([]datadogV1.WidgetTimeWindows, len(terraformTimeWindows))
		for i, timeWindows := range terraformTimeWindows {
			datadogTimeWindows[i] = datadogV1.WidgetTimeWindows(timeWindows.(string))
		}
		datadogDefinition.TimeWindows = &datadogTimeWindows
	}
	return datadogDefinition
}

func buildTerraformServiceLevelObjectiveDefinition(datadogDefinition datadogV1.SLOWidgetDefinition) map[string]interface{} {
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
	if datadogDefinition.SloId != nil {
		terraformDefinition["slo_id"] = datadogDefinition.GetSloId()
	}
	if showErrorBudget, ok := datadogDefinition.GetShowErrorBudgetOk(); ok {
		terraformDefinition["show_error_budget"] = showErrorBudget
	}
	if viewMode, ok := datadogDefinition.GetViewModeOk(); ok {
		terraformDefinition["view_mode"] = viewMode
	}
	if datadogDefinition.TimeWindows != nil {
		terraformTimeWindows := make([]string, len(datadogDefinition.GetTimeWindows()))
		for i, datadogTimeWindow := range datadogDefinition.GetTimeWindows() {
			terraformTimeWindows[i] = string(datadogTimeWindow)
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
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validateTimeseriesWidgetLegendSize,
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

func buildDatadogTimeseriesDefinition(terraformDefinition map[string]interface{}) *datadogV1.TimeseriesWidgetDefinition {
	datadogDefinition := datadogV1.NewTimeseriesWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogTimeseriesRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["marker"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Markers = buildDatadogWidgetMarkers(&v)
	}
	if v, ok := terraformDefinition["event"].([]interface{}); ok && len(v) > 0 {
		datadogDefinition.Events = buildDatadogWidgetEvents(&v)
	}
	if v, ok := terraformDefinition["yaxis"].([]interface{}); ok && len(v) > 0 {
		if axis, ok := v[0].(map[string]interface{}); ok && len(axis) > 0 {
			datadogDefinition.Yaxis = buildDatadogWidgetAxis(axis)
		}
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	if v, ok := terraformDefinition["show_legend"].(bool); ok {
		datadogDefinition.SetShowLegend(v)
	}
	if v, ok := terraformDefinition["legend_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetLegendSize(datadogV1.WidgetLegendSize(v))
	}
	return datadogDefinition
}

func buildTerraformTimeseriesDefinition(datadogDefinition datadogV1.TimeseriesWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["request"] = buildTerraformTimeseriesRequests(&datadogDefinition.Requests)
	// Optional params
	if v, ok := datadogDefinition.GetMarkersOk(); ok {
		terraformDefinition["marker"] = buildTerraformWidgetMarkers(v)
	}
	if v, ok := datadogDefinition.GetEventsOk(); ok {
		terraformDefinition["event"] = buildTerraformWidgetEvents(v)
	}
	if v, ok := datadogDefinition.GetYaxisOk(); ok {
		axis := buildTerraformWidgetAxis(*v)
		terraformDefinition["yaxis"] = []map[string]interface{}{axis}
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = *v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = *v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = *v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
	}
	if v, ok := datadogDefinition.GetShowLegendOk(); ok {
		terraformDefinition["show_legend"] = *v
	}
	if v, ok := datadogDefinition.GetLegendSizeOk(); ok {
		terraformDefinition["legend_size"] = *v
	}

	return terraformDefinition
}

func getTimeseriesRequestSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		// A request should implement exactly one of the following type of query
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"rum_query":      getApmLogNetworkOrRumQuerySchema(),
		"network_query":  getApmLogNetworkOrRumQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
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
func buildDatadogTimeseriesRequests(terraformRequests *[]interface{}) *[]datadogV1.TimeseriesWidgetRequest {
	datadogRequests := make([]datadogV1.TimeseriesWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build TimeseriesRequest
		datadogTimeseriesRequest := datadogV1.NewTimeseriesWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogTimeseriesRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["network_query"].([]interface{}); ok && len(v) > 0 {
			networkQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.NetworkQuery = buildDatadogApmOrLogQuery(networkQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogTimeseriesRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogTimeseriesRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}
		// Metadata
		if terraformMetadataList, ok := terraformRequest["metadata"].([]interface{}); ok && len(terraformMetadataList) > 0 {
			datadogMetadataList := make([]datadogV1.TimeseriesWidgetRequestMetadata, len(terraformMetadataList))
			for i, m := range terraformMetadataList {
				metadata := m.(map[string]interface{})
				// Expression
				datadogMetadata := datadogV1.NewTimeseriesWidgetRequestMetadata(metadata["expression"].(string))
				// AliasName
				if v, ok := metadata["alias_name"].(string); ok && len(v) != 0 {
					datadogMetadata.SetAliasName(v)
				}
				datadogMetadataList[i] = *datadogMetadata
			}
			datadogTimeseriesRequest.SetMetadata(datadogMetadataList)
		}
		if v, ok := terraformRequest["display_type"].(string); ok && len(v) != 0 {
			datadogTimeseriesRequest.SetDisplayType(datadogV1.WidgetDisplayType(v))
		}
		datadogRequests[i] = *datadogTimeseriesRequest
	}
	return &datadogRequests
}
func buildTerraformTimeseriesRequests(datadogTimeseriesRequests *[]datadogV1.TimeseriesWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogTimeseriesRequests))
	for i, datadogRequest := range *datadogTimeseriesRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetNetworkQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["network_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetRumQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetRequestStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
		}
		// Metadata
		if datadogRequest.Metadata != nil {
			terraformMetadataList := make([]map[string]interface{}, len(datadogRequest.GetMetadata()))
			for i, metadata := range datadogRequest.GetMetadata() {
				// Expression
				terraformMetadata := map[string]interface{}{
					"expression": metadata.GetExpression(),
				}
				// AliasName
				if metadata.AliasName != nil {
					terraformMetadata["alias_name"] = metadata.GetAliasName()
				}

				terraformMetadataList[i] = terraformMetadata
			}
			terraformRequest["metadata"] = &terraformMetadataList
		}
		if v, ok := datadogRequest.GetDisplayTypeOk(); ok {
			terraformRequest["display_type"] = v
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
func buildDatadogToplistDefinition(terraformDefinition map[string]interface{}) *datadogV1.ToplistWidgetDefinition {
	datadogDefinition := datadogV1.NewToplistWidgetDefinitionWithDefaults()
	// Required params
	terraformRequests := terraformDefinition["request"].([]interface{})
	datadogDefinition.Requests = *buildDatadogToplistRequests(&terraformRequests)
	// Optional params
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.Time = buildDatadogWidgetTime(v)
	}
	return datadogDefinition
}
func buildTerraformToplistDefinition(datadogDefinition datadogV1.ToplistWidgetDefinition) map[string]interface{} {
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
		"q":              getMetricQuerySchema(),
		"apm_query":      getApmOrLogQuerySchema(),
		"log_query":      getApmOrLogQuerySchema(),
		"process_query":  getProcessQuerySchema(),
		"rum_query":      getApmOrLogQuerySchema(),
		"security_query": getApmOrLogQuerySchema(),
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
func buildDatadogToplistRequests(terraformRequests *[]interface{}) *[]datadogV1.ToplistWidgetRequest {
	datadogRequests := make([]datadogV1.ToplistWidgetRequest, len(*terraformRequests))
	for i, r := range *terraformRequests {
		terraformRequest := r.(map[string]interface{})
		// Build ToplistRequest
		datadogToplistRequest := datadogV1.NewToplistWidgetRequest()
		if v, ok := terraformRequest["q"].(string); ok && len(v) != 0 {
			datadogToplistRequest.SetQ(v)
		} else if v, ok := terraformRequest["apm_query"].([]interface{}); ok && len(v) > 0 {
			apmQuery := v[0].(map[string]interface{})
			datadogToplistRequest.ApmQuery = buildDatadogApmOrLogQuery(apmQuery)
		} else if v, ok := terraformRequest["log_query"].([]interface{}); ok && len(v) > 0 {
			logQuery := v[0].(map[string]interface{})
			datadogToplistRequest.LogQuery = buildDatadogApmOrLogQuery(logQuery)
		} else if v, ok := terraformRequest["process_query"].([]interface{}); ok && len(v) > 0 {
			processQuery := v[0].(map[string]interface{})
			datadogToplistRequest.ProcessQuery = buildDatadogProcessQuery(processQuery)
		} else if v, ok := terraformRequest["rum_query"].([]interface{}); ok && len(v) > 0 {
			rumQuery := v[0].(map[string]interface{})
			datadogToplistRequest.RumQuery = buildDatadogApmOrLogQuery(rumQuery)
		} else if v, ok := terraformRequest["security_query"].([]interface{}); ok && len(v) > 0 {
			securityQuery := v[0].(map[string]interface{})
			datadogToplistRequest.SecurityQuery = buildDatadogApmOrLogQuery(securityQuery)
		}
		if v, ok := terraformRequest["conditional_formats"].([]interface{}); ok && len(v) != 0 {
			datadogToplistRequest.ConditionalFormats = buildDatadogWidgetConditionalFormat(&v)
		}
		if style, ok := terraformRequest["style"].([]interface{}); ok && len(style) > 0 {
			if v, ok := style[0].(map[string]interface{}); ok && len(v) > 0 {
				datadogToplistRequest.Style = buildDatadogWidgetRequestStyle(v)
			}
		}
		datadogRequests[i] = *datadogToplistRequest
	}
	return &datadogRequests
}
func buildTerraformToplistRequests(datadogToplistRequests *[]datadogV1.ToplistWidgetRequest) *[]map[string]interface{} {
	terraformRequests := make([]map[string]interface{}, len(*datadogToplistRequests))
	for i, datadogRequest := range *datadogToplistRequests {
		terraformRequest := map[string]interface{}{}
		if v, ok := datadogRequest.GetQOk(); ok {
			terraformRequest["q"] = v
		} else if v, ok := datadogRequest.GetApmQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["apm_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetLogQueryOk(); ok {
			terraformQuery := buildTerraformApmOrLogQuery(*v)
			terraformRequest["log_query"] = []map[string]interface{}{terraformQuery}
		} else if v, ok := datadogRequest.GetProcessQueryOk(); ok {
			terraformQuery := buildTerraformProcessQuery(*v)
			terraformRequest["process_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.RumQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.RumQuery)
			terraformRequest["rum_query"] = []map[string]interface{}{terraformQuery}
		} else if datadogRequest.SecurityQuery != nil {
			terraformQuery := buildTerraformApmOrLogQuery(*datadogRequest.SecurityQuery)
			terraformRequest["security_query"] = []map[string]interface{}{terraformQuery}
		}

		if v, ok := datadogRequest.GetConditionalFormatsOk(); ok {
			terraformConditionalFormats := buildTerraformWidgetConditionalFormat(v)
			terraformRequest["conditional_formats"] = terraformConditionalFormats
		}
		if v, ok := datadogRequest.GetStyleOk(); ok {
			style := buildTerraformWidgetRequestStyle(*v)
			terraformRequest["style"] = []map[string]interface{}{style}
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

func buildDatadogTraceServiceDefinition(terraformDefinition map[string]interface{}) *datadogV1.ServiceSummaryWidgetDefinition {
	datadogDefinition := datadogV1.NewServiceSummaryWidgetDefinitionWithDefaults()
	// Required params
	datadogDefinition.SetEnv(terraformDefinition["env"].(string))
	datadogDefinition.SetService(terraformDefinition["service"].(string))
	datadogDefinition.SetSpanName(terraformDefinition["span_name"].(string))
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
		datadogDefinition.SetSizeFormat(datadogV1.WidgetSizeFormat(v))
	}
	if v, ok := terraformDefinition["display_format"].(string); ok && len(v) != 0 {
		datadogDefinition.SetDisplayFormat(datadogV1.WidgetServiceSummaryDisplayFormat(v))
	}
	if v, ok := terraformDefinition["title"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitle(v)
	}
	if v, ok := terraformDefinition["title_size"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleSize(v)
	}
	if v, ok := terraformDefinition["title_align"].(string); ok && len(v) != 0 {
		datadogDefinition.SetTitleAlign(datadogV1.WidgetTextAlign(v))
	}
	if v, ok := terraformDefinition["time"].(map[string]interface{}); ok && len(v) > 0 {
		datadogDefinition.SetTime(*buildDatadogWidgetTime(v))
	}
	return datadogDefinition
}

func buildTerraformTraceServiceDefinition(datadogDefinition datadogV1.ServiceSummaryWidgetDefinition) map[string]interface{} {
	terraformDefinition := map[string]interface{}{}
	// Required params
	terraformDefinition["env"] = datadogDefinition.GetEnv()
	terraformDefinition["service"] = datadogDefinition.GetService()
	terraformDefinition["span_name"] = datadogDefinition.GetSpanName()
	// Optional params
	if v, ok := datadogDefinition.GetShowHitsOk(); ok {
		terraformDefinition["show_hits"] = v
	}
	if v, ok := datadogDefinition.GetShowErrorsOk(); ok {
		terraformDefinition["show_errors"] = v
	}
	if v, ok := datadogDefinition.GetShowLatencyOk(); ok {
		terraformDefinition["show_latency"] = v
	}
	if v, ok := datadogDefinition.GetShowBreakdownOk(); ok {
		terraformDefinition["show_breakdown"] = v
	}
	if v, ok := datadogDefinition.GetShowDistributionOk(); ok {
		terraformDefinition["show_distribution"] = v
	}
	if v, ok := datadogDefinition.GetShowResourceListOk(); ok {
		terraformDefinition["show_resource_list"] = v
	}
	if v, ok := datadogDefinition.GetSizeFormatOk(); ok {
		terraformDefinition["size_format"] = v
	}
	if v, ok := datadogDefinition.GetDisplayFormatOk(); ok {
		terraformDefinition["display_format"] = v
	}
	if v, ok := datadogDefinition.GetTitleOk(); ok {
		terraformDefinition["title"] = v
	}
	if v, ok := datadogDefinition.GetTitleSizeOk(); ok {
		terraformDefinition["title_size"] = v
	}
	if v, ok := datadogDefinition.GetTitleAlignOk(); ok {
		terraformDefinition["title_align"] = v
	}
	if v, ok := datadogDefinition.GetTimeOk(); ok {
		terraformDefinition["time"] = buildTerraformWidgetTime(*v)
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
func buildDatadogWidgetConditionalFormat(terraformWidgetConditionalFormat *[]interface{}) *[]datadogV1.WidgetConditionalFormat {
	datadogWidgetConditionalFormat := make([]datadogV1.WidgetConditionalFormat, len(*terraformWidgetConditionalFormat))
	for i, conditionalFormat := range *terraformWidgetConditionalFormat {
		terraformConditionalFormat := conditionalFormat.(map[string]interface{})
		datadogConditionalFormat := datadogV1.NewWidgetConditionalFormat(
			datadogV1.WidgetComparator(terraformConditionalFormat["comparator"].(string)),
			datadogV1.WidgetPalette(terraformConditionalFormat["palette"].(string)),
			terraformConditionalFormat["value"].(float64))
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
		datadogWidgetConditionalFormat[i] = *datadogConditionalFormat
	}
	return &datadogWidgetConditionalFormat
}
func buildTerraformWidgetConditionalFormat(datadogWidgetConditionalFormat *[]datadogV1.WidgetConditionalFormat) *[]map[string]interface{} {
	terraformWidgetConditionalFormat := make([]map[string]interface{}, len(*datadogWidgetConditionalFormat))
	for i, datadogConditionalFormat := range *datadogWidgetConditionalFormat {
		terraformConditionalFormat := map[string]interface{}{}
		// Required params
		terraformConditionalFormat["comparator"] = datadogConditionalFormat.GetComparator()
		terraformConditionalFormat["value"] = datadogConditionalFormat.GetValue()
		terraformConditionalFormat["palette"] = datadogConditionalFormat.GetPalette()
		// Optional params
		if datadogConditionalFormat.CustomBgColor != nil {
			terraformConditionalFormat["custom_bg_color"] = datadogConditionalFormat.GetCustomBgColor()
		}
		if v, ok := datadogConditionalFormat.GetCustomFgColorOk(); ok {
			terraformConditionalFormat["custom_fg_color"] = v
		}
		if v, ok := datadogConditionalFormat.GetImageUrlOk(); ok {
			terraformConditionalFormat["image_url"] = v
		}
		if v, ok := datadogConditionalFormat.GetHideValueOk(); ok {
			terraformConditionalFormat["hide_value"] = v
		}
		if v, ok := datadogConditionalFormat.GetTimeframeOk(); ok {
			terraformConditionalFormat["timeframe"] = v
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
		"tags_execution": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}
}
func buildDatadogWidgetEvents(terraformWidgetEvents *[]interface{}) *[]datadogV1.WidgetEvent {
	datadogWidgetEvents := make([]datadogV1.WidgetEvent, len(*terraformWidgetEvents))
	for i, event := range *terraformWidgetEvents {
		terraformEvent := event.(map[string]interface{})
		datadogWidgetEvent := datadogV1.NewWidgetEvent(terraformEvent["q"].(string))
		if v, ok := terraformEvent["tags_execution"].(string); ok && len(v) > 0 {
			datadogWidgetEvent.SetTagsExecution(v)
		}
		datadogWidgetEvents[i] = *datadogWidgetEvent
	}

	return &datadogWidgetEvents
}
func buildTerraformWidgetEvents(datadogWidgetEvents *[]datadogV1.WidgetEvent) *[]map[string]string {
	terraformWidgetEvents := make([]map[string]string, len(*datadogWidgetEvents))
	for i, datadogWidget := range *datadogWidgetEvents {
		terraformWidget := map[string]string{}
		// Required params
		terraformWidget["q"] = datadogWidget.GetQ()
		// Optional params
		if v, ok := datadogWidget.GetTagsExecutionOk(); ok {
			terraformWidget["tags_execution"] = *v
		}

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
func buildDatadogWidgetTime(terraformWidgetTime map[string]interface{}) *datadogV1.WidgetTime {
	datadogWidgetTime := &datadogV1.WidgetTime{}
	if v, ok := terraformWidgetTime["live_span"].(string); ok && len(v) != 0 {
		datadogWidgetTime.SetLiveSpan(datadogV1.WidgetLiveSpan(v))
	}
	return datadogWidgetTime
}
func buildTerraformWidgetTime(datadogWidgetTime datadogV1.WidgetTime) map[string]string {
	terraformWidgetTime := map[string]string{}
	if v, ok := datadogWidgetTime.GetLiveSpanOk(); ok {
		terraformWidgetTime["live_span"] = string(*v)
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
func buildDatadogWidgetMarkers(terraformWidgetMarkers *[]interface{}) *[]datadogV1.WidgetMarker {
	datadogWidgetMarkers := make([]datadogV1.WidgetMarker, len(*terraformWidgetMarkers))
	for i, marker := range *terraformWidgetMarkers {
		terraformMarker := marker.(map[string]interface{})
		// Required
		datadogMarker := datadogV1.NewWidgetMarker(terraformMarker["value"].(string))
		// Optional
		if v, ok := terraformMarker["display_type"].(string); ok && len(v) != 0 {
			datadogMarker.SetDisplayType(v)
		}
		if v, ok := terraformMarker["label"].(string); ok && len(v) != 0 {
			datadogMarker.SetLabel(v)
		}
		datadogWidgetMarkers[i] = *datadogMarker
	}
	return &datadogWidgetMarkers
}
func buildTerraformWidgetMarkers(datadogWidgetMarkers *[]datadogV1.WidgetMarker) *[]map[string]string {
	terraformWidgetMarkers := make([]map[string]string, len(*datadogWidgetMarkers))
	for i, datadogMarker := range *datadogWidgetMarkers {
		terraformMarker := map[string]string{}
		// Required params
		terraformMarker["value"] = datadogMarker.Value
		// Optional params
		if v, ok := datadogMarker.GetDisplayTypeOk(); ok {
			terraformMarker["display_type"] = *v
		}
		if v, ok := datadogMarker.GetLabelOk(); ok {
			terraformMarker["label"] = *v
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

// APM, Log, Network or RUM Query
func getApmLogNetworkOrRumQuerySchema() *schema.Schema {
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
				"compute": {
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
				"search": {
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
				"group_by": {
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
func buildDatadogApmOrLogQuery(terraformQuery map[string]interface{}) *datadogV1.LogQueryDefinition {
	// Index
	datadogQuery := datadogV1.NewLogQueryDefinition()
	datadogQuery.SetIndex(terraformQuery["index"].(string))

	// Compute
	terraformCompute := terraformQuery["compute"].(map[string]interface{})
	datadogCompute := datadogV1.NewLogsQueryComputeWithDefaults()
	if aggr, ok := terraformCompute["aggregation"].(string); ok && len(aggr) != 0 {
		datadogCompute.SetAggregation(aggr)
	}
	if facet, ok := terraformCompute["facet"].(string); ok && len(facet) != 0 {
		datadogCompute.SetFacet(facet)
	}
	if interval, ok := terraformCompute["interval"].(string); ok {
		if v, err := strconv.ParseInt(interval, 10, 64); err == nil {
			datadogCompute.SetInterval(v)
		}
	}
	datadogQuery.SetCompute(*datadogCompute)
	// Search
	if terraformSearch, ok := terraformQuery["search"].(map[string]interface{}); ok && len(terraformSearch) > 0 {
		datadogQuery.Search = &datadogV1.LogQueryDefinitionSearch{
			Query: terraformSearch["query"].(string),
		}
	}
	// GroupBy
	if terraformGroupBys, ok := terraformQuery["group_by"].([]interface{}); ok && len(terraformGroupBys) > 0 {
		datadogGroupBys := make([]datadogV1.LogQueryDefinitionGroupBy, len(terraformGroupBys))
		for i, g := range terraformGroupBys {
			groupBy := g.(map[string]interface{})
			// Facet
			datadogGroupBy := datadogV1.NewLogQueryDefinitionGroupBy(groupBy["facet"].(string))
			// Limit
			if v, ok := groupBy["limit"].(int); ok && v != 0 {
				datadogGroupBy.SetLimit(int64(v))
			}
			// Sort
			if sort, ok := groupBy["sort"].(map[string]interface{}); ok && len(sort) > 0 {

				datadogGroupBy.Sort = &datadogV1.LogQueryDefinitionSort{}
				if aggr, ok := sort["aggregation"].(string); ok && len(aggr) > 0 {
					datadogGroupBy.Sort.SetAggregation(aggr)
				}
				if order, ok := sort["order"].(string); ok && len(order) > 0 {
					datadogGroupBy.Sort.SetOrder(datadogV1.WidgetSort(order))
				}
				if facet, ok := sort["facet"].(string); ok && len(facet) > 0 {
					datadogGroupBy.Sort.SetFacet(facet)
				}
			}
			datadogGroupBys[i] = *datadogGroupBy
		}
		datadogQuery.SetGroupBy(datadogGroupBys)
	}
	return datadogQuery
}
func buildTerraformApmOrLogQuery(datadogQuery datadogV1.LogQueryDefinition) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	// Index
	terraformQuery["index"] = datadogQuery.GetIndex()
	// Compute
	terraformCompute := map[string]interface{}{
		"aggregation": datadogQuery.Compute.GetAggregation(),
	}
	if v, ok := datadogQuery.Compute.GetFacetOk(); ok {
		terraformCompute["facet"] = *v
	}
	if datadogQuery.Compute.Interval != nil {
		terraformCompute["interval"] = strconv.FormatInt(*datadogQuery.Compute.Interval, 10)
	}
	terraformQuery["compute"] = terraformCompute
	// Search
	if datadogQuery.Search != nil {
		terraformQuery["search"] = map[string]interface{}{
			"query": datadogQuery.Search.Query,
		}
	}
	// GroupBy
	if v, ok := datadogQuery.GetGroupByOk(); ok {
		terraformGroupBys := make([]map[string]interface{}, len(datadogQuery.GetGroupBy()))
		for i, groupBy := range *v {
			// Facet
			terraformGroupBy := map[string]interface{}{
				"facet": groupBy.GetFacet(),
			}
			// Limit
			if v, ok := groupBy.GetLimitOk(); ok {
				terraformGroupBy["limit"] = *v
			}
			// Sort
			if v, ok := groupBy.GetSortOk(); ok {
				sort := map[string]string{
					"aggregation": v.GetAggregation(),
					"order":       string(v.GetOrder()),
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
func buildDatadogProcessQuery(terraformQuery map[string]interface{}) *datadogV1.ProcessQueryDefinition {
	datadogQuery := datadogV1.NewProcessQueryDefinitionWithDefaults()
	if v, ok := terraformQuery["metric"].(string); ok && len(v) != 0 {
		datadogQuery.SetMetric(v)
	}
	if v, ok := terraformQuery["search_by"].(string); ok && len(v) != 0 {
		datadogQuery.SetSearchBy(v)
	}

	if terraformFilterBys, ok := terraformQuery["filter_by"].([]interface{}); ok && len(terraformFilterBys) > 0 {
		datadogFilterbys := make([]string, len(terraformFilterBys))
		for i, filterBy := range terraformFilterBys {
			datadogFilterbys[i] = filterBy.(string)
		}
		datadogQuery.SetFilterBy(datadogFilterbys)
	}

	if v, ok := terraformQuery["limit"].(int); ok && v != 0 {
		datadogQuery.SetLimit(int64(v))
	}

	return datadogQuery
}

func buildTerraformProcessQuery(datadogQuery datadogV1.ProcessQueryDefinition) map[string]interface{} {
	terraformQuery := map[string]interface{}{}
	if v, ok := datadogQuery.GetMetricOk(); ok {
		terraformQuery["metric"] = v
	}
	if v, ok := datadogQuery.GetSearchByOk(); ok {
		terraformQuery["search_by"] = v
	}
	if v, ok := datadogQuery.GetFilterByOk(); ok {
		terraformFilterBys := make([]string, len(*v))
		for i, datadogFilterBy := range *v {
			terraformFilterBys[i] = datadogFilterBy
		}
		terraformQuery["filter_by"] = terraformFilterBys
	}
	if v, ok := datadogQuery.GetLimitOk(); ok {
		terraformQuery["limit"] = v
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

func buildDatadogWidgetAxis(terraformWidgetAxis map[string]interface{}) *datadogV1.WidgetAxis {
	datadogWidgetAxis := &datadogV1.WidgetAxis{}
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

func buildTerraformWidgetAxis(datadogWidgetAxis datadogV1.WidgetAxis) map[string]interface{} {
	terraformWidgetAxis := map[string]interface{}{}
	if v, ok := datadogWidgetAxis.GetLabelOk(); ok {
		terraformWidgetAxis["label"] = v
	}
	if v, ok := datadogWidgetAxis.GetScaleOk(); ok {
		terraformWidgetAxis["scale"] = v
	}
	if v, ok := datadogWidgetAxis.GetMinOk(); ok {
		terraformWidgetAxis["min"] = v
	}
	if v, ok := datadogWidgetAxis.GetMaxOk(); ok {
		terraformWidgetAxis["max"] = v
	}
	if v, ok := datadogWidgetAxis.GetIncludeZeroOk(); ok {
		terraformWidgetAxis["include_zero"] = v
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
func buildDatadogWidgetStyle(terraformStyle map[string]interface{}) *datadogV1.WidgetStyle {
	datadogStyle := &datadogV1.WidgetStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}

	return datadogStyle
}
func buildTerraformWidgetStyle(datadogStyle datadogV1.WidgetStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if v, ok := datadogStyle.GetPaletteOk(); ok {
		terraformStyle["palette"] = v
	}
	return terraformStyle
}

// Timeseriest Style helpers

func buildDatadogWidgetRequestStyle(terraformStyle map[string]interface{}) *datadogV1.WidgetRequestStyle {
	datadogStyle := &datadogV1.WidgetRequestStyle{}
	if v, ok := terraformStyle["palette"].(string); ok && len(v) != 0 {
		datadogStyle.SetPalette(v)
	}
	if v, ok := terraformStyle["line_type"].(string); ok && len(v) != 0 {
		datadogStyle.SetLineType(datadogV1.WidgetLineType(v))
	}
	if v, ok := terraformStyle["line_width"].(string); ok && len(v) != 0 {
		datadogStyle.SetLineWidth(datadogV1.WidgetLineWidth(v))
	}

	return datadogStyle
}
func buildTerraformWidgetRequestStyle(datadogStyle datadogV1.WidgetRequestStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if v, ok := datadogStyle.GetPaletteOk(); ok {
		terraformStyle["palette"] = v
	}
	if v, ok := datadogStyle.GetLineTypeOk(); ok {
		terraformStyle["line_type"] = v
	}
	if v, ok := datadogStyle.GetLineWidthOk(); ok {
		terraformStyle["line_width"] = v
	}
	return terraformStyle
}

// Hostmap Style helpers

func buildDatadogHostmapRequestStyle(terraformStyle map[string]interface{}) *datadogV1.HostMapWidgetDefinitionStyle {
	datadogStyle := &datadogV1.HostMapWidgetDefinitionStyle{}
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
func buildTerraformHostmapRequestStyle(datadogStyle datadogV1.HostMapWidgetDefinitionStyle) map[string]interface{} {
	terraformStyle := map[string]interface{}{}
	if datadogStyle.Palette != nil {
		terraformStyle["palette"] = datadogStyle.GetPalette()
	}
	if datadogStyle.PaletteFlip != nil {
		terraformStyle["palette_flip"] = datadogStyle.GetPaletteFlip()
	}
	if datadogStyle.FillMin != nil {
		terraformStyle["fill_min"] = datadogStyle.GetFillMin()
	}
	if datadogStyle.FillMax != nil {
		terraformStyle["fill_max"] = datadogStyle.GetFillMax()
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
func validateTimeseriesWidgetLegendSize(val interface{}, key string) (warns []string, errs []error) {
	value := val.(string)
	switch value {
	case "0", "2", "4", "8", "16", "auto":
		break
	default:
		errs = append(errs, fmt.Errorf(
			"%q contains an invalid value %q. Valud values are `2`, `4`, `8`, `16`, or `auto`", key, value))
	}
	return
}
